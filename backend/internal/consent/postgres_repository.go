package consent

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

func (r *PostgresRepository) Record(ctx context.Context, userID uuid.UUID, docType DocumentType, version string) error {
	const query = `
		INSERT INTO consents (user_id, document_type, document_version)
		VALUES ($1, $2, $3)
	`

	if _, err := r.db.Exec(ctx, query, userID, string(docType), version); err != nil {
		return fmt.Errorf("consent: registrar consentimiento: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]Consent, error) {
	const query = `
		SELECT id, user_id, document_type, document_version, accepted_at
		FROM consents
		WHERE user_id = $1
		ORDER BY accepted_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("consent: listar consentimientos: %w", err)
	}
	defer rows.Close()

	var items []Consent
	for rows.Next() {
		var c Consent
		var docType string
		if err := rows.Scan(&c.ID, &c.UserID, &docType, &c.DocumentVersion, &c.AcceptedAt); err != nil {
			return nil, fmt.Errorf("consent: leer consentimiento: %w", err)
		}
		c.DocumentType = DocumentType(docType)
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("consent: listar consentimientos: %w", err)
	}

	return items, nil
}
