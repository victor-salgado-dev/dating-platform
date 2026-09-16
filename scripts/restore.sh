#!/bin/sh
# Restaura un backup creado por scripts/backup.sh.
#
# ATENCIÓN: sobrescribe la base de datos actual. Úsalo con la misma
# precaución que cualquier restauración de producción.
#
# Uso: ./scripts/restore.sh backups/dating_platform_20260101T000000Z.sql.gz
set -e

FILE="$1"
if [ -z "$FILE" ]; then
    echo "Uso: ./scripts/restore.sh <fichero-backup.sql.gz>"
    exit 1
fi
if [ ! -f "$FILE" ]; then
    echo "ERROR: no existe el fichero $FILE"
    exit 1
fi

echo "Vas a restaurar '$FILE' sobre la base de datos '${POSTGRES_DB:-dating_app}'."
echo "Esto SOBRESCRIBE los datos actuales. Escribe 'si' para continuar:"
read -r confirm
if [ "$confirm" != "si" ]; then
    echo "Cancelado."
    exit 1
fi

echo "==> Restaurando..."
gunzip -c "$FILE" | docker compose exec -T postgres psql \
    -U "${POSTGRES_USER:-dating_app}" \
    -d "${POSTGRES_DB:-dating_app}"

echo "==> Restauración completada."
