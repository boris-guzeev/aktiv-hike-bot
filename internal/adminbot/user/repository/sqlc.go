package repository

import (
	"context"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	queries *sqlc.Queries
}

func New(q *sqlc.Queries) service.Repository {
	return &repository{queries: q}
}

func (r *repository) UpsertTelegramUser(ctx context.Context, tgUser service.TelegramUser) (int32, error) {
	id, err := r.queries.UpsertTelegramUser(ctx, sqlc.UpsertTelegramUserParams{
		TgUserID:   tgUser.TgUserID,
		TgUsername: toPgText(tgUser.TgUsername),
		FullName:   toPgText(tgUser.FullName),
	})
	if err != nil {
		return 0, logger.WrapError(err)
	}

	return id, nil
}

func (r *repository) ListTelegramUsers(ctx context.Context) ([]service.TelegramUser, error) {
	rows, err := r.queries.ListTelegramUsers(ctx)
	if err != nil {
		return nil, logger.WrapError(err)
	}
	users := make([]service.TelegramUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, service.TelegramUser{ID: row.ID, TgUserID: row.TgUserID, TgUsername: row.TgUsername, FullName: row.FullName})
	}
	return users, nil
}

func (r *repository) GetTelegramUser(ctx context.Context, id int32) (service.TelegramUser, error) {
	row, err := r.queries.GetTelegramUser(ctx, id)
	if err != nil {
		return service.TelegramUser{}, logger.WrapError(err)
	}
	return service.TelegramUser{ID: row.ID, TgUserID: row.TgUserID, TgUsername: row.TgUsername, FullName: row.FullName}, nil
}

func (r *repository) ListUserAchievements(ctx context.Context, userID int32) ([]service.Achievement, error) {
	rows, err := r.queries.ListUserAchievements(ctx, userID)
	if err != nil {
		return nil, logger.WrapError(err)
	}
	items := make([]service.Achievement, 0, len(rows))
	for _, row := range rows {
		items = append(items, service.Achievement{ID: row.ID, Name: row.Name, Description: row.Description, Assigned: row.Assigned})
	}
	return items, nil
}

func (r *repository) ToggleUserAchievement(ctx context.Context, userID int32, achievementID int16) (bool, error) {
	assigned, err := r.queries.ToggleUserAchievement(ctx, sqlc.ToggleUserAchievementParams{UserIDArg: userID, AchievementIDArg: achievementID})
	return assigned, logger.WrapError(err)
}

// TODO: вынести отдельно в utils
func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
