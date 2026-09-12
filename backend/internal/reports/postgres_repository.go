package reports

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(ctx context.Context, reporterID, reportedID uuid.UUID, reason Reason, description *string) error {
	const query = `
		INSERT INTO reports (reporter_id, reported_id, reason, description)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := r.db.Exec(ctx, query, reporterID, reportedID, string(reason), description); err != nil {
		return fmt.Errorf("reports: crear reporte: %w", err)
	}
	return nil
}

func (r *PostgresRepository) List(ctx context.Context, page, pageSize int, statusFilter *Status) (*ListResult, error) {
	where := ""
	args := []any{}
	if statusFilter != nil {
		where = "WHERE r.status = $1"
		args = append(args, string(*statusFilter))
	}

	args = append(args, pageSize, (page-1)*pageSize)
	limitArg := len(args) - 1
	offsetArg := len(args)

	query := fmt.Sprintf(`
		SELECT
			r.id, r.reporter_id, p1.display_name, r.reported_id, p2.display_name,
			r.reason, r.description, r.status, r.created_at,
			COUNT(*) OVER() AS total_count
		FROM reports r
		LEFT JOIN profiles p1 ON p1.user_id = r.reporter_id
		LEFT JOIN profiles p2 ON p2.user_id = r.reported_id
		%s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArg, offsetArg)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("reports: listar reportes: %w", err)
	}
	defer rows.Close()

	var items []ListItem
	total := 0

	for rows.Next() {
		var (
			item       ListItem
			reasonStr  string
			statusStr  string
			totalCount int
		)

		if err := rows.Scan(
			&item.ID, &item.ReporterID, &item.ReporterName, &item.ReportedID, &item.ReportedName,
			&reasonStr, &item.Description, &statusStr, &item.CreatedAt, &totalCount,
		); err != nil {
			return nil, fmt.Errorf("reports: leer reporte: %w", err)
		}

		item.Reason = Reason(reasonStr)
		item.Status = Status(statusStr)
		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reports: listar reportes: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &ListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error {
	const query = `UPDATE reports SET status = $2 WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id, string(status))
	if err != nil {
		return fmt.Errorf("reports: actualizar estado: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
