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
	HasChildren      *bool
	WantsChildren    *bool
	Interests        []string
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
