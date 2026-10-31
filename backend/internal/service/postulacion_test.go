package service

import (
	"bytes"
	"context"
	"mime/multipart"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

type mockPostulacionRepo struct {
	postulacion *model.PostulacionCompleta
	estado      string
	codigo      string
	errGuardar  error
	errEnviar   error
}

func (m *mockPostulacionRepo) ObtenerPorPostulanteID(ctx context.Context, postulanteID int) (*model.PostulacionCompleta, error) {
	return m.postulacion, nil
}
func (m *mockPostulacionRepo) GuardarTransaccional(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error) {
	if m.errGuardar != nil {
		return "", m.errGuardar
	}
	return "BIESS-2026-000001", nil
}
func (m *mockPostulacionRepo) EnviarPostulacion(ctx context.Context, postulanteID int) (string, error) {
	if m.errEnviar != nil {
		return "", m.errEnviar
	}
	return "BIESS-2026-000001", nil
}
func (m *mockPostulacionRepo) ObtenerEstado(ctx context.Context, postulanteID int) (string, error) {
	return m.estado, nil
}

func TestPostulacionService_ValidarCargo_MensajesExactos(t *testing.T) {
	// Postulante solo afiliado
	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			EsAfiliado:   true,
			EsJubilado:   false,
			EsEmpleador:  false,
		},
	}
	svc := NewPostulacionService(&mockPostulacionRepo{}, repoPost, nil, &mockEmailService{}, "")

	// 1. AFILIADOS (Habilitado 1.4.1.1)
	resp, err := svc.ValidarCargo(context.Background(), 1, &model.ValidarCargoRequest{Cargo: "AFILIADOS"})
	if err != nil || !resp.Habilitado || resp.Mensaje != "Postulante afiliado al IESS" {
		t.Fatalf("fallo en validar AFILIADOS habilitado: resp=%+v err=%v", resp, err)
	}

	// 2. JUBILADOS (No habilitado 1.4.2.2 - mensaje exacto)
	respJub, err := svc.ValidarCargo(context.Background(), 1, &model.ValidarCargoRequest{Cargo: "JUBILADOS"})
	if err != nil || respJub.Habilitado {
		t.Fatalf("se esperaba deshabilitado para JUBILADOS: %+v", respJub)
	}
	msgJubEsperado := "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación"
	if respJub.Mensaje != msgJubEsperado {
		t.Errorf("mensaje de jubilado incorrecto.\nEsperado: %s\nObtenido: %s", msgJubEsperado, respJub.Mensaje)
	}

	// 3. EMPLEADORES (No habilitado 1.4.3.2 - mensaje exacto)
	respEmp, err := svc.ValidarCargo(context.Background(), 1, &model.ValidarCargoRequest{Cargo: "EMPLEADORES"})
	if err != nil || respEmp.Habilitado {
		t.Fatalf("se esperaba deshabilitado para EMPLEADORES: %+v", respEmp)
	}
	msgEmpEsperado := "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador"
	if respEmp.Mensaje != msgEmpEsperado {
		t.Errorf("mensaje de empleador incorrecto.\nEsperado: %s\nObtenido: %s", msgEmpEsperado, respEmp.Mensaje)
	}
}

func TestPostulacionService_CalcularDuracion(t *testing.T) {
	inicio := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	fin := time.Date(2022, 3, 10, 0, 0, 0, 0, time.UTC)

	dur := CalcularDuracion(inicio, &fin)
	if !strings.Contains(dur, "2 años") || !strings.Contains(dur, "2 meses") || !strings.Contains(dur, "9 días") {
		t.Errorf("cálculo de duración inesperado: %s", dur)
	}
}

func TestPostulacionService_GuardarPostulacion_Validaciones(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)
	svc := NewPostulacionService(&mockPostulacionRepo{}, &mockPostulantePerfilRepo{}, periodoSvc, &mockEmailService{}, "")

	// 1. Error V8: funciones > 1000 caracteres
	strLarga := strings.Repeat("A", 1001)
	pInvalida := &model.PostulacionCompleta{
		Experiencia: []model.ExperienciaLaboral{
			{
				Tipo:        "GENERAL",
				Cargo:       "Analista",
				Institucion: "Banco",
				FechaInicio: ahora.Add(-30 * 24 * time.Hour),
				Funciones:   &strLarga,
			},
		},
	}
	_, err := svc.GuardarPostulacion(context.Background(), 1, pInvalida)
	if err == nil {
		t.Error("se esperaba error de validación para funciones > 1000 caracteres (V8)")
	}

	// 2. Error V7: fecha fin < fecha inicio
	fechaInicio := ahora
	fechaFin := ahora.Add(-24 * time.Hour)
	pFechaInvalida := &model.PostulacionCompleta{
		Experiencia: []model.ExperienciaLaboral{
			{
				Tipo:        "GENERAL",
				Cargo:       "Analista",
				Institucion: "Banco",
				FechaInicio: fechaInicio,
				FechaFin:    &fechaFin,
			},
		},
	}
	_, err = svc.GuardarPostulacion(context.Background(), 1, pFechaInvalida)
	if err == nil {
		t.Error("se esperaba error de validación para fecha_fin < fecha_inicio (V7)")
	}
}

func TestPostulacionService_SubirDocumento_ValidacionPDF(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)
	tempDir, _ := os.MkdirTemp("", "uploads-subir-*")
	defer os.RemoveAll(tempDir)

	svc := NewPostulacionService(&mockPostulacionRepo{estado: "BORRADOR"}, &mockPostulantePerfilRepo{}, periodoSvc, &mockEmailService{}, tempDir)

	// Crear multipart dummy no PDF
	crearHeader := func(contenido []byte) *multipart.FileHeader {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		part, _ := w.CreateFormFile("archivo", "test.pdf")
		_, _ = part.Write(contenido)
		_ = w.Close()

		reader := multipart.NewReader(&b, w.Boundary())
		form, _ := reader.ReadForm(1024 * 1024)
		return form.File["archivo"][0]
	}

	// 1. Archivo que no empieza con %PDF (A2)
	headerInvalido := crearHeader([]byte("ESTO NO ES UN PDF SINO TEXTO PLANO"))
	_, err := svc.SubirDocumento(context.Background(), 1, "FORMACION", headerInvalido)
	if err == nil || !strings.Contains(err.Error(), "no es un documento PDF válido") {
		t.Errorf("se esperaba error por magic bytes inválidos (A2), se obtuvo: %v", err)
	}

	// 2. Archivo PDF válido (empieza con %PDF) (A1)
	headerValido := crearHeader([]byte("%PDF-1.4 dummy content"))
	ruta, err := svc.SubirDocumento(context.Background(), 1, "FORMACION", headerValido)
	if err != nil {
		t.Fatalf("error subiendo archivo válido: %v", err)
	}
	if !strings.HasPrefix(ruta, "uploads/1/") || !strings.HasSuffix(ruta, ".pdf") {
		t.Errorf("ruta devuelta inesperada: %s", ruta)
	}
}

func TestPostulacionService_EnviarPostulacion_ExigeConsentimiento(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)
	emailSvc := &mockEmailService{}
	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Email:        "juan.perez@example.com",
		},
	}
	svc := NewPostulacionService(&mockPostulacionRepo{estado: "BORRADOR"}, repoPost, periodoSvc, emailSvc, "")

	// 1. Sin consentimiento L1 -> debe fallar
	_, err := svc.EnviarPostulacion(context.Background(), 1, &model.EnviarPostulacionRequest{AceptaTerminos: false})
	if err == nil || !strings.Contains(err.Error(), "consentimiento") {
		t.Errorf("se esperaba error por no aceptar términos (L1), se obtuvo: %v", err)
	}

	// 2. Con consentimiento -> éxito y correo enviado a Mailpit (I2)
	codigo, err := svc.EnviarPostulacion(context.Background(), 1, &model.EnviarPostulacionRequest{AceptaTerminos: true})
	if err != nil {
		t.Fatalf("error inesperado enviando postulación: %v", err)
	}
	if codigo != "BIESS-2026-000001" {
		t.Errorf("código inesperado: %s", codigo)
	}
	if len(emailSvc.enviados) != 1 || emailSvc.enviados[0] != "juan.perez@example.com" {
		t.Errorf("se esperaba envío de correo de confirmación a Mailpit (I2), enviados: %v", emailSvc.enviados)
	}
}
