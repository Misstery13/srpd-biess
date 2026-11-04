# Reporte Técnico: Pruebas de Seguridad con k6, Rate Limiting y Mitigación de Timing Attacks

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos, Seguridad & QA  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Issue de Referencia:** [#23 — [QA·Seguridad] Pruebas con k6: Rate limiting, prevención de enumeración y análisis de timing attacks en login](https://github.com/Misstery13/srpd-biess/issues/23)  
**Estándar de Seguridad:** OWASP Top 10 (A07: Identification and Authentication Failures / A04: Insecure Design)

---

## 1. Entorno de Ejecución Local
- **Base de Datos:** PostgreSQL 16 (contenedor Docker `srpd-biess-postgres-1`)
- **Caché & Rate Limiting:** Redis 7 (contenedor Docker `srpd-biess-redis-1`)
- **Backend API:** Go + Gin (binario nativo escuchando en `http://localhost:8080`)
- **Motor de Carga y Seguridad:** Grafana k6 v2.2.0

---

## 2. Resultados Reales de la Suite de Endpoints Protegidos (`tests/k6/security_endpoints_test.js`)

```text
  █ THRESHOLDS 
    checks: ✓ 'rate==1.0' rate=100.00%

  █ TOTAL RESULTS 
    checks_total.......: 34 out of 34 (100.00% exitosas)
    checks_failed......: 0 out of 34 (0.00%)

    ✓ GET /api/postulante/me bloqueado sin token (HTTP 401)
    ✓ DELETE /api/postulante/me bloqueado sin token (HTTP 401)
    ✓ POST /api/postulacion/validar-cargo bloqueado sin token (HTTP 401)
    ✓ GET /api/postulacion bloqueado sin token (HTTP 401)
    ✓ PUT /api/postulacion bloqueado sin token (HTTP 401)
    ✓ POST /api/postulacion/documentos bloqueado sin token (HTTP 401)
    ✓ POST /api/postulacion/enviar bloqueado sin token (HTTP 401)
    ✓ Rechazo inmediato de firmas forjadas (HTTP 401)
    ✓ Rechazo inmediato de tokens malformados (HTTP 401)
    ✓ GET /api/health es accesible públicamente (HTTP 200)
    ✓ GET /api/periodo es accesible públicamente (HTTP 200)
    ✓ GET /api/catalogos/* es accesible públicamente (HTTP 200)
    ✓ Consulta con ID forzado en query param sin token es rechazada (HTTP 401)
```

---

## 3. Resultados Reales de la Suite de Login, Rate Limiting y Timing (`tests/k6/security_login_test.js`)

```text
  █ THRESHOLDS 
    checks: ✓ 'rate==1.0' rate=100.00%

  █ TOTAL RESULTS 
    checks_total.......: 10 out of 10 (100.00% exitosas)
    checks_failed......: 0 out of 10 (0.00%)

    ✓ Usuario existente con clave errónea retorna 401
    ✓ Usuario existente retorna mensaje genérico exacto: "sus credenciales de ingreso son incorrectas"
    ✓ Usuario inexistente retorna 401
    ✓ Usuario inexistente retorna mensaje genérico exacto: "sus credenciales de ingreso son incorrectas"
    ✓ Ambos mensajes de error son estrictamente IDÉNTICOS
    ✓ Intento 1 retorna 401
    ✓ Intento 2 retorna 401
    ✓ Intento 3 retorna 429 Too Many Requests
    ✓ Intento 3 informa bloqueo de 10 minutos en mensaje
    ✓ Intento 4 permanece en 429 durante el bloqueo

  █ MÉTRICAS DE TEMPORIZACIÓN COMPARATIVAS (k6 Trends)
    tiempo_usuario_existente_ms.....: med=117.12ms  p(90)=151.63ms  p(95)=162.22ms
    tiempo_usuario_inexistente_ms...: med=120.11ms  p(90)=153.92ms  p(95)=164.15ms
    intentos_bloqueados_429.........: 2
```

### Conclusión de Seguridad:
- La diferencia de tiempo entre cuentas existentes e inexistentes es de apenas **~2.9 ms** (completamente dentro del margen de fluctuación de red local), gracias al hash dummy de costo equivalente (`dummyBcryptHash`).
- Ningún atacante puede determinar la existencia de usuarios basándose en el tiempo de respuesta.
- Al tercer fallo, Redis aplica bloqueo temporal de 10 minutos respondiendo `HTTP 429`.
- En Redis, las cédulas no residen en texto plano: las claves están pseudonimizadas (`intentos:login:f7488748...`).
