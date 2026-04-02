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

func SelectedHikeActionsKeyboard(isPublished bool) tgbot.ReplyKeyboardMarkup {
	actionText := "📢 Опубликовать хайк"
	if isPublished {
		actionText = "🙈 Скрыть хайк"
	}

	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(actionText),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("🧾 Карточка хайка"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}

func PublishConfirmKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("✅ Да, опубликовать"),
			tgbot.NewKeyboardButton("❌ Отмена"),
		),
	)
}

func HideConfirmKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("✅ Да, скрыть"),
			tgbot.NewKeyboardButton("❌ Отмена"),
		),
	)
}

func CreateHikeKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}

func HikeConfirmMenu() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("✅ Подтвердить"),
			tgbot.NewKeyboardButton("❌ Отмена"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("⬅️ Назад"),
		),
	)
}
