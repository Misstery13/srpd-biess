#!/usr/bin/env sh
# Actualiza el despliegue en el VPS: ssh srpd-biess './srpd-biess/deploy.sh'
set -e
cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "Generando .env con secretos nuevos"
  cat > .env <<EOF
APP_PORT=8080
JWT_SECRET=$(openssl rand -hex 32)
POSTGRES_PORT=5432
POSTGRES_DB=srpd_biess
POSTGRES_USER=biess
POSTGRES_PASSWORD=$(openssl rand -hex 16)
REDIS_PASSWORD=$(openssl rand -hex 16)
LOGIN_MAX_ATTEMPTS=3
LOGIN_LOCK_MINUTES=10
CATALOG_CACHE_TTL_HOURS=24
SMTP_PORT=1025
EOF
  chmod 600 .env
fi

git pull --ff-only
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
docker image prune -f >/dev/null
docker compose -f docker-compose.prod.yml ps
