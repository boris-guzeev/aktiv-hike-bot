package booking

import (
	"fmt"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AdminBookingKeyboard(bookingID int32) tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"🟢 Взять в работу",
				fmt.Sprintf("booking_take:%d", bookingID),
			),
		),
	)
}
