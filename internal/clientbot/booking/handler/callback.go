package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	bookingService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/service"
	hikeService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/hike/service"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/user/service"

	bookingUI "github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/ui/booking"
)

func (h *Handler) BookHike(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.From == nil || q.Message == nil {
		return nil
	}

	idStr := strings.TrimPrefix(q.Data, "book_hike:")
	hikeID64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		h.replyCallback(q, "Не удалось обработать запрос.")
		return logger.WrapError(fmt.Errorf("parse hike id error: %v (data=%q)", err, q.Data))
	}
	hikeID := int32(hikeID64)

	tgUserID := q.From.ID
	username := q.From.UserName
	fullName := strings.TrimSpace(q.From.FirstName + " " + q.From.LastName)

	// 1) Check if user exists
	userID, err := h.userService.EnsureTelegramUser(ctx, userService.TelegramUser{
		TgUserID:   tgUserID,
		TgUsername: username,
		FullName:   fullName,
	})
	if err != nil {
		_ = h.replyCallback(q, "Ошибка. Пожалуйста, попробуйте позже.")
		return err
	}

	// 2) Get Hike
	hike, err := h.hikeService.GetHike(ctx, hikeID)
	if err != nil {
		if errors.Is(err, hikeService.ErrHikesNotFound) {
			_ = h.replyCallback(q, "К сожалению, этот хайк недоступен.")
		}
		return err
	}

	// 3) Create booking and set status to new
	bookingID, err := h.bookingService.Create(ctx, hikeID, userID)
	if err != nil {
		if errors.Is(err, bookingService.ErrBookingAlreadyExists) {
			_ = h.replyCallback(q, "У Вас уже есть заявка на этот хайк ✅ Мы её обрабатываем.")
			return err
		}
		_ = h.replyCallback(q, "Ошибка. Пожалуйста, попробуйте позже.")
		return logger.WrapError(fmt.Errorf("failed to create booking: %w", err))
	}

	// 4) Change inline-button text
	newKb := tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"⏳ Запрос отправлен",
				"booking_sent",
			),
		),
	)

	// 5) Send new Message with changed button
	edit := tgbot.NewEditMessageReplyMarkup(
		q.Message.Chat.ID,
		q.Message.MessageID,
		newKb,
	)
	if _, err := h.bot.Send(edit); err != nil {
		return logger.WrapError(err)
	}

	// 6) Info user if hike is booked successfully
	_ = h.replyCallback(q, "Ваша заявка отправлена ✅ Мы передали её менеджерам.")

	// 7) Form and send admin message
	msg := tgbot.NewMessage(h.cfg.AdminChatID, bookingUI.AdminBookingMessage(
		hike,
		bookingID,
		tgUserID,
		username,
		fullName,
	))
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = bookingUI.AdminBookingKeyboard(bookingID)

	if _, err := h.bot.Send(msg); err != nil {
		return logger.WrapError(fmt.Errorf("failed to send admin message to chat=%v: %w", h.cfg.AdminChatID, err))
	}

	return nil
}

func (h *Handler) BookSent(ctx context.Context, q *tgbot.CallbackQuery) error {
	return h.replyCallback(q, "Заявка уже отправлена ✅")
}

func (h *Handler) TakeBooking(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.From == nil || q.Message == nil {
		return nil
	}

	// Ensure telegram user (admin) exists
	tgUserID := q.From.ID
	tgUserName := q.From.UserName
	tgFullName := strings.TrimSpace(q.From.FirstName + " " + q.From.LastName)

	userID, err := h.userService.EnsureTelegramUser(ctx, userService.TelegramUser{
		TgUserID:   tgUserID,
		TgUsername: tgUserName,
		FullName:   tgFullName,
	})
	if err != nil {
		return err
	}

	// Ensure admin exists
	err = h.adminService.Ensure(ctx, userID)
	if err != nil {
		return err
	}

	// Get booking ID
	strBookingID := strings.TrimPrefix(q.Data, "booking_take:")
	bookingID64, err := strconv.ParseInt(strBookingID, 10, 32)
	if err != nil {
		return logger.WrapError(err)
	}
	bookingID := int32(bookingID64)

	// Take booking by admin
	_, err = h.bookingService.TakeInProgress(ctx, bookingID, userID)
	if err != nil {
		if errors.Is(err, bookingService.ErrBookingAlreadyTaken) {
			_ = h.replyCallback(q, "Эту заявку уже взяли в работу.")
			return err
		}

		return err
	}

	// Update admin-group booking button message
	err = h.updateBookingTakenMessage(q, tgFullName, tgUserName)
	if err != nil {
		return fmt.Errorf("failed to update admin booking message: %w", err)
	}

	return nil
}

func (h *Handler) updateBookingTakenMessage(q *tgbot.CallbackQuery, fullName, username string) error {
	text := q.Message.Text
	if text == "" {
		text = q.Message.Caption
	}

	newText := bookingUI.BookingTakenMessage(text, fullName, username)

	edit := tgbot.NewEditMessageText(
		q.Message.Chat.ID,
		q.Message.MessageID,
		newText,
	)
	edit.ParseMode = tgbot.ModeHTML

	_, err := h.bot.Request(edit)
	return logger.WrapError(err)
}

func (h *Handler) replyCallback(q *tgbot.CallbackQuery, text string) error {
	cfg := tgbot.CallbackConfig{
		CallbackQueryID: q.ID,
		Text:            text,
	}
	_, err := h.bot.Request(cfg)

	return err
}
