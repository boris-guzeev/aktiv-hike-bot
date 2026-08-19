package relay

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/config"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	defaultSubject  = "общее обращение"
	maxCaptionRunes = 1024
	maxTextRunes    = 4096
)

var telegramIDPattern = regexp.MustCompile(`(?m)Telegram ID:\s*(\d+)`)

type Handler struct {
	bot *tgbot.BotAPI
	cfg config.ClientBot
}

func New(bot *tgbot.BotAPI, cfg config.ClientBot) *Handler {
	return &Handler{bot: bot, cfg: cfg}
}

// HandleAdminReply relays a reply from the configured admin group to the
// client encoded in the replied-to Telegram message.
func (h *Handler) HandleAdminReply(message *tgbot.Message) (bool, error) {
	if message == nil || message.Chat == nil || message.Chat.ID != h.cfg.AdminChatID ||
		message.ReplyToMessage == nil || !h.isOwnMessage(message.ReplyToMessage) {
		return false, nil
	}

	clientID, ok := clientIDFromMessage(message.ReplyToMessage)
	if !ok {
		return false, nil
	}

	header := "💬 <b>Ответ от AktivHike</b>"

	return true, h.sendRelayedMessage(clientID, message, 0, header)
}

// HandleClientMessage relays any private client message to the admin group.
// A reply carries the subject from the manager's message; a standalone
// message is treated as a general enquiry.
func (h *Handler) HandleClientMessage(message *tgbot.Message) (bool, error) {
	if message == nil || message.Chat == nil || !message.Chat.IsPrivate() ||
		message.From == nil || message.From.IsBot {
		return false, nil
	}

	subject := defaultSubject
	if message.ReplyToMessage != nil && h.isOwnMessage(message.ReplyToMessage) {
		subject = subjectFromMessage(message.ReplyToMessage)
	}
	name := strings.TrimSpace(message.From.FirstName + " " + message.From.LastName)
	if name == "" {
		name = "Клиент"
	}

	username := ""
	if message.From.UserName != "" {
		username = " (@" + html.EscapeString(message.From.UserName) + ")"
	}

	header := fmt.Sprintf(
		"💬 <b>Сообщение от клиента</b>\n👤 <a href=\"tg://user?id=%d\">%s</a>%s\n🥾 Заявка: %s",
		message.From.ID,
		html.EscapeString(name),
		username,
		html.EscapeString(subject),
	)

	return true, h.sendRelayedMessage(h.cfg.AdminChatID, message, 0, header)
}

func (h *Handler) sendRelayedMessage(chatID int64, source *tgbot.Message, replyTo int, header string) error {
	content := strings.TrimSpace(messageText(source))
	if source.Text != "" {
		out := tgbot.NewMessage(chatID, formattedContent(header, content, maxTextRunes))
		out.ParseMode = tgbot.ModeHTML
		out.ReplyToMessageID = replyTo
		out.AllowSendingWithoutReply = true
		_, err := h.bot.Send(out)
		return err
	}

	out := tgbot.NewCopyMessage(chatID, source.Chat.ID, source.MessageID)
	out.Caption = formattedContent(header, content, maxCaptionRunes)
	out.ParseMode = tgbot.ModeHTML
	out.ReplyToMessageID = replyTo
	out.AllowSendingWithoutReply = true
	_, err := h.bot.CopyMessage(out)
	return err
}

func (h *Handler) isOwnMessage(message *tgbot.Message) bool {
	return message.From != nil && message.From.IsBot && message.From.ID == h.bot.Self.ID
}

func clientIDFromMessage(message *tgbot.Message) (int64, bool) {
	entities := message.Entities
	if message.Text == "" {
		entities = message.CaptionEntities
	}

	for _, entity := range entities {
		if entity.User != nil {
			return entity.User.ID, true
		}

		if entity.URL == "" {
			continue
		}
		parsed, err := url.Parse(entity.URL)
		if err != nil || parsed.Scheme != "tg" || parsed.Host != "user" {
			continue
		}

		id, err := strconv.ParseInt(parsed.Query().Get("id"), 10, 64)
		if err == nil && id > 0 {
			return id, true
		}
	}

	match := telegramIDPattern.FindStringSubmatch(messageText(message))
	if len(match) == 2 {
		id, err := strconv.ParseInt(match[1], 10, 64)
		if err == nil && id > 0 {
			return id, true
		}
	}

	return 0, false
}

func subjectFromMessage(message *tgbot.Message) string {
	for _, line := range strings.Split(messageText(message), "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"📍 Хайк:", "🥾 Заявка:", "💬 Ответ на"} {
			if strings.HasPrefix(line, prefix) {
				subject := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if subject != "" {
					return subject
				}
			}
		}
	}
	return defaultSubject
}

func messageText(message *tgbot.Message) string {
	if message == nil {
		return ""
	}
	if message.Text != "" {
		return message.Text
	}
	return message.Caption
}

func formattedContent(header, content string, limit int) string {
	if content == "" {
		return header
	}

	// Truncate before HTML escaping so an entity such as &amp; is never cut in
	// half. Counting the markup in the header is conservative but valid.
	available := limit - utf8.RuneCountInString(header) - 2
	if available <= 0 {
		return header
	}
	if utf8.RuneCountInString(content) > available {
		runes := []rune(content)
		content = string(runes[:available-1]) + "…"
	}
	return header + "\n\n" + html.EscapeString(content)
}
