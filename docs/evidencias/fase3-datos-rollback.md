# Evidencia Técnica: Persistencia Relacional, Transacciones y ROLLBACK (Fase 3 · Datos)

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos (Base de Datos, Repositorios y Caché)  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Criterio de Evaluación:** Criterio 2 — Persistencia Relacional & Transacciones (PostgreSQL) — **8.0 Puntos**  
**Issue de Referencia:** [#10 — [F3·Datos] Guardado transaccional de la postulación](https://github.com/Misstery13/srpd-biess/issues/10)

---

## 1. Arquitectura de Transaccionalidad Multi-Tabla

Para el Concurso de Méritos del Directorio del BIESS, el guardado de la postulación (`EF-12`) involucra 8 tablas dependientes:
1. `postulaciones` (Cabecera y código único `BIESS-2026-XXXXXX`)
2. `domicilios` (Ubicación nacional o extranjera)
3. `familiares_iess` (Declaración de parentesco institucional)
4. `formacion_academica` (Títulos nacionales y extranjeros)
5. `experiencia_laboral` (Cálculo dinámico de trayectoria)
6. `publicaciones` (Artículos e investigaciones)
7. `capacitaciones` (Cursos y seminarios con validación `horas > 0`)
8. `inhabilidades` (18 declaraciones juradas de inhabilidad legal)

### Implementación en Go (`backend/internal/repository/postulacion.go`):
```go
func (r *pgxPostulacionRepository) GuardarTransaccional(ctx context.Context, postulanteID int, p *model.PostulacionCompleta) (string, error) {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return "", fmt.Errorf("error al iniciar transaccion: %w", err)
    }
    defer tx.Rollback(ctx) // Garantía de ROLLBACK ante cualquier pánico o error intermedio

    // Inserción o actualización de cabecera...
    // Limpieza atómica de tablas hijas anteriores (B3)...
    // Inserción secuencial de domicilios, familiares, formación, experiencia, etc.

    if err := tx.Commit(ctx); err != nil {
        return "", fmt.Errorf("error al hacer commit del guardado: %w", err)
    }
    return codigo, nil
}
```

---

## 2. Demostración Práctica del ROLLBACK (Prueba de Fallo Forzado)

Para comprobar la **atomicidad (propiedad ACID)**, se implementó una prueba controlada que simula un fallo en la última inserción de la postulación:

### A. Prueba Unitaria en Go (`backend/internal/repository/postulacion_test.go`)
La prueba `TestPostulacionRepository_TransaccionalYRollback`:
1. Registra una postulación con datos válidos en borrador.
2. Intenta guardar una modificación que incluye una capacitación con `Horas: -10`, violando la restricción SQL `CHECK (horas > 0)`.
3. Comprueba que el método retorna error y que la base de datos ejecuta `ROLLBACK`.
4. Comprueba mediante `ObtenerPorPostulanteID` que los datos preexistentes permanecen **100% intactos**, sin datos huérfanos ni cabeceras corrompidas.

### B. Script de Ejecución SQL Directo (`db/prueba_rollback.sql`)
Permite reproducir el comportamiento transaccional en PostgreSQL:
```bash
./scripts/ejecutar_prueba_rollback.sh
```

**Salida de Ejecución:**
```text
========================================================================
1. Verificando estado inicial antes de la transacción...
========================================================================
 total_postulaciones | total_domicilios | total_familiares | total_capacitaciones 
---------------------+------------------+------------------+----------------------
                   0 |                0 |                0 |                    0

========================================================================
2. Iniciando transacción multi-tabla y forzando error en tabla hija...
========================================================================
NOTICE:  -> Cabecera creada exitosamente: ID=1, Código=BIESS-2026-000001
NOTICE:  -> Domicilio insertado exitosamente.
NOTICE:  -> Familiar IESS insertado exitosamente.
NOTICE:  -> Intentando insertar capacitación inválida con horas = -10 (forzando violación CHECK)...
WARNING: ¡Error de restricción CHECK capturado exitosamente! Motivo: Las horas de capacitación deben ser mayores a 0.
NOTICE:  Ejecutando ROLLBACK integral... Ningún registro debe persistir.
ERROR:   el nuevo registro para la relación «capacitaciones» viola la restricción «check» «capacitaciones_horas_check»

========================================================================
3. Verificando estado final tras el ROLLBACK (Atomicidad garantizada)...
========================================================================
 total_postulaciones | total_domicilios | total_familiares | total_capacitaciones 
---------------------+------------------+------------------+----------------------
                   0 |                0 |                0 |                    0

CONCLUSIÓN: No quedaron filas huérfanas ni cabeceras incompletas en la base de datos.
========================================================================
```

---

## 3. Cumplimiento de la Ley Orgánica de Protección de Datos Personales (LOPDP)

Conforme a lo acordado en el **PR #15**, la persistencia implementa el derecho de supresión y minimización mediante el método `SuprimirPostulante`:
1. **Baja Lógica del Postulante:** Se asigna `deleted_at = now()`.
2. **Anonimización:** El correo se sustituye por `eliminado-{id}@anonimo.invalid` y el `password_hash` se borra físicamente.
3. **Eliminación de Datos de Terceros:** Los datos de cónyuge y familiares IESS se eliminan físicamente de inmediato para proteger la privacidad de terceros (Art. 12 LOPDP).
4. **Preservación Legal:** Se conservan únicamente la cédula, el código `BIESS-2026-XXXXXX` y las fechas de auditoría para garantizar la transparencia del concurso público ante la Contraloría.
5. **Índices Únicos Parciales:** La restricción `WHERE deleted_at IS NULL` permite que si el ciudadano desea ejercer su derecho de postulación en un nuevo proceso, el índice no colisione con el registro suprimido.

