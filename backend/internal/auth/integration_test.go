//go:build integration

// Tests de integración: requieren PostgreSQL y Redis reales
// (`docker compose up -d postgres redis`). Se ejecutan con:
//
//	docker compose run --rm --entrypoint go backend test -tags=integration ./...
//
// o, más cómodo, `make test-integration`.
package auth_test

import (
	"context"
	"testing"
	"time"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/consent"
	"dating-platform/backend/internal/email"
	"dating-platform/backend/internal/testutil"
	"dating-platform/backend/internal/users"
)

// testDeps agrupa las piezas necesarias para probar auth.Service de
// extremo a extremo contra infraestructura real.
type testDeps struct {
	svc    *auth.Service
	tokens *auth.RedisTokenStore
}

func newTestDeps(t *testing.T) testDeps {
	t.Helper()

	pool := testutil.RequireDB(t)
	rdb := testutil.RequireRedis(t)

	usersRepo := users.NewPostgresRepository(pool)
	sessions := auth.NewRedisSessionStore(rdb, time.Hour)
	tokens := auth.NewRedisTokenStore(rdb)
	sender := email.NewNoopSender()
	consentService := consent.NewService(consent.NewPostgresRepository(pool))

	svc := auth.NewService(usersRepo, sessions, tokens, sender, consentService, time.Hour, time.Hour, time.Hour)
	return testDeps{svc: svc, tokens: tokens}
}

func TestIntegration_RegisterLoginLogout(t *testing.T) {
	deps := newTestDeps(t)
	ctx := context.Background()
	userEmail := testutil.UniqueEmail()

	u, sessionToken, err := deps.svc.Register(ctx, userEmail, "contraseña-larga-123", true)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if sessionToken == "" {
		t.Fatal("Register debería devolver un token de sesión")
	}
	if u.IsEmailVerified() {
		t.Error("una cuenta recién registrada no debería tener el email verificado")
	}

	// La sesión abierta al registrarse debe ser válida de inmediato.
	current, err := deps.svc.CurrentUser(ctx, sessionToken)
	if err != nil {
		t.Fatalf("CurrentUser tras registrar: %v", err)
	}
	if current.ID != u.ID {
		t.Error("CurrentUser devolvió una cuenta distinta a la registrada")
	}

	// Registrar el mismo email dos veces debe fallar.
	if _, _, err := deps.svc.Register(ctx, userEmail, "otra-contraseña-123", true); err == nil {
		t.Error("registrar un email ya usado debería fallar")
	}

	// Login con contraseña incorrecta debe fallar, sin distinguir el motivo.
	if _, _, err := deps.svc.Login(ctx, userEmail, "contraseña-incorrecta"); err == nil {
		t.Error("login con contraseña incorrecta debería fallar")
	}

	// Login correcto abre una sesión nueva e independiente.
	_, loginToken, err := deps.svc.Login(ctx, userEmail, "contraseña-larga-123")
	if err != nil {
		t.Fatalf("Login con credenciales correctas: %v", err)
	}

	// Logout invalida esa sesión concreta...
	if err := deps.svc.Logout(ctx, loginToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := deps.svc.CurrentUser(ctx, loginToken); err == nil {
		t.Error("una sesión cerrada no debería seguir siendo válida")
	}

	// ...pero no afecta a la sesión original abierta por Register.
	if _, err := deps.svc.CurrentUser(ctx, sessionToken); err != nil {
		t.Errorf("la sesión original no debería haberse visto afectada por el logout de otra: %v", err)
	}
}

func TestIntegration_RegisterRequiresAcceptedTerms(t *testing.T) {
	deps := newTestDeps(t)
	ctx := context.Background()

	if _, _, err := deps.svc.Register(ctx, testutil.UniqueEmail(), "contraseña-123456", false); err == nil {
		t.Error("registrar sin aceptar los Términos debería fallar (auth.ErrTermsNotAccepted)")
	}
}

func TestIntegration_RegisterRecordsConsents(t *testing.T) {
	pool := testutil.RequireDB(t)
	deps := newTestDeps(t)
	ctx := context.Background()

	u, _, err := deps.svc.Register(ctx, testutil.UniqueEmail(), "contraseña-123456", true)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	consentsRepo := consent.NewPostgresRepository(pool)
	items, err := consentsRepo.ListByUser(ctx, u.ID)
	if err != nil {
		t.Fatalf("listar consentimientos: %v", err)
	}

	gotTypes := map[consent.DocumentType]bool{}
	for _, c := range items {
		gotTypes[c.DocumentType] = true
	}
	if !gotTypes[consent.DocumentTerms] {
		t.Error("el registro debería haber dejado un consentimiento de tipo 'terms'")
	}
	if !gotTypes[consent.DocumentPrivacyPolicy] {
		t.Error("el registro debería haber dejado un consentimiento de tipo 'privacy_policy' (separado de 'terms')")
	}
}

func TestIntegration_PasswordResetConsumesTokenOnce(t *testing.T) {
	deps := newTestDeps(t)
	ctx := context.Background()
	userEmail := testutil.UniqueEmail()

	u, _, err := deps.svc.Register(ctx, userEmail, "contraseña-original-123", true)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Un email inexistente no debe dar error (evita enumeración de cuentas).
	if err := deps.svc.RequestPasswordReset(ctx, testutil.UniqueEmail()); err != nil {
		t.Errorf("RequestPasswordReset con email inexistente no debería fallar: %v", err)
	}

	// Generamos el token directamente con el mismo TokenStore que usa el
	// servicio (simulando "el email que se habría enviado"): así se
	// prueba el consumo del token sin depender de interceptar un email.
	token, err := deps.tokens.Create(ctx, auth.PurposePasswordReset, u.ID, time.Hour)
	if err != nil {
		t.Fatalf("crear token de reset de prueba: %v", err)
	}

	if err := deps.svc.ResetPassword(ctx, token, "contraseña-nueva-456"); err != nil {
		t.Fatalf("ResetPassword con token válido: %v", err)
	}

	// El token es de un solo uso: reutilizarlo debe fallar.
	if err := deps.svc.ResetPassword(ctx, token, "otra-contraseña-789"); err == nil {
		t.Error("reutilizar un token de reset ya consumido debería fallar")
	}

	// La contraseña antigua ya no sirve; la nueva sí.
	if _, _, err := deps.svc.Login(ctx, userEmail, "contraseña-original-123"); err == nil {
		t.Error("la contraseña antigua no debería seguir funcionando tras el reset")
	}
	if _, _, err := deps.svc.Login(ctx, userEmail, "contraseña-nueva-456"); err != nil {
		t.Errorf("la contraseña nueva debería funcionar tras el reset: %v", err)
	}
}

func TestIntegration_EmailVerification(t *testing.T) {
	deps := newTestDeps(t)
	ctx := context.Background()
	userEmail := testutil.UniqueEmail()

	u, _, err := deps.svc.Register(ctx, userEmail, "contraseña-123456", true)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	token, err := deps.tokens.Create(ctx, auth.PurposeEmailVerification, u.ID, time.Hour)
	if err != nil {
		t.Fatalf("crear token de verificación de prueba: %v", err)
	}

	if err := deps.svc.VerifyEmail(ctx, token); err != nil {
		t.Fatalf("VerifyEmail con token válido: %v", err)
	}

	if err := deps.svc.ResendVerification(ctx, u.ID); err == nil {
		t.Error("reenviar verificación a un email ya verificado debería devolver ErrEmailAlreadyVerified")
	}
}
