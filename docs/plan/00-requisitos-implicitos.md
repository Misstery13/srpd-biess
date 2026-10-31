# Requisitos implícitos y decisiones

Lo que el documento no dice pero el desarrollo necesita. Si alguien no está de acuerdo con una decisión, se discute **antes** de la Fase 1.

## Fuera de alcance

- **EF-13** (módulo de administración) y **EF-14** (banco de preguntas). El enunciado pide solo el "módulo inicial de registro y postulación".
- Firma manuscrita e impresión del formulario (1.10.2): se muestra el apartado, sin generar PDF.
- **RBAC completo** (tablas `opcion`, `endpoint`, `rol_opcion`, `opcion_endpoint`, menú dinámico y botones condicionales por permiso): solo como mejora de la Fase 4 si EF-12 está terminado. Con un único rol activo (postulante) no cambiaría nada visible. Ver R1–R4.

## Seguridad

| # | Decisión |
|---|----------|
| S1 | Contraseñas con **bcrypt**; nunca se guardan ni se devuelven en texto plano. |
| S2 | Sesión con **JWT** (expira en 2 h) en header `Authorization: Bearer`. El frontend lo guarda en memoria + `sessionStorage`. |
| S3 | Contraseña mínima: 8 caracteres, al menos 1 letra y 1 número. |
| S4 | Bloqueo de 3 intentos / 10 min aplica a **registro (verificación de identidad) y a login** (la rúbrica habla de "intentos de autenticación"). |
| S5 | El contador de intentos se **reinicia** al tener éxito. |
| S6 | El mensaje de login fallido es genérico ("sus credenciales de ingreso son incorrectas"), no revela si la cédula existe. |
| S7 | Un postulante solo accede a **su propia** postulación (se toma el id del JWT, nunca del body). |

## Validaciones

| # | Decisión |
|---|----------|
| V1 | Cédula ecuatoriana: 10 dígitos + **algoritmo módulo 10**. Se valida en frontend (feedback) y backend (regla). |
| V2 | Fecha de expedición en formato `DD/MM/AAAA` en la UI; `YYYY-MM-DD` en la API. |
| V3 | Código dactilar: 10 caracteres alfanuméricos, se compara en mayúsculas. |
| V4 | Email único entre postulantes **activos**, sin distinguir mayúsculas (índice parcial sobre `lower(email)`). |
| V5 | Nacionalidad ≠ ECUATORIANA → `403` y la UI cierra la postulación (no muestra el formulario). |
| V6 | Si el país ≠ Ecuador, provincia y ciudad son texto libre (ya contemplado en la tabla `domicilios`). |
| V7 | Experiencia: `fecha_fin ≥ fecha_inicio`; la duración (años, meses, días) se **calcula, no se guarda**. |
| V8 | Funciones de experiencia: máx. 1000 caracteres. |

## Datos y persistencia

| # | Decisión |
|---|----------|
| P1 | **Una postulación activa por postulante**: índice único parcial `postulaciones (postulante_id) WHERE deleted_at IS NULL` (ya en `init.sql`). |
| P2 | Código único con formato `BIESS-2026-000001` (a partir del id). |
| P3 | Estados: `BORRADOR` → `ENVIADA`. Una postulación `ENVIADA` ya no se edita. |
| P4 | Guardar = **una transacción** que reemplaza las secciones (domicilio, familiares, formación, experiencia…). Si una falla, `ROLLBACK` de todo. |
| P5 | Faltan tablas: `publicaciones`, `capacitaciones`, `inhabilidades`. |
| P6 | Los datos del Registro Civil se muestran **solo lectura**; no se copian a `postulaciones`. |

## Archivos PDF

| # | Decisión |
|---|----------|
| A1 | Se guardan en disco en `backend/uploads/{postulante_id}/{uuid}.pdf` (carpeta en `.gitignore`, **no** servida como estática); en BD solo la ruta. El nombre original del usuario nunca se usa como ruta (path traversal). |
| A2 | Se valida **tipo real** (los primeros bytes son `%PDF`), no solo la extensión. |
| A3 | Límites: 4 MB respaldo de organización, 2 MB los demás, **50 MB total por postulante**. |

## Integraciones simuladas

| # | Decisión |
|---|----------|
| I1 | **Registro Civil** e **Historial Laboral IESS** = tabla `registro_civil` (columnas `es_afiliado`, `es_jubilado`, `es_empleador`). |
| I2 | **Correo** (recuperar contraseña, enviar postulación): **Mailpit** en docker-compose (bandeja local en `:8025`). El log solo dice "correo enviado a postulación BIESS-2026-000001", nunca el contenido (ver L4). |
| I3 | Token de recuperación aleatorio (`crypto/rand`) en Redis con TTL de 15 min, **de un solo uso** (se borra al usarlo). |

## Redis

| Clave | Valor | TTL |
|-------|-------|-----|
| `intentos:registro:{cedula}` | contador (`INCR`) | 600 s desde el primer fallo |
| `intentos:login:{cedula}` | contador (`INCR`) | 600 s |
| `catalogo:paises` | JSON | 24 h |
| `catalogo:provincias:{pais_id}` | JSON | 24 h |
| `catalogo:cantones:{provincia_id}` | JSON | 24 h |
| `reset:{token}` | cédula | 900 s |

Si Redis no responde, los catálogos se leen de PostgreSQL. El bloqueo **no** se omite: si Redis falla, la autenticación devuelve `503`.

## Seguridad — OWASP Top 10

| # | Riesgo | Decisión |
|---|--------|----------|
| O1 | A01 Control de acceso | Todo recurso (postulación, PDF) se busca filtrando por el `postulante_id` del JWT. Pedir el de otro → `404`. |
| O2 | A02 Criptografía | `JWT_SECRET` de ≥ 32 caracteres aleatorios; el backend **no arranca** si falta o es el valor de ejemplo. |
| O3 | A03 Inyección | Solo consultas parametrizadas (`$1, $2`), nunca `fmt.Sprintf` con SQL. En React, prohibido `dangerouslySetInnerHTML`. |
| O4 | A04 Diseño inseguro | Límite de tamaño del body (5 MB) y de longitud de cada campo, igual que en la BD. |
| O5 | A05 Configuración | Cabeceras `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`; errores `500` sin detalles internos (el detalle va al log). |
| O6 | A06 Dependencias | `npm audit` y `govulncheck ./...` antes de la entrega. |
| O7 | A07 Autenticación | Cubierto por S1–S6. |
| O8 | A09 Logs | Se registran intentos fallidos y bloqueos con la cédula **enmascarada** (`17******65`). |

## Protección de datos — LOPDP (Ecuador, 2021)

| # | Principio | Decisión |
|---|-----------|----------|
| L1 | Consentimiento | Casillas de los 5 textos de 1.10.1, **sin marcar por defecto**; se guarda `acepta_terminos` + `fecha_consentimiento`. Sin consentimiento no se envía. |
| L2 | Minimización | Datos del Registro Civil se consultan, no se copian (P6). No se piden datos fuera del formulario EF-12. |
| L3 | Finalidad | Los datos solo se usan para el concurso; nada de analítica ni servicios externos (CDN de fuentes, trackers). |
| L4 | Seguridad | Logs sin datos personales (O8, I2); PDFs fuera de carpetas públicas (A1); contraseñas con bcrypt (S1). |
| L5 | Derechos del titular | Acceso y rectificación: el postulante ve y edita su formulario mientras está en `BORRADOR`. Supresión: `DELETE /api/postulante/me` aplica B4 (baja lógica + anonimización + borrado de PDFs y datos de terceros). |
| L6 | Datos de terceros | Cónyuge y familiares se registran sin su consentimiento directo. **Limitación conocida**, se menciona en la sustentación. |
| L7 | Datos de prueba | Todas las personas de `init.sql` son ficticias. Nunca usar cédulas reales en el repo. |

## Roles y permisos — RBAC mínimo

Acordado por el equipo (observaciones de Gino en PR #1, versión reducida).

| # | Decisión |
|---|----------|
| R1 | Tablas `rol` y `permiso` (`rol_id`, `metodo`, `ruta`). Roles semilla: `POSTULANTE` y `ADMIN` (este último sin pantallas: EF-13 fuera de alcance). |
| R2 | `postulantes.rol_id`; al registrarse se asigna `POSTULANTE`. El JWT lleva `sub` (id) y `rol` (código). |
| R3 | Middleware de Gin en las rutas protegidas: busca un permiso activo para (`rol`, `c.Request.Method`, `c.FullPath()`). Sin permiso → `403` "No tiene permisos para realizar esta acción". Se aplica **después** del middleware JWT. |
| R4 | Las rutas públicas (`/api/auth/*`, `/api/catalogos/*`, `/api/periodo`, `/api/health`) no pasan por RBAC. Toda ruta protegida nueva necesita su fila en `permiso`. |

## Borrado lógico (soft delete)

| # | Decisión |
|---|----------|
| B1 | Solo `deleted_at TIMESTAMPTZ` (NULL = activo). No se agrega un booleano `deleted`: duplicaría el dato y podría quedar inconsistente. |
| B2 | Aplica a `rol`, `permiso`, `periodo_postulacion`, `postulantes` y `postulaciones`. Toda consulta operativa filtra `WHERE deleted_at IS NULL`; "eliminar" es un `UPDATE ... SET deleted_at = now()`. |
| B3 | **No** aplica a las tablas hijas de la postulación (domicilio, familiares, formación, experiencia, publicaciones, capacitación, inhabilidades): se reemplazan dentro de la transacción de guardado (P4). Con soft delete acumularían filas muertas en cada guardado. |
| B4 | Supresión LOPDP de un postulante, **en una transacción**: `deleted_at = now()` en postulante y postulación; email → `eliminado-{id}@anonimo.invalid`; `password_hash` vacío; se borran físicamente las tablas hijas (incluye datos de terceros, L6) y los PDFs. Se conserva solo cédula, código y fechas para auditoría del concurso. |
| B5 | Unicidad solo entre activos con **índices únicos parciales** (`WHERE deleted_at IS NULL`): `rol.codigo`, `permiso (rol_id, metodo, ruta)`, `postulantes.cedula`, `lower(postulantes.email)`, `postulaciones.postulante_id`. El `codigo` de postulación es único global: nunca se reutiliza. |
| B6 | Reactivación: en **RBAC** se reactiva la fila existente (`deleted_at = NULL`) en vez de insertar otra. En **postulantes**, si alguien dado de baja vuelve a registrarse con su cédula, se crea un registro nuevo (el anterior quedó anonimizado y no se puede restaurar) y debe aceptar de nuevo el consentimiento (L1). |

## Período de postulación

Requisito del documento (EF-11 p.1, restricciones temporales de EF-1, EF-13 p.2) que no estaba en el plan.

| # | Decisión |
|---|----------|
| T1 | Tabla `periodo_postulacion` (`inicio`, `fin`); semilla: 01/10/2026 – 30/11/2026. Sin pantalla de administración (EF-13). |
| T2 | `GET /api/periodo` (pública) → `{abierto, inicio, fin}`. Fuera del período, `PUT /api/postulacion`, `documentos` y `enviar` devuelven `403` "El período de postulación no está habilitado". El login y la consulta siguen disponibles. |

## Manejo de errores

| # | Decisión |
|---|----------|
| E1 | **Backend:** los handlers devuelven errores tipados (`ErrValidacion`, `ErrNoAutorizado`, `ErrProhibido`, `ErrNoEncontrado`, `ErrConflicto`, `ErrBloqueado`); un middleware los traduce al código HTTP y a `{"error": "..."}`. Cualquier otro error → `500` con mensaje genérico y detalle solo en el log (O5). |
| E2 | Nunca se devuelven nombres de tablas, restricciones, SQL ni trazas. Una violación de índice único se traduce a `409` con mensaje para el usuario. |
| E3 | **Frontend:** el interceptor de axios convierte toda falla en un `Error` con mensaje listo para mostrar y su `status` (ya en `services/api.js`). `401` con sesión iniciada → cierra sesión y redirige a `/login`. |
| E4 | **Frontend:** `ErrorBoundary` en el layout para errores de render: muestra "Ocurrió un problema" con botón para reintentar, sin pantalla en blanco. |
| E5 | Cada formulario muestra el error junto al campo (validación) o arriba del formulario (servidor), con `role="alert"`. |

## API

- Errores con formato único: `{"error": "mensaje para el usuario"}`.
- Códigos: `400` validación, `401` credenciales o sesión vencida, `403` regla de negocio, permiso (R3) o período cerrado (T2), `404`, `409` duplicado, `429` bloqueo, `500`.
- CORS: en desarrollo no hace falta (proxy de Vite en `/api`).
