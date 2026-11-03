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
           │                     │                            │ • Token de reset cifrado  │
           │ (Cache Miss)        │                            │ • Cédula pseudonimizada   │
           ▼                     └────────────────────────────┴───────────────────────────┘
   [ PostgreSQL 16 ]
```

---

## 2. Optimización y Reducción de Carga a PostgreSQL

**Catálogos geográficos ([`backend/internal/cache/catalogo.go`](../../backend/internal/cache/catalogo.go)):** países, provincias y cantones se cachean con TTL de 24 horas y timeout de lectura de 200 ms. Son datos que no cambian durante el concurso y se consultan en cada formulario.

### Decisión: período y permisos RBAC se consultan siempre en PostgreSQL
Se probó cachear el período de postulación y los permisos RBAC con TTL de 1 hora, y se descartó:
- **Período (T2):** al cerrar el concurso en la BD, la API seguía respondiendo `abierto: true` hasta que expiraba el caché, y los postulantes podían seguir guardando. El E2E contra el backend real lo detectó (51/53).
- **Permisos (R3):** si se revoca un permiso en la tabla `permiso`, seguía vigente hasta 1 hora.

Las dos son consultas de una sola fila por índice (menos de 1 ms). Cachearlas casi no ahorra tiempo y cambia reglas de negocio que deben aplicarse al instante. Con las dos consultas directas, la prueba k6 de carga sostiene unas 1 290 peticiones por segundo con p95 = 19 ms y 0 % de errores ([reporte](fase4-rendimiento-k6-concurrencia.md)).

---

## 3. Cifrado y Pseudonimización de Datos en Memoria (LOPDP y OWASP A02)

### A. Riesgo Identificado:
Redis guarda datos en memoria volátil y genera volcados periódicos en disco (`dump.rdb` y `appendonly.aof`). Si las cédulas o los tokens de reseteo se guardan en texto plano:
- Un atacante con acceso al servidor o a un volcado de memoria puede extraer las cédulas de los ciudadanos.
- El comando `KEYS *` o `SCAN` expondría la lista de postulantes.

### B. Solución Implementada ([`backend/internal/cache/crypto.go`](../../backend/internal/cache/crypto.go)):

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

Se añadieron pruebas en [`backend/internal/cache/crypto_test.go`](../../backend/internal/cache/crypto_test.go) que verifican:
- Cifrado y descifrado fiel de datos con AES-256-GCM.
- Propiedad de no-determinismo (mismo dato cifrado dos veces produce ciphertexts distintos por el nonce aleatorio).
- Detección inmediata de manipulación de bytes (tampering / error de tag GCM).
- Determinismo seguro de la pseudonimización de claves.

