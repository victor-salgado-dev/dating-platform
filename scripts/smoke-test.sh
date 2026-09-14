#!/bin/sh
# Smoke test del stack Docker completo (Fase 12: "comprobaciones Docker").
#
# No sustituye a los tests de Go: comprueba que `docker compose up`
# realmente deja el stack en un estado usable —contenedores arriba,
# migraciones aplicables desde cero, endpoints de salud respondiendo—
# que es justo el tipo de fallo que un test de Go no detecta.
#
# Uso: ./scripts/smoke-test.sh   (o `make smoke-test`)
set -e

BASE_URL="${SMOKE_TEST_BASE_URL:-http://localhost}"
MAX_WAIT_SECONDS=60

echo "==> Levantando el stack (docker compose up -d)..."
docker compose up -d

echo "==> Esperando a que el backend responda (máx. ${MAX_WAIT_SECONDS}s)..."
waited=0
until curl -fsS "${BASE_URL}/healthz" > /dev/null 2>&1; do
    waited=$((waited + 2))
    if [ "$waited" -ge "$MAX_WAIT_SECONDS" ]; then
        echo "FALLO: el backend no respondió a /healthz tras ${MAX_WAIT_SECONDS}s."
        echo "---- docker compose ps ----"
        docker compose ps
        echo "---- logs del backend ----"
        docker compose logs --tail=50 backend
        exit 1
    fi
    sleep 2
done
echo "    backend arriba tras ${waited}s."

echo "==> Aplicando migraciones desde el estado actual..."
docker compose run --rm migrate -path=/migrations \
    -database "postgres://${POSTGRES_USER:-dating_app}:${POSTGRES_PASSWORD:-changeme_dev_password}@postgres:5432/${POSTGRES_DB:-dating_app}?sslmode=disable" \
    up

echo "==> Comprobando /api/v1/health (Postgres + Redis)..."
health_json=$(curl -fsS "${BASE_URL}/api/v1/health")
echo "    ${health_json}"

case "$health_json" in
    *'"status":"ok"'*)
        echo "    readiness: ok"
        ;;
    *)
        echo "FALLO: /api/v1/health no informa status=ok: ${health_json}"
        exit 1
        ;;
esac

echo "==> Comprobando que el frontend responde..."
if ! curl -fsS "${BASE_URL}/" > /dev/null; then
    echo "FALLO: el frontend no respondió en ${BASE_URL}/"
    exit 1
fi

echo ""
echo "TODO OK: el stack arranca, migra y responde correctamente."
