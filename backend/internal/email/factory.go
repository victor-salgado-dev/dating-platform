package email

import "fmt"

// NewSender construye el Sender configurado mediante EMAIL_DRIVER:
// "noop" (desarrollo, registra en logs) o "smtp" (producción, Fase 14).
// smtpCfg se ignora si driver no es "smtp".
func NewSender(driver string, smtpCfg SMTPConfig) (Sender, error) {
	switch driver {
	case "", "noop":
		return NewNoopSender(), nil
	case "smtp":
		if smtpCfg.Host == "" || smtpCfg.From == "" {
			return nil, fmt.Errorf("email: configuración SMTP incompleta (host/from)")
		}
		return NewSMTPSender(smtpCfg), nil
	default:
		return nil, fmt.Errorf("email: driver desconocido %q", driver)
	}
}
