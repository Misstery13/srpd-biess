package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// O4: un texto más largo que la columna es error del cliente (400), no 500
func TestErrorMiddleware_TextoMuyLargo_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorMiddleware())
	r.PUT("/x", func(c *gin.Context) {
		_ = c.Error(fmt.Errorf("error al guardar domicilio: %w", &pgconn.PgError{Code: "22001"}))
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/x", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, obtuvo %d", w.Code)
	}
}

// S2: solo HS256 y con expiración obligatoria
func TestJWTMiddleware_RechazaOtroAlgoritmoYSinExp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secreto := "01234567890123456789012345678901"
	r := gin.New()
	r.GET("/x", JWTMiddleware(secreto), func(c *gin.Context) { c.Status(http.StatusOK) })

	firmar := func(m jwt.SigningMethod, claims jwt.MapClaims) string {
		s, err := jwt.NewWithClaims(m, claims).SignedString([]byte(secreto))
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	exp := time.Now().Add(time.Hour).Unix()
	casos := map[string]struct {
		token  string
		codigo int
	}{
		"válido":  {firmar(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "rol": "POSTULANTE", "exp": exp}), 200},
		"HS512":   {firmar(jwt.SigningMethodHS512, jwt.MapClaims{"sub": "1", "rol": "POSTULANTE", "exp": exp}), 401},
		"sin exp": {firmar(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "rol": "POSTULANTE"}), 401},
	}
	for nombre, c := range casos {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", "Bearer "+c.token)
		r.ServeHTTP(w, req)
		if w.Code != c.codigo {
			t.Errorf("%s: esperaba %d, obtuvo %d", nombre, c.codigo, w.Code)
		}
	}
}
