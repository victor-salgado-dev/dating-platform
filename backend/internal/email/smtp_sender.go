package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPConfig son los parámetros de conexión al servidor SMTP. Pensado
// para proveedores transaccionales habituales (SendGrid, Mailgun,
// Postmark, Amazon SES...) que exponen SMTP con STARTTLS en el puerto
// 587: net/smtp.SendMail negocia STARTTLS automáticamente si el
// servidor lo anuncia, así que no hace falta gestionar TLS a mano.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPSender envía emails de verdad vía SMTP. Es el driver pensado
// para producción (EMAIL_DRIVER=smtp); NoopSender sigue siendo el de
// desarrollo.
type SMTPSender struct {
	cfg SMTPConfig
}

func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

var _ Sender = (*SMTPSender)(nil)

func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	body := buildMIMEMessage(s.cfg.From, msg)

	// net/smtp.SendMail no acepta un context.Context (es una API de
	// hace muchos años, previa a context); para V1 es una limitación
	// aceptada — el timeout real lo pone el propio servidor SMTP al
	// cortar conexiones colgadas. Si hiciera falta cancelación explícita
	// más adelante, tocaría sustituirlo por un cliente SMTP de más bajo
	// nivel (smtp.Dial + comandos manuales).
	if err := smtp.SendMail(addr, auth, s.cfg.From, []string{msg.To}, body); err != nil {
		return fmt.Errorf("email: enviar por SMTP: %w", err)
	}

	return nil
}

func buildMIMEMessage(from string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)
	return []byte(b.String())
}
