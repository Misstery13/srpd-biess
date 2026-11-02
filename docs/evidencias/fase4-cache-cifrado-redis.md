# Reporte Técnico: Optimización de Caché en Redis y Cifrado AES-256-GCM de Datos Sensibles

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos, Rendimiento & Seguridad  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Issue de Referencia:** [#24 — [F4·Datos/Seguridad] Optimización de caché en Redis (RBAC/Período) y cifrado AES-256-GCM de datos sensibles](https://github.com/Misstery13/srpd-biess/issues/24)  
**Estándares:** OWASP Top 10 (A02: Cryptographic Failures) & LOPDP Art. 40 (Seguridad del Tratamiento)

---

## 1. El Doble Rol de Redis en la Arquitectura

En este proyecto, Redis 7 cumple dos funciones primordiales:

```
                                 ┌────────────────────────────────────────────────────────┐
                                 │                   REDIS 7 (In-Memory)                  │
                                 ├────────────────────────────┬───────────────────────────┤
                                 │    A. ALTO RENDIMIENTO     │   B. SEGURIDAD & LOPDP    │
                                 │    (Cache-Aside Pattern)   │ (Rate Limit & Encrypted)  │
                                 ├────────────────────────────┼───────────────────────────┤
[ Handlers / Middlewares ] ────► │ • Catálogos geográficos    │ • Bloqueo de 3 intentos   │
           │                     │ • Período de postulación   │ • Token de reset cifrado  │
           │ (Cache Miss)        │ • Permisos RBAC por ruta   │ • Cédula pseudonimizada   │
           ▼                     └────────────────────────────┴───────────────────────────┘
   [ PostgreSQL 16 ]
```

---

## 2. Optimización y Reducción de Carga a PostgreSQL

Antes de esta optimización, cada petición HTTP entrante a una ruta protegida (`/api/postulacion/*`, `/api/postulante/*`) ejecutaba dos consultas a PostgreSQL:
1. `SELECT EXISTS(FROM permiso ...)` para el middleware RBAC.
2. `SELECT inicio, fin FROM periodo_postulacion ...` para validar si el concurso estaba abierto.

### Mejoras Implementadas:
1. **Caché de Permisos RBAC ([`backend/internal/cache/permiso.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/cache/permiso.go)):**
   - Clave: `rbac:permiso:{rol}:{metodo}:{ruta}` con TTL de 1 hora.
   - Si la regla está en Redis, la autorización toma **< 0.5 ms** en RAM, evitando golpear el pool de conexiones de PostgreSQL.
2. **Caché de Período ([`backend/internal/cache/periodo.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/cache/periodo.go)):**
   - Clave: `periodo:activo` con TTL de 1 hora.
   - Serialización JSON atómica con fallback transparente a base de datos.
3. **Catálogos Geográficos ([`backend/internal/cache/catalogo.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/cache/catalogo.go)):**
   - Cachean países, provincias y cantones con TTL de 24 horas y timeout de lectura de 200 ms.

---

## 3. Cifrado y Pseudonimización de Datos en Memoria (LOPDP y OWASP A02)

### A. Riesgo Identificado:
Redis guarda datos en memoria volátil y genera volcados periódicos en disco (`dump.rdb` y `appendonly.aof`). Si las cédulas o los tokens de reseteo se guardan en texto plano:
- Un atacante con acceso al servidor o a un volcado de memoria puede extraer las cédulas de los ciudadanos.
- El comando `KEYS *` o `SCAN` expondría la lista de postulantes.

### B. Solución Implementada ([`backend/internal/cache/crypto.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/cache/crypto.go)):

1. **Cifrado Simétrico Autenticado (AES-256-GCM):**
   - Se utiliza una clave de 256 bits derivada mediante SHA-256 a partir de `JWT_SECRET`.
   - En `GuardarResetToken`, la cédula se cifra con un nonce único aleatorio de 12 bytes (`crypto/rand`).
   - El payload almacenado en Redis tiene la estructura: `base64( [nonce (12B)] + [ciphertext] + [tag de autenticación (16B)] )`.
   - Al recuperar el token con `ObtenerYBorrarResetToken`, se descifra y autentica; cualquier manipulación de datos en memoria es detectada y rechazada.

2. **Pseudonimización Criptográfica de Claves (HMAC-SHA256):**
   - Para no exponer la cédula en los nombres de clave de Redis, se aplica una función resumen HMAC con secreto:
     ```
     intentos:login:1710034065  ──►  intentos:login:a3f81bc945e82b71...
     registro:verificado:1710034065 ──►  registro:verificado:c8192a54...
     ```
   - Quien inspeccione Redis ve únicamente cadenas seudoaleatorias sin posibilidad de correlacionarlas con cédulas de identidad reales sin poseer la clave maestra del servidor.

---

## 4. Pruebas Unitarias de Verificación

Se añadieron pruebas en [`backend/internal/cache/crypto_test.go`](file:///c:/Users/Gino/Proyectos/srpd-biess/backend/internal/cache/crypto_test.go) que verifican:
- Cifrado y descifrado fiel de datos con AES-256-GCM.
- Propiedad de no-determinismo (mismo dato cifrado dos veces produce ciphertexts distintos por el nonce aleatorio).
- Detección inmediata de manipulación de bytes (tampering / error de tag GCM).
- Determinismo seguro de la pseudonimización de claves.

