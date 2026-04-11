package handler

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/hike"
)

func (h *Handler) DetailsHike(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.From == nil || q.Message == nil {
		return nil
	}

	idStr := strings.TrimPrefix(q.Data, "details_hike:")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return logger.WrapError(err)
	}

	hike, err := h.service.GetHike(ctx, int32(id))
	if err != nil {
		return logger.WrapError(err)
	}

	text := fmt.Sprintf(
		"<b>%s</b>\n\n%s",
		html.EscapeString(hike.TitleRu),
		html.EscapeString(hike.DescriptionRu),
	)

	kb := hikeUI.DetailsHikeActions(hike)

	msg := tgbot.NewMessage(q.Message.Chat.ID, text)
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = kb

	if _, err = h.bot.Send(msg); err != nil {
		return logger.WrapError(err)
	}

	return nil
}
