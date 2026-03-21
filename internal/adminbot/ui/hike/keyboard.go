package hike

import tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func HikeMenu() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("➕ Создать хайк"),
			tgbot.NewKeyboardButton("📋 Список хайков"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("❓ Помощь"),
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}

func backBtn() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}

func btn(text, data string) tgbot.InlineKeyboardButton {
	b := tgbot.NewInlineKeyboardButtonData(text, data)
	return b
}
