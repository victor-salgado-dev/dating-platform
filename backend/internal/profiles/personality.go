package profiles

import (
	"time"

	"github.com/google/uuid"
)

// PersonalityTrait son los 5 rasgos del modelo (Big Five). La lista es
// cerrada a propósito: los rasgos no cambian, solo las afirmaciones
// dentro de cada uno (por eso esas sí van en un catálogo en BD).
type PersonalityTrait string

const (
	TraitExtraversion       PersonalityTrait = "extraversion"
	TraitEmotionalStability PersonalityTrait = "emotional_stability"
	TraitConscientiousness  PersonalityTrait = "conscientiousness"
	TraitAgreeableness      PersonalityTrait = "agreeableness"
	TraitOpenness           PersonalityTrait = "openness"
)

// PersonalityStatement es una afirmación del catálogo (tabla
// personality_statements), agrupada bajo un rasgo. Dato semi-estático,
// igual que InterestDefinition.
type PersonalityStatement struct {
	Key       string
	TraitKey  PersonalityTrait
	Label     string
	SortOrder int
}

// ProfilePersonalityAnswer es la puntuación (1-5) que un perfil da a
// una afirmación concreta. Si no hay fila para un (perfil, afirmación),
// significa "no contestado".
type ProfilePersonalityAnswer struct {
	ProfileID    uuid.UUID
	StatementKey string
	Score        int
	UpdatedAt    time.Time
}

// PersonalityTraitScore es el agregado ("Gesamt") de un rasgo: la media
// de las afirmaciones contestadas para ese rasgo. Se calcula en la BD
// (vista profile_personality_trait_scores) a partir de las respuestas
// individuales — nunca se guarda de forma independiente, para que no
// pueda desincronizarse de ellas.
type PersonalityTraitScore struct {
	ProfileID     uuid.UUID
	TraitKey      PersonalityTrait
	AverageScore  float64
	AnsweredCount int
}
