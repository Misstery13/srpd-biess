package service

import (
	"context"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

// mockPeriodoRepo implementa PeriodoRepository para testing.
type mockPeriodoRepo struct {
	row *repository.PeriodoRow
	err error
}

func (m *mockPeriodoRepo) ObtenerPeriodoActivo(_ context.Context) (*repository.PeriodoRow, error) {
	return m.row, m.err
}

func TestObtenerPeriodo_Abierto(t *testing.T) {
	ahora := time.Now()
	repo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-24 * time.Hour),
			Fin:    ahora.Add(24 * time.Hour),
		},
	}
	svc := NewPeriodoService(repo)

	p, err := svc.ObtenerPeriodo(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !p.Abierto {
		t.Error("se esperaba abierto=true, pero fue false")
	}
}

func TestObtenerPeriodo_Cerrado_Futuro(t *testing.T) {
	ahora := time.Now()
	repo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(24 * time.Hour),
			Fin:    ahora.Add(48 * time.Hour),
		},
	}
	svc := NewPeriodoService(repo)

	p, err := svc.ObtenerPeriodo(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Abierto {
		t.Error("se esperaba abierto=false para periodo futuro, pero fue true")
	}
}

func TestObtenerPeriodo_Cerrado_Pasado(t *testing.T) {
	ahora := time.Now()
	repo := &mockPeriodoRepo{
		row: &repository.PeriodoRow{
			Inicio: ahora.Add(-48 * time.Hour),
			Fin:    ahora.Add(-24 * time.Hour),
		},
	}
	svc := NewPeriodoService(repo)

	p, err := svc.ObtenerPeriodo(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Abierto {
		t.Error("se esperaba abierto=false para periodo pasado, pero fue true")
	}
}

func TestObtenerPeriodo_SinPeriodo(t *testing.T) {
	repo := &mockPeriodoRepo{row: nil}
	svc := NewPeriodoService(repo)

	p, err := svc.ObtenerPeriodo(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Abierto {
		t.Error("se esperaba abierto=false sin periodo, pero fue true")
	}
	if !p.Inicio.IsZero() || !p.Fin.IsZero() {
		t.Error("se esperaban fechas cero sin periodo configurado")
	}
}
