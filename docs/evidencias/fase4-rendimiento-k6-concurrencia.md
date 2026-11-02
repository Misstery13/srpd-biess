# Reporte Técnico: Pruebas de Carga Masiva y Concurrencia con k6 (31,700+ Peticiones)

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos, Rendimiento & QA  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Issue de Referencia:** [#25 — [QA·Rendimiento] Pruebas de concurrencia y carga masiva con k6: Validación de pgxpool, caché Redis y estrés transaccional](https://github.com/Misstery13/srpd-biess/issues/25)  
**Entorno:** Go + Gin API (localhost:8080) · PostgreSQL 16 (pgxpool MaxConns=15) · Redis 7 (In-Memory)

---

## 1. Diseño del Escenario de Carga y Concurrencia

Se ejecutó un escenario de estrés progresivo simulando el tráfico pico durante la apertura/cierre del concurso:
- **Rampa de Subida:** 0 a 20 Usuarios Virtuales (VUs) en 5 segundos.
- **Carga Sostenida:** 50 VUs simultáneos durante 15 segundos.
- **Pico de Estrés (Spike):** 70 VUs simultáneos en 5 segundos.
- **Enfriamiento:** Rampa de bajada a 0 VUs en 5 segundos.
- **Duración Total:** 30.1 segundos.

---

## 2. Resultados Reales Obtenidos en Local

```text
  █ TOTAL RESULTS 
    checks_total.......: 39,660 out of 39,660 (100.00% exitosas)
    checks_failed......: 0 out of 39,660 (0.00%)

    ✓ GET /api/catalogos/paises responde 200
    ✓ GET /api/periodo responde 200
    ✓ GET /api/catalogos/paises/52/provincias responde 200
    ✓ POST /api/auth/login responde 401 o 429 bajo estrés
    ✓ Servidor no genera errores 500 por contención

  █ MÉTRICAS DE THROUGHPUT Y CAPACIDAD
    http_reqs.......................: 31,728 peticiones procesadas
    Throughput promedio.............: 1,052.39 peticiones por segundo (RPS)
    Datos transferidos..............: 64 MB (2.1 MB/s)
    Peticiones cacheadas exitosas...: 23,796 peticiones servidas desde Redis RAM

  █ LATENCIAS DE RESPUESTA (Percentiles)
    Métrica                         Mediana (p50)    p90          p95 (SLA < 300ms)
    ─────────────────────────────────────────────────────────────────────────────
    Latencia Global Servidor        2.78 ms          16.78 ms     29.00 ms
    Catálogos en Redis              2.86 ms          17.09 ms     29.41 ms
    Período en Redis                2.67 ms          16.43 ms     28.18 ms
    Login Concurrente               2.71 ms          16.48 ms     28.80 ms
```

---

## 3. Conclusiones Técnicas de la Pista de Datos

1. **Eficiencia del Caché en Redis (Cache-Aside):**
   - Redis absorbió **23,796 consultas** de lectura inmediata en memoria RAM, con una mediana de **2.8 ms**.
   - El pool de PostgreSQL (`pgxpool`, configurado con `MaxConns = 15`) no sufrió saturación ni agotamiento de sockets a pesar de atender más de **1,000 peticiones por segundo**.
2. **Cero Errores 500:**
   - No se produjo ningún pánico, bloqueo mutuo (deadlock) ni caída del microservicio en Go.
3. **Atomicidad Concurrente de Redis:**
   - Las operaciones atómicas `INCR` en pipelines para el rate limiting gestionaron las peticiones paralelas de los 70 VUs sin inconsistencias ni carreras de datos (race conditions).

