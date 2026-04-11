package hike

import (
	"fmt"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
)

func PreviewHikeActions(hike service.Hike) tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"🥾 Забронировать",
				fmt.Sprintf("book_hike:%d", hike.ID),
			),
			tgbot.NewInlineKeyboardButtonData(
				"🔍 Подробнее",
				fmt.Sprintf("details_hike:%d", hike.ID),
			),
		),
	)
}

func DetailsHikeActions(hike service.Hike) tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"🥾 Забронировать",
				fmt.Sprintf("book_hike:%d", hike.ID),
			),
		),
	)
}
