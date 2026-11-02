package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

const (
	PeriodoCacheKey = "periodo:activo"
	PeriodoTTL      = 1 * time.Hour
)

type CachedPeriodoRepository struct {
	repo repository.PeriodoRepository
	rdb  *redis.Client
}

// NewCachedPeriodoRepository decora el repositorio de período con caché Redis
// para evitar consultas constantes a PostgreSQL (Issue #24).
func NewCachedPeriodoRepository(repo repository.PeriodoRepository, rdb *redis.Client) repository.PeriodoRepository {
	return &CachedPeriodoRepository{
		repo: repo,
		rdb:  rdb,
	}
}

func (c *CachedPeriodoRepository) ObtenerPeriodoActivo(ctx context.Context) (*repository.PeriodoRow, error) {
	if c.rdb != nil {
		ctxGet, cancel := context.WithTimeout(ctx, redisOpTimeout)
		val, err := c.rdb.Get(ctxGet, PeriodoCacheKey).Result()
		cancel()

		if err == nil {
			var row repository.PeriodoRow
			if err := json.Unmarshal([]byte(val), &row); err == nil {
				return &row, nil
			}
		} else if !errors.Is(err, redis.Nil) {
			log.Printf("[CACHE WARN] Redis no disponible para GET %s: %v. Fallback a PostgreSQL.", PeriodoCacheKey, err)
		}
	}

	// Fallback a PostgreSQL
	row, err := c.repo.ObtenerPeriodoActivo(ctx)
	if err != nil {
		return nil, err
	}

	// Guardar en Redis
	if c.rdb != nil && row != nil {
		if data, err := json.Marshal(row); err == nil {
			ctxSet, cancel := context.WithTimeout(ctx, redisOpTimeout)
			defer cancel()
			_ = c.rdb.Set(ctxSet, PeriodoCacheKey, data, PeriodoTTL).Err()
		}
	}

	return row, nil
}

