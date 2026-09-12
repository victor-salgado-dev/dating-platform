package messaging

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// FindOrCreateConversation devuelve el ID de la conversación entre
	// userA y userB, creándola si todavía no existe. El orden de los
	// argumentos no importa: internamente se canonicaliza.
	FindOrCreateConversation(ctx context.Context, userA, userB uuid.UUID) (uuid.UUID, error)

	// IsParticipant indica si userID es parte de la conversación conversationID.
	IsParticipant(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)

	// Participants devuelve los dos usuarios de una conversación (sin
	// orden particular respecto a quién la inició). Se usa para
	// comprobar bloqueos (Fase 9) antes de dejar continuar un chat.
	Participants(ctx context.Context, conversationID uuid.UUID) (userA, userB uuid.UUID, err error)

	// InsertMessage añade un mensaje a una conversación ya existente.
	InsertMessage(ctx context.Context, conversationID, senderID uuid.UUID, body string) (*Message, error)

	// MarkRead marca como leídos todos los mensajes de conversationID
	// que NO envió readerUserID (es decir, los que le escribieron a él).
	MarkRead(ctx context.Context, conversationID, readerUserID uuid.UUID) error

	// ListMessages pagina los mensajes de una conversación, del más
	// antiguo al más reciente.
	ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) (*MessageListResult, error)

	// ListConversations pagina las conversaciones de userID, ordenadas
	// por el mensaje más reciente primero.
	ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ConversationListResult, error)
}
