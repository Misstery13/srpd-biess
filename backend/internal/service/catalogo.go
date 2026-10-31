package service

import (
	"context"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/Misstery13/srpd-biess/backend/internal/repository"
)

type CatalogoService interface {
	ObtenerPaises(ctx context.Context) ([]model.Pais, error)
	ObtenerProvincias(ctx context.Context, paisID int) ([]model.Provincia, error)
	ObtenerCantones(ctx context.Context, provinciaID int) ([]model.Canton, error)
}

type catalogoService struct {
	repo repository.CatalogoRepository
}

func NewCatalogoService(repo repository.CatalogoRepository) CatalogoService {
	return &catalogoService{repo: repo}
}

func (s *catalogoService) ObtenerPaises(ctx context.Context) ([]model.Pais, error) {
	return s.repo.ListarPaises(ctx)
}

func (s *catalogoService) ObtenerProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	if paisID <= 0 {
		return nil, NewErrValidacion("El identificador de país no es válido")
	}
	return s.repo.ListarProvincias(ctx, paisID)
}

func (s *catalogoService) ObtenerCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	if provinciaID <= 0 {
		return nil, NewErrValidacion("El identificador de provincia no es válido")
	}
	return s.repo.ListarCantones(ctx, provinciaID)
}
