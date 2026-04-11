package repository

import (
	"context"
	"errors"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	queries *client.Queries
}

func New(q *client.Queries) service.Repository {
	return &repository{queries: q}
}

func (r *repository) ListActualHikes(ctx context.Context, limit, offset int32) ([]service.Hike, error) {
	rawHikes, err := r.queries.ListActualHikes(ctx, client.ListActualHikesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, logger.WrapError(err)
	}

	serviceHikes := make([]service.Hike, 0, len(rawHikes))
	for _, rawHike := range rawHikes {
		var imagePath *string
		if rawHike.ImagePath.Valid {
			s := rawHike.ImagePath.String
			imagePath = &s
		}

		var distance float64
		if rawHike.DistanceKm.Valid {
			result, err := rawHike.DistanceKm.Float64Value()
			if err != nil {
				return nil, logger.WrapError(err)
			}
			distance = result.Float64
		}

		serviceHikes = append(serviceHikes, service.Hike{
			ID:             rawHike.ID,
			TitleRu:        rawHike.TitleRu,
			PreviewRu:      rawHike.PreviewRu,
			StartsAt:       rawHike.StartsAt,
			EndsAt:         rawHike.EndsAt,
			ImagePath:      imagePath,
			PriceGel:       rawHike.PriceGel,
			DistanceKm:     distance,
			ElevationGainM: int(rawHike.ElevationGainM.Int32),
		})
	}

	return serviceHikes, nil
}

func (r *repository) GetHike(ctx context.Context, id int32) (service.Hike, error) {
	hikeRaw, err := r.queries.GetHike(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.Hike{}, service.ErrHikesNotFound
		}
		return service.Hike{}, logger.WrapError(err)
	}
	hike := service.Hike{
		ID:       hikeRaw.ID,
		TitleRu:  hikeRaw.TitleRu,
		StartsAt: hikeRaw.StartsAt,
		EndsAt:   hikeRaw.EndsAt,
	}

	return hike, nil
}
