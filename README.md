# Dating Platform

Plataforma internacional de dating/relaciones. Monolito modular:
Go (backend) + Next.js (frontend) + PostgreSQL + Redis + Caddy, todo
orquestado con Docker Compose.

Proyecto construido por fases. Estado actual: **Fase 1 — Base del proyecto**.

## Requisitos

* Docker y Docker Compose (plugin `docker compose`).
* Nada más: no hace falta instalar Go, Node, PostgreSQL ni Redis en el host.

## Arrancar el proyecto

```bash
cp .env.example .env
# Edita .env si quieres cambiar contraseñas/puertos por defecto.

make up-build   # primera vez, o tras cambiar dependencias/Dockerfiles
make up         # las siguientes veces
```

Servicios expuestos:

| Servicio  | URL                                    |
|-----------|-----------------------------------------|
| App (vía Caddy) | http://localhost                  |
| Frontend directo | http://localhost:3000            |
| Backend directo  | http://localhost:8080            |
| PostgreSQL       | localhost:5432                   |
| Redis            | localhost:6379                   |

## Detener el proyecto

```bash
make down
```

## Migraciones de base de datos

El esquema de PostgreSQL se gestiona con migraciones SQL versionadas en
`backend/migrations/`, aplicadas con [golang-migrate](https://github.com/golang-migrate/migrate)
a través de un servicio Docker dedicado (no requiere instalar nada en el host).

```bash
# Levanta al menos postgres antes de migrar
docker compose up -d postgres

make migrate-new name=add_profiles   # crea 00000X_add_profiles.up.sql / .down.sql
make migrate-up                      # aplica todas las migraciones pendientes
make migrate-up-one                  # aplica solo la siguiente
make migrate-down-one                # revierte solo la última aplicada
make migrate-down                    # revierte todas (cuidado en producción)
make migrate-version                 # muestra la versión actual aplicada
```

Reglas:

* Nunca se edita una migración ya usada; los cambios posteriores se hacen
  en una migración nueva.
* Cada migración incluye `.up.sql` y `.down.sql`.

## Tests

Se añadirán en la Fase 12 (Tests y calidad).

## Variables de entorno

Ver [`.env.example`](./.env.example) para la lista completa y comentada.
Nunca se sube `.env` (con secretos reales) a Git.

## Estructura del proyecto

```
.
├── docker-compose.yml       # Orquestación de todos los servicios
├── Caddyfile                 # Reverse proxy: /api/* -> backend, resto -> frontend
├── Makefile                  # Comandos de arranque y migraciones
├── .env.example
├── backend/                  # Monolito modular en Go
│   ├── cmd/api/               # Punto de entrada
│   ├── internal/
│   │   ├── config/             # Lectura de variables de entorno
│   │   ├── db/                  # Pool de conexión a PostgreSQL
│   │   ├── redisclient/         # Cliente Redis
│   │   ├── server/              # Router HTTP y middlewares
│   │   ├── health/              # Endpoints de liveness/readiness
│   │   └── users/               # Dominio de cuenta: modelo + repositorio
│   └── migrations/            # Migraciones SQL versionadas
└── frontend/                 # Next.js (App Router, TypeScript)
    ├── app/
    └── messages/               # Esqueleto i18n (es/en), aún sin enrutar
```

## Roadmap de fases

El proyecto se construye fase a fase. Ver el documento de especificación
del proyecto para el listado completo (Fase 1 a Fase 14). No se implementan
fases futuras por adelantado.
