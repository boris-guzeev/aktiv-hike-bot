package handlers

import (
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	log     logger.Logger
	bot     *tgbot.BotAPI
	queries *sqlc.Queries
	cfg     config.ClientBot
}

func New(l logger.Logger, q *sqlc.Queries, b *tgbot.BotAPI, c config.ClientBot) *Handler {
	return &Handler{
		log:     l,
		bot:     b,
		queries: q,
		cfg:     c,
	}
}

func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}
