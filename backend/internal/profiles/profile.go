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
	RelationshipCasual     RelationshipGoal = "casual"
	RelationshipLongTerm   RelationshipGoal = "long_term"
	RelationshipFriendship RelationshipGoal = "friendship"
	RelationshipMarriage   RelationshipGoal = "marriage"
	RelationshipNotSure    RelationshipGoal = "not_sure"
)

// Profile es la entidad de dominio de perfil.
//
// REGLA DE LOS DATOS FALTANTES: todos los campos opcionales son punteros
// o slices nil. nil significa literalmente "el usuario no lo ha dicho":
// nunca se debe tratar como false, como vacío-que-cuenta, ni inventarse
// un valor. La Fase 5 (búsqueda) debe respetar esto al filtrar.
//
// Languages e Interests ya NO viven aquí (antes eran TEXT[] libres):
// pasaron a profile_languages y profile_interests respectivamente, cada
// uno con su propio nivel/intensidad. Consultarlos es responsabilidad
// de Service.ListMyLanguages / Service.ListMyInterests, no de este
// struct — igual que ya pasaba con las fotos.
type Profile struct {
	ID     uuid.UUID
	UserID uuid.UUID

	DisplayName string
	BirthDate   time.Time
	Gender      Gender
	CountryCode string

	Region *string

	// Antes era un único valor (*RelationshipGoal); ahora se puede
	// buscar más de una cosa a la vez (ej. casual y long_term).
	RelationshipGoals []RelationshipGoal

	// Han pasado a string para soportar opciones extra (ej: "not_sure", "prefer_not_to_say")
	HasChildren   *string
	WantsChildren *string
	Bio           *string

	// --- Físico y Apariencia ---
	Height           *int
	Weight           *int
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
	ChildrenCount         *int
	YoungestChildAge      *int
	OldestChildAge        *int
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

	// --- Über mich / estilo de vida ---
	FutureVision       []string
	Sports             []string
	LikesPets          *string
	PetsOwned          []string
	FavoriteSeason     *string
	IdealVacationStyle []string
	VacationActivities []string
	ProfileQuote       *string
	DreamWish          *string

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
