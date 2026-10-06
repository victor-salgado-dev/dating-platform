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

## Estructura del proyecto

### Raíz

```text
.
├── docker-compose.yml          # Stack de desarrollo
├── docker-compose.prod.yml     # Override de producción
├── Caddyfile                   # Proxy de desarrollo
├── Caddyfile.prod              # Proxy de producción (HTTPS automático)
├── Makefile                    # Targets principales
├── seed_profiles.ps1           # Generación de perfiles de demostración
├── backend/                    # Servicio Go
├── frontend/                   # Aplicación Next.js
└── scripts/                    # Scripts auxiliares
```

### Backend (`backend/`)

```text
backend/
├── cmd/
│   ├── api/                     # Servidor HTTP de la API
│   └── promote-admin/           # Comando para promover un usuario a admin
├── internal/
│   ├── activity/                # Actividad reciente del usuario
│   ├── admin/                   # Gestión administrativa y validaciones
│   ├── apperr/                  # Errores globales de aplicación
│   ├── auth/                    # Autenticación, sesiones y contraseñas
│   ├── blocking/                # Bloqueo de usuarios
│   ├── config/                  # Carga de configuración desde variables de entorno
│   ├── consent/                 # Tipos de consentimiento
│   ├── contact/                 # Formulario de contacto
│   ├── email/                   # Envío de correos (noop/smtp)
│   ├── favorites/               # Gestión de favoritos
│   ├── httpx/                   # Utilidades HTTP (JSON, errores, IP)
│   ├── likes/                   # Likes enviados y recibidos
│   ├── messaging/               # Mensajería y conversaciones
│   ├── profiles/                # Perfiles, fotos, preferencias, idiomas, etc.
│   ├── ratelimit/               # Limitación de peticiones usando Redis
│   ├── redisclient/             # Cliente Redis
│   ├── reports/                 # Reportes de usuarios
│   ├── search/                  # Búsqueda y filtros de perfiles
│   ├── server/                  # Middlewares y utilidades de servidor HTTP
│   ├── storage/                 # Almacenamiento de archivos (local/s3)
│   └── visits/                  # Visitas a perfiles
└── migrations/                  # Migraciones SQL versionadas
```

#### Módulos backend destacados

- **profiles**: contiene los tipos de `Gender`, `RelationshipGoal`,
  `BaseListItem`, `ProfileDetails`, `PartnerPreferences`, `PersonalityTrait`,
  así como la resolución de IDs, URLs de fotos y validaciones de género.
- **search**: define los filtros de búsqueda (`ViewerScope`, `RawBounds`,
  `PersonalityFilter`, `InterestFilter`) y los tipos de resultado.
- **auth**: expone la configuración de contraseñas (`MinPasswordLength`),
  sesiones y resumen de cuenta (`AccountSummary`).
- **storage**: define la interfaz `Storage` con `Save`, `Open`, `Delete` y
  los drivers `local` y `s3`.
- **httpx**: provee `WriteJSON`, `WriteError` y `ClientIP` para respuestas
  HTTP consistentes.
- **ratelimit**: implementa el límite de peticiones apoyado en Redis.

### Frontend (`frontend/`)

```text
frontend/
├── app/                         # Rutas y páginas de Next.js
│   ├── discover/                # Página de descubrimiento de perfiles
│   ├── legal/contact/           # Página de contacto legal
│   ├── new-members/             # Redirección a la pestaña de nuevos miembros
│   ├── online-now/              # Redirección a la pestaña de online now
│   └── popular/                 # Redirección a la pestaña de populares
├── components/                  # Componentes React reutilizables
│   ├── InteractionListSection.tsx
│   └── ProfileCard.tsx
├── lib/
│   ├── api.ts                   # Cliente de API tipado y caché de GET
│   ├── geocoding.ts             # Autocompletado de ubicaciones con Photon
│   ├── i18n/
│   │   ├── config.ts            # Configuración del locale
│   │   ├── context.tsx          # Provider de idioma
│   │   ├── dictionaries/
│   │   │   └── es.ts            # Diccionario base en español
│   │   └── options.ts           # Resolución de etiquetas traducidas
│   ├── seekingGenders.ts        # Definición de opciones "Qué busco"
│   ├── useProfileInteractions.ts # Estado global de likes/favoritos
│   └── useProfileList.ts        # Paginación de listas de perfiles
├── messages/                    # Archivos JSON de traducción (opcionales)
│   ├── es.json
│   └── en.json
├── package.json                 # Dependencias y scripts del frontend
└── tsconfig.json                # Configuración de TypeScript
```

#### Utilidades frontend relevantes

- **api.ts**: implementa `apiFetch` y `apiMutateQuiet`, maneja errores
  `ApiError`, caché de GET de 60 s, deduplicación de peticiones en vuelo y
  limpieza de caché tras mutaciones. Exporta los tipos principales de las
  respuestas de la API (`SearchResponse`, `PublicProfile`, `MessageItem`,
  `PartnerPreferences`, etc.).
- **useProfileList.ts**: hook usado por las pestañas de Home para cargar
  páginas de `recommended`, `popular`, `online-now` y `new-members`,
  manteniendo los datos anteriores mientras se carga una página nueva.
- **useProfileInteractions.ts**: hook global que centraliza la carga de
  likes/favoritos enviados y recibidos y expone los conjuntos de IDs y las
  funciones `toggleLike` y `toggleFavorite`.
- **geocoding.ts**: autocompletado de ubicaciones usando Photon
  (OpenStreetMap). Devuelve `PlaceSuggestion` con `label`, `city`, `state`,
  `country` y `countryCode`.
- **seekingGenders.ts**: mapea las opciones seleccionadas por el usuario
  (`male`, `female`, `trans`) a los valores de `Gender` del backend.
- **i18n/context.tsx**: define `I18nProvider` y `useI18n` para inyectar el
  locale y el diccionario en los componentes cliente.
- **i18n/options.ts**: funciones `tOption` y `tOptionList` para resolver
  labels traducidas de valores técnicos como `relationshipGoal`,
  `bodyType`, etc.

### Scripts (`scripts/`)

```text
scripts/
├── smoke-test.sh
├── backup.sh
└── restore.sh
```

## Datos de demostración

`seed_profiles.ps1` genera **5.000** perfiles de demostración en un entorno
local. Requiere PowerShell y fotos `.jpg` en la ruta configurada dentro del
script.

El script registra o inicia sesión con cuentas `user1@datingdemo.com` ...
`user5000@datingdemo.com`, crea perfiles, idiomas, intereses, personalidad,
preferencias, fotos e interacciones masivas. Está pensado únicamente para
desarrollo.

## Limitaciones conocidas

- No hay tokens CSRF explícitos. `SameSite=Lax` cubre el caso habitual, pero
  no sustituye una solución CSRF completa.
- El enrutado i18n no está activado, aunque ya existen diccionarios
  TypeScript en `frontend/lib/i18n/dictionaries` y archivos JSON en
  `frontend/messages`.
- El almacenamiento local (`backend_uploads`) es adecuado para desarrollo y
  una sola instancia; para varias instancias debe usarse S3 u otro
  almacenamiento compatible.
