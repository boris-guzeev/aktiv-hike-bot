package handlers

import (
	"fmt"
	"html"
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func formatAdminBookingMessage(
	hike sqlc.GetHikeRow,
	bookingID int32,
	tgUserID int64,
	username string,
	fullName string,
) string {

	title := html.EscapeString(hike.TitleRu)
	fullNameEsc := html.EscapeString(strings.TrimSpace(fullName))

	if fullNameEsc == "" {
		fullNameEsc = "—"
	}

	unameLine := "—"
	if strings.TrimSpace(username) != "" {
		unameLine = "@" + html.EscapeString(username)
	}

	start := hike.StartsAt.Format("02.01.2006 15:04")
	end := hike.EndsAt.Format("02.01.2006 15:04")

	userLink := fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, tgUserID, fullNameEsc)

	return fmt.Sprintf(
		"🆕 <b>Новая заявка на хайк</b>\n\n"+
			"📍 <b>Хайк:</b> %s\n"+
			"🗓 <b>Дата:</b> %s → %s\n\n"+
			"👤 <b>Пользователь:</b> %s\n"+
			"🔗 <b>Username:</b> %s\n"+
			"🆔 <b>tg_user_id:</b> %d\n\n"+
			"📦 <b>Booking ID:</b> %d\n"+
			"🟡 <b>Статус:</b> pending",
		title,
		start, end,
		userLink,
		unameLine,
		tgUserID,
		bookingID,
	)
}

func adminBookingKeyboard(bookingID int32) tgbot.InlineKeyboardMarkup {
	return tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"🟢 Взять в работу",
				fmt.Sprintf("booking_take:%d", bookingID),
			),
		),
	)
}
