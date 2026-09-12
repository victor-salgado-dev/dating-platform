package favorites

import "errors"

// ErrCannotFavoriteSelf: no tiene sentido añadirte a ti mismo a favoritos.
var ErrCannotFavoriteSelf = errors.New("favorites: no puedes añadirte a ti mismo a favoritos")
