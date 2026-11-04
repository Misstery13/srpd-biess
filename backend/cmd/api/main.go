package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error de configuración: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool := initPostgres(ctx, cfg.PostgresDSN())
	defer pgPool.Close()

	rdb := initRedis(ctx, cfg)
	defer rdb.Close()

	router := buildRouter(cfg, pgPool, rdb)

	log.Printf("Servidor API escuchando en el puerto :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func initPostgres(ctx context.Context, dsn string) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Error al parsear configuración de PostgreSQL: %v", err)
	}
	poolConfig.MaxConns = 15
	poolConfig.MinConns = 2

	pgPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Error al crear pool de PostgreSQL: %v", err)
	}

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL")
	return pgPool
}

func initRedis(ctx context.Context, cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
		MaxRetries:   0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("[ALERTA] No se pudo conectar a Redis (%v). Se usará fallback directo a PostgreSQL.", err)
	} else {
		log.Println("Conexión exitosa a Redis")
	}
	return rdb
}

func buildRouter(cfg *config.Config, pgPool *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Logger())
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("[PANIC RECUPERADO] %v", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ocurrió un error interno en el servidor."})
	}))
	router.Use(handler.SecurityHeadersMiddleware())
	router.Use(handler.ErrorMiddleware())

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada."})
	})
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerAppRoutes(router, cfg, pgPool, rdb)
	return router
}

func registerAppRoutes(router *gin.Engine, cfg *config.Config, pgPool *pgxpool.Pool, rdb *redis.Client) {
	// Catálogos
	catalogoRepo := repository.NewCatalogoRepository(pgPool)
	cachedCatalogoRepo := cache.NewCachedCatalogoRepository(catalogoRepo, rdb)
	catalogoSvc := service.NewCatalogoService(cachedCatalogoRepo)
	catalogoHdl := handler.NewCatalogoHandler(catalogoSvc)

	catalogosGroup := router.Group("/api/catalogos")
	{
		catalogosGroup.GET("/paises", catalogoHdl.ListarPaises)
		catalogosGroup.GET("/paises/:id/provincias", catalogoHdl.ListarProvincias)
		catalogosGroup.GET("/provincias/:id/cantones", catalogoHdl.ListarCantones)
	}

	// Autenticación (EF-11)
	rcRepo := repository.NewRegistroCivilRepository(pgPool)
	postRepo := repository.NewPostulanteRepository(pgPool)
	permisoRepo := repository.NewPermisoRepository(pgPool)
	authCache := cache.NewAuthCache(rdb, cfg.JWTSecret)
	emailSvc := service.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.AppURL, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom)
	authSvc := service.NewAuthService(service.AuthConfig{
		RCRepo:    rcRepo,
		PostRepo:  postRepo,
		Cache:     authCache,
		EmailSvc:  emailSvc,
		JWTSecret: cfg.JWTSecret,
	})
	authHdl := handler.NewAuthHandler(authSvc)

	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/verificar", authHdl.Verificar)
		authGroup.POST("/registro", authHdl.Registrar)
		authGroup.POST("/login", authHdl.Login)
		authGroup.POST("/recuperar", authHdl.Recuperar)
		authGroup.POST("/restablecer", authHdl.Restablecer)
	}

	// Periodo (T2 - Consulta directa a BD sin caché para reflejar apertura/cierre en tiempo real)
	periodoRepo := repository.NewPeriodoRepository(pgPool)
	periodoSvc := service.NewPeriodoService(periodoRepo)
	periodoHdl := handler.NewPeriodoHandler(periodoSvc)
	router.GET("/api/periodo", periodoHdl.ObtenerPeriodo)

	// Postulante y Postulación protegidas (R3, R4)
	postulanteSvc := service.NewPostulanteService(postRepo, cfg.UploadsDir)
	postulanteHdl := handler.NewPostulanteHandler(postulanteSvc)
	postulacionRepo := repository.NewPostulacionRepository(pgPool)
	postulacionSvc := service.NewPostulacionService(postulacionRepo, postRepo, periodoSvc, emailSvc, cfg.UploadsDir)
	postulacionHdl := handler.NewPostulacionHandler(postulacionSvc)

	postulanteGroup := router.Group("/api/postulante")
	postulanteGroup.Use(handler.JWTMiddleware(cfg.JWTSecret))
	postulanteGroup.Use(handler.RBACMiddleware(permisoRepo))
	{
		postulanteGroup.GET("/me", postulanteHdl.ObtenerPerfil)
		postulanteGroup.DELETE("/me", postulanteHdl.SuprimirCuenta)
	}

	postulacionGroup := router.Group("/api/postulacion")
	postulacionGroup.Use(handler.JWTMiddleware(cfg.JWTSecret))
	postulacionGroup.Use(handler.RBACMiddleware(permisoRepo))
	{
		postulacionGroup.POST("/validar-cargo", postulacionHdl.ValidarCargo)
		postulacionGroup.GET("", postulacionHdl.Obtener)
		postulacionGroup.PUT("", postulacionHdl.Guardar)
		postulacionGroup.POST("/documentos", postulacionHdl.SubirDocumento)
		postulacionGroup.POST("/enviar", postulacionHdl.Enviar)
	}
}
