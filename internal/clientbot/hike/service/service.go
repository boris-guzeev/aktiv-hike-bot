package service

import (
	"context"
	"errors"
	"time"
)

type Hike struct {
	ID             int32
	TitleRu        string
	DescriptionRu  string
	StartsAt       time.Time
	EndsAt         time.Time
	ImagePath      *string
	PriceGel       int32
	DistanceKm     float64
	ElevationGainM int
}

var (
	ErrHikesNotFound = errors.New("hikes not found")
)

type Repository interface {
	GetHike(ctx context.Context, id int32) (Hike, error)
	ListActualHikes(ctx context.Context, limit, offset int32) ([]Hike, error)
}

type Service interface {
	GetHike(ctx context.Context, id int32) (Hike, error)
	ListActualHikes(ctx context.Context, page, size int32) ([]Hike, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) ListActualHikes(ctx context.Context, page, size int32) ([]Hike, error) {
	offset := (page - 1) * size
	return s.repo.ListActualHikes(ctx, size, offset)
}

func (s *service) GetHike(ctx context.Context, id int32) (Hike, error) {
	return s.repo.GetHike(ctx, id)
}
