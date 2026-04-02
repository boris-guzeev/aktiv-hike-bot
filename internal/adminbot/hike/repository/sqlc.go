package repository

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
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
		return service.Hike{}, err
	}
	return service.Hike{
		ID:            rawHike.ID,
		TitleRu:       rawHike.TitleRu,
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
		return nil, err
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
		return nil, err
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
	photoFileID := pgtype.Text{
		String: hike.PhotoFileID,
		Valid:  hike.PhotoFileID != "",
	}

	return r.queries.CreateHike(ctx, admin.CreateHikeParams{
		TitleRu:       hike.TitleRu,
		DescriptionRu: hike.DescriptionRu,
		StartsAt:      hike.StartsAt,
		EndsAt:        hike.EndsAt,
		PhotoFileID:   photoFileID,
	})
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
