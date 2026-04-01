package service

import "context"

type TelegramUser struct {
	ID         int32
	TgUserID   int64
	TgUsername string
	FullName   string
}

type Repository interface {
	UpsertTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
}

type Service interface {
	EnsureTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) EnsureTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error) {
	return s.repo.UpsertTelegramUser(ctx, tgUser)
}
