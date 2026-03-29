package service

import (
	"context"
)

type Admin struct {
	ID int32
}

type Repository interface {
	CreateIfNotExists(ctx context.Context, id int32) error
}

type Service interface {
	Ensure(ctx context.Context, id int32) error
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) Ensure(ctx context.Context, id int32) error {
	return s.repo.CreateIfNotExists(ctx, id)
}
