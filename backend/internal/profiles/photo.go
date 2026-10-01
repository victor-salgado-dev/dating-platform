package profiles

import (
	"time"

	"github.com/google/uuid"
)

// Photo es una foto de un perfil. StorageKey es opaco: lo resuelve la
// abstracción internal/storage (fichero local en V1, S3 en producción).
type Photo struct {
	ID         uuid.UUID
	ProfileID  uuid.UUID
	StorageKey string
	// ThumbStorageKey es la miniatura (vacía en fotos anteriores a la
	// migración 000021: en ese caso se sirve la original).
	ThumbStorageKey string
	ContentType     string
	Position        int
	CreatedAt       time.Time
}
