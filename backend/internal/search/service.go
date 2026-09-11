package search

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// RawQuery son los parámetros de búsqueda tal como llegan de la query
// string HTTP, todavía sin validar ni convertir a tipos de dominio.
// Mantener esta capa separada permite testear la validación (Service)
// sin depender de net/http.
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

	if raw.HasChildren != "" {
		b, err := parseBool("has_children", raw.HasChildren)
		if err != nil {
			return Params{}, err
		}
		f.HasChildren = &b
	}
	if raw.WantsChildren != "" {
		b, err := parseBool("wants_children", raw.WantsChildren)
		if err != nil {
			return Params{}, err
		}
		f.WantsChildren = &b
	}

	for _, i := range raw.Interests {
		i = strings.TrimSpace(i)
		if i != "" {
			f.Interests = append(f.Interests, i)
		}
	}

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

func parseBool(field, raw string) (bool, error) {
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, invalidParam(field, "debe ser true o false")
	}
	return b, nil
}
