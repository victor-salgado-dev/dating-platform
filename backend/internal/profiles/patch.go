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

	HasChildrenSet bool
	HasChildren    *bool

	WantsChildrenSet bool
	WantsChildren    *bool

	BioSet bool
	Bio    *string

	InterestsSet bool
	Interests    []string
}
