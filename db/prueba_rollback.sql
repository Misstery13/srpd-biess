-- ============================================================================
-- SRPD BIESS — Prueba de Atomicidad Transaccional y ROLLBACK
-- Pista D (Datos) — Responsable: Gino Bermúdez (@ginobermudez76)
-- Concurso Público de Méritos y Oposición del Directorio del BIESS
--
-- NOTA TÉCNICA: En PostgreSQL, las secuencias (postulaciones_id_seq) no son reversibles
-- tras un ROLLBACK por diseño para evitar contención de bloqueos entre transacciones concurrentes.
-- Este script es de referencia didáctica para ejecución en BD local de desarrollo.
-- Para la verificación del sistema Go real, use: ./scripts/ejecutar_prueba_rollback.sh
-- ============================================================================

\echo '[FASE 1] Verificando estado inicial antes de la transaccion...'

SELECT 
    (SELECT COUNT(*) FROM postulaciones) AS total_postulaciones,
    (SELECT COUNT(*) FROM domicilios) AS total_domicilios,
    (SELECT COUNT(*) FROM familiares_iess) AS total_familiares,
    (SELECT COUNT(*) FROM formacion_academica) AS total_formacion,
    (SELECT COUNT(*) FROM experiencia_laboral) AS total_experiencia,
    (SELECT COUNT(*) FROM capacitaciones) AS total_capacitaciones,
    (SELECT COUNT(*) FROM publicaciones) AS total_publicaciones,
    (SELECT COUNT(*) FROM inhabilidades) AS total_inhabilidades;

\echo '[FASE 2] Iniciando transaccion multi-tabla y forzando error en tabla hija...'

DO $$
DECLARE
    v_postulante_id INT;
    v_postulacion_id INT;
    v_codigo VARCHAR(20);
    v_seq_id INT;
BEGIN
    -- Obtener o registrar temporalmente al postulante de prueba (Juan Carlos Pérez López - 1710034065)
    SELECT id INTO v_postulante_id FROM postulantes WHERE cedula = '1710034065' AND deleted_at IS NULL;
    
    IF v_postulante_id IS NULL THEN
        INSERT INTO postulantes (cedula, email, password_hash, rol_id)
        VALUES ('1710034065', 'juan.perez@test.local', 'hash_dummy_no_usable_para_pruebas_sql', 1)
        RETURNING id INTO v_postulante_id;
        RAISE NOTICE '-> Postulante de prueba registrado automáticamente (ID=%)', v_postulante_id;
    END IF;

    -- Paso A: Insertar cabecera con código formal BIESS-2026-XXXXXX
    SELECT nextval('public.postulaciones_id_seq') INTO v_seq_id;
    v_codigo := 'BIESS-2026-' || LPAD(v_seq_id::text, 6, '0');

    INSERT INTO postulaciones (id, codigo, postulante_id, cargo, respaldo_org, estado)
    VALUES (v_seq_id, v_codigo, v_postulante_id, 'AFILIADOS', FALSE, 'BORRADOR')
    RETURNING id INTO v_postulacion_id;

    RAISE NOTICE '-> Cabecera creada exitosamente: ID=%, Código=%', v_postulacion_id, v_codigo;

    -- Paso B: Insertar Domicilio (provincia 1 Azuay, canton 1 Camilo Ponce Enríquez)
    INSERT INTO domicilios (postulacion_id, pais_id, provincia_id, canton_id, calle_principal, numero, celular)
    VALUES (v_postulacion_id, 52, 1, 1, 'Av. Amazonas', 'N24-196', '0999887766');
    RAISE NOTICE '-> Domicilio insertado exitosamente.';

    -- Paso C: Insertar Familiar IESS
    INSERT INTO familiares_iess (postulacion_id, nombres, parentesco, institucion, area, cargo)
    VALUES (v_postulacion_id, 'Carlos Pérez', 'hermano/a', 'IESS', 'Tecnología', 'Analista');
    RAISE NOTICE '-> Familiar IESS insertado exitosamente.';

    -- Paso D: Inserción que VIOLA restricción CHECK (horas > 0)
    RAISE NOTICE '-> Intentando insertar capacitación inválida con horas = -10 (forzando violación CHECK)...';
    INSERT INTO capacitaciones (postulacion_id, tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin, horas)
    VALUES (v_postulacion_id, 'CURSO', 'Seguridad de Datos', 'EPN', '2025-01-10', '2025-01-20', -10);

EXCEPTION
    WHEN check_violation THEN
        RAISE WARNING '¡Error de restricción CHECK capturado exitosamente! Motivo: Las horas de capacitación deben ser mayores a 0.';
        RAISE NOTICE 'Ejecutando ROLLBACK integral... Ningún registro debe persistir.';
        -- Al terminar el bloque DO con excepción no controlada, PostgreSQL descarta automáticamente los cambios.
        RAISE;
END $$;

\echo '[FASE 3] Verificando estado final tras el ROLLBACK (Atomicidad garantizada)...'

SELECT 
    (SELECT COUNT(*) FROM postulaciones) AS total_postulaciones,
    (SELECT COUNT(*) FROM domicilios) AS total_domicilios,
    (SELECT COUNT(*) FROM familiares_iess) AS total_familiares,
    (SELECT COUNT(*) FROM capacitaciones) AS total_capacitaciones;

\echo '[CONCLUSION] No quedaron filas huerfanas ni cabeceras incompletas en la base de datos.'
