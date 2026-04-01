package booking

import (
	"context"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bookingUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/booking"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"

	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
)

func (h *BookingHandler) HandleCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil {
		return nil
	}

	switch {
	case strings.HasPrefix(q.Data, "booking:confirm:"):
		return h.AskConfirmAction(ctx, q, "confirm")

	case strings.HasPrefix(q.Data, "booking:cancel:"):
		return h.AskConfirmAction(ctx, q, "cancel")

	case strings.HasPrefix(q.Data, "booking:complete:"):
		return h.AskConfirmAction(ctx, q, "complete")

	case strings.HasPrefix(q.Data, "booking:apply:"):
		return h.ApplyAction(ctx, q)

	case strings.HasPrefix(q.Data, "booking:back:"):
		return h.RestoreActions(ctx, q)
	}

	return nil
}

func (h *BookingHandler) AskConfirmAction(ctx context.Context, q *tgbot.CallbackQuery, action string) error {
	if q == nil || q.Message == nil {
		return nil
	}

	var text string

	switch action {
	case "confirm":
		text = "Подтвердить заявку?"
	case "cancel":
		text = "Отменить заявку?"
	case "complete":
		text = "Завершить заявку?"
	default:
		return nil
	}

	parsedAction, bookingID, ok := parseBookingAction(q.Data)
	if !ok {
		return nil
	}

	msg := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		bookingUI.ConfirmActionKeyboard(parsedAction, bookingID),
	)

	if _, err := h.bot.Send(msg); err != nil {
		return logger.WrapError(err)
	}

	cb := tgbot.NewCallback(q.ID, text)
	_, err := h.bot.Request(cb)

	return logger.WrapError(err)
}

func (h *BookingHandler) ApplyAction(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.Message == nil || q.From == nil {
		return nil
	}

	action, bookingID, ok := parseBookingApplyAction(q.Data)
	if !ok {
		return h.answerCallback(q.ID, "Не удалось обработать действие.")
	}

	adminID, err := h.userService.EnsureTelegramUser(ctx, userService.TelegramUser{
		TgUserID:   q.From.ID,
		TgUsername: q.From.UserName,
		FullName:   strings.TrimSpace(q.From.FirstName + " " + q.From.LastName),
	})
	if err != nil {
		return err
	}

	var newStatus bookingService.BookingStatus
	var successText string

	switch action {
	case "confirm":
		newStatus = bookingService.StatusConfirmed
		successText = "Заявка подтверждена."
	case "cancel":
		newStatus = bookingService.StatusCanceled
		successText = "Заявка отменена."
	case "complete":
		newStatus = bookingService.StatusCompleted
		successText = "Заявка завершена."
	default:
		return h.answerCallback(q.ID, "Неизвестное действие.")
	}

	updatedBooking, err := h.bookingService.UpdateStatus(ctx, bookingID, adminID, newStatus)
	if err != nil {
		switch err {
		case bookingService.ErrNotYourBooking:
			return h.answerCallback(q.ID, "Это не ваша заявка.")
		case bookingService.ErrInvalidStatusTransition:
			return h.answerCallback(q.ID, "Недопустимая смена статуса.")
		default:
			return h.answerCallback(q.ID, "Не удалось изменить статус заявки.")
		}
	}

	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		bookingUI.AdminBookingActions(updatedBooking),
	)

	if _, err := h.bot.Send(edit); err != nil {
		return logger.WrapError(err)
	}

	return h.answerCallback(q.ID, successText)
}

func (h *BookingHandler) RestoreActions(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.Message == nil {
		return nil
	}

	bookingID, ok := parseBookingBackAction(q.Data)
	if !ok {
		return h.answerCallback(q.ID, "Не удалось восстановить действия.")
	}

	booking, err := h.bookingService.GetByID(ctx, bookingID)
	if err != nil {
		return h.answerCallback(q.ID, "Не удалось загрузить заявку.")
	}

	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		bookingUI.AdminBookingActions(booking),
	)

	if _, err := h.bot.Send(edit); err != nil {
		return logger.WrapError(err)
	}

	return h.answerCallback(q.ID, "Действие отменено.")
}

func (h *BookingHandler) answerCallback(callbackID, text string) error {
	cb := tgbot.NewCallback(callbackID, text)
	_, err := h.bot.Request(cb)
	return logger.WrapError(err)
}

func parseBookingApplyAction(data string) (action string, bookingID int32, ok bool) {
	// booking:apply:confirm:15

	parts := strings.Split(data, ":")
	if len(parts) != 4 {
		return "", 0, false
	}

	if parts[0] != "booking" || parts[1] != "apply" {
		return "", 0, false
	}

	id64, err := strconv.ParseInt(parts[3], 10, 32)
	if err != nil {
		return "", 0, false
	}

	return parts[2], int32(id64), true
}

func parseBookingBackAction(data string) (bookingID int32, ok bool) {
	// booking:back:15

	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		return 0, false
	}

	if parts[0] != "booking" || parts[1] != "back" {
		return 0, false
	}

	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return 0, false
	}

	return int32(id64), true
}

func parseBookingAction(data string) (action string, bookingID int32, ok bool) {
	// booking:confirm:15

	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		return "", 0, false
	}

	if parts[0] != "booking" {
		return "", 0, false
	}

	id64, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return "", 0, false
	}

	return parts[1], int32(id64), true
}
