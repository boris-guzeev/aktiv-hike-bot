package router

import (
	"context"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/admin"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/notify"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
type Router struct {
	bot *tgbot.BotAPI
	adminChatID int64
	adminHandler *admin.Handler
	clientHandler *client.Handler
}

func New(l *time.Location, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *Router {
	return &Router{
		bot: b,
		adminChatID: acID,
		adminHandler: admin.New(l, b, q, acID),
		clientHandler: client.New(b, q),
	}
}

func (r *Router) Handle(ctx context.Context, u tgbot.Update) error {
	// Private commands
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		// if r.adminHandler.IsAdmin(m.From.ID) {
		// 		return r.adminHandler.HandleMessage(ctx, m)
		// }
		return r.clientHandler.HandleMessage(ctx, m)
	}

	// Private callbacks
	if q := u.CallbackQuery; q != nil && q.Message.Chat.IsPrivate() {
		// if r.adminHandler.IsAdmin(q.From.ID) {
		// 	return r.adminHandler.HandleCallback(ctx, q)
		// }
		return r.clientHandler.HandleCallback(ctx, q)
	}

	if m := u.Message; m != nil && m.Chat.ID == r.adminChatID {
		return nil
	}

	return nil
}

func (r *Router) Notifier() notify.Notifier {
	return notify.New(r.bot, r.adminChatID)
}