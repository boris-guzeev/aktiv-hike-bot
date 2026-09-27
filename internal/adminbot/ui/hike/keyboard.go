package hike

import (
	"strconv"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func ListHikesPaginationKeyboard(page int32, hasNext bool) tgbot.InlineKeyboardMarkup {
	buttons := make([]tgbot.InlineKeyboardButton, 0, 2)
	if page > 1 {
		buttons = append(buttons, tgbot.NewInlineKeyboardButtonData(
			"⬅️ Назад",
			CallbackListHikesPagePrefix+strconv.FormatInt(int64(page-1), 10),
		))
	}
	if hasNext {
		buttons = append(buttons, tgbot.NewInlineKeyboardButtonData(
			"Вперёд ➡️",
			CallbackListHikesPagePrefix+strconv.FormatInt(int64(page+1), 10),
		))
	}

	if len(buttons) == 0 {
		return tgbot.InlineKeyboardMarkup{}
	}

	return tgbot.NewInlineKeyboardMarkup(tgbot.NewInlineKeyboardRow(buttons...))
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

// TODO: заменить на common.OnlyBackKeyboard
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

func HikeDetailsKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditTitleRu),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditPreviewRu),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditDescriptionRu),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditDates),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditPriceGel),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonEditType),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(ButtonBack),
		),
	)
}
