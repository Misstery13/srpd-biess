package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	PermisoTTL = 1 * time.Hour
)

type CachedPermisoRepository struct {
	repo repository.PermisoRepository
	rdb  *redis.Client
}

// NewCachedPermisoRepository optimiza la autorización RBAC cacheando las reglas
// activas en Redis para evitar consultas recurrentes a PostgreSQL en cada petición HTTP (Issue #24).
func NewCachedPermisoRepository(repo repository.PermisoRepository, rdb *redis.Client) repository.PermisoRepository {
	return &CachedPermisoRepository{
		repo: repo,
		rdb:  rdb,
	}
}

func (c *CachedPermisoRepository) TienePermiso(ctx context.Context, rolCodigo, metodo, ruta string) (bool, error) {
	key := fmt.Sprintf("rbac:permiso:%s:%s:%s", rolCodigo, metodo, ruta)

	if cached, ok := c.obtenerDeCache(ctx, key); ok {
		return cached, nil
	}

	tiene, err := c.repo.TienePermiso(ctx, rolCodigo, metodo, ruta)
	if err != nil {
		return false, err
	}

	c.guardarEnCache(ctx, key, tiene)
	return tiene, nil
}

func (c *CachedPermisoRepository) obtenerDeCache(ctx context.Context, key string) (bool, bool) {
	if c.rdb == nil {
		return false, false
	}

	ctxGet, cancel := context.WithTimeout(ctx, redisOpTimeout)
	defer cancel()

	val, err := c.rdb.Get(ctxGet, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("[CACHE WARN] Redis GET %s fallo: %v. Fallback a BD.", key, err)
		}
		return false, false
	}

	return val == "1", true
}

func (c *CachedPermisoRepository) guardarEnCache(ctx context.Context, key string, tiene bool) {
	if c.rdb == nil {
		return
	}

	val := "0"
	if tiene {
		val = "1"
	}

	ctxSet, cancel := context.WithTimeout(ctx, redisOpTimeout)
	defer cancel()

	_ = c.rdb.Set(ctxSet, key, val, PermisoTTL).Err()
}
