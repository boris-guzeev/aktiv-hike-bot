package handlers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/bot/admin/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	loc         *time.Location
	bot         *tgbot.BotAPI
	queries     *sqlc.Queries
	adminChatID int64
	fsm         *fsm.FSM
}

func New(l *time.Location, b *tgbot.BotAPI, q *sqlc.Queries, acID int64) *Handler {
	return &Handler{
		loc:         l,
		bot:         b,
		queries:     q,
		adminChatID: acID,
		fsm:         fsm.NewFSM(),
	}
}

func (h *Handler) IsAdmin(userID int64) bool {
	m, err := h.bot.GetChatMember(tgbot.GetChatMemberConfig{
		ChatConfigWithUser: tgbot.ChatConfigWithUser{
			ChatID: h.adminChatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false
	}

	return m.Status != "left" && m.Status != "kicked"
}

func (h *Handler) HandleMessage(ctx context.Context, m *tgbot.Message) error {
	// Handle Back button
	if m.Text == backButtonText && h.fsm.State(m.From.ID) != fsm.StateIdle {
		h.fsm.Reset(m.From.ID)

		msg := tgbot.NewMessage(m.Chat.ID, "Создание хайка отменено. Возвращаю в админ-меню.")
		msg.ReplyMarkup = adminMenuKB()
		_, err := h.bot.Send(msg)
		return err
	}

	// FSM step. If is any active states
	if st := h.fsm.State(m.From.ID); st != fsm.StateIdle && !m.IsCommand() {
		return h.handleFSM(ctx, m)
	}

	// Handle Commands
	if m.IsCommand() {
		switch m.Command() {
		case "start":
			return h.sendAdminMenu(m.Chat.ID)

		case "help":
			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Админ-помощь: /newhike — создать хайк, /hikes — список, /admin — меню."))
			return err
		case "hikes":
			// Покажем пока что все. Их пока все равно очень мало
			return h.sendHikesList(m.Chat.ID)
		case "newhike":
			h.fsm.Set(m.From.ID, fsm.StateCreateTitleRU)
			msg := tgbot.NewMessage(m.Chat.ID, "Введите название RU:")
			msg.ReplyMarkup = backKeyboard()
			_, err := h.bot.Send(msg)
			return err
		}
	}

	// Handle Reply Keyboard
	switch m.Text {
	case "➕ Создать хайк":
		h.fsm.Set(m.From.ID, fsm.StateCreateTitleRU)
		msg := tgbot.NewMessage(m.Chat.ID, "Введите название RU:")
		msg.ReplyMarkup = backKeyboard()
		_, err := h.bot.Send(msg)
		return err

	case "📋 Список хайков":
		return h.sendHikesList(m.Chat.ID)

	case "❓ Помощь":
		_, err := h.bot.Send(tgbot.NewMessage(
			m.Chat.ID,
			"Что умеет админ-бот:\n"+
				"• «Создать хайк» — запускает мастера создания\n"+
				"• «Список хайков» — показывает все актуальные\n"+
				"• /newhike, /hikes — те же действия командами",
		))
		return err
	}

	return h.sendAdminMenu(m.Chat.ID)
}

func (h *Handler) sendAdminMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Админ-меню. Выберите действие:")
	msg.ReplyMarkup = adminMenuKB()
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) sendBackBtn(chatID int64, text string) error {
	msg := tgbot.NewMessage(chatID, text)
	msg.ReplyMarkup = backKeyboard()
	_, err := h.bot.Send(msg)
	return err
}

func adminMenuKB() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("➕ Создать хайк"),
			tgbot.NewKeyboardButton("📋 Список хайков"),
		),
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton("❓ Помощь"),
		),
	)
}

const backButtonText = "⬅️ Назад"

func backKeyboard() tgbot.ReplyKeyboardMarkup {
	return tgbot.NewReplyKeyboard(
		tgbot.NewKeyboardButtonRow(
			tgbot.NewKeyboardButton(backButtonText),
		),
	)
}

func (h *Handler) saveCreatedHike(ctx context.Context, userID int64) error {
	data := h.fsm.Data(userID)

	// Parse time
	startAt, err := time.ParseInLocation("02.01.2006 15:04", data["starts_at"], h.loc)
	if err != nil {
		return err
	}
	endsAt, err := time.ParseInLocation("02.01.2006 15:04", data["ends_at"], h.loc)
	if err != nil {
		return err
	}

	// Set params
	args := sqlc.CreateHikeParams{
		TitleRu:       data["title_ru"],
		DescriptionRu: data["description_ru"],
		StartsAt:      startAt,
		EndsAt:        endsAt,
		CreatedAt:     time.Now().In(h.loc),
	}

	// Create Hike
	hike, err := h.queries.CreateHike(ctx, args)
	if err != nil {
		return err
	}

	fmt.Printf("hike: %#v\n", hike)
	return nil
}

// TODO: Вынести парсеры в отдельный файл
func parseHikeDates(input string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	s := strings.TrimSpace(input)
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")
	s = strings.ReplaceAll(s, "  ", " ")

	// Normalize range delimimiters
	// cases: "10 12", "10-12", "10 - 12", "03.02 04.02", "03.02-04.02"
	rangeDelim := regexp.MustCompile(`\s*-\s*|\s+`)
	tokens := rangeDelim.Split(s, -1)

	switch len(tokens) {
	case 1:
		// One value: either "10" or "03.02"
		return parseSingle(tokens[0], now, loc)

	case 2:
		// Range: "10 12", "31 3", "03.02 04.02", "15.12 16.12", "03.02-04.02"
		return parseRange(tokens[0], tokens[1], now, loc)

	default:
		return time.Time{}, time.Time{}, errors.New("не получилось распознать даты (слишком много частей)")
	}
}

var (
	reDay      = regexp.MustCompile(`^\d{1,2}$`)             // 1..31
	reDayMonth = regexp.MustCompile(`(^\d{1,2})\.(\d{1,2})`) // dd.mm
)

func parseSingle(token string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	if reDayMonth.MatchString(token) {
		start, err := parseDDMM(token, now.Year(), loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		today := truncateToDay(now.In(loc))
		if start.Before(today) {
			start, err = parseDDMM(token, now.Year()+1, loc)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
		}
		// один день
		end := time.Date(start.Year(), start.Month(), start.Day(), 22, 0, 0, 0, loc)
		return start, end, nil
	}

	if reDay.MatchString(token) {
		day, _ := strconv.Atoi(token)

		var start time.Time
		// выбираем месяц: если день уже прошёл — берём следующий
		if now.Day() > day {
			start = time.Date(now.Year(), now.Month()+1, day, 8, 0, 0, 0, loc)
		} else {
			start = time.Date(now.Year(), now.Month(), day, 8, 0, 0, 0, loc)
		}

		// однодневный хайк
		end := time.Date(start.Year(), start.Month(), start.Day(), 22, 0, 0, 0, loc)
		return start, end, nil
	}

	return time.Time{}, time.Time{}, errors.New("ожидал формат 'dd' или 'dd.mm'")
}

func parseRange(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	switch {
	case reDay.MatchString(a) && reDay.MatchString(b):
		// both are days without months
		return parseDayDay(a, b, now, loc)
	case reDayMonth.MatchString(a) && reDayMonth.MatchString(b):
		// both are dd.mm
		return parseDDMM_DDMM(a, b, now, loc)
	default:
		return time.Time{}, time.Time{}, errors.New("диапазон должен быть 'dd dd' или 'dd.mm dd.mm'")
	}
}

func parseDayDay(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	dayA, _ := strconv.Atoi(a)
	dayB, _ := strconv.Atoi(b)

	start := time.Date(now.Year(), now.Month(), dayA, 8, 0, 0, 0, loc)

	var end time.Time
	if dayB >= dayA {
		// the same month
		end = time.Date(now.Year(), now.Month(), dayB, 22, 0, 0, 0, loc)
	} else {
		// the next month
		year, month := now.Year(), now.Month()+1
		if month > 12 {
			month = 1
			year++
		}
		end = time.Date(year, month, dayB, 22, 0, 0, 0, loc)
	}
	return start, end, nil
}

func parseDDMM(token string, year int, loc *time.Location) (time.Time, error) {
	m := reDayMonth.FindStringSubmatch(token)
	if len(m) != 3 {
		return time.Time{}, errors.New("ожидал dd.mm")
	}
	day, _ := strconv.Atoi(m[1])
	mon, _ := strconv.Atoi(m[2])

	return time.Date(year, time.Month(mon), day, 8, 0, 0, 0, loc), nil
}

func parseDDMM_DDMM(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	// Try current year
	start, err := parseDDMM(a, now.Year(), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDDMM(b, now.Year(), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// If both dates before Now
	today := truncateToDay(now.In(loc))
	if end.Before(today) && start.Before(today) {
		start, err = parseDDMM(a, now.Year()+1, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err = parseDDMM(b, now.Year()+1, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// If the end before start
	if end.Before(start) {
		// TODO: обмозговать
		end = start.Add(24 * time.Hour)
	}

	end = time.Date(end.Year(), end.Month(), end.Day(), 22, 0, 0, 0, loc)
	return start, end, nil
}

func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func (h *Handler) handleFSM(ctx context.Context, m *tgbot.Message) error {
	switch h.fsm.State(m.From.ID) {

	case fsm.StateCreateTitleRU:
		h.fsm.Put(m.From.ID, "title_ru", m.Text)

		// Пропускаем title_en
		h.fsm.Set(m.From.ID, fsm.StateCreateDescRU)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите описание RU:"))
		return err

	case fsm.StateCreateDescRU:
		h.fsm.Put(m.From.ID, "description_ru", m.Text)

		// Пропускаем description_en
		h.fsm.Set(m.From.ID, fsm.StateCreateDates)
		examples := "Введите даты начала и завершения хайка (примеры: `10`, `10 12`, `10-12`, `31 3`, `03.02-04.02`, `15.12 16.12`)."
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, examples))
		return err

	case fsm.StateCreateDates:
		loc := h.loc
		start, end, err := parseHikeDates(m.Text, time.Now().In(loc), loc)
		if err != nil {
			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не получилось распознать даты. Попробуйте ещё раз.\nПримеры: 10 · 10 12 · 10-12 · 31 3 · 03.02-04.02 · 15.12 16.12"))
			return nil
		}

		h.fsm.Put(m.From.ID, "starts_at", start.Format("02.01.2006 15:04"))
		h.fsm.Put(m.From.ID, "ends_at", end.Format("02.01.2006 15:04"))

		h.fsm.Set(m.From.ID, fsm.StateConfirm)

		preview := fmt.Sprintf(
			"Проверьте данные:\n\nНазвание: %s\nОписание: %s\nДаты: %s → %s\n\nОтправьте 'ok' для сохранения или 'cancel' для отмены.",
			h.fsm.Data(m.From.ID)["title_ru"],
			h.fsm.Data(m.From.ID)["description_ru"],
			format(start), format(end),
		)
		_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, preview))
		return err

	case fsm.StateConfirm:
		txt := strings.TrimSpace(strings.ToLower(m.Text))
		if txt == "cancel" {
			h.fsm.Reset(m.From.ID)
			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Создание отменено. Введите /newhike чтобы начать заново."))
			return err
		}
		if txt != "ok" {
			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Напишите 'ok' для сохранения или 'cancel' для отмены."))
			return nil
		}

		if err := h.saveCreatedHike(ctx, m.From.ID); err != nil {
			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Ошибка при сохранении хайка :("))
			return err
		}

		h.fsm.Reset(m.From.ID)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Хайк создан!"))
		return err

	default:
		h.fsm.Reset(m.From.ID)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Сбросил состояние. Введите /newhike"))
		return err
	}
}

// func (h *Handler) handleFSM(ctx context.Context, m *tgbot.Message) error {
// 	switch h.fsm.State(m.From.ID) {
// 	case StateCreateTitleRU:
// 		// Save current state
// 		h.fsm.Put(m.From.ID, "title_ru", m.Text)

// 		// Set next state
// 		h.fsm.Set(m.From.ID, StateCreateTitleEN)
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите название EN:"))
// 		return err

// 	case StateCreateTitleEN:
// 		// Save current state
// 		h.fsm.Put(m.From.ID, "title_en", m.Text)

// 		// Set the next state
// 		h.fsm.Set(m.From.ID, StateCreateDescRU)
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите описание RU:"))
// 		return err

// 	case StateCreateDescRU:
// 		// Save current state
// 		h.fsm.Put(m.From.ID, "description_ru", m.Text)

// 		// Set the next state
// 		h.fsm.Set(m.From.ID, StateCreateDescEN)
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите описание EN:"))
// 		return err

// 	case StateCreateDescEN:
// 		h.fsm.Put(m.From.ID, "description_en", m.Text)
// 		h.fsm.Set(m.From.ID, StateCreateDates)
// 		examples := "Введите даты начала и завершения хайка (примеры: `10`, `10 12`, `10-12`, `31 3`, `03.02-04.02`, `15.12 16.12`)."
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, examples))
// 		return err

// 	case StateCreateDates:
// 		loc := h.loc
// 		start, end, err := parseHikeDates(m.Text, time.Now(), loc)
// 		if err != nil {
// 			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не получилось распознать даты. Попробуйте ещё раз.\nПримеры: 10 · 10 12 · 10-12 · 31 3 · 03.02-04.02 · 15.12 16.12"))
// 			return nil
// 		}
// 		h.fsm.Put(m.From.ID, "starts_at", start.Format("02.01.2006 15:04"))
// 		h.fsm.Put(m.From.ID, "ends_at", end.Format("02.01.2006 15:04"))

// 		// Proceed to Confirm
// 		h.fsm.Set(m.From.ID, StateConfirm)
// 		// TODO: Добавить метод h.fsm.Get(m.From.ID, "title_en")
// 		preview := fmt.Sprintf(
// 			"Проверьте данные:\n\nНазвание: %s / %s\nОписание: %s / %s\nДаты: %s → %s\n\nОтправьте 'ok' для сохранения или 'cancel' для отмены.",
// 			h.fsm.Data(m.From.ID)["title_ru"],
// 			h.fsm.Data(m.From.ID)["title_en"],
// 			h.fsm.Data(m.From.ID)["description_ru"],
// 			h.fsm.Data(m.From.ID)["description_en"],
// 			format(start), format(end),
// 		)
// 		_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, preview))
// 		return err

// 	case StateConfirm:
// 		txt := strings.TrimSpace(strings.ToLower(m.Text))
// 		if txt == "cancel" {
// 			h.fsm.Reset(m.From.ID)
// 			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Создание отменено. Введите /newhike чтобы начать заново."))
// 			return err
// 		}
// 		if txt != "ok" {
// 			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Напишите 'ok' для сохранения или 'cancel' для отмены."))
// 			return nil
// 		}
// 		// Saving Hike
// 		if err := h.saveCreatedHike(ctx, m.From.ID); err != nil {
// 			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Ошибка при сохранении хайка :("))
// 			return err
// 		}
// 		h.fsm.Reset(m.From.ID)
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Хайк создан!"))
// 		return err

// 	default:
// 		h.fsm.Reset(m.From.ID)
// 		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Сбросил состояние. Введите /newhike"))
// 		return err
// 	}
// }

// func (h *Handler) sendAdminMenu(chatID int64) error {
// 	msg := tgbot.NewMessage(chatID, "Админ-меню:")
// 	msg.ReplyMarkup = h.adminMenuKB()
// 	_, err := h.bot.Send(msg)
// 	return err
// }

// func (h *Handler) adminMenuKB() tgbot.InlineKeyboardMarkup {
// 	return tgbot.NewInlineKeyboardMarkup(
// 		tgbot.NewInlineKeyboardRow(
// 			btn("➕ Новый хайк", "a:new"),
// 			btn("📋 Актуальные", "a:list:actual:1"),
// 		),
// 		tgbot.NewInlineKeyboardRow(
// 			btn("🗂 Все хайки", "a:list:all:1"),
// 		),
// 	)
// }

func (h *Handler) showActual(ctx context.Context, q *tgbot.CallbackQuery) error {
	rows, err := h.queries.ListActualHikes(ctx, sqlc.ListActualHikesParams{Limit: 20, Offset: 0})
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("Актуальные хайки:\n\n")
	for _, r := range rows {
		b.WriteString(
			fmt.Sprintf("• #%d %s • %s → %s\n",
				r.ID, r.TitleRu, format(r.StartsAt), format(r.EndsAt),
			),
		)
	}
	return h.editText(q, b.String(), tgbot.NewInlineKeyboardMarkup())
}

func (h *Handler) showCard(ctx context.Context, q *tgbot.CallbackQuery, id int32) error {
	hike, err := h.queries.GetHikeByID(ctx, id)
	if err != nil {
		return err
	}
	pub := "🚫 Нет"
	next := "1"
	nextTxt := "📤 Опубликовать"

	if hike.IsPublished {
		pub = "✅ Да"
		next = "0"
		nextTxt = "🚫 Снять с публикации"
	}
	text := fmt.Sprintf(
		"🏔 %s\n%s → %s\nОпубликован: %s",
		hike.TitleRu,
		format(hike.StartsAt),
		format(hike.EndsAt),
		pub,
	)
	kb := tgbot.NewInlineKeyboardMarkup(
		tgbot.NewInlineKeyboardRow(
			btn("✏️ Редактировать", fmt.Sprintf("a:edit:%d", id)),
		),
		tgbot.NewInlineKeyboardRow(
			btn(nextTxt, fmt.Sprintf("a:pub:%d:%s", id, next)),
		),
		tgbot.NewInlineKeyboardRow(
			btn("⬅️ Назад", "a:menu"),
		),
	)
	return h.editText(q, text, kb)
}

func (h *Handler) editText(q *tgbot.CallbackQuery, text string, kb tgbot.InlineKeyboardMarkup) error {
	_, _ = h.bot.Request(tgbot.NewCallback(q.ID, ""))
	edit := tgbot.NewEditMessageTextAndMarkup(q.Message.Chat.ID, q.Message.MessageID, text, kb)
	_, err := h.bot.Send(edit)
	return err
}

func format(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}

func btn(text, data string) tgbot.InlineKeyboardButton {
	b := tgbot.NewInlineKeyboardButtonData(text, data)
	return b
}
