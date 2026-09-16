#!/bin/sh
# Backup de PostgreSQL (Fase 14). Ejecuta pg_dump DENTRO del contenedor
# de postgres (misma versión que la base de datos real, sin tener que
# instalar herramientas de PostgreSQL en el host) y copia el resultado
# a ./backups en el host.
#
# Uso: ./scripts/backup.sh   (o `make backup`)
set -e

BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
FILENAME="dating_platform_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "==> Volcando la base de datos (pg_dump dentro del contenedor)..."
docker compose exec -T postgres pg_dump \
    -U "${POSTGRES_USER:-dating_app}" \
    -d "${POSTGRES_DB:-dating_app}" \
    --no-owner --no-privileges \
    | gzip > "${BACKUP_DIR}/${FILENAME}"

size=$(du -h "${BACKUP_DIR}/${FILENAME}" | cut -f1)
echo "==> Backup guardado en ${BACKUP_DIR}/${FILENAME} (${size})"
