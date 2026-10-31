package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostulacionRepository define las operaciones de persistencia de la postulación (EF-12).
type PostulacionRepository interface {
	ObtenerPorPostulanteID(ctx context.Context, postulanteID int) (*model.PostulacionCompleta, error)
	GuardarTransaccional(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error)
	EnviarPostulacion(ctx context.Context, postulanteID int) (string, error)
	ObtenerEstado(ctx context.Context, postulanteID int) (string, error)
}

type pgxPostulacionRepository struct {
	pool *pgxpool.Pool
}

// NewPostulacionRepository crea una nueva instancia respaldada por PostgreSQL.
func NewPostulacionRepository(pool *pgxpool.Pool) PostulacionRepository {
	return &pgxPostulacionRepository{pool: pool}
}

func (r *pgxPostulacionRepository) ObtenerEstado(ctx context.Context, postulanteID int) (string, error) {
	query := `SELECT estado FROM postulaciones WHERE postulante_id = $1 AND deleted_at IS NULL`
	var estado string
	err := r.pool.QueryRow(ctx, query, postulanteID).Scan(&estado)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return estado, nil
}

func (r *pgxPostulacionRepository) ObtenerPorPostulanteID(ctx context.Context, postulanteID int) (*model.PostulacionCompleta, error) {
	// B2: Filtrar deleted_at IS NULL
	queryHeader := `
		SELECT id, codigo, postulante_id, cargo, respaldo_org, org_nombre, org_documento,
		       estado, acepta_terminos, fecha_consentimiento, created_at, updated_at
		FROM postulaciones
		WHERE postulante_id = $1 AND deleted_at IS NULL
	`
	var res model.PostulacionCompleta
	err := r.pool.QueryRow(ctx, queryHeader, postulanteID).Scan(
		&res.ID,
		&res.Codigo,
		&res.PostulanteID,
		&res.Cargo,
		&res.RespaldoOrg,
		&res.OrgNombre,
		&res.OrgDocumento,
		&res.Estado,
		&res.AceptaTerminos,
		&res.FechaConsentimiento,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener postulacion: %w", err)
	}

	postulacionID := res.ID

	// 1. Domicilio
	queryDom := `
		SELECT postulacion_id, pais_id, provincia_id, canton_id, provincia_texto, ciudad_texto,
		       celular, email2, calle_principal, numero, calle_secundaria, sector, referencia, telefono
		FROM domicilios
		WHERE postulacion_id = $1
	`
	var dom model.Domicilio
	err = r.pool.QueryRow(ctx, queryDom, postulacionID).Scan(
		&dom.PostulacionID,
		&dom.PaisID,
		&dom.ProvinciaID,
		&dom.CantonID,
		&dom.ProvinciaTexto,
		&dom.CiudadTexto,
		&dom.Celular,
		&dom.Email2,
		&dom.CallePrincipal,
		&dom.Numero,
		&dom.CalleSecundaria,
		&dom.Sector,
		&dom.Referencia,
		&dom.Telefono,
	)
	if err == nil {
		res.Domicilio = &dom
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error al obtener domicilio: %w", err)
	}

	// 2. Familiares IESS
	queryFam := `
		SELECT id, postulacion_id, nombres, parentesco, institucion, area, cargo
		FROM familiares_iess
		WHERE postulacion_id = $1
		ORDER BY id ASC
	`
	rowsFam, err := r.pool.Query(ctx, queryFam, postulacionID)
	if err == nil {
		defer rowsFam.Close()
		for rowsFam.Next() {
			var f model.FamiliarIESS
			if err := rowsFam.Scan(&f.ID, &f.PostulacionID, &f.Nombres, &f.Parentesco, &f.Institucion, &f.Area, &f.Cargo); err == nil {
				res.Familiares = append(res.Familiares, f)
			}
		}
	}

	// 3. Formación Académica
	queryForm := `
		SELECT id, postulacion_id, nivel, institucion, tipo, titulo, area, senescyt_numero, senescyt_fecha, documento
		FROM formacion_academica
		WHERE postulacion_id = $1
		ORDER BY id ASC
	`
	rowsForm, err := r.pool.Query(ctx, queryForm, postulacionID)
	if err == nil {
		defer rowsForm.Close()
		for rowsForm.Next() {
			var fa model.FormacionAcademica
			if err := rowsForm.Scan(&fa.ID, &fa.PostulacionID, &fa.Nivel, &fa.Institucion, &fa.Tipo, &fa.Titulo, &fa.Area, &fa.SenescytNumero, &fa.SenescytFecha, &fa.Documento); err == nil {
				res.Formacion = append(res.Formacion, fa)
			}
		}
	}

	// 4. Experiencia Laboral
	queryExp := `
		SELECT id, postulacion_id, tipo, cargo, institucion, fecha_inicio, fecha_fin, area, funciones, documento
		FROM experiencia_laboral
		WHERE postulacion_id = $1
		ORDER BY fecha_inicio DESC
	`
	rowsExp, err := r.pool.Query(ctx, queryExp, postulacionID)
	if err == nil {
		defer rowsExp.Close()
		for rowsExp.Next() {
			var el model.ExperienciaLaboral
			if err := rowsExp.Scan(&el.ID, &el.PostulacionID, &el.Tipo, &el.Cargo, &el.Institucion, &el.FechaInicio, &el.FechaFin, &el.Area, &el.Funciones, &el.Documento); err == nil {
				res.Experiencia = append(res.Experiencia, el)
			}
		}
	}

	// 5. Publicaciones
	queryPub := `
		SELECT id, postulacion_id, titulo, tipo, medio, fecha, relacion_profesional, documento
		FROM publicaciones
		WHERE postulacion_id = $1
		ORDER BY fecha DESC
	`
	rowsPub, err := r.pool.Query(ctx, queryPub, postulacionID)
	if err == nil {
		defer rowsPub.Close()
		for rowsPub.Next() {
			var pb model.Publicacion
			if err := rowsPub.Scan(&pb.ID, &pb.PostulacionID, &pb.Titulo, &pb.Tipo, &pb.Medio, &pb.Fecha, &pb.RelacionProfesional, &pb.Documento); err == nil {
				res.Publicaciones = append(res.Publicaciones, pb)
			}
		}
	}

	// 6. Capacitaciones
	queryCap := `
		SELECT id, postulacion_id, tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin, horas, documento
		FROM capacitaciones
		WHERE postulacion_id = $1
		ORDER BY fecha_inicio DESC
	`
	rowsCap, err := r.pool.Query(ctx, queryCap, postulacionID)
	if err == nil {
		defer rowsCap.Close()
		for rowsCap.Next() {
			var cp model.Capacitacion
			if err := rowsCap.Scan(&cp.ID, &cp.PostulacionID, &cp.TipoEvento, &cp.NombreEvento, &cp.Institucion, &cp.FechaInicio, &cp.FechaFin, &cp.Horas, &cp.Documento); err == nil {
				res.Capacitaciones = append(res.Capacitaciones, cp)
			}
		}
	}

	// 7. Inhabilidades
	queryInh := `
		SELECT postulacion_id, goce_derechos_politicos, inhabilitado_comercio, mora_obligaciones_estado,
		       vinculo_sistema_financiero, funcionario_iess_biess, interes_aseguradoras,
		       removido_control, removido_institucion, removido_cargo, removido_fecha_fin, removido_motivo,
		       obligaciones_sri, castigo_obligaciones_financieras, litigio_iess_biess,
		       procesado_corrupcion, responsabilidad_contraloria, registrado_uafe
		FROM inhabilidades
		WHERE postulacion_id = $1
	`
	var inh model.Inhabilidades
	err = r.pool.QueryRow(ctx, queryInh, postulacionID).Scan(
		&inh.PostulacionID,
		&inh.GoceDerechosPoliticos,
		&inh.InhabilitadoComercio,
		&inh.MoraObligacionesEstado,
		&inh.VinculoSistemaFinanciero,
		&inh.FuncionarioIessBiess,
		&inh.InteresAseguradoras,
		&inh.RemovidoControl,
		&inh.RemovidoInstitucion,
		&inh.RemovidoCargo,
		&inh.RemovidoFechaFin,
		&inh.RemovidoMotivo,
		&inh.ObligacionesSRI,
		&inh.CastigoObligacionesFinancieras,
		&inh.LitigioIessBiess,
		&inh.ProcesadoCorrupcion,
		&inh.ResponsabilidadContraloria,
		&inh.RegistradoUAFE,
	)
	if err == nil {
		res.Inhabilidades = &inh
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error al obtener inhabilidades: %w", err)
	}

	return &res, nil
}

func (r *pgxPostulacionRepository) GuardarTransaccional(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error) {
	// P4: Una sola transacción. Si algo falla -> ROLLBACK completo
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error al iniciar transaccion: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verificar si existe postulación activa
	var postulacionID int
	var codigo string
	var estado string

	queryBuscar := `SELECT id, codigo, estado FROM postulaciones WHERE postulante_id = $1 AND deleted_at IS NULL`
	err = tx.QueryRow(ctx, queryBuscar, postulanteID).Scan(&postulacionID, &codigo, &estado)

	if errors.Is(err, pgx.ErrNoRows) {
		// P2: Código con formato BIESS-2026-000001
		insertHeader := `
			WITH next_seq AS (
				SELECT nextval('public.postulaciones_id_seq') AS seq_id
			)
			INSERT INTO postulaciones (id, codigo, postulante_id, cargo, respaldo_org, org_nombre, org_documento, estado)
			SELECT 
				seq_id,
				'BIESS-2026-' || LPAD(seq_id::text, 6, '0'),
				$1, $2, $3, $4, $5, 'BORRADOR'
			FROM next_seq
			RETURNING id, codigo
		`
		err = tx.QueryRow(ctx, insertHeader, postulanteID, p.Cargo, p.RespaldoOrg, p.OrgNombre, p.OrgDocumento).
			Scan(&postulacionID, &codigo)
		if err != nil {
			return "", fmt.Errorf("error al crear cabecera de postulacion: %w", err)
		}
	} else if err == nil {
		// P3: Estado ENVIADA no editable
		if estado == "ENVIADA" || estado == "CERRADA" {
			return "", errors.New("POSTULACION_ENVIADA_NO_EDITABLE")
		}

		updateHeader := `
			UPDATE postulaciones
			SET cargo = $2, respaldo_org = $3, org_nombre = $4, org_documento = $5, updated_at = now()
			WHERE id = $1 AND deleted_at IS NULL
			RETURNING codigo
		`
		err = tx.QueryRow(ctx, updateHeader, postulacionID, p.Cargo, p.RespaldoOrg, p.OrgNombre, p.OrgDocumento).
			Scan(&codigo)
		if err != nil {
			return "", fmt.Errorf("error al actualizar cabecera de postulacion: %w", err)
		}

		// B3: Eliminar filas hijas anteriores antes de reinsertar
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
			if _, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE postulacion_id = $1", tabla), postulacionID); err != nil {
				return "", fmt.Errorf("error al limpiar tabla %s: %w", tabla, err)
			}
		}
	} else {
		return "", fmt.Errorf("error al verificar postulacion: %w", err)
	}

	// Inserción de tablas hijas
	// 1. Domicilio
	if p.Domicilio != nil {
		insertDom := `
			INSERT INTO domicilios (
				postulacion_id, pais_id, provincia_id, canton_id, provincia_texto, ciudad_texto,
				celular, email2, calle_principal, numero, calle_secundaria, sector, referencia, telefono
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`
		_, err = tx.Exec(ctx, insertDom,
			postulacionID, p.Domicilio.PaisID, p.Domicilio.ProvinciaID, p.Domicilio.CantonID,
			p.Domicilio.ProvinciaTexto, p.Domicilio.CiudadTexto, p.Domicilio.Celular, p.Domicilio.Email2,
			p.Domicilio.CallePrincipal, p.Domicilio.Numero, p.Domicilio.CalleSecundaria, p.Domicilio.Sector,
			p.Domicilio.Referencia, p.Domicilio.Telefono,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar domicilio: %w", err)
		}
	}

	// 2. Familiares IESS
	for _, f := range p.Familiares {
		insertFam := `
			INSERT INTO familiares_iess (postulacion_id, nombres, parentesco, institucion, area, cargo)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, insertFam, postulacionID, f.Nombres, f.Parentesco, f.Institucion, f.Area, f.Cargo)
		if err != nil {
			return "", fmt.Errorf("error al guardar familiar IESS: %w", err)
		}
	}

	// 3. Formación Académica
	for _, fa := range p.Formacion {
		insertForm := `
			INSERT INTO formacion_academica (
				postulacion_id, nivel, institucion, tipo, titulo, area, senescyt_numero, senescyt_fecha, documento
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, insertForm,
			postulacionID, fa.Nivel, fa.Institucion, fa.Tipo, fa.Titulo, fa.Area,
			fa.SenescytNumero, fa.SenescytFecha, fa.Documento,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar formacion academica: %w", err)
		}
	}

	// 4. Experiencia Laboral
	for _, el := range p.Experiencia {
		insertExp := `
			INSERT INTO experiencia_laboral (
				postulacion_id, tipo, cargo, institucion, fecha_inicio, fecha_fin, area, funciones, documento
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(ctx, insertExp,
			postulacionID, el.Tipo, el.Cargo, el.Institucion, el.FechaInicio, el.FechaFin,
			el.Area, el.Funciones, el.Documento,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar experiencia laboral: %w", err)
		}
	}

	// 5. Publicaciones
	for _, pb := range p.Publicaciones {
		insertPub := `
			INSERT INTO publicaciones (
				postulacion_id, titulo, tipo, medio, fecha, relacion_profesional, documento
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = tx.Exec(ctx, insertPub,
			postulacionID, pb.Titulo, pb.Tipo, pb.Medio, pb.Fecha, pb.RelacionProfesional, pb.Documento,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar publicacion: %w", err)
		}
	}

	// 6. Capacitaciones
	for _, cp := range p.Capacitaciones {
		insertCap := `
			INSERT INTO capacitaciones (
				postulacion_id, tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin, horas, documento
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.Exec(ctx, insertCap,
			postulacionID, cp.TipoEvento, cp.NombreEvento, cp.Institucion, cp.FechaInicio, cp.FechaFin, cp.Horas, cp.Documento,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar capacitacion: %w", err)
		}
	}

	// 7. Inhabilidades
	if p.Inhabilidades != nil {
		insertInh := `
			INSERT INTO inhabilidades (
				postulacion_id, goce_derechos_politicos, inhabilitado_comercio, mora_obligaciones_estado,
				vinculo_sistema_financiero, funcionario_iess_biess, interes_aseguradoras,
				removido_control, removido_institucion, removido_cargo, removido_fecha_fin, removido_motivo,
				obligaciones_sri, castigo_obligaciones_financieras, litigio_iess_biess,
				procesado_corrupcion, responsabilidad_contraloria, registrado_uafe
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		`
		_, err = tx.Exec(ctx, insertInh,
			postulacionID,
			p.Inhabilidades.GoceDerechosPoliticos,
			p.Inhabilidades.InhabilitadoComercio,
			p.Inhabilidades.MoraObligacionesEstado,
			p.Inhabilidades.VinculoSistemaFinanciero,
			p.Inhabilidades.FuncionarioIessBiess,
			p.Inhabilidades.InteresAseguradoras,
			p.Inhabilidades.RemovidoControl,
			p.Inhabilidades.RemovidoInstitucion,
			p.Inhabilidades.RemovidoCargo,
			p.Inhabilidades.RemovidoFechaFin,
			p.Inhabilidades.RemovidoMotivo,
			p.Inhabilidades.ObligacionesSRI,
			p.Inhabilidades.CastigoObligacionesFinancieras,
			p.Inhabilidades.LitigioIessBiess,
			p.Inhabilidades.ProcesadoCorrupcion,
			p.Inhabilidades.ResponsabilidadContraloria,
			p.Inhabilidades.RegistradoUAFE,
		)
		if err != nil {
			return "", fmt.Errorf("error al guardar inhabilidades: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error al hacer commit del guardado: %w", err)
	}

	return codigo, nil
}

func (r *pgxPostulacionRepository) EnviarPostulacion(ctx context.Context, postulanteID int) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error al iniciar transaccion: %w", err)
	}
	defer tx.Rollback(ctx)

	var postulacionID int
	var codigo string
	var estado string

	query := `SELECT id, codigo, estado FROM postulaciones WHERE postulante_id = $1 AND deleted_at IS NULL`
	err = tx.QueryRow(ctx, query, postulanteID).Scan(&postulacionID, &codigo, &estado)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("POSTULACION_NO_ENCONTRADA")
		}
		return "", err
	}

	if estado == "ENVIADA" {
		return "", errors.New("POSTULACION_YA_ENVIADA")
	}

	// Verificar secciones mínimas requeridas: domicilio, cargo e inhabilidades
	var tieneDomicilio bool
	_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM domicilios WHERE postulacion_id = $1)`, postulacionID).Scan(&tieneDomicilio)
	if !tieneDomicilio {
		return "", errors.New("DEBE_COMPLETAR_DOMICILIO")
	}

	var cargo *string
	_ = tx.QueryRow(ctx, `SELECT cargo FROM postulaciones WHERE id = $1`, postulacionID).Scan(&cargo)
	if cargo == nil || *cargo == "" {
		return "", errors.New("DEBE_SELECCIONAR_CARGO")
	}

	var tieneInhabilidades bool
	_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM inhabilidades WHERE postulacion_id = $1)`, postulacionID).Scan(&tieneInhabilidades)
	if !tieneInhabilidades {
		return "", errors.New("DEBE_DECLARAR_INHABILIDADES")
	}

	// L1: Guarda acepta_terminos = true y fecha_consentimiento = now()
	updateQuery := `
		UPDATE postulaciones 
		SET estado = 'ENVIADA', 
		    acepta_terminos = TRUE, 
		    fecha_consentimiento = now(), 
		    updated_at = now()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, updateQuery, postulacionID)
	if err != nil {
		return "", fmt.Errorf("error al enviar postulacion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error al hacer commit del envio: %w", err)
	}

	return codigo, nil
}
