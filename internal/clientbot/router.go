package clientbot

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/handlers"
	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	log     logger.Logger
	bot     *tgbot.BotAPI
	cfg     config.ClientBot
	handler *handlers.Handler
}

func NewRouter(l logger.Logger, b *tgbot.BotAPI, q *sqlc.Queries, c config.ClientBot) *router {
	return &router{
		log:     l,
		bot:     b,
		cfg:     c,
		handler: handlers.New(l, q, b, c),
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
	if q := u.CallbackQuery; q != nil && q.Message.Chat.ID == r.cfg.AdminChatID {
		return r.handler.HandleAdminCallback(ctx, q)
	}

	return nil
}
