package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// SecurityHeadersMiddleware aplica las cabeceras recomendadas por OWASP (O5)
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}

// ErrorMiddleware traduce errores tipados a respuestas HTTP uniformes (E1, E2, O5)
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// 1. Error de cliente tipado (400, 401, 403, 404, 409, 429)
			var clientErr service.ClientError
			if errors.As(err, &clientErr) {
				c.JSON(clientErr.StatusCode(), gin.H{"error": clientErr.Error()})
				return
			}

			// 2. Violación de restricción única en PostgreSQL (código 23505) -> 409 (E2)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				c.JSON(http.StatusConflict, gin.H{"error": "El registro ya existe o se encuentra duplicado."})
				return
			}

			// 3. Cualquier otro error no controlado -> 500 sin detalles internos en la respuesta (O5)
			log.Printf("[ERROR INTERNO] %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ocurrió un error interno en el servidor."})
		}
	}
}
