package profiles

import (
	"errors"

	"dating-platform/backend/internal/apperr"
)

var (
	// ErrNotFound: no existe perfil para ese usuario/ID.
	ErrNotFound = errors.New("profiles: perfil no encontrado")

	// ErrAlreadyExists: el usuario ya tiene un perfil creado (relación 1:1).
	ErrAlreadyExists = errors.New("profiles: el usuario ya tiene un perfil")

	// ErrPhotoNotFound: no existe esa foto (o no pertenece al perfil indicado).
	ErrPhotoNotFound = errors.New("profiles: foto no encontrada")

	// ErrTooManyPhotos: se alcanzó el límite de fotos por perfil.
	ErrTooManyPhotos = errors.New("profiles: se alcanzó el número máximo de fotos")

	// ErrInterestNotFound: no existe ese interés en el catálogo. Se usa
	// para dar un mensaje claro en Service.SetInterest antes de escribir,
	// en vez de esperar a que la FK de profile_interests falle.
	ErrInterestNotFound = errors.New("profiles: interés no encontrado en el catálogo")
)

// ValidationError es un alias de apperr.ValidationError: la definición vive
// ahora en un único paquete y errors.As sigue funcionando desde fuera
// (profiles.ValidationError y apperr.ValidationError son el mismo tipo).
type ValidationError = apperr.ValidationError

func invalidField(field, message string) error {
	return apperr.InvalidField(field, message)
}
