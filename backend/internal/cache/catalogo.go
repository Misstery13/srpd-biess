package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const catalogoTTL = 24 * time.Hour

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

func (c *CachedCatalogoRepository) ListarPaises(ctx context.Context) ([]model.Pais, error) {
	key := "catalogo:paises"

	if c.rdb != nil {
		val, err := c.rdb.Get(ctx, key).Result()
		if err == nil {
			var paises []model.Pais
			if err := json.Unmarshal([]byte(val), &paises); err == nil {
				return paises, nil
			}
		}
	}

	// Fallback a PostgreSQL
	paises, err := c.repo.ListarPaises(ctx)
	if err != nil {
		return nil, err
	}

	// Guardar en caché
	if c.rdb != nil && len(paises) > 0 {
		if data, err := json.Marshal(paises); err == nil {
			if err := c.rdb.Set(ctx, key, data, catalogoTTL).Err(); err != nil {
				log.Printf("[CACHE WARN] No se pudo cachear %s en Redis: %v", key, err)
			}
		}
	}

	return paises, nil
}

func (c *CachedCatalogoRepository) ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	key := fmt.Sprintf("catalogo:provincias:%d", paisID)

	if c.rdb != nil {
		val, err := c.rdb.Get(ctx, key).Result()
		if err == nil {
			var provincias []model.Provincia
			if err := json.Unmarshal([]byte(val), &provincias); err == nil {
				return provincias, nil
			}
		}
	}

	// Fallback a PostgreSQL
	provincias, err := c.repo.ListarProvincias(ctx, paisID)
	if err != nil {
		return nil, err
	}

	// Guardar en caché
	if c.rdb != nil && len(provincias) > 0 {
		if data, err := json.Marshal(provincias); err == nil {
			if err := c.rdb.Set(ctx, key, data, catalogoTTL).Err(); err != nil {
				log.Printf("[CACHE WARN] No se pudo cachear %s en Redis: %v", key, err)
			}
		}
	}

	return provincias, nil
}

func (c *CachedCatalogoRepository) ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	key := fmt.Sprintf("catalogo:cantones:%d", provinciaID)

	if c.rdb != nil {
		val, err := c.rdb.Get(ctx, key).Result()
		if err == nil {
			var cantones []model.Canton
			if err := json.Unmarshal([]byte(val), &cantones); err == nil {
				return cantones, nil
			}
		}
	}

	// Fallback a PostgreSQL
	cantones, err := c.repo.ListarCantones(ctx, provinciaID)
	if err != nil {
		return nil, err
	}

	// Guardar en caché
	if c.rdb != nil && len(cantones) > 0 {
		if data, err := json.Marshal(cantones); err == nil {
			if err := c.rdb.Set(ctx, key, data, catalogoTTL).Err(); err != nil {
				log.Printf("[CACHE WARN] No se pudo cachear %s en Redis: %v", key, err)
			}
		}
	}

	return cantones, nil
}
