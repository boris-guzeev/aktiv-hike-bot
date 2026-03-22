package service

import (
	"context"
	"time"
)

type Hike struct {
	ID            int32
	TitleRu       string
	DescriptionRu string
	StartsAt      time.Time
	EndsAt        time.Time
	PhotoFileID   string
	IsPublished   bool
}

type Repository interface {
	GetHike(ctx context.Context, id int32) (Hike, error)
	ListHikes(ctx context.Context, limit, offset int32) ([]Hike, error)
	ListActualHikes(ctx context.Context, limit, offset int32) ([]Hike, error)
	PublishHike(ctx context.Context, id int32) error
	CreateHike(ctx context.Context, hike Hike) error
	HideHike(ctx context.Context, id int32) error
	DeleteHike(ctx context.Context, id int32) error
}

type Service interface {
	GetHike(ctx context.Context, id int32) (Hike, error)
	ListHikes(ctx context.Context, page, size int32) ([]Hike, error)
	ListActualHikes(ctx context.Context, page, size int32) ([]Hike, error)
	PublishHike(ctx context.Context, id int32) error
	CreateHike(ctx context.Context, hike Hike) error
	HideHike(ctx context.Context, id int32) error
	DeleteHike(ctx context.Context, id int32) error
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s service) GetHike(ctx context.Context, id int32) (Hike, error) {
	return s.repo.GetHike(ctx, id)
}

func (s service) ListHikes(ctx context.Context, page, size int32) ([]Hike, error) {
	offset := (page - 1) * size
	return s.repo.ListHikes(ctx, size, offset)
}

func (s service) ListActualHikes(ctx context.Context, page, size int32) ([]Hike, error) {
	offset := (page - 1) * size
	return s.repo.ListActualHikes(ctx, size, offset)
}

func (s service) PublishHike(ctx context.Context, id int32) error {
	return s.repo.PublishHike(ctx, id)
}

func (s service) CreateHike(ctx context.Context, hike Hike) error {
	return s.repo.CreateHike(ctx, hike)
}

func (s service) HideHike(ctx context.Context, id int32) error {
	return s.repo.HideHike(ctx, id)
}

func (s service) DeleteHike(ctx context.Context, id int32) error {
	return s.repo.DeleteHike(ctx, id)
}
