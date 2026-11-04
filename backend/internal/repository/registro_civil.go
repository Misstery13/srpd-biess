package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistroCivilRepository interface {
	BuscarPorCedula(ctx context.Context, cedula string) (*model.RegistroCivil, error)
}

type pgxRegistroCivilRepository struct {
	pool *pgxpool.Pool
}

func NewRegistroCivilRepository(pool *pgxpool.Pool) RegistroCivilRepository {
	return &pgxRegistroCivilRepository{pool: pool}
}

func (r *pgxRegistroCivilRepository) BuscarPorCedula(ctx context.Context, cedula string) (*model.RegistroCivil, error) {
	// Consulta parametrizada (O3)
	query := `
		SELECT cedula, nombres, apellidos, genero, fecha_nacimiento, nacionalidad,
		       estado_civil, conyuge_nombres, conyuge_cedula, fecha_expedicion,
		       codigo_dactilar, es_afiliado, es_jubilado, es_empleador
		FROM registro_civil
		WHERE cedula = $1
	`
	var rc model.RegistroCivil
	err := r.pool.QueryRow(ctx, query, cedula).Scan(
		&rc.Cedula,
		&rc.Nombres,
		&rc.Apellidos,
		&rc.Genero,
		&rc.FechaNacimiento,
		&rc.Nacionalidad,
		&rc.EstadoCivil,
		&rc.ConyugeNombres,
		&rc.ConyugeCedula,
		&rc.FechaExpedicion,
		&rc.CodigoDactilar,
		&rc.EsAfiliado,
		&rc.EsJubilado,
		&rc.EsEmpleador,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al consultar registro civil: %w", err)
	}

	return &rc, nil
}
