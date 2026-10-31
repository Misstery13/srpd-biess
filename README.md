# SRPD BIESS

Módulo de registro y postulación para el **Concurso Público de Méritos y Oposición del Directorio del BIESS** (Resolución C.D. 701).
Proyecto de la asignatura *Frameworks de Programación Web*. Especificación completa en [`docs/`](docs/).

Alcance: **EF-11** (registro y autenticación) y **EF-12** (formulario de postulación).

**Plan de implementación por fases:** [`docs/plan/`](docs/plan/README.md)

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
├── docker-compose.yml      # PostgreSQL + Redis
├── .env.example
└── docs/                   # enunciado y requerimiento funcional
```

## Reglas de negocio clave

- **EF-11**: validación de cédula contra Registro Civil (simulado en la tabla `registro_civil`) con *fecha de expedición* o *código dactilar*. **3 intentos fallidos → bloqueo de 10 min en Redis.**
- **EF-12**: solo nacionalidad **ecuatoriana** puede postular. Cascada **País → Provincia → Cantón** (si el país no es Ecuador, provincia y ciudad son texto libre). Catálogos cacheados en Redis.
- Guardado de la postulación en **una transacción** (`BEGIN / COMMIT / ROLLBACK`) sobre varias tablas.

## Prerrequisitos

- Go 1.23+
- Node.js 20+ (probado con 24)
- Docker + Docker Compose (para PostgreSQL 16 y Redis 7), o instalarlos localmente

## Ejecución local

```bash
# 1. Variables de entorno
cp .env.example .env

# 2. Base de datos y Redis (init.sql se ejecuta solo la primera vez)
docker compose up -d

# 3. Backend
cd backend
go mod tidy
go run ./cmd/api        # http://localhost:8080/api/health

# 4. Frontend (otra terminal)
cd frontend
npm install
npm run dev             # http://localhost:5173  (proxy /api -> :8080)
```

Para reiniciar la base desde cero: `docker compose down -v && docker compose up -d`.

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
