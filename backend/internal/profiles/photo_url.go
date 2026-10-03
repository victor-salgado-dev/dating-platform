package profiles

import (
	"fmt"

	"github.com/google/uuid"
)

// URLs de los ficheros de foto. Son las rutas que registra server/routes.go
// para ServePhoto (propias) y ServePublicPhoto (de otro perfil); aquí es el
// ÚNICO sitio que las construye (las usan el handler y AccountSummary).
//
// El sufijo de miniatura debe coincidir con lo que reconoce wantsThumb;
// photo_url_test.go lo comprueba.
const thumbQuery = "?size=thumb"

// ownPhotoURL es la URL de una foto del propio usuario.
func ownPhotoURL(photoID uuid.UUID, thumb bool) string {
	return withThumb(fmt.Sprintf("/api/v1/profiles/me/photos/%s/file", photoID), thumb)
}

// publicPhotoURL es la URL de una foto de un perfil (visible para quien mira).
func publicPhotoURL(profileID, photoID uuid.UUID, thumb bool) string {
	return withThumb(fmt.Sprintf("/api/v1/profiles/%s/photos/%s/file", profileID, photoID), thumb)
}

func withThumb(url string, thumb bool) string {
	if thumb {
		return url + thumbQuery
	}
	return url
}
