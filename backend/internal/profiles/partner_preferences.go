package profiles

import (
	"time"

	"github.com/google/uuid"
)

// PartnerPreferences es "lo que busco", deliberadamente separado de
// Profile ("quién soy"): son dos dominios distintos aunque compartan
// tabla de origen visual en la pestaña "Partner".
//
// Relación 1:1 opcional con profiles: mientras el usuario no rellene
// nada, no existe fila en BD. GetMyPartnerPreferences nunca devuelve
// ErrNotFound por esto — devuelve un valor con todos los campos a
// nil/zero, igual que cualquier otro campo opcional sin contestar.
type PartnerPreferences struct {
	ProfileID uuid.UUID

	AgeMin    *int
	AgeMax    *int
	HeightMin *int
	HeightMax *int

	DesiredTraits []string

	PartnerMayHaveChildren    *string
	PartnerReligionPreference *string
	AboutPartnerText          *string
	FirstMeetingPreference    *string
	DesiredLivingPlace        []string

	// "Welche Aspekte sind mir in einer Partnerschaft wie wichtig?"
	ImportanceSharedThoughts    *int
	ImportanceSharedHobbies     *int
	ImportanceIntimacy          *int
	ImportanceRomanticLove      *int
	ImportanceFinancialSecurity *int
	ImportanceFun               *int
	ImportanceSharedFriends     *int
	ImportanceSharedHumor       *int
	ImportancePersonalSpace     *int
	ImportanceIndependence      *int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// PartnerPreferencesPatch sigue la misma convención que ProfilePatch:
// todos los campos son opcionales, y cada uno tiene su flag ...Set para
// distinguir "no viene en el body" (no tocar) de "viene a null" (borrar).
type PartnerPreferencesPatch struct {
	AgeMinSet bool
	AgeMin    *int

	AgeMaxSet bool
	AgeMax    *int

	HeightMinSet bool
	HeightMin    *int

	HeightMaxSet bool
	HeightMax    *int

	DesiredTraitsSet bool
	DesiredTraits    []string

	PartnerMayHaveChildrenSet bool
	PartnerMayHaveChildren    *string

	PartnerReligionPreferenceSet bool
	PartnerReligionPreference    *string

	AboutPartnerTextSet bool
	AboutPartnerText    *string

	FirstMeetingPreferenceSet bool
	FirstMeetingPreference    *string

	DesiredLivingPlaceSet bool
	DesiredLivingPlace    []string

	ImportanceSharedThoughtsSet bool
	ImportanceSharedThoughts    *int

	ImportanceSharedHobbiesSet bool
	ImportanceSharedHobbies    *int

	ImportanceIntimacySet bool
	ImportanceIntimacy    *int

	ImportanceRomanticLoveSet bool
	ImportanceRomanticLove    *int

	ImportanceFinancialSecuritySet bool
	ImportanceFinancialSecurity    *int

	ImportanceFunSet bool
	ImportanceFun    *int

	ImportanceSharedFriendsSet bool
	ImportanceSharedFriends    *int

	ImportanceSharedHumorSet bool
	ImportanceSharedHumor    *int

	ImportancePersonalSpaceSet bool
	ImportancePersonalSpace    *int

	ImportanceIndependenceSet bool
	ImportanceIndependence    *int
}
