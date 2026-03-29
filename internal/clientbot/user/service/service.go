package service

import "context"

type TelegramUser struct {
	ID         int32
	TgUserID   int64
	TgUsername string
	FullName   string
	Lang       string
}

type Repository interface {
	GetByID(ctx context.Context, id int32) (TelegramUser, error)
	UpsertTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
}

type Service interface {
	GetByID(ctx context.Context, id int32) (TelegramUser, error)
	EnsureTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) GetByID(ctx context.Context, id int32) (TelegramUser, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) EnsureTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error) {
	return s.repo.UpsertTelegramUser(ctx, tgUser)
}
