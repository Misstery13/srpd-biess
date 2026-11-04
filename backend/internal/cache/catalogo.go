package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	catalogoTTL    = 24 * time.Hour
	redisOpTimeout = 300 * time.Millisecond
)

type CachedCatalogoRepository struct {
	repo repository.CatalogoRepository
	rdb  *redis.Client
}

func NewCachedCatalogoRepository(repo repository.CatalogoRepository, rdb *redis.Client) repository.CatalogoRepository {
	return &CachedCatalogoRepository{
		repo: repo,
		rdb:  rdb,
	}
}

// obtenerConCache unifica la lectura con timeout corto, fallback a PostgreSQL
// y evita reintentos de Set si Redis no responde (Bloqueante 2 y Menor 4).
func obtenerConCache[T any](
	ctx context.Context,
	rdb *redis.Client,
	key string,
	cargar func(context.Context) ([]T, error),
) ([]T, error) {
	redisOperativo := false

	if rdb != nil {
		ctxGet, cancelGet := context.WithTimeout(ctx, redisOpTimeout)
		val, err := rdb.Get(ctxGet, key).Result()
		cancelGet()

		if err == nil {
			var items []T
			if err := json.Unmarshal([]byte(val), &items); err == nil {
				return items, nil
			}
		} else if errors.Is(err, redis.Nil) {
			// Clave no encontrada, pero Redis respondió con normalidad
			redisOperativo = true
		} else {
			// Error de conexión, timeout o Redis caído
			log.Printf("[CACHE WARN] Redis no disponible para GET %s: %v. Fallback a PostgreSQL.", key, err)
		}
	}

	// Fallback a PostgreSQL
	items, err := cargar(ctx)
	if err != nil {
		return nil, err
	}

	// Solo intenta guardar en Redis si no falló por conexión
	if rdb != nil && redisOperativo && len(items) > 0 {
		if data, err := json.Marshal(items); err == nil {
			ctxSet, cancelSet := context.WithTimeout(ctx, redisOpTimeout)
			defer cancelSet()
			if err := rdb.Set(ctxSet, key, data, catalogoTTL).Err(); err != nil {
				log.Printf("[CACHE WARN] No se pudo cachear %s en Redis: %v", key, err)
			}
		}
	}

	return items, nil
}

func (c *CachedCatalogoRepository) ListarPaises(ctx context.Context) ([]model.Pais, error) {
	return obtenerConCache(ctx, c.rdb, "catalogo:paises", c.repo.ListarPaises)
}

func (c *CachedCatalogoRepository) ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	key := fmt.Sprintf("catalogo:provincias:%d", paisID)
	return obtenerConCache(ctx, c.rdb, key, func(ctx context.Context) ([]model.Provincia, error) {
		return c.repo.ListarProvincias(ctx, paisID)
	})
}

func (c *CachedCatalogoRepository) ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	key := fmt.Sprintf("catalogo:cantones:%d", provinciaID)
	return obtenerConCache(ctx, c.rdb, key, func(ctx context.Context) ([]model.Canton, error) {
		return c.repo.ListarCantones(ctx, provinciaID)
	})
}
