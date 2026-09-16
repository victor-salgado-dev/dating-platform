# Dating Platform

Plataforma internacional de dating y relaciones. Es un monolito modular
compuesto por Go, Next.js, PostgreSQL, Redis y Caddy, orquestado con Docker
Compose.

## Estado actual

La implementación cubre las fases 1 a 14 del roadmap original:

- cuentas, sesiones, verificación de email y recuperación de contraseña;
- perfiles propios con fotos y perfiles públicos;
- búsqueda paginada con filtros;
- favoritos, mensajería 1:1, bloqueos y reportes;
- administración de usuarios y moderación de reportes;
- rate limiting, cabeceras de seguridad, límites de peticiones y validación
  del contenido real de las imágenes;
- consentimiento de términos y privacidad, páginas legales y contacto;
- imágenes de producción, HTTPS automático, email SMTP, almacenamiento S3 y
  backups de PostgreSQL.

El proyecto sigue siendo una V1. Hay limitaciones conocidas documentadas al
final de este archivo.

## Requisitos

- Docker Desktop con el plugin `docker compose`.
- `make`.
- Git Bash, WSL o un shell POSIX para los scripts `.sh` y algunos targets de
  `make` en Windows.

No es necesario instalar Go, Node.js, PostgreSQL ni Redis en el host.

En el primer arranque el contenedor de desarrollo del backend ejecuta
`go mod tidy`, porque `./backend` se monta como volumen. Necesita acceso a
Internet para descargar dependencias la primera vez; después se usa el volumen
de caché `backend_go_cache`. El `go.sum` generado queda en `backend/go.sum`.

## Desarrollo local

### Configuración

```bash
cp .env.example .env
# En PowerShell: Copy-Item .env.example .env
```

Edita `.env` si quieres cambiar contraseñas, puertos o límites. Los valores
del archivo de ejemplo son solo para desarrollo y no deben reutilizarse en
producción.

### Arranque

```bash
make up-build   # primer arranque o después de cambiar Dockerfiles/dependencias
make up         # arranques posteriores
```

Los dos targets dejan `docker compose up` en primer plano. Para preparar la
base de datos desde cero:

```bash
docker compose up -d postgres redis
make migrate-up
make up-build
```

Servicios expuestos en desarrollo:

| Servicio | URL/puerto |
| --- | --- |
| Aplicación a través de Caddy | http://localhost |
| Frontend directo | http://localhost:3000 |
| Backend directo | http://localhost:8080 |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |

El backend, PostgreSQL y Redis están publicados en desarrollo para facilitar
las pruebas directas. En producción solo Caddy publica puertos.

### Detener y consultar el stack

```bash
make down
make ps
make logs
```

## Migraciones y datos

Las migraciones SQL versionadas viven en `backend/migrations/` y se ejecutan
con el servicio Docker de `golang-migrate`. `docker compose up` no aplica
migraciones automáticamente.

```bash
docker compose up -d postgres

make migrate-new name=add_example
make migrate-up
make migrate-up-one
make migrate-down-one
make migrate-down
make migrate-version
make migrate-force VERSION=3
```

El esquema actual incluye extensión `pgcrypto`, usuarios con estados y soft
delete, perfiles 1:1, fotos, índices de búsqueda, favoritos, likes, matches,
conversaciones y mensajes, bloqueos, reportes, roles `user/admin` y
consentimientos.

No se deben editar migraciones ya aplicadas: los cambios nuevos deben ir en
otra pareja `.up.sql`/`.down.sql`.

## Funcionalidad del backend

Todos los endpoints de dominio usan `/api/v1`. Los errores tienen la forma
`{"error":{"code":"...","message":"..."}}`. Las peticiones JSON rechazan
campos desconocidos y existe un límite global configurable mediante
`MAX_REQUEST_BODY_MB`.

### Salud

| Método | Ruta | Descripción |
| --- | --- | --- |
| GET | `/healthz` | Liveness para Docker; no comprueba dependencias. |
| GET | `/api/v1/health` | Readiness; comprueba PostgreSQL y Redis. |

### Autenticación y cuenta

Las rutas autenticadas usan la cookie httpOnly `session_id`, cuyo nombre se
puede cambiar con `SESSION_COOKIE_NAME`.

| Método | Ruta | Descripción |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | Registra una cuenta y abre sesión. Requiere `accepted_terms: true`. |
| POST | `/api/v1/auth/login` | Inicia sesión. |
| POST | `/api/v1/auth/logout` | Cierra la sesión actual. |
| GET | `/api/v1/auth/me` | Devuelve la cuenta autenticada. |
| POST | `/api/v1/auth/password/forgot` | Solicita recuperación sin enumerar cuentas. |
| POST | `/api/v1/auth/password/reset` | Consume un token y cambia la contraseña. |
| POST | `/api/v1/auth/email/verify` | Consume el token de verificación. |
| POST | `/api/v1/auth/email/resend` | Reenvía la verificación. |
| DELETE | `/api/v1/auth/account` | Soft delete tras confirmar la contraseña. |
| GET | `/api/v1/consents/me` | Historial propio de consentimientos. |
| POST | `/api/v1/contact` | Contacto público con rate limiting estricto. |

Con `EMAIL_DRIVER=noop`, los emails se escriben en los logs del backend:

```bash
docker compose logs -f backend
```

### Perfiles y fotos

| Método | Ruta | Descripción |
| --- | --- | --- |
| GET | `/api/v1/profiles/me` | Perfil propio. |
| POST | `/api/v1/profiles/me` | Crea el perfil propio. |
| PATCH | `/api/v1/profiles/me` | Actualización parcial; `null` borra un dato. |
| POST | `/api/v1/profiles/me/photos` | Sube JPEG, PNG o WebP de hasta 5 MB; máximo 6 fotos. |
| GET | `/api/v1/profiles/me/photos` | Lista las fotos propias. |
| GET | `/api/v1/profiles/me/photos/{id}/file` | Sirve una foto propia. |
| DELETE | `/api/v1/profiles/me/photos/{id}` | Elimina una foto. |
| GET | `/api/v1/profiles/{profileID}` | Perfil público de otra persona. |
| GET | `/api/v1/profiles/{profileID}/photos` | Fotos públicas. |
| GET | `/api/v1/profiles/{profileID}/photos/{photoID}/file` | Sirve una foto pública. |

Los campos opcionales no indicados se guardan como `NULL`, no como valores
inventados. Los perfiles de cuentas suspendidas, eliminadas o bloqueadas se
ocultan con un `404` genérico. El tipo de imagen se detecta por sus bytes, no
por el `Content-Type` declarado por el cliente.

### Búsqueda

`GET /api/v1/search/profiles` devuelve fichas resumidas de otros perfiles,
paginadas. Admite `gender`, `min_age`, `max_age`, `country`, `language`,
`relationship_goal`, `has_children`, `wants_children`, `interests`, `sort`,
`page` y `page_size` (máximo 50). Los valores de `relationship_goal` son
`casual`, `long_term`, `friendship`, `marriage` y `not_sure`.

Los filtros que exigen un dato excluyen perfiles que lo tienen en `NULL`; sin
filtro, esos perfiles siguen siendo visibles.

### Favoritos, likes, matches, mensajes, bloqueos y reportes

| Área | Rutas |
| --- | --- |
| Favoritos | `GET /api/v1/favorites`, `POST|DELETE /api/v1/favorites/{profileID}`, `GET /api/v1/favorites/{profileID}` |
| Likes | `POST|DELETE /api/v1/likes/{profileID}`, `GET /api/v1/likes/{profileID}`, `GET /api/v1/likes/sent`, `GET /api/v1/likes/received` |
| Matches | `GET /api/v1/matches`, `GET /api/v1/matches/{profileID}` |
| Mensajes | `POST /api/v1/messages/to/{profileID}`, `GET /api/v1/messages/conversations`, `GET|POST /api/v1/messages/conversations/{conversationID}/messages` |
| Bloqueos | `GET /api/v1/blocks`, `POST|DELETE /api/v1/blocks/{profileID}`, `GET /api/v1/blocks/{profileID}` |
| Reportes | `POST /api/v1/reports/{profileID}` |

Favoritos, likes, bloqueos y sus operaciones de estado son idempotentes. Dos
likes mutuos crean un match dentro de la misma transacción; quitar cualquiera
de los likes elimina el match. Las listas ocultan perfiles suspendidos,
eliminados o bloqueados, pero bloquear no borra el historial. Los mensajes
son 1:1 y tienen un máximo de 2000 caracteres. Al listar una conversación se
marcan como leídos los mensajes recibidos. Un bloqueo es mutuo a efectos de
búsqueda, favoritos, likes, matches, perfiles, conversaciones y nuevos
mensajes, pero no borra los datos históricos.

Los reportes aceptan `spam`, `fake_profile`, `harassment`,
`inappropriate_content`, `underage` u `other`, con una descripción opcional de
hasta 2000 caracteres.

### Administración

No existe un registro separado para administradores. Registra primero una
cuenta normal y promociónala:

```bash
make admin-promote email=persona@example.com
```

Las rutas siguientes requieren sesión y rol `admin`:

| Método | Ruta | Descripción |
| --- | --- | --- |
| GET | `/api/v1/admin/users?status=&page=&page_size=` | Lista usuarios. |
| POST | `/api/v1/admin/users/{userID}/suspend` | Suspende un usuario. |
| POST | `/api/v1/admin/users/{userID}/reactivate` | Reactiva un usuario. |
| GET | `/api/v1/admin/reports?status=&page=&page_size=` | Lista reportes. |
| POST | `/api/v1/admin/reports/{reportID}/resolve` | Marca `reviewed` o `dismissed`. |

Una suspensión invalida el acceso de inmediato, incluso para sesiones ya
abiertas. Un administrador no puede suspenderse a sí mismo.

## Frontend

La interfaz usa Next.js App Router, TypeScript y `frontend/lib/api.ts` como
cliente compartido. La cookie de sesión se envía automáticamente al usar
Caddy y el mismo origen.

| Ruta | Función |
| --- | --- |
| `/` | Página inicial con perfiles populares/paginados. |
| `/discover` | Resultados de búsqueda. |
| `/profiles/[id]` | Perfil público, fotos, favorito, bloqueo, reporte y primer mensaje. |
| `/profile` | Visualización del perfil propio. |
| `/profile/edit` | Creación, edición y gestión de fotos del perfil propio. |
| `/favorites` | Lista de favoritos. |
| `/likes` | Likes enviados y recibidos, paginados. |
| `/matches` | Matches paginados y enlace a conversaciones existentes. |
| `/messages` | Lista de conversaciones. |
| `/messages/[id]` | Hilo de conversación y respuesta. |
| `/blocked` | Lista y desbloqueo de perfiles. |
| `/admin` | Panel de usuarios y reportes para administradores. |
| `/login` | Inicio de sesión. |
| `/register` | Registro con aceptación de términos y privacidad. |
| `/settings` | Estado de cuenta, consentimientos, logout y eliminación de cuenta. |
| `/legal/terms` | Términos. |
| `/legal/privacy` | Política de privacidad. |
| `/legal/impressum` | Aviso legal. |
| `/legal/contact` | Página de contacto. |

Las páginas legales son plantillas con placeholders y deben revisarse con
asesoramiento legal antes de una publicación real. Los archivos de
traducciones `frontend/messages/es.json` y `en.json` existen como base, pero
no hay routing i18n activo todavía.

La navegación contiene enlaces a `/activity`, `/online` y `/new`, pero esas
páginas todavía no tienen implementación propia.

## Seguridad

- Rate limiting global por IP y un límite más estricto para autenticación y
  contacto, ambos configurables con Redis.
- Cabeceras `X-Content-Type-Options`, `X-Frame-Options` y `Referrer-Policy`.
- Límite global de cuerpo de petición y límites específicos por endpoint.
- Cookies de sesión httpOnly y `SameSite=Lax`; en producción se marca
  `Secure` automáticamente.
- Logs con la IP de origen de cada petición.

El rate limiting funciona fail-open si Redis no está disponible. No existe
protección CSRF explícita con tokens; `SameSite=Lax` cubre el caso común, pero
no sustituye una solución CSRF completa. En desarrollo el backend está
publicado directamente en el host, por lo que `X-Forwarded-For` puede ser
falseado; en producción el tráfico externo debe entrar solo por Caddy.

## Tests y calidad

```bash
make test              # unitarios; no requiere Postgres ni Redis
make test-integration  # levanta Postgres/Redis; requiere migraciones aplicadas
make test-all          # unitarios + integración
make smoke-test        # stack Docker, migraciones y endpoints de salud
```

Los tests de integración usan la base de datos de desarrollo y la build tag
`integration`; no existe una base separada en V1. En una base vacía hay que
ejecutar `make migrate-up` antes de `make test-integration`. El smoke test
requiere shell POSIX, `curl` y utilidades Unix.

## Producción

La producción usa `docker-compose.yml` junto con
`docker-compose.prod.yml`. Requiere Docker Compose 2.24 o posterior por los
tags `!override` y `!reset`.

1. Instala Docker y apunta el DNS del dominio a la VM.
2. Clona el repositorio y copia `.env.production.example` como `.env`.
3. Rellena dominio, contraseñas, SMTP y credenciales S3.
4. Valida y construye el stack:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml config
make prod-build
make prod-up
```

5. Aplica las migraciones:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm migrate \
  -path=/migrations \
  -database "postgres://USUARIO:PASSWORD@postgres:5432/BASE_DE_DATOS?sslmode=disable" up
```

6. Registra la primera cuenta y promuévela. En desarrollo basta con
   `make admin-promote email=persona@example.com`; con el stack de producción
   combinado usa:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml run --rm \
  --entrypoint go backend run ./cmd/promote-admin \
  --email=persona@example.com
```

En producción solo Caddy publica `80` y `443`. Caddy obtiene y renueva el
certificado de Let's Encrypt mediante `Caddyfile.prod`. El backend y frontend
usan imágenes compiladas, el email usa SMTP y las fotos usan un proveedor
compatible con S3. El bucket debe existir y ser accesible al arrancar.

```bash
make prod-logs
make prod-down
```

### Backups y restauración

```bash
make backup
make restore file=backups/dating_platform_YYYYMMDDTHHMMSSZ.sql.gz
```

El backup se guarda en `./backups` como `.sql.gz`. La restauración pide
escribir `si` y sobrescribe la base de datos actual. Los scripts dependen de
variables de entorno; `make` carga `.env`, pero si se ejecutan directamente
hay que exportar las variables necesarias. No se incluye un cron automático.

## Variables de entorno

Las listas completas y comentadas están en [`.env.example`](.env.example) y
[`.env.production.example`](.env.production.example). Las principales áreas
son:

- PostgreSQL y Redis;
- puertos y nivel de logs;
- duración de sesiones y tokens;
- driver de email (`noop` o `smtp`);
- límites y rate limiting;
- URL pública e interna del frontend;
- almacenamiento local o S3.

`NEXT_PUBLIC_API_URL` es la URL que usa el navegador; `INTERNAL_API_URL` es la
URL interna para llamadas desde el servidor de Next.js. `REDIS_PORT` configura
al backend, mientras que el puerto publicado por Compose sigue fijo en
`6379:6379`.

## Estructura

```text
.
├── docker-compose.yml          # Stack de desarrollo
├── docker-compose.prod.yml     # Override de producción
├── Caddyfile                   # Proxy de desarrollo
├── Caddyfile.prod              # Proxy HTTPS de producción
├── Makefile                    # Arranque, migraciones, tests y operaciones
├── seed_profiles.ps1           # Carga local de perfiles de demostración
├── backend/
│   ├── cmd/api/                # Servidor HTTP
│   ├── cmd/promote-admin/      # Promoción de cuentas a admin
│   ├── internal/               # Módulos de dominio, infraestructura y tests
│   └── migrations/             # Esquema SQL versionado
├── frontend/
│   ├── app/                    # Rutas y páginas Next.js
│   ├── lib/                    # Cliente de API
│   └── messages/               # Base de traducciones
└── scripts/
    ├── smoke-test.sh           # Smoke test del stack
    ├── backup.sh               # Backup de PostgreSQL
    └── restore.sh              # Restauración de backups
```

## Datos de demostración

`seed_profiles.ps1` intenta registrar o iniciar sesión con 50 cuentas,
crear sus perfiles y subir fotos a `http://localhost`. Requiere PowerShell,
`curl.exe` y fotos `.jpg` en:

```text
C:\Users\Victor\Downloads\fotos\Nueva carpeta (4)
```

El script usa credenciales de demostración y está pensado solo para un
entorno local. Actualmente incluye el valor `short_term` en sus datos, pero
el backend acepta `casual`, `long_term`, `friendship`, `marriage` y `not_sure`;
ese valor debe corregirse antes de usar el script para una carga completa.

## Limitaciones conocidas

- No hay tokens CSRF explícitos.
- `/activity`, `/online` y `/new` aparecen en la navegación, pero aún no son
  páginas implementadas.
- El routing i18n no está activado.
- El almacenamiento local con el volumen `backend_uploads` es apropiado para
  desarrollo y una sola instancia, no para varias instancias.
- `INTERNAL_API_URL` está preparado en Compose, aunque las vistas actuales
  usan principalmente componentes cliente y `NEXT_PUBLIC_API_URL`.
- El entorno de integración reutiliza la base de datos de desarrollo.