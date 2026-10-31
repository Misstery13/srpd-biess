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

	// 3. Conexión a Redis con timeouts cortos y sin reintentos (Bloqueante 2)
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		DialTimeout:  200 * time.Millisecond,
		ReadTimeout:  200 * time.Millisecond,
		WriteTimeout: 200 * time.Millisecond,
		MaxRetries:   0,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[ALERTA] No se pudo conectar a Redis (%v). Se usará fallback directo a PostgreSQL.", err)
	} else {
		log.Println("Conexión exitosa a Redis")
	}

	// 4. Inyección de dependencias (Repository -> Service -> Handler)
	// Catálogos
	catalogoRepo := repository.NewCatalogoRepository(pgPool)
	cachedCatalogoRepo := cache.NewCachedCatalogoRepository(catalogoRepo, redisClient)
	catalogoSvc := service.NewCatalogoService(cachedCatalogoRepo)
	catalogoHdl := handler.NewCatalogoHandler(catalogoSvc)

	// Autenticación (EF-11, Issue #8)
	rcRepo := repository.NewRegistroCivilRepository(pgPool)
	postRepo := repository.NewPostulanteRepository(pgPool)
	permisoRepo := repository.NewPermisoRepository(pgPool)
	authCache := cache.NewAuthCache(redisClient)
	emailSvc := service.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.AppURL)
	authSvc := service.NewAuthService(rcRepo, postRepo, authCache, emailSvc, cfg.JWTSecret)
	authHdl := handler.NewAuthHandler(authSvc)

	// Periodo de postulación (T2)
	periodoRepo := repository.NewPeriodoRepository(pgPool)
	periodoSvc := service.NewPeriodoService(periodoRepo)
	periodoHdl := handler.NewPeriodoHandler(periodoSvc)

	// Postulante y Postulación (EF-12, Issue #11)
	postulanteSvc := service.NewPostulanteService(postRepo, cfg.UploadsDir)
	postulanteHdl := handler.NewPostulanteHandler(postulanteSvc)

	postulacionRepo := repository.NewPostulacionRepository(pgPool)
	postulacionSvc := service.NewPostulacionService(postulacionRepo, postRepo, periodoSvc, emailSvc, cfg.UploadsDir)
	postulacionHdl := handler.NewPostulacionHandler(postulacionSvc)

	// 5. Configuración de Gin (Menor 1 y Menor 2)
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

	// Ruta inexistente uniforme JSON (E1)
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada."})
	})

	// Ruta de health check (mantenida para despliegue y monitoreo)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Endpoints de catálogos (Públicos)
	catalogosGroup := router.Group("/api/catalogos")
	{
		catalogosGroup.GET("/paises", catalogoHdl.ListarPaises)
		catalogosGroup.GET("/paises/:id/provincias", catalogoHdl.ListarProvincias)
		catalogosGroup.GET("/provincias/:id/cantones", catalogoHdl.ListarCantones)
	}

	// Endpoints de autenticación (Públicos - EF-11)
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/verificar", authHdl.Verificar)
		authGroup.POST("/registro", authHdl.Registrar)
		authGroup.POST("/login", authHdl.Login)
		authGroup.POST("/recuperar", authHdl.Recuperar)
		authGroup.POST("/restablecer", authHdl.Restablecer)
	}

	// Periodo de postulación (Público - T2)
	router.GET("/api/periodo", periodoHdl.ObtenerPeriodo)

	// Rutas protegidas de postulante (Requieren JWT y validación RBAC - R3, R4)
	postulanteGroup := router.Group("/api/postulante")
	postulanteGroup.Use(handler.JWTMiddleware(cfg.JWTSecret))
	postulanteGroup.Use(handler.RBACMiddleware(permisoRepo))
	{
		postulanteGroup.GET("/me", postulanteHdl.ObtenerPerfil)
		postulanteGroup.DELETE("/me", postulanteHdl.SuprimirCuenta)
	}

	// Rutas protegidas de postulación (Requieren JWT y validación RBAC - R3, R4)
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

	log.Printf("Servidor API escuchando en el puerto :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
