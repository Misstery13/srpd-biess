# ==============================================================================
# Script de Demostracion y Validacion de ROLLBACK Transaccional (PowerShell)
# Concurso Publico BIESS 2026 - Pista D (Datos): Gino Bermudez
# Valida la atomicidad de GuardarTransaccional (Issue #10) con PostgreSQL real
# ==============================================================================

Write-Host "======================================================================"
Write-Host " SRPD BIESS - Validacion de Atomicidad y ROLLBACK Transaccional (Go)"
Write-Host "======================================================================"

# 1. Asegurar contenedor de PostgreSQL activo y listo para conexiones
Write-Host "[1/3] Asegurando servicio PostgreSQL..."
docker compose up -d postgres

Write-Host "Esperando que PostgreSQL este completamente listo para aceptar conexiones..."
$ready = $false
$retries = 30
while (-not $ready -and $retries -gt 0) {
    docker compose exec -T postgres pg_isready -U biess -d srpd_biess *>$null
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        Write-Host "PostgreSQL listo y aceptando conexiones."
    } else {
        Start-Sleep -Seconds 1
        $retries--
    }
}

if (-not $ready) {
    Write-Error "ERROR: PostgreSQL no estuvo listo a tiempo."
    exit 1
}

# 2. Ejecutar prueba con go en host o contenedor Docker si no hay go instalado
Write-Host "[2/3] Ejecutando TestPostulacionRepository_TransaccionalYRollback..."

$hasGo = Get-Command go -ErrorAction SilentlyContinue
$testOutput = @()
$testExit = 0

if ($hasGo) {
    Write-Host "Ejecutando con binario 'go' local..."
    $env:POSTGRES_HOST = "localhost"
    $env:POSTGRES_PORT = "5432"
    $env:POSTGRES_USER = "biess"
    $env:POSTGRES_PASSWORD = "biess"
    $env:POSTGRES_DB = "srpd_biess"

    Push-Location "backend"
    $testOutput = go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository 2>&1
    $testExit = $LASTEXITCODE
    Pop-Location
} else {
    Write-Host "'go' no detectado en el host. Ejecutando mediante contenedor Docker golang:1.24-alpine..."
    $postgresId = (docker compose ps -q postgres).Trim()
    $network = (docker inspect -f "{{range `$k, `$v := .NetworkSettings.Networks}}{{`$k}}{{end}}" $postgresId).Trim()
    $currentDir = (Get-Location).Path
    $backendPath = Join-Path $currentDir "backend"

    $dockerCmd = "docker run --rm --network $network -e POSTGRES_HOST=postgres -e POSTGRES_PORT=5432 -e POSTGRES_USER=biess -e POSTGRES_PASSWORD=biess -e POSTGRES_DB=srpd_biess -v `"${backendPath}:/src`" -w /src golang:1.24-alpine go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository 2>&1"
    $testOutput = Invoke-Expression $dockerCmd
    $testExit = $LASTEXITCODE
}

$testOutputText = $testOutput -join "`n"
Write-Host $testOutputText

# 3. Validar resultado: NO permitir falso positivo por SKIP o fallo
if ($testExit -ne 0) {
    Write-Error "ERROR: La prueba de ROLLBACK transaccional FALLO (codigo $testExit)."
    exit 1
}

if ($testOutputText -match "--- SKIP:") {
    Write-Error "ERROR: La prueba fue OMITIDA (SKIP). No se pudo validar la transaccion real con la BD."
    exit 1
}

if ($testOutputText -notmatch "--- PASS: TestPostulacionRepository_TransaccionalYRollback") {
    Write-Error "ERROR: La prueba no paso con PASS exitoso."
    exit 1
}

Write-Host "======================================================================"
Write-Host "[3/3] EXITO: Transaccion probada contra el repositorio real de Go."
Write-Host "Se verifico que ante error de restriccion (CHECK horas > 0), PostgreSQL ejecuto ROLLBACK completo"
Write-Host "y no quedaron registros huerfanos en ninguna tabla hija."
Write-Host "======================================================================"
