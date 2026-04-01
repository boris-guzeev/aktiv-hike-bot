package booking

import (
	"fmt"

	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AdminBookingActions(b bookingService.Booking) tgbot.InlineKeyboardMarkup {
	switch b.Status {
	case bookingService.StatusInProgress:
		return tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("✅ Подтвердить", fmt.Sprintf("booking:confirm:%d", b.ID)),
				tgbot.NewInlineKeyboardButtonData("❌ Отменить", fmt.Sprintf("booking:cancel:%d", b.ID)),
			),
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("🏁 Завершить", fmt.Sprintf("booking:complete:%d", b.ID)),
			),
		)

	case bookingService.StatusConfirmed:
		return tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("🏁 Завершить", fmt.Sprintf("booking:complete:%d", b.ID)),
				tgbot.NewInlineKeyboardButtonData("❌ Отменить", fmt.Sprintf("booking:cancel:%d", b.ID)),
			),
		)

	default:
		return tgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbot.InlineKeyboardButton{},
		}
	}
}

func ConfirmActionKeyboard(action string, bookingID int32) tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData("✅ Да", makeBookingApplyData(action, bookingID)),
			tgbot.NewInlineKeyboardButtonData("↩️ Нет", makeBookingBackData(bookingID)),
		),
	)
}
