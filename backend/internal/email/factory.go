package email

import "fmt"

// NewSender construye el Sender configurado mediante EMAIL_DRIVER.
// En V1 solo existe el driver "noop". Fases posteriores (producción)
// podrán añadir un driver SMTP/API real aquí sin que auth ni ningún
// otro consumidor de Sender tengan que cambiar una sola línea.
func NewSender(driver string) (Sender, error) {
	switch driver {
	case "", "noop":
		return NewNoopSender(), nil
	default:
		return nil, fmt.Errorf("email: driver desconocido %q", driver)
	}
}
