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

func (r *repository) ListActualHikes(ctx context.Context, limit, offset int32) ([]service.HikeListItem, error) {
	rawHikes, err := r.queries.ListActualHikes(ctx, client.ListActualHikesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, logger.WrapError(err)
	}

	serviceHikes := make([]service.HikeListItem, 0, len(rawHikes))
	for _, rawHike := range rawHikes {
		var distance float64
		if rawHike.DistanceKm.Valid {
			result, err := rawHike.DistanceKm.Float64Value()
			if err != nil {
				return nil, logger.WrapError(err)
			}
			distance = result.Float64
		}

		serviceHikes = append(serviceHikes, service.HikeListItem{
			ID:             rawHike.ID,
			TitleRu:        rawHike.TitleRu,
			StartsAt:       rawHike.StartsAt,
			EndsAt:         rawHike.EndsAt,
			PriceGel:       rawHike.PriceGel,
			DistanceKm:     distance,
			ElevationGainM: int(rawHike.ElevationGainM.Int32),
		})
	}

	return serviceHikes, nil
}

func (r *repository) GetHikeCard(ctx context.Context, id int32) (service.HikeCard, error) {
	rawHike, err := r.queries.GetHikeCard(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.HikeCard{}, service.ErrHikesNotFound
		}
		return service.HikeCard{}, logger.WrapError(err)
	}

	var distance float64
	if rawHike.DistanceKm.Valid {
		v, err := rawHike.DistanceKm.Float64Value()
		if err != nil {
			return service.HikeCard{}, logger.WrapError(err)
		}
		distance = v.Float64
	}

	var elevationGainM int32
	if rawHike.ElevationGainM.Valid {
		elevationGainM = rawHike.ElevationGainM.Int32
	}

	var imagePath string
	if rawHike.ImagePath.Valid {
		imagePath = rawHike.ImagePath.String
	}

	hike := service.HikeCard{
		ID:             rawHike.ID,
		TitleRu:        rawHike.TitleRu,
		StartsAt:       rawHike.StartsAt,
		EndsAt:         rawHike.EndsAt,
		PriceGel:       rawHike.PriceGel,
		DistanceKm:     &distance,
		ElevationGainM: &elevationGainM,
		ImagePath:      &imagePath,
	}

	return hike, nil
}

func (r *repository) GetHikeDetails(ctx context.Context, id int32) (service.HikeDetails, error) {
	rawHike, err := r.queries.GetHikeDetails(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return service.HikeDetails{}, service.ErrHikesNotFound
		}
		return service.HikeDetails{}, logger.WrapError(err)
	}

	return service.HikeDetails{
		ID:            rawHike.ID,
		TitleRu:       rawHike.TitleRu,
		DescriptionRu: rawHike.DescriptionRu,
	}, nil
}
