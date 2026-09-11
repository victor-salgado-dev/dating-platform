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
// de imágenes. En V1 solo se usa el driver "local"; en producción se
// podrá cambiar a un driver S3-compatible sin tocar la lógica de negocio.
type StorageConfig struct {
	Driver    string
	LocalPath string
}

// EmailConfig es la configuración de la abstracción de envío de email.
// El driver "noop" (usado hasta ahora) no envía nada, solo registra en logs.
type EmailConfig struct {
	Driver string
	From   string
}

// AuthConfig agrupa la configuración del módulo de autenticación:
// duración de sesiones y de los tokens de un solo uso (verificación de
// email, reseteo de contraseña), y si las cookies deben marcarse Secure.
type AuthConfig struct {
	SessionTTL             time.Duration
	EmailVerificationTTL   time.Duration
	PasswordResetTTL       time.Duration
	CookieSecure           bool
	CookieName             string
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
		},

		Email: EmailConfig{
			Driver: getEnv("EMAIL_DRIVER", "noop"),
			From:   getEnv("EMAIL_FROM", "no-reply@example.com"),
		},

		Auth: AuthConfig{
			SessionTTL:           time.Duration(getEnvHours("SESSION_TTL_HOURS", 720)) * time.Hour,
			EmailVerificationTTL: time.Duration(getEnvHours("EMAIL_VERIFICATION_TTL_HOURS", 24)) * time.Hour,
			PasswordResetTTL:     time.Duration(getEnvHours("PASSWORD_RESET_TTL_HOURS", 1)) * time.Hour,
			CookieSecure:         getEnv("APP_ENV", "development") == "production",
			CookieName:           getEnv("SESSION_COOKIE_NAME", "session_id"),
		},
	}
}

func getEnvHours(key string, fallback int) int {
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

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
