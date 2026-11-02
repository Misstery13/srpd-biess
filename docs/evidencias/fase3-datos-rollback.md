# Evidencia Técnica: Persistencia Relacional, Transacciones y ROLLBACK (Fase 3 · Datos)

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos (Base de Datos, Repositorios y Caché)  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Criterio de Evaluación:** Criterio 2 — Persistencia Relacional & Transacciones (PostgreSQL) — **8.0 Puntos**  
**Issue de Referencia:** [#10 — [F3·Datos] Guardado transaccional de la postulación](https://github.com/Misstery13/srpd-biess/issues/10)

---

## 1. Arquitectura de Transaccionalidad Multi-Tabla

Para el Concurso de Méritos del Directorio del BIESS, el guardado de la postulación (`EF-12`) involucra 8 tablas relacionales dependientes:
1. `postulaciones` (Cabecera y código único `BIESS-2026-XXXXXX`)
2. `domicilios` (Ubicación nacional o extranjera)
3. `familiares_iess` (Declaración de parentesco institucional de 4to grado de consanguinidad o 2do de afinidad)
4. `formacion_academica` (Títulos de 3er y 4to nivel)
5. `experiencia_laboral` (Cálculo dinámico de trayectoria)
6. `publicaciones` (Artículos e investigaciones)
7. `capacitaciones` (Cursos y seminarios con validación `CHECK (horas > 0)`)
8. `inhabilidades` (**13** declaraciones juradas de inhabilidad legal, conforme al Art. 1.9 del pliego)

### Implementación Oficial del Sistema
La lógica transaccional está implementada directamente en el repositorio de Go:  
👉 Ver código fuente en [`backend/internal/repository/postulacion.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/repository/postulacion.go) (`GuardarTransaccional`).

El flujo inicia una transacción de base de datos con `r.pool.Begin(ctx)` y garantiza mediante `defer tx.Rollback(ctx)` que si ocurre cualquier error durante la inserción secuencial de cabecera o de alguna de las tablas hijas, la transacción se revierte íntegramente.

---

## 2. Validación de Consistencia y Demostración del ROLLBACK

Para verificar la **atomicidad (propiedad ACID)** contra el código real del sistema y evitar duplicaciones de sentencias SQL, la validación se apoya en el test de integración de Go:
👉 [`backend/internal/repository/postulacion_test.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/repository/postulacion_test.go) (`TestPostulacionRepository_TransaccionalYRollback`).

### Metodología de la Prueba:
1. Registra un postulante temporal de prueba y guarda una postulación completa con éxito en estado `BORRADOR`.
2. Intenta guardar una modificación que incluye una capacitación inválida con `Horas: -10`, provocando intencionalmente la excepción SQL `capacitaciones_horas_check`.
3. Comprueba que el método `GuardarTransaccional` captura el error y dispara el `ROLLBACK`.
4. Consulta las tablas dependientes y verifica que los datos preexistentes permanecen intactos, sin registros huérfanos ni cabeceras inconsistentes.
5. Limpia los datos de prueba en la cláusula `defer`.

### Scripts de Ejecución Automatizada:
Para ejecutar la prueba localmente o en un contenedor Docker con PostgreSQL levantado desde cero:
```bash
# En Linux / macOS / Bash:
./scripts/ejecutar_prueba_rollback.sh

# En Windows PowerShell:
.\scripts\ejecutar_prueba_rollback.ps1
```

**Resultado de Ejecución Real (`go test` con PostgreSQL en Docker):**
```text
SRPD BIESS - Validacion de Atomicidad y ROLLBACK Transaccional (Go)
Contenedor PostgreSQL activo.
Ejecutando TestPostulacionRepository_TransaccionalYRollback...
=== RUN   TestPostulacionRepository_TransaccionalYRollback
--- PASS: TestPostulacionRepository_TransaccionalYRollback (0.22s)
PASS
ok  	github.com/Misstery13/srpd-biess/backend/internal/repository	0.506s
EXITO: Transaccion probada contra el repositorio real de Go.
Se verifico que ante error de restriccion (CHECK horas > 0), PostgreSQL ejecuto ROLLBACK completo
y no quedaron registros huerfanos en ninguna tabla hija.
```

> [!NOTE]
> **Comportamiento de Secuencias en PostgreSQL:**
> En PostgreSQL, las secuencias (`postulaciones_id_seq`) son transaccionalmente no reversibles (`nextval` no retrocede ante un `ROLLBACK` para evitar contención de bloqueos entre sesiones concurrentes). El test automatizado de Go maneja la limpieza y el estado sin comprometer la correlación numérica en producción.

---

## 3. Cumplimiento de la Ley Orgánica de Protección de Datos Personales (LOPDP)

Conforme a lo consensuado en el **PR #15** y en estricto cumplimiento del principio de minimización y derecho de supresión (Art. 12 LOPDP):

1. **Baja Lógica y Despersonalización:**  
   La cuenta del postulante no se destruye de inmediato para fines de auditoría; se marca con `deleted_at = now()`, su correo se anonimiza (`eliminado-{id}@anonimo.invalid`) y su `password_hash` se borra físicamente.
2. **Supresión Física de Datos de Postulación:**  
   Al solicitar la baja mediante `SuprimirPostulacion`, se eliminan físicamente en cascada las tablas dependientes de la postulación (`domicilios`, `familiares_iess`, `formacion_academica`, `experiencia_laboral`, `publicaciones`, `capacitaciones`, `documentos_postulacion` e `inhabilidades`).
3. **Datos de Identidad y Cónyuge:**  
   Los datos personales de identidad y del cónyuge provienen de la simulación oficial del Registro Civil (`registro_civil`) y no son modificados ni eliminados por el sistema de postulación, preservando la inmutabilidad de la fuente de identidad fidedigna.
