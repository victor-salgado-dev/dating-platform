# Dating Platform

Plataforma internacional de dating/relaciones. Monolito modular:
Go (backend) + Next.js (frontend) + PostgreSQL + Redis + Caddy, todo
orquestado con Docker Compose.

Proyecto construido por fases. Estado actual: **Fase 1 — Base del proyecto**.

## Requisitos

* Docker y Docker Compose (plugin `docker compose`).
* Nada más: no hace falta instalar Go, Node, PostgreSQL ni Redis en el host.

> **Nota sobre `go.sum`:** en desarrollo, el contenedor del backend
> regenera `go.sum` automáticamente en cada arranque (`go mod tidy`,
> ver `backend/docker-entrypoint.dev.sh`), porque el volumen
> `./backend:/app` sustituye lo que se construyó en la imagen por el
> código del host. La primera vez que ejecutes `make up-build` necesitará
> red para resolver dependencias; las siguientes usarán la cache
> (`backend_go_cache`). Una vez generado, `backend/go.sum` queda en tu
> copia local — te recomendamos commitearlo a Git para builds reproducibles.

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

## API de autenticación (Fase 3)

Todos los endpoints devuelven JSON. Los que requieren sesión leen la
cookie httpOnly `session_id` (nombre configurable vía `SESSION_COOKIE_NAME`).

| Método | Ruta | Auth | Descripción |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Crea cuenta, envía email de verificación, abre sesión |
| POST | `/api/v1/auth/login` | No | Verifica credenciales, abre sesión |
| POST | `/api/v1/auth/logout` | Cookie | Cierra la sesión actual |
| GET | `/api/v1/auth/me` | Cookie | Devuelve la cuenta autenticada |
| POST | `/api/v1/auth/password/forgot` | No | Solicita reset (siempre 202, evita enumeración) |
| POST | `/api/v1/auth/password/reset` | No | Consume el token y fija contraseña nueva |
| POST | `/api/v1/auth/email/verify` | No | Consume el token de verificación de email |
| POST | `/api/v1/auth/email/resend` | Cookie | Reenvía el email de verificación |
| DELETE | `/api/v1/auth/account` | Cookie | Elimina (soft delete) la cuenta, requiere confirmar contraseña |

En desarrollo (`EMAIL_DRIVER=noop`) los emails no se envían de verdad:
se registran en los logs del backend (`docker compose logs -f backend`),
lo que permite copiar el token de verificación/reset directamente de ahí.

## API de perfil (Fase 4)

Todos requieren sesión (cookie). En V1 solo se gestiona el **propio**
perfil; ver perfiles de otras personas llega en la Fase 6.

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/profiles/me` | Perfil propio (404 si aún no se ha creado) |
| POST | `/api/v1/profiles/me` | Crea el perfil (una sola vez, relación 1:1 con la cuenta) |
| PATCH | `/api/v1/profiles/me` | Edición parcial. Una clave ausente no se toca; una clave a `null` borra ese dato (pasa a "desconocido") |
| POST | `/api/v1/profiles/me/photos` | Sube una foto (`multipart/form-data`, campo `photo`; JPEG/PNG/WebP, máx. 5 MB, máx. 6 fotos) |
| GET | `/api/v1/profiles/me/photos` | Lista las fotos propias |
| GET | `/api/v1/profiles/me/photos/{id}/file` | Sirve el contenido binario de una foto |
| DELETE | `/api/v1/profiles/me/photos/{id}` | Borra una foto |

La regla de los datos faltantes (sección 8) se aplica de raíz: los
campos opcionales del perfil (región, idiomas, objetivo de relación,
hijos, deseo de hijos, bio, intereses) se guardan como `NULL` cuando no
se han indicado, nunca con un valor por defecto inventado. La Fase 5
(búsqueda) debe respetar esta distinción al filtrar.

## API de búsqueda (Fase 5)

`GET /api/v1/search/profiles` — requiere sesión. Devuelve fichas
resumidas (no el perfil completo: eso es la Fase 6) de otros perfiles,
paginadas.

| Parámetro | Formato | Descripción |
|---|---|---|
| `gender` | `female,male` (repetible o por comas) | Filtra por género |
| `min_age`, `max_age` | entero (18–120) | Rango de edad |
| `country` | `ES` | Código de país ISO 3166-1 alpha-2 |
| `language` | `es,en` (repetible o por comas) | Habla al menos uno de estos idiomas |
| `relationship_goal` | uno de `casual\|long_term\|friendship\|marriage\|not_sure` | Objetivo de relación |
| `has_children`, `wants_children` | `true\|false` | Información familiar |
| `interests` | `chess,hiking` (repetible o por comas) | Coincide con al menos uno |
| `sort` | `recent` (por defecto) \| `age_asc` \| `age_desc` | Orden |
| `page`, `page_size` | entero | Paginación (`page_size` máx. 50) |

**Regla de los datos faltantes aplicada a la búsqueda:** si se envía un
filtro (p. ej. `language=fr`) y un perfil no ha indicado ese dato
(`languages IS NULL`), ese perfil **no** aparece en los resultados. Si
el filtro no se envía, el perfil aparece con normalidad aunque le falte
ese dato. La implementación se apoya en que SQL nunca evalúa a `TRUE`
una comparación contra `NULL`, así que esto ocurre automáticamente en
cada cláusula, sin lógica especial por campo (ver comentarios en
`backend/internal/search/postgres_repository.go`).

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
│   │   ├── httpx/                # Helpers de respuesta/error JSON compartidos
│   │   ├── health/              # Endpoints de liveness/readiness
│   │   ├── users/               # Dominio de cuenta: modelo + repositorio
│   │   ├── auth/                 # Registro, login, sesiones, tokens, email de verificación/reset
│   │   ├── email/                # Abstracción de envío de email (driver "noop" en V1)
│   │   ├── profiles/              # Perfil propio: datos estructurados + fotos
│   │   ├── search/                 # Búsqueda de perfiles con filtros estructurados
│   │   └── storage/               # Abstracción de almacenamiento de ficheros (driver "local" en V1)
│   └── migrations/            # Migraciones SQL versionadas
└── frontend/                 # Next.js (App Router, TypeScript)
    ├── app/
    └── messages/               # Esqueleto i18n (es/en), aún sin enrutar
```

## Roadmap de fases

El proyecto se construye fase a fase. Ver el documento de especificación
del proyecto para el listado completo (Fase 1 a Fase 14). No se implementan
fases futuras por adelantado.
