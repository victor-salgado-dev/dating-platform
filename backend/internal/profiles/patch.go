package profiles

import "time"

// ProfilePatch representa una actualización parcial de un perfil
// (PATCH semántico). Para los campos obligatorios, un puntero nil
// significa "no tocar" y un puntero no-nil trae el nuevo valor.
//
// Para los campos OPCIONALES la ambigüedad de un simple puntero no
// basta (un JSON con la clave ausente y un JSON con la clave a `null`
// deben comportarse distinto: el primero no toca nada, el segundo borra
// el dato). Por eso cada campo opcional tiene un flag `...Set` aparte:
// si es false, el campo no se toca; si es true, se aplica el valor
// (que puede ser nil, es decir, "bórralo, ahora es desconocido").
//
// Quien construye este struct a partir del JSON de la petición
// (internal/profiles/handler.go) es responsable de fijar los flags
// según qué claves vinieran realmente en el body.
type ProfilePatch struct {
	DisplayName *string
	BirthDate   *time.Time
	Gender      *Gender
	CountryCode *string

	RegionSet bool
	Region    *string

	LanguagesSet bool
	Languages    []string

	RelationshipGoalSet bool
	RelationshipGoal    *RelationshipGoal

	// Cambiado de *bool a *string para soportar más opciones
	HasChildrenSet bool
	HasChildren    *string

	// Cambiado de *bool a *string
	WantsChildrenSet bool
	WantsChildren    *string

	BioSet bool
	Bio    *string

	InterestsSet bool
	Interests    []string

	// --- Físico y Apariencia ---
	HeightSet bool
	Height    *int

	WeightSet bool
	Weight    *int

	BodyTypeSet bool
	BodyType    *string

	EthnicitySet bool
	Ethnicity    *string

	AppearanceRatingSet bool
	AppearanceRating    *string

	HairColorSet bool
	HairColor    *string

	EyeColorSet bool
	EyeColor    *string

	BodyArtSet bool
	BodyArt    []string

	// --- Estilo de Vida y Familia ---
	SmokingHabitSet bool
	SmokingHabit    *string

	DrinkingHabitSet bool
	DrinkingHabit    *string

	RelocationWillingnessSet bool
	RelocationWillingness    []string

	MaritalStatusSet bool
	MaritalStatus    *string

	ChildrenCountSet bool
	ChildrenCount    *int

	YoungestChildAgeSet bool
	YoungestChildAge    *int

	OldestChildAgeSet bool
	OldestChildAge    *int

	OccupationSet bool
	Occupation    *string

	EmploymentStatusSet bool
	EmploymentStatus    *string

	IncomeLevelSet bool
	IncomeLevel    *string

	LivingSituationSet bool
	LivingSituation    *string

	// --- Fondo, Cultura y Valores ---
	NationalitySet bool
	Nationality    *string

	EducationLevelSet bool
	EducationLevel    *string

	EnglishAbilitySet bool
	EnglishAbility    *string

	ReligionSet bool
	Religion    *string

	ReligiousValuesSet bool
	ReligiousValues    *string

	StarSignSet bool
	StarSign    *string

	// --- NUEVOS CAMPOS (Über mich / estilo de vida) -----------------
	FutureVisionSet bool
	FutureVision    []string

	SportsSet bool
	Sports    []string

	LikesPetsSet bool
	LikesPets    *string

	PetsOwnedSet bool
	PetsOwned    []string

	FavoriteSeasonSet bool
	FavoriteSeason    *string

	IdealVacationStyleSet bool
	IdealVacationStyle    []string

	VacationActivitiesSet bool
	VacationActivities    []string

	ProfileQuoteSet bool
	ProfileQuote    *string

	DreamWishSet bool
	DreamWish    *string
}
