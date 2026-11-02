#!/usr/bin/env bash
# ==============================================================================
# Script de Demostración y Validación de ROLLBACK Transaccional
# Concurso Público BIESS 2026 — Pista D (Datos): Gino Bermúdez
# Valida la atomicidad de GuardarTransaccional (Issue #10) con PostgreSQL real
# ==============================================================================

set -euo pipefail

echo "======================================================================"
echo " SRPD BIESS — Validación de Atomicidad y ROLLBACK Transaccional (Go)"
echo "======================================================================"

# 1. Asegurar contenedor de PostgreSQL activo
if ! docker ps --filter "name=postgres" --filter "status=running" --format "{{.Names}}" | grep -q "postgres"; then
    echo "[1/3] Contenedor PostgreSQL no activo. Levantando con docker compose..."
    docker compose up -d postgres
    sleep 3
else
    echo "[1/3] Contenedor PostgreSQL detectado y activo."
fi

# 2. Configurar variables de conexión
export POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
export POSTGRES_PORT="${POSTGRES_PORT:-5432}"
export POSTGRES_USER="${POSTGRES_USER:-biess}"
export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-biess}"
export POSTGRES_DB="${POSTGRES_DB:-srpd_biess}"

echo "[2/3] Ejecutando TestPostulacionRepository_TransaccionalYRollback..."

(cd backend && go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository)

echo "[3/3] ÉXITO: Transacción probada contra el repositorio real de Go."
echo "Se verificó que ante error de restricción (CHECK horas > 0), PostgreSQL ejecutó ROLLBACK completo"
echo "y no quedaron registros huérfanos en ninguna tabla hija."
