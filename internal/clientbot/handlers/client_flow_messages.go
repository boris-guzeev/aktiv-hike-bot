package handlers

import (
	"context"
	"fmt"
	"html"
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	btnActual = "🥾 Актуальные хайки"
	btnMy     = "🧾 Мои записи"
	btnHelp   = "ℹ️ Помощь"
)

// Handles client-bot reply-buttons clicks
func (h *Handler) HandleClientMessage(ctx context.Context, m *tgbot.Message) error {
	switch strings.TrimSpace(m.Text) {
	case btnActual:
		return h.showActual(ctx, m.Chat.ID)
	case btnMy:
		return h.showMyBookings(ctx, m.Chat.ID)
	case btnHelp:
		return h.showHelp(m.Chat.ID)
	default:
		return h.showMainMenu(m.Chat.ID)
	}
}

func (h *Handler) showMainMenu(chatID int64) error {
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
	msg.ReplyMarkup = kb
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) showHelp(chatID int64) error {
	txt := "ℹ️ <b>Помощь</b>\n\n• «🥾 Актуальные хайки» — ближайшие походы.\n• «🧾 Мои записи» — Ваши заявки.\n• Вопросы: @your_support"
	msg := tgbot.NewMessage(chatID, txt)
	msg.ParseMode = "HTML"
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) showActual(ctx context.Context, chatID int64) error {
	rows, err := h.queries.ListActualHikes(ctx, sqlc.ListActualHikesParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(chatID, "Пока нет актуальных хайков."))
		return err
	}

	for _, r := range rows {
		var b strings.Builder

		b.WriteString("🏔 <b>")
		b.WriteString(html.EscapeString(r.TitleRu))
		b.WriteString("</b>\n")

		b.WriteString("📅 ")
		b.WriteString(r.StartsAt.Format("02 January 2006"))
		b.WriteString("\n")

		b.WriteString("📅 ")
		b.WriteString(r.EndsAt.Format("02 January 2006"))
		b.WriteString("\n")

		if r.DescriptionRu != "" {
			b.WriteString("\n")
			b.WriteString(html.EscapeString(r.DescriptionRu))
			b.WriteString("\n")
		}

		kb := tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("🥾 Забронировать", fmt.Sprintf("book_hike:%d", r.ID)),
			),
		)

		msg := tgbot.NewMessage(chatID, b.String())
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = kb

		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) showMyBookings(ctx context.Context, chatID int64) error {
	_, err := h.bot.Send(tgbot.NewMessage(chatID, "Здесь появятся Ваши записи на хайки."))
	return err
}
