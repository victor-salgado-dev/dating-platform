package profiles

import (
	"errors"
	"fmt"
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

// ValidationError señala que un campo del perfil no cumple las reglas
// de negocio (formato, rango, valor permitido). El handler HTTP la
// traduce a un 400 con el nombre de campo incluido.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("profiles: campo %q inválido: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}
