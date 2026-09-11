package email

import (
	"context"
	"log/slog"
)

// NoopSender no envía correos reales: los registra en el log.
// Es el driver por defecto en desarrollo (EMAIL_DRIVER=noop) y permite
// probar flujos completos (verificación de email, reset de contraseña)
// sin depender de un proveedor de email externo.
type NoopSender struct{}

func NewNoopSender() *NoopSender {
	return &NoopSender{}
}

var _ Sender = (*NoopSender)(nil)

func (s *NoopSender) Send(ctx context.Context, msg Message) error {
	slog.Info("email (noop, no enviado realmente)",
		"to", msg.To,
		"subject", msg.Subject,
		"body", msg.Body,
	)
	return nil
}
