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

## Seguridad y hardening (Fase 11)

Añadido transversal a toda la API existente, sin cambiar ningún contrato:

- **Rate limiting (Redis, ventana fija).** Dos niveles, por IP:
  - Global sobre toda la API: `RATE_LIMIT_GLOBAL_MAX` peticiones cada
    `RATE_LIMIT_GLOBAL_WINDOW_SECONDS` (300/300s por defecto).
  - Estricto sobre los endpoints de auth más sensibles a fuerza bruta o
    spam (`register`, `login`, `password/forgot`, `password/reset`,
    `email/verify`, `email/resend`): `RATE_LIMIT_AUTH_MAX` cada
    `RATE_LIMIT_AUTH_WINDOW_SECONDS` (20/900s por defecto). Al superarse,
    `429 rate_limited`. Si Redis fallara, el límite se salta (fail-open):
    es una capa de defensa, no debe poder tumbar la API entera.
- **Cabeceras de seguridad** en toda respuesta: `X-Content-Type-Options: nosniff`,
  `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`.
- **Límite global de tamaño de petición** (`MAX_REQUEST_BODY_MB`, 10 MB por
  defecto): backstop de memoria/DoS: cada endpoint sigue aplicando además
  sus propios límites semánticos (2000 caracteres en un mensaje, 5 MB en
  una foto...).
- **Seguridad de imágenes reforzada.** El `Content-Type` que declara el
  formulario de subida ya NO es de fiar (se falsifica con solo cambiar
  una cabecera): ahora se inspeccionan los primeros bytes reales del
  fichero (`http.DetectContentType`) y se guarda/sirve siempre ese tipo
  detectado, nunca el declarado. Un desajuste se registra en logs pero
  no bloquea por sí solo — lo que sí bloquea es que el contenido real no
  sea JPEG/PNG/WebP.
- **Suspensión efectiva de inmediato** (cerrada en la Fase 10, forma
  parte del mismo endurecimiento): una cuenta suspendida pierde el
  acceso aunque tenga una sesión ya abierta.
- **Logs** incluyen ahora la IP de origen de cada petición.

**Limitaciones conocidas, documentadas a propósito en vez de resueltas
a medias:**
- No hay protección CSRF explícita (tokens de formulario). La cookie de
  sesión usa `SameSite=Lax`, que ya bloquea el caso más común (POST
  cross-site), pero no es una solución completa. Añadir CSRF tokens
  reales es una mejora futura razonable si se necesitara reforzar esto.
- `docker-compose.yml` publica el puerto del backend directamente
  (`8080`) para poder probarlo sin pasar por Caddy en desarrollo. Eso
  significa que alguien podría saltarse Caddy y falsificar
  `X-Forwarded-For` para intentar evadir el rate limiting por IP. En
  producción (Fase 14) el backend no debería exponerse fuera de la red
  de Docker: solo Caddy debería poder llegar a él.

## Tests (Fase 12)

Dos niveles, deliberadamente separados:

### Unitarios — rápidos, sin red

Lógica pura (validaciones, cálculo de edad, hashing, normalización de
email, construcción/validación de parámetros de búsqueda, helpers
HTTP). No tocan Postgres ni Redis, así que corren siempre, incluso sin
`docker compose up`.

```bash
make test
```

### Integración — requieren Postgres y Redis reales

Usan la build tag `integration` a propósito: `go test ./...` normal
(sin esa tag) ni siquiera los compila, así que nunca fallan por
sorpresa si alguien olvida levantar la infraestructura. Cubren:

- **Autenticación** (`backend/internal/auth`): registro, login, logout,
  que una sesión cerrada no siga siendo válida pero otras del mismo
  usuario sí, reset de contraseña con consumo de un solo uso del token,
  verificación de email.
- **Búsqueda** (`backend/internal/search`): sobre todo, **la regla de
  los datos faltantes contra PostgreSQL real** — un perfil que no
  indicó un dato no aparece cuando la búsqueda exige ese dato, pero sí
  aparece cuando no lo exige. También que nadie se encuentra a sí mismo.
- **Rate limiting** (`backend/internal/ratelimit`): el límite se agota
  a la cuenta esperada y cada key es independiente.
- **Integración básica** (`backend/internal/server`): levanta el
  router HTTP real (el mismo de `cmd/api`) y ejercita un flujo de
  extremo a extremo — registrar dos cuentas, crear sus perfiles, y
  comprobar que cada una encuentra a la otra en `/search/profiles` pero
  nunca a sí misma — más los endpoints de salud.

Usan la misma base de datos de desarrollo (vía `config.Load()`, igual
que la app): no hay una base de datos de test separada en V1. Es
aceptable para el flujo de desarrollo actual; si hiciera falta
aislarlos (p. ej. en CI), lo natural sería una base de datos de test
dedicada, no antes.

```bash
make test-integration   # levanta postgres/redis si hace falta
make test-all           # unitarios + integración
```

### Smoke test del stack Docker

No es un test de Go: comprueba que `docker compose up` deja el stack
en un estado realmente usable (contenedores arriba, migraciones
aplicables, endpoints de salud respondiendo) — el tipo de fallo que un
test de Go no ve.

```bash
make smoke-test
```

## Legal y privacidad (Fase 13)

### Consentimiento en el registro

`POST /api/v1/auth/register` ahora exige `accepted_terms: true` en el
cuerpo — si falta o es `false`, `400 terms_not_accepted`, **antes** de
crear la cuenta. Al aceptar, se registran dos consentimientos
separados (`terms` y `privacy_policy`, misma marca de tiempo): son
documentos distintos y no se asume que aceptar uno implica el otro
(sección 14). Cada fila guarda versión y fecha/hora exactas; nunca se
sobrescribe un consentimiento ya dado, solo se añaden nuevos si
volviera a hacer falta re-consentir una versión posterior.

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/api/v1/consents/me` | Tu propio historial de consentimientos (transparencia) |

### Formulario de contacto

`POST /api/v1/contact` es la única ruta de escritura de toda la API
que **no requiere sesión** (alguien sin cuenta también tiene que poder
escribir). Por eso lleva el mismo rate limit estricto que login/registro,
para no convertirse en un vector de spam. Cuerpo: `{"name", "email", "message"}`;
se reenvía por email (abstracción `email.Sender`, driver `noop` en dev)
a `CONTACT_INBOX_EMAIL`.

### Páginas legales (frontend)

`/legal/terms`, `/legal/privacy`, `/legal/impressum`, `/legal/contact`.
**Son plantillas de contenido**, con placeholders (`[NOMBRE DE LA
EMPRESA]`, `[DIRECCIÓN]`, etc.) y un aviso visible en cada página: no
son asesoramiento legal y deben revisarse con un abogado antes de
publicarse de verdad, adaptadas a la jurisdicción y la empresa reales.

### Login, registro y cuenta (frontend)

Hasta esta fase, todo el frontend asumía una sesión ya iniciada por
`curl`. Como la casilla de consentimiento solo tiene sentido dentro de
un formulario de registro real, esta fase añade las páginas que
faltaban para completar el ciclo:

- `/register` — email, contraseña y la casilla de aceptación (enlaza a
  `/legal/terms` y `/legal/privacy`); el botón de enviar está
  deshabilitado hasta marcarla.
- `/login` — email y contraseña.
- `/account` — email, estado de verificación, cerrar sesión, historial
  de consentimientos, y **eliminar cuenta** (pide confirmar la
  contraseña, con aviso de que es irreversible).

**Nota de alcance:** estas páginas cierran el ciclo de cuenta/legal que
pedía esta fase. No incluyen un formulario de creación/edición de
perfil en el frontend (la Fase 4 solo pedía la API): sigue siendo
razonable añadirlo más adelante, pero no lo pide ninguna fase todavía,
así que no se ha construido por adelantado.

## Producción (Fase 14)

Última fase del roadmap original. Todo lo anterior seguía funcionando
igual en desarrollo (`docker compose up`); esto es exclusivamente
infraestructura nueva para desplegar de verdad.

### Qué cambia respecto a desarrollo

| | Desarrollo | Producción |
|---|---|---|
| Backend/frontend | `Dockerfile.dev`, código montado, `go run`/`next dev` | `Dockerfile`, build compilado, sin código fuente en la imagen |
| Puertos | postgres/redis/backend publicados en el host, para poder probarlos sueltos | **solo Caddy** publica 80/443; el resto solo habla por la red interna de Docker |
| Caddy | `:80`, sin TLS | dominio real, HTTPS automático (Let's Encrypt) vía `Caddyfile.prod` |
| Storage | `STORAGE_DRIVER=local` (disco del contenedor) | `STORAGE_DRIVER=s3` (cualquier proveedor compatible con S3) |
| Email | `EMAIL_DRIVER=noop` (logs) | `EMAIL_DRIVER=smtp` (envío real) |
| Logs | `LOG_LEVEL=debug` | `LOG_LEVEL=info` (configurable; antes de esta fase la variable existía pero no se aplicaba de verdad al logger — quedó corregido aquí) |

### Desplegar

Pensado para el caso más simple y barato: **una sola VM con Docker**
(coincide con la prioridad de coste bajo de la sección 1). Nada de esto
impide migrar después a un orquestador más sofisticado si hiciera
falta, pero no se construye por adelantado.

```bash
# En el servidor, con Docker y el plugin compose instalados:
git clone <tu-repo> && cd dating-platform

cp .env.production.example .env
# Rellena TODOS los [CAMBIAR]: dominio, contraseñas, SMTP, bucket S3...

# Apunta el DNS de tu dominio (registro A) a la IP del servidor ANTES
# de arrancar Caddy, o la emisión del certificado HTTPS fallará.

# (Opcional pero recomendable) comprobar cómo se combina el override
# antes de construir nada:
#   docker compose -f docker-compose.yml -f docker-compose.prod.yml config

make prod-build
make prod-up

# Migraciones (igual que en desarrollo, contra la base de datos real):
docker compose run --rm migrate -path=/migrations \
  -database "postgres://USUARIO:PASSWORD@postgres:5432/BASE_DE_DATOS?sslmode=disable" up

# Primer admin (igual que en desarrollo):
make admin-promote email=tu-email@tu-dominio.com
```

Caddy obtiene y renueva el certificado HTTPS solo; no hace falta
certbot ni configuración de TLS manual.

### Backups

```bash
make backup                          # -> ./backups/dating_platform_<fecha>.sql.gz
make restore file=backups/xxx.sql.gz # pide confirmación explícita
```

`pg_dump`/`psql` corren dentro del contenedor de `postgres` (misma
versión que la base de datos real): no hace falta instalar herramientas
de PostgreSQL en el servidor. No hay una tarea programada (cron)
incluida — en la mayoría de VMs basta un `cron` del sistema operativo
llamando a `make backup`; automatizarlo dentro de la propia app sería
sobrearquitectura para V1.

### Almacenamiento S3-compatible

`STORAGE_DRIVER=s3` (además de `local`, que sigue siendo el de
desarrollo). Funciona con cualquier proveedor compatible con la API de
S3 — AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2... — vía
[`minio-go`](https://github.com/minio/minio-go), un cliente ligero, no
solo para MinIO pese al nombre. El bucket debe existir de antemano: el
backend comprueba que es accesible al arrancar y falla pronto (antes de
aceptar tráfico) si no. Ver variables `STORAGE_S3_*` en
`.env.production.example`.

### Email real (SMTP)

`EMAIL_DRIVER=smtp` usa `net/smtp` de la librería estándar de Go (sin
dependencias nuevas): compatible con cualquier proveedor transaccional
que hable SMTP con STARTTLS en el puerto 587 (SendGrid, Mailgun,
Postmark, Amazon SES...). Ver variables `SMTP_*`.

### Health checks

Las imágenes de producción (`backend/Dockerfile`, `frontend/Dockerfile`)
incluyen `HEALTHCHECK` propio (antes solo lo tenían postgres/redis):
Docker puede saber si el backend o el frontend dejaron de responder,
no solo si el proceso sigue vivo.

## Variables de entorno

Ver [`.env.example`](./.env.example) (desarrollo) y
[`.env.production.example`](./.env.production.example) (producción,
Fase 14) para la lista completa y comentada. Nunca se sube `.env` (con
secretos reales) a Git.

## Estructura del proyecto

```
.
├── docker-compose.yml       # Orquestación (desarrollo)
├── docker-compose.prod.yml   # Override de producción (Fase 14)
├── Caddyfile                 # Reverse proxy de desarrollo (:80, sin TLS)
├── Caddyfile.prod              # Reverse proxy de producción (dominio real, HTTPS automático)
├── Makefile                  # Comandos de arranque, migraciones, tests, producción, backups
├── .env.example
├── .env.production.example    # Plantilla de variables de producción (Fase 14)
├── backend/                  # Monolito modular en Go
│   ├── cmd/
│   │   ├── api/                # Punto de entrada del servidor
│   │   └── promote-admin/       # CLI para promocionar una cuenta a admin
│   ├── Dockerfile.dev           # Imagen de desarrollo (go run, código montado)
│   ├── Dockerfile                # Imagen de producción (build compilado, Fase 14)
│   ├── internal/
│   │   ├── config/             # Lectura de variables de entorno
│   │   ├── db/                  # Pool de conexión a PostgreSQL
│   │   ├── redisclient/         # Cliente Redis
│   │   ├── server/              # Router HTTP y middlewares
│   │   ├── httpx/                # Helpers de respuesta/error JSON compartidos
│   │   ├── health/              # Endpoints de liveness/readiness
│   │   ├── users/               # Dominio de cuenta: modelo + repositorio
│   │   ├── auth/                 # Registro, login, sesiones, tokens, email de verificación/reset
│   │   ├── email/                # Abstracción de email (driver "noop" en dev, "smtp" en producción)
│   │   ├── profiles/              # Perfil propio: datos estructurados + fotos
│   │   ├── search/                 # Búsqueda de perfiles con filtros estructurados
│   │   ├── favorites/               # Favoritos: añadir, eliminar, listar
│   │   ├── messaging/                # Conversaciones y mensajes 1:1
│   │   ├── blocking/                  # Bloqueo entre usuarios y reglas de visibilidad
│   │   ├── reports/                   # Creación de reportes
│   │   ├── admin/                     # Panel de administración (Fase 10)
│   │   ├── ratelimit/                  # Rate limiting en Redis (Fase 11)
│   │   ├── testutil/                    # Helpers para tests de integración (Fase 12)
│   │   ├── consent/                     # Registro de aceptación de Términos/Privacidad (Fase 13)
│   │   ├── contact/                      # Formulario de contacto público (Fase 13)
│   │   └── storage/               # Abstracción de storage (driver "local" en dev, "s3" en producción)
│   └── migrations/            # Migraciones SQL versionadas
└── frontend/                 # Next.js (App Router, TypeScript)
    ├── Dockerfile.dev           # Imagen de desarrollo (next dev, código montado)
    ├── Dockerfile                # Imagen de producción (build standalone, Fase 14)
    ├── app/
    │   ├── discover/           # Resultados de búsqueda (Fase 6)
    │   ├── profiles/[id]/      # Página de perfil público (Fase 6)
    │   ├── favorites/          # Lista de favoritos (Fase 7)
    │   ├── messages/           # Conversaciones y mensajes (Fase 8)
    │   ├── blocked/            # Perfiles bloqueados (Fase 9)
    │   ├── admin/              # Panel de administración (Fase 10)
    │   ├── login/              # Inicio de sesión (Fase 13)
    │   ├── register/           # Registro, con casilla de consentimiento (Fase 13)
    │   ├── account/            # Cuenta: cerrar sesión, consentimientos, eliminar cuenta (Fase 13)
    │   └── legal/              # Términos, Privacidad, Aviso legal, Contacto (Fase 13)
    ├── lib/                    # Helpers compartidos (cliente de API)
    └── messages/               # Esqueleto i18n (es/en), aún sin enrutar

scripts/
├── smoke-test.sh             # Comprobación de que el stack Docker arranca sano (Fase 12)
├── backup.sh                  # Backup de PostgreSQL (Fase 14)
└── restore.sh                 # Restauración de un backup (Fase 14)
```

## Roadmap de fases

El proyecto se construye fase a fase. Ver el documento de especificación
del proyecto para el listado completo (Fase 1 a Fase 14). No se implementan
fases futuras por adelantado.
