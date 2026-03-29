package clientbot

import (
	"context"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/app/config"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bookingHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/handler"
	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/handler"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/common"
)

type router struct {
	bot         *tgbot.BotAPI
	cfg         config.ClientBot
	hikeHandler *hikeHandler.Handler
	bookHandler *bookingHandler.Handler
}

func NewRouter(b *tgbot.BotAPI, c config.ClientBot, hH *hikeHandler.Handler, bH *bookingHandler.Handler) *router {
	return &router{
		bot:         b,
		cfg:         c,
		hikeHandler: hH,
		bookHandler: bH,
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private messages -> client flow
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		return r.routeMessage(ctx, m)
	}

	// Private callbacks -> client flow
	if q := u.CallbackQuery; q != nil && q.Message != nil && q.Message.Chat.IsPrivate() {
		return r.routeClientCallback(ctx, q)
	}

	// Admin chat callbacks -> admin flow
	if q := u.CallbackQuery; q != nil && q.Message != nil && q.Message.Chat.ID == r.cfg.AdminChatID {
		return r.routeAdminCallback(ctx, q)
	}

	return nil
}

func (r *router) routeMessage(ctx context.Context, m *tgbot.Message) error {
	switch m.Text {
	case "🥾 Актуальные хайки":
		return r.hikeHandler.ListActualHikes(ctx, m)
		// case: "🧾 Мои записи":
		// 	// TODO

		// case: "ℹ️ Помощь"
		// 	// TODO
		// }

	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) showMainMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Выберите раздел")
	msg.ReplyMarkup = common.MainMenu()

	_, err := r.bot.Send(msg)
	return err
}

func (r *router) routeClientCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "book_hike:"):
		return r.bookHandler.BookHike(ctx, q)
	case q.Data == "booking_sent":
		return r.bookHandler.BookSent(ctx, q)
	}

	return nil
}

func (r *router) routeAdminCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "booking_take:"):
		return r.bookHandler.TakeBooking(ctx, q)
	default:
		return nil
	}
}
