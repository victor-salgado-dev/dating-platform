package messaging

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"dating-platform/backend/internal/blocking"
	"dating-platform/backend/internal/profiles"
)

type Service struct {
	repo     Repository
	profiles profiles.Repository
	blocks   blocking.Repository
}

func NewService(repo Repository, profilesRepo profiles.Repository, blocksRepo blocking.Repository) *Service {
	return &Service{repo: repo, profiles: profilesRepo, blocks: blocksRepo}
}

// SendToProfile envía un mensaje a la persona detrás de targetProfileID,
// creando la conversación si es la primera vez que se escriben. Valida
// lo mismo que favorites.Add: el perfil debe ser visible públicamente
// (profiles.ErrNotFound cubre "no existe", "cuenta inactiva" y "hay un
// bloqueo mutuo" por igual) y no puede ser el propio perfil del remitente.
func (s *Service) SendToProfile(ctx context.Context, senderUserID, targetProfileID uuid.UUID, rawBody string) (*Message, error) {
	body, err := validateBody(rawBody)
	if err != nil {
		return nil, err
	}

	target, err := s.profiles.GetPublicByID(ctx, targetProfileID, senderUserID)
	if err != nil {
		return nil, err
	}
	if target.UserID == senderUserID {
		return nil, ErrCannotMessageSelf
	}

	conversationID, err := s.repo.FindOrCreateConversation(ctx, senderUserID, target.UserID)
	if err != nil {
		return nil, err
	}

	return s.repo.InsertMessage(ctx, conversationID, senderUserID, body)
}

// SendInConversation añade un mensaje a una conversación en la que
// senderUserID ya es participante (continuar un chat abierto). A
// diferencia de SendToProfile, esta ruta no pasa por GetPublicByID (no
// hace falta volver a resolver el perfil), así que el bloqueo se
// comprueba aquí explícitamente: si cualquiera de los dos ha bloqueado
// al otro desde que empezó la conversación, no se puede seguir escribiendo.
func (s *Service) SendInConversation(ctx context.Context, senderUserID, conversationID uuid.UUID, rawBody string) (*Message, error) {
	body, err := validateBody(rawBody)
	if err != nil {
		return nil, err
	}

	userA, userB, err := s.repo.Participants(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if senderUserID != userA && senderUserID != userB {
		return nil, ErrConversationNotFound
	}

	other := userA
	if senderUserID == userA {
		other = userB
	}

	blocked, err := s.blocks.IsBlocked(ctx, senderUserID, other)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrConversationNotFound
	}

	return s.repo.InsertMessage(ctx, conversationID, senderUserID, body)
}

// ListMessages pagina los mensajes de una conversación y, como efecto
// secundario (la "lectura" de la Fase 8), marca como leídos los que le
// escribieron a userID.
func (s *Service) ListMessages(ctx context.Context, userID, conversationID uuid.UUID, page, pageSize int) (*MessageListResult, error) {
	isParticipant, err := s.repo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrConversationNotFound
	}

	if err := s.repo.MarkRead(ctx, conversationID, userID); err != nil {
		return nil, err
	}

	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListMessages(ctx, conversationID, page, pageSize)
}

func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ConversationListResult, error) {
	page, pageSize = clampPaging(page, pageSize)
	return s.repo.ListConversations(ctx, userID, page, pageSize)
}

func validateBody(raw string) (string, error) {
	body := strings.TrimSpace(raw)
	if body == "" {
		return "", invalidField("body", "no puede estar vacío")
	}
	if len([]rune(body)) > MaxBodyLength {
		return "", invalidField("body", "no puede superar 2000 caracteres")
	}
	return body, nil
}

func clampPaging(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
