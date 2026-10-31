package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	JWTSecret        string
	SMTPHost         string
	SMTPPort         string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string
	RedisAddr        string
	RedisPassword    string
	UploadsDir       string
}

func Load() (*Config, error) {
	// Busca .env en el directorio actual o en el directorio raíz padre
	for _, envPath := range []string{".env", "../.env"} {
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
	}

	port := getEnv("APP_PORT", "8080")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Regla O2: JWT_SECRET de al menos 32 caracteres y no el valor por defecto
	if err := ValidateJWTSecret(jwtSecret); err != nil {
		return nil, err
	}

	cfg := &Config{
		AppPort:          port,
		JWTSecret:        jwtSecret,
		SMTPHost:         getEnv("SMTP_HOST", "localhost"),
		SMTPPort:         getEnv("SMTP_PORT", "1025"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnv("POSTGRES_DB", "srpd_biess"),
		PostgresUser:     getEnv("POSTGRES_USER", "biess"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "biess"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		UploadsDir:       getEnv("UPLOADS_DIR", "uploads"),
	}

	return cfg, nil
}

func ValidateJWTSecret(secret string) error {
	if secret == "" || secret == "cambia-esto" || len(secret) < 32 {
		return errors.New("JWT_SECRET inválido o débil: debe tener al menos 32 caracteres y no ser el valor de ejemplo")
	}
	return nil
}

func (c *Config) PostgresDSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   fmt.Sprintf("%s:%s", c.PostgresHost, c.PostgresPort),
		Path:   c.PostgresDB,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
