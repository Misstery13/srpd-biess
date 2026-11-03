#!/usr/bin/env bash
# ==============================================================================
# Script de Demostración y Validación de ROLLBACK Transaccional
# Concurso Público BIESS 2026 — Pista D (Datos): Gino Bermúdez
# Valida la atomicidad de GuardarTransaccional (Issue #10) con PostgreSQL real
# ==============================================================================

set -euo pipefail
# Git Bash (Windows) convierte rutas como /src en C:/Program Files/Git/src; esto lo evita
export MSYS_NO_PATHCONV=1

echo "======================================================================"
echo " SRPD BIESS — Validación de Atomicidad y ROLLBACK Transaccional (Go)"
echo "======================================================================"

# 1. Asegurar contenedor de PostgreSQL activo y listo para conexiones
echo "[1/3] Asegurando servicio PostgreSQL..."
docker compose up -d postgres

echo "Esperando que PostgreSQL este completamente listo para aceptar conexiones..."
RETRIES=30
while [ $RETRIES -gt 0 ]; do
    if docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-biess}" -d "${POSTGRES_DB:-srpd_biess}" > /dev/null 2>&1; then
        echo "PostgreSQL listo y aceptando conexiones."
        break
    fi
    RETRIES=$((RETRIES - 1))
    sleep 1
done

if [ $RETRIES -eq 0 ]; then
    echo "ERROR: PostgreSQL no estuvo listo a tiempo."
    exit 1
fi

# 2. Ejecutar prueba con go en host o contenedor Docker si no hay go instalado
echo "[2/3] Ejecutando TestPostulacionRepository_TransaccionalYRollback..."

TEST_OUTPUT=""
TEST_EXIT=0

if command -v go >/dev/null 2>&1; then
    echo "Ejecutando con binario 'go' local..."
    export POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
    export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
    export POSTGRES_USER="${POSTGRES_USER:-biess}"
    export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-biess}"
    export POSTGRES_DB="${POSTGRES_DB:-srpd_biess}"
    set +e
    TEST_OUTPUT=$(cd backend && go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository 2>&1)
    TEST_EXIT=$?
    set -e
else
    echo "'go' no detectado en el host. Ejecutando mediante contenedor Docker golang:1.26-alpine..."
    PG_CONTAINER_ID=$(docker compose ps -q postgres)
    DOCKER_NET=$(docker inspect -f '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}' "$PG_CONTAINER_ID")
    REPO_DIR="$(pwd)"
    set +e
    TEST_OUTPUT=$(docker run --rm \
        --network "$DOCKER_NET" \
        -e POSTGRES_HOST=postgres \
        -e POSTGRES_PORT=5432 \
        -e POSTGRES_USER="${POSTGRES_USER:-biess}" \
        -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-biess}" \
        -e POSTGRES_DB="${POSTGRES_DB:-srpd_biess}" \
        -v "$REPO_DIR/backend:/src" \
        -w /src \
        golang:1.26-alpine \
        go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository 2>&1)
    TEST_EXIT=$?
    set -e
fi

echo "$TEST_OUTPUT"

# 3. Validar resultado: NO permitir falso positivo por SKIP o fallo
if [ $TEST_EXIT -ne 0 ]; then
    echo "ERROR: La prueba de ROLLBACK transaccional FALLÓ (código de salida $TEST_EXIT)."
    exit 1
fi

if echo "$TEST_OUTPUT" | grep -q -- "--- SKIP:"; then
    echo "ERROR: La prueba fue OMITIDA (SKIP). No se pudo validar la transacción real con la BD."
    exit 1
fi

if ! echo "$TEST_OUTPUT" | grep -q -- "--- PASS: TestPostulacionRepository_TransaccionalYRollback"; then
    echo "ERROR: La prueba no pasó con PASS exitoso."
    exit 1
fi

echo "======================================================================"
echo "[3/3] ÉXITO: Transacción probada contra el repositorio real de Go."
echo "Se verificó que ante error de restricción (CHECK horas > 0), PostgreSQL ejecutó ROLLBACK completo"
echo "y no quedaron registros huérfanos en ninguna tabla hija."
echo "======================================================================"
