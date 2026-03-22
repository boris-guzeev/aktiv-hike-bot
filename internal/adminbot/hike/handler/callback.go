package handler

import (
	"context"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *HikeHandler) HandleConfirm(ctx context.Context, q *tgbot.CallbackQuery) error {
	userID := q.From.ID

	if h.fsm.State(userID) != fsm.StateConfirm {
		return nil
	}

	if err := h.saveCreatedHike(ctx, userID); err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "Ошибка при сохранении хайка :("))
		return err
	}

	h.fsm.Reset(userID)

	// delete buttons
	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		tgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbot.InlineKeyboardButton{},
		},
	)
	if _, err := h.bot.Send(edit); err != nil {
		return err
	}

	_, err := h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "Хайк создан!"))
	return err
}

func (h *HikeHandler) HandleCancel(ctx context.Context, q *tgbot.CallbackQuery) error {
	userID := q.From.ID

	h.fsm.Reset(userID)

	// delete buttons
	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		tgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbot.InlineKeyboardButton{},
		},
	)

	if _, err := h.bot.Send(edit); err != nil {
		return err
	}
	_, err := h.bot.Send(tgbot.NewMessage(q.Message.Chat.ID, "Создание отменено."))
	return err
}
