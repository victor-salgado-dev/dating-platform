.PHONY: up up-build down logs ps \
        migrate-new migrate-up migrate-down migrate-up-one migrate-down-one migrate-version migrate-force

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
