// Package messaging implementa mensajería 1:1 entre usuarios:
// conversaciones y mensajes, con lectura y paginación básica. Depende
// de profiles.Repository (interfaz) para resolver y validar a quién se
// puede escribir, igual que favorites; profiles no depende de
// messaging, así que no hay ciclos.
package messaging

import (
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

const (
	MaxBodyLength   = 2000
	DefaultPageSize = 20
	MaxPageSize     = 50
)

// Message es un mensaje dentro de una conversación. SenderID es interno:
// los handlers nunca lo serializan tal cual, solo un booleano "is_mine"
// calculado contra el usuario autenticado, para no filtrar IDs de otras
// cuentas (igual criterio que en profiles/favorites).
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Body           string
	CreatedAt      time.Time
	ReadAt         *time.Time
}

// Participant es la ficha resumida del otro participante de una
// conversación (mismo nivel de detalle que un resultado de búsqueda).
type Participant struct {
	ProfileID   uuid.UUID
	DisplayName string
	Age         int
	Gender      profiles.Gender
	CountryCode string
	Region      *string
	HasPhoto    bool
	// PhotoID es el id de la foto principal (la de position más baja),
	// resuelto en la misma consulta que lista las conversaciones para
	// evitar una consulta por ítem (N+1). Es nil si el perfil no tiene
	// fotos.
	PhotoID *uuid.UUID
}

// ConversationSummary es una fila de la lista de conversaciones.
type ConversationSummary struct {
	ConversationID    uuid.UUID
	OtherParticipant  Participant
	LastMessageBody   string
	LastMessageAt     time.Time
	LastMessageIsMine bool
	UnreadCount       int
}

type ConversationListResult struct {
	Items      []ConversationSummary
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

type MessageListResult struct {
	ConversationID uuid.UUID
	Items          []Message
	Total          int
	Page           int
	PageSize       int
	TotalPages     int
}
