// Package storage abstrae dónde y cómo se guardan los ficheros binarios
// (por ahora, fotos de perfil). La lógica de negocio (internal/profiles)
// depende únicamente de la interfaz Storage, nunca de un driver
// concreto, para poder pasar de almacenamiento local (desarrollo) a un
// object storage S3-compatible (producción, Fase 14) sin tocarla.
package storage

import (
	"context"
	"io"
)

// Storage guarda, sirve y borra ficheros identificados por una clave
// opaca (key). El formato de la key lo decide quien la genera (en V1,
// el módulo profiles); Storage no le da ningún significado especial.
type Storage interface {
	// Save escribe el contenido de r bajo key. Si ya existe algo con esa
	// key, lo sobrescribe.
	Save(ctx context.Context, key string, r io.Reader) error

	// Open abre el contenido guardado bajo key para lectura. Quien lo usa
	// es responsable de cerrarlo. Devuelve ErrNotFound si no existe.
	Open(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete borra el contenido bajo key. No es un error borrar una key
	// que no existe (idempotente).
	Delete(ctx context.Context, key string) error
}
