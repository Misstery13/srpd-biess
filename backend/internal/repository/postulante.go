package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostulanteRepository interface {
	BuscarActivoPorCedula(ctx context.Context, cedula string) (*model.Postulante, error)
	BuscarActivoPorEmail(ctx context.Context, email string) (*model.Postulante, error)
	Crear(ctx context.Context, p *model.Postulante) error
	ActualizarPassword(ctx context.Context, cedula, passwordHash string) error
	BuscarRolPorCodigo(ctx context.Context, codigo string) (*model.Rol, error)
}

type pgxPostulanteRepository struct {
	pool *pgxpool.Pool
}

func NewPostulanteRepository(pool *pgxpool.Pool) PostulanteRepository {
	return &pgxPostulanteRepository{pool: pool}
}

func (r *pgxPostulanteRepository) BuscarActivoPorCedula(ctx context.Context, cedula string) (*model.Postulante, error) {
	// B2: Solo registros activos (deleted_at IS NULL)
	query := `
		SELECT id, cedula, email, password_hash, rol_id, created_at, deleted_at
		FROM postulantes
		WHERE cedula = $1 AND deleted_at IS NULL
	`
	var p model.Postulante
	err := r.pool.QueryRow(ctx, query, cedula).Scan(
		&p.ID,
		&p.Cedula,
		&p.Email,
		&p.PasswordHash,
		&p.RolID,
		&p.CreatedAt,
		&p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar postulante por cedula: %w", err)
	}

	return &p, nil
}

func (r *pgxPostulanteRepository) BuscarActivoPorEmail(ctx context.Context, email string) (*model.Postulante, error) {
	// V4: Email único sin distinguir mayúsculas entre activos
	query := `
		SELECT id, cedula, email, password_hash, rol_id, created_at, deleted_at
		FROM postulantes
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`
	var p model.Postulante
	err := r.pool.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&p.ID,
		&p.Cedula,
		&p.Email,
		&p.PasswordHash,
		&p.RolID,
		&p.CreatedAt,
		&p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar postulante por email: %w", err)
	}

	return &p, nil
}

func (r *pgxPostulanteRepository) Crear(ctx context.Context, p *model.Postulante) error {
	query := `
		INSERT INTO postulantes (cedula, email, password_hash, rol_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query, p.Cedula, strings.ToLower(strings.TrimSpace(p.Email)), p.PasswordHash, p.RolID).
		Scan(&p.ID, &p.CreatedAt)
}

func (r *pgxPostulanteRepository) ActualizarPassword(ctx context.Context, cedula, passwordHash string) error {
	query := `
		UPDATE postulantes
		SET password_hash = $1
		WHERE cedula = $2 AND deleted_at IS NULL
	`
	cmd, err := r.pool.Exec(ctx, query, passwordHash, cedula)
	if err != nil {
		return fmt.Errorf("error al actualizar contraseña: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("no se encontró ningún postulante activo con esa cédula")
	}
	return nil
}

func (r *pgxPostulanteRepository) BuscarRolPorCodigo(ctx context.Context, codigo string) (*model.Rol, error) {
	query := `
		SELECT id, codigo, nombre, deleted_at
		FROM rol
		WHERE codigo = $1 AND deleted_at IS NULL
	`
	var rol model.Rol
	err := r.pool.QueryRow(ctx, query, codigo).Scan(&rol.ID, &rol.Codigo, &rol.Nombre, &rol.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al buscar rol: %w", err)
	}
	return &rol, nil
}
