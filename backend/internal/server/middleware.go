package server

import (
	"log/slog"
	"net/http"
	"time"

	"dating-platform/backend/internal/httpx"
)

// withLogging registra método, ruta, código de estado, duración e IP
// de origen de cada petición. No registra cuerpos de petición/respuesta
// para evitar filtrar datos sensibles en los logs.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", httpx.ClientIP(r),
		)
	})
}

// withRecover evita que un panic en un handler tumbe todo el proceso.
// Nunca expone detalles internos (stack traces) al cliente.
func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recuperado", "error", err, "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal_server_error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withSecurityHeaders añade cabeceras HTTP de seguridad estándar a
// toda respuesta. Son cabeceras "gratis": no cambian el comportamiento
// de la API para un cliente legítimo, solo reducen la superficie de
// ataque en el navegador (MIME-sniffing, clickjacking, fuga de
// referrer). No sustituyen a CSRF/CSP específicos si se necesitaran
// más adelante, pero son la base razonable para V1.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// withMaxBodySize limita el tamaño de CUALQUIER cuerpo de petición.
// Es un backstop de memoria/DoS, no una validación semántica: cada
// endpoint sigue aplicando sus propios límites más estrictos (2000
// caracteres en un mensaje, 5 MB en una foto...). El límite aquí debe
// ser mayor que el límite más grande de cualquier endpoint (fotos),
// para no romperlo.
func withMaxBodySize(next http.Handler, maxBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
