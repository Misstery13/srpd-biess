Write-Host "SRPD BIESS - Validacion de Atomicidad y ROLLBACK Transaccional (Go)"

$postgresRunning = docker ps --filter "name=postgres" --filter "status=running" --format "{{.Names}}"
if (-not $postgresRunning) {
    Write-Host "Contenedor PostgreSQL inactivo. Levantando..."
    docker compose up -d postgres
    Start-Sleep -Seconds 3
} else {
    Write-Host "Contenedor PostgreSQL activo."
}

$env:POSTGRES_HOST = "localhost"
$env:POSTGRES_PORT = "5432"
$env:POSTGRES_USER = "biess"
$env:POSTGRES_PASSWORD = "biess"
$env:POSTGRES_DB = "srpd_biess"

Write-Host "Ejecutando TestPostulacionRepository_TransaccionalYRollback..."

Push-Location "backend"
go test -count=1 -v -run TestPostulacionRepository_TransaccionalYRollback ./internal/repository
$testExit = $LASTEXITCODE
Pop-Location

if ($testExit -ne 0) {
    Write-Host "ERROR: La prueba de ROLLBACK transaccional FALLO (codigo $testExit)."
    exit 1
}

Write-Host "EXITO: Transaccion probada contra el repositorio real de Go."
Write-Host "Se verifico que ante error de restriccion (CHECK horas > 0), PostgreSQL ejecuto ROLLBACK completo"
Write-Host "y no quedaron registros huerfanos en ninguna tabla hija."
