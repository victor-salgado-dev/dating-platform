package storage

import (
	"context"
	"fmt"
)

// Config son los parámetros de configuración de storage, ya resueltos
// desde variables de entorno (ver internal/config). Se define aquí, no
// en el paquete config, para que storage siga siendo un paquete
// autocontenido que no dependa del resto de la app.
type Config struct {
	Driver    string
	LocalPath string
	S3        S3Config
}

// New construye el Storage configurado mediante STORAGE_DRIVER:
// "local" (por defecto, desarrollo) o "s3" (producción, Fase 14).
// Cambiar de uno a otro es solo configuración: ningún consumidor de
// Storage (p. ej. profiles) tiene que cambiar una sola línea.
func New(ctx context.Context, cfg Config) (Storage, error) {
	switch cfg.Driver {
	case "", "local":
		return NewLocalStorage(cfg.LocalPath)
	case "s3":
		return NewS3Storage(ctx, cfg.S3)
	default:
		return nil, fmt.Errorf("storage: driver desconocido %q", cfg.Driver)
	}
}

