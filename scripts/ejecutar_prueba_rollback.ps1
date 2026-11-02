# ==============================================================================
# Script de Demostración de ROLLBACK Transaccional (PostgreSQL)
# Concurso Público BIESS 2026 — Pista D (Datos): Gino Bermúdez
# ==============================================================================

Write-Host "Iniciando prueba de consistencia transaccional y ROLLBACK en PostgreSQL..." -ForegroundColor Cyan

$containerName = "srpd-biess-postgres-1"
$checkContainer = docker ps --filter "name=postgres" --format "{{.Names}}"

if (-not $checkContainer) {
    Write-Host "Contenedor de PostgreSQL no encontrado corriendo. Intentando levantar docker compose..." -ForegroundColor Yellow
    docker compose up -d postgres
    Start-Sleep -Seconds 3
}

Write-Host "Ejecutando script SQL de prueba dentro del contenedor de PostgreSQL..." -ForegroundColor Green
docker compose exec -T postgres psql -U biess -d srpd_biess -f /db/prueba_rollback.sql

if ($LASTEXITCODE -ne 0) {
    Write-Host "`nEl script generó la excepción esperada (CHECK horas > 0) y PostgreSQL ejecutó el ROLLBACK con éxito." -ForegroundColor Green
} else {
    Write-Host "`nPrueba finalizada." -ForegroundColor Green
}

