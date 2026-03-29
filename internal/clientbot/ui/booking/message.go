package booking

import (
	"fmt"
	"html"
	"strings"
	"time"

	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
)

func AdminBookingMessage(hike hikeService.Hike, bookingID int32, tgUserID int64, username, fullName string) string {
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

func BookingTakenMessage(text, fullName, username string) string {
	managerName := strings.TrimSpace(fullName)
	if managerName == "" {
		managerName = "Менеджер"
	}

	managerLine := managerName
	username = strings.TrimSpace(username)
	if username != "" {
		managerLine = fmt.Sprintf("%s (@%s)", managerName, username)
	}

	statusLine := fmt.Sprintf(
		"\n\n🟡 <b>Взято в работу менеджером</b>\n%s",
		html.EscapeString(managerLine),
	)

	if strings.Contains(text, "🟡 <b>Взято в работу менеджером</b>") {
		return text
	}

	return text + statusLine
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}
