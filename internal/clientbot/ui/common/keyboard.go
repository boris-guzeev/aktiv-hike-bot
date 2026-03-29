package common

import tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func MainMenu() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("🥾 Актуальные хайки"),
			tgbot.NewKeyboardButton("🧾 Мои записи"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("ℹ️ Помощь"),
		),
	)
}
