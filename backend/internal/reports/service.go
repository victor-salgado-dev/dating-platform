package reports

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

type Service struct {
	repo     Repository
	profiles profiles.Repository
}

func NewService(repo Repository, profilesRepo profiles.Repository) *Service {
	return &Service{repo: repo, profiles: profilesRepo}
}

// Create reporta al usuario detrás de targetProfileID. Usa GetByIDAny
// (no GetPublicByID): reportar tiene que funcionar aunque esa persona
// te haya bloqueado o su cuenta ya esté suspendida.
func (s *Service) Create(ctx context.Context, reporterUserID, targetProfileID uuid.UUID, reason Reason, rawDescription string) error {
	if !IsValidReason(reason) {
		return invalidField("reason", "valor no permitido")
	}

	var description *string
	if trimmed := strings.TrimSpace(rawDescription); trimmed != "" {
		if len([]rune(trimmed)) > MaxDescriptionLength {
			return invalidField("description", "no puede superar 2000 caracteres")
		}
		description = &trimmed
	}

	target, err := s.profiles.GetByIDAny(ctx, targetProfileID)
	if err != nil {
		return err
	}
	if target.UserID == reporterUserID {
		return ErrCannotReportSelf
	}

	return s.repo.Create(ctx, reporterUserID, target.UserID, reason, description)
}
