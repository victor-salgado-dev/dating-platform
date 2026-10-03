package profiles

// ProfilePatch es una actualización parcial de un perfil (PATCH semántico).
//
// Cada campo es un Field[T]: Set=false significa "la clave no vino en el body,
// no tocar" y Set=true significa "aplicar Value" (que puede ser nil/vacío si
// vino a null: "bórralo, ahora es desconocido"). Ver Field.
//
// El body JSON se decodifica directamente en este struct (ver
// decodePatchOrWrite), y el repositorio recorre sus campos para construir el
// UPDATE (ver setColumns): no hay código por campo en ningún otro sitio.
//
// CONTRATO: la etiqueta json de cada campo es a la vez el nombre de la clave
// en la API y el de la columna de profiles (override con la etiqueta db si
// algún día divergen). patch_test.go comprueba que no falta ni sobra ninguna
// columna respecto a allProfileCols y a los DTOs de respuesta.
//
// Languages e Interests no están aquí: se gestionan aparte, un ítem cada vez
// (Service.SetLanguage / Service.SetInterest).
type ProfilePatch struct {
	DisplayName       Field[string]             `json:"display_name"`
	BirthDate         Field[DateOnly]           `json:"birth_date"`
	Gender            Field[Gender]             `json:"gender"`
	CountryCode       Field[string]             `json:"country_code"`
	Region            Field[*string]            `json:"region"`
	RelationshipGoals Field[[]RelationshipGoal] `json:"relationship_goals"`
	HasChildren       Field[*string]            `json:"has_children"`
	WantsChildren     Field[*string]            `json:"wants_children"`
	Bio               Field[*string]            `json:"bio"`

	// --- Físico y apariencia ---
	Height           Field[*int]     `json:"height"`
	Weight           Field[*int]     `json:"weight"`
	BodyType         Field[*string]  `json:"body_type"`
	Ethnicity        Field[*string]  `json:"ethnicity"`
	AppearanceRating Field[*string]  `json:"appearance_rating"`
	HairColor        Field[*string]  `json:"hair_color"`
	EyeColor         Field[*string]  `json:"eye_color"`
	BodyArt          Field[[]string] `json:"body_art"`

	// --- Estilo de vida y familia ---
	SmokingHabit          Field[*string]  `json:"smoking_habit"`
	DrinkingHabit         Field[*string]  `json:"drinking_habit"`
	RelocationWillingness Field[[]string] `json:"relocation_willingness"`
	MaritalStatus         Field[*string]  `json:"marital_status"`
	ChildrenCount         Field[*int]     `json:"children_count"`
	YoungestChildAge      Field[*int]     `json:"youngest_child_age"`
	OldestChildAge        Field[*int]     `json:"oldest_child_age"`
	Occupation            Field[*string]  `json:"occupation"`
	EmploymentStatus      Field[*string]  `json:"employment_status"`
	IncomeLevel           Field[*string]  `json:"income_level"`
	LivingSituation       Field[*string]  `json:"living_situation"`

	// --- Fondo, cultura y valores ---
	Nationality     Field[*string] `json:"nationality"`
	EducationLevel  Field[*string] `json:"education_level"`
	EnglishAbility  Field[*string] `json:"english_ability"`
	Religion        Field[*string] `json:"religion"`
	ReligiousValues Field[*string] `json:"religious_values"`
	StarSign        Field[*string] `json:"star_sign"`

	// --- Über mich / estilo de vida ---
	FutureVision       Field[[]string] `json:"future_vision"`
	Sports             Field[[]string] `json:"sports"`
	LikesPets          Field[*string]  `json:"likes_pets"`
	PetsOwned          Field[[]string] `json:"pets_owned"`
	FavoriteSeason     Field[*string]  `json:"favorite_season"`
	IdealVacationStyle Field[[]string] `json:"ideal_vacation_style"`
	VacationActivities Field[[]string] `json:"vacation_activities"`
	ProfileQuote       Field[*string]  `json:"profile_quote"`
	DreamWish          Field[*string]  `json:"dream_wish"`
}

// PartnerPreferencesPatch es el equivalente de ProfilePatch para las
// preferencias de pareja ("lo que busco"), con el mismo contrato: etiqueta json
// == columna de profile_partner_preferences.
type PartnerPreferencesPatch struct {
	AgeMin                      Field[*int]     `json:"age_min"`
	AgeMax                      Field[*int]     `json:"age_max"`
	HeightMin                   Field[*int]     `json:"height_min"`
	HeightMax                   Field[*int]     `json:"height_max"`
	DesiredTraits               Field[[]string] `json:"desired_traits"`
	PartnerMayHaveChildren      Field[*string]  `json:"partner_may_have_children"`
	PartnerReligionPreference   Field[*string]  `json:"partner_religion_preference"`
	AboutPartnerText            Field[*string]  `json:"about_partner_text"`
	FirstMeetingPreference      Field[*string]  `json:"first_meeting_preference"`
	DesiredLivingPlace          Field[[]string] `json:"desired_living_place"`
	ImportanceSharedThoughts    Field[*int]     `json:"importance_shared_thoughts"`
	ImportanceSharedHobbies     Field[*int]     `json:"importance_shared_hobbies"`
	ImportanceIntimacy          Field[*int]     `json:"importance_intimacy"`
	ImportanceRomanticLove      Field[*int]     `json:"importance_romantic_love"`
	ImportanceFinancialSecurity Field[*int]     `json:"importance_financial_security"`
	ImportanceFun               Field[*int]     `json:"importance_fun"`
	ImportanceSharedFriends     Field[*int]     `json:"importance_shared_friends"`
	ImportanceSharedHumor       Field[*int]     `json:"importance_shared_humor"`
	ImportancePersonalSpace     Field[*int]     `json:"importance_personal_space"`
	ImportanceIndependence      Field[*int]     `json:"importance_independence"`
}
