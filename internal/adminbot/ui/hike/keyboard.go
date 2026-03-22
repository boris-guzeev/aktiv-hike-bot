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

func ConfirmKeyboard() tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData("✅ Подтвердить", "confirm"),
			tgbot.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
		),
	)
}
