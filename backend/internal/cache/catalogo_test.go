package cache

import (
	"context"
	"testing"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
)

type mockRepo struct {
	paises     []model.Pais
	provincias []model.Provincia
	cantones   []model.Canton
}

func (m *mockRepo) ListarPaises(ctx context.Context) ([]model.Pais, error) {
	return m.paises, nil
}

func (m *mockRepo) ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	return m.provincias, nil
}

func (m *mockRepo) ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	return m.cantones, nil
}

func TestCachedCatalogoRepository_FallbackSinRedis(t *testing.T) {
	repo := &mockRepo{
		paises:     []model.Pais{{ID: 52, Nombre: "Ecuador"}},
		provincias: []model.Provincia{{ID: 19, PaisID: 52, Nombre: "Pichincha"}},
		cantones:   []model.Canton{{ID: 170, ProvinciaID: 19, Nombre: "Quito"}},
	}

	// Con redis.Client nil debe consultar directamente al repositorio de PostgreSQL sin errores
	cachedRepo := NewCachedCatalogoRepository(repo, nil)

	paises, err := cachedRepo.ListarPaises(context.Background())
	if err != nil || len(paises) != 1 {
		t.Fatalf("error en ListarPaises fallback: %v", err)
	}

	provincias, err := cachedRepo.ListarProvincias(context.Background(), 52)
	if err != nil || len(provincias) != 1 {
		t.Fatalf("error en ListarProvincias fallback: %v", err)
	}

	cantones, err := cachedRepo.ListarCantones(context.Background(), 19)
	if err != nil || len(cantones) != 1 {
		t.Fatalf("error en ListarCantones fallback: %v", err)
	}
}
