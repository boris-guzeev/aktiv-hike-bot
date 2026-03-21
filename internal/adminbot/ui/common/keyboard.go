package common

import tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func MainMenu() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("🏔 Хайки"),
			tgbot.NewKeyboardButton("📥 Заявки"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("❓ Помощь"),
		),
	)
}
