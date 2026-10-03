package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"dating-platform/backend/internal/email"
	"dating-platform/backend/internal/users"
)

// --- Dobles de prueba --------------------------------------------------------
//
// fakeUsers embebe la interfaz: solo se implementa lo que usan estos tests y
// cualquier otro método entra en pánico, lo que delata un uso inesperado.

type fakeUsers struct {
	users.Repository
	byEmail map[string]*users.User
	err     error // si no es nil, GetByEmail falla con él (error de infraestructura)
}

func (f *fakeUsers) GetByEmail(_ context.Context, e string) (*users.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.byEmail[e]; ok {
		return u, nil
	}
	return nil, users.ErrNotFound
}

func (f *fakeUsers) TouchLastActive(context.Context, uuid.UUID) error { return nil }

type fakeSessions struct{}

func (fakeSessions) Create(context.Context, uuid.UUID) (string, error) { return "session-token", nil }
func (fakeSessions) Get(context.Context, string) (*Session, error)     { return nil, ErrTokenInvalid }
func (fakeSessions) Delete(context.Context, string) error             { return nil }

type fakeTokens struct{}

func (fakeTokens) Create(context.Context, TokenPurpose, uuid.UUID, time.Duration) (string, error) {
	return "reset-token-123", nil
}
func (fakeTokens) Consume(context.Context, TokenPurpose, string) (uuid.UUID, error) {
	return uuid.Nil, ErrTokenInvalid
}

// fakeSender simula un SMTP lento: si release no es nil, Send espera a que se cierre.
type fakeSender struct {
	release chan struct{}
	sent    chan email.Message
}

func newFakeSender(slow bool) *fakeSender {
	s := &fakeSender{sent: make(chan email.Message, 4)}
	if slow {
		s.release = make(chan struct{})
	}
	return s
}

func (f *fakeSender) Send(_ context.Context, m email.Message) error {
	if f.release != nil {
		<-f.release
	}
	f.sent <- m
	return nil
}

// unblock libera un Send pendiente (idempotente).
func (f *fakeSender) unblock() {
	if f.release == nil {
		return
	}
	select {
	case <-f.release:
	default:
		close(f.release)
	}
}

func newTestService(t *testing.T, repo *fakeUsers, sender *fakeSender) *Service {
	t.Helper()
	t.Cleanup(sender.unblock)
	return NewService(repo, fakeSessions{}, fakeTokens{}, sender, nil, time.Hour, time.Hour, time.Hour)
}

func activeUser(t *testing.T, password string) *users.User {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	return &users.User{ID: uuid.New(), Email: "ana@example.com", PasswordHash: hash, Status: users.StatusActive}
}

// --- Login: el tiempo no debe revelar si el email existe ----------------------

func TestDummyPasswordHash_HasTheSameCostAsRealHashes(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(dummyPasswordHash))
	if err != nil {
		t.Fatalf("dummyPasswordHash no es un hash bcrypt válido: %v", err)
	}
	if cost != bcrypt.DefaultCost {
		t.Errorf("coste = %d, debe ser bcrypt.DefaultCost (%d): HashPassword usa ese coste y la comparación ficticia debe tardar lo mismo", cost, bcrypt.DefaultCost)
	}
	if VerifyPassword(dummyPasswordHash, "cualquier-contraseña") {
		t.Error("el hash ficticio no debe validar ninguna contraseña razonable")
	}
}

func TestLogin_UnknownEmailTakesAsLongAsWrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("mide tiempos de bcrypt")
	}
	user := activeUser(t, "correcta-123456")
	svc := newTestService(t, &fakeUsers{byEmail: map[string]*users.User{user.Email: user}}, newFakeSender(false))

	best := func(emailAddr string) (d time.Duration, err error) {
		d = time.Hour
		for i := 0; i < 3; i++ { // el mínimo de 3 intentos filtra ruido de la máquina
			start := time.Now()
			_, _, err = svc.Login(context.Background(), emailAddr, "incorrecta-123456")
			if el := time.Since(start); el < d {
				d = el
			}
		}
		return d, err
	}

	wrongPassword, err1 := best(user.Email)
	unknownEmail, err2 := best("nadie@example.com")

	if !errors.Is(err1, ErrInvalidCredentials) || !errors.Is(err2, ErrInvalidCredentials) {
		t.Fatalf("ambos casos deben devolver ErrInvalidCredentials: %v / %v", err1, err2)
	}
	// bcrypt tarda decenas de ms; sin comparación ficticia el email desconocido
	// respondería en microsegundos. Se exige al menos un tercio, con mucho margen.
	if unknownEmail < wrongPassword/3 {
		t.Errorf("email desconocido: %v, contraseña incorrecta: %v. La diferencia delata qué emails están registrados", unknownEmail, wrongPassword)
	}
}

func TestLogin_NormalBehaviourIsUnchanged(t *testing.T) {
	user := activeUser(t, "correcta-123456")
	suspended := activeUser(t, "correcta-123456")
	suspended.Email, suspended.Status = "baneada@example.com", users.StatusSuspended
	svc := newTestService(t, &fakeUsers{byEmail: map[string]*users.User{user.Email: user, suspended.Email: suspended}}, newFakeSender(false))
	ctx := context.Background()

	if _, token, err := svc.Login(ctx, "  ANA@example.com ", "correcta-123456"); err != nil || token != "session-token" {
		t.Errorf("login correcto: token=%q err=%v", token, err)
	}
	if _, _, err := svc.Login(ctx, user.Email, "mal"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("contraseña incorrecta: %v", err)
	}
	if _, _, err := svc.Login(ctx, "no-es-un-email", "x"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("email con formato inválido: %v", err)
	}
	if _, _, err := svc.Login(ctx, suspended.Email, "correcta-123456"); !errors.Is(err, ErrAccountSuspended) {
		t.Errorf("cuenta suspendida con contraseña correcta: %v", err)
	}
	if _, _, err := svc.Login(ctx, suspended.Email, "incorrecta"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("cuenta suspendida con contraseña incorrecta no debe revelar la suspensión: %v", err)
	}
}

// --- Reset de contraseña: el envío del correo no debe retrasar la respuesta ------

func TestRequestPasswordReset_DoesNotWaitForEmailDelivery(t *testing.T) {
	user := activeUser(t, "x-123456789")
	sender := newFakeSender(true) // SMTP lento
	svc := newTestService(t, &fakeUsers{byEmail: map[string]*users.User{user.Email: user}}, sender)

	returned := make(chan error, 1)
	go func() { returned <- svc.RequestPasswordReset(context.Background(), user.Email) }()

	select {
	case err := <-returned:
		if err != nil {
			t.Fatalf("RequestPasswordReset: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("RequestPasswordReset esperó a que se enviara el correo: con una cuenta existente tarda mucho más que con una inexistente y delata que el email está registrado")
	}

	// Aun así el correo acaba enviándose, con el token.
	sender.unblock()
	select {
	case m := <-sender.sent:
		if m.To != user.Email || !strings.Contains(m.Body, "reset-token-123") {
			t.Errorf("correo inesperado: %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("el correo de restablecimiento nunca se envió")
	}
}

func TestRequestPasswordReset_UnknownOrInvalidEmailSendsNothing(t *testing.T) {
	sender := newFakeSender(false)
	svc := newTestService(t, &fakeUsers{byEmail: map[string]*users.User{}}, sender)

	for _, addr := range []string{"nadie@example.com", "no-es-un-email", ""} {
		if err := svc.RequestPasswordReset(context.Background(), addr); err != nil {
			t.Errorf("RequestPasswordReset(%q) no debe fallar: %v", addr, err)
		}
	}
	select {
	case m := <-sender.sent:
		t.Fatalf("no debía enviarse ningún correo, se envió %+v", m)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestRequestPasswordReset_StillReportsInfrastructureErrors(t *testing.T) {
	boom := errors.New("postgres caído")
	svc := newTestService(t, &fakeUsers{err: boom}, newFakeSender(false))

	if err := svc.RequestPasswordReset(context.Background(), "ana@example.com"); !errors.Is(err, boom) {
		t.Errorf("un error real de infraestructura debe propagarse, se obtuvo %v", err)
	}
}
