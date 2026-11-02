#!/usr/bin/env bash
# ==============================================================================
# Script para Ejecución de Pruebas de Seguridad k6 (Login & Rate Limiting)
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

BASE_URL="${1:-http://localhost:8080}"

echo "======================================================================"
echo "Iniciando suite de pruebas de seguridad k6 contra: $BASE_URL"
echo "======================================================================"

if ! command -v k6 &> /dev/null; then
    echo "Error: k6 no está instalado en el sistema."
    exit 1
fi

export BASE_URL="$BASE_URL"
k6 run tests/k6/security_login_test.js
