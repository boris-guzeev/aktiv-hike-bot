package adminbot

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/handler"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/common"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	bot         *tgbot.BotAPI
	adminChatID int64
	hikeHandler *handler.HikeHandler
}

func NewRouter(b *tgbot.BotAPI, acID int64, h *handler.HikeHandler) *router {
	return &router{
		bot:         b,
		adminChatID: acID,
		hikeHandler: h,
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private messages
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		if r.isAdmin(m.From.ID) {
			return r.routeMessage(ctx, m)
		}
	}

	// Private callbacks
	if q := u.CallbackQuery; q != nil && q.Message.Chat.IsPrivate() {
		if r.isAdmin(q.From.ID) {
			return r.routeCallback(ctx, q)
		}
	}

	return nil
}

func (r *router) routeMessage(ctx context.Context, m *tgbot.Message) error {
	if r.hikeHandler.InProgress(m.From.ID) {
		return r.hikeHandler.HandleCreateHike(ctx, m)
	}

	switch m.Text {
	case "🏔 Хайки":
		return r.hikeHandler.ShowMenu(ctx, m)
	case "➕ Создать хайк":
		return r.hikeHandler.StartCreateHike(ctx, m)
	case "📋 Список хайков":
		return r.hikeHandler.ListHikes(ctx, m)
	case "📥 Заявки":
		// TODO
	case "Список заявок":
		// TODO
	case "❓ Помощь":

	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) showMainMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Выберите раздел")
	msg.ReplyMarkup = common.MainMenu()

	_, err := r.bot.Send(msg)
	return err
}

func (r *router) routeCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch q.Data {

	}

	return nil
}

func (r *router) isAdmin(userID int64) bool {
	m, err := r.bot.GetChatMember(tgbot.GetChatMemberConfig{
		ChatConfigWithUser: tgbot.ChatConfigWithUser{
			ChatID: r.adminChatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false
	}

	return m.Status != "left" && m.Status != "kicked"
}
