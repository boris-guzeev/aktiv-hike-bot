package admin

import (
	"context"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/admin/handlers"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/notify"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	bot         *tgbot.BotAPI
	adminChatID int64
	handler     *handlers.Handler
}

func NewRouter(l *time.Location, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *router {
	return &router{
		bot:         b,
		adminChatID: acID,
		handler:     handlers.New(l, b, q, acID),
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private commands
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		if r.handler.IsAdmin(m.From.ID) {
			return r.handler.HandleMessage(ctx, m)
		}
	}

	// Private callbacks
	if q := u.CallbackQuery; q != nil && q.Message.Chat.IsPrivate() {
		// if r.adminHandler.IsAdmin(q.From.ID) {
		// 	return r.adminHandler.HandleCallback(ctx, q)
		// }
	}

	return nil
}

func (r *router) Notifier() notify.Notifier {
	return notify.New(r.bot, r.adminChatID)
}
