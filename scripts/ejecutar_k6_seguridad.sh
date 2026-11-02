#!/usr/bin/env bash
# ==============================================================================
# Script para Ejecución de Pruebas de Seguridad k6 (Suite Completa)
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

BASE_URL="${1:-http://localhost:8080}"
SUITE="${2:-all}"

echo "======================================================================"
echo "Iniciando suite de pruebas de seguridad k6 contra: $BASE_URL"
echo "======================================================================"

if ! command -v k6 &> /dev/null; then
    echo "Error: k6 no está instalado en el sistema."
    exit 1
fi

export BASE_URL="$BASE_URL"

if [ "$SUITE" = "all" ] || [ "$SUITE" = "endpoints" ]; then
    echo -e "\n[1/2] Ejecutando pruebas de Broken Access Control y Endpoints Protegidos..."
    k6 run tests/k6/security_endpoints_test.js
fi

if [ "$SUITE" = "all" ] || [ "$SUITE" = "login" ]; then
    echo -e "\n[2/2] Ejecutando pruebas de Rate Limiting, Login y Timing Attacks..."
    k6 run tests/k6/security_login_test.js
fi

echo -e "\nPruebas de seguridad completadas."
