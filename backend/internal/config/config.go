// Package config centraliza la lectura de configuración desde variables
// de entorno. Es el único lugar del backend que debe llamar a os.Getenv
// directamente: el resto de paquetes reciben la configuración ya resuelta.
package config

import (
	"fmt"
	"os"
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
// El driver "noop" (usado en Fase 1) no envía nada, solo registra en logs.
type EmailConfig struct {
	Driver string
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
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
