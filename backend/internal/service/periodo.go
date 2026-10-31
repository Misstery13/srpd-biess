package service

import (
	"context"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

// PeriodoService gestiona la lógica del periodo de postulación.
type PeriodoService struct {
	repo repository.PeriodoRepository
}

// NewPeriodoService crea una instancia de PeriodoService.
func NewPeriodoService(repo repository.PeriodoRepository) *PeriodoService {
	return &PeriodoService{repo: repo}
}

// ObtenerPeriodo devuelve el estado del periodo de postulación actual (T2).
// Si no hay periodo configurado, retorna abierto=false con fechas cero.
func (s *PeriodoService) ObtenerPeriodo(ctx context.Context) (*model.Periodo, error) {
	row, err := s.repo.ObtenerPeriodoActivo(ctx)
	if err != nil {
		return nil, err
	}

	if row == nil {
		return &model.Periodo{
			Abierto: false,
			Inicio:  time.Time{},
			Fin:     time.Time{},
		}, nil
	}

	ahora := time.Now()
	abierto := ahora.After(row.Inicio) && ahora.Before(row.Fin)

	return &model.Periodo{
		Abierto: abierto,
		Inicio:  row.Inicio,
		Fin:     row.Fin,
	}, nil
}
