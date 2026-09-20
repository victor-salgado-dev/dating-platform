package profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
	pgForeignKeyViolation = "23503"
)

// Listado de todas las columnas (excepto id, created_at, updated_at para inserts)
const allProfileCols = `
	user_id, display_name, birth_date, gender, country_code,
	region, languages, relationship_goal, has_children, wants_children, bio, interests,
	height, weight, body_type, ethnicity, appearance_rating, hair_color, eye_color, body_art,
	smoking_habit, drinking_habit, relocation_willingness, marital_status,
	children_count, youngest_child_age, oldest_child_age, occupation, employment_status,
	income_level, living_situation, nationality, education_level, english_ability,
	religion, religious_values, star_sign,
	future_vision, sports, likes_pets, pets_owned, favorite_season,
	ideal_vacation_style, vacation_activities, profile_quote, dream_wish
`

// profileColCount tiene que coincidir exactamente con el número de
// columnas listadas en allProfileCols. Se usa solo para generar los
// placeholders ($1, $2...) del INSERT sin tener que contarlos ni
// renumerarlos a mano cada vez que se añade un campo nuevo.
const profileColCount = 46

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(ctx context.Context, p *Profile) error {
	placeholders := make([]string, profileColCount)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		INSERT INTO profiles (%s)
		VALUES (%s)
		RETURNING id, created_at, updated_at
	`, allProfileCols, strings.Join(placeholders, ", "))

	err := r.db.QueryRow(ctx, query,
		p.UserID, p.DisplayName, p.BirthDate, string(p.Gender), p.CountryCode,
		p.Region, p.Languages, relationshipGoalToDB(p.RelationshipGoal),
		p.HasChildren, p.WantsChildren, p.Bio, p.Interests,
		p.Height, p.Weight, p.BodyType, p.Ethnicity, p.AppearanceRating, p.HairColor, p.EyeColor, p.BodyArt,
		p.SmokingHabit, p.DrinkingHabit, p.RelocationWillingness, p.MaritalStatus,
		p.ChildrenCount, p.YoungestChildAge, p.OldestChildAge, p.Occupation, p.EmploymentStatus,
		p.IncomeLevel, p.LivingSituation, p.Nationality, p.EducationLevel, p.EnglishAbility,
		p.Religion, p.ReligiousValues, p.StarSign,
		p.FutureVision, p.Sports, p.LikesPets, p.PetsOwned, p.FavoriteSeason,
		p.IdealVacationStyle, p.VacationActivities, p.ProfileQuote, p.DreamWish,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgUniqueViolation:
				return ErrAlreadyExists
			case pgCheckViolation:
				return invalidField(pgErr.ConstraintName, "no cumple las restricciones del perfil")
			}
		}
		return fmt.Errorf("profiles: crear perfil: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	query := fmt.Sprintf(`
		SELECT id, %s, created_at, updated_at
		FROM profiles
		WHERE user_id = $1
	`, allProfileCols)

	return r.scanOne(ctx, query, userID)
}

func (r *PostgresRepository) GetPublicByID(ctx context.Context, id, viewerUserID uuid.UUID) (*Profile, error) {
	// Para GetPublicByID usamos un alias "p." en todas las columnas para que funcione con el JOIN
	aliasedCols := strings.ReplaceAll(allProfileCols, " ", " p.")
	aliasedCols = strings.ReplaceAll(aliasedCols, "\t", "")
	aliasedCols = strings.ReplaceAll(aliasedCols, "\n", "")
	aliasedCols = strings.ReplaceAll(aliasedCols, ",", ", p.")

	query := fmt.Sprintf(`
		SELECT p.id, p.%s, p.created_at, p.updated_at
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.id = $1
		  AND u.status = 'active' AND u.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM blocks b
		      WHERE (b.blocker_id = $2 AND b.blocked_id = p.user_id)
		         OR (b.blocker_id = p.user_id AND b.blocked_id = $2)
		  )
	`, aliasedCols)

	return r.scanOne(ctx, query, id, viewerUserID)
}

func (r *PostgresRepository) GetByIDAny(ctx context.Context, id uuid.UUID) (*Profile, error) {
	query := fmt.Sprintf(`
		SELECT id, %s, created_at, updated_at
		FROM profiles
		WHERE id = $1
	`, allProfileCols)

	return r.scanOne(ctx, query, id)
}

func (r *PostgresRepository) Update(ctx context.Context, userID uuid.UUID, patch ProfilePatch) (*Profile, error) {
	var setClauses []string
	var args []any
	argN := 1

	add := func(col string, val any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if patch.DisplayName != nil { add("display_name", *patch.DisplayName) }
	if patch.BirthDate != nil { add("birth_date", *patch.BirthDate) }
	if patch.Gender != nil { add("gender", string(*patch.Gender)) }
	if patch.CountryCode != nil { add("country_code", *patch.CountryCode) }
	if patch.RegionSet { add("region", patch.Region) }
	if patch.LanguagesSet { add("languages", patch.Languages) }
	if patch.RelationshipGoalSet { add("relationship_goal", relationshipGoalToDB(patch.RelationshipGoal)) }
	if patch.HasChildrenSet { add("has_children", patch.HasChildren) }
	if patch.WantsChildrenSet { add("wants_children", patch.WantsChildren) }
	if patch.BioSet { add("bio", patch.Bio) }
	if patch.InterestsSet { add("interests", patch.Interests) }

	// Nuevos campos (000012)
	if patch.HeightSet { add("height", patch.Height) }
	if patch.WeightSet { add("weight", patch.Weight) }
	if patch.BodyTypeSet { add("body_type", patch.BodyType) }
	if patch.EthnicitySet { add("ethnicity", patch.Ethnicity) }
	if patch.AppearanceRatingSet { add("appearance_rating", patch.AppearanceRating) }
	if patch.HairColorSet { add("hair_color", patch.HairColor) }
	if patch.EyeColorSet { add("eye_color", patch.EyeColor) }
	if patch.BodyArtSet { add("body_art", patch.BodyArt) }
	if patch.SmokingHabitSet { add("smoking_habit", patch.SmokingHabit) }
	if patch.DrinkingHabitSet { add("drinking_habit", patch.DrinkingHabit) }
	if patch.RelocationWillingnessSet { add("relocation_willingness", patch.RelocationWillingness) }
	if patch.MaritalStatusSet { add("marital_status", patch.MaritalStatus) }
	if patch.ChildrenCountSet { add("children_count", patch.ChildrenCount) }
	if patch.YoungestChildAgeSet { add("youngest_child_age", patch.YoungestChildAge) }
	if patch.OldestChildAgeSet { add("oldest_child_age", patch.OldestChildAge) }
	if patch.OccupationSet { add("occupation", patch.Occupation) }
	if patch.EmploymentStatusSet { add("employment_status", patch.EmploymentStatus) }
	if patch.IncomeLevelSet { add("income_level", patch.IncomeLevel) }
	if patch.LivingSituationSet { add("living_situation", patch.LivingSituation) }
	if patch.NationalitySet { add("nationality", patch.Nationality) }
	if patch.EducationLevelSet { add("education_level", patch.EducationLevel) }
	if patch.EnglishAbilitySet { add("english_ability", patch.EnglishAbility) }
	if patch.ReligionSet { add("religion", patch.Religion) }
	if patch.ReligiousValuesSet { add("religious_values", patch.ReligiousValues) }
	if patch.StarSignSet { add("star_sign", patch.StarSign) }

	// Nuevos campos (000013 - Über mich)
	if patch.FutureVisionSet { add("future_vision", patch.FutureVision) }
	if patch.SportsSet { add("sports", patch.Sports) }
	if patch.LikesPetsSet { add("likes_pets", patch.LikesPets) }
	if patch.PetsOwnedSet { add("pets_owned", patch.PetsOwned) }
	if patch.FavoriteSeasonSet { add("favorite_season", patch.FavoriteSeason) }
	if patch.IdealVacationStyleSet { add("ideal_vacation_style", patch.IdealVacationStyle) }
	if patch.VacationActivitiesSet { add("vacation_activities", patch.VacationActivities) }
	if patch.ProfileQuoteSet { add("profile_quote", patch.ProfileQuote) }
	if patch.DreamWishSet { add("dream_wish", patch.DreamWish) }

	if len(setClauses) == 0 {
		return r.GetByUserID(ctx, userID)
	}

	query := fmt.Sprintf(
		"UPDATE profiles SET %s WHERE user_id = $%d",
		strings.Join(setClauses, ", "), argN,
	)
	args = append(args, userID)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCheckViolation {
			return nil, invalidField(pgErr.ConstraintName, "no cumple las restricciones del perfil")
		}
		return nil, fmt.Errorf("profiles: actualizar perfil: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return r.GetByUserID(ctx, userID)
}

func (r *PostgresRepository) scanOne(ctx context.Context, query string, args ...any) (*Profile, error) {
	var p Profile
	var genderStr string
	var relGoalStr *string

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&p.ID, &p.UserID, &p.DisplayName, &p.BirthDate, &genderStr, &p.CountryCode,
		&p.Region, &p.Languages, &relGoalStr, &p.HasChildren, &p.WantsChildren, &p.Bio, &p.Interests,
		&p.Height, &p.Weight, &p.BodyType, &p.Ethnicity, &p.AppearanceRating, &p.HairColor, &p.EyeColor, &p.BodyArt,
		&p.SmokingHabit, &p.DrinkingHabit, &p.RelocationWillingness, &p.MaritalStatus,
		&p.ChildrenCount, &p.YoungestChildAge, &p.OldestChildAge, &p.Occupation, &p.EmploymentStatus,
		&p.IncomeLevel, &p.LivingSituation, &p.Nationality, &p.EducationLevel, &p.EnglishAbility,
		&p.Religion, &p.ReligiousValues, &p.StarSign,
		&p.FutureVision, &p.Sports, &p.LikesPets, &p.PetsOwned, &p.FavoriteSeason,
		&p.IdealVacationStyle, &p.VacationActivities, &p.ProfileQuote, &p.DreamWish,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("profiles: consultar perfil: %w", err)
	}

	p.Gender = Gender(genderStr)
	if relGoalStr != nil {
		g := RelationshipGoal(*relGoalStr)
		p.RelationshipGoal = &g
	}

	return &p, nil
}

func relationshipGoalToDB(g *RelationshipGoal) *string {
	if g == nil {
		return nil
	}
	v := string(*g)
	return &v
}

// --- Fotos ------------------------------------------------------------
// (El código de las fotos se queda idéntico porque no lo hemos tocado)

func (r *PostgresRepository) AddPhoto(ctx context.Context, profileID uuid.UUID, photo *Photo) error {
	const query = `
		INSERT INTO profile_photos (profile_id, storage_key, content_type, position)
		VALUES (
			$1, $2, $3,
			COALESCE((SELECT MAX(position) + 1 FROM profile_photos WHERE profile_id = $1), 0)
		)
		RETURNING id, position, created_at
	`

	err := r.db.QueryRow(ctx, query, profileID, photo.StorageKey, photo.ContentType).
		Scan(&photo.ID, &photo.Position, &photo.CreatedAt)
	if err != nil {
		return fmt.Errorf("profiles: guardar foto: %w", err)
	}

	photo.ProfileID = profileID
	return nil
}

func (r *PostgresRepository) ListPhotos(ctx context.Context, profileID uuid.UUID) ([]Photo, error) {
	const query = `
		SELECT id, profile_id, storage_key, content_type, position, created_at
		FROM profile_photos
		WHERE profile_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar fotos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var ph Photo
		if err := rows.Scan(&ph.ID, &ph.ProfileID, &ph.StorageKey, &ph.ContentType, &ph.Position, &ph.CreatedAt); err != nil {
			return nil, fmt.Errorf("profiles: leer foto: %w", err)
		}
		photos = append(photos, ph)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("profiles: listar fotos: %w", err)
	}

	return photos, nil
}

func (r *PostgresRepository) CountPhotos(ctx context.Context, profileID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM profile_photos WHERE profile_id = $1`

	var count int
	if err := r.db.QueryRow(ctx, query, profileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("profiles: contar fotos: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) GetPhoto(ctx context.Context, profileID, photoID uuid.UUID) (*Photo, error) {
	const query = `
		SELECT id, profile_id, storage_key, content_type, position, created_at
		FROM profile_photos
		WHERE id = $1 AND profile_id = $2
	`

	var ph Photo
	err := r.db.QueryRow(ctx, query, photoID, profileID).
		Scan(&ph.ID, &ph.ProfileID, &ph.StorageKey, &ph.ContentType, &ph.Position, &ph.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, fmt.Errorf("profiles: consultar foto: %w", err)
	}

	return &ph, nil
}

func (r *PostgresRepository) DeletePhoto(ctx context.Context, profileID, photoID uuid.UUID) error {
	const query = `DELETE FROM profile_photos WHERE id = $1 AND profile_id = $2`

	tag, err := r.db.Exec(ctx, query, photoID, profileID)
	if err != nil {
		return fmt.Errorf("profiles: borrar foto: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPhotoNotFound
	}

	return nil
}

// --- Catálogos ----------------------------------------------------------

func (r *PostgresRepository) ListHobbyDefinitions(ctx context.Context) ([]HobbyDefinition, error) {
	const query = `
		SELECT key, category, label, sort_order
		FROM hobby_definitions
		ORDER BY category, sort_order
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar catálogo de hobbies: %w", err)
	}
	defer rows.Close()

	var defs []HobbyDefinition
	for rows.Next() {
		var d HobbyDefinition
		if err := rows.Scan(&d.Key, &d.Category, &d.Label, &d.SortOrder); err != nil {
			return nil, fmt.Errorf("profiles: leer hobby del catálogo: %w", err)
		}
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (r *PostgresRepository) ListPersonalityStatements(ctx context.Context) ([]PersonalityStatement, error) {
	const query = `
		SELECT key, trait_key, label, sort_order
		FROM personality_statements
		ORDER BY trait_key, sort_order
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar catálogo de personalidad: %w", err)
	}
	defer rows.Close()

	var stmts []PersonalityStatement
	for rows.Next() {
		var s PersonalityStatement
		var trait string
		if err := rows.Scan(&s.Key, &trait, &s.Label, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("profiles: leer afirmación del catálogo: %w", err)
		}
		s.TraitKey = PersonalityTrait(trait)
		stmts = append(stmts, s)
	}
	return stmts, rows.Err()
}

// --- Hobbies del perfil ---------------------------------------------------

func (r *PostgresRepository) UpsertProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string, liked bool, intensity *int) (*ProfileHobby, error) {
	const query = `
		INSERT INTO profile_hobbies (profile_id, hobby_key, liked, intensity)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (profile_id, hobby_key)
		DO UPDATE SET liked = EXCLUDED.liked, intensity = EXCLUDED.intensity, updated_at = now()
		RETURNING profile_id, hobby_key, liked, intensity, updated_at
	`

	var ph ProfileHobby
	err := r.db.QueryRow(ctx, query, profileID, hobbyKey, liked, intensity).
		Scan(&ph.ProfileID, &ph.HobbyKey, &ph.Liked, &ph.Intensity, &ph.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgForeignKeyViolation:
				return nil, invalidField("hobby_key", "no existe ese hobby en el catálogo")
			case pgCheckViolation:
				return nil, invalidField("intensity", "debe estar entre 1 y 5, y solo si liked=true")
			}
		}
		return nil, fmt.Errorf("profiles: guardar hobby: %w", err)
	}

	return &ph, nil
}

func (r *PostgresRepository) ListProfileHobbies(ctx context.Context, profileID uuid.UUID) ([]ProfileHobby, error) {
	const query = `
		SELECT profile_id, hobby_key, liked, intensity, updated_at
		FROM profile_hobbies
		WHERE profile_id = $1
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar hobbies del perfil: %w", err)
	}
	defer rows.Close()

	var hobbies []ProfileHobby
	for rows.Next() {
		var ph ProfileHobby
		if err := rows.Scan(&ph.ProfileID, &ph.HobbyKey, &ph.Liked, &ph.Intensity, &ph.UpdatedAt); err != nil {
			return nil, fmt.Errorf("profiles: leer hobby del perfil: %w", err)
		}
		hobbies = append(hobbies, ph)
	}
	return hobbies, rows.Err()
}

func (r *PostgresRepository) DeleteProfileHobby(ctx context.Context, profileID uuid.UUID, hobbyKey string) error {
	const query = `DELETE FROM profile_hobbies WHERE profile_id = $1 AND hobby_key = $2`

	if _, err := r.db.Exec(ctx, query, profileID, hobbyKey); err != nil {
		return fmt.Errorf("profiles: borrar hobby: %w", err)
	}
	return nil
}

// --- Personalidad del perfil -----------------------------------------------

func (r *PostgresRepository) UpsertPersonalityAnswer(ctx context.Context, profileID uuid.UUID, statementKey string, score int) (*ProfilePersonalityAnswer, error) {
	const query = `
		INSERT INTO profile_personality_answers (profile_id, statement_key, score)
		VALUES ($1, $2, $3)
		ON CONFLICT (profile_id, statement_key)
		DO UPDATE SET score = EXCLUDED.score, updated_at = now()
		RETURNING profile_id, statement_key, score, updated_at
	`

	var a ProfilePersonalityAnswer
	err := r.db.QueryRow(ctx, query, profileID, statementKey, score).
		Scan(&a.ProfileID, &a.StatementKey, &a.Score, &a.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgForeignKeyViolation:
				return nil, invalidField("statement_key", "no existe esa afirmación en el catálogo")
			case pgCheckViolation:
				return nil, invalidField("score", "debe estar entre 1 y 5")
			}
		}
		return nil, fmt.Errorf("profiles: guardar respuesta de personalidad: %w", err)
	}

	return &a, nil
}

func (r *PostgresRepository) ListPersonalityAnswers(ctx context.Context, profileID uuid.UUID) ([]ProfilePersonalityAnswer, error) {
	const query = `
		SELECT profile_id, statement_key, score, updated_at
		FROM profile_personality_answers
		WHERE profile_id = $1
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("profiles: listar respuestas de personalidad: %w", err)
	}
	defer rows.Close()

	var answers []ProfilePersonalityAnswer
	for rows.Next() {
		var a ProfilePersonalityAnswer
		if err := rows.Scan(&a.ProfileID, &a.StatementKey, &a.Score, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("profiles: leer respuesta de personalidad: %w", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

func (r *PostgresRepository) GetPersonalityTraitScores(ctx context.Context, profileID uuid.UUID) ([]PersonalityTraitScore, error) {
	const query = `
		SELECT profile_id, trait_key, avg_score, answered_count
		FROM profile_personality_trait_scores
		WHERE profile_id = $1
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("profiles: calcular puntuaciones de personalidad: %w", err)
	}
	defer rows.Close()

	var scores []PersonalityTraitScore
	for rows.Next() {
		var s PersonalityTraitScore
		var trait string
		if err := rows.Scan(&s.ProfileID, &trait, &s.AverageScore, &s.AnsweredCount); err != nil {
			return nil, fmt.Errorf("profiles: leer puntuación de personalidad: %w", err)
		}
		s.TraitKey = PersonalityTrait(trait)
		scores = append(scores, s)
	}
	return scores, rows.Err()
}

// --- Preferencias de pareja ---------------------------------------------

func (r *PostgresRepository) GetPartnerPreferences(ctx context.Context, profileID uuid.UUID) (*PartnerPreferences, error) {
	const query = `
		SELECT profile_id, age_min, age_max, height_min, height_max, desired_traits,
			partner_may_have_children, partner_religion_preference, about_partner_text,
			first_meeting_preference, desired_living_place,
			importance_shared_thoughts, importance_shared_hobbies, importance_intimacy,
			importance_romantic_love, importance_financial_security, importance_fun,
			importance_shared_friends, importance_shared_humor, importance_personal_space,
			importance_independence, created_at, updated_at
		FROM profile_partner_preferences
		WHERE profile_id = $1
	`

	var pp PartnerPreferences
	err := r.db.QueryRow(ctx, query, profileID).Scan(
		&pp.ProfileID, &pp.AgeMin, &pp.AgeMax, &pp.HeightMin, &pp.HeightMax, &pp.DesiredTraits,
		&pp.PartnerMayHaveChildren, &pp.PartnerReligionPreference, &pp.AboutPartnerText,
		&pp.FirstMeetingPreference, &pp.DesiredLivingPlace,
		&pp.ImportanceSharedThoughts, &pp.ImportanceSharedHobbies, &pp.ImportanceIntimacy,
		&pp.ImportanceRomanticLove, &pp.ImportanceFinancialSecurity, &pp.ImportanceFun,
		&pp.ImportanceSharedFriends, &pp.ImportanceSharedHumor, &pp.ImportancePersonalSpace,
		&pp.ImportanceIndependence, &pp.CreatedAt, &pp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Todavía no las ha rellenado: no es un error, es el mismo
			// "no contestado" que cualquier campo opcional en nil.
			return &PartnerPreferences{ProfileID: profileID}, nil
		}
		return nil, fmt.Errorf("profiles: consultar preferencias de pareja: %w", err)
	}

	return &pp, nil
}

func (r *PostgresRepository) UpsertPartnerPreferences(ctx context.Context, profileID uuid.UUID, patch PartnerPreferencesPatch) (*PartnerPreferences, error) {
	cols := []string{"profile_id"}
	placeholders := []string{"$1"}
	var setClauses []string
	args := []any{profileID}
	argN := 2

	addCol := func(col string, val any) {
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", argN))
		setClauses = append(setClauses, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		args = append(args, val)
		argN++
	}

	if patch.AgeMinSet { addCol("age_min", patch.AgeMin) }
	if patch.AgeMaxSet { addCol("age_max", patch.AgeMax) }
	if patch.HeightMinSet { addCol("height_min", patch.HeightMin) }
	if patch.HeightMaxSet { addCol("height_max", patch.HeightMax) }
	if patch.DesiredTraitsSet { addCol("desired_traits", patch.DesiredTraits) }
	if patch.PartnerMayHaveChildrenSet { addCol("partner_may_have_children", patch.PartnerMayHaveChildren) }
	if patch.PartnerReligionPreferenceSet { addCol("partner_religion_preference", patch.PartnerReligionPreference) }
	if patch.AboutPartnerTextSet { addCol("about_partner_text", patch.AboutPartnerText) }
	if patch.FirstMeetingPreferenceSet { addCol("first_meeting_preference", patch.FirstMeetingPreference) }
	if patch.DesiredLivingPlaceSet { addCol("desired_living_place", patch.DesiredLivingPlace) }
	if patch.ImportanceSharedThoughtsSet { addCol("importance_shared_thoughts", patch.ImportanceSharedThoughts) }
	if patch.ImportanceSharedHobbiesSet { addCol("importance_shared_hobbies", patch.ImportanceSharedHobbies) }
	if patch.ImportanceIntimacySet { addCol("importance_intimacy", patch.ImportanceIntimacy) }
	if patch.ImportanceRomanticLoveSet { addCol("importance_romantic_love", patch.ImportanceRomanticLove) }
	if patch.ImportanceFinancialSecuritySet { addCol("importance_financial_security", patch.ImportanceFinancialSecurity) }
	if patch.ImportanceFunSet { addCol("importance_fun", patch.ImportanceFun) }
	if patch.ImportanceSharedFriendsSet { addCol("importance_shared_friends", patch.ImportanceSharedFriends) }
	if patch.ImportanceSharedHumorSet { addCol("importance_shared_humor", patch.ImportanceSharedHumor) }
	if patch.ImportancePersonalSpaceSet { addCol("importance_personal_space", patch.ImportancePersonalSpace) }
	if patch.ImportanceIndependenceSet { addCol("importance_independence", patch.ImportanceIndependence) }

	// Si no hay nada que cambiar, nos limitamos a garantizar que exista
	// la fila (para que GetPartnerPreferences pueda seguir leyendo de
	// una sola tabla sin casos especiales) sin tocar ningún valor.
	setSQL := "updated_at = profile_partner_preferences.updated_at"
	if len(setClauses) > 0 {
		setSQL = strings.Join(setClauses, ", ")
	}

	query := fmt.Sprintf(`
		INSERT INTO profile_partner_preferences (%s)
		VALUES (%s)
		ON CONFLICT (profile_id) DO UPDATE SET %s
	`, strings.Join(cols, ", "), strings.Join(placeholders, ", "), setSQL)

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCheckViolation {
			return nil, invalidField(pgErr.ConstraintName, "no cumple las restricciones de preferencias de pareja")
		}
		return nil, fmt.Errorf("profiles: guardar preferencias de pareja: %w", err)
	}

	return r.GetPartnerPreferences(ctx, profileID)
}
