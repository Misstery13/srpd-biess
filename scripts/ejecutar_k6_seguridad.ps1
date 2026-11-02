# ==============================================================================
# Script para Ejecución de Pruebas de Seguridad k6 (Suite Completa)
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Suite = "all" # 'all', 'login', 'endpoints'
)

Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "Iniciando suite de pruebas de seguridad k6 contra: $BaseUrl" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
    Write-Host "Error: k6 no está instalado o no está en el PATH." -ForegroundColor Red
    exit 1
}

$env:BASE_URL = $BaseUrl

if ($Suite -eq "all" -or $Suite -eq "endpoints") {
    Write-Host "`n[1/2] Ejecutando pruebas de Broken Access Control y Endpoints Protegidos..." -ForegroundColor Yellow
    k6 run tests/k6/security_endpoints_test.js
}

if ($Suite -eq "all" -or $Suite -eq "login") {
    Write-Host "`n[2/2] Ejecutando pruebas de Rate Limiting, Login y Timing Attacks..." -ForegroundColor Yellow
    k6 run tests/k6/security_login_test.js
}

Write-Host "`nPruebas de seguridad completadas." -ForegroundColor Green
