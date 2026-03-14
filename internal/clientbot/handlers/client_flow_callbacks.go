package handlers

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handles client-bot inline-button clicks
func (h *Handler) HandleClientCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "book_hike:"):
		h.onClientBookHike(ctx, q)
	case q.Data == "booking_sent":
		h.replyCallback(q, "Заявка уже отправлена ✅")
	}
	return nil
}

func (h *Handler) onClientBookHike(ctx context.Context, q *tgbot.CallbackQuery) {
	idStr := strings.TrimPrefix(q.Data, "book_hike:")
	hikeID64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		h.log.Errorf("parse hike id error: %v (data=%q)", err, q.Data)
		h.replyCallback(q, "Не удалось обработать запрос.")
		return
	}
	hikeID := int32(hikeID64)

	tgUserID := q.From.ID
	username := q.From.UserName
	fullName := strings.TrimSpace(q.From.FirstName + " " + q.From.LastName)

	// 1) Check if user exists
	userID, err := h.queries.UpsertTelegramUser(ctx, sqlc.UpsertTelegramUserParams{
		TgUserID:   tgUserID,
		TgUsername: toPgText(username),
		FullName:   toPgText(fullName),
	})
	if err != nil {
		h.log.Errorf("failed to upsert user: %v", err)
		h.replyCallback(q, "Ошибка. Пожалуйста, попробуйте позже.")
		return
	}

	// 2) Get Hike
	hike, err := h.queries.GetHike(ctx, hikeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.replyCallback(q, "К сожалению, этот хайк недоступен.")
			return
		}
		h.log.Errorf("failed to get hike: %v", err)
		return
	}

	// 3) Create booking and set status to pending
	bookingID, err := h.queries.CreateBookingPending(ctx, sqlc.CreateBookingPendingParams{
		HikeID: hikeID,
		UserID: userID,
	})
	if err != nil || bookingID == 0 {
		if errors.Is(err, sql.ErrNoRows) {
			h.replyCallback(q, "У Вас уже есть заявка на этот хайк ✅ Мы её обрабатываем.")
			return
		}
		h.log.Errorf("failed to create booking: %v", err)
		return
	}

	// 4) Change inline-button text
	newKb := tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			tgbot.NewInlineKeyboardButtonData(
				"⏳ Запрос отправлен",
				"booking_sent",
				//fmt.Sprintf("book_hike:%d", hikeID),
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
		h.log.Error(err)
	}

	// 8) Info user if hike is booked successfully
	h.replyCallback(q, "Ваша заявка отправлена ✅ Мы передали её менеджерам.")

	// 9) Form and send admin message
	msg := tgbot.NewMessage(h.adminChatID, formatAdminBookingMessage(
		hike,
		bookingID,
		tgUserID,
		username,
		fullName,
	))
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = adminBookingKeyboard(bookingID)

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Errorf("failed to send admin message to chat=%v: %v", h.adminChatID, err)
	}
}

func (h *Handler) replyCallback(q *tgbot.CallbackQuery, text string) {
	cfg := tgbot.CallbackConfig{
		CallbackQueryID: q.ID,
		Text:            text,
	}
	_, _ = h.bot.Request(cfg)
}
