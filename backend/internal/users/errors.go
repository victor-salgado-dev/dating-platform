package users

import "errors"

var (
	// ErrNotFound se devuelve cuando no existe ninguna cuenta activa
	// (no eliminada) que coincida con la búsqueda.
	ErrNotFound = errors.New("users: cuenta no encontrada")

	// ErrDuplicateEmail se devuelve al intentar crear una cuenta con un
	// email que ya está en uso por otra cuenta no eliminada.
	ErrDuplicateEmail = errors.New("users: el email ya está registrado")
)
