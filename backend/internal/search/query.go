package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// onlineNowWindowSeconds es la ventana en la que un usuario cuenta como "en línea".
const onlineNowWindowSeconds = 15 * 60

// builtQuery es una búsqueda lista para ejecutar.
//
// La búsqueda se reparte para que su coste no crezca con el número de perfiles:
//   - SQL devuelve solo los ids de la página, sin recuento: con LIMIT, Postgres
//     se detiene en cuanto junta PageSize coincidencias en lugar de recorrerlas
//     todas.
//   - El repositorio carga después los datos de tarjeta de esos pocos ids
//     (loadCards).
//   - El total sale de CountSQL, solo cuando hace falta y cacheado unos segundos
//     (ver count_cache.go).
type builtQuery struct {
	SQL  string // fase 1: ids de la página pedida
	Args []any

	// CountSQL/CountArgs cuentan los perfiles que cumplen los filtros, sin
	// paginación. Se usan para el total cuando la página no lo da por sí sola.
	CountSQL  string
	CountArgs []any
}

// queryBuilder acumula condiciones WHERE y sus argumentos manteniendo los
// placeholders ($1, $2...) alineados con args.
type queryBuilder struct {
	where []string
	args  []any
}

// add añade una condición con UN parámetro: clause lleva "$%d" donde va el
// placeholder. El texto de clause son siempre literales del código, nunca
// entrada del usuario; los valores van siempre como argumentos.
func (b *queryBuilder) add(clause string, val any) {
	b.args = append(b.args, val)
	b.where = append(b.where, fmt.Sprintf(clause, len(b.args)))
}

// eq añade `p.<col> = $n` si el filtro viene informado. Un filtro nil no añade
// nada: así un perfil sin ese dato sigue apareciendo (ver REGLA DE LOS DATOS
// FALTANTES en Repository); y cuando sí se filtra, NULL = valor nunca es TRUE.
func eq[T any](b *queryBuilder, col string, v *T) {
	if v != nil {
		b.add("p."+col+" = $%d", *v)
	}
}

// bound añade `p.<col> <op> $n` (op es "<=" o ">=") si el límite viene informado.
func bound(b *queryBuilder, col, op string, v *int) {
	if v != nil {
		b.add("p."+col+" "+op+" $%d", *v)
	}
}

// overlaps añade `p.<col> && $n` (las listas comparten algún elemento) si hay valores.
func overlaps(b *queryBuilder, col string, vals []string) {
	if len(vals) > 0 {
		b.add("p."+col+" && $%d", vals)
	}
}

// buildSearchQuery arma la consulta de búsqueda. Es pura (no toca la base de
// datos y recibe `now`) para poder probarla sin Postgres.
//
// El orden de las condiciones es el de siempre; los filtros que no vienen
// informados no añaden nada.
func buildSearchQuery(params Params, now time.Time) builtQuery {
	// $1 = quien busca. Visibilidad pública (cuenta activa, sin bloqueos en ningún
	// sentido) con la MISMA definición que el resto de la app, y nunca uno mismo.
	b := &queryBuilder{
		where: []string{profiles.VisibleSQL("$1"), "p.user_id <> $1"},
		args:  []any{params.ExcludeUserID},
	}
	f := params.Filters

	// Género: si la petición trae gender= explícito manda ese; si no, se usa
	// el género buscado por quien mira (obligatorio en su perfil).
	wantedGenders := f.Genders
	if len(wantedGenders) == 0 {
		wantedGenders = params.Viewer.SeekingGenders
	}
	if len(wantedGenders) > 0 {
		genders := make([]string, len(wantedGenders))
		for i, g := range wantedGenders {
			genders[i] = string(g)
		}
		b.add("p.gender = ANY($%d)", genders)
	}

	// Edad: el rango se trata como una unidad. Si en la búsqueda se rellenó
	// min_age y/o max_age, manda ese rango (el límite vacío = sin límite). Solo
	// si no se rellenó ninguno se usa el rango de las preferencias de pareja.
	minAge, maxAge := f.MinAge, f.MaxAge
	if minAge == nil && maxAge == nil {
		minAge, maxAge = params.Viewer.MinAge, params.Viewer.MaxAge
	}
	minBirth, maxBirth := ageBounds(now, minAge, maxAge)
	if minBirth != nil {
		b.add("p.birth_date <= $%d", *minBirth)
	}
	if maxBirth != nil {
		b.add("p.birth_date >= $%d", *maxBirth)
	}

	if f.CountryCode != nil {
		b.add("p.country_code = $%d", *f.CountryCode)
	}

	// Idiomas: viven en la tabla profile_languages, no en profiles.
	if len(f.Languages) > 0 {
		b.add(`EXISTS (
			SELECT 1 FROM profile_languages pl
			WHERE pl.profile_id = p.id AND pl.language_code = ANY($%d)
		)`, f.Languages)
	}

	// Intereses: viven en profile_interests. Cada filtro es pertenencia simple, o
	// pertenencia + rango de nivel si trae Min/Max (solo con has_level=true).
	for _, itf := range f.Interests {
		b.interestExists(itf)
	}

	// relationship_goals es una columna text[] (índice GIN).
	if len(f.RelationshipGoals) > 0 {
		goals := make([]string, len(f.RelationshipGoals))
		for i, g := range f.RelationshipGoals {
			goals[i] = string(g)
		}
		b.add("p.relationship_goals && $%d", goals)
	}

	eq(b, "has_children", f.HasChildren)
	eq(b, "wants_children", f.WantsChildren)

	// --- Físico y apariencia ---
	bound(b, "height", ">=", f.MinHeight)
	bound(b, "height", "<=", f.MaxHeight)
	bound(b, "weight", ">=", f.MinWeight)
	bound(b, "weight", "<=", f.MaxWeight)
	eq(b, "body_type", f.BodyType)
	eq(b, "ethnicity", f.Ethnicity)
	eq(b, "appearance_rating", f.AppearanceRating)
	eq(b, "hair_color", f.HairColor)
	eq(b, "eye_color", f.EyeColor)
	overlaps(b, "body_art", f.BodyArt)

	// --- Estilo de vida y familia ---
	eq(b, "smoking_habit", f.SmokingHabit)
	eq(b, "drinking_habit", f.DrinkingHabit)
	overlaps(b, "relocation_willingness", f.RelocationWillingness)
	eq(b, "marital_status", f.MaritalStatus)
	bound(b, "children_count", "<=", f.MaxChildren)
	eq(b, "occupation", f.Occupation)
	eq(b, "employment_status", f.EmploymentStatus)
	eq(b, "income_level", f.IncomeLevel)
	eq(b, "living_situation", f.LivingSituation)

	// --- Fondo, cultura y valores ---
	eq(b, "nationality", f.Nationality)
	eq(b, "education_level", f.EducationLevel)
	eq(b, "english_ability", f.EnglishAbility)
	eq(b, "religion", f.Religion)
	eq(b, "religious_values", f.ReligiousValues)
	eq(b, "star_sign", f.StarSign)

	// --- Estilo de vida adicional ---
	overlaps(b, "future_vision", f.FutureVision)
	overlaps(b, "sports", f.Sports)
	eq(b, "likes_pets", f.LikesPets)
	overlaps(b, "pets_owned", f.PetsOwned)
	eq(b, "favorite_season", f.FavoriteSeason)
	overlaps(b, "ideal_vacation_style", f.IdealVacationStyle)
	overlaps(b, "vacation_activities", f.VacationActivities)

	// --- Personalidad ---
	for _, pf := range f.PersonalityTraits {
		b.personalityExists(pf)
	}

	// --- Online now ---
	if f.OnlineNow {
		b.add("u.last_active_at >= now() - make_interval(secs => $%d::double precision)", float64(onlineNowWindowSeconds))
	}

	// Las condiciones acaban aquí: la consulta de recuento usa exactamente los
	// mismos argumentos, sin los de paginación.
	whereSQL := strings.Join(b.where, " AND ")
	filterArgs := len(b.args)

	// La popularidad viene de la vista materializada profile_popularity; solo se
	// une cuando el orden la necesita.
	joins := `JOIN users u ON u.id = p.user_id`
	orderBy := orderByClause(params.Sort)
	if f.OnlineNow {
		// En online-now prima la actividad reciente, no la creación reciente.
		orderBy = "u.last_active_at DESC, p.id ASC"
	} else if params.Sort == SortPopular {
		joins += `
		LEFT JOIN profile_popularity pop ON pop.profile_id = p.id`
	}

	pageLimit := params.PageSize
	if params.SkipTotal {
		pageLimit++
	}
	limitArg := len(b.args) + 1
	offsetArg := len(b.args) + 2
	args := append(b.args, pageLimit, (params.Page-1)*params.PageSize)

	sql := fmt.Sprintf(`
		SELECT p.id
		FROM profiles p
		%s
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, joins, whereSQL, orderBy, limitArg, offsetArg)

	countSQL := fmt.Sprintf(`
			SELECT COUNT(*)
			FROM profiles p
			JOIN users u ON u.id = p.user_id
			WHERE %s
		`, whereSQL)

	// El recuento se hace SIN el usuario que busca (uuid.Nil: no excluye bloqueos
	// ni a uno mismo), igual que hace la caché de listados. Así el mismo recuento
	// lo comparten todos los usuarios con los mismos filtros. A cambio, el total
	// puede sobrar en unas pocas unidades (uno mismo y sus bloqueos).
	countArgs := append([]any(nil), args[:filterArgs]...)
	countArgs[0] = uuid.Nil

	return builtQuery{SQL: sql, Args: args, CountSQL: countSQL, CountArgs: countArgs}
}

func (b *queryBuilder) interestExists(itf InterestFilter) {
	b.args = append(b.args, itf.Key)
	keyArg := len(b.args)

	var sb strings.Builder
	fmt.Fprintf(&sb, `EXISTS (
		SELECT 1 FROM profile_interests pi
		WHERE pi.profile_id = p.id AND pi.interest_key = $%d`, keyArg)

	if itf.Min != nil {
		b.args = append(b.args, *itf.Min)
		fmt.Fprintf(&sb, " AND pi.level >= $%d", len(b.args))
	}
	if itf.Max != nil {
		b.args = append(b.args, *itf.Max)
		fmt.Fprintf(&sb, " AND pi.level <= $%d", len(b.args))
	}
	sb.WriteString(")")

	b.where = append(b.where, sb.String())
}

func (b *queryBuilder) personalityExists(pf PersonalityFilter) {
	b.args = append(b.args, pf.TraitKey)
	keyArg := len(b.args)

	var sb strings.Builder
	fmt.Fprintf(&sb, `EXISTS (
		SELECT 1 FROM profile_personality_trait_scores pts
		WHERE pts.profile_id = p.id AND pts.trait_key = $%d`, keyArg)

	if pf.Min != nil {
		b.args = append(b.args, *pf.Min)
		fmt.Fprintf(&sb, " AND pts.avg_score >= $%d", len(b.args))
	}
	if pf.Max != nil {
		b.args = append(b.args, *pf.Max)
		fmt.Fprintf(&sb, " AND pts.avg_score <= $%d", len(b.args))
	}
	sb.WriteString(")")

	b.where = append(b.where, sb.String())
}

func orderByClause(sort Sort) string {
	switch sort {
	case SortAgeAsc:
		return "p.birth_date DESC, p.id ASC"
	case SortAgeDesc:
		return "p.birth_date ASC, p.id ASC"
	case SortPopular:
		// pop viene de un LEFT JOIN a profile_popularity: los perfiles sin
		// actividad reciente no tienen fila y cuentan como 0.
		return "COALESCE(pop.score, 0) DESC, p.created_at DESC, p.id ASC"
	default: // SortRecent
		return "p.created_at DESC, p.id ASC"
	}
}

// ageBounds convierte un rango de edad en límites sobre birth_date, coherentes
// con profiles.AgeAt (la edad que ve el usuario en las tarjetas):
//
//	minBirth: nació en esa fecha o antes  <=> ya tiene al menos minAge años
//	maxBirth: nació en esa fecha o después <=> todavía no ha cumplido maxAge+1
//
// Son fechas (medianoche UTC): birth_date es un DATE. "Hace n años" se calcula
// sin desbordar: el 29 de febrero hace n años que no existe (n no múltiplo de 4)
// es el 28, no el 1 de marzo que daría time.AddDate. Con AddDate, ese único día
// de cada 4 años, quien cumple justo al día siguiente quedaba mal clasificado.
func ageBounds(now time.Time, minAge, maxAge *int) (minBirth, maxBirth *time.Time) {
	y, m, d := now.Date()
	if minAge != nil {
		t := yearsAgo(y, m, d, *minAge)
		minBirth = &t
	}
	if maxAge != nil {
		t := yearsAgo(y, m, d, *maxAge+1).AddDate(0, 0, 1)
		maxBirth = &t
	}
	return minBirth, maxBirth
}

// yearsAgo devuelve la fecha (y-n, m, d) a medianoche UTC; si ese día no existe
// (29 de febrero en un año no bisiesto) devuelve el último día del mes.
func yearsAgo(y int, m time.Month, d, n int) time.Time {
	t := time.Date(y-n, m, d, 0, 0, 0, 0, time.UTC)
	if t.Month() != m {
		t = time.Date(y-n, m+1, 0, 0, 0, 0, 0, time.UTC)
	}
	return t
}
