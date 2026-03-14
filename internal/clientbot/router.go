package clientbot

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/handlers"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type router struct {
	log         *logrus.Logger
	bot         *tgbot.BotAPI
	adminChatID int64
	handler     *handlers.Handler
}

func NewRouter(l *logrus.Logger, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *router {
	return &router{
		log:         l,
		bot:         b,
		adminChatID: acID,
		handler:     handlers.New(l, b, q, acID),
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private messages -> client flow
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		return r.handler.HandleClientMessage(ctx, m)
	}

	// Private callbacks -> client flow
	if q := u.CallbackQuery; q != nil && q.Message.Chat.IsPrivate() {
		return r.handler.HandleClientCallback(ctx, q)
	}

	// Admin chat callbacks -> admin flow
	if q := u.CallbackQuery; q != nil && q.Message.Chat.ID == r.adminChatID {
		return r.handler.HandleAdminCallback(ctx, q)
	}

	return nil
}
