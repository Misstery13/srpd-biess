package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

type mockRepo struct {
	paises     []model.Pais
	provincias []model.Provincia
	cantones   []model.Canton
	err        error
}

func (m *mockRepo) ListarPaises(ctx context.Context) ([]model.Pais, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.paises, nil
}

func (m *mockRepo) ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.provincias, nil
}

func (m *mockRepo) ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cantones, nil
}

func TestCatalogoService_ObtenerPaises(t *testing.T) {
	repo := &mockRepo{
		paises: []model.Pais{{ID: 52, Nombre: "Ecuador"}},
	}
	svc := NewCatalogoService(repo)

	paises, err := svc.ObtenerPaises(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(paises) != 1 || paises[0].Nombre != "Ecuador" {
		t.Fatalf("datos incorrectos: %+v", paises)
	}
}

func TestCatalogoService_Validaciones(t *testing.T) {
	repo := &mockRepo{}
	svc := NewCatalogoService(repo)

	// País ID <= 0
	_, err := svc.ObtenerProvincias(context.Background(), 0)
	if err == nil {
		t.Fatal("esperaba error para paisID = 0")
	}
	var clientErr ClientError
	if !errors.As(err, &clientErr) || clientErr.StatusCode() != 400 {
		t.Fatalf("esperaba error de cliente 400, obtuvo %v", err)
	}

	// Provincia ID <= 0
	_, err = svc.ObtenerCantones(context.Background(), -5)
	if err == nil {
		t.Fatal("esperaba error para provinciaID = -5")
	}
	if !errors.As(err, &clientErr) || clientErr.StatusCode() != 400 {
		t.Fatalf("esperaba error de cliente 400, obtuvo %v", err)
	}
}
