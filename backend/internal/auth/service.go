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

// Service implementa los casos de uso de autenticación. Depende de
// interfaces (users.Repository, SessionStore, TokenStore, email.Sender),
// nunca de implementaciones concretas, para poder sustituirlas en tests.
type Service struct {
	users    users.Repository
	sessions SessionStore
	tokens   TokenStore
	sender   email.Sender

	sessionTTL   time.Duration
	verifyTTL    time.Duration
	resetTTL     time.Duration
}

func NewService(
	usersRepo users.Repository,
	sessions SessionStore,
	tokens TokenStore,
	sender email.Sender,
	sessionTTL, verifyTTL, resetTTL time.Duration,
) *Service {
	return &Service{
		users:      usersRepo,
		sessions:   sessions,
		tokens:     tokens,
		sender:     sender,
		sessionTTL: sessionTTL,
		verifyTTL:  verifyTTL,
		resetTTL:   resetTTL,
	}
}

// Register crea una cuenta nueva, dispara el email de verificación
// (best-effort: un fallo de envío no impide el registro) y abre sesión
// inmediatamente. Devuelve ErrWeakPassword, users.ErrDuplicateEmail, o
// un error de validación de email envuelto.
func (s *Service) Register(ctx context.Context, rawEmail, password string) (*users.User, string, error) {
	normEmail, err := normalizeEmail(rawEmail)
	if err != nil {
		return nil, "", err
	}
	if len(password) < MinPasswordLength {
		return nil, "", ErrWeakPassword
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	u := &users.User{Email: normEmail, PasswordHash: hash}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, "", err // p.ej. users.ErrDuplicateEmail
	}

	s.sendVerificationEmailBestEffort(ctx, u)

	token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return nil, "", err
	}

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

	return u, token, nil
}

// Logout invalida una sesión. No es un error llamar a Logout con un
// token que ya no existe (es idempotente).
func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	return s.sessions.Delete(ctx, sessionToken)
}

// CurrentUser resuelve la sesión y devuelve la cuenta asociada.
// Devuelve ErrTokenInvalid si la sesión no existe o ha caducado.
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

	return u, nil
}

// RequestPasswordReset genera un token de reset y envía el email.
// Deliberadamente no revela si el email existe o no: siempre devuelve
// nil salvo un error real de infraestructura, para no facilitar
// enumeración de cuentas registradas.
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

	token, err := s.tokens.Create(ctx, PurposePasswordReset, u.ID, s.resetTTL)
	if err != nil {
		return err
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
		slog.Error("no se pudo enviar el email de reset de contraseña", "error", err)
	}

	return nil
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
// No cierra la sesión: eso lo hace el handler HTTP, que es quien conoce
// la cookie concreta a invalidar.
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !VerifyPassword(u.PasswordHash, password) {
		return ErrInvalidCredentials
	}

	return s.users.SoftDelete(ctx, userID)
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
