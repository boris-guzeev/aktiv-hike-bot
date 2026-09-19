package service

import "context"

type TelegramUser struct {
	ID         int32
	TgUserID   int64
	TgUsername string
	FullName   string
}

type Achievement struct {
	ID          int16
	Name        string
	Description string
	Assigned    bool
}

type Repository interface {
	UpsertTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
	ListTelegramUsers(ctx context.Context) ([]TelegramUser, error)
	GetTelegramUser(ctx context.Context, id int32) (TelegramUser, error)
	ListUserAchievements(ctx context.Context, userID int32) ([]Achievement, error)
	ToggleUserAchievement(ctx context.Context, userID int32, achievementID int16) (bool, error)
}

type Service interface {
	EnsureTelegramUser(ctx context.Context, tgUser TelegramUser) (int32, error)
	ListTelegramUsers(ctx context.Context) ([]TelegramUser, error)
	GetTelegramUser(ctx context.Context, id int32) (TelegramUser, error)
	ListUserAchievements(ctx context.Context, userID int32) ([]Achievement, error)
	ToggleUserAchievement(ctx context.Context, userID int32, achievementID int16) (bool, error)
}

func (s *service) ListTelegramUsers(ctx context.Context) ([]TelegramUser, error) {
	return s.repo.ListTelegramUsers(ctx)
}

func (s *service) GetTelegramUser(ctx context.Context, id int32) (TelegramUser, error) {
	return s.repo.GetTelegramUser(ctx, id)
}

func (s *service) ListUserAchievements(ctx context.Context, userID int32) ([]Achievement, error) {
	return s.repo.ListUserAchievements(ctx, userID)
}

func (s *service) ToggleUserAchievement(ctx context.Context, userID int32, achievementID int16) (bool, error) {
	return s.repo.ToggleUserAchievement(ctx, userID, achievementID)
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
