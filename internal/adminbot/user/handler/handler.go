package handler

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/common"
	userUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/user"
	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type state uint8

const (
	stateIdle state = iota
	stateSelectUser
	stateViewUser
	stateEditAchievements
	stateEditAchievementValue
)

const seasonsAchievementName = "SEASONS"

const maxSeasons = 20

type session struct {
	state         state
	userIDs       []int32
	userID        int32
	achievementID int16
}

type Handler struct {
	bot      *tgbot.BotAPI
	service  userService.Service
	mu       sync.Mutex
	sessions map[int64]session
}

func New(bot *tgbot.BotAPI, service userService.Service) *Handler {
	return &Handler{bot: bot, service: service, sessions: make(map[int64]session)}
}

func (h *Handler) InProgress(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sessions[userID].state != stateIdle
}

func (h *Handler) Reset(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.sessions, userID)
}

func (h *Handler) get(userID int64) session {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sessions[userID]
}

func (h *Handler) set(userID int64, value session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[userID] = value
}

func (h *Handler) ListUsers(ctx context.Context, m *tgbot.Message) error {
	users, err := h.service.ListTelegramUsers(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Список пользователей пуст."))
		return logger.WrapError(err)
	}

	ids := make([]int32, 0, len(users))
	var b strings.Builder
	b.WriteString("👥 <b>Пользователи</b>\n\n")
	for i, user := range users {
		ids = append(ids, user.ID)
		username := "—"
		if strings.TrimSpace(user.TgUsername) != "" {
			username = "@" + strings.TrimPrefix(user.TgUsername, "@")
		}
		name := strings.TrimSpace(user.FullName)
		if name == "" {
			name = "—"
		}
		fmt.Fprintf(&b, "%d. %s · %s\n", i+1, html.EscapeString(username), html.EscapeString(name))
	}
	b.WriteString("\nОтправьте номер пользователя из списка.")
	h.set(m.From.ID, session{state: stateSelectUser, userIDs: ids})
	msg := tgbot.NewMessage(m.Chat.ID, b.String())
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = userUI.BackKeyboard()
	_, err = h.bot.Send(msg)
	return logger.WrapError(err)
}

func (h *Handler) Handle(ctx context.Context, m *tgbot.Message) error {
	s := h.get(m.From.ID)
	switch s.state {
	case stateSelectUser:
		if m.Text == userUI.ButtonBack {
			h.Reset(m.From.ID)
			msg := tgbot.NewMessage(m.Chat.ID, "Выберите раздел")
			msg.ReplyMarkup = common.MainMenu()
			_, err := h.bot.Send(msg)
			return logger.WrapError(err)
		}
		n, err := strconv.Atoi(strings.TrimSpace(m.Text))
		if err != nil || n < 1 || n > len(s.userIDs) {
			_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите номер пользователя из списка."))
			return logger.WrapError(err)
		}
		s.state, s.userID = stateViewUser, s.userIDs[n-1]
		h.set(m.From.ID, s)
		return h.showCard(ctx, m.Chat.ID, s.userID)
	case stateViewUser:
		switch m.Text {
		case userUI.ButtonEditAchievements:
			s.state = stateEditAchievements
			h.set(m.From.ID, s)
			return h.showAchievements(ctx, m.Chat.ID, s.userID)
		case userUI.ButtonBack:
			return h.ListUsers(ctx, m)
		default:
			return h.showCard(ctx, m.Chat.ID, s.userID)
		}
	case stateEditAchievements:
		if m.Text == userUI.ButtonBack {
			s.state = stateViewUser
			h.set(m.From.ID, s)
			return h.showCard(ctx, m.Chat.ID, s.userID)
		}
		id, err := strconv.ParseInt(strings.TrimSpace(m.Text), 10, 16)
		achievements, listErr := h.service.ListUserAchievements(ctx, s.userID)
		if listErr != nil {
			return listErr
		}
		var selected *userService.Achievement
		if err == nil {
			for i := range achievements {
				if achievements[i].ID == int16(id) {
					selected = &achievements[i]
					break
				}
			}
		}
		if selected == nil {
			_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите ID достижения из списка."))
			return logger.WrapError(err)
		}
		if selected.HasValue {
			s.state = stateEditAchievementValue
			s.achievementID = selected.ID
			h.set(m.From.ID, s)
			current := ""
			if selected.Value != nil {
				current = fmt.Sprintf(" Текущее значение: %d.", *selected.Value)
			}
			msg := tgbot.NewMessage(m.Chat.ID, "Введите количество сезонов (0 — убрать достижение)."+current)
			msg.ReplyMarkup = userUI.BackKeyboard()
			_, err = h.bot.Send(msg)
			return logger.WrapError(err)
		}
		if _, err := h.service.ToggleUserAchievement(ctx, s.userID, selected.ID); err != nil {
			return err
		}
		return h.showAchievements(ctx, m.Chat.ID, s.userID)
	case stateEditAchievementValue:
		if m.Text == userUI.ButtonBack {
			s.state = stateEditAchievements
			s.achievementID = 0
			h.set(m.From.ID, s)
			return h.showAchievements(ctx, m.Chat.ID, s.userID)
		}
		value, err := strconv.ParseInt(strings.TrimSpace(m.Text), 10, 32)
		if err != nil || value < 0 || value > maxSeasons {
			_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите целое число от 0 до 20."))
			return logger.WrapError(err)
		}
		if value == 0 {
			err = h.service.RemoveUserAchievement(ctx, s.userID, s.achievementID)
		} else {
			err = h.service.SetUserAchievementValue(ctx, s.userID, s.achievementID, int32(value))
		}
		if err != nil {
			return err
		}
		s.state = stateEditAchievements
		s.achievementID = 0
		h.set(m.From.ID, s)
		return h.showAchievements(ctx, m.Chat.ID, s.userID)
	default:
		return h.ListUsers(ctx, m)
	}
}

func (h *Handler) showCard(ctx context.Context, chatID int64, userID int32) error {
	user, err := h.service.GetTelegramUser(ctx, userID)
	if err != nil {
		return err
	}
	achievements, err := h.service.ListUserAchievements(ctx, userID)
	if err != nil {
		return err
	}
	username := "—"
	if strings.TrimSpace(user.TgUsername) != "" {
		username = "@" + strings.TrimPrefix(user.TgUsername, "@")
	}
	name := strings.TrimSpace(user.FullName)
	if name == "" {
		name = "—"
	}
	var b strings.Builder
	b.WriteString("👤 <b>Карточка пользователя</b>\n\n")
	fmt.Fprintf(&b, "<b>Telegram:</b> %s\n<b>Полное имя:</b> %s\n\n<b>Достижения:</b>\n", html.EscapeString(username), html.EscapeString(name))
	count := 0
	for _, a := range achievements {
		if a.Assigned {
			fmt.Fprintf(&b, "🏅 <b>%s</b> — %s\n", html.EscapeString(achievementDisplayName(a)), html.EscapeString(a.Description))
			count++
		}
	}
	if count == 0 {
		b.WriteString("Пока нет достижений.\n")
	}
	msg := tgbot.NewMessage(chatID, b.String())
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = userUI.UserCardKeyboard()
	_, err = h.bot.Send(msg)
	return logger.WrapError(err)
}

func (h *Handler) showAchievements(ctx context.Context, chatID int64, userID int32) error {
	items, err := h.service.ListUserAchievements(ctx, userID)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("🏅 <b>Редактирование достижений</b>\n\n")
	for _, a := range items {
		mark := "⬜"
		if a.Assigned {
			mark = "✅"
		}
		name := achievementDisplayName(a)
		fmt.Fprintf(
			&b,
			"ID %d · %s <b>%s</b>\n%s\n\n",
			a.ID,
			mark,
			html.EscapeString(name),
			html.EscapeString(a.Description),
		)
	}
	b.WriteString("Отправьте ID, чтобы изменить достижение.")
	msg := tgbot.NewMessage(chatID, b.String())
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = userUI.BackKeyboard()
	_, err = h.bot.Send(msg)
	return logger.WrapError(err)
}

func achievementDisplayName(a userService.Achievement) string {
	if a.Name != seasonsAchievementName || a.Value == nil || *a.Value <= 0 {
		return a.Name
	}
	season := "SEASONS"
	if *a.Value == 1 {
		season = "SEASON"
	}
	return fmt.Sprintf("%s %s", toRoman(*a.Value), season)
}

func toRoman(value int32) string {
	if value <= 0 {
		return ""
	}
	values := [...]int32{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	symbols := [...]string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var b strings.Builder
	for i, n := range values {
		for value >= n {
			b.WriteString(symbols[i])
			value -= n
		}
	}
	return b.String()
}
