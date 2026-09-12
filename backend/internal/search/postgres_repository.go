package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"dating-platform/backend/internal/profiles"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Search(ctx context.Context, params Params) (*Result, error) {
	where := []string{
		"u.status = 'active'", "u.deleted_at IS NULL", "p.user_id <> $1",
		`NOT EXISTS (
			SELECT 1 FROM blocks b
			WHERE (b.blocker_id = $1 AND b.blocked_id = p.user_id)
			   OR (b.blocker_id = p.user_id AND b.blocked_id = $1)
		)`,
	}
	args := []any{params.ExcludeUserID}

	add := func(clause string, val any) {
		args = append(args, val)
		where = append(where, fmt.Sprintf(clause, len(args)))
	}

	f := params.Filters
	now := time.Now()

	if len(f.Genders) > 0 {
		genders := make([]string, len(f.Genders))
		for i, g := range f.Genders {
			genders[i] = string(g)
		}
		add("p.gender = ANY($%d)", genders)
	}

	// Los filtros de edad se traducen a un rango de birth_date con la
	// MISMA lógica de "cumplir años" que Profile.Age() (profiles.AgeAt),
	// para que el filtro y el campo "age" devuelto sean siempre coherentes.
	if f.MinAge != nil {
		// Nacido en esta fecha o antes => ya tiene al menos MinAge años.
		bound := now.AddDate(-*f.MinAge, 0, 0)
		add("p.birth_date <= $%d", bound)
	}
	if f.MaxAge != nil {
		// Nacido en esta fecha o después => todavía no supera MaxAge años.
		bound := now.AddDate(-(*f.MaxAge + 1), 0, 1)
		add("p.birth_date >= $%d", bound)
	}

	if f.CountryCode != nil {
		add("p.country_code = $%d", *f.CountryCode)
	}

	// languages/interests son arrays NULLABLE: el operador && con una
	// columna NULL evalúa a NULL (no TRUE), así que estas cláusulas ya
	// excluyen automáticamente los perfiles sin ese dato, cumpliendo la
	// regla de los datos faltantes sin necesidad de un IS NOT NULL aparte.
	if len(f.Languages) > 0 {
		add("p.languages && $%d", f.Languages)
	}
	if len(f.Interests) > 0 {
		add("p.interests && $%d", f.Interests)
	}

	// relationship_goal/has_children/wants_children son NULLABLE con
	// comparación de igualdad: NULL = valor también evalúa a NULL, así
	// que estas cláusulas excluyen igualmente los perfiles sin ese dato.
	if f.RelationshipGoal != nil {
		add("p.relationship_goal = $%d", string(*f.RelationshipGoal))
	}
	if f.HasChildren != nil {
		add("p.has_children = $%d", *f.HasChildren)
	}
	if f.WantsChildren != nil {
		add("p.wants_children = $%d", *f.WantsChildren)
	}

	orderBy := orderByClause(params.Sort)

	limitArg := len(args) + 1
	offsetArg := len(args) + 2
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	query := fmt.Sprintf(`
		SELECT
			p.id, p.display_name, p.birth_date, p.gender, p.country_code, p.region,
			p.relationship_goal, p.created_at,
			EXISTS (SELECT 1 FROM profile_photos ph WHERE ph.profile_id = p.id) AS has_photo,
			COUNT(*) OVER() AS total_count
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), orderBy, limitArg, offsetArg)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search: consultar perfiles: %w", err)
	}
	defer rows.Close()

	var items []ResultItem
	total := 0

	for rows.Next() {
		var (
			item       ResultItem
			genderStr  string
			relGoalStr *string
			birthDate  time.Time
			totalCount int
		)

		if err := rows.Scan(
			&item.ProfileID, &item.DisplayName, &birthDate, &genderStr, &item.CountryCode,
			&item.Region, &relGoalStr, &item.CreatedAt, &item.HasPhoto, &totalCount,
		); err != nil {
			return nil, fmt.Errorf("search: leer resultado: %w", err)
		}

		item.Age = profiles.AgeAt(birthDate, now)
		item.Gender = profiles.Gender(genderStr)
		if relGoalStr != nil {
			g := profiles.RelationshipGoal(*relGoalStr)
			item.RelationshipGoal = &g
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer resultados: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return &Result{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func orderByClause(sort Sort) string {
	switch sort {
	case SortAgeAsc:
		return "p.birth_date DESC, p.id ASC"
	case SortAgeDesc:
		return "p.birth_date ASC, p.id ASC"
	default: // SortRecent
		return "p.created_at DESC, p.id ASC"
	}
}
