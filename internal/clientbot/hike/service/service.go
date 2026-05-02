package service

import (
	"context"
	"errors"
	"time"
)

type HikeListItem struct {
	ID             int32
	TitleRu        string
	StartsAt       time.Time
	EndsAt         time.Time
	PriceGel       int32
	DistanceKm     float64
	ElevationGainM int
}

type HikeCard struct {
	ID             int32
	TitleRu        string
	PreviewRu      string
	ImagePath      *string
	StartsAt       time.Time
	EndsAt         time.Time
	PriceGel       int32
	DistanceKm     *float64
	ElevationGainM *int32
}

type HikeDetails struct {
	ID            int32
	TitleRu       string
	DescriptionRu string
}

var (
	ErrHikesNotFound = errors.New("hikes not found")
)

type Repository interface {
	ListActualHikes(ctx context.Context, limit, offset int32) ([]HikeListItem, error)
	GetHikeCard(ctx context.Context, id int32) (HikeCard, error)
	GetHikeDetails(ctx context.Context, id int32) (HikeDetails, error)
}

type Service interface {
	ListActualHikes(ctx context.Context, page, size int32) ([]HikeListItem, error)
	GetHikeCard(ctx context.Context, id int32) (HikeCard, error)
	GetHikeDetails(ctx context.Context, id int32) (HikeDetails, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) ListActualHikes(ctx context.Context, page, size int32) ([]HikeListItem, error) {
	offset := (page - 1) * size
	return s.repo.ListActualHikes(ctx, size, offset)
}

func (s *service) GetHikeCard(ctx context.Context, id int32) (HikeCard, error) {
	return s.repo.GetHikeCard(ctx, id)
}

func (s *service) GetHikeDetails(ctx context.Context, id int32) (HikeDetails, error) {
	return s.repo.GetHikeDetails(ctx, id)
}
