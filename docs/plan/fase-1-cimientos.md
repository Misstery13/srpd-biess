# Fase 1 — Cimientos

**Fecha:** sábado 31 (mañana) · **Hito H1:** en el navegador, al elegir Ecuador → Pichincha aparecen sus cantones, y la segunda consulta sale de Redis (verificable con `redis-cli KEYS catalogo:*`).

## D — Datos
- [ ] Agregar a `db/init.sql`: tablas `publicaciones`, `capacitaciones`, `inhabilidades` (P5). Son hijas de la postulación: sin `deleted_at` (B3).
- [x] `rol`, `permiso`, `periodo_postulacion`, `deleted_at` e índices únicos parciales en `init.sql` (R1, B1–B5, T1).
- [ ] Conexión a PostgreSQL con `pgxpool` y a Redis con `go-redis`.
- [ ] `CatalogoRepository`: países, provincias por país, cantones por provincia.
- [ ] Caché de catálogos con TTL 24 h y fallback a PostgreSQL.

## B — Backend
- [ ] `internal/config`: leer `.env`.
- [ ] Gin en `main.go` con **inyección de dependencias**: `repo := ...; svc := service.New(repo); h := handler.New(svc)`.
- [ ] Errores tipados + middleware que los traduce a código HTTP y `{"error": ...}` (E1, E2).
- [ ] Handlers de los 3 endpoints de catálogos.

## F — Frontend
- [ ] `react-router`: rutas `/login`, `/registro`, `/recuperar`, `/postulacion`.
- [ ] `services/api.js`: instancia de axios con `baseURL=/api` e interceptor del token.
- [ ] Layout base con Tailwind (encabezado, contenedor responsivo).
- [ ] `ErrorBoundary` en el layout (E4).
- [ ] Componente `SelectorUbicacion` (País → Provincia → Cantón, texto libre si no es Ecuador), con loader y estado de error.

## Criterio de aceptación
- [ ] `go vet ./...` y `npm run lint` sin errores.
- [ ] Cascada funcionando contra la API real.
