// Package pagination reúne las constantes y helpers compartidos por todos
// los listados paginados. Antes cada dominio reimplementaba su propio
// totalPages con semánticas ligeramente distintas en el borde (total == 0
// vs total <= 0); aquí hay una única fuente de verdad.
package pagination

const (
	// DefaultPageSize es el tamaño de página por defecto si el cliente no
	// indica page_size.
	DefaultPageSize = 20

	// MaxPageSize acota el tamaño máximo de página para que no se pueda
	// pedir una página gigante (defensa básica frente a abuso).
	MaxPageSize = 50
)

// TotalPages calcula el número de páginas para un total y un tamaño de
// página dados. Devuelve 0 cuando no hay elementos o cuando los valores de
// entrada no tienen sentido.
func TotalPages(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
