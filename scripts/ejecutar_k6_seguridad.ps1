# ==============================================================================
# Script para Ejecución de Pruebas de Seguridad k6 (Login & Rate Limiting)
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "Iniciando suite de pruebas de seguridad k6 contra: $BaseUrl" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
    Write-Host "Error: k6 no está instalado o no está en el PATH." -ForegroundColor Red
    exit 1
}

$env:BASE_URL = $BaseUrl
k6 run tests/k6/security_login_test.js
