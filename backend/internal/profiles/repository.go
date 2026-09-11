package profiles

import (
	"context"

	"github.com/google/uuid"
)

// Repository persiste perfiles y sus fotos.
type Repository interface {
	// Create inserta un perfil nuevo. Rellena p.ID/CreatedAt/UpdatedAt.
	// Devuelve ErrAlreadyExists si el usuario ya tiene perfil.
	Create(ctx context.Context, p *Profile) error

	// GetByUserID devuelve el perfil de un usuario.
	// Devuelve ErrNotFound si todavía no se ha creado.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)

	// Update aplica un patch parcial al perfil de userID y devuelve el
	// perfil resultante. Devuelve ErrNotFound si no existe.
	Update(ctx context.Context, userID uuid.UUID, patch ProfilePatch) (*Profile, error)

	// --- Fotos --------------------------------------------------------

	// AddPhoto inserta una foto para profileID en la siguiente posición
	// disponible. Rellena photo.ID/Position/CreatedAt.
	AddPhoto(ctx context.Context, profileID uuid.UUID, photo *Photo) error

	// ListPhotos devuelve las fotos de un perfil ordenadas por posición.
	ListPhotos(ctx context.Context, profileID uuid.UUID) ([]Photo, error)

	// CountPhotos devuelve cuántas fotos tiene ya un perfil (para aplicar
	// el límite máximo antes de aceptar una subida).
	CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error)

	// GetPhoto devuelve una foto por ID, solo si pertenece a profileID.
	// Devuelve ErrPhotoNotFound en caso contrario.
	GetPhoto(ctx context.Context, profileID, photoID uuid.UUID) (*Photo, error)

	// DeletePhoto borra el registro de una foto (no el fichero: eso lo
	// hace el Service llamando a storage.Storage por separado).
	// Devuelve ErrPhotoNotFound si no existe o no pertenece al perfil.
	DeletePhoto(ctx context.Context, profileID, photoID uuid.UUID) error
}
