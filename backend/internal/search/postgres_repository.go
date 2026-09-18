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

	// Filtros de edad
	if f.MinAge != nil {
		bound := now.AddDate(-*f.MinAge, 0, 0)
		add("p.birth_date <= $%d", bound)
	}
	if f.MaxAge != nil {
		bound := now.AddDate(-(*f.MaxAge + 1), 0, 1)
		add("p.birth_date >= $%d", bound)
	}

	if f.CountryCode != nil {
		add("p.country_code = $%d", *f.CountryCode)
	}

	if len(f.Languages) > 0 {
		add("p.languages && $%d", f.Languages)
	}
	if len(f.Interests) > 0 {
		add("p.interests && $%d", f.Interests)
	}

	if f.RelationshipGoal != nil {
		add("p.relationship_goal = $%d", string(*f.RelationshipGoal))
	}

	// Cambiado de booleano a string
	if f.HasChildren != nil {
		add("p.has_children = $%d", *f.HasChildren)
	}
	if f.WantsChildren != nil {
		add("p.wants_children = $%d", *f.WantsChildren)
	}

	// --- NUEVOS FILTROS: Físico y Apariencia ---
	if f.MinHeight != nil {
		add("p.height >= $%d", *f.MinHeight)
	}
	if f.MaxHeight != nil {
		add("p.height <= $%d", *f.MaxHeight)
	}
	if f.MinWeight != nil {
		add("p.weight >= $%d", *f.MinWeight)
	}
	if f.MaxWeight != nil {
		add("p.weight <= $%d", *f.MaxWeight)
	}
	if f.BodyType != nil {
		add("p.body_type = $%d", *f.BodyType)
	}
	if f.Ethnicity != nil {
		add("p.ethnicity = $%d", *f.Ethnicity)
	}
	if f.AppearanceRating != nil {
		add("p.appearance_rating = $%d", *f.AppearanceRating)
	}
	if f.HairColor != nil {
		add("p.hair_color = $%d", *f.HairColor)
	}
	if f.EyeColor != nil {
		add("p.eye_color = $%d", *f.EyeColor)
	}
	if len(f.BodyArt) > 0 {
		add("p.body_art && $%d", f.BodyArt)
	}

	// --- NUEVOS FILTROS: Estilo de Vida y Familia ---
	if f.SmokingHabit != nil {
		add("p.smoking_habit = $%d", *f.SmokingHabit)
	}
	if f.DrinkingHabit != nil {
		add("p.drinking_habit = $%d", *f.DrinkingHabit)
	}
	if len(f.RelocationWillingness) > 0 {
		add("p.relocation_willingness && $%d", f.RelocationWillingness)
	}
	if f.MaritalStatus != nil {
		add("p.marital_status = $%d", *f.MaritalStatus)
	}
	if f.MaxChildren != nil {
		add("p.children_count <= $%d", *f.MaxChildren)
	}
	if f.Occupation != nil {
		add("p.occupation = $%d", *f.Occupation)
	}
	if f.EmploymentStatus != nil {
		add("p.employment_status = $%d", *f.EmploymentStatus)
	}
	if f.IncomeLevel != nil {
		add("p.income_level = $%d", *f.IncomeLevel)
	}
	if f.LivingSituation != nil {
		add("p.living_situation = $%d", *f.LivingSituation)
	}

	// --- NUEVOS FILTROS: Fondo, Cultura y Valores ---
	if f.Nationality != nil {
		add("p.nationality = $%d", *f.Nationality)
	}
	if f.EducationLevel != nil {
		add("p.education_level = $%d", *f.EducationLevel)
	}
	if f.EnglishAbility != nil {
		add("p.english_ability = $%d", *f.EnglishAbility)
	}
	if f.Religion != nil {
		add("p.religion = $%d", *f.Religion)
	}
	if f.ReligiousValues != nil {
		add("p.religious_values = $%d", *f.ReligiousValues)
	}
	if f.StarSign != nil {
		add("p.star_sign = $%d", *f.StarSign)
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