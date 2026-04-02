package adminbot

import (
	"context"
	"strings"

	bookingHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/handler"
	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/handler"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/common"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	bot            *tgbot.BotAPI
	adminChatID    int64
	hikeHandler    *hikeHandler.HikeHandler
	bookingHandler *bookingHandler.BookingHandler
}

func NewRouter(b *tgbot.BotAPI, acID int64, hH *hikeHandler.HikeHandler, bH *bookingHandler.BookingHandler) *router {
	return &router{
		bot:            b,
		adminChatID:    acID,
		hikeHandler:    hH,
		bookingHandler: bH,
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
	if q := u.CallbackQuery; q != nil && q.Message != nil && q.Message.Chat.IsPrivate() {
		if r.isAdmin(q.From.ID) {
			// TODO: возможно уже тут разделять роутинг по назначению hike_ или booking_
			return r.routeCallback(ctx, q)
		}
	}

	return nil
}

func (r *router) routeMessage(ctx context.Context, m *tgbot.Message) error {
	if m.Text == "⬅️ Назад" {
		r.hikeHandler.ResetFSM(m.From.ID)
		return r.showMainMenu(m.Chat.ID)
	}

	if r.hikeHandler.InProgressFSM(m.From.ID) {
		return r.hikeHandler.HandleFSM(ctx, m)
	}

	switch m.Text {
	case "🏔 Хайки", "➕ Создать хайк", "📋 Список хайков":
		return r.routeHikeMessage(ctx, m)

	case "📥 Заявки", "📋 Список заявок", "📊 Статистика заявок":
		return r.routeBookingMessage(ctx, m)

	case "⬅️ Назад":
		// Сбросить любое текущее состояние
		return r.showMainMenu(m.Chat.ID)

	case "❓ Помощь":
		// TODO
		return nil
	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) routeHikeMessage(ctx context.Context, m *tgbot.Message) error {
	switch m.Text {
	case "🏔 Хайки":
		return r.hikeHandler.ShowMenu(ctx, m)
	case "➕ Создать хайк":
		return r.hikeHandler.StartCreateHike(ctx, m)
	case "📋 Список хайков":
		return r.hikeHandler.ListHikes(ctx, m)
	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) routeBookingMessage(ctx context.Context, m *tgbot.Message) error {
	switch m.Text {
	case "📥 Заявки":
		return r.bookingHandler.ShowMenu(ctx, m)
	case "📋 Список заявок":
		return r.bookingHandler.ListBookings(ctx, m)
	case "📊 Статистика заявок":
		// TODO: return r.bookingHandler.Stat(ctx, m)
		return nil
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
	switch {
	case strings.HasPrefix(q.Data, "hike:"):
		return nil
		// TODO: return r.hikeHandler.HandleCallback(ctx, q)

	case strings.HasPrefix(q.Data, "booking:"):
		return r.bookingHandler.HandleCallback(ctx, q)
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
