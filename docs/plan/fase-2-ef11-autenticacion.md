# Fase 2 — EF-11 Registro y autenticación

**Fecha:** sábado 31 – domingo 1 · **Hito H2:** con la cédula `1710034065` se verifica, registra e inicia sesión; 3 intentos fallidos muestran el bloqueo de 10 min; sin token, `/postulacion` redirige a `/login`.

## D — Datos
- [ ] `PostulanteRepository`: crear, buscar por cédula.
- [ ] `RegistroCivilRepository`: buscar por cédula.
- [ ] `IntentosCache`: `Incrementar(clave)`, `EstaBloqueado(clave)`, `Reiniciar(clave)` con TTL de 600 s (S4, S5).
- [ ] Token de recuperación en Redis con TTL de 15 min (I3).

## B — Backend
- [ ] `AuthService.Verificar`: valida cédula (V1), compara fecha de expedición **o** código dactilar, aplica el bloqueo.
- [ ] `AuthService.Registrar`: bcrypt (S1), contraseña válida (S3), email único (V4).
- [ ] `AuthService.Login` → JWT de 2 h (S2), con bloqueo.
- [ ] Recuperar / restablecer contraseña (correo simulado en el log, I2).
- [ ] Middleware JWT para las rutas protegidas.

## F — Frontend
- [ ] Pantalla inicial: "Iniciar sesión" / "Registrarse".
- [ ] Registro en 2 pasos: (1) cédula + fecha de expedición **o** código dactilar → (2) email + contraseña.
- [ ] Login con link "¿Olvidaste tu contraseña?".
- [ ] Mensajes exactos del documento para 401 y 429; el botón se deshabilita mientras dura el bloqueo.
- [ ] `AuthContext` + `RutaProtegida`.
- [ ] Al primer ingreso, redirige directo a `/postulacion` (EF-12).

## Criterio de aceptación
- [ ] Los 3 mensajes del documento aparecen tal cual.
- [ ] Después de 10 min (o de borrar la clave en Redis) se puede volver a intentar.
