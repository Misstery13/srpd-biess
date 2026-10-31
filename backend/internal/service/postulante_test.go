package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

type mockPostulantePerfilRepo struct {
	perfil      *model.PostulantePerfil
	err         error
	suprimidoID int
}

func (m *mockPostulantePerfilRepo) BuscarActivoPorCedula(ctx context.Context, cedula string) (*model.Postulante, error) {
	return nil, nil
}
func (m *mockPostulantePerfilRepo) BuscarActivoPorEmail(ctx context.Context, email string) (*model.Postulante, error) {
	return nil, nil
}
func (m *mockPostulantePerfilRepo) Crear(ctx context.Context, p *model.Postulante) error {
	return nil
}
func (m *mockPostulantePerfilRepo) ActualizarPassword(ctx context.Context, cedula, passwordHash string) error {
	return nil
}
func (m *mockPostulantePerfilRepo) BuscarRolPorCodigo(ctx context.Context, codigo string) (*model.Rol, error) {
	return nil, nil
}
func (m *mockPostulantePerfilRepo) ObtenerDetalleRegistroCivil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error) {
	return m.perfil, m.err
}
func (m *mockPostulantePerfilRepo) SuprimirPostulante(ctx context.Context, postulanteID int) error {
	m.suprimidoID = postulanteID
	return m.err
}

func TestPostulanteService_ObtenerPerfil_Ecuatoriano(t *testing.T) {
	repo := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 1,
			Cedula:       "1710034065",
			Nombres:      "JUAN CARLOS",
			Apellidos:    "PÉREZ LÓPEZ",
			Nacionalidad: "ECUATORIANA",
		},
	}
	svc := NewPostulanteService(repo, "")

	perfil, err := svc.ObtenerPerfil(context.Background(), 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if perfil.Nacionalidad != "ECUATORIANA" {
		t.Errorf("se esperaba nacionalidad ECUATORIANA, se obtuvo %s", perfil.Nacionalidad)
	}
}

func TestPostulanteService_ObtenerPerfil_NoEcuatoriano_403(t *testing.T) {
	// V5: Nacionalidad != ECUATORIANA debe responder 403 Forbidden
	repo := &mockPostulantePerfilRepo{
		perfil: &model.PostulantePerfil{
			PostulanteID: 2,
			Cedula:       "1754321980",
			Nombres:      "CARLOS ANDRÉS",
			Apellidos:    "GÓMEZ RESTREPO",
			Nacionalidad: "COLOMBIANA",
		},
	}
	svc := NewPostulanteService(repo, "")

	_, err := svc.ObtenerPerfil(context.Background(), 2)
	if err == nil {
		t.Fatal("se esperaba error 403 para ciudadano no ecuatoriano, pero no ocurrió ningún error")
	}

	appErr, ok := err.(ClientError)
	if !ok || appErr.StatusCode() != 403 {
		t.Fatalf("se esperaba código 403, se obtuvo: %v", err)
	}
}

func TestPostulanteService_SuprimirCuenta(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "uploads-test-*")
	if err != nil {
		t.Fatalf("error creando temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	userDir := filepath.Join(tempDir, "1")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatalf("error creando dir de usuario: %v", err)
	}
	testFile := filepath.Join(userDir, "test.pdf")
	_ = os.WriteFile(testFile, []byte("%PDF-dummy"), 0644)

	repo := &mockPostulantePerfilRepo{}
	svc := NewPostulanteService(repo, tempDir)

	if err := svc.SuprimirCuenta(context.Background(), 1); err != nil {
		t.Fatalf("error suprimiendo cuenta: %v", err)
	}

	if repo.suprimidoID != 1 {
		t.Errorf("se esperaba que se suprimiera el postulante 1, se recibió %d", repo.suprimidoID)
	}

	// Verificar borrado físico del directorio (B4, L5)
	if _, err := os.Stat(userDir); !os.IsNotExist(err) {
		t.Error("el directorio de uploads del usuario debería haber sido eliminado físicamente")
	}
}
