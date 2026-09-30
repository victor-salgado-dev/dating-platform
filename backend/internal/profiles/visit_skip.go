package profiles

import (
	"context"
	"net/http"
)

type skipVisitKey struct{}

// SkipVisitFromQuery marca la petición para que GetFullPublicProfile NO
// registre una visita cuando llega ?visit=0. Lo usa el cliente al PRECARGAR
// el siguiente perfil de Quick Match: sin esto, precargar contaría como
// visita a alguien que el usuario aún no ha visto. Se monta en routes.go
// sobre GET /profiles/{profileID}/full, así no hace falta tocar el handler.
func SkipVisitFromQuery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("visit") == "0" {
			r = r.WithContext(context.WithValue(r.Context(), skipVisitKey{}, true))
		}
		next.ServeHTTP(w, r)
	})
}

func visitSkipped(ctx context.Context) bool {
	v, _ := ctx.Value(skipVisitKey{}).(bool)
	return v
}
