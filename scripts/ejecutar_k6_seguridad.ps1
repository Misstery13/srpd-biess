# ==============================================================================
# Script para Ejecución de Pruebas de Rendimiento y Seguridad k6
# Concurso Público BIESS 2026 — Pista D (Datos & QA): Gino Bermúdez
# ==============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Suite = "all" # 'all', 'load', 'security', 'login', 'endpoints'
)

Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "Iniciando suite de pruebas con k6 contra: $BaseUrl" -ForegroundColor Cyan
Write-Host "Modalidad: $Suite" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan

if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
    Write-Host "Error: k6 no está instalado o no está en el PATH." -ForegroundColor Red
    exit 1
}

$env:BASE_URL = $BaseUrl

if ($Suite -eq "all" -or $Suite -eq "security" -or $Suite -eq "endpoints") {
    Write-Host "`n[1/3] Pruebas de Broken Access Control y Endpoints Protegidos..." -ForegroundColor Yellow
    k6 run tests/k6/security_endpoints_test.js
}

if ($Suite -eq "all" -or $Suite -eq "security" -or $Suite -eq "login") {
    Write-Host "`n[2/3] Pruebas de Rate Limiting, Login y Timing Attacks..." -ForegroundColor Yellow
    k6 run tests/k6/security_login_test.js
}

if ($Suite -eq "all" -or $Suite -eq "load") {
    Write-Host "`n[3/3] Pruebas de Carga Masiva y Concurrencia (pgxpool + Redis)..." -ForegroundColor Yellow
    k6 run tests/k6/load_concurrency_test.js
}

Write-Host "`nPruebas k6 finalizadas exitosamente." -ForegroundColor Green
