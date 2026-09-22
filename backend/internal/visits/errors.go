package visits

import "errors"

var (
ErrCannotVisitSelf = errors.New("visits: no puedes visitar tu propio perfil")
)
