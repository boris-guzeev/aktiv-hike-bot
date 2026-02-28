package handlers

import (
	"context"
	"html"
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbot.BotAPI
	queries *sqlc.Queries
}

func New(b *tgbot.BotAPI, q *sqlc.Queries) *Handler {
	return &Handler{bot: b, queries: q}
}

const (
	btnActual = "🥾 Актуальные хайки"
	btnMy     = "🧾 Мои записи"
	btnHelp   = "ℹ️ Помощь"
)

func (h *Handler) HandleMessage(ctx context.Context, m *tgbot.Message) error {
	switch strings.TrimSpace(m.Text) {
	case btnActual:
		return h.showActual(ctx, m.Chat.ID)
	case btnMy:
		return h.showMyBookings(ctx, m.Chat.ID)
	case btnHelp:
		return h.sendHelp(m.Chat.ID)
	default:
		return h.sendMainMenu(m.Chat.ID)
	}
}

func (h *Handler) HandleCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	// Пока колбэков у клиента нет
	return nil
}

func (h *Handler) sendMainMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Выберите действие:")
	kb := tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(btnActual),
			tgbot.NewKeyboardButton(btnMy),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(btnHelp),
		),
	)
	kb.ResizeKeyboard = true
	kb.InputFieldPlaceholder = "Например: 🥾 Актуальные хайки"
	msg.ReplyMarkup = kb
	msg.DisableWebPagePreview = true
	msg.ParseMode = "HTML"
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) sendHelp(chatID int64) error {
	txt := "ℹ️ <b>Помощь</b>\n\n• «🥾 Актуальные хайки» — ближайшие походы.\n• «🧾 Мои записи» — ваши заявки.\n• Вопросы: @your_support"
	msg := tgbot.NewMessage(chatID, txt)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) showActual(ctx context.Context, chatID int64) error {
	rows, err := h.queries.ListActualHikes(ctx, sqlc.ListActualHikesParams{
		Limit: 20, Offset: 0,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(chatID, "Пока нет актуальных хайков."))
		return err
	}
	
	{
		msg := tgbot.NewMessage(chatID, "🥾 <b>Актуальные хайки</b>")
		msg.ParseMode = "HTML"
		_, err = h.bot.Send(msg)
		if err != nil {
			return err
		}
	}

	for _, r := range rows {
		var b strings.Builder

		// Title
		b.WriteString("🏔 <b>")
		b.WriteString(html.EscapeString(r.TitleRu))
		b.WriteString("</b>\n")

		// Starts At
		b.WriteString("📅 ")
		b.WriteString(r.StartsAt.Format("02 January 2006"))
		b.WriteString("\n")

		// Ends At
		b.WriteString("📅 ")
		b.WriteString(r.EndsAt.Format("02 January 2006"))
		b.WriteString("\n")

		// Description Ru
		if r.DescriptionRu != "" {
			b.WriteString("\n")
			b.WriteString(html.EscapeString(r.DescriptionRu))
			b.WriteString("\n")
		}
	}

	return err
}

func (h *Handler) showMyBookings(ctx context.Context, chatID int64) error {
	// Заглушка — позже подключите вашу таблицу
	_, err := h.bot.Send(tgbot.NewMessage(chatID, "Здесь появятся ваши записи на хайки."))
	return err
}