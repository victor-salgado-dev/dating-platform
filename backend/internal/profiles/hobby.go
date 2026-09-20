package profiles

import (
	"time"

	"github.com/google/uuid"
)

// HobbyDefinition es un ítem del catálogo de hobbies (tabla
// hobby_definitions). Es dato semi-estático: se puebla por migración o
// por un futuro panel de administración, nunca lo escribe el usuario
// final directamente.
type HobbyDefinition struct {
	Key       string
	Category  string
	Label     string
	SortOrder int
}

// ProfileHobby es la respuesta de un perfil a un hobby del catálogo.
//
// REGLA: si Liked es false, Intensity siempre es nil — no tiene sentido
// puntuar algo que no gusta (lo impone también un CHECK en BD). Si no
// existe ninguna fila para un (perfil, hobby), significa "no
// contestado": no se debe inventar un valor por defecto ni tratarlo
// como "no me gusta".
type ProfileHobby struct {
	ProfileID uuid.UUID
	HobbyKey  string
	Liked     bool
	Intensity *int
	UpdatedAt time.Time
}
