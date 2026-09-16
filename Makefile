.PHONY: up up-build down logs ps \
        migrate-new migrate-up migrate-down migrate-up-one migrate-down-one migrate-version migrate-force \
        admin-promote \
        test test-integration test-all smoke-test \
        prod-build prod-up prod-down prod-logs backup restore

# Carga las variables de .env para poder usarlas en los comandos de migración.
-include .env
export

DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

# ---------------------------------------------------------
# Ciclo de vida del entorno de desarrollo
# ---------------------------------------------------------

## Arranca todos los servicios (usa las imágenes ya construidas si existen)
up:
	docker compose up

## Arranca todos los servicios reconstruyendo las imágenes
up-build:
	docker compose up --build

## Detiene y elimina los contenedores
down:
	docker compose down

## Muestra logs en vivo de todos los servicios
logs:
	docker compose logs -f

## Lista el estado de los servicios
ps:
	docker compose ps

# ---------------------------------------------------------
# Migraciones de base de datos
#
# Requiere que 'postgres' esté levantado y saludable:
#   docker compose up -d postgres
# ---------------------------------------------------------

## Crea un nuevo par de archivos de migración .up.sql / .down.sql
## Uso: make migrate-new name=add_profiles
migrate-new:
	@if [ -z "$(name)" ]; then echo "ERROR: falta name. Uso: make migrate-new name=add_profiles"; exit 1; fi
	docker compose run --rm migrate create -ext sql -dir /migrations -seq -digits 3 $(name)

## Aplica todas las migraciones pendientes
migrate-up:
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" up

## Revierte todas las migraciones (usar con cuidado, especialmente en producción)
migrate-down:
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" down

## Aplica únicamente la siguiente migración pendiente
migrate-up-one:
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" up 1

## Revierte únicamente la última migración aplicada
migrate-down-one:
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" down 1

## Muestra la versión de migración actual aplicada en la base de datos
migrate-version:
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" version

## Fuerza una versión de migración concreta (recuperación tras un fallo)
## Uso: make migrate-force VERSION=3
migrate-force:
	@if [ -z "$(VERSION)" ]; then echo "ERROR: falta VERSION. Uso: make migrate-force VERSION=3"; exit 1; fi
	docker compose run --rm migrate -path=/migrations -database "$(DB_URL)" force $(VERSION)

# ---------------------------------------------------------
# Administración (Fase 10)
# ---------------------------------------------------------

## Promociona una cuenta YA REGISTRADA a rol admin. Es la única forma
## soportada de crear un administrador.
## Uso: make admin-promote email=persona@example.com
admin-promote:
	@if [ -z "$(email)" ]; then echo "ERROR: falta email. Uso: make admin-promote email=persona@example.com"; exit 1; fi
	docker compose run --rm --entrypoint go backend run ./cmd/promote-admin --email=$(email)

# ---------------------------------------------------------
# Tests y calidad (Fase 12)
# ---------------------------------------------------------

## Tests unitarios: rápidos, sin red, se pueden ejecutar siempre.
test:
	docker compose run --rm --entrypoint go backend test ./...

## Tests de integración: requieren Postgres y Redis levantados
## (usan la misma base de datos de desarrollo; ver README).
test-integration:
	docker compose up -d postgres redis
	docker compose run --rm --entrypoint go backend test -tags=integration ./...

## Unitarios + integración en una sola llamada.
test-all: test test-integration

## Smoke test del stack Docker completo: arranca, migra y comprueba
## los endpoints de salud. No sustituye a `test`/`test-integration`.
smoke-test:
	./scripts/smoke-test.sh

# ---------------------------------------------------------
# Producción (Fase 14)
#
# Requiere un .env de producción (ver .env.production.example) y
# Docker Compose >= 2.24 (por el uso de !override/!reset en
# docker-compose.prod.yml).
# ---------------------------------------------------------

## Construye las imágenes de producción (backend/Dockerfile, frontend/Dockerfile).
prod-build:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml build

## Arranca el stack de producción (Caddy con HTTPS real en 80/443).
prod-up:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

## Detiene el stack de producción.
prod-down:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml down

## Logs del stack de producción.
prod-logs:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

## Backup de PostgreSQL a ./backups (pg_dump dentro del contenedor).
backup:
	./scripts/backup.sh

## Restaura un backup. Uso: make restore file=backups/xxx.sql.gz
restore:
	@if [ -z "$(file)" ]; then echo "ERROR: falta file. Uso: make restore file=backups/xxx.sql.gz"; exit 1; fi
	./scripts/restore.sh $(file)
