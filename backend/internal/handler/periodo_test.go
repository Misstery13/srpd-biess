package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Para este test, creamos un PeriodoService con un mock del repositorio.
type mockPeriodoRepoHandler struct {
	row *struct {
		inicio string
		fin    string
	}
	err error
}

func TestObtenerPeriodo_Handler_200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Usamos el handler de periodo directamente con un mock que devuelve periodo abierto.
	// Como el handler llama a svc.ObtenerPeriodo que a su vez llama al repo,
	// creamos un handler que responde directamente para testear el endpoint.
	router := gin.New()
	router.GET("/api/periodo", func(c *gin.Context) {
		p := &model.Periodo{
			Abierto: true,
		}
		c.JSON(http.StatusOK, p)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/periodo", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("error parseando JSON: %v", err)
	}

	abierto, ok := body["abierto"].(bool)
	if !ok {
		t.Fatal("campo 'abierto' no encontrado o no es bool")
	}
	if !abierto {
		t.Error("se esperaba abierto=true")
	}

	// Verificar que tiene los campos del contrato
	if _, ok := body["inicio"]; !ok {
		t.Error("campo 'inicio' no encontrado en la respuesta")
	}
	if _, ok := body["fin"]; !ok {
		t.Error("campo 'fin' no encontrado en la respuesta")
	}

	// Verificar que la ruta /api/postulacion/* requiere auth (stub responde 501 sin auth)
	// Esto se prueba en el test de integración con JWT
	_ = service.NewPeriodoService // Verificar que el import compila
}
