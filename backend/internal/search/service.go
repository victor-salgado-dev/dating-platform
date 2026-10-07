package search

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"dating-platform/backend/internal/profiles"
)

// RawQuery son los parámetros de búsqueda tal como llegan de la query string,
// sin validar.
type RawQuery struct {
	Genders           []string
	MinAge            string
	MaxAge            string
	CountryCode       string
	Languages         []string
	RelationshipGoals []string
	HasChildren       string
	WantsChildren     string
	Interests         []string
	Page              string
	PageSize          string
	Sort              string
	SkipTotal         bool

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

	// --- Estilo de vida adicional ---
	FutureVision       []string
	Sports             []string
	LikesPets          string
	PetsOwned          []string
	FavoriteSeason     string
	IdealVacationStyle []string
	VacationActivities []string

	// InterestBounds: límites de nivel por clave de interés (?interest_<clave>_min / _max),
	// solo con sentido para intereses con has_level=true.
	InterestBounds map[string]RawBounds

	// PersonalityTraitBounds: límites por rasgo (?trait_<rasgo>_min / _max).
	PersonalityTraitBounds map[string]RawBounds

	// OnlineNow no proviene del query string. Lo asigna directamente el
	// handler OnlineNow para activar el filtrado por last_active_at.
	OnlineNow bool

	// UseAgePrefs tampoco viene del query string: lo fija el handler
	// Recommended para aplicar el rango de edad de las preferencias de pareja.
	UseAgePrefs bool
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

// buildParams valida y normaliza la query string. Las secciones se validan en un
// orden fijo, así que ante varios errores siempre se devuelve el mismo.
func buildParams(excludeUserID uuid.UUID, raw RawQuery) (Params, error) {
	var f Filters
	var err error

	if err = fillDemographicFilters(&f, raw); err != nil {
		return Params{}, err
	}
	if err = fillBodyRangeFilters(&f, raw); err != nil {
		return Params{}, err
	}
	fillTextFilters(&f, raw)

	if f.Interests, err = buildInterestFilters(raw); err != nil {
		return Params{}, err
	}
	if f.PersonalityTraits, err = buildPersonalityFilters(raw); err != nil {
		return Params{}, err
	}

	f.OnlineNow = raw.OnlineNow

	sortValue, err := parseSort(raw.Sort)
	if err != nil {
		return Params{}, err
	}
	page, pageSize, err := parsePaging(raw.Page, raw.PageSize)
	if err != nil {
		return Params{}, err
	}

	return Params{
		Filters:       f,
		ExcludeUserID: excludeUserID,
		Sort:          sortValue,
		Page:          page,
		PageSize:      pageSize,
		SkipTotal:     raw.SkipTotal,
		UseAgePrefs:   raw.UseAgePrefs,
	}, nil
}

// fillDemographicFilters: género, edad, país, idiomas, objetivos e hijos.
func fillDemographicFilters(f *Filters, raw RawQuery) error {
	for _, g := range raw.Genders {
		gender := profiles.Gender(strings.TrimSpace(g))
		if !profiles.IsValidGender(gender) {
			return invalidParam("gender", "valor no permitido: "+g)
		}
		f.Genders = append(f.Genders, gender)
	}

	if raw.MinAge != "" {
		n, err := parseAge("min_age", raw.MinAge)
		if err != nil {
			return err
		}
		f.MinAge = &n
	}
	if raw.MaxAge != "" {
		n, err := parseAge("max_age", raw.MaxAge)
		if err != nil {
			return err
		}
		f.MaxAge = &n
	}
	if f.MinAge != nil && f.MaxAge != nil && *f.MinAge > *f.MaxAge {
		return invalidParam("min_age", "no puede ser mayor que max_age")
	}

	if raw.CountryCode != "" {
		cc := strings.ToUpper(strings.TrimSpace(raw.CountryCode))
		if len(cc) != 2 {
			return invalidParam("country", "debe ser un código ISO 3166-1 alpha-2 (p.ej. ES)")
		}
		f.CountryCode = &cc
	}

	f.Languages = cleanList(raw.Languages)

	for _, rg := range raw.RelationshipGoals {
		rg = strings.TrimSpace(rg)
		if rg == "" {
			continue
		}
		goal := profiles.RelationshipGoal(rg)
		if !profiles.IsValidRelationshipGoal(goal) {
			return invalidParam("relationship_goal", "valor no permitido: "+rg)
		}
		f.RelationshipGoals = append(f.RelationshipGoals, goal)
	}

	f.HasChildren = trimmedPtr(raw.HasChildren)
	f.WantsChildren = trimmedPtr(raw.WantsChildren)
	return nil
}

// fillBodyRangeFilters: alturas, pesos y máximo de hijos.
func fillBodyRangeFilters(f *Filters, raw RawQuery) error {
	var err error

	if f.MinHeight, err = parseInt("min_height", raw.MinHeight, 50, 300); err != nil {
		return err
	}
	if f.MaxHeight, err = parseInt("max_height", raw.MaxHeight, 50, 300); err != nil {
		return err
	}
	if f.MinHeight != nil && f.MaxHeight != nil && *f.MinHeight > *f.MaxHeight {
		return invalidParam("min_height", "no puede ser mayor que max_height")
	}

	if f.MinWeight, err = parseInt("min_weight", raw.MinWeight, 20, 400); err != nil {
		return err
	}
	if f.MaxWeight, err = parseInt("max_weight", raw.MaxWeight, 20, 400); err != nil {
		return err
	}
	if f.MinWeight != nil && f.MaxWeight != nil && *f.MinWeight > *f.MaxWeight {
		return invalidParam("min_weight", "no puede ser mayor que max_weight")
	}

	f.MaxChildren, err = parseInt("max_children", raw.MaxChildren, 0, 30)
	return err
}

// fillTextFilters: filtros de texto y de listas, que no tienen validación de
// valores (se comparan tal cual contra la base de datos).
func fillTextFilters(f *Filters, raw RawQuery) {
	f.BodyType = trimmedPtr(raw.BodyType)
	f.Ethnicity = trimmedPtr(raw.Ethnicity)
	f.AppearanceRating = trimmedPtr(raw.AppearanceRating)
	f.HairColor = trimmedPtr(raw.HairColor)
	f.EyeColor = trimmedPtr(raw.EyeColor)
	f.BodyArt = cleanList(raw.BodyArt)

	f.SmokingHabit = trimmedPtr(raw.SmokingHabit)
	f.DrinkingHabit = trimmedPtr(raw.DrinkingHabit)
	f.RelocationWillingness = cleanList(raw.RelocationWillingness)
	f.MaritalStatus = trimmedPtr(raw.MaritalStatus)
	f.Occupation = trimmedPtr(raw.Occupation)
	f.EmploymentStatus = trimmedPtr(raw.EmploymentStatus)
	f.IncomeLevel = trimmedPtr(raw.IncomeLevel)
	f.LivingSituation = trimmedPtr(raw.LivingSituation)

	f.Nationality = upperPtr(raw.Nationality)
	f.EducationLevel = trimmedPtr(raw.EducationLevel)
	f.EnglishAbility = trimmedPtr(raw.EnglishAbility)
	f.Religion = trimmedPtr(raw.Religion)
	f.ReligiousValues = trimmedPtr(raw.ReligiousValues)
	f.StarSign = trimmedPtr(raw.StarSign)

	f.FutureVision = cleanList(raw.FutureVision)
	f.Sports = cleanList(raw.Sports)
	f.LikesPets = trimmedPtr(raw.LikesPets)
	f.PetsOwned = cleanList(raw.PetsOwned)
	f.FavoriteSeason = trimmedPtr(raw.FavoriteSeason)
	f.IdealVacationStyle = cleanList(raw.IdealVacationStyle)
	f.VacationActivities = cleanList(raw.VacationActivities)
}

// buildInterestFilters une los intereses sueltos (raw.Interests: "lo tiene
// marcado, sin importar el nivel") con los que traen límites de nivel
// (raw.InterestBounds). Si una clave aparece en ambas fuentes, se funden en un
// único filtro. El resultado sale ordenado por clave, para que la consulta sea
// determinista.
func buildInterestFilters(raw RawQuery) ([]InterestFilter, error) {
	filters := map[string]InterestFilter{}
	for _, key := range raw.Interests {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := filters[key]; !exists {
			filters[key] = InterestFilter{Key: key}
		}
	}
	for key := range raw.InterestBounds {
		if key = strings.TrimSpace(key); key != "" {
			filters[key] = InterestFilter{Key: key}
		}
	}

	// Cada filtro es un EXISTS en la consulta: sin tope, una URL con cientos de
	// ?interest_x_min= generaría cientos de subconsultas.
	if len(filters) > MaxInterestFilters {
		return nil, invalidParam("interests", fmt.Sprintf("máximo %d intereses por búsqueda", MaxInterestFilters))
	}

	// Se validan en orden alfabético (un map se recorre al azar): ante varios
	// límites inválidos siempre se informa del mismo.
	boundKeys := make([]string, 0, len(raw.InterestBounds))
	for rawKey := range raw.InterestBounds {
		boundKeys = append(boundKeys, rawKey)
	}
	sort.Strings(boundKeys)

	for _, rawKey := range boundKeys {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}
		bounds := raw.InterestBounds[rawKey]
		itf := filters[key]

		var err error
		if itf.Min, err = parseInt(fmt.Sprintf("interest_%s_min", key), bounds.Min, MinInterestLevel, MaxInterestLevel); err != nil {
			return nil, err
		}
		if itf.Max, err = parseInt(fmt.Sprintf("interest_%s_max", key), bounds.Max, MinInterestLevel, MaxInterestLevel); err != nil {
			return nil, err
		}
		if itf.Min != nil && itf.Max != nil && *itf.Min > *itf.Max {
			return nil, invalidParam(fmt.Sprintf("interest_%s_min", key), "no puede ser mayor que el máximo")
		}
		filters[key] = itf
	}

	keys := make([]string, 0, len(filters))
	for key := range filters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var out []InterestFilter
	for _, key := range keys {
		out = append(out, filters[key])
	}
	return out, nil
}

// buildPersonalityFilters valida los rasgos de personalidad y sus límites, en
// orden alfabético de rasgo.
func buildPersonalityFilters(raw RawQuery) ([]PersonalityFilter, error) {
	keys := make([]string, 0, len(raw.PersonalityTraitBounds))
	for key := range raw.PersonalityTraitBounds {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var out []PersonalityFilter
	for _, rawKey := range keys {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}
		if !profiles.IsValidPersonalityTrait(profiles.PersonalityTrait(key)) {
			return nil, invalidParam("trait_"+key, "rasgo de personalidad no reconocido")
		}

		bounds := raw.PersonalityTraitBounds[rawKey]
		pf := PersonalityFilter{TraitKey: key}

		var err error
		if pf.Min, err = parseFloat(fmt.Sprintf("trait_%s_min", key), bounds.Min, float64(profiles.MinPersonalityScore), float64(profiles.MaxPersonalityScore)); err != nil {
			return nil, err
		}
		if pf.Max, err = parseFloat(fmt.Sprintf("trait_%s_max", key), bounds.Max, float64(profiles.MinPersonalityScore), float64(profiles.MaxPersonalityScore)); err != nil {
			return nil, err
		}
		if pf.Min != nil && pf.Max != nil && *pf.Min > *pf.Max {
			return nil, invalidParam(fmt.Sprintf("trait_%s_min", key), "no puede ser mayor que el máximo")
		}

		out = append(out, pf)
	}
	return out, nil
}

func parseSort(value string) (Sort, error) {
	switch s := Sort(value); s {
	case "":
		return SortRecent, nil
	case SortRecent, SortAgeAsc, SortAgeDesc, SortPopular:
		return s, nil
	default:
		return "", invalidParam("sort", "valores permitidos: recent, age_asc, age_desc, popular")
	}
}

// parsePaging valida page y page_size. Un page_size mayor que el máximo se
// recorta; un page absurdo se rechaza (el OFFSET crece con él y un entero enorme
// llegaba a desbordar el cálculo).
func parsePaging(rawPage, rawPageSize string) (page, pageSize int, err error) {
	page = 1
	if rawPage != "" {
		n, convErr := strconv.Atoi(rawPage)
		if convErr != nil || n < 1 {
			return 0, 0, invalidParam("page", "debe ser un entero >= 1")
		}
		if n > MaxPage {
			return 0, 0, invalidParam("page", fmt.Sprintf("no puede ser mayor que %d", MaxPage))
		}
		page = n
	}

	pageSize = DefaultPageSize
	if rawPageSize != "" {
		n, convErr := strconv.Atoi(rawPageSize)
		if convErr != nil || n < 1 {
			return 0, 0, invalidParam("page_size", "debe ser un entero >= 1")
		}
		if n > MaxPageSize {
			n = MaxPageSize
		}
		pageSize = n
	}
	return page, pageSize, nil
}

func parseAge(field, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, invalidParam(field, "debe ser un número entero")
	}
	if n < MinSearchAge || n > MaxSearchAge {
		return 0, invalidParam(field, "debe estar entre 18 y 120")
	}
	return n, nil
}

// parseInt convierte un entero opcional dentro de [min, max]; vacío => nil.
func parseInt(field, value string, min, max int) (*int, error) {
	trimmed := strings.TrimSpace(value)
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

// parseFloat convierte un decimal opcional dentro de [min, max]; vacío => nil.
func parseFloat(field, value string, min, max float64) (*float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil, invalidParam(field, "debe ser un número")
	}
	if n < min || n > max {
		return nil, invalidParam(field, "valor fuera de rango permitido")
	}
	return &n, nil
}

// trimmedPtr devuelve el texto sin espacios, o nil si queda vacío.
func trimmedPtr(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// upperPtr es trimmedPtr en mayúsculas (p. ej. nacionalidad).
func upperPtr(value string) *string {
	upper := strings.ToUpper(strings.TrimSpace(value))
	if upper == "" {
		return nil
	}
	return &upper
}

// cleanList descarta los elementos vacíos tras recortar espacios; nil si no queda ninguno.
func cleanList(items []string) []string {
	var out []string
	for _, it := range items {
		if trimmed := strings.TrimSpace(it); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
