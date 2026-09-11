package search

import "context"

// Repository ejecuta búsquedas de perfiles contra la fuente de verdad
// (PostgreSQL).
//
// REGLA DE LOS DATOS FALTANTES (sección 8): si un filtro exige un dato
// concreto (p. ej. idioma, objetivo de relación, deseo de hijos) y el
// perfil no lo ha indicado (NULL en la base de datos), ese perfil NO
// debe aparecer en esa búsqueda. Si el filtro no se aplica, el perfil
// puede aparecer con normalidad aunque le falte ese dato. La
// implementación PostgreSQL se apoya en la lógica de tres valores de
// SQL (una comparación contra NULL nunca es TRUE) para que esto ocurra
// de forma natural en cada cláusula, sin lógica especial por campo.
type Repository interface {
	Search(ctx context.Context, params Params) (*Result, error)
}
