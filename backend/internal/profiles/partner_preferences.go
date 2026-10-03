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
