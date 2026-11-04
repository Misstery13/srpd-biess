package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PermisoRepository interface {
	TienePermiso(ctx context.Context, rolCodigo, metodo, ruta string) (bool, error)
}

type pgxPermisoRepository struct {
	pool *pgxpool.Pool
}

func NewPermisoRepository(pool *pgxpool.Pool) PermisoRepository {
	return &pgxPermisoRepository{pool: pool}
}

func (r *pgxPermisoRepository) TienePermiso(ctx context.Context, rolCodigo, metodo, ruta string) (bool, error) {
	// R3: Busca permiso activo para (rol, metodo, ruta)
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM permiso p
			JOIN rol r ON p.rol_id = r.id
			WHERE r.codigo = $1
			  AND p.metodo = $2
			  AND p.ruta = $3
			  AND p.deleted_at IS NULL
			  AND r.deleted_at IS NULL
		)
	`
	var tiene bool
	err := r.pool.QueryRow(ctx, query, rolCodigo, metodo, ruta).Scan(&tiene)
	if err != nil {
		return false, fmt.Errorf("error al verificar permisos: %w", err)
	}
	return tiene, nil
}
