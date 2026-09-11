package profiles

import (
	"time"

	"github.com/google/uuid"
)

// Photo es una foto de un perfil. StorageKey es opaco: lo resuelve la
// abstracción internal/storage (fichero local en V1, S3 en producción).
type Photo struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	StorageKey  string
	ContentType string
	Position    int
	CreatedAt   time.Time
}
