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

## API de perfiles públicos (Fase 6)

Mismo formato que `GET /api/v1/profiles/me`, pero sobre el perfil de
otra persona, identificado por su `profile_id` (el que devuelve la
búsqueda). Todo requiere sesión.

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/profiles/{profileID}` | Perfil público de otra persona |
| GET | `/api/v1/profiles/{profileID}/photos` | Sus fotos |
| GET | `/api/v1/profiles/{profileID}/photos/{photoID}/file` | Contenido de una foto suya |

**Privacidad básica:** solo son visibles los perfiles de cuentas
activas (`status = 'active'`, no eliminadas). Ver el perfil de alguien
suspendido o dado de baja devuelve el mismo `404` genérico que un ID
inexistente, para no filtrar por qué no aparece.

## Frontend: resultados y perfil (Fase 6)

- `/discover` — lista paginada de resultados de `GET /api/v1/search/profiles`, con enlace a cada perfil.
- `/profiles/[id]` — página de un perfil público (datos + fotos), con navegación de vuelta a `/discover`.
- `frontend/lib/api.ts` — helper `apiFetch` compartido por ambas páginas (usa `NEXT_PUBLIC_API_URL`, mismo origen que el navegador vía Caddy, así que la cookie de sesión viaja sola).

Estas páginas requieren una sesión iniciada (cookie `session_id`). Como
todavía no existe una pantalla de login en el frontend (no la pide
ninguna fase hasta ahora), para probarlas primero inicia sesión con
`curl` guardando cookies, y usa esas mismas cookies en el navegador (o
copia la cookie `session_id` con las herramientas de desarrollador).

## API de favoritos (Fase 7)

Los favoritos se referencian por `profile_id` (igual que búsqueda y
perfiles públicos). `POST`/`DELETE` son **idempotentes**: repetir la
misma llamada no es un error, para que el botón de favorito del cliente
no tenga que rastrear su propio estado antes de llamar.

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/favorites` | Lista mis favoritos (paginada, `page`/`page_size`) |
| POST | `/api/v1/favorites/{profileID}` | Añade a favoritos |
| DELETE | `/api/v1/favorites/{profileID}` | Quita de favoritos |
| GET | `/api/v1/favorites/{profileID}` | `{"favorited": true\|false}` |

No puedes añadirte a ti mismo (`400 cannot_favorite_self`), y solo se
pueden favoritear perfiles visibles públicamente (misma regla de la
Fase 6: `404` genérico si no existe o la cuenta no está activa). Un
favorito hacia una cuenta que luego se suspende o elimina deja de
aparecer en el listado automáticamente (se filtra igual que en
búsqueda), sin necesidad de borrar la fila.

`/favorites` en el frontend lista los favoritos, y la página de perfil
(`/profiles/[id]`) incluye un botón para añadir/quitar.

## API de mensajería (Fase 8)

Mensajería 1:1 (sin grupos en V1). Toda ruta requiere sesión.

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/v1/messages/to/{profileID}` | Envía un mensaje a esa persona; crea la conversación si es la primera vez |
| GET | `/api/v1/messages/conversations` | Lista mis conversaciones (paginada), con último mensaje y no leídos |
| GET | `/api/v1/messages/conversations/{conversationID}/messages` | Pagina los mensajes de una conversación (más antiguo primero) |
| POST | `/api/v1/messages/conversations/{conversationID}/messages` | Continúa una conversación ya abierta |

**Lectura:** al listar los mensajes de una conversación (`GET .../messages`),
el backend marca como leídos, como efecto secundario, todos los que te
escribieron a ti — no hace falta un endpoint aparte para "marcar como leído".

**Controles básicos:** el cuerpo del mensaje no puede estar vacío ni
superar 2000 caracteres; no puedes escribirte a ti mismo
(`400 cannot_message_self`); solo puedes iniciar conversación con
perfiles visibles públicamente (misma regla de la Fase 6). El bloqueo
de usuarios (impedir que alguien te escriba) es explícitamente la
Fase 9, no esta.

`/messages` lista las conversaciones; `/messages/[id]` es el hilo, con
caja para responder. Desde `/profiles/[id]` hay un pequeño formulario
para mandar el primer mensaje, que te lleva directo al hilo abierto.

## API de bloqueo y reportes (Fase 9)

### Bloqueo

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/blocks` | Lista a quién he bloqueado yo (paginada) |
| POST | `/api/v1/blocks/{profileID}` | Bloquea (idempotente) |
| DELETE | `/api/v1/blocks/{profileID}` | Desbloquea (idempotente) |
| GET | `/api/v1/blocks/{profileID}` | `{"blocked": true\|false}` |

**Reglas de visibilidad (el efecto es mutuo, no solo desde quien bloquea):**
si A bloquea a B, ninguno de los dos aparece en los resultados de
búsqueda, la lista de favoritos ni la lista de conversaciones del otro;
tampoco pueden verse el perfil público ni sus fotos (mismo `404`
genérico que "no existe" o "cuenta inactiva"), ni escribirse mensajes
nuevos — ni siquiera continuando una conversación que ya existía antes
del bloqueo. Nada de esto borra datos: favoritos y mensajes previos
siguen en la base de datos, solo se filtran al mostrarlos.

Solo quien bloqueó puede desbloquear. Bloquear y reportar funcionan
incluso hacia alguien que ya te ha bloqueado a ti (usan una resolución
de perfil que ignora las reglas de visibilidad a propósito).

### Reportes

| Método | Ruta | Descripción |
|--------|------|-------------|
| POST | `/api/v1/reports/{profileID}` | Crea un reporte: `{"reason": "...", "description": "..."}` |

`reason` debe ser uno de: `spam`, `fake_profile`, `harassment`,
`inappropriate_content`, `underage`, `other`. `description` es opcional
(máx. 2000 caracteres). Los reportes quedan en estado `pending`; su
revisión es el panel de moderación de la Fase 10.

`/blocked` en el frontend lista y permite desbloquear; `/profiles/[id]`
tiene botones de bloquear/desbloquear y un formulario de reporte.

## Administración y moderación (Fase 10)

### Crear el primer administrador

No hay registro público de admins. Registra una cuenta normal
(`POST /api/v1/auth/register` o el frontend) y luego promociónala:

```bash
make admin-promote email=tu-email@example.com
```

Esto ejecuta `backend/cmd/promote-admin` dentro del contenedor del
backend ya construido (sin tocar SQL a mano). A partir de aquí, esa
cuenta puede usar `/admin` y las rutas `/api/v1/admin/*` con su sesión
normal — no hay un login separado para admins.

### API

Estas rutas exigen sesión **y** rol `admin` (`403 forbidden` si no lo
tienes). A diferencia del resto de la API, identifican cuentas por
`user_id` (no `profile_id`): son herramientas internas, no cara al público.

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/admin/users?status=&page=&page_size=` | Lista cuentas (incluye suspendidas/eliminadas) |
| POST | `/api/v1/admin/users/{userID}/suspend` | Suspende una cuenta |
| POST | `/api/v1/admin/users/{userID}/reactivate` | Reactiva una cuenta suspendida |
| GET | `/api/v1/admin/reports?status=&page=&page_size=` | Lista reportes (`pending` por defecto si no se filtra) |
| POST | `/api/v1/admin/reports/{reportID}/resolve` | `{"status": "reviewed"\|"dismissed"}` |

**Efecto de suspender:** una cuenta suspendida deja de poder usar la
API de inmediato, incluso con una sesión ya abierta — no hace falta
esperar a que expire la cookie (`auth.Service.CurrentUser` comprueba el
estado en cada petición autenticada, no solo al hacer login). Un admin
no puede suspenderse a sí mismo (`400 cannot_act_on_self`).

`/admin` en el frontend es el panel: pestañas de Reportes y Usuarios,
con las acciones de arriba. Si tu cuenta no es admin, verás un mensaje
de permisos en vez del panel (el backend ya lo protege; el frontend solo
refleja el `403`).

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
│   ├── cmd/
│   │   ├── api/                # Punto de entrada del servidor
│   │   └── promote-admin/       # CLI para promocionar una cuenta a admin
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
│   │   ├── favorites/               # Favoritos: añadir, eliminar, listar
│   │   ├── messaging/                # Conversaciones y mensajes 1:1
│   │   ├── blocking/                  # Bloqueo entre usuarios y reglas de visibilidad
│   │   ├── reports/                   # Creación de reportes
│   │   ├── admin/                     # Panel de administración (Fase 10)
│   │   └── storage/               # Abstracción de almacenamiento de ficheros (driver "local" en V1)
│   └── migrations/            # Migraciones SQL versionadas
└── frontend/                 # Next.js (App Router, TypeScript)
    ├── app/
    │   ├── discover/           # Resultados de búsqueda (Fase 6)
    │   ├── profiles/[id]/      # Página de perfil público (Fase 6)
    │   ├── favorites/          # Lista de favoritos (Fase 7)
    │   ├── messages/           # Conversaciones y mensajes (Fase 8)
    │   ├── blocked/            # Perfiles bloqueados (Fase 9)
    │   └── admin/              # Panel de administración (Fase 10)
    ├── lib/                    # Helpers compartidos (cliente de API)
    └── messages/               # Esqueleto i18n (es/en), aún sin enrutar
```

## Roadmap de fases

El proyecto se construye fase a fase. Ver el documento de especificación
del proyecto para el listado completo (Fase 1 a Fase 14). No se implementan
fases futuras por adelantado.
