package profiles

import (
	"time"

	"github.com/google/uuid"
)

// ProfileLanguage es un idioma que un perfil ha indicado, con un nivel
// opcional (1-5: "no contestado el nivel" es distinto de "nivel bajo").
//
// La lista de códigos permitidos vive como CHECK en la base de datos
// (profile_languages.language_code): es una lista cerrada y estable de
// códigos ISO, así que no hace falta duplicarla aquí — un código no
// reconocido se traduce en un ValidationError vía ese CHECK, igual que
// el resto de campos de selección de profiles (body_type, religion...).
type ProfileLanguage struct {
	ProfileID    uuid.UUID
	LanguageCode string
	Level        *int
	UpdatedAt    time.Time
}
