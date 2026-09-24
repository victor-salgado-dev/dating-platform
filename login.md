# Autenticación (login)

Documento de referencia sobre cómo funciona el sistema de autenticación del
proyecto, cómo probarlo a mano y qué cambios se contemplan a futuro.

Estado actual: **implementado y en uso**. La verificación de email **no bloquea**
el acceso (ver sección "Verificación de email").

---

## 1. Dónde vive el código

Todo el módulo está en `backend/internal/auth/`:

| Fichero | Responsabilidad |
| --- | --- |
| `service.go` | Casos de uso: register, login, logout, sesión actual, reset de contraseña, verificación de email, borrado de cuenta. |
| `handler.go` | Capa HTTP: DTOs, traducción de errores a códigos y gestión de la cookie de sesión. |
| `password.go` | Hash y verificación de contraseñas, `MinPasswordLength`. |
| `session.go` | `Session` y la interfaz `SessionStore`. |
| `redis_session_store.go` | Implementación de `SessionStore` sobre Redis. |
| `token.go` | Generación y hash de tokens de un solo uso (verificación / reset). |
| `middleware.go` | `ContextWithUserID` / `UserIDFromContext` y el middleware que resuelve la sesión. |

La configuración asociada está en `backend/internal/config/config.go` (`AuthConfig`):
`SessionTTL`, `EmailVerificationTTL`, `PasswordResetTTL`, `CookieSecure`,
`CookieName`.

---

## 2. Endpoints

Todos cuelgan de `/api/v1/auth` (detrás de Caddy, que además enruta
`/api/v1/*` hacia el backend).

| Método | Ruta | Autenticado | Descripción |
| --- | --- | --- | --- |
| POST | `/register` | No | Crea la cuenta, registra el consentimiento legal y abre sesión. |
| POST | `/login` | No | Valida credenciales y abre sesión. |
| POST | `/logout` | No (usa la cookie) | Invalida la sesión actual. Idempotente. |
| GET | `/me` | Sí | Devuelve la cuenta de la sesión. |
| POST | `/password/forgot` | No | Genera token de reset y envía email. Siempre responde 202. |
| POST | `/password/reset` | No | Consume el token y fija contraseña nueva. |
| POST | `/email/verify` | No | Consume el token y marca el email como verificado. |
| POST | `/email/resend` | Sí | Reenvía el email de verificación de la cuenta autenticada. |
| DELETE | `/account` | Sí | Soft delete de la cuenta tras confirmar la contraseña. |

Las respuestas de error tienen siempre la forma:

    {"error":{"code":"...","message":"..."}}

---

## 3. Flujo de login

1. El cliente hace `POST /api/v1/auth/login` con `{"email","password"}`.
2. `Service.Login`:
   - normaliza el email (trim + minúsculas) y valida el formato; si no es válido
     devuelve `ErrInvalidCredentials` (no `invalid_email`, para no delatar nada);
   - busca el usuario por email; si no existe → `ErrInvalidCredentials`;
   - verifica la contraseña; si no coincide → `ErrInvalidCredentials`;
   - si la cuenta está suspendida → `ErrAccountSuspended`;
   - crea una **sesión nueva** (`SessionStore.Create`) y devuelve el token en claro.
3. `Handler.Login` coloca la cookie de sesión y responde `200` con los datos
   básicos de la cuenta (`id`, `email`, `email_verified`, `status`, `created_at`).

Puntos de diseño:

- **Mismo error para email inexistente y contraseña incorrecta** (`invalid_credentials`),
  para dificultar la enumeración de cuentas registradas.
- La suspensión **corta el acceso de inmediato**: además de impedir logins nuevos,
  `CurrentUser` vuelve a comprobar el estado en cada petición y devuelve
  `ErrAccountSuspended` si la cuenta fue suspendida con la sesión ya abierta.

---

## 4. Sesión y cookie

- La sesión se persiste en **Redis** (`RedisSessionStore`), no en Postgres.
- El token que viaja en la cookie **no se guarda en claro** en Redis: se guarda
  su hash (`token.go` → `hashToken`). La clave de Redis se construye con un
  prefijo (`session:`) más ese hash.
- La cookie se llama según `AuthConfig.CookieName` (en la práctica `session_id`) y
  se emite con:
  - `Path=/`
  - `HttpOnly` (no accesible desde JavaScript)
  - `Secure` según `AuthConfig.CookieSecure`
  - `SameSite=Lax`
  - `Max-Age` = `SessionTTL`
- La duración de la sesión es `AuthConfig.SessionTTL`.
- **Logout** borra la entrada de Redis y manda una cookie vacía con `Max-Age=-1`.
  Es idempotente: llamarlo con una sesión ya inexistente no es error.

Las peticiones autenticadas pasan por el middleware, que lee la cookie, resuelve
la sesión (`SessionStore.Get`), coloca el `userID` en el contexto
(`ContextWithUserID`) y deja que los handlers lo lean con `UserIDFromContext`.

---

## 5. Contraseñas

- Longitud mínima: `auth.MinPasswordLength` (8). La validación la hace quien
  llama; el paquete expone la constante y el helper de hash.
- Almacenamiento: hash **con sal** (`password.go`). Nunca se guarda la contraseña
  en claro.
- `VerifyPassword` compara en tiempo (suficientemente) constante.
- `DeleteAccount` exige reconfirmar la contraseña antes del soft delete.

---

## 6. Tokens de un solo uso (verificación y reset)

Verificación de email y reset de contraseña comparten el mismo mecanismo
(`TokenStore`), con distinto `purpose`:

- `PurposeEmailVerification`
- `PurposePasswordReset`

`Create(ctx, purpose, userID, ttl)` genera el token; `Consume(ctx, purpose, token)`
lo valida y lo **consume** (un solo uso). Los TTL salen de `EmailVerificationTTL`
y `PasswordResetTTL`.

En el estado actual, el envío del email es **best-effort**:

- En el registro, un fallo al enviar el email de verificación **no impide** crear
  la cuenta (se registra el error y se continúa).
- En `password/forgot`, tanto si la cuenta existe como si no, la respuesta es
  202 (`{"status":"if_account_exists_email_sent"}`), para no filtrar qué emails
  están registrados.

---

## 7. Códigos de error

Los más relevantes de la capa HTTP:

| Código | HTTP | Dónde |
| --- | --- | --- |
| `invalid_body` | 400 | Cuerpo malformado en cualquier endpoint. |
| `invalid_email` | 400 | Formato de email inválido en register. |
| `weak_password` | 400 | Contraseña por debajo de `MinPasswordLength`. |
| `terms_not_accepted` | 400 | Registro sin aceptar términos. |
| `email_taken` | 409 | Registro con email ya existente. |
| `invalid_credentials` | 401 / 403 | Login (401) o borrado de cuenta (403). |
| `account_suspended` | 403 | Cuenta suspendida. |
| `invalid_token` | 400 | Token de verificación o reset inválido/caducado. |
| `unauthenticated` | 401 | Falta cookie de sesión o sesión no válida. |
| `internal_error` | 500 | Fallo inesperado. |

---

## 8. Rate limiting

Definido en `SecurityConfig`:

- **Límite global** aplicado a todas las peticiones (defensa básica frente a abuso).
- **Límite estricto por IP** aplicado solo a los endpoints sensibles de auth:
  `register`, `login`, `password/forgot`, `password/reset`, `email/verify`,
  `email/resend`.

Un exceso devuelve `429`. Nota práctica: al probar a mano, una racha de intentos
seguidos puede disparar el límite; conviene esperar antes de repetir.

---

## 9. Verificación de email: estado actual

**Decisión vigente (opción C): el email no verificado NO bloquea el uso.**

- `Register` y `Login` crean sesión con independencia de `email_verified`.
- La respuesta de `/me` y de login expone `email_verified` para que el cliente
  pueda mostrar avisos, pero no condiciona el acceso.

Para reenviar la verificación, hoy `POST /email/resend` **exige sesión**. Esto es
coherente con la opción C, pero es un punto a revisar si algún día se bloquea el
login de cuentas no verificadas (ver siguiente sección).

---

## 10. Comprobación manual (curl)

Útil para validar el flujo completo sin navegador. En Windows/PowerShell hay que
usar `curl.exe` (el alias `curl` apunta a `Invoke-WebRequest` y no entiende las
opciones de curl real).

1. Cuerpo del login en un fichero (evita problemas de comillas en PowerShell):

       Set-Content -Path login.json -Value '{"email":"demo@demo.com","password":"DemoPass123!"}' -Encoding ascii -NoNewline

2. Login y guardado de cookies:

       curl.exe -i -c cookies.txt -X POST http://localhost/api/v1/auth/login -H "Content-Type: application/json" --data-binary=@login.json

   Esperado: `200` con cabecera `Set-Cookie: session_id=...`.

3. Comprobar que la sesión se lee:

       curl.exe -i -b cookies.txt http://localhost/api/v1/profiles/me

   Esperado: `200` con el perfil.

4. Logout y re-comprobación:

       curl.exe -i -b cookies.txt -c cookies.txt -X POST http://localhost/api/v1/auth/logout
       curl.exe -i -b cookies.txt http://localhost/api/v1/profiles/me

   Esperado: el segundo comando devuelve `401 unauthenticated`.

`cookies.txt` y `login.json` son temporales y están en `.gitignore`. No deben
versionarse (el segundo contiene la contraseña en claro).

---

## 11. Cambios futuros contemplados

Nada de esto está implementado; se listan como opciones sobre la mesa.

### 11.1 Endurecer la verificación de email

- **Opción A — bloquear login sin verificar.** `Login` devolvería `403`
  `email_not_verified`. Implica repensar el reenvío: hoy `/email/resend` exige
  sesión, y con A el usuario no podría entrar a reenviar. Alternativas:
  - añadir un endpoint de reenvío **sin sesión** que reciba el email en el cuerpo
    (con respuesta neutra tipo 202 para no filtrar existencia), o
  - permitir un login "limitado" cuyo único destino válido sea reenviar/verificar.
- **Opción B — permitir login, restringir la app.** Un middleware bloquearía los
  endpoints "de uso" mientras `email_verified` sea falso, dejando libres los de
  auth (login, reenvío). No tocaría `service.go`.

### 11.2 Reenvío de verificación sin sesión

Cambio necesario si se adopta la opción A. Implica nuevo endpoint y ajuste del
rate limiting para ese caso.

### 11.3 Verificación por enlace en lugar de código

Hoy se envía un **código**. Un enlace con `GET`/página de confirmación mejora la
UX y evita copiar/pegar, a cambio de gestionar la caducidad y el estado en la
página.

### 11.4 Rotación y revocación de sesiones

- Rotar el token de sesión tras cambios sensibles (p. ej. cambio de contraseña).
- Revocar **todas** las sesiones de un usuario al cambiar la contraseña o al ser
  suspendido. Hoy la suspensión se corta en cada petición vía `CurrentUser`, pero
  las sesiones siguen existiendo en Redis hasta su TTL.
- Listado de sesiones activas y "cerrar sesión en todos los dispositivos".

### 11.5 Rate limiting por cuenta

Complementar el límite por IP con un contador por cuenta/email para frenar
ataques distribuidos contra una misma cuenta.

### 11.6 Segundo factor (2FA)

TOTP o similar, a valorar para cuentas con privilegios (por ejemplo, admin).

### 11.7 Registro/auditoría de eventos de auth

Traza de logins, logouts, resets y fallos, para detección de abuso y soporte.
