package handler

import (
	"context"
	"fmt"
	"html"
	"path/filepath"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	hikeFormatter "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/hike"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ListActualHikes(ctx context.Context, m *tgbot.Message) error {
	rows, err := h.service.ListActualHikes(ctx, 1, 20)

	if err != nil {
		return err
	}

	if len(rows) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Пока нет актуальных хайков."))
		return err
	}

	for _, r := range rows {
		caption := buildHikeCaption(r)

		kb := tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData(
					"🥾 Забронировать",
					fmt.Sprintf("book_hike:%d", r.ID),
				),
			),
		)

		if r.ImagePath != nil && *r.ImagePath != "" {
			imagePath := filepath.Join(h.cfg.StorageRoot, *r.ImagePath)
			msg := tgbot.NewPhoto(m.Chat.ID, tgbot.FilePath(imagePath))
			msg.Caption = caption
			msg.ParseMode = tgbot.ModeHTML
			msg.ReplyMarkup = kb

			if _, err := h.bot.Send(msg); err != nil {
				return logger.WrapError(err)
			}

			continue
		}

		msg := tgbot.NewMessage(m.Chat.ID, caption)
		msg.ParseMode = tgbot.ModeHTML
		msg.ReplyMarkup = kb

		if _, err := h.bot.Send(msg); err != nil {
			return logger.WrapError(err)
		}
	}

	return nil
}

func buildHikeCaption(hike service.Hike) string {
	var b strings.Builder

	// Title
	b.WriteString("🏔 <b>")
	b.WriteString(html.EscapeString(hike.TitleRu))
	b.WriteString("</b>\n")

	// StartsAt | EndsAt
	b.WriteString("🗓 ")
	b.WriteString(hikeFormatter.FormatDateRange(hike.StartsAt, hike.EndsAt))
	b.WriteString("\n")

	// Description
	if hike.DescriptionRu != "" {
		b.WriteString("\n")
		b.WriteString(html.EscapeString(hike.DescriptionRu))
	}

	return b.String()
}
