#!/usr/bin/env bash
# ==============================================================================
# Script de Demostración de ROLLBACK Transaccional (PostgreSQL)
# Concurso Público BIESS 2026 — Pista D (Datos): Gino Bermúdez
# ==============================================================================

echo "Iniciando prueba de consistencia transaccional y ROLLBACK en PostgreSQL..."

if ! docker ps --filter "name=postgres" --format "{{.Names}}" | grep -q "postgres"; then
    echo "Contenedor de PostgreSQL no está levantado. Iniciando con docker compose..."
    docker compose up -d postgres
    sleep 3
fi

echo "Ejecutando db/prueba_rollback.sql en PostgreSQL..."
docker compose exec -T postgres psql -U biess -d srpd_biess < db/prueba_rollback.sql

echo "Demostración de ROLLBACK finalizada."
