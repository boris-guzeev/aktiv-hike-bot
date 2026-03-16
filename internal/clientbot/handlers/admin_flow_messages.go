package handlers

import (
	"fmt"
	"html"
	"strings"
	"time"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BookingStatus string

const (
	StatusNew        BookingStatus = "new"
	StatusInProgress BookingStatus = "in_progress"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusCompleted  BookingStatus = "completed"
	StatusCanceled   BookingStatus = "canceled"
)

var bookingStatuses = map[BookingStatus]string{
	StatusNew:        "новая",
	StatusInProgress: "в работе",
	StatusConfirmed:  "подтверждена",
	StatusCompleted:  "завершена",
	StatusCanceled:   "отменена",
}

func (b BookingStatus) String() string {
	return bookingStatuses[b]
}

func formatAdminBookingMessage(hike sqlc.GetHikeRow, bookingID int32, tgUserID int64, username, fullName string) string {
	title := html.EscapeString(hike.TitleRu)
	fullNameEsc := html.EscapeString(strings.TrimSpace(fullName))

	if fullNameEsc == "" {
		fullNameEsc = "—"
	}

	unameLine := "—"
	if strings.TrimSpace(username) != "" {
		unameLine = "@" + html.EscapeString(username)
	}

	var start, end string
	if sameDate(hike.StartsAt, hike.EndsAt) {
		start = hike.StartsAt.Format("02.01.2006 15:04")
		end = hike.EndsAt.Format("15:04")
	} else {
		start = hike.StartsAt.Format("02.01.2006 15:04")
		end = hike.EndsAt.Format("02.01.2006 15:04")
	}
	dateRange := fmt.Sprintf("%s → %s", start, end)

	userLink := fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, tgUserID, fullNameEsc)

	return fmt.Sprintf(
		"🆕 <b>Новая заявка на хайк</b>\n\n"+
			"📦 ID заявки: %d\n"+
			"📍 Хайк: %s\n"+
			"🗓 Дата: %s\n\n"+
			"Данные клиента\n"+
			"🔗 Username: %s\n"+
			"👤 Пользователь: %s\n",
		bookingID,
		title,
		dateRange,
		userLink,
		unameLine,
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

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}
