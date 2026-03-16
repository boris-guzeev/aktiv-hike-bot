package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) HandleAdminCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "booking_take:"):
		return h.handleTakeBooking(ctx, q)
	default:
		return nil
	}
}

func (h *Handler) handleTakeBooking(ctx context.Context, q *tgbot.CallbackQuery) error {
	if q == nil || q.From == nil || q.Message == nil {
		return nil
	}

	// Ensure telegram user (admin) exists
	tgUserID := q.From.ID
	tgUserName := q.From.UserName
	tgFullName := strings.TrimSpace(q.From.FirstName + " " + q.From.LastName)

	userID, err := h.queries.UpsertTelegramUser(ctx, client.UpsertTelegramUserParams{
		TgUserID:   tgUserID,
		TgUsername: toPgText(tgUserName),
		FullName:   toPgText(tgFullName),
	})
	if err != nil {
		return err
	}

	// Ensure admin exists
	err = h.queries.CreateAdminIfNotExists(ctx, userID)
	if err != nil {
		return err
	}

	// Get booking ID
	strBookingID := strings.TrimPrefix(q.Data, "booking_take:")
	bookingID64, err := strconv.ParseInt(strBookingID, 10, 32)
	if err != nil {
		return err
	}
	bookingID := int32(bookingID64)

	// Take booking by admin
	_, err = h.queries.TakeBookingInProgress(ctx, client.TakeBookingInProgressParams{
		ID:             bookingID,
		TakenByAdminID: toPgInt4(userID),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.replyCallback(q, "Эту заявку уже взяли в работу.")
			return nil
		}
		return err
	}

	// Update admin-group booking button message
	err = h.updateBookingTakenMessage(q, bookingID, tgFullName, tgUserName)
	if err != nil {
		h.log.Errorf("failed to update admin booking message: %v", err)
	}

	// Send to admin
	return nil
}

func (h *Handler) updateBookingTakenMessage(q *tgbot.CallbackQuery, bookingID int32, fullName, username string) error {
	managerName := strings.TrimSpace(fullName)
	if managerName == "" {
		managerName = "Менеджер"
	}

	managerLine := managerName
	username = strings.TrimSpace(username)
	if username != "" {
		managerLine = fmt.Sprintf("%s (@%s)", managerName, username)
	}

	text := q.Message.Text
	if text == "" {
		text = q.Message.Caption
	}

	statusLine := fmt.Sprintf(
		"\n\n🟡 <b>Взято в работу менеджером</b>\n%s",
		html.EscapeString(managerLine),
	)

	if !strings.Contains(text, "🟡 <b>Взято в работу</b>") {
		text += statusLine
	}

	edit := tgbot.NewEditMessageText(
		q.Message.Chat.ID,
		q.Message.MessageID,
		text,
	)
	edit.ParseMode = tgbot.ModeHTML
	_, err := h.bot.Request(edit)
	return err
}
