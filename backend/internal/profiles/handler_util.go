package profiles

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"dating-platform/backend/internal/auth"
	"dating-platform/backend/internal/httpx"
)

// requireUser devuelve el user_id autenticado de la petición. Si no hay,
// escribe el 401 y devuelve false: el handler solo tiene que hacer `return`.
func requireUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated", "Inicia sesión para continuar.")
	}
	return id, ok
}

// pathUUID lee un UUID de un parámetro del path. Si es inválido escribe el 400
// ("ID de <what> inválido.") y devuelve false.
func pathUUID(w http.ResponseWriter, r *http.Request, name, what string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_id", "ID de "+what+" inválido.")
		return uuid.Nil, false
	}
	return id, true
}

// decodePatchOrWrite decodifica el body de un PATCH en dst (*ProfilePatch o
// *PartnerPreferencesPatch). Si falla escribe el error HTTP y devuelve false.
//
// Primero lee el body como mapa de claves con httpx.DecodeJSON (así los límites
// de tamaño y la validación del JSON siguen siendo los de siempre) y después lo
// decodifica en el struct: ver unmarshalPatch.
func decodePatchOrWrite(w http.ResponseWriter, r *http.Request, dst any) bool {
	var raw map[string]json.RawMessage
	if err := httpx.DecodeJSON(r, &raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "El cuerpo de la petición no es válido.")
		return false
	}
	if err := unmarshalPatch(raw, dst); err != nil {
		writeProfileError(w, err)
		return false
	}
	return true
}

// mapSlice aplica f a cada elemento. Devuelve SIEMPRE un slice no nil (vacío
// si in lo está): en JSON sale `[]`, nunca `null`.
func mapSlice[T, R any](in []T, f func(T) R) []R {
	out := make([]R, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}
