# Dating Platform

Plataforma de dating y relaciones. Backend en Go, frontend en Next.js,
PostgreSQL, Redis y Caddy como reverse proxy. Todo se ejecuta con Docker
Compose.

El repositorio es una V1 en evolución: el backend está organizado en módulos
Go, el frontend usa Next.js App Router y TypeScript, y las rutas de la API se
gestionan bajo `/api/v1`. El cliente compartido del frontend está en
`frontend/lib/api.ts`.

## Requisitos

- Docker Desktop con `docker compose`.
- `make`.
- En Windows, Git Bash, WSL o cualquier shell POSIX para los scripts `.sh`.

No es necesario instalar Go, Node.js, PostgreSQL ni Redis en el host.

## Configuración inicial

```bash
cp .env.example .env
# En PowerShell: Copy-Item .env.example .env
```

Edita `.env` si quieres cambiar contraseñas, puertos o límites. Para
producción usa `.env.production.example`.

## Comandos principales

Todos los targets están definidos en el `Makefile`.

| Target | Descripción |
| --- | --- |
| `make up-build` | Construye e inicia los contenedores de desarrollo. |
| `make up` | Inicia el stack de desarrollo sin reconstruir. |
| `make down` | Detiene y elimina los contenedores. |
| `make logs` | Sigue los logs de todos los servicios. |
| `make ps` | Lista el estado de los servicios. |
| `make migrate-new name=add_example` | Crea una pareja de migraciones `.up.sql` / `.down.sql`. |
| `make migrate-up` | Aplica las migraciones pendientes. |
| `make migrate-up-one` | Aplica la siguiente migración pendiente. |
| `make migrate-down-one` | Revierte la última migración aplicada. |
| `make migrate-down` | Revierte todas las migraciones. |
| `make migrate-version` | Muestra la versión actual de migración. |
| `make migrate-force VERSION=3` | Fuerza la versión indicada. |
| `make admin-promote email=persona@example.com` | Promociona una cuenta existente a rol `admin`. |
| `make test` | Tests unitarios. |
| `make test-integration` | Tests de integración. |
| `make test-all` | Unitarios + integración. |
| `make smoke-test` | Smoke test del stack Docker completo. |
| `make prod-build` | Construye las imágenes de producción. |
| `make prod-up` | Inicia el stack de producción. |
| `make prod-down` | Detiene el stack de producción. |
| `make prod-logs` | Sigue los logs de producción. |
| `make backup` | Crea un backup de PostgreSQL. |
| `make restore FILE=backups/xxx.sql.gz` | Restaura un backup. |

## Servicios de desarrollo

Con `make up` estarán disponibles:

| Servicio | URL/puerto |
| --- | --- |
| Aplicación a través de Caddy | http://localhost |
| Frontend directo | http://localhost:3000 |
| Backend directo | http://localhost:8080 |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

En producción solo Caddy publica puertos; PostgreSQL, Redis y el backend no
quedan expuestos al host.

## Migraciones

Las migraciones SQL están en `backend/migrations/` y se ejecutan con el
servicio `migrate` definido en `docker-compose.yml`. `docker compose up` **no**
aplica migraciones automáticamente.

Para preparar una base de datos desde cero:

```bash
docker compose up -d postgres
make migrate-up
```

No edites migraciones ya aplicadas. Los cambios deben incluirse en una nueva
pareja de migraciones `.up.sql` / `.down.sql`.

## Tests

```bash
make test               # unitarios, no requieren Postgres ni Redis
make test-integration   # levanta Postgres/Redis y ejecuta tests de integración
make test-all           # unitarios + integración
make smoke-test         # arranca el stack, aplica migraciones y comprueba salud
```

`make test-integration` levanta los contenedores necesarios, pero no aplica
migraciones. Antes de ejecutarlo sobre una base vacía, asegúrate de haber
aplicado las migraciones:

```bash
docker compose up -d postgres
make migrate-up
```

## Producción

La producción usa `docker-compose.yml` junto con
`docker-compose.prod.yml`. Requiere Docker Compose 2.24 o posterior por los
tags `!override` y `!reset`.

Resumen:

1. Copia `.env.production.example` como `.env` y rellena los valores marcados.
2. Construye e inicia:

```bash
make prod-build
make prod-up
```

3. Aplica migraciones:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm migrate \
  -path=/migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable" up
```

4. Registra una cuenta normal y promuévela a administradora:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm \
  --entrypoint go backend run ./cmd/promote-admin \
  --email=persona@example.com
```

En producción Caddy usa `Caddyfile.prod`, obtiene el certificado Let's Encrypt
automáticamente y publica los puertos `80` y `443`.

## Backups y restauración

```bash
make backup
make restore FILE=backups/dating_platform_YYYYMMDDTHHMMSSZ.sql.gz
```

El backup se guarda en `./backups` como `.sql.gz`. La restauración pide
escribir `si` y sobrescribe la base de datos actual.

## Variables de entorno

Las listas comentadas están en `.env.example` y `.env.production.example`.
Áreas principales:

- PostgreSQL y Redis
- puertos y nivel de logs
- sesión y tokens
- email (`noop` en desarrollo o `smtp` en producción)
- rate limiting
- URL pública del frontend y URL interna para SSR
- storage local o S3

`NEXT_PUBLIC_API_URL` es la URL que usa el navegador. `INTERNAL_API_URL` es la
URL interna para llamadas desde el servidor de Next.js dentro de la red
Docker.

## Estructura

```text
.
├── docker-compose.yml          # Stack de desarrollo
├── docker-compose.prod.yml     # Override de producción
├── Caddyfile                   # Proxy de desarrollo
├── Caddyfile.prod              # Proxy de producción (HTTPS automático)
├── Makefile                    # Targets principales
├── seed_profiles.ps1           # Generación de perfiles de demostración
├── backend/
│   ├── cmd/api/                # Servidor HTTP
│   ├── cmd/promote-admin/      # Promoción a administrador
│   ├── internal/               # Módulos de dominio e infraestructura
│   └── migrations/             # Migraciones SQL versionadas
├── frontend/
│   ├── app/                    # Páginas y rutas de Next.js
│   ├── lib/                    # Cliente de API compartido
│   └── messages/               # Primeros archivos de traducción
└── scripts/
    ├── smoke-test.sh
    ├── backup.sh
    └── restore.sh
```

## Datos de demostración

`seed_profiles.ps1` genera **5.000** perfiles de demostración en un entorno
local. Requiere PowerShell y fotos `.jpg` en la ruta configurada dentro del
script:

```text
C:\Users\Victor\Downloads\fotos\Nueva carpeta (4)
```

El script registra o inicia sesión con cuentas `user1@datingdemo.com` ...
`user5000@datingdemo.com`, crea perfiles, idiomas, intereses, personalidad,
preferencias, fotos e interacciones masivas. Está pensado únicamente para
desarrollo.

## Limitaciones conocidas

- No hay tokens CSRF explícitos. `SameSite=Lax` cubre el caso habitual, pero
  no sustituye una solución CSRF completa.
- El enrutado i18n no está activado, aunque ya existen
  `frontend/messages/es.json` y `frontend/messages/en.json`.
- El almacenamiento local (`backend_uploads`) es adecuado para desarrollo y
  una sola instancia; para varias instancias debe usarse S3 u otro
  almacenamiento compatible.
