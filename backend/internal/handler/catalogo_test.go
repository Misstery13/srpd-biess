package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// mockCatalogoService simula service.CatalogoService
type mockCatalogoService struct {
	paises     []model.Pais
	provincias []model.Provincia
	cantones   []model.Canton
	err        error
}

func (m *mockCatalogoService) ObtenerPaises(ctx context.Context) ([]model.Pais, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.paises, nil
}

func (m *mockCatalogoService) ObtenerProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.provincias, nil
}

func (m *mockCatalogoService) ObtenerCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cantones, nil
}

func setupTestRouter(svc service.CatalogoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.Use(ErrorMiddleware())

	h := NewCatalogoHandler(svc)
	grp := r.Group("/api/catalogos")
	{
		grp.GET("/paises", h.ListarPaises)
		grp.GET("/paises/:id/provincias", h.ListarProvincias)
		grp.GET("/provincias/:id/cantones", h.ListarCantones)
	}

	return r
}

func TestListarPaises_Exito(t *testing.T) {
	mockSvc := &mockCatalogoService{
		paises: []model.Pais{
			{ID: 52, Nombre: "Ecuador"},
			{ID: 42, Nombre: "Colombia"},
		},
	}
	r := setupTestRouter(mockSvc)

	req, _ := http.NewRequest(http.MethodGet, "/api/catalogos/paises", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba status 200, obtuvo %d", w.Code)
	}

	// Verificar cabeceras de seguridad (O5)
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("esperaba cabecera X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("esperaba cabecera X-Frame-Options: DENY")
	}

	var res []model.Pais
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("error al deserializar respuesta: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("esperaba 2 países, obtuvo %d", len(res))
	}
}

func TestListarProvincias_IDInvalido(t *testing.T) {
	mockSvc := &mockCatalogoService{}
	r := setupTestRouter(mockSvc)

	req, _ := http.NewRequest(http.MethodGet, "/api/catalogos/paises/abc/provincias", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperaba status 400 para ID inválido, obtuvo %d", w.Code)
	}

	var errRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &errRes)
	if errRes["error"] == "" {
		t.Errorf("esperaba mensaje de error en JSON")
	}
}

func TestErrorMiddleware_ErrorTipadoYErrorInterno(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorMiddleware())

	// Ruta que lanza error tipado de bloqueo (429)
	r.GET("/test-bloqueado", func(c *gin.Context) {
		_ = c.Error(service.NewErrBloqueado("registro fallido, inténtelo nuevamente en 10 minutos"))
	})

	// Ruta que lanza error tipado de validación (400)
	r.GET("/test-validacion", func(c *gin.Context) {
		_ = c.Error(service.NewErrValidacion("datos de formulario incompletos"))
	})

	// Ruta que lanza error desconocido (500)
	r.GET("/test-interno", func(c *gin.Context) {
		_ = c.Error(errors.New("conexion sql cerrada inesperadamente"))
	})

	// 1. Probar 429
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/test-bloqueado", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusTooManyRequests {
		t.Errorf("esperaba status 429, obtuvo %d", w1.Code)
	}

	// 2. Probar 400
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test-validacion", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Errorf("esperaba status 400, obtuvo %d", w2.Code)
	}

	// 3. Probar 500 (O5: sin detalles internos expuestos al usuario)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/test-interno", nil)
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusInternalServerError {
		t.Errorf("esperaba status 500, obtuvo %d", w3.Code)
	}

	var res500 map[string]string
	_ = json.Unmarshal(w3.Body.Bytes(), &res500)
	if res500["error"] != "Ocurrió un error interno en el servidor." {
		t.Errorf("error 500 expuso detalles internos: %s", res500["error"])
	}
}
func TestNoRouteYPanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Ocurrió un error interno en el servidor."})
	}))
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada."})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("error inesperado")
	})

	// Probar 404 NoRoute
	w404 := httptest.NewRecorder()
	req404, _ := http.NewRequest(http.MethodGet, "/ruta-inexistente", nil)
	r.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("esperaba status 404, obtuvo %d", w404.Code)
	}
	var res404 map[string]string
	_ = json.Unmarshal(w404.Body.Bytes(), &res404)
	if res404["error"] != "Ruta no encontrada." {
		t.Errorf("respuesta 404 incorrecta: %+v", res404)
	}

	// Probar 500 Panic recovery
	wPanic := httptest.NewRecorder()
	reqPanic, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(wPanic, reqPanic)
	if wPanic.Code != http.StatusInternalServerError {
		t.Errorf("esperaba status 500 tras panic, obtuvo %d", wPanic.Code)
	}
	var resPanic map[string]string
	_ = json.Unmarshal(wPanic.Body.Bytes(), &resPanic)
	if resPanic["error"] != "Ocurrió un error interno en el servidor." {
		t.Errorf("respuesta 500 tras panic incorrecta: %+v", resPanic)
	}
}
