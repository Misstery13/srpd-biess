package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	IntentosBloqueoTTL = 600 * time.Second  // 10 minutos (S4)
	VerificadoTTL      = 600 * time.Second  // 10 minutos (para permitir completar paso 2)
	ResetTokenTTL      = 900 * time.Second  // 15 minutos (I3)
	MaxIntentos        = 3
)

var ErrRedisNoDisponible = errors.New("servicio de caché temporalmente no disponible")

type AuthCache interface {
	IncrementarIntentos(ctx context.Context, tipo, cedula string) (int64, error)
	EstaBloqueado(ctx context.Context, tipo, cedula string) (bool, error)
	ReiniciarIntentos(ctx context.Context, tipo, cedula string) error
	GuardarVerificado(ctx context.Context, cedula string) error
	EsVerificado(ctx context.Context, cedula string) (bool, error)
	BorrarVerificado(ctx context.Context, cedula string) error
	GuardarResetToken(ctx context.Context, token, cedula string) error
	ObtenerYBorrarResetToken(ctx context.Context, token string) (string, error)
}

type redisAuthCache struct {
	rdb *redis.Client
}

func NewAuthCache(rdb *redis.Client) AuthCache {
	return &redisAuthCache{rdb: rdb}
}

func (c *redisAuthCache) IncrementarIntentos(ctx context.Context, tipo, cedula string) (int64, error) {
	if c.rdb == nil {
		return 0, ErrRedisNoDisponible
	}
	key := fmt.Sprintf("intentos:%s:%s", tipo, cedula)

	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if n == 1 {
		c.rdb.Expire(ctx, key, IntentosBloqueoTTL)
	}

	return n, nil
}

func (c *redisAuthCache) EstaBloqueado(ctx context.Context, tipo, cedula string) (bool, error) {
	if c.rdb == nil {
		return false, ErrRedisNoDisponible
	}
	key := fmt.Sprintf("intentos:%s:%s", tipo, cedula)

	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	intentos, err := strconv.Atoi(val)
	if err != nil {
		return false, nil
	}

	return intentos >= MaxIntentos, nil
}

func (c *redisAuthCache) ReiniciarIntentos(ctx context.Context, tipo, cedula string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("intentos:%s:%s", tipo, cedula)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisAuthCache) GuardarVerificado(ctx context.Context, cedula string) error {
	if c.rdb == nil {
		return ErrRedisNoDisponible
	}
	key := fmt.Sprintf("registro:verificado:%s", cedula)
	return c.rdb.Set(ctx, key, "1", VerificadoTTL).Err()
}

func (c *redisAuthCache) EsVerificado(ctx context.Context, cedula string) (bool, error) {
	if c.rdb == nil {
		return false, ErrRedisNoDisponible
	}
	key := fmt.Sprintf("registro:verificado:%s", cedula)
	n, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *redisAuthCache) BorrarVerificado(ctx context.Context, cedula string) error {
	if c.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("registro:verificado:%s", cedula)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisAuthCache) GuardarResetToken(ctx context.Context, token, cedula string) error {
	if c.rdb == nil {
		return ErrRedisNoDisponible
	}
	key := fmt.Sprintf("reset:%s", token)
	return c.rdb.Set(ctx, key, cedula, ResetTokenTTL).Err()
}

func (c *redisAuthCache) ObtenerYBorrarResetToken(ctx context.Context, token string) (string, error) {
	if c.rdb == nil {
		return "", ErrRedisNoDisponible
	}
	key := fmt.Sprintf("reset:%s", token)
	cedula, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	// De un solo uso: se borra de inmediato (I3)
	_ = c.rdb.Del(ctx, key).Err()
	return cedula, nil
}
