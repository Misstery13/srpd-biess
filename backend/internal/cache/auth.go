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
	IntentosBloqueoTTL = 600 * time.Second // 10 minutos (S4)
	VerificadoTTL      = 600 * time.Second // 10 minutos (para permitir completar paso 2)
	ResetTokenTTL      = 900 * time.Second // 15 minutos (I3)
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
	rdb        *redis.Client
	hmacSecret string
	aesKey     []byte
}

// NewAuthCache crea una instancia de caché respaldada por Redis con soporte para
// cifrado AES-256-GCM y pseudonimización de identificadores sensibles (Issue #24).
// Utiliza derivación de claves con dominios separados para AES y HMAC.
func NewAuthCache(rdb *redis.Client, secret string) AuthCache {
	return &redisAuthCache{
		rdb:        rdb,
		hmacSecret: "hmac|" + secret,
		aesKey:     DerivarClaveAES32("aes|" + secret),
	}
}

func (c *redisAuthCache) keyIntentos(tipo, cedula string) string {
	// Pseudonimización HMAC-SHA256 para no exponer la cédula en texto plano en las claves de Redis
	pseudo := PseudonimizarClave(c.hmacSecret, cedula)
	return fmt.Sprintf("intentos:%s:%s", tipo, pseudo)
}

func (c *redisAuthCache) keyVerificado(cedula string) string {
	pseudo := PseudonimizarClave(c.hmacSecret, cedula)
	return fmt.Sprintf("registro:verificado:%s", pseudo)
}

func (c *redisAuthCache) IncrementarIntentos(ctx context.Context, tipo, cedula string) (int64, error) {
	if c.rdb == nil {
		return 0, ErrRedisNoDisponible
	}
	key := c.keyIntentos(tipo, cedula)

	pipe := c.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, IntentosBloqueoTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (c *redisAuthCache) EstaBloqueado(ctx context.Context, tipo, cedula string) (bool, error) {
	if c.rdb == nil {
		return false, ErrRedisNoDisponible
	}
	key := c.keyIntentos(tipo, cedula)

	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	intentos, err := strconv.Atoi(val)
	return err == nil && intentos >= MaxIntentos, nil
}

func (c *redisAuthCache) ReiniciarIntentos(ctx context.Context, tipo, cedula string) error {
	if c.rdb == nil {
		return nil
	}
	key := c.keyIntentos(tipo, cedula)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisAuthCache) GuardarVerificado(ctx context.Context, cedula string) error {
	if c.rdb == nil {
		return ErrRedisNoDisponible
	}
	key := c.keyVerificado(cedula)
	return c.rdb.Set(ctx, key, "1", VerificadoTTL).Err()
}

func (c *redisAuthCache) EsVerificado(ctx context.Context, cedula string) (bool, error) {
	if c.rdb == nil {
		return false, ErrRedisNoDisponible
	}
	key := c.keyVerificado(cedula)
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
	key := c.keyVerificado(cedula)
	return c.rdb.Del(ctx, key).Err()
}

func (c *redisAuthCache) GuardarResetToken(ctx context.Context, token, cedula string) error {
	if c.rdb == nil {
		return ErrRedisNoDisponible
	}

	// LOPDP / OWASP A02: Cifrado autenticado AES-256-GCM antes de persistir en Redis
	cifrado, err := CifrarAESGCM(c.aesKey, cedula)
	if err != nil {
		return fmt.Errorf("error al cifrar dato sensible para cache: %w", err)
	}

	key := fmt.Sprintf("reset:%s", token)
	return c.rdb.Set(ctx, key, cifrado, ResetTokenTTL).Err()
}

func (c *redisAuthCache) ObtenerYBorrarResetToken(ctx context.Context, token string) (string, error) {
	if c.rdb == nil {
		return "", ErrRedisNoDisponible
	}
	key := fmt.Sprintf("reset:%s", token)
	cifrado, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}

	_ = c.rdb.Del(ctx, key).Err()

	descifrado, err := DescifrarAESGCM(c.aesKey, cifrado)
	if err != nil {
		// Token corrupto o manipulado: se invalida
		return "", nil
	}
	return descifrado, nil
}
