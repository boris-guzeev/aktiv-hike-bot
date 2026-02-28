package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	hikesPageSize = 10
	descMaxLen    = 140 // сколько символов показывать в списке
)

func (h *Handler) HandleListCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	data := q.Data

	switch {
	// --- Hikes: Confirm ---
	case strings.HasPrefix(data, "h:pub:confirm:"):
		idStr := strings.TrimPrefix(data, "h:pub:confirm:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.handlePublishConfirm(ctx, q, int32(id))

	case strings.HasPrefix(data, "h:hide:confirm:"):
		idStr := strings.TrimPrefix(data, "h:hide:confirm:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.handleHideConfirm(ctx, q, int32(id))

	case strings.HasPrefix(data, "h:del:confirm:"):
		idStr := strings.TrimPrefix(data, "h:del:confirm:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.handleDeleteConfirm(ctx, q, int32(id))

	// --- Hikes: Request Confirm ---
	case strings.HasPrefix(data, "h:pub:"):
		idStr := strings.TrimPrefix(data, "h:pub:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.askConfirmAction(
			q,
			"Опубликовать этот хайк?",
			fmt.Sprintf("h:pub:confirm:%d", id),
			"h:cancel",
		)

	case strings.HasPrefix(data, "h:hide:"):
		idStr := strings.TrimPrefix(data, "h:hide:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.askConfirmAction(
			q,
			"Скрыть этот хайк?",
			fmt.Sprintf("h:hide:confirm:%d", id),
			"h:cancel",
		)

	case strings.HasPrefix(data, "h:del:"):
		idStr := strings.TrimPrefix(data, "h:del:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		return h.askConfirmAction(
			q,
			"Удалить этот хайк?",
			fmt.Sprintf("h:del:confirm:%d", id),
			"h:cancel",
		)

	case data == "h:cancel":
		// Просто вернуть исходную клавиатуру нельзя, т.к. мы её не храним.
		// Проще убрать клаву и ничего не делать.
		h.removeKeyboard(q)
		_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "Действие отменено"))
		return nil
	}

	return nil

}

// Public
func (h *Handler) handlePublishConfirm(ctx context.Context, q *tgbot.CallbackQuery, id int32) error {
	err := h.queries.SetPublished(ctx, sqlc.SetPublishedParams{
		ID:          id,
		IsPublished: true,
	})
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "⚠️ Не удалось опубликовать хайк"))
		return err
	}

	h.removeKeyboard(q)
	_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "✅ Хайк опубликован"))
	return nil
}

// Hide from publication
func (h *Handler) handleHideConfirm(ctx context.Context, q *tgbot.CallbackQuery, id int32) error {
	err := h.queries.SetPublished(ctx, sqlc.SetPublishedParams{
		ID:          id,
		IsPublished: false,
	})
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "⚠️ Не удалось скрыть хайк"))
		return err
	}

	h.removeKeyboard(q)
	_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "🙈 Хайк скрыт"))
	return nil
}

// Delete
func (h *Handler) handleDeleteConfirm(ctx context.Context, q *tgbot.CallbackQuery, id int32) error {
	err := h.queries.DeleteHike(ctx, id)
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "⚠️ Не удалось удалить хайк"))
		return err
	}

	h.removeKeyboard(q)
	_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "🗑 Хайк удалён"))
	return nil
}

// sendHikesList shows last hikesPageSize
func (h *Handler) sendHikesList(chatID int64) error {
	ctx := context.Background()

	args := sqlc.ListHikesParams{
		Limit:  int32(hikesPageSize),
		Offset: 0,
	}

	rows, err := h.queries.ListHikes(ctx, args)
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(chatID, "⚠️ Не удалось получить список хайков"))
		return err
	}

	if len(rows) == 0 {
		_, err := h.bot.Send(tgbot.NewMessage(chatID, "Пока нет ни одного хайка. Нажми /newhike, чтобы создать."))
		return err
	}

	// Common title
	header := tgbot.NewMessage(chatID, "🌄 <b>Хайки</b>\n\nПоказаны последние хайки:")
	header.ParseMode = "HTML"
	if _, err := h.bot.Send(header); err != nil {
		return err
	}

	// Send every hike with inline-buttons
	for i, r := range rows {
		pub := "🕗 Черновик"
		if r.IsPublished {
			pub = "✅ Опубликован"
		}

		// Title
		title := strings.TrimSpace(r.TitleRu)
		if title == "" && r.TitleEn.Valid {
			title = strings.TrimSpace(r.TitleEn.String)
		}
		if title == "" {
			title = "(без названия)"
		}

		// Short desc
		desc := normalizeOneLine(strings.TrimSpace(r.DescriptionRu))
		if desc == "" {
			desc = "(описание не заполнено)"
		}
		desc = truncateRunes(desc, descMaxLen)

		// Dates
		start := r.StartsAt.In(time.Local).Format("02.01 15:04")
		end := r.EndsAt.In(time.Local).Format("02.01 15:04")
		created := r.CreatedAt.In(time.Local).Format("02.01 15:04")

		var b strings.Builder
		fmt.Fprintf(&b, "%d. <b>%s</b>\n", i+1, title)
		fmt.Fprintf(&b, "📝 %s\n", desc)
		fmt.Fprintf(&b, "📅 %s — %s\n", start, end)
		fmt.Fprintf(&b, "📤 %s   •   🕰 создано %s\n", pub, created)
		fmt.Fprintf(&b, "\nID: <code>%d</code>\n", r.ID)

		msg := tgbot.NewMessage(chatID, b.String())
		msg.ParseMode = "HTML"

		// Inline-buttons
		kb := tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("✅ Опубликовать", fmt.Sprintf("h:pub:%d", r.ID)),
				tgbot.NewInlineKeyboardButtonData("🙈 Скрыть", fmt.Sprintf("h:hide:%d", r.ID)),
			),
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("🗑 Удалить", fmt.Sprintf("h:del:%d", r.ID)),
			),
		)
		msg.ReplyMarkup = kb

		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) askConfirmAction(q *tgbot.CallbackQuery, question, confirmData, cancelData string) error {
	kb := tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData("✅ Да", confirmData),
			tgbot.NewInlineKeyboardButtonData("❌ Отмена", cancelData),
		),
	)

	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		kb,
	)

	_, err := h.bot.Send(edit)
	return err
}

// remove message keyboard after a success action
func (h *Handler) removeKeyboard(q *tgbot.CallbackQuery) {
	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		tgbot.InlineKeyboardMarkup{InlineKeyboard: [][]tgbot.InlineKeyboardButton{}},
	)
	_, _ = h.bot.Send(edit)
}

// truncateRunes truncates runes quantity and adds "..."
func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit-1]) + "…"
}

// normalizeOneLine replaces newlines/tabs/double spaces with a single space
func normalizeOneLine(s string) string {
	spacey := regexp.MustCompile(`[\r\n\t]+`)
	multi := regexp.MustCompile(`\s{2,}`)
	out := spacey.ReplaceAllString(s, " ")
	out = multi.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}
