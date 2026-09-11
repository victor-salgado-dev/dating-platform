package storage

import "errors"

// ErrNotFound se devuelve cuando no existe ningún fichero bajo la key
// solicitada, independientemente del driver concreto.
var ErrNotFound = errors.New("storage: fichero no encontrado")
