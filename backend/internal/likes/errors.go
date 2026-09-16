package likes

import "errors"

var ErrCannotLikeSelf = errors.New("likes: no puedes dar like a tu propio perfil")
