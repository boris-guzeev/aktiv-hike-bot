package notify

import tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Notifier interface {
	Text(text string) error
}

type notifier struct {
	bot *tgbot.BotAPI
	adminChatID int64
}

func New(b *tgbot.BotAPI, acID int64) Notifier {
	return &notifier{
		bot: b,
		adminChatID: acID,
	}
}

func (n *notifier) Text(text string) error {
	msg := tgbot.NewMessage(n.adminChatID, text)
	_, err := n.bot.Send(msg)
	return err
}