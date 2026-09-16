// Package config centraliza la lectura de configuración desde variables
// de entorno. Es el único lugar del backend que debe llamar a os.Getenv
// directamente: el resto de paquetes reciben la configuración ya resuelta.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config contiene toda la configuración de la aplicación, agrupada por área.
type Config struct {
	AppEnv   string
	LogLevel string

	Backend  BackendConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Storage  StorageConfig
	Email    EmailConfig
	Auth     AuthConfig
	Security SecurityConfig
}

type BackendConfig struct {
	Host string
	Port string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN construye la cadena de conexión de PostgreSQL para pgx.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// StorageConfig es la configuración de la abstracción de almacenamiento
// de imágenes. driver=local (V1/desarrollo) o driver=s3 (producción,
// Fase 14: cualquier proveedor compatible con S3 — AWS S3, MinIO,
// DigitalOcean Spaces, Backblaze B2...). Los campos S3* se ignoran
// completamente si driver=local.
type StorageConfig struct {
	Driver    string
	LocalPath string

	S3Endpoint       string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3UseSSL         bool
	S3ForcePathStyle bool
}

// EmailConfig es la configuración de la abstracción de envío de email.
// El driver "noop" (desarrollo) registra en logs en vez de enviar; el
// driver "smtp" (producción, Fase 14) envía de verdad.
type EmailConfig struct {
	Driver string
	From   string
	// ContactInbox es la dirección a la que llegan los mensajes del
	// formulario de contacto (Fase 13).
	ContactInbox string

	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

// AuthConfig agrupa la configuración del módulo de autenticación:
// duración de sesiones y de los tokens de un solo uso (verificación de
// email, reseteo de contraseña), y si las cookies deben marcarse Secure.
type AuthConfig struct {
	SessionTTL           time.Duration
	EmailVerificationTTL time.Duration
	PasswordResetTTL     time.Duration
	CookieSecure         bool
	CookieName           string
}

// SecurityConfig agrupa los parámetros de hardening de la Fase 11:
// rate limiting (Redis) y el límite general de tamaño de petición.
type SecurityConfig struct {
	// Límite global, generoso, aplicado a TODAS las peticiones (defensa
	// básica frente a abuso/DoS por fuerza bruta general).
	GlobalRateLimit  int
	GlobalRateWindow time.Duration

	// Límite estricto aplicado solo a los endpoints sensibles de auth
	// (register, login, password/forgot, password/reset, email/verify,
	// email/resend), por IP.
	AuthRateLimit  int
	AuthRateWindow time.Duration

	// Tamaño máximo de cualquier cuerpo de petición, en bytes. Es un
	// backstop de memoria/DoS: los límites semánticos más estrictos
	// (2000 caracteres en un mensaje, 5 MB en una foto...) los sigue
	// aplicando cada módulo.
	MaxRequestBodyBytes int64
}

// Load lee la configuración desde variables de entorno, aplicando valores
// por defecto razonables para desarrollo. No debe inventar secretos: si
// falta una variable sensible en producción, debe fallar de forma explícita
// en una fase posterior (auth) cuando esa variable sea realmente necesaria.
func Load() Config {
	return Config{
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		Backend: BackendConfig{
			Host: getEnv("BACKEND_HOST", "0.0.0.0"),
			Port: getEnv("BACKEND_PORT", "8080"),
		},

		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "postgres"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "dating_app"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "dating_app"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "redis"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},

		Storage: StorageConfig{
			Driver:    getEnv("STORAGE_DRIVER", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "/data/uploads"),

			S3Endpoint:       getEnv("STORAGE_S3_ENDPOINT", ""),
			S3Region:         getEnv("STORAGE_S3_REGION", "us-east-1"),
			S3Bucket:         getEnv("STORAGE_S3_BUCKET", ""),
			S3AccessKey:      getEnv("STORAGE_S3_ACCESS_KEY", ""),
			S3SecretKey:      getEnv("STORAGE_S3_SECRET_KEY", ""),
			S3UseSSL:         getEnvBool("STORAGE_S3_USE_SSL", true),
			S3ForcePathStyle: getEnvBool("STORAGE_S3_FORCE_PATH_STYLE", false),
		},

		Email: EmailConfig{
			Driver:       getEnv("EMAIL_DRIVER", "noop"),
			From:         getEnv("EMAIL_FROM", "no-reply@example.com"),
			ContactInbox: getEnv("CONTACT_INBOX_EMAIL", "contact@example.com"),

			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		},

		Auth: AuthConfig{
			SessionTTL:           time.Duration(getEnvHours("SESSION_TTL_HOURS", 720)) * time.Hour,
			EmailVerificationTTL: time.Duration(getEnvHours("EMAIL_VERIFICATION_TTL_HOURS", 24)) * time.Hour,
			PasswordResetTTL:     time.Duration(getEnvHours("PASSWORD_RESET_TTL_HOURS", 1)) * time.Hour,
			CookieSecure:         getEnv("APP_ENV", "development") == "production",
			CookieName:           getEnv("SESSION_COOKIE_NAME", "session_id"),
		},

		Security: SecurityConfig{
			GlobalRateLimit:     getEnvInt("RATE_LIMIT_GLOBAL_MAX", 300),
			GlobalRateWindow:    time.Duration(getEnvInt("RATE_LIMIT_GLOBAL_WINDOW_SECONDS", 300)) * time.Second,
			AuthRateLimit:       getEnvInt("RATE_LIMIT_AUTH_MAX", 20),
			AuthRateWindow:      time.Duration(getEnvInt("RATE_LIMIT_AUTH_WINDOW_SECONDS", 900)) * time.Second,
			MaxRequestBodyBytes: int64(getEnvInt("MAX_REQUEST_BODY_MB", 10)) * 1024 * 1024,
		},
	}
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvHours(key string, fallback int) int {
	return getEnvInt(key, fallback)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
