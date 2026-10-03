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
// REGLA DE LOS DATOS FALTANTES: todos los campos opcionales (ProfileDetails)
// son punteros o slices nil. nil significa literalmente "el usuario no lo ha
// dicho": nunca se debe tratar como false, como vacío-que-cuenta, ni
// inventarse un valor. La búsqueda debe respetar esto al filtrar.
//
// Languages e Interests no viven aquí: están en profile_languages y
// profile_interests, cada uno con su propio nivel/intensidad. Se consultan con
// Service.ListMyLanguages / Service.ListMyInterests, igual que las fotos.
type Profile struct {
	ID     uuid.UUID
	UserID uuid.UUID

	DisplayName string
	BirthDate   time.Time
	Gender      Gender
	CountryCode string

	ProfileDetails

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProfileDetails son los datos que el usuario puede dejar sin contestar. Los
// cuatro obligatorios (nombre, nacimiento, género y país) están en Profile.
//
// Se declara aparte para que el cuerpo de POST /profiles se decodifique
// directamente en este struct (lo embebe createProfileRequest) y llegue tal
// cual al servicio: añadir un campo opcional nuevo es una línea aquí, más su
// columna, su Field en ProfilePatch y su campo en profileResponse. Las
// etiquetas json son los nombres de la API; patch_test.go vigila que
// coincidan con ProfilePatch y con la respuesta.
type ProfileDetails struct {
	Region *string `json:"region"`

	// Se puede buscar más de una cosa a la vez (ej. casual y long_term).
	RelationshipGoals []RelationshipGoal `json:"relationship_goals"`

	// Strings (no bool) para soportar opciones extra: "not_sure", "prefer_not_to_say"...
	HasChildren   *string `json:"has_children"`
	WantsChildren *string `json:"wants_children"`
	Bio           *string `json:"bio"`

	// --- Físico y apariencia ---
	Height           *int     `json:"height"`
	Weight           *int     `json:"weight"`
	BodyType         *string  `json:"body_type"`
	Ethnicity        *string  `json:"ethnicity"`
	AppearanceRating *string  `json:"appearance_rating"`
	HairColor        *string  `json:"hair_color"`
	EyeColor         *string  `json:"eye_color"`
	BodyArt          []string `json:"body_art"`

	// --- Estilo de vida y familia ---
	SmokingHabit          *string  `json:"smoking_habit"`
	DrinkingHabit         *string  `json:"drinking_habit"`
	RelocationWillingness []string `json:"relocation_willingness"`
	MaritalStatus         *string  `json:"marital_status"`
	ChildrenCount         *int     `json:"children_count"`
	YoungestChildAge      *int     `json:"youngest_child_age"`
	OldestChildAge        *int     `json:"oldest_child_age"`
	Occupation            *string  `json:"occupation"`
	EmploymentStatus      *string  `json:"employment_status"`
	IncomeLevel           *string  `json:"income_level"`
	LivingSituation       *string  `json:"living_situation"`

	// --- Fondo, cultura y valores ---
	Nationality     *string `json:"nationality"`
	EducationLevel  *string `json:"education_level"`
	EnglishAbility  *string `json:"english_ability"`
	Religion        *string `json:"religion"`
	ReligiousValues *string `json:"religious_values"`
	StarSign        *string `json:"star_sign"`

	// --- Über mich / estilo de vida ---
	FutureVision       []string `json:"future_vision"`
	Sports             []string `json:"sports"`
	LikesPets          *string  `json:"likes_pets"`
	PetsOwned          []string `json:"pets_owned"`
	FavoriteSeason     *string  `json:"favorite_season"`
	IdealVacationStyle []string `json:"ideal_vacation_style"`
	VacationActivities []string `json:"vacation_activities"`
	ProfileQuote       *string  `json:"profile_quote"`
	DreamWish          *string  `json:"dream_wish"`
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
