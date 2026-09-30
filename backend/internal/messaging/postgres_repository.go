package messaging

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) FindOrCreateConversation(ctx context.Context, userA, userB uuid.UUID) (uuid.UUID, error) {
	one, two := orderPair(userA, userB)

	const insertQuery = `
		INSERT INTO conversations (user_one_id, user_two_id)
		VALUES ($1, $2)
		ON CONFLICT (user_one_id, user_two_id) DO NOTHING
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, insertQuery, one, two).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("messaging: crear conversación: %w", err)
	}

	// ON CONFLICT DO NOTHING no devolvió fila: ya existía, la buscamos.
	const selectQuery = `SELECT id FROM conversations WHERE user_one_id = $1 AND user_two_id = $2`
	if err := r.db.QueryRow(ctx, selectQuery, one, two).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("messaging: buscar conversación existente: %w", err)
	}

	return id, nil
}

func (r *PostgresRepository) IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM conversations
			WHERE id = $1 AND (user_one_id = $2 OR user_two_id = $2)
		)
	`

	var ok bool
	if err := r.db.QueryRow(ctx, query, conversationID, userID).Scan(&ok); err != nil {
		return false, fmt.Errorf("messaging: comprobar participante: %w", err)
	}
	return ok, nil
}

func (r *PostgresRepository) Participants(ctx context.Context, conversationID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	const query = `SELECT user_one_id, user_two_id FROM conversations WHERE id = $1`

	var a, b uuid.UUID
	if err := r.db.QueryRow(ctx, query, conversationID).Scan(&a, &b); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, uuid.Nil, ErrConversationNotFound
		}
		return uuid.Nil, uuid.Nil, fmt.Errorf("messaging: leer participantes: %w", err)
	}
	return a, b, nil
}

func (r *PostgresRepository) InsertMessage(ctx context.Context, conversationID, senderID uuid.UUID, body string) (*Message, error) {
	const query = `
		INSERT INTO messages (conversation_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, read_at
	`

	msg := &Message{ConversationID: conversationID, SenderID: senderID, Body: body}
	err := r.db.QueryRow(ctx, query, conversationID, senderID, body).
		Scan(&msg.ID, &msg.CreatedAt, &msg.ReadAt)
	if err != nil {
		return nil, fmt.Errorf("messaging: enviar mensaje: %w", err)
	}

	return msg, nil
}

func (r *PostgresRepository) MarkRead(ctx context.Context, conversationID, readerUserID uuid.UUID) error {
	const query = `
		UPDATE messages
		SET read_at = now()
		WHERE conversation_id = $1 AND sender_id <> $2 AND read_at IS NULL
	`

	if _, err := r.db.Exec(ctx, query, conversationID, readerUserID); err != nil {
		return fmt.Errorf("messaging: marcar como leído: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) (*MessageListResult, error) {
	const query = `
		SELECT id, sender_id, body, created_at, read_at, COUNT(*) OVER() AS total_count
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, conversationID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("messaging: listar mensajes: %w", err)
	}
	defer rows.Close()

	var items []Message
	total := 0

	for rows.Next() {
		var m Message
		var totalCount int
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Body, &m.CreatedAt, &m.ReadAt, &totalCount); err != nil {
			return nil, fmt.Errorf("messaging: leer mensaje: %w", err)
		}
		m.ConversationID = conversationID
		total = totalCount
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messaging: listar mensajes: %w", err)
	}

	return &MessageListResult{
		ConversationID: conversationID,
		Items:          items,
		Total:          total,
		Page:           page,
		PageSize:       pageSize,
		TotalPages:     totalPages(total, pageSize),
	}, nil
}

func (r *PostgresRepository) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ConversationListResult, error) {
	const query = `
		SELECT
			c.id,
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			(SELECT ph.id FROM profile_photos ph
			     WHERE ph.profile_id = p.id
			     ORDER BY ph.position ASC, ph.id ASC
			     LIMIT 1) AS photo_id,
			lm.body, lm.created_at, lm.sender_id,
			(
				SELECT COUNT(*) FROM messages um
				WHERE um.conversation_id = c.id AND um.sender_id <> $1 AND um.read_at IS NULL
			) AS unread_count,
			COUNT(*) OVER() AS total_count
		FROM conversations c
		JOIN profiles p
			ON p.user_id = (CASE WHEN c.user_one_id = $1 THEN c.user_two_id ELSE c.user_one_id END)
		JOIN LATERAL (
			SELECT body, created_at, sender_id
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON TRUE
		WHERE (c.user_one_id = $1 OR c.user_two_id = $1)
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = $1 AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = $1)
		  )
		ORDER BY lm.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("messaging: listar conversaciones: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var items []ConversationSummary
	total := 0

	for rows.Next() {
		var (
			item         ConversationSummary
			genderStr    string
			birthDate    time.Time
			photoID      *uuid.UUID
			lastSenderID uuid.UUID
			totalCount   int
		)

		if err := rows.Scan(
			&item.ConversationID,
			&item.OtherParticipant.ProfileID, &item.OtherParticipant.DisplayName, &birthDate,
			&genderStr, &item.OtherParticipant.CountryCode, &item.OtherParticipant.Region,
			&item.OtherParticipant.HasPhoto,
			&photoID,
			&item.LastMessageBody, &item.LastMessageAt, &lastSenderID,
			&item.UnreadCount, &totalCount,
		); err != nil {
			return nil, fmt.Errorf("messaging: leer conversación: %w", err)
		}

		item.OtherParticipant.Age = profiles.AgeAt(birthDate, now)
		item.OtherParticipant.Gender = profiles.Gender(genderStr)
		item.OtherParticipant.PhotoID = photoID
		item.LastMessageIsMine = lastSenderID == userID

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("messaging: listar conversaciones: %w", err)
	}

	return &ConversationListResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages(total, pageSize),
	}, nil
}

// orderPair devuelve (a, b) en un orden canónico y estable (comparación
// byte a byte, igual que el operador < de PostgreSQL sobre uuid), para
// que buscar la conversación entre dos personas sea siempre una
// igualdad exacta sin importar quién escribió a quién primero.
func orderPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}

func totalPages(total, pageSize int) int {
	if total <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}
