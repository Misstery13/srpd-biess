package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PeriodoRow representa la fila cruda de la tabla periodo_postulacion.
type PeriodoRow struct {
	Inicio time.Time
	Fin    time.Time
}

// PeriodoRepository define las operaciones de consulta del periodo de postulación.
type PeriodoRepository interface {
	ObtenerPeriodoActivo(ctx context.Context) (*PeriodoRow, error)
}

type periodoRepository struct {
	pool *pgxpool.Pool
}

// NewPeriodoRepository crea una instancia de PeriodoRepository respaldada por PostgreSQL.
func NewPeriodoRepository(pool *pgxpool.Pool) PeriodoRepository {
	return &periodoRepository{pool: pool}
}

// ObtenerPeriodoActivo devuelve el periodo de postulación activo más reciente (no eliminado).
// Si no existe ninguno, retorna nil sin error.
func (r *periodoRepository) ObtenerPeriodoActivo(ctx context.Context) (*PeriodoRow, error) {
	query := `
		SELECT inicio, fin
		FROM periodo_postulacion
		WHERE deleted_at IS NULL
		ORDER BY fin DESC
		LIMIT 1`

	var row PeriodoRow
	err := r.pool.QueryRow(ctx, query).Scan(&row.Inicio, &row.Fin)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
