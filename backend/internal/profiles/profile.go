// Package profiles contiene el dominio de PERFIL: los datos
// estructurados y públicos de una cuenta (display name, edad, país,
// intereses, fotos...), separados de la cuenta en sí (paquete users).
package profiles

import (
	"time"

	"github.com/google/uuid"
)

// Gender son los valores permitidos para el campo género. La lista es
// deliberadamente corta para V1; ampliarla no requiere cambiar el tipo,
// solo añadir constantes y actualizar la validación y el CHECK de la BD.
type Gender string

const (
	GenderFemale    Gender = "female"
	GenderMale      Gender = "male"
	GenderNonBinary Gender = "non_binary"
	GenderOther     Gender = "other"
)

// RelationshipGoal son los valores permitidos para el objetivo de relación.
type RelationshipGoal string

const (
	RelationshipCasual    RelationshipGoal = "casual"
	RelationshipLongTerm  RelationshipGoal = "long_term"
	RelationshipFriendship RelationshipGoal = "friendship"
	RelationshipMarriage  RelationshipGoal = "marriage"
	RelationshipNotSure   RelationshipGoal = "not_sure"
)

// Profile es la entidad de dominio de perfil.
//
// REGLA DE LOS DATOS FALTANTES: todos los campos opcionales son punteros
// o slices nil. nil significa literalmente "el usuario no lo ha dicho":
// nunca se debe tratar como false, como vacío-que-cuenta, ni inventarse
// un valor. La Fase 5 (búsqueda) debe respetar esto al filtrar.
type Profile struct {
	ID     uuid.UUID
	UserID uuid.UUID

	DisplayName string
	BirthDate   time.Time
	Gender      Gender
	CountryCode string

	Region            *string
	Languages         []string
	RelationshipGoal  *RelationshipGoal
	HasChildren       *bool
	WantsChildren     *bool
	Bio               *string
	Interests         []string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Age calcula la edad actual a partir de la fecha de nacimiento. No se
// almacena en la base de datos para evitar que quede desactualizada.
func (p *Profile) Age() int {
	return AgeAt(p.BirthDate, time.Now())
}

// AgeAt calcula la edad a partir de una fecha de nacimiento, a fecha
// `now`. Exportada porque otros módulos (p. ej. search, para traducir
// filtros de edad en rangos de fecha de nacimiento) necesitan la misma
// lógica exacta: debe haber un único sitio que decida qué es "cumplir
// años".
func AgeAt(birthDate, now time.Time) int {
	age := now.Year() - birthDate.Year()
	hadBirthdayThisYear := now.Month() > birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() >= birthDate.Day())
	if !hadBirthdayThisYear {
		age--
	}
	return age
}
