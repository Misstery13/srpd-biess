package service

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"os"
	"path/filepath"
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
			Nacionalidad: "ECUATORIANA",
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

func TestPostulacionService_PostulanteHabilitado_Reglas(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)

	// Caso 1: No ecuatoriano (colombiano 1754321980) -> 403 (V5)
	repoCol := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 2,
			Nacionalidad: "COLOMBIANA",
		},
	}
	svcCol := NewPostulacionService(&mockPostulacionRepo{}, repoCol, periodoSvc, &mockEmailService{}, "")
	_, err := svcCol.GuardarPostulacion(context.Background(), 2, &model.PostulacionCompleta{})
	if err == nil || !strings.Contains(err.Error(), "nacionalidad ecuatoriana") {
		t.Errorf("se esperaba error de nacionalidad ecuatoriana (V5), se obtuvo: %v", err)
	}

	// Caso 2: Inhabilitado para el cargo (jubilada postulando a AFILIADOS) -> 403 (1.4.x)
	repoJub := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 3,
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   false,
			EsJubilado:   true,
		},
	}
	cargoAfiliados := "AFILIADOS"
	svcJub := NewPostulacionService(&mockPostulacionRepo{}, repoJub, periodoSvc, &mockEmailService{}, "")
	_, err = svcJub.GuardarPostulacion(context.Background(), 3, &model.PostulacionCompleta{
		PostulacionHeader: model.PostulacionHeader{Cargo: &cargoAfiliados},
	})
	if err == nil || !strings.Contains(err.Error(), "no mantiene la calidad de afiliado") {
		t.Errorf("se esperaba bloqueo de cargo con texto exacto, se obtuvo: %v", err)
	}

	// Caso 3: Postulante suprimido (perfil == nil) -> 401 (E3, L5)
	repoSuprimido := &mockPostulantePerfilRepo{perfil: nil}
	svcSuprimido := NewPostulacionService(&mockPostulacionRepo{}, repoSuprimido, periodoSvc, &mockEmailService{}, "")
	_, err = svcSuprimido.GuardarPostulacion(context.Background(), 99, &model.PostulacionCompleta{})
	if err == nil || !strings.Contains(err.Error(), "Su sesión ya no es válida") {
		t.Errorf("se esperaba error de sesión no válida para cuenta suprimida (E3), se obtuvo: %v", err)
	}
}

func TestPostulacionService_ValidacionListasFijas(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)
	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   true,
		},
	}
	svc := NewPostulacionService(&mockPostulacionRepo{}, repoPost, periodoSvc, &mockEmailService{}, "")

	// 1. Parentesco inválido ("vecino")
	pInvalidoFam := &model.PostulacionCompleta{
		Familiares: []model.FamiliarIESS{
			{Nombres: "Pedro", Parentesco: "vecino", Institucion: "IESS"},
		},
	}
	_, err := svc.GuardarPostulacion(context.Background(), 1, pInvalidoFam)
	if err == nil || !strings.Contains(err.Error(), "parentesco") {
		t.Errorf("se esperaba error por parentesco inválido (1.3.1.2), se obtuvo: %v", err)
	}

	// 2. Nivel de formación inválido ("PRIMARIA")
	pInvalidoNivel := &model.PostulacionCompleta{
		Formacion: []model.FormacionAcademica{
			{Nivel: "PRIMARIA", Institucion: "Escuela", Tipo: "Nacional", Titulo: "Bachiller", Area: "Economía"},
		},
	}
	_, err = svc.GuardarPostulacion(context.Background(), 1, pInvalidoNivel)
	if err == nil || !strings.Contains(err.Error(), "nivel de instrucción") {
		t.Errorf("se esperaba error por nivel inválido (1.5.1), se obtuvo: %v", err)
	}

	// 3. Área inválida ("Astrología")
	pInvalidaArea := &model.PostulacionCompleta{
		Formacion: []model.FormacionAcademica{
			{Nivel: "Tercer Nivel de Grado", Institucion: "U", Tipo: "Nacional", Titulo: "Licenciado", Area: "Astrología"},
		},
	}
	_, err = svc.GuardarPostulacion(context.Background(), 1, pInvalidaArea)
	if err == nil || !strings.Contains(err.Error(), "área de estudios") {
		t.Errorf("se esperaba error por área inválida (1.5.1.4), se obtuvo: %v", err)
	}
}

func TestPostulacionService_ValidacionRutasDocumentos(t *testing.T) {
	ahora := time.Now()
	periodoRepo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	periodoSvc := NewPeriodoService(periodoRepo)
	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   true,
		},
	}
	svc := NewPostulacionService(&mockPostulacionRepo{}, repoPost, periodoSvc, &mockEmailService{}, "")

	respaldoTrue := true
	rutaTraversal := "../../etc/passwd"
	pTraversal := &model.PostulacionCompleta{
		PostulacionHeader: model.PostulacionHeader{
			RespaldoOrg:  &respaldoTrue,
			OrgNombre:    &rutaTraversal,
			OrgDocumento: &rutaTraversal,
		},
	}
	_, err := svc.GuardarPostulacion(context.Background(), 1, pTraversal)
	if err == nil || !strings.Contains(err.Error(), "no es válida o no pertenece al postulante") {
		t.Errorf("se esperaba rechazo de path traversal (O1), se obtuvo: %v", err)
	}

	// Documento de otro usuario (usuario 999 intentando ser referenciado por usuario 1)
	rutaOtroUsuario := "uploads/999/12345678-1234-1234-1234-1234567890ab.pdf"
	pOtroUsuario := &model.PostulacionCompleta{
		PostulacionHeader: model.PostulacionHeader{
			RespaldoOrg:  &respaldoTrue,
			OrgNombre:    &rutaTraversal,
			OrgDocumento: &rutaOtroUsuario,
		},
	}
	_, err = svc.GuardarPostulacion(context.Background(), 1, pOtroUsuario)
	if err == nil || !strings.Contains(err.Error(), "no es válida o no pertenece al postulante") {
		t.Errorf("se esperaba rechazo de documento de otro postulante (O1), se obtuvo: %v", err)
	}
}

func TestPostulacionService_LimpiezaArchivosHuerfanos(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "uploads-huerfanos-*")
	defer os.RemoveAll(tempDir)

	userDir := filepath.Join(tempDir, "1")
	_ = os.MkdirAll(userDir, 0750)

	docViejoPath := filepath.Join(userDir, "00000000-0000-0000-0000-000000000001.pdf")
	docNuevoPath := filepath.Join(userDir, "00000000-0000-0000-0000-000000000002.pdf")
	_ = os.WriteFile(docViejoPath, []byte("viejo"), 0640)
	_ = os.WriteFile(docNuevoPath, []byte("nuevo"), 0640)

	docViejoRel := "uploads/1/00000000-0000-0000-0000-000000000001.pdf"
	docNuevoRel := "uploads/1/00000000-0000-0000-0000-000000000002.pdf"

	oldP := &model.PostulacionCompleta{
		Experiencia: []model.ExperienciaLaboral{
			{Tipo: "GENERAL", Cargo: "Dev", Institucion: "Corp", FechaInicio: time.Now(), Area: "Finanzas", Documento: &docViejoRel},
		},
	}
	newP := &model.PostulacionCompleta{
		Experiencia: []model.ExperienciaLaboral{
			{Tipo: "GENERAL", Cargo: "Dev", Institucion: "Corp", FechaInicio: time.Now(), Area: "Finanzas", Documento: &docNuevoRel},
		},
	}

	ahora := time.Now()
	periodoSvc := NewPeriodoService(&mockPeriodoRepo{
		row: &repository.PeriodoRow{Inicio: ahora.Add(-time.Hour), Fin: ahora.Add(time.Hour)},
	})
	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{PostulanteID: 1, Nacionalidad: "ECUATORIANA", EsAfiliado: true},
	}
	repoPostulacion := &mockPostulacionRepo{postulacion: oldP}
	svc := NewPostulacionService(repoPostulacion, repoPost, periodoSvc, &mockEmailService{}, tempDir)

	_, err := svc.GuardarPostulacion(context.Background(), 1, newP)
	if err != nil {
		t.Fatalf("error guardando postulación: %v", err)
	}

	// Comprobar que docViejoPath fue eliminado para liberar cuota
	if _, err := os.Stat(docViejoPath); !os.IsNotExist(err) {
		t.Errorf("el archivo reemplazado debía ser eliminado de disco para no consumir cuota")
	}
	// Comprobar que docNuevoPath aún existe
	if _, err := os.Stat(docNuevoPath); os.IsNotExist(err) {
		t.Errorf("el archivo nuevo debe preservarse en disco")
	}
}

func TestPostulantePerfil_MinimizacionJSON(t *testing.T) {
	perfil := &model.PostulantePerfil{
		PostulanteID:    1,
		Cedula:          "1710034065",
		Nombres:         "JUAN",
		Apellidos:       "PEREZ",
		FechaExpedicion: time.Now(),
		CodigoDactilar:  "E1234I5678",
	}

	data, err := json.Marshal(perfil)
	if err != nil {
		t.Fatalf("error serializando perfil: %v", err)
	}
	jsonStr := string(data)
	if strings.Contains(jsonStr, "fecha_expedicion") || strings.Contains(jsonStr, "codigo_dactilar") {
		t.Errorf("fecha_expedicion y codigo_dactilar no deben exponerse en json (L2 minimización): %s", jsonStr)
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

	repoPost := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   true,
		},
	}
	svc := NewPostulacionService(&mockPostulacionRepo{estado: "BORRADOR"}, repoPost, periodoSvc, &mockEmailService{}, tempDir)

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
			Nacionalidad: "ECUATORIANA",
			EsAfiliado:   true,
		},
	}
	cargoAfiliados := "AFILIADOS"
	repoPostulacion := &mockPostulacionRepo{
		estado: "BORRADOR",
		postulacion: &model.PostulacionCompleta{
			PostulacionHeader: model.PostulacionHeader{
				ID:     1,
				Codigo: "BIESS-2026-000001",
				Cargo:  &cargoAfiliados,
			},
		},
	}
	svc := NewPostulacionService(repoPostulacion, repoPost, periodoSvc, emailSvc, "")

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
