// Package search implementa la búsqueda de perfiles con filtros
// estructurados. Lee de las tablas de `profiles` (y `profile_photos`,
// `users`), pero no depende de la lógica interna de esos módulos: solo
// reutiliza sus tipos de valor (Gender, RelationshipGoal), que son
// vocabulario compartido del dominio, no implementación.
package search

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// Sort son los criterios de ordenación soportados. Es una lista blanca
// deliberadamente corta: nunca se interpola un valor de ordenación
// recibido del cliente directamente en SQL.
type Sort string

const (
	SortRecent  Sort = "recent"   // perfiles creados más recientemente primero (por defecto)
	SortAgeAsc  Sort = "age_asc"  // más jóvenes primero
	SortAgeDesc Sort = "age_desc" // más mayores primero
	SortPopular Sort = "popular"  // mayor actividad social de los últimos 30 días (vista profile_popularity)
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
	MinSearchAge    = 18 // V1 es solo para mayores de edad (sección 1)
	MaxSearchAge    = 120

	// MaxPage y MaxInterestFilters acotan lo que una sola petición puede pedir:
	// el OFFSET crece con la página, y cada filtro de interés es una subconsulta.
	MaxPage            = 10_000
	MaxInterestFilters = 25

	// Límites del nivel de un interés con has_level=true: los mismos que valida
	// profiles al guardarlo, para que no puedan desincronizarse.
	MinInterestLevel = profiles.MinInterestLevel
	MaxInterestLevel = profiles.MaxInterestLevel
)

// PersonalityFilter exige que la media de un rasgo de personalidad
// (calculada por profiles sobre las afirmaciones contestadas) caiga en
// el rango [Min, Max] (ambos límites inclusive si no son nil). Un
// perfil que no ha contestado ninguna afirmación de ese rasgo no tiene
// media y nunca cumple el filtro — misma regla de datos faltantes.
//
// A diferencia de un filtro de pertenencia simple, aquí no tiene
// sentido un filtro sin ningún límite, así que siempre debe traer
// Min y/o Max.
type PersonalityFilter struct {
	TraitKey string
	Min      *float64
	Max      *float64
}

// InterestFilter exige que el perfil tenga marcado un interés concreto
// del catálogo (tabla profile_interests). Si Min y/o Max no son nil,
// además exige que profile_interests.level caiga en ese rango (ambos
// límites inclusive) — solo tiene sentido para intereses cuya
// definición trae has_level=true; para el resto basta con Min=Max=nil
// (solo pertenencia, "le gusta/tiene marcado").
//
// Un perfil que no marcó ese interés, o que lo marcó sin nivel cuando
// el filtro exige un rango, nunca cumple este filtro: misma REGLA DE
// LOS DATOS FALTANTES del resto de Filters.
type InterestFilter struct {
	Key string
	Min *int
	Max *int
}

// RawBounds son los límites min/max de un filtro con clave dinámica
// (un interés o un rasgo de personalidad concretos), todavía sin
// parsear ni validar: llegan como texto desde la query string
// (?interest_travel_min=4, ?trait_openness_max=3...).
type RawBounds struct {
	Min string
	Max string
}

// Filters son los criterios de búsqueda. Todos son opcionales: un
// campo vacío/nil significa "no filtrar por esto", no "buscar valores
// vacíos". Ver REGLA DE LOS DATOS FALTANTES en Repository.Search.
type Filters struct {
	Genders           []profiles.Gender
	MinAge            *int
	MaxAge            *int
	CountryCode       *string
	Languages         []string
	RelationshipGoals []profiles.RelationshipGoal

	HasChildren   *string
	WantsChildren *string
	// Interests: varios filtros se combinan con AND (el perfil tiene que
	// cumplir TODOS los intereses que se pidan). Cada uno es pertenencia
	// simple, o pertenencia + rango de nivel si trae Min/Max.
	Interests []InterestFilter

	// --- Físico y Apariencia ---
	MinHeight        *int
	MaxHeight        *int
	MinWeight        *int
	MaxWeight        *int
	BodyType         *string
	Ethnicity        *string
	AppearanceRating *string
	HairColor        *string
	EyeColor         *string
	BodyArt          []string

	// --- Estilo de Vida y Familia ---
	SmokingHabit          *string
	DrinkingHabit         *string
	RelocationWillingness []string
	MaritalStatus         *string
	MaxChildren           *int
	Occupation            *string
	EmploymentStatus      *string
	IncomeLevel           *string
	LivingSituation       *string

	// --- Fondo, Cultura y Valores ---
	Nationality     *string
	EducationLevel  *string
	EnglishAbility  *string
	Religion        *string
	ReligiousValues *string
	StarSign        *string

	// --- Estilo de vida adicional ---
	FutureVision       []string
	Sports             []string
	LikesPets          *string
	PetsOwned          []string
	FavoriteSeason     *string
	IdealVacationStyle []string
	VacationActivities []string

	// --- Personalidad (Fase 2) --------------------------------------
	// Varios filtros se combinan con AND: el perfil tiene que cumplir
	// TODOS los rasgos que se pidan a la vez.
	PersonalityTraits []PersonalityFilter

	// OnlineNow es true cuando la búsqueda procede del endpoint
	// /search/online-now y debe limitar los resultados a perfiles con
	// usuarios activos recientemente. No se puede activar desde el
	// query string: lo fija el handler.
	OnlineNow bool
}

// ViewerScope son las restricciones que salen de QUIEN MIRA (sus
// seeking_genders obligatorios y, en "recommended", su rango de edad de
// profile_partner_preferences). No son filtros elegidos en la petición, por
// eso NO van en Filters: así la caché de listados sigue activa y se aplican
// en memoria sobre la lista compartida. Vacío/nil = sin restricción.
type ViewerScope struct {
	SeekingGenders []profiles.Gender
	MinAge         *int
	MaxAge         *int
}

// Params agrupa los filtros con la paginación/ordenación y quién busca
// (para excluirlo de sus propios resultados).
type Params struct {
	Filters       Filters
	ExcludeUserID uuid.UUID
	Sort          Sort
	Page          int
	PageSize      int

	// UseAgePrefs: aplicar también el rango de edad de las preferencias de
	// pareja de quien mira. Solo lo activa el listado "recommended".
	UseAgePrefs bool
	// Viewer lo rellena el repositorio (loadViewerScope), no el servicio.
	Viewer ViewerScope
}

// ResultItem es la ficha resumida de un perfil en una lista de
// resultados. Deliberadamente no incluye bio/intereses/idiomas
// completos: eso pertenece a la vista de perfil público (Fase 6).
type ResultItem struct {
	ProfileID         uuid.UUID
	DisplayName       string
	Age               int
	Gender            profiles.Gender
	CountryCode       string
	Region            *string
	RelationshipGoals []profiles.RelationshipGoal
	HasPhoto          bool
	// PhotoID es el id de la foto principal (la de position más baja),
	// como subconsulta escalar dentro del mismo SELECT. Es nil si el
	// perfil no tiene ninguna foto. El handler lo usa para armar
	// photo_url sin una petición adicional por resultado.
	PhotoID   *uuid.UUID
	CreatedAt time.Time

	// Relación entre quien busca y este perfil, calculada en la misma
	// consulta (ver profiles.ViewerFlagsSQL). Evita que el cliente tenga
	// que descargar sus listas de likes y favoritos para pintar las tarjetas.
	Liked            bool
	Favorited        bool
	ReceivedLike     bool
	ReceivedFavorite bool
}

// Result es una página de resultados de búsqueda.
type Result struct {
	Items      []ResultItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
