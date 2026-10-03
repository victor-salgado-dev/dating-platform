package search

import (
	"context"
	"fmt"
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
	now := time.Now()
	q := buildSearchQuery(params, now)

	rows, err := r.db.Query(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, fmt.Errorf("search: consultar perfiles: %w", err)
	}
	defer rows.Close()

	var items []ResultItem
	total := 0

	for rows.Next() {
		var (
			item         ResultItem
			genderStr    string
			relGoalsList []string
			birthDate    time.Time
			totalCount   int
		)

		if err := rows.Scan(
			&item.ProfileID, &item.DisplayName, &birthDate, &genderStr, &item.CountryCode,
			&item.Region, &relGoalsList, &item.CreatedAt, &item.PhotoID,
			&item.Liked, &item.Favorited, &item.ReceivedLike, &item.ReceivedFavorite,
			&totalCount,
		); err != nil {
			return nil, fmt.Errorf("search: leer resultado: %w", err)
		}

		item.HasPhoto = item.PhotoID != nil
		item.Age = profiles.AgeAt(birthDate, now)
		item.Gender = profiles.Gender(genderStr)
		if len(relGoalsList) > 0 {
			item.RelationshipGoals = make([]profiles.RelationshipGoal, len(relGoalsList))
			for i, rg := range relGoalsList {
				item.RelationshipGoals[i] = profiles.RelationshipGoal(rg)
			}
		}

		total = totalCount
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search: leer resultados: %w", err)
	}

	// COUNT(*) OVER() vive en las filas devueltas: si la página pedida está
	// fuera de rango no hay ninguna y total saldría 0, con lo que el cliente
	// no sabría a qué página volver. En ese caso (y solo ese) se cuenta aparte.
	if len(items) == 0 && params.Page > 1 {
		if err := r.db.QueryRow(ctx, q.CountSQL, q.CountArgs...).Scan(&total); err != nil {
			return nil, fmt.Errorf("search: contar perfiles: %w", err)
		}
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
