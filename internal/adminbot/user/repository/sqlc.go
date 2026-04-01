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

// TODO: вынести отдельно в utils
func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
