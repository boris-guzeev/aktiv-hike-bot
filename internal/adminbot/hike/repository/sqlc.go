package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	queries *admin.Queries
}

func New(q *admin.Queries) service.Repository {
	return &repository{queries: q}
}

func (r repository) GetHike(ctx context.Context, id int32) (service.Hike, error) {
	rawHike, err := r.queries.GetHikeByID(ctx, id)
	if err != nil {
		return service.Hike{}, logger.WrapError(err)
	}
	return service.Hike{
		ID:            rawHike.ID,
		TitleRu:       rawHike.TitleRu,
		PreviewRu:     rawHike.PreviewRu,
		DescriptionRu: rawHike.DescriptionRu,
		StartsAt:      rawHike.StartsAt,
		EndsAt:        rawHike.EndsAt,
		IsPublished:   rawHike.IsPublished,
	}, nil
}

func (r repository) ListHikes(ctx context.Context, limit, offset int32) ([]service.Hike, error) {
	rawHikes, err := r.queries.ListHikes(ctx, admin.ListHikesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, logger.WrapError(err)
	}

	var hikes []service.Hike
	for _, rawHike := range rawHikes {
		h := service.Hike{
			ID:          rawHike.ID,
			TitleRu:     rawHike.TitleRu,
			StartsAt:    rawHike.StartsAt,
			EndsAt:      rawHike.EndsAt,
			IsPublished: rawHike.IsPublished,
		}
		hikes = append(hikes, h)
	}

	return hikes, nil
}

func (r repository) ListActualHikes(ctx context.Context, limit, offset int32) ([]service.Hike, error) {
	rawActHikes, err := r.queries.ListActualHikes(ctx, admin.ListActualHikesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, logger.WrapError(err)
	}

	var hikes []service.Hike
	for _, rawHike := range rawActHikes {
		h := service.Hike{
			ID:          rawHike.ID,
			TitleRu:     rawHike.TitleRu,
			StartsAt:    rawHike.StartsAt,
			EndsAt:      rawHike.EndsAt,
			IsPublished: rawHike.IsPublished,
		}
		hikes = append(hikes, h)
	}

	return hikes, nil
}

func (r repository) PublishHike(ctx context.Context, id int32) error {
	return r.queries.SetPublished(ctx, admin.SetPublishedParams{
		ID:          id,
		IsPublished: true,
	})
}

func (r repository) HideHike(ctx context.Context, id int32) error {
	return r.queries.SetPublished(ctx, admin.SetPublishedParams{
		ID:          id,
		IsPublished: false,
	})
}

func (r repository) DeleteHike(ctx context.Context, id int32) error {
	return r.queries.DeleteHike(ctx, id)
}

func (r repository) CreateHike(ctx context.Context, hike service.Hike) (int32, error) {
	distanceKm := pgtype.Numeric{}
	if hike.DistanceKm != 0 {
		if err := distanceKm.Scan(fmt.Sprintf("%.2f", hike.DistanceKm)); err != nil {
			return 0, logger.WrapError(err)
		}
	}

	elevationGainM := pgtype.Int4{
		Int32: int32(hike.ElevationGainM),
		Valid: hike.ElevationGainM != 0,
	}

	return r.queries.CreateHike(ctx, admin.CreateHikeParams{
		TitleRu:        hike.TitleRu,
		PreviewRu:      hike.PreviewRu,
		DescriptionRu:  hike.DescriptionRu,
		StartsAt:       hike.StartsAt,
		EndsAt:         hike.EndsAt,
		PriceGel:       hike.PriceGel,
		DistanceKm:     distanceKm,
		ElevationGainM: elevationGainM,
	})
}

func (r repository) UpdateTitleRu(ctx context.Context, hikeID int32, title string) error {
	err := r.queries.UpdateTitleRu(ctx, admin.UpdateTitleRuParams{
		ID:      hikeID,
		TitleRu: title,
	})
	return logger.WrapError(err)
}

func (r repository) UpdatePreviewRu(ctx context.Context, hikeID int32, preview string) error {
	err := r.queries.UpdatePreviewRu(ctx, admin.UpdatePreviewRuParams{
		ID:        hikeID,
		PreviewRu: preview,
	})
	return logger.WrapError(err)
}

func (r repository) UpdateDescriptionRu(ctx context.Context, hikeID int32, description string) error {
	err := r.queries.UpdateDescriptionRu(ctx, admin.UpdateDescriptionRuParams{
		ID:            hikeID,
		DescriptionRu: description,
	})
	return logger.WrapError(err)
}

func (r repository) UpdateDates(ctx context.Context, hikeID int32, startsAt, endsAt time.Time) error {
	err := r.queries.UpdateDates(ctx, admin.UpdateDatesParams{
		ID:       hikeID,
		StartsAt: startsAt,
		EndsAt:   endsAt,
	})
	return logger.WrapError(err)
}

func (r repository) UpdateImagePath(ctx context.Context, hikeID int32, imagePath string) error {
	imagePathText := pgtype.Text{
		String: imagePath,
		Valid:  imagePath != "",
	}

	return r.queries.UpdateImagePath(ctx, admin.UpdateImagePathParams{
		ID:        hikeID,
		ImagePath: imagePathText,
	})
}
