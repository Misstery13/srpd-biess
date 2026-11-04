# SRPD BIESS

Módulo de registro y postulación para el **Concurso Público de Méritos y Oposición del Directorio del BIESS** (Resolución C.D. 701).
Proyecto de la asignatura *Frameworks de Programación Web*. Especificación completa en [`docs/`](docs/).

Alcance: **EF-11** (registro y autenticación) y **EF-12** (formulario de postulación).

**Plan de implementación por fases:** [`docs/plan/`](docs/plan/README.md)

## Enlaces del proyecto

- **Repositorio de Git:** [https://github.com/Misstery13/srpd-biess](https://github.com/Misstery13/srpd-biess)
- **Despliegue en producción:** [https://postulacionesbiess.dev](https://postulacionesbiess.dev) (VPS EC2 de AWS)

## Arquitectura

| Capa | Tecnología |
|------|------------|
| Frontend | React + Vite + Tailwind CSS (SPA) |
| Backend | Go + Gin — Handler → Service → Repository con inyección de dependencias |
| Base de datos | PostgreSQL 16 |
| Caché / rate limiting | Redis 7 |

```
.
├── backend/
│   ├── cmd/api/            # main.go: arranque y cableado de dependencias
│   └── internal/
│       ├── config/         # lectura de variables de entorno
│       ├── handler/        # HTTP (controladores)
│       ├── service/        # reglas de negocio (EF-11, EF-12)
│       ├── repository/     # acceso a PostgreSQL
│       ├── cache/          # Redis: bloqueo de intentos y caché de catálogos
│       └── model/          # structs de dominio
├── frontend/
│   └── src/
│       ├── components/     # componentes reutilizables
│       ├── pages/          # Login, Registro, Postulación...
│       ├── services/       # llamadas a la API (axios)
│       ├── context/        # estado global (sesión)
│       └── routes/         # rutas protegidas
├── db/init.sql             # DDL + datos semilla (países, provincias, cantones, registro civil simulado)
├── docker-compose.yml      # PostgreSQL + Redis + Mailpit (desarrollo)
├── docker-compose.prod.yml # stack completo con Nginx (VPS)
├── .env.example
└── docs/                   # enunciado y requerimiento funcional
```

## Reglas de negocio clave

- **EF-11**: validación de cédula contra Registro Civil (simulado en la tabla `registro_civil`) con *fecha de expedición* o *código dactilar*. **3 intentos fallidos → bloqueo de 10 min en Redis.**
- **EF-12**: solo nacionalidad **ecuatoriana** puede postular. Cascada **País → Provincia → Cantón** (si el país no es Ecuador, provincia y ciudad son texto libre). Catálogos cacheados en Redis.
- Guardado de la postulación en **una transacción** (`BEGIN / COMMIT / ROLLBACK`) sobre varias tablas.

## Prerrequisitos

- **Go:** 1.26+ (definido en `backend/go.mod`)
- **Node.js:** 20+ (probado con Node 24 y npm 10+)
- **PostgreSQL:** 16+
- **Redis:** 7+
- **Docker y Docker Compose:** v2+ (recomendado para levantar PostgreSQL 16, Redis 7 y Mailpit sin instalaciones locales)

## Ejecución local

```bash
# 1. Variables de entorno
cp .env.example .env
# Cambiar JWT_SECRET: el backend no arranca con el valor de ejemplo (mínimo 32 caracteres)
#   openssl rand -hex 32

# 2. Base de datos, Redis y Mailpit (init.sql se ejecuta solo la primera vez)
docker compose up -d

# 3. Backend
cd backend
go mod tidy
go run ./cmd/api        # http://localhost:8080/api/health

# 4. Frontend (otra terminal)
cd frontend
npm install
npm run dev             # http://localhost:5173  (proxy /api -> APP_PORT)
```

- **Correos** (recuperación de contraseña y confirmación de envío): bandeja de Mailpit en http://localhost:8025.
- **Puertos ocupados:** si ya hay un PostgreSQL o un servidor web local en 5432 u 8080, cambiar `POSTGRES_PORT` y `APP_PORT` en `.env`. Docker Compose, el backend y el proxy de Vite leen ese mismo archivo.
- **Producción** (Nginx + backend + BD en contenedores): `docker compose -f docker-compose.prod.yml up -d --build`. En el VPS se usa `deploy.sh`.

### Pruebas

```bash
cd backend && go vet ./... && go test ./...
cd frontend && npm run lint && npm run build
# E2E con Chrome headless (ver el encabezado de cada script en frontend/scripts/)
node scripts/e2e-postulacion.mjs "<ruta a chrome>" http://localhost:5173   # con VITE_MOCK=true, o E2E_REAL=1 contra el backend
```

Para reiniciar la base desde cero: `docker compose down -v && docker compose up -d`.

### Variables de entorno (`.env.example`)

El archivo `.env.example` contiene las variables requeridas para desplegar el stack local y en producción:

| Variable | Valor por defecto | Descripción |
|---|---|---|
| `APP_PORT` | `8080` | Puerto HTTP en el que escucha el backend en Go |
| `APP_URL` | `http://localhost:5173` | URL pública de la aplicación (usada en enlaces de recuperación de contraseña) |
| `JWT_SECRET` | *(requerido)* | Clave secreta (≥ 32 caracteres) para firmar tokens JWT (`openssl rand -hex 32`) |
| `SMTP_HOST` | `localhost` / `mailpit` | Host del servidor SMTP para notificaciones y recuperación de clave |
| `SMTP_PORT` | `1025` | Puerto SMTP (Mailpit en desarrollo; 587 para un servidor real con STARTTLS) |
| `SMTP_USER` / `SMTP_PASSWORD` | *(vacío)* | Credenciales del servidor SMTP real; vacío = sin autenticación (Mailpit) |
| `SMTP_FROM` | `noreply@biess.gob.ec` | Remitente; con un servidor real debe ser una dirección autorizada por ese servidor |
| `POSTGRES_HOST` | `localhost` / `postgres` | Host del motor PostgreSQL |
| `POSTGRES_PORT` | `5432` | Puerto de conexión a PostgreSQL |
| `POSTGRES_DB` | `srpd_biess` | Nombre de la base de datos |
| `POSTGRES_USER` | `biess` | Usuario de PostgreSQL |
| `POSTGRES_PASSWORD` | `biess` | Contraseña de PostgreSQL |
| `REDIS_ADDR` | `localhost:6379` / `redis:6379` | Host y puerto del servidor Redis |
| `REDIS_PASSWORD` | *(vacío en local)* | Contraseña de autenticación de Redis |
| `LOGIN_MAX_ATTEMPTS` | `3` | Intentos fallidos antes de bloquear la cédula en Redis |
| `LOGIN_LOCK_MINUTES` | `10` | Duración del bloqueo en minutos tras exceder intentos |
| `CATALOG_CACHE_TTL_HOURS` | `24` | Tiempo de vida (TTL) del caché de catálogos en Redis |
| `VITE_API_URL` | `/api` | Prefijo de la API consumida por el frontend |
| `VITE_MOCK` | `false` | `true` activa simulación en memoria sin backend |

### Script DDL y datos semilla (`db/init.sql`)

El archivo [`db/init.sql`](db/init.sql) contiene el esquema relacional DDL completo y las semillas de datos:

1. **División política territorial (Catálogos normalizados):**
   - Tablas `paises`, `provincias` y `cantones` con foreign keys compuestas e integridad referencial.
   - Datos semilla con el registro oficial de Ecuador (sus 24 provincias y cantones) y países extranjeros para postulantes no residentes.
2. **Interoperabilidad simulada (Registro Civil e Historial IESS):**
   - Tabla `registro_civil` con registros de prueba simulando datos de identidad (cédula, nombres, código dactilar, fecha de expedición, estado civil, cónyuge) y condiciones laborales del IESS (`es_afiliado`, `es_jubilado`, `es_empleador`).
3. **Módulo de Postulación transaccional (EF-12):**
   - Tablas de postulaciones: `postulantes`, `periodos`, `postulaciones`, `domicilios`, `familiares`, `formacion_academica`, `experiencias`, `publicaciones`, `capacitaciones` e `inhabilidades`.
   - Soporte para transacciones atómicas `BEGIN / COMMIT / ROLLBACK` en cada guardado.
4. **Seguridad y Control de Acceso (RBAC):**
   - Tablas `rol`, `permiso` y `rol_permiso` con soporte de borrado lógico (`deleted_at`) e índices únicos parciales.

### Usuarios de prueba (Registro Civil simulado)

| Cédula | Fecha expedición | Código dactilar | Caso |
|--------|------------------|-----------------|------|
| 1710034065 | 15/03/2020 | V4443V4444 | Afiliado, casado |
| 0912345675 | 02/11/2019 | E1234I5678 | Jubilado |
| 0102030400 | 08/07/2021 | A1111B2222 | Empleador |
| 1754321980 | 10/01/2022 | C3333D4444 | Colombiano → postulación bloqueada |

## Flujo de trabajo en equipo

- `main`: versión estable para la entrega. No se hace push directo.
- `develop`: integración.
- Ramas de trabajo desde `develop`: `feature/ef11-login`, `feature/ef12-formulario`, `fix/...`
- Se integra con Pull Request hacia `develop`; al final `develop` → `main`.

Commits cortos y en presente: `feat: bloqueo de intentos en redis`, `fix: cascada de cantones`.
