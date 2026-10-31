# Requisitos implícitos y decisiones

Lo que el documento no dice pero el desarrollo necesita. Si alguien no está de acuerdo con una decisión, se discute **antes** de la Fase 1.

## Fuera de alcance

- **EF-13** (módulo de administración) y **EF-14** (banco de preguntas). El enunciado pide solo el "módulo inicial de registro y postulación".
- Firma manuscrita e impresión del formulario (1.10.2): se muestra el apartado, sin generar PDF.

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
| V4 | Email único por postulante. |
| V5 | Nacionalidad ≠ ECUATORIANA → `403` y la UI cierra la postulación (no muestra el formulario). |
| V6 | Si el país ≠ Ecuador, provincia y ciudad son texto libre (ya contemplado en la tabla `domicilios`). |
| V7 | Experiencia: `fecha_fin ≥ fecha_inicio`; la duración (años, meses, días) se **calcula, no se guarda**. |
| V8 | Funciones de experiencia: máx. 1000 caracteres. |

## Datos y persistencia

| # | Decisión |
|---|----------|
| P1 | **Una postulación por postulante**: falta `UNIQUE (postulante_id)` en `postulaciones`. |
| P2 | Código único con formato `BIESS-2026-000001` (a partir del id). |
| P3 | Estados: `BORRADOR` → `ENVIADA`. Una postulación `ENVIADA` ya no se edita. |
| P4 | Guardar = **una transacción** que reemplaza las secciones (domicilio, familiares, formación, experiencia…). Si una falla, `ROLLBACK` de todo. |
| P5 | Faltan tablas: `publicaciones`, `capacitaciones`, `inhabilidades`. |
| P6 | Los datos del Registro Civil se muestran **solo lectura**; no se copian a `postulaciones`. |

## Archivos PDF

| # | Decisión |
|---|----------|
| A1 | Se guardan en disco en `backend/uploads/{cedula}/` (carpeta en `.gitignore`); en BD solo la ruta. |
| A2 | Se valida **tipo real** (los primeros bytes son `%PDF`), no solo la extensión. |
| A3 | Límites: 4 MB respaldo de organización, 2 MB los demás, **50 MB total por postulante**. |

## Integraciones simuladas

| # | Decisión |
|---|----------|
| I1 | **Registro Civil** e **Historial Laboral IESS** = tabla `registro_civil` (columnas `es_afiliado`, `es_jubilado`, `es_empleador`). |
| I2 | **Correo** (recuperar contraseña, enviar postulación): se escribe en el log del backend. Opcional: Mailpit en docker-compose. |
| I3 | Token de recuperación de contraseña en Redis con TTL de 15 min. |

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

## API

- Errores con formato único: `{"error": "mensaje para el usuario"}`.
- Códigos: `400` validación, `401` credenciales, `403` regla de negocio, `404`, `409` duplicado, `429` bloqueo, `500`.
- CORS: en desarrollo no hace falta (proxy de Vite en `/api`).
