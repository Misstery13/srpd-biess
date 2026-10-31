package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware valida el token JWT en el encabezado Authorization: Bearer <token> (S2)
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Debe iniciar sesión para acceder a este recurso"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Su sesión ha expirado o es inválida"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token de autenticación no válido"})
			return
		}

		sub, _ := claims["sub"].(string)
		rol, _ := claims["rol"].(string)

		if sub == "" || rol == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Claims de autenticación incompletos"})
			return
		}

		c.Set("postulante_id", sub)
		c.Set("rol", rol)

		c.Next()
	}
}

// RBACMiddleware valida que el rol tenga permiso sobre el método y ruta actual (R3)
func RBACMiddleware(permisoRepo repository.PermisoRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		rol := c.GetString("rol")
		if rol == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No se encontró el rol de usuario autenticado"})
			return
		}

		metodo := c.Request.Method
		ruta := c.FullPath() // Ruta patrón de Gin (p. ej. /api/postulante/me)

		tiene, err := permisoRepo.TienePermiso(c.Request.Context(), rol, metodo, ruta)
		if err != nil || !tiene {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No tiene permisos para realizar esta acción"})
			return
		}

		c.Next()
	}
}
