package handler

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Servicio simulado: el handler depende de la interfaz service.PostulacionService, no de la implementación
type postulacionSvcSimulado struct{ errGuardar error }

func (postulacionSvcSimulado) ValidarCargo(context.Context, int, *model.ValidarCargoRequest) (*model.ValidarCargoResponse, error) {
	return nil, nil
}
func (postulacionSvcSimulado) ObtenerPostulacion(context.Context, int) (*model.PostulacionCompleta, error) {
	return nil, nil
}
func (m postulacionSvcSimulado) GuardarPostulacion(context.Context, int, *model.PostulacionCompleta) (string, error) {
	return "BIESS-2026-000001", m.errGuardar
}
func (postulacionSvcSimulado) SubirDocumento(context.Context, int, string, *multipart.FileHeader) (string, error) {
	return "", nil
}
func (postulacionSvcSimulado) EnviarPostulacion(context.Context, int, *model.EnviarPostulacionRequest) (string, error) {
	return "", nil
}

func TestPostulacionHandler_ConServicioSimulado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	casos := []struct {
		nombre string
		err    error
		codigo int
	}{
		{"guardado", nil, http.StatusOK},
		{"error de validación del servicio", service.NewErrValidacion("dato inválido"), http.StatusBadRequest},
	}
	for _, c := range casos {
		var svc service.PostulacionService = postulacionSvcSimulado{errGuardar: c.err}
		h := NewPostulacionHandler(svc)
		r := gin.New()
		r.Use(ErrorMiddleware())
		r.PUT("/x", func(ctx *gin.Context) { ctx.Set("postulante_id", "1") }, h.Guardar)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/x", strings.NewReader(`{}`)))
		if w.Code != c.codigo {
			t.Errorf("%s: esperaba %d, obtuvo %d (%s)", c.nombre, c.codigo, w.Code, w.Body.String())
		}
	}
}
