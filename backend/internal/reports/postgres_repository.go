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
