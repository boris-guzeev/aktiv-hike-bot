package user

import tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func UserCardKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(tgbot.NewKeyboardButton(ButtonEditAchievements)),
		tgbot.NewKeyboardButtonRow(tgbot.NewKeyboardButton(ButtonBack)),
	)
}

func BackKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(tgbot.NewKeyboardButton(ButtonBack)),
	)
}
