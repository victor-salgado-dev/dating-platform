// Package httpx contiene helpers HTTP mínimos y compartidos entre
// módulos, para que todos los endpoints de la API devuelvan JSON y
// errores con el mismo formato (sección 9: estructura consistente).
package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse es el formato estándar de error de toda la API.
// Nunca debe incluir stack traces ni detalles internos.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON serializa payload como JSON con el status indicado.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError escribe un error en el formato estándar de la API.
// message debe ser un texto seguro para mostrar al cliente: nunca el
// error interno crudo (err.Error()) que pueda filtrar detalles.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

// DecodeJSON parsea el cuerpo de la petición como JSON en dst.
func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
