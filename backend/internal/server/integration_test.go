//go:build integration

package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/config"
	"dating-platform/backend/internal/consent"
	"dating-platform/backend/internal/email"
	"dating-platform/backend/internal/profiles"
	"dating-platform/backend/internal/ratelimit"
	"dating-platform/backend/internal/search"
	"dating-platform/backend/internal/server"
	"dating-platform/backend/internal/storage"
	"dating-platform/backend/internal/testutil"
	"dating-platform/backend/internal/users"
)

// newTestServer levanta el router HTTP real (el mismo que arranca
// cmd/api) con dependencias reales, para un test de "integración
// básica" de extremo a extremo. Deja sin usar (nil) los módulos que
// esta prueba no ejercita (messaging, blocking, reports, admin): sus
// rutas seguirían registradas, pero no se llaman aquí.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	pool := testutil.RequireDB(t)
	rdb := testutil.RequireRedis(t)
	cfg := config.Load()

	usersRepo := users.NewPostgresRepository(pool)
	sessions := auth.NewRedisSessionStore(rdb, time.Hour)
	tokens := auth.NewRedisTokenStore(rdb)
	consentService := consent.NewService(consent.NewPostgresRepository(pool))
	authService := auth.NewService(usersRepo, sessions, tokens, email.NewNoopSender(), consentService, time.Hour, time.Hour, time.Hour)
	authHandler := auth.NewHandler(authService, cfg.Auth)

	fileStorage, err := storage.NewLocalStorage(t.TempDir())
	if err != nil {
		t.Fatalf("crear storage local de prueba: %v", err)
	}
	profilesRepo := profiles.NewPostgresRepository(pool)
	profilesService := profiles.NewService(profilesRepo, fileStorage)
	profilesHandler := profiles.NewHandler(profilesService)

	searchRepo := search.NewPostgresRepository(pool)
	searchService := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchService)

	router := server.NewRouter(server.Dependencies{
		DB:              pool,
		Redis:           rdb,
		AuthService:     authService,
		AuthHandler:     authHandler,
		SessionCookie:   cfg.Auth.CookieName,
		ProfilesHandler: profilesHandler,
		SearchHandler:   searchHandler,
		RateLimiter:     ratelimit.New(rdb),
		Security:        cfg.Security,
	})

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)
	return ts
}

// newClientWithCookies crea un http.Client con su propio cookiejar,
// simulando la sesión de UN usuario (igual que un navegador).
func newClientWithCookies(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("crear cookiejar: %v", err)
	}
	return &http.Client{Jar: jar}
}

func postJSON(t *testing.T, client *http.Client, url string, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("serializar cuerpo de petición: %v", err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

// TestIntegration_RegisterProfileAndFindEachOtherInSearch es el "flujo
// feliz" básico de la plataforma: dos personas se registran, crean su
// perfil, y cada una encuentra a la otra en la búsqueda (pero no se
// encuentra a sí misma).
func TestIntegration_RegisterProfileAndFindEachOtherInSearch(t *testing.T) {
	ts := newTestServer(t)

	clientA := newClientWithCookies(t)
	clientB := newClientWithCookies(t)

	// --- Registro ---------------------------------------------------
	respA := postJSON(t, clientA, ts.URL+"/api/v1/auth/register", map[string]any{
		"email": testutil.UniqueEmail(), "password": "contraseña-a-123", "accepted_terms": true,
	})
	if respA.StatusCode != http.StatusCreated {
		t.Fatalf("registro de A: status = %d, se esperaba 201", respA.StatusCode)
	}
	respA.Body.Close()

	respB := postJSON(t, clientB, ts.URL+"/api/v1/auth/register", map[string]any{
		"email": testutil.UniqueEmail(), "password": "contraseña-b-123", "accepted_terms": true,
	})
	if respB.StatusCode != http.StatusCreated {
		t.Fatalf("registro de B: status = %d, se esperaba 201", respB.StatusCode)
	}
	respB.Body.Close()

	// --- Crear perfil -------------------------------------------------
	profileInput := map[string]any{
		"display_name": "Persona de prueba",
		"birth_date":   "1995-05-20",
		"gender":       "other",
		"country_code": "ES",
	}

	respProfileA := postJSON(t, clientA, ts.URL+"/api/v1/profiles/me", profileInput)
	if respProfileA.StatusCode != http.StatusCreated {
		t.Fatalf("crear perfil de A: status = %d, se esperaba 201", respProfileA.StatusCode)
	}
	var profileA struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(respProfileA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decodificar perfil de A: %v", err)
	}
	respProfileA.Body.Close()

	respProfileB := postJSON(t, clientB, ts.URL+"/api/v1/profiles/me", profileInput)
	if respProfileB.StatusCode != http.StatusCreated {
		t.Fatalf("crear perfil de B: status = %d, se esperaba 201", respProfileB.StatusCode)
	}
	var profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(respProfileB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decodificar perfil de B: %v", err)
	}
	respProfileB.Body.Close()

	// --- Cada quien encuentra al otro, no a sí mismo -------------------
	searchResp, err := clientA.Get(ts.URL + "/api/v1/search/profiles?page_size=50")
	if err != nil {
		t.Fatalf("buscar como A: %v", err)
	}
	defer searchResp.Body.Close()
	if searchResp.StatusCode != http.StatusOK {
		t.Fatalf("buscar como A: status = %d, se esperaba 200", searchResp.StatusCode)
	}

	var results struct {
		Items []struct {
			ProfileID string `json:"profile_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(searchResp.Body).Decode(&results); err != nil {
		t.Fatalf("decodificar resultados de búsqueda: %v", err)
	}

	foundB := false
	for _, it := range results.Items {
		if it.ProfileID == profileB.ID {
			foundB = true
		}
		if it.ProfileID == profileA.ID {
			t.Error("A no debería encontrarse a sí mismo en sus propios resultados de búsqueda")
		}
	}
	if !foundB {
		t.Error("A debería encontrar el perfil de B en la búsqueda")
	}
}

// TestIntegration_HealthEndpoint comprueba la "comprobación Docker" más
// básica a nivel de aplicación: con Postgres y Redis arriba, /healthz y
// /api/v1/health deben responder 200.
func TestIntegration_HealthEndpoint(t *testing.T) {
	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/healthz status = %d, se esperaba 200", resp.StatusCode)
	}

	resp2, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /api/v1/health: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("/api/v1/health status = %d, se esperaba 200 (Postgres y Redis deberían estar arriba)", resp2.StatusCode)
	}
}
