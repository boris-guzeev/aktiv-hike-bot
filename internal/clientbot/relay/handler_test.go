package relay

import (
	"testing"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestClientIDFromMessage(t *testing.T) {
	message := &tgbot.Message{
		Text: "Клиент: Boris",
		Entities: []tgbot.MessageEntity{
			{Type: "text_link", URL: "tg://user?id=123456789"},
		},
	}

	got, ok := clientIDFromMessage(message)
	if !ok || got != 123456789 {
		t.Fatalf("clientIDFromMessage() = %d, %v", got, ok)
	}
}

func TestClientIDFromEditedBookingMessage(t *testing.T) {
	message := &tgbot.Message{Text: "👤 Пользователь: @boris\n🆔 Telegram ID: 987654321"}

	got, ok := clientIDFromMessage(message)
	if !ok || got != 987654321 {
		t.Fatalf("clientIDFromMessage() = %d, %v", got, ok)
	}
}

func TestSubjectFromMessages(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"📍 Хайк: Тушети\nДата: завтра", "Тушети"},
		{"💬 Сообщение\n🥾 Заявка: Чирокки → Муха", "Чирокки → Муха"},
		{"💬 Ответ на Хайкинг-тур Тушети", "Хайкинг-тур Тушети"},
		{"Без контекста", defaultSubject},
	}

	for _, test := range tests {
		if got := subjectFromMessage(&tgbot.Message{Text: test.text}); got != test.want {
			t.Errorf("subjectFromMessage(%q) = %q, want %q", test.text, got, test.want)
		}
	}
}
