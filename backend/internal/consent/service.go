package consent

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RecordRegistrationConsents registra la aceptación de Términos y
// Política de Privacidad en el momento del registro. accepted debe ser
// true explícitamente (una casilla marcada por la persona, nunca un
// valor por defecto) o devuelve ErrTermsNotAccepted.
//
// Se registran como DOS consentimientos separados con el mismo
// instante, no uno combinado: son documentos distintos y la sección 14
// pide poder tratarlos por separado más adelante (p. ej. si solo
// cambiara la Política de Privacidad, no los Términos).
func (s *Service) RecordRegistrationConsents(ctx context.Context, userID uuid.UUID, accepted bool) error {
	if !accepted {
		return ErrTermsNotAccepted
	}

	if err := s.repo.Record(ctx, userID, DocumentTerms, CurrentTermsVersion); err != nil {
		return err
	}
	return s.repo.Record(ctx, userID, DocumentPrivacyPolicy, CurrentPrivacyPolicyVersion)
}

func (s *Service) ListMyConsents(ctx context.Context, userID uuid.UUID) ([]Consent, error) {
	return s.repo.ListByUser(ctx, userID)
}
