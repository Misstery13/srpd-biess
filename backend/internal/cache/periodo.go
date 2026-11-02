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
// para optimizar el acceso recurrente y evitar sobrecarga en PostgreSQL (Issue #24).
func NewCachedPeriodoRepository(repo repository.PeriodoRepository, rdb *redis.Client) repository.PeriodoRepository {
	return &CachedPeriodoRepository{
		repo: repo,
		rdb:  rdb,
	}
}

func (c *CachedPeriodoRepository) ObtenerPeriodoActivo(ctx context.Context) (*repository.PeriodoRow, error) {
	if cached := c.obtenerDeCache(ctx); cached != nil {
		return cached, nil
	}

	row, err := c.repo.ObtenerPeriodoActivo(ctx)
	if err != nil {
		return nil, err
	}

	c.guardarEnCache(ctx, row)
	return row, nil
}

func (c *CachedPeriodoRepository) obtenerDeCache(ctx context.Context) *repository.PeriodoRow {
	if c.rdb == nil {
		return nil
	}

	ctxGet, cancel := context.WithTimeout(ctx, redisOpTimeout)
	defer cancel()

	val, err := c.rdb.Get(ctxGet, PeriodoCacheKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Printf("[CACHE WARN] Redis GET %s fallo: %v. Fallback a BD.", PeriodoCacheKey, err)
		}
		return nil
	}

	var row repository.PeriodoRow
	if err := json.Unmarshal([]byte(val), &row); err != nil {
		return nil
	}
	return &row
}

func (c *CachedPeriodoRepository) guardarEnCache(ctx context.Context, row *repository.PeriodoRow) {
	if c.rdb == nil || row == nil {
		return
	}

	data, err := json.Marshal(row)
	if err != nil {
		return
	}

	ctxSet, cancel := context.WithTimeout(ctx, redisOpTimeout)
	defer cancel()

	_ = c.rdb.Set(ctxSet, PeriodoCacheKey, data, PeriodoTTL).Err()
}
