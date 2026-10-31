# Fase 0 — Acuerdos

**Fecha:** viernes 30 · **Hito H0:** contrato de API aprobado por los 3 y entorno corriendo en las 3 máquinas.

## Tareas (todos)

- [ ] Leer `docs/requerimiento.pdf` (EF-11 y EF-12) y [requisitos implícitos](00-requisitos-implicitos.md).
- [ ] Aprobar o cambiar el contrato de abajo.
- [x] Elegir framework: **Gin**. Usa `net/http` y `context.Context` nativos (compatibles con `pgx` y `go-redis`); Fiber (`fasthttp`) reutiliza memoria entre peticiones y obliga a copiar valores del contexto.
- [x] Asignar quién lleva cada pista: D = @ginobermudez76, B = @SaidJuniorProgrammer, F = @Misstery13.
- [ ] Clonar, `cp .env.example .env`, `docker compose up -d`, `npm run dev`, `go run ./cmd/api`.

## Contrato de API

| Método | Ruta | Body | Respuestas |
|--------|------|------|------------|
| POST | `/api/auth/verificar` | `{cedula, fecha_expedicion?, codigo_dactilar?}` | 200 `{nombres}` · 401 "la información ingresada es incorrecta" · 429 "registro fallido, inténtelo nuevamente en 10 minutos" |
| POST | `/api/auth/registro` | `{cedula, email, password}` | 201 · 409 ya registrado |
| POST | `/api/auth/login` | `{cedula, password}` | 200 `{token}` · 401 "sus credenciales de ingreso son incorrectas" · 429 |
| POST | `/api/auth/recuperar` | `{cedula, fecha_expedicion?, codigo_dactilar?}` | 200 (correo simulado en log) |
| POST | `/api/auth/restablecer` | `{token, password}` | 200 · 400 token inválido |
| GET | `/api/catalogos/paises` | — | 200 `[{id, nombre}]` |
| GET | `/api/catalogos/paises/:id/provincias` | — | 200 `[{id, nombre}]` |
| GET | `/api/catalogos/provincias/:id/cantones` | — | 200 `[{id, nombre}]` |
| GET | `/api/postulante/me` | — | 200 datos Registro Civil · 403 no ecuatoriano |
| POST | `/api/postulacion/validar-cargo` | `{cargo}` | 200 `{habilitado, mensaje}` |
| GET | `/api/postulacion` | — | 200 postulación completa · 404 si no existe |
| PUT | `/api/postulacion` | postulación completa | 200 `{codigo}` (transacción) |
| POST | `/api/postulacion/documentos` | `multipart: archivo, tipo` | 201 `{ruta}` · 400 no es PDF / excede tamaño |
| POST | `/api/postulacion/enviar` | `{acepta_terminos: true}` | 200 · 409 ya enviada |

Rutas `/api/postulante/*` y `/api/postulacion/*` requieren JWT.
