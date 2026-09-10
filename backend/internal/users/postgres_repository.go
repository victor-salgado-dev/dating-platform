package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgUniqueViolation es el código de error de PostgreSQL para violaciones
// de restricciones UNIQUE (23505).
const pgUniqueViolation = "23505"

// PostgresRepository implementa Repository usando PostgreSQL como
// fuente principal de verdad, vía un pool de conexiones pgx.
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository crea un repositorio de users respaldado por
// PostgreSQL. db normalmente proviene de internal/db.NewPool.
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(ctx context.Context, u *User) error {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, status, email_verified_at, created_at, updated_at, deleted_at
	`

	err := r.db.QueryRow(ctx, query, u.Email, u.PasswordHash).Scan(
		&u.ID, &u.Status, &u.EmailVerifiedAt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("users: crear cuenta: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const query = `
		SELECT id, email, password_hash, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanOne(ctx, query, id)
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = `
		SELECT id, email, password_hash, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE lower(email) = lower($1) AND deleted_at IS NULL
	`

	return r.scanOne(ctx, query, email)
}

func (r *PostgresRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE users
		SET deleted_at = now(), status = 'deleted'
		WHERE id = $1 AND deleted_at IS NULL
	`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("users: eliminar cuenta: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) scanOne(ctx context.Context, query string, args ...any) (*User, error) {
	var u User

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Status, &u.EmailVerifiedAt,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("users: consultar cuenta: %w", err)
	}

	return &u, nil
}
