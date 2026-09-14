package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]string{"hello": "world"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, se esperaba %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, se esperaba application/json", ct)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("la respuesta no es JSON válido: %v", err)
	}
	if body["hello"] != "world" {
		t.Errorf("cuerpo = %v, no contiene el valor esperado", body)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "invalid_field", "el campo X es inválido")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, se esperaba %d", rec.Code, http.StatusBadRequest)
	}

	var body ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("la respuesta no es JSON válido: %v", err)
	}
	if body.Error.Code != "invalid_field" {
		t.Errorf("error.code = %q, se esperaba invalid_field", body.Error.Code)
	}
	if body.Error.Message != "el campo X es inválido" {
		t.Errorf("error.message = %q, no coincide", body.Error.Message)
	}
}

func TestClientIPPrefersForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")

	if got := ClientIP(r); got != "203.0.113.9" {
		t.Errorf("ClientIP = %q, se esperaba el primer valor de X-Forwarded-For", got)
	}
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "198.51.100.7:54321"

	if got := ClientIP(r); got != "198.51.100.7" {
		t.Errorf("ClientIP = %q, se esperaba el host de RemoteAddr sin el puerto", got)
	}
}
