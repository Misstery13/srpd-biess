package repository

import (
	"context"
	"fmt"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CatalogoRepository interface {
	ListarPaises(ctx context.Context) ([]model.Pais, error)
	ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error)
	ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error)
}

type pgxCatalogoRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogoRepository(pool *pgxpool.Pool) CatalogoRepository {
	return &pgxCatalogoRepository{pool: pool}
}

func (r *pgxCatalogoRepository) ListarPaises(ctx context.Context) ([]model.Pais, error) {
	query := `SELECT id, nombre FROM paises ORDER BY nombre ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar paises: %w", err)
	}
	defer rows.Close()

	paises := make([]model.Pais, 0)
	for rows.Next() {
		var p model.Pais
		if err := rows.Scan(&p.ID, &p.Nombre); err != nil {
			return nil, fmt.Errorf("error al escanear pais: %w", err)
		}
		paises = append(paises, p)
	}

	return paises, rows.Err()
}

func (r *pgxCatalogoRepository) ListarProvincias(ctx context.Context, paisID int) ([]model.Provincia, error) {
	// Consulta parametrizada (O3)
	query := `SELECT id, pais_id, nombre FROM provincias WHERE pais_id = $1 ORDER BY nombre ASC`
	rows, err := r.pool.Query(ctx, query, paisID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar provincias: %w", err)
	}
	defer rows.Close()

	provincias := make([]model.Provincia, 0)
	for rows.Next() {
		var p model.Provincia
		if err := rows.Scan(&p.ID, &p.PaisID, &p.Nombre); err != nil {
			return nil, fmt.Errorf("error al escanear provincia: %w", err)
		}
		provincias = append(provincias, p)
	}

	return provincias, rows.Err()
}

func (r *pgxCatalogoRepository) ListarCantones(ctx context.Context, provinciaID int) ([]model.Canton, error) {
	// Consulta parametrizada (O3)
	query := `SELECT id, provincia_id, nombre FROM cantones WHERE provincia_id = $1 ORDER BY nombre ASC`
	rows, err := r.pool.Query(ctx, query, provinciaID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar cantones: %w", err)
	}
	defer rows.Close()

	cantones := make([]model.Canton, 0)
	for rows.Next() {
		var c model.Canton
		if err := rows.Scan(&c.ID, &c.ProvinciaID, &c.Nombre); err != nil {
			return nil, fmt.Errorf("error al escanear canton: %w", err)
		}
		cantones = append(cantones, c)
	}

	return cantones, rows.Err()
}
