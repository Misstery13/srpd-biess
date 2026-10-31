package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/cache"
	"github.com/Misstery13/srpd-biess/backend/internal/config"
	"github.com/Misstery13/srpd-biess/backend/internal/handler"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Cargar y validar configuración (.env)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error de configuración: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Conexión a PostgreSQL (pgxpool)
	poolConfig, err := pgxpool.ParseConfig(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("Error al parsear configuración de PostgreSQL: %v", err)
	}
	poolConfig.MaxConns = 15
	poolConfig.MinConns = 2

	pgPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Error al crear pool de PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL")

	// 3. Conexión a Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[ALERTA] No se pudo conectar a Redis (%v). Se usará fallback directo a PostgreSQL.", err)
	} else {
		log.Println("Conexión exitosa a Redis")
	}

	// 4. Inyección de dependencias (Repository -> Service -> Handler)
	catalogoRepo := repository.NewCatalogoRepository(pgPool)
	cachedCatalogoRepo := cache.NewCachedCatalogoRepository(catalogoRepo, redisClient)
	catalogoSvc := service.NewCatalogoService(cachedCatalogoRepo)
	catalogoHdl := handler.NewCatalogoHandler(catalogoSvc)

	// 5. Configuración de Gin
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(handler.SecurityHeadersMiddleware())
	router.Use(handler.ErrorMiddleware())

	// Ruta de health check (mantenida para despliegue y monitoreo)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Endpoints de catálogos
	catalogosGroup := router.Group("/api/catalogos")
	{
		catalogosGroup.GET("/paises", catalogoHdl.ListarPaises)
		catalogosGroup.GET("/paises/:id/provincias", catalogoHdl.ListarProvincias)
		catalogosGroup.GET("/provincias/:id/cantones", catalogoHdl.ListarCantones)
	}

	log.Printf("Servidor API escuchando en el puerto :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
