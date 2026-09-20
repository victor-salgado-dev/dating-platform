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
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
	MinSearchAge    = 18 // V1 es solo para mayores de edad (sección 1)
	MaxSearchAge    = 120
)

// HobbyFilter exige que el perfil haya marcado un hobby del catálogo
// como "me gusta" (liked=true). Si Min y/o Max no son nil, además exige
// que la intensidad caiga en ese rango (ambos límites inclusive).
//
// Un perfil que no contestó ese hobby, o que contestó "no me gusta",
// nunca cumple este filtro: es la misma REGLA DE LOS DATOS FALTANTES
// del resto de Filters, aplicada a una tabla aparte (profile_hobbies)
// en vez de a una columna de profiles.
type HobbyFilter struct {
	Key string
	Min *int
	Max *int
}

// PersonalityFilter exige que la media de un rasgo de personalidad
// (calculada por profiles sobre las afirmaciones contestadas) caiga en
// el rango [Min, Max] (ambos límites inclusive si no son nil). Un
// perfil que no ha contestado ninguna afirmación de ese rasgo no tiene
// media y nunca cumple el filtro — misma regla de datos faltantes.
//
// A diferencia de HobbyFilter, aquí no tiene sentido un filtro sin
// ningún límite (no hay equivalente a "liked" para un rasgo), así que
// siempre debe traer Min y/o Max.
type PersonalityFilter struct {
	TraitKey string
	Min      *float64
	Max      *float64
}

// RawBounds son los límites min/max de un filtro con clave dinámica
// (un hobby o un rasgo de personalidad concretos), todavía sin
// parsear ni validar: llegan como texto desde la query string
// (?hobby_travelling_min=4, ?trait_openness_max=3...).
type RawBounds struct {
	Min string
	Max string
}

// Filters son los criterios de búsqueda. Todos son opcionales: un
// campo vacío/nil significa "no filtrar por esto", no "buscar valores
// vacíos". Ver REGLA DE LOS DATOS FALTANTES en Repository.Search.
type Filters struct {
	Genders          []profiles.Gender
	MinAge           *int
	MaxAge           *int
	CountryCode      *string
	Languages        []string
	RelationshipGoal *profiles.RelationshipGoal

	// Cambiados de *bool a *string
	HasChildren   *string
	WantsChildren *string
	Interests     []string

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

	// --- NUEVOS: Hobbies y personalidad (Fase 2) -------------------
	// Varios filtros se combinan con AND: el perfil tiene que cumplir
	// TODOS los que se pidan (ej: jardinería<4 Y viajar>3 a la vez).
	Hobbies           []HobbyFilter
	PersonalityTraits []PersonalityFilter
}

// Params agrupa los filtros con la paginación/ordenación y quién busca
// (para excluirlo de sus propios resultados).
type Params struct {
	Filters       Filters
	ExcludeUserID uuid.UUID
	Sort          Sort
	Page          int
	PageSize      int
}

// ResultItem es la ficha resumida de un perfil en una lista de
// resultados. Deliberadamente no incluye bio/intereses/idiomas
// completos: eso pertenece a la vista de perfil público (Fase 6).
type ResultItem struct {
	ProfileID        uuid.UUID
	DisplayName      string
	Age              int
	Gender           profiles.Gender
	CountryCode      string
	Region           *string
	RelationshipGoal *profiles.RelationshipGoal
	HasPhoto         bool
	CreatedAt        time.Time
}

// Result es una página de resultados de búsqueda.
type Result struct {
	Items      []ResultItem
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}
