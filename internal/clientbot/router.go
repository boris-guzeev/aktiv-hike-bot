package clientbot

import (
	"context"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/handlers"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	bot *tgbot.BotAPI
	handler     *handlers.Handler
}

func NewRouter(l *time.Location, b *tgbot.BotAPI, q *sqlc.Queries) *router {
	return &router{
		bot:         b,
		handler:     handlers.New(b, q),
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private commands
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		return r.handler.HandleMessage(ctx, m)
	}

	// Private callbacks
	if q := u.CallbackQuery; q != nil && q.Message.Chat.IsPrivate() {
		return r.handler.HandleCallback(ctx, q)
	}

	return nil
}