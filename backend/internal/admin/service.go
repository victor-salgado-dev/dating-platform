package admin

import (
	"context"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/reports"
	"dating-platform/backend/internal/users"
)

// SessionBlocker corta o restablece el acceso de una cuenta en el almacén de
// sesiones (Redis). Lo satisface *auth.RedisSessionStore. Es opcional: si es
// nil, la suspensión solo se aplica en el siguiente login.
type SessionBlocker interface {
	BlockUser(ctx context.Context, userID uuid.UUID, reason string) error
	UnblockUser(ctx context.Context, userID uuid.UUID) error
}

type Service struct {
	users   users.Repository
	reports reports.Repository
	blocker SessionBlocker
}

func NewService(usersRepo users.Repository, reportsRepo reports.Repository, blocker SessionBlocker) *Service {
	return &Service{users: usersRepo, reports: reportsRepo, blocker: blocker}
}

func (s *Service) ListUsers(ctx context.Context, page, pageSize int, statusFilter *users.Status) (*users.ListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.users.List(ctx, page, pageSize, statusFilter)
}

// SuspendUser suspende targetUserID. adminUserID es quien ejecuta la
// acción: no se puede suspender a uno mismo. Además de cambiar el estado,
// marca la cuenta como bloqueada para que sus sesiones abiertas dejen de
// valer de inmediato (RequireAuth ya no consulta Postgres en cada petición).
func (s *Service) SuspendUser(ctx context.Context, adminUserID, targetUserID uuid.UUID) error {
	if adminUserID == targetUserID {
		return ErrCannotActOnSelf
	}
	if err := s.users.SetStatus(ctx, targetUserID, users.StatusSuspended); err != nil {
		return err
	}
	if s.blocker != nil {
		return s.blocker.BlockUser(ctx, targetUserID, auth.BlockSuspended)
	}
	return nil
}

// ReactivateUser devuelve una cuenta suspendida a estado activo.
func (s *Service) ReactivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	if err := s.users.SetStatus(ctx, targetUserID, users.StatusActive); err != nil {
		return err
	}
	if s.blocker != nil {
		return s.blocker.UnblockUser(ctx, targetUserID)
	}
	return nil
}

func (s *Service) ListReports(ctx context.Context, page, pageSize int, statusFilter *reports.Status) (*reports.ListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.reports.List(ctx, page, pageSize, statusFilter)
}

// ResolveReport marca un reporte como revisado o descartado.
func (s *Service) ResolveReport(ctx context.Context, reportID uuid.UUID, status reports.Status) error {
	if !reports.IsValidResolution(status) {
		return invalidField("status", "debe ser 'reviewed' o 'dismissed'")
	}
	return s.reports.UpdateStatus(ctx, reportID, status)
}

const (
	defaultPageSize = 20
	maxPageSize     = 50
)

func clampPaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
