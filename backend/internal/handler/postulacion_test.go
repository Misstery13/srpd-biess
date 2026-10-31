package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
	"github.com/Misstery13/srpd-biess/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// Mocks para handler tests
type mockPostulanteRepoForHandler struct {
	perfil *model.PostulantePerfil
}

func (m *mockPostulanteRepoForHandler) BuscarActivoPorCedula(ctx context.Context, cedula string) (*model.Postulante, error) {
	return nil, nil
}
func (m *mockPostulanteRepoForHandler) BuscarActivoPorEmail(ctx context.Context, email string) (*model.Postulante, error) {
	return nil, nil
}
func (m *mockPostulanteRepoForHandler) Crear(ctx context.Context, p *model.Postulante) error {
	return nil
}
func (m *mockPostulanteRepoForHandler) ActualizarPassword(ctx context.Context, cedula, passwordHash string) error {
	return nil
}
func (m *mockPostulanteRepoForHandler) BuscarRolPorCodigo(ctx context.Context, codigo string) (*model.Rol, error) {
	return nil, nil
}
func (m *mockPostulanteRepoForHandler) ObtenerDetalleRegistroCivil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error) {
	return m.perfil, nil
}
func (m *mockPostulanteRepoForHandler) SuprimirPostulante(ctx context.Context, postulanteID int) error {
	return nil
}

type mockPostulacionRepoForHandler struct {
	postulacion *model.PostulacionCompleta
	estado      string
}

func (m *mockPostulacionRepoForHandler) ObtenerPorPostulanteID(ctx context.Context, postulanteID int) (*model.PostulacionCompleta, error) {
	return m.postulacion, nil
}
func (m *mockPostulacionRepoForHandler) GuardarTransaccional(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error) {
	return "BIESS-2026-000001", nil
}
func (m *mockPostulacionRepoForHandler) EnviarPostulacion(ctx context.Context, postulanteID int) (string, error) {
	return "BIESS-2026-000001", nil
}
func (m *mockPostulacionRepoForHandler) ObtenerEstado(ctx context.Context, postulanteID int) (string, error) {
	return m.estado, nil
}

type mockPeriodoRepoForHandler struct{}

func (m *mockPeriodoRepoForHandler) ObtenerPeriodoActivo(ctx context.Context) (*repository.PeriodoRow, error) {
	ahora := time.Now()
	return &repository.PeriodoRow{
		Inicio: ahora.Add(-24 * time.Hour),
		Fin:    ahora.Add(24 * time.Hour),
	}, nil
}

type mockEmailServiceForHandler struct{}

func (m *mockEmailServiceForHandler) EnviarRecuperacion(destinatario, token string) error {
	return nil
}
func (m *mockEmailServiceForHandler) EnviarConfirmacionPostulacion(destinatario, codigo string) error {
	return nil
}

func setupPostulacionTestRouter() (*gin.Engine, *mockPostulanteRepoForHandler, *mockPostulacionRepoForHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorMiddleware())

	postRepo := &mockPostulanteRepoForHandler{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Cedula:       "1710034065",
			Nombres:      "JUAN CARLOS",
			Apellidos:    "PÉREZ LÓPEZ",
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   true,
			Email:        "juan.perez@example.com",
		},
	}
	postulanteSvc := service.NewPostulanteService(postRepo, "")
	postulanteHdl := NewPostulanteHandler(postulanteSvc)

	periodoSvc := service.NewPeriodoService(&mockPeriodoRepoForHandler{})
	emailSvc := &mockEmailServiceForHandler{}
	postulacionRepo := &mockPostulacionRepoForHandler{
		estado: "BORRADOR",
		postulacion: &model.PostulacionCompleta{
			PostulacionHeader: model.PostulacionHeader{
				ID:           1,
				Codigo:       "BIESS-2026-000001",
				PostulanteID: 1,
				Estado:       "BORRADOR",
			},
		},
	}
	tempDir, _ := os.MkdirTemp("", "handler-uploads-*")
	postulacionSvc := service.NewPostulacionService(postulacionRepo, postRepo, periodoSvc, emailSvc, tempDir)
	postulacionHdl := NewPostulacionHandler(postulacionSvc)

	// Inyectar postulante_id en contexto simulando auth
	authMock := func(c *gin.Context) {
		c.Set("postulante_id", "1")
		c.Next()
	}

	postulanteGroup := router.Group("/api/postulante", authMock)
	{
		postulanteGroup.GET("/me", postulanteHdl.ObtenerPerfil)
		postulanteGroup.DELETE("/me", postulanteHdl.SuprimirCuenta)
	}

	postulacionGroup := router.Group("/api/postulacion", authMock)
	{
		postulacionGroup.POST("/validar-cargo", postulacionHdl.ValidarCargo)
		postulacionGroup.GET("", postulacionHdl.Obtener)
		postulacionGroup.PUT("", postulacionHdl.Guardar)
		postulacionGroup.POST("/documentos", postulacionHdl.SubirDocumento)
		postulacionGroup.POST("/enviar", postulacionHdl.Enviar)
	}

	return router, postRepo, postulacionRepo
}

func TestPostulanteHandler_Perfil_200(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/postulante/me", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d", w.Code)
	}

	var perfil model.PostulantePerfil
	_ = json.Unmarshal(w.Body.Bytes(), &perfil)
	if perfil.Nacionalidad != "ECUATORIANA" {
		t.Errorf("nacionalidad incorrecta: %s", perfil.Nacionalidad)
	}
}

func TestPostulanteHandler_Perfil_NoEcuatoriano_403(t *testing.T) {
	router, postRepo, _ := setupPostulacionTestRouter()
	postRepo.perfil.Nacionalidad = "COLOMBIANA"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/postulante/me", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 (V5), obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestPostulanteHandler_SuprimirCuenta_204(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/postulante/me", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("se esperaba 204 (B4), obtenido %d", w.Code)
	}
}

func TestPostulacionHandler_ValidarCargo_200(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()
	body, _ := json.Marshal(map[string]string{"cargo": "AFILIADOS"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/postulacion/validar-cargo", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d: %s", w.Code, w.Body.String())
	}

	var resp model.ValidarCargoResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Habilitado || resp.Mensaje != "Postulante afiliado al IESS" {
		t.Errorf("respuesta inesperada: %+v", resp)
	}
}

func TestPostulacionHandler_Guardar_200(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()
	postulacion := model.PostulacionCompleta{}
	cargo := "AFILIADOS"
	postulacion.Cargo = &cargo
	body, _ := json.Marshal(postulacion)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/postulacion", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestPostulacionHandler_SubirDocumento_201(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	_ = writer.WriteField("tipo", "FORMACION")
	part, _ := writer.CreateFormFile("archivo", "titulo.pdf")
	_, _ = part.Write([]byte("%PDF-1.4 dummy valid PDF"))
	_ = writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/postulacion/documentos", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("se esperaba 201, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestPostulacionHandler_Enviar_200(t *testing.T) {
	router, _, _ := setupPostulacionTestRouter()
	body, _ := json.Marshal(map[string]bool{"acepta_terminos": true})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/postulacion/enviar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d: %s", w.Code, w.Body.String())
	}
}
