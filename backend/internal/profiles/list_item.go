package profiles

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// BaseListItem son los campos que comparten TODAS las fichas resumidas de
// perfil que devuelven los distintos listados (favoritos, likes, visitas,
// bloqueados…): identidad, edad/género ya calculados y si el perfil tiene
// foto. Los dominios embeben este struct y añaden su propio timestamp
// (FavoritedAt, LikedAt, VisitedAt, BlockedAt…) y, si procede, campos
// como RelationshipGoal.
type BaseListItem struct {
	ProfileID   uuid.UUID
	DisplayName string
	Age         int
	Gender      Gender
	CountryCode string
	Region      *string
	HasPhoto    bool
}

// ScanBaseListItem escanea las 7 columnas base comunes a los listados de
// fichas, en este orden exacto:
//
//	p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region, <has_photo>
//
// y calcula Age a partir de birth_date. Los extras (timestamp del listado,
// COUNT(*) OVER(), etc.) se pasan como punteros adicionales y se escanean
// a continuación, en el mismo orden en que aparezcan en el SELECT.
func ScanBaseListItem(row pgx.Row, extras ...any) (BaseListItem, error) {
	var (
		base   BaseListItem
		birth  time.Time
		gender string
	)

	dest := append([]any{
		&base.ProfileID,
		&base.DisplayName,
		&birth,
		&gender,
		&base.CountryCode,
		&base.Region,
		&base.HasPhoto,
	}, extras...)

	if err := row.Scan(dest...); err != nil {
		return BaseListItem{}, err
	}

	base.Age = AgeAt(birth, time.Now())
	base.Gender = Gender(gender)
	return base, nil
}
