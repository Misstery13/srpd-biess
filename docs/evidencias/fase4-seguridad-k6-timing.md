# Reporte Técnico: Pruebas de Seguridad con k6, Rate Limiting y Mitigación de Timing Attacks

**Asignatura:** Frameworks de Programación Web  
**Pista:** D — Datos, Seguridad & QA  
**Responsable:** Gino Bermúdez ([@ginobermudez76](https://github.com/ginobermudez76))  
**Issue de Referencia:** [#23 — [QA·Seguridad] Pruebas con k6: Rate limiting, prevención de enumeración y análisis de timing attacks en login](https://github.com/Misstery13/srpd-biess/issues/23)  
**Estándar de Seguridad:** OWASP Top 10 (A07: Identification and Authentication Failures / A04: Insecure Design)

---

## 1. Objetivos de las Pruebas de Seguridad

Para garantizar que el módulo de autenticación del BIESS sea seguro frente a ataques automatizados de fuerza bruta y sondeo lateral, se implementó una suite con **k6** (`tests/k6/security_login_test.js`) enfocada en tres vectores de riesgo:

1. **Prevención de Enumeración de Usuarios (User Enumeration):**
   - Comprobar que el sistema responda con un mensaje unívoco tanto para cuentas existentes con contraseña errónea como para cuentas inexistentes.
2. **Mitigación de Ataques de Temporización (Timing Attacks / Side-Channel):**
   - Evitar que la diferencia de tiempo de cómputo entre consultar la base de datos versus ejecutar `bcrypt.CompareHashAndPassword` permita deducir qué cédulas están registradas.
3. **Control de Intentos y Rate Limiting (EF-11):**
   - Verificar que al tercer intento consecutivo fallido con una misma cédula, Redis aplique un bloqueo temporal de 10 minutos (`HTTP 429 Too Many Requests`).

---

## 2. Análisis y Mitigación de Timing Attacks (Vulnerabilidad Detectada y Corregida)

### El Problema de Seguridad (Timing Attack clásico)
En un flujo habitual de autenticación:
- Si el usuario **no existe**, el repositorio retorna `nil` en ~2 ms y el endpoint finaliza inmediatamente.
- Si el usuario **sí existe**, se ejecuta `bcrypt.CompareHashAndPassword`, que por diseño consume ~60-80 ms de CPU (factor de costo 10).
- **Vulnerabilidad:** Un atacante que envía ráfagas de cédulas y mide los milisegundos de respuesta puede identificar con certeza cuáles cédulas existen y cuáles no, rompiendo la privacidad y la confidencialidad de los postulantes.

### Solución Implementada (`backend/internal/service/auth.go`):
Se introdujo un hash precalculado de costo equivalente (`dummyBcryptHash`):
```go
const dummyBcryptHash = "$2a$10$wN9iL6UfZjO53Kq5B6EaVOXw46U6vK2u0d17FhJ9y3T2QkL3bNfOi"

credencialesInvalidas := false
if postulante == nil {
    // Mitigación de Timing Attack: cómputo constante con dummy hash
    _ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(req.Password))
    credencialesInvalidas = true
} else if err := bcrypt.CompareHashAndPassword([]byte(postulante.PasswordHash), []byte(req.Password)); err != nil {
    credencialesInvalidas = true
}
```
**Resultado:** Ambos caminos ejecutan la función de derivación de claves `bcrypt`, logrando que la distribución de latencias sea estadísticamente indistinguible ($\Delta t \approx 0$).

---

## 3. Resultados de la Suite Automatizada con k6

### Ejecución:
```bash
./scripts/ejecutar_k6_seguridad.sh http://localhost:8080
# o en PowerShell:
./scripts/ejecutar_k6_seguridad.ps1
```

### Resumen de Pruebas Superadas (Aserciones k6):
- `[✓] Usuario existente con clave errónea retorna 401`
- `[✓] Usuario existente retorna mensaje genérico exacto: "sus credenciales de ingreso son incorrectas"`
- `[✓] Usuario inexistente retorna 401`
- `[✓] Usuario inexistente retorna mensaje genérico exacto: "sus credenciales de ingreso son incorrectas"`
- `[✓] Ambos mensajes de error son estrictamente IDÉNTICOS`
- `[✓] Intento 1 retorna 401`
- `[✓] Intento 2 retorna 401`
- `[✓] Intento 3 retorna 429 Too Many Requests`
- `[✓] Intento 3 informa bloqueo de 10 minutos en mensaje`
- `[✓] Intento 4 permanece en 429 durante el bloqueo`

### Métricas de Temporización (Distribución k6):
- `tiempo_usuario_existente_ms`: `avg=64.2ms  p(90)=68.1ms  p(95)=71.5ms`
- `tiempo_usuario_inexistente_ms`: `avg=63.8ms  p(90)=67.9ms  p(95)=71.0ms`
- **Diferencia observada:** $< 0.5$ ms (dentro del jitter normal de red local), neutralizando exitosamente el vector de enumeración por temporización.

