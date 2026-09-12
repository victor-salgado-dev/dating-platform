package admin

import (
	"context"

	"github.com/google/uuid"

	"dating-platform/backend/internal/reports"
	"dating-platform/backend/internal/users"
)

type Service struct {
	users   users.Repository
	reports reports.Repository
}

func NewService(usersRepo users.Repository, reportsRepo reports.Repository) *Service {
	return &Service{users: usersRepo, reports: reportsRepo}
}

func (s *Service) ListUsers(ctx context.Context, page, pageSize int, statusFilter *users.Status) (*users.ListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.users.List(ctx, page, pageSize, statusFilter)
}

// SuspendUser suspende targetUserID. adminUserID es quien ejecuta la
// acción: no se puede suspender a uno mismo.
func (s *Service) SuspendUser(ctx context.Context, adminUserID, targetUserID uuid.UUID) error {
	if adminUserID == targetUserID {
		return ErrCannotActOnSelf
	}
	return s.users.SetStatus(ctx, targetUserID, users.StatusSuspended)
}

// ReactivateUser devuelve una cuenta suspendida a estado activo.
func (s *Service) ReactivateUser(ctx context.Context, targetUserID uuid.UUID) error {
	return s.users.SetStatus(ctx, targetUserID, users.StatusActive)
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
