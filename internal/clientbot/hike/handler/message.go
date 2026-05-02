package handler

import (
	"context"
	"fmt"
	"html"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/hike"
)

func (h *Handler) ListActualHikes(ctx context.Context, m *tgbot.Message) error {
	rows, err := h.service.ListActualHikes(ctx, 1, 20)
	if err != nil {
		return err
	}

	ids := make([]int32, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}

	// Set FSM
	h.fsm.SetActualHikes(m.From.ID, ids)

	if len(rows) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Пока нет актуальных хайков."))
		return err
	}

	msg := tgbot.NewMessage(m.Chat.ID, buildActualHikesListMessage(rows))

	msg.ParseMode = tgbot.ModeHTML
	if _, err := h.bot.Send(msg); err != nil {
		return logger.WrapError(err)
	}
	return nil
}

func (h *Handler) InProgressFSM(userID int64) bool {
	return h.fsm.GetState(userID) != fsm.StateIdle
}

func (h *Handler) HandleFSM(ctx context.Context, m *tgbot.Message) error {
	userID := m.From.ID
	state := h.fsm.GetState(m.From.ID)

	switch state {
	case fsm.StateWaitingActualHikeNumber:
		return h.HandleActualHikeNumber(ctx, m)
	default:
		h.ResetFSM(userID)
		msg := tgbot.NewMessage(m.Chat.ID, "Неизвестный сценарий. Попробуйте выбрать действие заново.")
		_, err := h.bot.Send(msg)
		return err
	}
}

func (h *Handler) ResetFSM(userID int64) {
	h.fsm.Reset(userID)
}

func (h *Handler) HandleActualHikeNumber(ctx context.Context, m *tgbot.Message) error {
	userID := m.From.ID

	number, err := strconv.ParseInt(strings.TrimSpace(m.Text), 10, 32)
	if err != nil {
		return h.sendText(m.Chat.ID, "Введите номер хайка из списка, например: 1")
	}

	hikeID, ok := h.fsm.GetHikeID(userID, int32(number))
	if !ok {
		return h.sendText(m.Chat.ID, "Такого номера нет в списке. Введите номер из списка.")
	}

	h.fsm.Reset(userID)

	hike, err := h.service.GetHikeCard(ctx, hikeID)
	if err != nil {
		return err
	}

	return h.sendHikeCard(m.From.ID, hike)
}

func (h *Handler) sendHikeCard(chatID int64, hike service.HikeCard) error {
	keyboard := hikeUI.PreviewHikeActions(hike)

	caption := h.buildHikeCaption(hike)

	if hike.ImagePath != nil && *hike.ImagePath != "" {
		imagePath := filepath.Join(h.cfg.StorageRoot, *hike.ImagePath)

		msg := tgbot.NewPhoto(chatID, tgbot.FilePath(imagePath))
		msg.Caption = caption
		msg.ParseMode = tgbot.ModeHTML
		msg.ReplyMarkup = keyboard

		_, err := h.bot.Send(msg)
		return err
	}

	msg := tgbot.NewMessage(chatID, caption)
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) buildHikeCaption(hike service.HikeCard) string {
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

	if hike.DistanceKm != nil && *hike.DistanceKm > 0 {
		meta = append(meta, fmt.Sprintf("🥾 %.1f км", *hike.DistanceKm))
	}

	if hike.ElevationGainM != nil && *hike.ElevationGainM > 0 {
		meta = append(meta, fmt.Sprintf("⛰ %d м набор", *hike.ElevationGainM))
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

func (h *Handler) sendText(chatID int64, text string) error {
	msg := tgbot.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

func buildActualHikesListMessage(hikes []service.HikeListItem) string {
	var b strings.Builder

	b.WriteString("🥾 <b>Актуальные хайки</b>\n\n")

	for i, hike := range hikes {
		b.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", i+1, html.EscapeString(formatListTitle(hike.TitleRu))))

		var meta []string

		dateRange := hikeUI.FormatDateRange(hike.StartsAt, hike.EndsAt)
		if dateRange != "" {
			meta = append(meta, "🗓 "+html.EscapeString(dateRange))
		}

		if hike.DistanceKm > 0 {
			meta = append(meta, fmt.Sprintf("📏 %s км", formatDistance(hike.DistanceKm)))
		}

		if hike.ElevationGainM > 0 {
			meta = append(meta, fmt.Sprintf("⛰ +%d м", hike.ElevationGainM))
		}

		if hike.PriceGel > 0 {
			meta = append(meta, fmt.Sprintf("💰 %d GEL", hike.PriceGel))
		}

		if len(meta) > 0 {
			b.WriteString(strings.Join(meta, " • "))
			b.WriteString("\n")
		}

		if i < len(hikes)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n━━━━━━━━━━━━━━━\n\n")
	b.WriteString("✍️ Отправьте номер хайка, например: <b>1</b>,\n")
	b.WriteString("чтобы посмотреть подробности и забронировать")

	return b.String()
}

func formatDistance(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%.0f", v)
	}

	return fmt.Sprintf("%.1f", v)
}

func formatListTitle(title string) string {
	title = strings.TrimSpace(title)

	// Чтобы строка была чище в списке.
	title = strings.TrimPrefix(title, `Тур "`)
	title = strings.TrimSuffix(title, `"`)

	return title
}
