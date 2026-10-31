# Plan de implementación

Entrega: **martes 3 de noviembre de 2026**. Alcance: **EF-11** y **EF-12** (EF-13 y EF-14 quedan fuera, ver [requisitos implícitos](00-requisitos-implicitos.md)).

Cada fase termina en un **hito verificable de punta a punta** (BD → API → pantalla). Dentro de cada fase el trabajo se divide en tres pistas que avanzan en paralelo:

| Pista | Responsable | Alcance |
|-------|-------------|---------|
| **D — Datos** | Gino Bermúdez (@ginobermudez76) | PostgreSQL, Redis, `internal/repository`, `internal/cache` |
| **B — Backend** | Said (@SaidJuniorProgrammer) | `cmd/api`, `internal/config`, `internal/handler`, `internal/service` (framework: **Gin**) |
| **F — Frontend** | Diana Melena (@Misstery13) | `frontend/` |

Las tareas de cada fase están como **issues** en GitHub, agrupadas por milestone (Fase 1, Fase 2…).

| Fase | Fecha | Hito (se cumple cuando…) |
|------|-------|--------------------------|
| [0 — Acuerdos](fase-0-acuerdos.md) | Vie 30 | Contrato de API aprobado y los 3 levantan el entorno |
| [1 — Cimientos](fase-1-cimientos.md) | Sáb 31 | La cascada País → Provincia → Cantón se ve en pantalla, servida desde Redis |
| [2 — EF-11 Autenticación](fase-2-ef11-autenticacion.md) | Sáb 31 – Dom 1 | Registro, login, bloqueo de 10 min y ruta protegida funcionan |
| [3 — EF-12 Postulación](fase-3-ef12-postulacion.md) | Dom 1 – Lun 2 | Se guarda y envía una postulación completa en una transacción |
| [4 — Integración y entrega](fase-4-entrega.md) | Lun 2 – Mar 3 | `main` con tag `v1.0`, README al día, sustentación ensayada |

## Reglas del equipo

- Una rama por tarea desde `develop`: `feature/f2-login`, `fix/cascada`…
- PR a `develop` con al menos una revisión de otro integrante.
- Si algo bloquea más de 1 hora, se avisa en el grupo.
- Todos deben poder explicar todo el código en la sustentación.
