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

	if c.rdb != nil {
		ctxGet, cancel := context.WithTimeout(ctx, redisOpTimeout)
		val, err := c.rdb.Get(ctxGet, key).Result()
		cancel()

		if err == nil {
			return val == "1", nil
		} else if !errors.Is(err, redis.Nil) {
			log.Printf("[CACHE WARN] Redis no disponible para GET %s: %v. Fallback a PostgreSQL.", key, err)
		}
	}

	// Fallback a PostgreSQL
	tiene, err := c.repo.TienePermiso(ctx, rolCodigo, metodo, ruta)
	if err != nil {
		return false, err
	}

	// Cachear resultado en Redis
	if c.rdb != nil {
		valToSet := "0"
		if tiene {
			valToSet = "1"
		}
		ctxSet, cancel := context.WithTimeout(ctx, redisOpTimeout)
		defer cancel()
		_ = c.rdb.Set(ctxSet, key, valToSet, PermisoTTL).Err()
	}

	return tiene, nil
}

