package booking

import (
	"fmt"
	"html"
	"strings"

	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
)

func AdminBookingCard(b bookingService.Booking) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📋 <b>Заявка #%d</b>\n", b.ID))
	sb.WriteString(fmt.Sprintf("Статус: %s\n", statusLabel(b.Status)))
	sb.WriteString(fmt.Sprintf("Хайк: <b>%s</b>\n", html.EscapeString(b.HikeTitle)))

	clientName := html.EscapeString(strings.TrimSpace(b.UserName))
	if clientName == "" {
		clientName = "—"
	}
	sb.WriteString(fmt.Sprintf("Клиент: %s\n", clientName))

	if b.UserTgID != 0 {
		sb.WriteString(fmt.Sprintf("Telegram ID: <code>%d</code>\n", b.UserTgID))
	}

	sb.WriteString(fmt.Sprintf("Создана: %s", b.CreatedAt.Format("02.01.2006 15:04")))

	return sb.String()
}

func statusLabel(status bookingService.BookingStatus) string {
	switch status {
	case bookingService.StatusInProgress:
		return "🟡 В работе"
	case bookingService.StatusConfirmed:
		return "✅ Подтверждена"
	case bookingService.StatusCompleted:
		return "🏁 Завершена"
	case bookingService.StatusCanceled:
		return "❌ Отменена"
	default:
		return "🆕 Новая"
	}
}
