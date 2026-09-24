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

