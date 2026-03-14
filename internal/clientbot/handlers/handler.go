package handlers

import (
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	log         *logrus.Logger
	bot         *tgbot.BotAPI
	queries     *sqlc.Queries
	adminChatID int64
}

func New(l *logrus.Logger, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *Handler {
	return &Handler{
		log:         l,
		bot:         b,
		queries:     q,
		adminChatID: acID,
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
