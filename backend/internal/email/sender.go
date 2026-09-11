// Package email abstrae el envío de correos electrónicos. La lógica de
// negocio (auth, notificaciones futuras) depende únicamente de la
// interfaz Sender, nunca de un proveedor concreto. Esto permite cambiar
// de un driver "noop"/local en desarrollo a un proveedor real (SMTP,
// SES, Postmark, ...) en producción sin tocar el resto del código.
package email

import "context"

// Message es un correo simple. Body se trata como texto plano en V1;
// una fase futura podrá añadir una versión HTML si hace falta.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender envía mensajes de email.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}
