# Fase 1 — Cimientos

**Fecha:** sábado 31 (mañana) · **Hito H1:** en el navegador, al elegir Ecuador → Pichincha aparecen sus cantones, y la segunda consulta sale de Redis (verificable con `redis-cli KEYS catalogo:*`).

## D — Datos
- [ ] Agregar a `db/init.sql`: `UNIQUE (postulante_id)`, tablas `publicaciones`, `capacitaciones`, `inhabilidades` (P1, P5).
- [ ] Conexión a PostgreSQL con `pgxpool` y a Redis con `go-redis`.
- [ ] `CatalogoRepository`: países, provincias por país, cantones por provincia.
- [ ] Caché de catálogos con TTL 24 h y fallback a PostgreSQL.

## B — Backend
- [ ] `internal/config`: leer `.env`.
- [ ] Fiber/Gin en `main.go` con **inyección de dependencias**: `repo := ...; svc := service.New(repo); h := handler.New(svc)`.
- [ ] Middleware de errores con formato `{"error": ...}`.
- [ ] Handlers de los 3 endpoints de catálogos.

## F — Frontend
- [ ] `react-router`: rutas `/login`, `/registro`, `/recuperar`, `/postulacion`.
- [ ] `services/api.js`: instancia de axios con `baseURL=/api` e interceptor del token.
- [ ] Layout base con Tailwind (encabezado, contenedor responsivo).
- [ ] Componente `SelectorUbicacion` (País → Provincia → Cantón, texto libre si no es Ecuador), con loader y estado de error.

## Criterio de aceptación
- [ ] `go vet ./...` y `npm run lint` sin errores.
- [ ] Cascada funcionando contra la API real.
