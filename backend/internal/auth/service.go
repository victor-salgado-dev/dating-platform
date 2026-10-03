package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/email"
	"dating-platform/backend/internal/users"
)

// emailPattern es una validación de formato deliberadamente simple.
// No pretende cubrir el RFC 5322 completo; su objetivo es rechazar
// entradas obviamente inválidas antes de tocar la base de datos.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// dummyPasswordHash es un hash bcrypt válido, con el coste por defecto (el mismo
// que usa HashPassword), de una contraseña aleatoria que se descartó. Login lo
// compara cuando el email no existe, para que ese camino cueste lo mismo que una
// contraseña incorrecta: sin esto, la respuesta tarda microsegundos en vez de
// decenas de milisegundos y el tiempo delata qué emails están registrados.
const dummyPasswordHash = "$2a$10$i2nfzpT3SmFFSxWYxwHGzO4SwtWAFjY0y7dD/0IvEFKLS5d1uL4P2"

// ConsentRecorder es la interfaz mínima que auth necesita del módulo
// consent: persistir que se aceptaron los documentos legales al
// registrarse. auth valida el booleano accepted por su cuenta (ver
// Register) y solo delega en esta interfaz el efecto de guardarlo;
// consent.Service la satisface sin que ninguno de los dos paquetes
// importe al otro más que en este único punto.
type ConsentRecorder interface {
	RecordRegistrationConsents(ctx context.Context, userID uuid.UUID, accepted bool) error
}

// Service implementa los casos de uso de autenticación. Depende de
// interfaces (users.Repository, SessionStore, TokenStore, email.Sender),
// nunca de implementaciones concretas, para poder sustituirlas en tests.
type Service struct {
	users    users.Repository
	sessions SessionStore
	tokens   TokenStore
	sender   email.Sender
	consents ConsentRecorder

	sessionTTL time.Duration
	verifyTTL  time.Duration
	resetTTL   time.Duration
}

func NewService(
	usersRepo users.Repository,
	sessions SessionStore,
	tokens TokenStore,
	sender email.Sender,
	consents ConsentRecorder,
	sessionTTL, verifyTTL, resetTTL time.Duration,
) *Service {
	return &Service{
		users:      usersRepo,
		sessions:   sessions,
		tokens:     tokens,
		sender:     sender,
		consents:   consents,
		sessionTTL: sessionTTL,
		verifyTTL:  verifyTTL,
		resetTTL:   resetTTL,
	}
}

// Register crea una cuenta nueva, registra la aceptación de Términos y
// Política de Privacidad (sección 14: no se asume, tiene que venir
// marcada explícitamente), dispara el email de verificación
// (best-effort: un fallo de envío no impide el registro) y abre sesión
// inmediatamente. Devuelve ErrTermsNotAccepted, ErrWeakPassword,
// users.ErrDuplicateEmail, o un error de validación de email envuelto.
func (s *Service) Register(ctx context.Context, rawEmail, password string, acceptedTerms bool) (*users.User, string, error) {
	normEmail, err := normalizeEmail(rawEmail)
	if err != nil {
		return nil, "", err
	}
	if len(password) < MinPasswordLength {
		return nil, "", ErrWeakPassword
	}
	if !acceptedTerms {
		return nil, "", ErrTermsNotAccepted
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	u := &users.User{Email: normEmail, PasswordHash: hash}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err // p.ej. users.ErrDuplicateEmail
	}

	if err := s.consents.RecordRegistrationConsents(ctx, u.ID, acceptedTerms); err != nil {
		// La cuenta ya existe en este punto (acceptedTerms es true, así
		// que esto solo podría fallar por un problema real de base de
		// datos, no de validación). No revertimos la creación: no hay
		// una transacción cruzando ambos repositorios en V1. Se registra
		// para poder detectarlo, y se corta el registro en vez de dejar
		// pasar una cuenta sin consentimiento registrado.
		slog.Error("no se pudo registrar el consentimiento en el alta", "user_id", u.ID, "error", err)
		return nil, "", err
	}

	s.sendVerificationEmailBestEffort(ctx, u)

	token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return nil, "", err
	}

	s.touchAsync(u.ID)
	return u, token, nil
}

// Login verifica credenciales y abre una sesión nueva.
// Devuelve ErrInvalidCredentials (sin distinguir email/contraseña) o
// ErrAccountSuspended.
func (s *Service) Login(ctx context.Context, rawEmail, password string) (*users.User, string, error) {
	normEmail, err := normalizeEmail(rawEmail)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	u, err := s.users.GetByEmail(ctx, normEmail)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			_ = VerifyPassword(dummyPasswordHash, password) // mismo coste que una contraseña incorrecta
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !VerifyPassword(u.PasswordHash, password) {
		return nil, "", ErrInvalidCredentials
	}

	if u.Status == users.StatusSuspended {
		return nil, "", ErrAccountSuspended
	}

	token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return nil, "", err
	}

	s.touchAsync(u.ID)
	return u, token, nil
}

// Logout invalida una sesión. No es un error llamar a Logout con un
// token que ya no existe (es idempotente).
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	return s.sessions.Delete(ctx, sessionToken)
}

// Authenticate es lo que usa RequireAuth en CADA petición: devuelve solo el
// ID del usuario. Si el SessionStore ofrece el camino rápido
// (SessionResolver: Redis en un único viaje, sin tocar Postgres) lo usa; si
// no, cae en CurrentUser. Una suspensión o una eliminación de cuenta cortan
// el acceso de inmediato porque BlockUser deja una marca que Resolve lee.
func (s *Service) Authenticate(ctx context.Context, sessionToken string) (uuid.UUID, error) {
	if resolver, ok := s.sessions.(SessionResolver); ok {
		userID, touch, err := resolver.Resolve(ctx, sessionToken)
		if err != nil {
			return uuid.Nil, err
		}
		if touch {
			s.touchAsync(userID)
		}
		return userID, nil
	}

	u, err := s.CurrentUser(ctx, sessionToken)
	if err != nil {
		return uuid.Nil, err
	}
	return u.ID, nil
}

// CurrentUser resuelve la sesión y devuelve la cuenta asociada (lectura
// completa de Postgres). Devuelve ErrTokenInvalid si la sesión no existe o
// ha caducado, o ErrAccountSuspended si la cuenta fue suspendida después de
// abrir la sesión. Lo usan los caminos que necesitan el usuario entero, como
// admin.RequireAdmin (rol); RequireAuth usa Authenticate.
func (s *Service) CurrentUser(ctx context.Context, sessionToken string) (*users.User, error) {
	sess, err := s.sessions.Get(ctx, sessionToken)
	if err != nil {
		return nil, err
	}

	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil, ErrTokenInvalid
		}
		return nil, err
	}

	if u.Status == users.StatusSuspended {
		return nil, ErrAccountSuspended
	}

	return u, nil
}

// touchAsync refresca users.last_active_at sin retrasar la petición. Se
// llama como mucho una vez cada pocos minutos por usuario (ver
// resolveScript), así que es una escritura barata.
func (s *Service) touchAsync(userID uuid.UUID) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.users.TouchLastActive(ctx, userID); err != nil {
			slog.Warn("no se pudo actualizar last_active_at", "user_id", userID, "error", err)
		}
	}()
}

// RequestPasswordReset genera un token de reset y envía el email.
//
// Deliberadamente no revela si el email existe o no: ni en el resultado (siempre
// devuelve nil salvo un error real de infraestructura al buscar la cuenta) ni en
// el tiempo de respuesta. Por eso, una vez localizada la cuenta, crear el token y
// enviar el correo (SMTP, lento) se hace en segundo plano: ambos caminos cuestan
// una sola consulta. Un fallo posterior solo se registra; si el proceso se apaga
// justo en ese momento el correo puede perderse y el usuario lo pide de nuevo.
func (s *Service) RequestPasswordReset(ctx context.Context, rawEmail string) error {
	normEmail, err := normalizeEmail(rawEmail)
	if err != nil {
		return nil // formato inválido: no hay nada que enviar, pero no lo delatamos
	}

	u, err := s.users.GetByEmail(ctx, normEmail)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return nil
		}
		return err
	}

	// El contexto de la petición se cancela al responder: el envío necesita el suyo.
	go s.sendPasswordResetEmail(context.WithoutCancel(ctx), u)
	return nil
}

func (s *Service) sendPasswordResetEmail(ctx context.Context, u *users.User) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, err := s.tokens.Create(ctx, PurposePasswordReset, u.ID, s.resetTTL)
	if err != nil {
		slog.Error("no se pudo crear el token de reset de contraseña", "user_id", u.ID, "error", err)
		return
	}

	err = s.sender.Send(ctx, email.Message{
		To:      u.Email,
		Subject: "Restablece tu contraseña",
		Body: fmt.Sprintf(
			"Usa este código para restablecer tu contraseña: %s\nCaduca en %s.",
			token, s.resetTTL,
		),
	})
	if err != nil {
		slog.Error("no se pudo enviar el email de reset de contraseña", "user_id", u.ID, "error", err)
	}
}

// ResetPassword consume un token de reset y fija una contraseña nueva.
// Devuelve ErrTokenInvalid o ErrWeakPassword.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < MinPasswordLength {
		return ErrWeakPassword
	}

	userID, err := s.tokens.Consume(ctx, PurposePasswordReset, token)
	if err != nil {
		return err
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePasswordHash(ctx, userID, hash); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return ErrTokenInvalid
		}
		return err
	}

	return nil
}

// VerifyEmail consume un token de verificación y marca el email como
// verificado. Devuelve ErrTokenInvalid si el token no es válido.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	userID, err := s.tokens.Consume(ctx, PurposeEmailVerification, token)
	if err != nil {
		return err
	}

	if err := s.users.MarkEmailVerified(ctx, userID); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return ErrTokenInvalid
		}
		return err
	}

	return nil
}

// ResendVerification reenvía el email de verificación a un usuario ya
// autenticado. Devuelve ErrEmailAlreadyVerified si no hace falta.
func (s *Service) ResendVerification(ctx context.Context, userID uuid.UUID) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if u.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}

	s.sendVerificationEmailBestEffort(ctx, u)
	return nil
}

// DeleteAccount confirma la contraseña y elimina (soft delete) la cuenta.
// No cierra la sesión de la cookie actual: eso lo hace el handler HTTP, que
// es quien conoce esa cookie. Las sesiones en otros dispositivos se cortan
// con la marca de bloqueo (ver Authenticate).
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !VerifyPassword(u.PasswordHash, password) {
		return ErrInvalidCredentials
	}

	if err := s.users.SoftDelete(ctx, userID); err != nil {
		return err
	}

	if blocker, ok := s.sessions.(AccountBlocker); ok {
		if err := blocker.BlockUser(ctx, userID, BlockDeleted); err != nil {
			slog.Error("no se pudo marcar la cuenta eliminada como bloqueada", "user_id", userID, "error", err)
		}
	}
	return nil
}

func (s *Service) sendVerificationEmailBestEffort(ctx context.Context, u *users.User) {
	token, err := s.tokens.Create(ctx, PurposeEmailVerification, u.ID, s.verifyTTL)
	if err != nil {
		slog.Error("no se pudo generar el token de verificación de email", "error", err)
		return
	}

	err = s.sender.Send(ctx, email.Message{
		To:      u.Email,
		Subject: "Verifica tu email",
		Body: fmt.Sprintf(
			"Usa este código para verificar tu email: %s\nCaduca en %s.",
			token, s.verifyTTL,
		),
	})
	if err != nil {
		slog.Error("no se pudo enviar el email de verificación", "error", err)
	}
}

func normalizeEmail(raw string) (string, error) {
	e := strings.TrimSpace(strings.ToLower(raw))
	if !emailPattern.MatchString(e) {
		return "", ErrInvalidEmail
	}
	return e, nil
}
