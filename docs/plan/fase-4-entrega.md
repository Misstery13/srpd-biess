# Fase 4 — Integración y entrega

**Fecha:** lunes 2 (noche) – martes 3 · **Hito H4:** `main` con tag `v1.0`, un compañero clona desde cero y levanta todo siguiendo solo el README.

## Todos
- [ ] Merge de todas las ramas a `develop`; PR `develop` → `main`.
- [ ] Prueba completa desde un clon limpio: `docker compose down -v && docker compose up -d`.
- [ ] README: prerrequisitos, ejecución, `.env.example`, usuarios de prueba (lo pide la entrega).
- [ ] Revisar que no haya `console.log`, credenciales ni `.env` en el repo.
- [ ] `git tag v1.0 && git push --tags`.

## Sustentación (10 min)
- [ ] 1 min — Problema y alcance (EF-11 y EF-12).
- [ ] 3 min — Arquitectura: diagrama handler → service → repository, inyección de dependencias.
- [ ] 4 min — Demo: bloqueo de 3 intentos, cascada, colombiano bloqueado, guardado transaccional.
- [ ] 2 min — Redis (claves y TTL) y transacción (BEGIN / COMMIT / ROLLBACK).
- [ ] Ensayo: cada integrante explica una parte que **no** programó.
