package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

// PostulanteService gestiona la lógica de perfil y supresión de postulantes.
type PostulanteService struct {
	repo       repository.PostulanteRepository
	uploadsDir string
}

// NewPostulanteService crea una instancia de PostulanteService.
func NewPostulanteService(repo repository.PostulanteRepository, uploadsDir string) *PostulanteService {
	return &PostulanteService{
		repo:       repo,
		uploadsDir: uploadsDir,
	}
}

// ObtenerPerfil obtiene los datos de solo lectura del postulante desde el Registro Civil (P6).
// Si la nacionalidad no es ECUATORIANA, devuelve error prohibido 403 (V5).
func (s *PostulanteService) ObtenerPerfil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error) {
	perfil, err := s.repo.ObtenerDetalleRegistroCivil(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	if perfil == nil {
		return nil, NewErrNoEncontrado("Postulante no encontrado.")
	}

	// V5: Restricción estricta de nacionalidad ecuatoriana
	if perfil.Nacionalidad != "ECUATORIANA" {
		return nil, NewErrProhibido("Solo ciudadanos con nacionalidad ecuatoriana pueden postular.")
	}

	return perfil, nil
}

// SuprimirCuenta aplica la supresión de datos personales según la LOPDP (B4, L5).
// Da de baja lógica al postulante y postulación, anonimiza datos, elimina filas de tablas hijas
// y borra físicamente los archivos PDF subidos al disco.
func (s *PostulanteService) SuprimirCuenta(ctx context.Context, postulanteID int) error {
	// 1. Supresión en base de datos en una sola transacción
	if err := s.repo.SuprimirPostulante(ctx, postulanteID); err != nil {
		return err
	}

	// 2. Borrado físico del directorio de PDFs del usuario (A1, B4, L5)
	if s.uploadsDir != "" {
		dirPostulante := filepath.Join(s.uploadsDir, fmt.Sprintf("%d", postulanteID))
		_ = os.RemoveAll(dirPostulante)
	}

	return nil
}
