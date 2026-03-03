package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	log         *logrus.Logger
	bot         *tgbot.BotAPI
	queries     *sqlc.Queries
	adminChatID int64
}

func New(l *logrus.Logger, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *Handler {
	return &Handler{
		log:         l,
		bot:         b,
		queries:     q,
		adminChatID: acID,
	}
}

const (
	btnActual = "🥾 Актуальные хайки"
	btnMy     = "🧾 Мои записи"
	btnHelp   = "ℹ️ Помощь"
)

func (h *Handler) HandleMessage(ctx context.Context, m *tgbot.Message) error {
	switch strings.TrimSpace(m.Text) {
	case btnActual:
		return h.showActual(ctx, m.Chat.ID)
	case btnMy:
		return h.showMyBookings(ctx, m.Chat.ID)
	case btnHelp:
		return h.sendHelp(m.Chat.ID)
	default:
		return h.sendMainMenu(m.Chat.ID)
	}
}

func (h *Handler) HandleCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "book_hike:"):
		h.onCallbackBookHike(ctx, q)
	}
	return nil
}

func (h *Handler) sendMainMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Выберите действие:")
	kb := tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(btnActual),
			tgbot.NewKeyboardButton(btnMy),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(btnHelp),
		),
	)
	kb.ResizeKeyboard = true
	msg.ReplyMarkup = kb
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) sendHelp(chatID int64) error {
	txt := "ℹ️ <b>Помощь</b>\n\n• «🥾 Актуальные хайки» — ближайшие походы.\n• «🧾 Мои записи» — Ваши заявки.\n• Вопросы: @your_support"
	msg := tgbot.NewMessage(chatID, txt)
	msg.ParseMode = "HTML"
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) showActual(ctx context.Context, chatID int64) error {
	rows, err := h.queries.ListActualHikes(ctx, sqlc.ListActualHikesParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(chatID, "Пока нет актуальных хайков."))
		return err
	}

	for _, r := range rows {
		var b strings.Builder

		b.WriteString("🏔 <b>")
		b.WriteString(html.EscapeString(r.TitleRu))
		b.WriteString("</b>\n")

		b.WriteString("📅 ")
		b.WriteString(r.StartsAt.Format("02 January 2006"))
		b.WriteString("\n")

		b.WriteString("📅 ")
		b.WriteString(r.EndsAt.Format("02 January 2006"))
		b.WriteString("\n")

		if r.DescriptionRu != "" {
			b.WriteString("\n")
			b.WriteString(html.EscapeString(r.DescriptionRu))
			b.WriteString("\n")
		}

		kb := tgbot.NewInlineKeyboardMarkup(
			tgbot.NewInlineKeyboardRow(
				tgbot.NewInlineKeyboardButtonData("✅ Забронировать", fmt.Sprintf("book_hike:%d", r.ID)),
			),
		)

		msg := tgbot.NewMessage(chatID, b.String())
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = kb

		if _, err := h.bot.Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handler) showMyBookings(ctx context.Context, chatID int64) error {
	_, err := h.bot.Send(tgbot.NewMessage(chatID, "Здесь появятся Ваши записи на хайки."))
	return err
}

func (h *Handler) onCallbackBookHike(ctx context.Context, q *tgbot.CallbackQuery) {
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

	userID, err := h.queries.UpsertTgUser(ctx, sqlc.UpsertTgUserParams{
		TgUserID:   tgUserID,
		TgUsername: toPgText(username),
		FullName:   toPgText(fullName),
	})
	if err != nil {
		h.log.Errorf("failed to upsert user: %v", err)
		h.replyCallback(q, "Ошибка. Пожалуйста, попробуйте позже.")
		return
	}

	hike, err := h.queries.GetHike(ctx, hikeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.replyCallback(q, "К сожалению, этот хайк недоступен.")
			return
		}
		h.log.Errorf("failed to get hike: %v", err)
		return
	}

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

	h.replyCallback(q, "Ваша заявка отправлена ✅ Мы передали её менеджерам.")

	msg := formatAdminBookingMessage(hike, bookingID, tgUserID, username, fullName)
	adminMsg := tgbot.NewMessage(h.adminChatID, msg)
	adminMsg.ParseMode = "HTML"
	adminMsg.DisableWebPagePreview = true

	if _, err := h.bot.Send(adminMsg); err != nil {
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

func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}