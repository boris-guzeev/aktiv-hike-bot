package booking

import (
	"fmt"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BookingMenuKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("📋 Список заявок"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("📊 Статистика заявок"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}

func makeBookingApplyData(action string, bookingID int32) string {
	return fmt.Sprintf("booking:apply:%s:%d", action, bookingID)
}

func makeBookingBackData(bookingID int32) string {
	return fmt.Sprintf("booking:back:%d", bookingID)
}
