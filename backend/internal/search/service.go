package search

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// RawQuery son los parámetros de búsqueda tal como llegan de la query string
type RawQuery struct {
	Genders          []string
	MinAge           string
	MaxAge           string
	CountryCode      string
	Languages        []string
	RelationshipGoal string
	HasChildren      string
	WantsChildren    string
	Interests        []string
	Page             string
	PageSize         string
	Sort             string

	// --- Nuevos campos ---
	MinHeight             string
	MaxHeight             string
	MinWeight             string
	MaxWeight             string
	BodyType              string
	Ethnicity             string
	AppearanceRating      string
	HairColor             string
	EyeColor              string
	BodyArt               []string
	SmokingHabit          string
	DrinkingHabit         string
	RelocationWillingness []string
	MaritalStatus         string
	MaxChildren           string
	Occupation            string
	EmploymentStatus      string
	IncomeLevel           string
	LivingSituation       string
	Nationality           string
	EducationLevel        string
	EnglishAbility        string
	Religion              string
	ReligiousValues       string
	StarSign              string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Search(ctx context.Context, excludeUserID uuid.UUID, raw RawQuery) (*Result, error) {
	params, err := buildParams(excludeUserID, raw)
	if err != nil {
		return nil, err
	}
	return s.repo.Search(ctx, params)
}

func buildParams(excludeUserID uuid.UUID, raw RawQuery) (Params, error) {
	var f Filters

	for _, g := range raw.Genders {
		gender := profiles.Gender(strings.TrimSpace(g))
		if !profiles.IsValidGender(gender) {
			return Params{}, invalidParam("gender", "valor no permitido: "+g)
		}
		f.Genders = append(f.Genders, gender)
	}

	if raw.MinAge != "" {
		n, err := parseAge("min_age", raw.MinAge)
		if err != nil {
			return Params{}, err
		}
		f.MinAge = &n
	}
	if raw.MaxAge != "" {
		n, err := parseAge("max_age", raw.MaxAge)
		if err != nil {
			return Params{}, err
		}
		f.MaxAge = &n
	}
	if f.MinAge != nil && f.MaxAge != nil && *f.MinAge > *f.MaxAge {
		return Params{}, invalidParam("min_age", "no puede ser mayor que max_age")
	}

	if raw.CountryCode != "" {
		cc := strings.ToUpper(strings.TrimSpace(raw.CountryCode))
		if len(cc) != 2 {
			return Params{}, invalidParam("country", "debe ser un código ISO 3166-1 alpha-2 (p.ej. ES)")
		}
		f.CountryCode = &cc
	}

	for _, l := range raw.Languages {
		l = strings.TrimSpace(l)
		if l != "" {
			f.Languages = append(f.Languages, l)
		}
	}

	if raw.RelationshipGoal != "" {
		g := profiles.RelationshipGoal(strings.TrimSpace(raw.RelationshipGoal))
		if !profiles.IsValidRelationshipGoal(g) {
			return Params{}, invalidParam("relationship_goal", "valor no permitido")
		}
		f.RelationshipGoal = &g
	}

	// Strings simples opcionales
	nonEmptyPtr := func(val string) *string {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return nil
		}
		return &trimmed
	}

	f.HasChildren = nonEmptyPtr(raw.HasChildren)
	f.WantsChildren = nonEmptyPtr(raw.WantsChildren)

	for _, i := range raw.Interests {
		i = strings.TrimSpace(i)
		if i != "" {
			f.Interests = append(f.Interests, i)
		}
	}

	// --- Validación de números (Alturas, Pesos, Hijos) ---
	parseInt := func(field, raw string, min, max int) (*int, error) {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			return nil, nil
		}
		n, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, invalidParam(field, "debe ser un número entero")
		}
		if n < min || n > max {
			return nil, invalidParam(field, "valor fuera de rango permitido")
		}
		return &n, nil
	}

	var err error
	if f.MinHeight, err = parseInt("min_height", raw.MinHeight, 50, 300); err != nil { return Params{}, err }
	if f.MaxHeight, err = parseInt("max_height", raw.MaxHeight, 50, 300); err != nil { return Params{}, err }
	if f.MinHeight != nil && f.MaxHeight != nil && *f.MinHeight > *f.MaxHeight {
		return Params{}, invalidParam("min_height", "no puede ser mayor que max_height")
	}

	if f.MinWeight, err = parseInt("min_weight", raw.MinWeight, 20, 400); err != nil { return Params{}, err }
	if f.MaxWeight, err = parseInt("max_weight", raw.MaxWeight, 20, 400); err != nil { return Params{}, err }
	if f.MinWeight != nil && f.MaxWeight != nil && *f.MinWeight > *f.MaxWeight {
		return Params{}, invalidParam("min_weight", "no puede ser mayor que max_weight")
	}

	if f.MaxChildren, err = parseInt("max_children", raw.MaxChildren, 0, 30); err != nil { return Params{}, err }

	// --- Mapeo de opciones de texto simples ---
	f.BodyType = nonEmptyPtr(raw.BodyType)
	f.Ethnicity = nonEmptyPtr(raw.Ethnicity)
	f.AppearanceRating = nonEmptyPtr(raw.AppearanceRating)
	f.HairColor = nonEmptyPtr(raw.HairColor)
	f.EyeColor = nonEmptyPtr(raw.EyeColor)
	f.SmokingHabit = nonEmptyPtr(raw.SmokingHabit)
	f.DrinkingHabit = nonEmptyPtr(raw.DrinkingHabit)
	f.MaritalStatus = nonEmptyPtr(raw.MaritalStatus)
	f.Occupation = nonEmptyPtr(raw.Occupation)
	f.EmploymentStatus = nonEmptyPtr(raw.EmploymentStatus)
	f.IncomeLevel = nonEmptyPtr(raw.IncomeLevel)
	f.LivingSituation = nonEmptyPtr(raw.LivingSituation)

	if raw.Nationality != "" {
		nat := strings.ToUpper(strings.TrimSpace(raw.Nationality))
		f.Nationality = &nat
	}
	f.EducationLevel = nonEmptyPtr(raw.EducationLevel)
	f.EnglishAbility = nonEmptyPtr(raw.EnglishAbility)
	f.Religion = nonEmptyPtr(raw.Religion)
	f.ReligiousValues = nonEmptyPtr(raw.ReligiousValues)
	f.StarSign = nonEmptyPtr(raw.StarSign)

	// --- Mapeo de listas múltiples (Arrays) ---
	cleanSlice := func(items []string) []string {
		var out []string
		for _, it := range items {
			trimmed := strings.TrimSpace(it)
			if trimmed != "" {
				out = append(out, trimmed)
			}
		}
		return out
	}
	f.BodyArt = cleanSlice(raw.BodyArt)
	f.RelocationWillingness = cleanSlice(raw.RelocationWillingness)

	// Ordenación y Paginación
	sort := Sort(raw.Sort)
	switch sort {
	case "":
		sort = SortRecent
	case SortRecent, SortAgeAsc, SortAgeDesc:
		// válido
	default:
		return Params{}, invalidParam("sort", "valores permitidos: recent, age_asc, age_desc")
	}

	page := 1
	if raw.Page != "" {
		n, err := strconv.Atoi(raw.Page)
		if err != nil || n < 1 {
			return Params{}, invalidParam("page", "debe ser un entero >= 1")
		}
		page = n
	}

	pageSize := DefaultPageSize
	if raw.PageSize != "" {
		n, err := strconv.Atoi(raw.PageSize)
		if err != nil || n < 1 {
			return Params{}, invalidParam("page_size", "debe ser un entero >= 1")
		}
		if n > MaxPageSize {
			n = MaxPageSize
		}
		pageSize = n
	}

	return Params{
		Filters:       f,
		ExcludeUserID: excludeUserID,
		Sort:          sort,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

func parseAge(field, raw string) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, invalidParam(field, "debe ser un número entero")
	}
	if n < MinSearchAge || n > MaxSearchAge {
		return 0, invalidParam(field, "debe estar entre 18 y 120")
	}
	return n, nil
}