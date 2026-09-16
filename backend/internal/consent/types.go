// Package consent registra la aceptación de documentos legales
// (Términos, Política de Privacidad) por parte de cada usuario: qué
// versión aceptó y cuándo (sección 14). Es un registro de auditoría de
// solo-inserción: nunca se actualiza ni se borra una fila salvo por la
// cascada de eliminar la cuenta.
package consent

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType distingue QUÉ se aceptó. Deliberadamente separado: no
// se puede asumir que aceptar los Términos implica consentir
// cualquier otro tratamiento de datos (sección 14).
type DocumentType string

const (
	DocumentTerms         DocumentType = "terms"
	DocumentPrivacyPolicy DocumentType = "privacy_policy"
)

// Versiones actuales de cada documento. Si el contenido legal cambia
// de forma sustancial, esto debe incrementarse (p. ej. a "2026-02-01")
// para que quede constancia de qué versión concreta aceptó cada
// usuario y cuándo — nunca se reescribe un consentimiento ya dado.
const (
	CurrentTermsVersion         = "v1"
	CurrentPrivacyPolicyVersion = "v1"
)

// Consent es un consentimiento ya registrado.
type Consent struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	DocumentType    DocumentType
	DocumentVersion string
	AcceptedAt      time.Time
}
