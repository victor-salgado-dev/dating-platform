package profiles

import (
	"context"

	"github.com/google/uuid"
)

// Repository persiste perfiles y todo lo que cuelga de ellos: fotos,
// hobbies, respuestas de personalidad y preferencias de pareja.
type Repository interface {
	// Create inserta un perfil nuevo. Rellena p.ID/CreatedAt/UpdatedAt.
	// Devuelve ErrAlreadyExists si el usuario ya tiene perfil.
	Create(ctx context.Context, p *Profile) error

	// GetByUserID devuelve el perfil de un usuario.
	// Devuelve ErrNotFound si todavía no se ha creado.
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)

	// GetPublicByID devuelve un perfil por su ID, solo si pertenece a una
	// cuenta activa (no suspendida, no eliminada) y no hay ningún bloqueo
	// (en cualquier sentido) entre viewerUserID y ese perfil. Es la
	// puerta de entrada a "perfiles públicos" (Fase 6) y a las reglas de
	// visibilidad del bloqueo (Fase 9): deliberadamente no distingue
	// entre "no existe", "cuenta inactiva" y "hay un bloqueo" — los tres
	// casos devuelven ErrNotFound, para no filtrar por qué un perfil no
	// aparece.
	GetPublicByID(ctx context.Context, id, viewerUserID uuid.UUID) (*Profile, error)

	// GetByIDAny devuelve un perfil por su ID sin aplicar ninguna regla
	// de visibilidad (ni estado de cuenta ni bloqueos). Solo debe usarse
	// para acciones que deben poder realizarse precisamente EN CONTRA de
	// esas reglas: bloquear a alguien, o reportarlo, tiene que funcionar
	// aunque esa persona ya te haya bloqueado a ti. Devuelve ErrNotFound
	// solo si el perfil no existe en absoluto.
	GetByIDAny(ctx context.Context, id uuid.UUID) (*Profile, error)

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

	// --- Catálogos (semi-estáticos, iguales para todos los perfiles) --

	// ListHobbyDefinitions devuelve el catálogo completo de hobbies,
	// ordenado por categoría y posición.
	ListHobbyDefinitions(ctx context.Context) ([]HobbyDefinition, error)

	// ListPersonalityStatements devuelve el catálogo completo de
	// afirmaciones de personalidad, ordenado por rasgo y posición.
	ListPersonalityStatements(ctx context.Context) ([]PersonalityStatement, error)

	// --- Hobbies del perfil --------------------------------------------

	// UpsertProfileHobby crea o actualiza la respuesta de un perfil a un
	// hobby del catálogo. Devuelve invalidField si hobbyKey no existe en
	// el catálogo o si liked/intensity no cumplen la regla de negocio
	// (intensity solo tiene sentido si liked = true).
	UpsertProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string, liked bool, intensity *int) (*ProfileHobby, error)

	// ListProfileHobbies devuelve todas las respuestas de hobbies de un
	// perfil. Los hobbies sin respuesta simplemente no aparecen.
	ListProfileHobbies(ctx context.Context, profileID uuid.UUID) ([]ProfileHobby, error)

	// DeleteProfileHobby borra la respuesta a un hobby (vuelve a "no
	// contestado"). Operación idempotente: no falla si no había fila.
	DeleteProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string) error

	// --- Personalidad del perfil ----------------------------------------

	// UpsertPersonalityAnswer crea o actualiza la puntuación (1-5) que
	// un perfil da a una afirmación del catálogo.
	UpsertPersonalityAnswer(ctx context.Context, profileID uuid.UUID, statementKey string, score int) (*ProfilePersonalityAnswer, error)

	// ListPersonalityAnswers devuelve todas las respuestas de
	// personalidad de un perfil.
	ListPersonalityAnswers(ctx context.Context, profileID uuid.UUID) ([]ProfilePersonalityAnswer, error)

	// GetPersonalityTraitScores devuelve el agregado ("Gesamt") por
	// rasgo, calculado a partir de las respuestas individuales. Un rasgo
	// sin ninguna respuesta contestada no aparece en el resultado.
	GetPersonalityTraitScores(ctx context.Context, profileID uuid.UUID) ([]PersonalityTraitScore, error)

	// --- Preferencias de pareja -------------------------------------------

	// GetPartnerPreferences devuelve las preferencias de pareja de un
	// perfil. Si el usuario todavía no las ha rellenado, no es un error:
	// devuelve un valor con todos los campos a nil/zero.
	GetPartnerPreferences(ctx context.Context, profileID uuid.UUID) (*PartnerPreferences, error)

	// UpsertPartnerPreferences aplica un patch parcial (crea la fila si
	// todavía no existía) y devuelve el resultado.
	UpsertPartnerPreferences(ctx context.Context, profileID uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error)
}
