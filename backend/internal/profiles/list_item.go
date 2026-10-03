package profiles

import (
	"fmt"
	"regexp"
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
//
// Los cuatro flags de interacción (Liked, Favorited, ReceivedLike,
// ReceivedFavorite) describen la relación entre quien mira y el perfil de
// la ficha. Solo los rellenan los listados que usan ScanListItem; con
// ScanBaseListItem se quedan en false.
type BaseListItem struct {
	ProfileID   uuid.UUID
	DisplayName string
	Age         int
	Gender      Gender
	CountryCode string
	Region      *string
	HasPhoto    bool

	Liked            bool // yo le di like
	Favorited        bool // yo lo tengo en favoritos
	ReceivedLike     bool // me dio like
	ReceivedFavorite bool // me tiene en favoritos
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
	return scanListItem(row, false, extras)
}

// ScanListItem es ScanBaseListItem más los cuatro flags de interacción, que
// deben ser las ÚLTIMAS cuatro columnas del SELECT (después de los extras),
// tal y como las genera ViewerFlagsSQL:
//
//	<7 columnas base>, <extras...>, liked, favorited, received_like, received_favorite
func ScanListItem(row pgx.Row, extras ...any) (BaseListItem, error) {
	return scanListItem(row, true, extras)
}

func scanListItem(row pgx.Row, withFlags bool, extras []any) (BaseListItem, error) {
	var (
		base   BaseListItem
		birth  time.Time
		gender string
	)

	dest := make([]any, 0, 7+len(extras)+4)
	dest = append(dest,
		&base.ProfileID,
		&base.DisplayName,
		&birth,
		&gender,
		&base.CountryCode,
		&base.Region,
		&base.HasPhoto,
	)
	dest = append(dest, extras...)
	if withFlags {
		dest = append(dest, &base.Liked, &base.Favorited, &base.ReceivedLike, &base.ReceivedFavorite)
	}

	if err := row.Scan(dest...); err != nil {
		return BaseListItem{}, err
	}

	base.Age = AgeAt(birth, time.Now())
	base.Gender = Gender(gender)
	return base, nil
}

var (
	sqlAliasPattern     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	reservedFlagAliases = map[string]bool{"fl": true, "ff": true, "rl": true, "rf": true}
)

// ViewerFlagsSQL devuelve las cuatro columnas de interacción entre quien
// mira (alias de una fila de profiles, con .id y .user_id) y el perfil de la
// ficha (alias `profile`). Cada columna es un EXISTS por índice único
// (likes UNIQUE(from,to), favorites UNIQUE(user_id, favorite_profile_id)),
// así que el coste depende del tamaño de la página, no de cuántos likes o
// favoritos acumule el usuario. Si el alias del visitante viene de un
// LEFT JOIN y no tiene fila, los cuatro flags salen false.
//
// viewer y profile se interpolan en el SQL: deben ser alias literales del
// código (nunca entrada de usuario). Se valida que sean identificadores simples
// y que no choquen con los alias internos (fl, ff, rl, rf); si no, entra en
// pánico, porque es un error de programación y no de datos.
func ViewerFlagsSQL(viewer, profile string) string {
	for _, alias := range []string{viewer, profile} {
		if !sqlAliasPattern.MatchString(alias) || reservedFlagAliases[alias] {
			panic(fmt.Sprintf("profiles: alias SQL no válido en ViewerFlagsSQL: %q", alias))
		}
	}
	return fmt.Sprintf(`EXISTS (SELECT 1 FROM likes fl WHERE fl.from_profile_id = %[1]s.id AND fl.to_profile_id = %[2]s.id) AS liked,
			EXISTS (SELECT 1 FROM favorites ff WHERE ff.user_id = %[1]s.user_id AND ff.favorite_profile_id = %[2]s.id) AS favorited,
			EXISTS (SELECT 1 FROM likes rl WHERE rl.from_profile_id = %[2]s.id AND rl.to_profile_id = %[1]s.id) AS received_like,
			EXISTS (SELECT 1 FROM favorites rf WHERE rf.user_id = %[2]s.user_id AND rf.favorite_profile_id = %[1]s.id) AS received_favorite`,
		viewer, profile)
}
