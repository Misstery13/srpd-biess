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

# Enlace de recuperación de contraseña (I3): APP_URL = IP pública del VPS, leída en el momento del despliegue
# (metadatos de EC2 con IMDSv2). No se escribe en el repo; se agrega al .env solo si falta.
if ! grep -q '^APP_URL=' .env; then
  imds=http://169.254.169.254/latest
  token=$(curl -fsS -m 2 -X PUT "$imds/api/token" -H 'X-aws-ec2-metadata-token-ttl-seconds: 60' || true)
  ip=$(curl -fsS -m 2 -H "X-aws-ec2-metadata-token: $token" "$imds/meta-data/public-ipv4" || true)
  if [ -n "$ip" ]; then
    echo "APP_URL=http://$ip" >> .env  # NOSONAR: sin dominio no hay certificado; con dominio poner APP_URL=https://... a mano
    echo "APP_URL agregado al .env"
  else
    echo "AVISO: no se pudo leer la IP pública; agregue APP_URL=http://<ip-o-dominio> al .env a mano" >&2  # NOSONAR: solo es un mensaje de ayuda
  fi
fi

git pull --ff-only
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
docker image prune -f >/dev/null
docker compose -f docker-compose.prod.yml ps
