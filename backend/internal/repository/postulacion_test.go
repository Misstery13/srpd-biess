package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Misstery13/srpd-biess/backend/internal/config"
	"github.com/Misstery13/srpd-biess/backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getTestPool(t *testing.T) *pgxpool.Pool {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "super-secret-key-for-test-at-least-32-chars-long")
	}
	if os.Getenv("POSTGRES_PORT") == "" {
		_ = os.Setenv("POSTGRES_PORT", "5433")
	}
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Omitiendo test de integración con BD: no se pudo cargar config (%v)", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN())
	if err != nil {
		t.Skipf("Omitiendo test de integración: %v", err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL no disponible en %s: %v", cfg.PostgresDSN(), err)
		return nil
	}

	return pool
}

func TestPostulacionRepository_TransaccionalYRollback(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()

	// Usamos la cédula semilla de Juan Pérez '1710034065'
	postRepo := NewPostulanteRepository(pool)
	rol, err := postRepo.BuscarRolPorCodigo(ctx, "POSTULANTE")
	if err != nil || rol == nil {
		t.Fatalf("rol POSTULANTE no encontrado: %v", err)
	}

	// Limpiar cualquier postulante de prueba previo
	_, _ = pool.Exec(ctx, "DELETE FROM postulantes WHERE cedula = '1710034065'")

	// Crear postulante activo
	p := &model.Postulante{
		Cedula:       "1710034065",
		Email:        "test.transaccion@biess.fin.ec",
		PasswordHash: "hash-dummy",
		RolID:        rol.ID,
	}
	if err := postRepo.Crear(ctx, p); err != nil {
		t.Fatalf("error creando postulante de prueba: %v", err)
	}
	defer func() {
		// Limpieza final
		_, _ = pool.Exec(ctx, "DELETE FROM postulantes WHERE id = $1", p.ID)
	}()

	repo := NewPostulacionRepository(pool)

	// 1. Guardado exitoso con todas las secciones
	cargo := "AFILIADOS"
	respaldo := true
	orgNombre := "Federación de Trabajadores"
	provinciaID := 17 // Pichincha
	cantonID := 1     // Quito

	postulacion := &model.PostulacionCompleta{
		PostulacionHeader: model.PostulacionHeader{
			Cargo:        &cargo,
			RespaldoOrg:  &respaldo,
			OrgNombre:    &orgNombre,
		},
		Domicilio: &model.Domicilio{
			PaisID:         52, // Ecuador
			ProvinciaID:    &provinciaID,
			CantonID:       &cantonID,
			Celular:        "0991234567",
			CallePrincipal: "Av. 10 de Agosto",
		},
		Familiares: []model.FamiliarIESS{
			{
				Nombres:     "Pedro Pérez",
				Parentesco:  "hermano/a",
				Institucion: "IESS",
			},
		},
		Inhabilidades: &model.Inhabilidades{
			GoceDerechosPoliticos:          true,
			InhabilitadoComercio:           false,
			MoraObligacionesEstado:         false,
			VinculoSistemaFinanciero:       false,
			FuncionarioIessBiess:           false,
			InteresAseguradoras:            false,
			RemovidoControl:                false,
			ObligacionesSRI:                false,
			CastigoObligacionesFinancieras: false,
			LitigioIessBiess:               false,
			ProcesadoCorrupcion:            false,
			ResponsabilidadContraloria:     false,
			RegistradoUAFE:                 false,
		},
	}

	codigo, err := repo.GuardarTransaccional(ctx, p.ID, postulacion)
	if err != nil {
		t.Fatalf("error guardando postulacion: %v", err)
	}
	if codigo == "" {
		t.Error("se esperaba código generado con formato BIESS-2026-XXXXXX")
	}

	// Comprobar que se guardó correctamente
	obtenida, err := repo.ObtenerPorPostulanteID(ctx, p.ID)
	if err != nil || obtenida == nil {
		t.Fatalf("error obteniendo postulación guardada: %v", err)
	}
	if len(obtenida.Familiares) != 1 {
		t.Errorf("se esperaba 1 familiar guardado, se obtuvieron %d", len(obtenida.Familiares))
	}

	// 2. PRUEBA DE ROLLBACK (Criterio de aceptación H3 / P4):
	// Intentar guardar forzando un error en una tabla hija (ej: violación de clave foránea o horas negativas)
	postulacionConError := &model.PostulacionCompleta{
		PostulacionHeader: model.PostulacionHeader{
			Cargo: &cargo,
		},
		Capacitaciones: []model.Capacitacion{
			{
				TipoEvento:   "CURSO",
				NombreEvento: "Gestión de Riesgos",
				Institucion:  "Universidad",
				FechaInicio:  time.Now(),
				FechaFin:     time.Now().Add(24 * time.Hour),
				Horas:        -10, // Viola CHECK (horas > 0) -> Forzará error SQL en inserción hija
			},
		},
	}

	_, errFallo := repo.GuardarTransaccional(ctx, p.ID, postulacionConError)
	if errFallo == nil {
		t.Fatal("se esperaba error en la transacción debido a CHECK (horas > 0)")
	}

	// Verificar atomicidad: tras el ROLLBACK, los datos originales siguen intactos y no quedaron filas huérfanas
	trasFallo, err := repo.ObtenerPorPostulanteID(ctx, p.ID)
	if err != nil || trasFallo == nil {
		t.Fatalf("error obteniendo postulación tras fallo: %v", err)
	}
	if len(trasFallo.Familiares) != 1 {
		t.Errorf("el rollback falló: se esperaban los datos previos (1 familiar), se obtuvo: %d", len(trasFallo.Familiares))
	}

	// 3. Prueba de Supresión LOPDP (B4)
	if err := postRepo.SuprimirPostulante(ctx, p.ID); err != nil {
		t.Fatalf("error suprimiendo postulante: %v", err)
	}

	// Verificar que el postulante quedó con deleted_at NOT NULL y anonimizado
	var emailAnonimo string
	var deletedAt *time.Time
	err = pool.QueryRow(ctx, "SELECT email, deleted_at FROM postulantes WHERE id = $1", p.ID).Scan(&emailAnonimo, &deletedAt)
	if err != nil || deletedAt == nil {
		t.Errorf("el postulante debería tener deleted_at asignado: %v", err)
	}

	// Verificar que las tablas hijas fueron eliminadas físicamente (B4)
	var countFamiliares int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM familiares_iess WHERE postulacion_id = $1", obtenida.ID).Scan(&countFamiliares)
	if countFamiliares != 0 {
		t.Errorf("las tablas hijas deberían haberse eliminado físicamente en la supresión, quedaron: %d", countFamiliares)
	}

	// Verificar que en postulaciones se conserva solo cédula, código y fechas (B4)
	var cargoPostBaja *string
	var orgNombreBaja *string
	_ = pool.QueryRow(ctx, "SELECT cargo, org_nombre FROM postulaciones WHERE id = $1", obtenida.ID).Scan(&cargoPostBaja, &orgNombreBaja)
	if cargoPostBaja != nil || orgNombreBaja != nil {
		t.Errorf("postulación dada de baja debe limpiar cargo y organización (B4)")
	}
}
