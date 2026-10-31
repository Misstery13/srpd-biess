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
	ObtenerDetalleRegistroCivil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error)
	SuprimirPostulante(ctx context.Context, postulanteID int) error
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

func (r *pgxPostulanteRepository) ObtenerDetalleRegistroCivil(ctx context.Context, postulanteID int) (*model.PostulantePerfil, error) {
	query := `
		SELECT 
			p.id, p.email, rc.cedula, rc.nombres, rc.apellidos, rc.genero, rc.fecha_nacimiento,
			rc.nacionalidad, rc.estado_civil, rc.conyuge_nombres, rc.conyuge_cedula,
			rc.fecha_expedicion, rc.codigo_dactilar, rc.es_afiliado, rc.es_jubilado, rc.es_empleador
		FROM postulantes p
		JOIN registro_civil rc ON p.cedula = rc.cedula
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`
	var perfil model.PostulantePerfil
	err := r.pool.QueryRow(ctx, query, postulanteID).Scan(
		&perfil.PostulanteID,
		&perfil.Email,
		&perfil.Cedula,
		&perfil.Nombres,
		&perfil.Apellidos,
		&perfil.Genero,
		&perfil.FechaNacimiento,
		&perfil.Nacionalidad,
		&perfil.EstadoCivil,
		&perfil.ConyugeNombres,
		&perfil.ConyugeCedula,
		&perfil.FechaExpedicion,
		&perfil.CodigoDactilar,
		&perfil.EsAfiliado,
		&perfil.EsJubilado,
		&perfil.EsEmpleador,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener detalle de registro civil: %w", err)
	}

	return &perfil, nil
}

func (r *pgxPostulanteRepository) SuprimirPostulante(ctx context.Context, postulanteID int) error {
	// B4: Supresión LOPDP en una sola transacción
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error iniciando transaccion de supresion: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Soft delete y anonimización de datos de contacto del postulante
	cmd, err := tx.Exec(ctx, `
		UPDATE postulantes 
		SET deleted_at = now(), 
		    email = 'eliminado-' || id || '@anonimo.invalid', 
		    password_hash = ''
		WHERE id = $1 AND deleted_at IS NULL
	`, postulanteID)
	if err != nil {
		return fmt.Errorf("error al anonimizar postulante: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("postulante no encontrado o ya eliminado")
	}

	// 2. Buscar si tiene postulación activa
	var postulacionID int
	err = tx.QueryRow(ctx, `
		SELECT id FROM postulaciones 
		WHERE postulante_id = $1 AND deleted_at IS NULL
	`, postulanteID).Scan(&postulacionID)

	if err == nil {
		// Baja lógica de la postulación
		_, err = tx.Exec(ctx, `UPDATE postulaciones SET deleted_at = now() WHERE id = $1`, postulacionID)
		if err != nil {
			return fmt.Errorf("error al dar de baja postulacion: %w", err)
		}

		// Borrado físico de tablas hijas (terceros, L6, B4)
		tablas := []string{
			"domicilios",
			"familiares_iess",
			"formacion_academica",
			"experiencia_laboral",
			"publicaciones",
			"capacitaciones",
			"inhabilidades",
		}
		for _, tabla := range tablas {
			query := fmt.Sprintf("DELETE FROM %s WHERE postulacion_id = $1", tabla)
			if _, err := tx.Exec(ctx, query, postulacionID); err != nil {
				return fmt.Errorf("error al eliminar datos de %s: %w", tabla, err)
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("error al consultar postulacion: %w", err)
	}

	return tx.Commit(ctx)
}
