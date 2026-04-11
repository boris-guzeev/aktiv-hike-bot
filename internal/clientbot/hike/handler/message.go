package handler

import (
	"context"
	"fmt"
	"html"
	"path/filepath"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/hike"
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

		kb := hikeUI.PreviewHikeActions(r)

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

	// Dates
	b.WriteString("🗓 ")
	b.WriteString(hikeUI.FormatDateRange(hike.StartsAt, hike.EndsAt))
	b.WriteString("\n")

	// Meta
	var meta []string

	if hike.PriceGel > 0 {
		meta = append(meta, fmt.Sprintf("💵 %d GEL", hike.PriceGel))
	}

	if hike.DistanceKm > 0 {
		meta = append(meta, fmt.Sprintf("🥾 %.1f км", hike.DistanceKm))
	}

	if hike.ElevationGainM > 0 {
		meta = append(meta, fmt.Sprintf("⛰ %d м набор", hike.ElevationGainM))
	}

	if len(meta) > 0 {
		b.WriteString("\n")
		b.WriteString(strings.Join(meta, " • "))
		b.WriteString("\n")
	}

	// Preview field
	if hike.PreviewRu != "" {
		b.WriteString("\n")
		b.WriteString(html.EscapeString(hike.PreviewRu))
	}

	return b.String()
}
