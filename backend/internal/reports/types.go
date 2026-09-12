// Package reports implementa la creación de reportes de usuarios. La
// revisión/gestión de reportes (panel de moderación) es la Fase 10;
// este módulo solo permite crearlos. Depende de profiles.Repository
// (vía GetByIDAny) para resolver profile_id -> user_id sin aplicar
// reglas de visibilidad: reportar a alguien tiene que funcionar incluso
// si esa persona te ha bloqueado o su cuenta ya está suspendida.
package reports

import "github.com/google/uuid"

// Reason son los motivos de reporte permitidos en V1.
type Reason string

const (
	ReasonSpam                 Reason = "spam"
	ReasonFakeProfile          Reason = "fake_profile"
	ReasonHarassment           Reason = "harassment"
	ReasonInappropriateContent Reason = "inappropriate_content"
	ReasonUnderage             Reason = "underage"
	ReasonOther                Reason = "other"
)

var allowedReasons = map[Reason]bool{
	ReasonSpam: true, ReasonFakeProfile: true, ReasonHarassment: true,
	ReasonInappropriateContent: true, ReasonUnderage: true, ReasonOther: true,
}

func IsValidReason(r Reason) bool {
	return allowedReasons[r]
}

const MaxDescriptionLength = 2000

// Report es un reporte ya creado.
type Report struct {
	ID          uuid.UUID
	ReporterID  uuid.UUID
	ReportedID  uuid.UUID
	Reason      Reason
	Description *string
}
