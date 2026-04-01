package booking

import (
	"context"
	"fmt"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bookingUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/booking"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
)

func (h *BookingHandler) ShowMenu(ctx context.Context, m *tgbot.Message) error {
	msg := tgbot.NewMessage(m.Chat.ID, "Раздел заявок. Выберите действие:")
	msg.ReplyMarkup = bookingUI.BookingMenuKeyboard()

	_, err := h.bot.Send(msg)
	return logger.WrapError(err)
}

func (h *BookingHandler) ListBookings(ctx context.Context, m *tgbot.Message) error {
	if m == nil || m.From == nil {
		return nil
	}

	tgUserID := m.From.ID
	tgUsername := m.From.UserName
	fullName := strings.TrimSpace(m.From.FirstName + " " + m.From.LastName)

	adminID, err := h.userService.EnsureTelegramUser(ctx, userService.TelegramUser{
		TgUserID:   tgUserID,
		TgUsername: tgUsername,
		FullName:   fullName,
	})
	if err != nil {
		return err
	}

	bookings, err := h.bookingService.ListAdminBookings(ctx, adminID)
	if err != nil {
		return err
	}

	if len(bookings) == 0 {
		msg := tgbot.NewMessage(
			m.Chat.ID,
			"📋 <b>Мои заявки</b>\n\nУ вас пока нет активных заявок.",
		)
		msg.ParseMode = "HTML"

		_, err = h.bot.Send(msg)
		return err
	}

	for _, booking := range bookings {
		msg := tgbot.NewMessage(m.Chat.ID, bookingUI.AdminBookingCard(booking))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = bookingUI.AdminBookingActions(booking)

		if _, err := h.bot.Send(msg); err != nil {
			return fmt.Errorf("send booking message: %w", err)
		}
	}

	return nil
}
