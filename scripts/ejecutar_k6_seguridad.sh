#!/usr/bin/env bash
# ==============================================================================
# Script para Ejecución de Pruebas de Rendimiento y Seguridad k6
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
SUITE="${2:-all}"

echo "======================================================================"
echo "Iniciando suite de pruebas con k6 contra: $BASE_URL"
echo "Modalidad: $SUITE"
echo "======================================================================"

if ! command -v k6 &> /dev/null; then
    echo "Error: k6 no está instalado en el sistema." >&2
    exit 1
fi

export BASE_URL="$BASE_URL"

if [[ "$SUITE" == "all" || "$SUITE" == "security" || "$SUITE" == "endpoints" ]]; then
    echo -e "\n[1/3] Pruebas de Broken Access Control y Endpoints Protegidos..."
    k6 run tests/k6/security_endpoints_test.js
fi

if [[ "$SUITE" == "all" || "$SUITE" == "security" || "$SUITE" == "login" ]]; then
    echo -e "\n[2/3] Pruebas de Rate Limiting, Login y Timing Attacks..."
    k6 run tests/k6/security_login_test.js
fi

if [[ "$SUITE" == "all" || "$SUITE" == "load" ]]; then
    echo -e "\n[3/3] Pruebas de Carga Masiva y Concurrencia (pgxpool + Redis)..."
    k6 run tests/k6/load_concurrency_test.js
fi

echo -e "\nPruebas k6 finalizadas exitosamente."
