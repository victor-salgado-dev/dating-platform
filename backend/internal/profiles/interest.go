package profiles

import (
	"time"

	"github.com/google/uuid"
)

// Compatibilidad temporal con el paquete search hasta su migración
const (
	MinHobbyIntensity = 1
	MaxHobbyIntensity = 5
)

// InterestDefinition es un ítem del catálogo de intereses (tabla
// interests). Es dato semi-estático: se puebla por migración o por un
// futuro panel de administración, nunca lo escribe el usuario final.
//
// HasLevel dice si este interés se puntúa 1-5 (ej. "Cocina": cuánto te
// gusta) o si es una simple etiqueta presente/ausente (ej. "Guitarra":
// la tocas o no la tocas, no tiene sentido puntuarla). La fila de
// respuesta (ProfileInterest) debe respetar esta bandera — esa
// correspondencia no se puede expresar como CHECK de Postgres porque
// cruza dos tablas, así que la aplica el Service antes de escribir
// (ver Service.SetInterest).
type InterestDefinition struct {
	Key       string
	Category  string
	Label     string
	HasLevel  bool
	SortOrder int
}

// ProfileInterest es la respuesta de un perfil a un interés del
// catálogo. Si no existe ninguna fila para un (perfil, interés),
// significa "no seleccionado": no se debe inventar un valor por
// defecto ni tratarlo como "no le interesa".
type ProfileInterest struct {
	ProfileID   uuid.UUID
	InterestKey string
	Level       *int
	UpdatedAt   time.Time
}