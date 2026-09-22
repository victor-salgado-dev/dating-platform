package profiles

import (
	"context"

	"github.com/google/uuid"
)

// Repository persiste perfiles y todo lo que cuelga de ellos: fotos,
// idiomas, intereses, respuestas de personalidad y preferencias de
// pareja.
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

	AddPhoto(ctx context.Context, profileID uuid.UUID, photo *Photo) error
	ListPhotos(ctx context.Context, profileID uuid.UUID) ([]Photo, error)
	CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error)
	GetPhoto(ctx context.Context, profileID, photoID uuid.UUID) (*Photo, error)
	DeletePhoto(ctx context.Context, profileID, photoID uuid.UUID) error

	// --- Idiomas del perfil ---------------------------------------------
	//
	// Sin catálogo propio (a diferencia de intereses/hobbies): la lista
	// de códigos válidos es cerrada y estable, se valida con un CHECK
	// directamente en profile_languages.

	// UpsertProfileLanguage crea o actualiza el nivel de un idioma para
	// un perfil. Devuelve invalidField si languageCode no está en la
	// lista permitida o si level está fuera de 1-5.
	UpsertProfileLanguage(ctx context.Context, profileID uuid.UUID, languageCode string, level *int) (*ProfileLanguage, error)

	// ListProfileLanguages devuelve todos los idiomas indicados por un
	// perfil.
	ListProfileLanguages(ctx context.Context, profileID uuid.UUID) ([]ProfileLanguage, error)

	// DeleteProfileLanguage borra un idioma (vuelve a "no indicado").
	// Operación idempotente: no falla si no había fila.
	DeleteProfileLanguage(ctx context.Context, profileID uuid.UUID, languageCode string) error

	// --- Catálogo de intereses (semi-estático) --------------------------

	// ListInterestDefinitions devuelve el catálogo completo de
	// intereses, ordenado por categoría y posición.
	ListInterestDefinitions(ctx context.Context) ([]InterestDefinition, error)

	// GetInterestDefinition devuelve un único ítem del catálogo.
	// Devuelve ErrInterestNotFound si no existe.
	GetInterestDefinition(ctx context.Context, key string) (*InterestDefinition, error)

	// --- Intereses del perfil -------------------------------------------

	// UpsertProfileInterest crea o actualiza la respuesta de un perfil a
	// un interés del catálogo. Devuelve invalidField si interestKey no
	// existe en el catálogo o si level está fuera de 1-5. NO valida la
	// correspondencia has_level<->level (cruza dos tablas): de eso se
	// encarga el Service antes de llamar aquí.
	UpsertProfileInterest(ctx context.Context, profileID uuid.UUID, interestKey string, level *int) (*ProfileInterest, error)

	// ListProfileInterests devuelve todas las respuestas de intereses de
	// un perfil. Los intereses sin respuesta simplemente no aparecen.
	ListProfileInterests(ctx context.Context, profileID uuid.UUID) ([]ProfileInterest, error)

	// DeleteProfileInterest borra la respuesta a un interés (vuelve a
	// "no seleccionado"). Operación idempotente.
	DeleteProfileInterest(ctx context.Context, profileID uuid.UUID, interestKey string) error

	// --- Personalidad ----------------------------------------------------

	ListPersonalityStatements(ctx context.Context) ([]PersonalityStatement, error)
	UpsertPersonalityAnswer(ctx context.Context, profileID uuid.UUID, statementKey string, score int) (*ProfilePersonalityAnswer, error)
	ListPersonalityAnswers(ctx context.Context, profileID uuid.UUID) ([]ProfilePersonalityAnswer, error)
	GetPersonalityTraitScores(ctx context.Context, profileID uuid.UUID) ([]PersonalityTraitScore, error)

	// --- Preferencias de pareja -------------------------------------------

	GetPartnerPreferences(ctx context.Context, profileID uuid.UUID) (*PartnerPreferences, error)
	UpsertPartnerPreferences(ctx context.Context, profileID uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error)
}
