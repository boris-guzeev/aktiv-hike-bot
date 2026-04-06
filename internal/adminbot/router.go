package adminbot

import (
	"context"
	"strings"

	bookingHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/handler"
	hikeHandler "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/handler"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/common"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type router struct {
	bot            *tgbot.BotAPI
	adminChatID    int64
	hikeHandler    *hikeHandler.HikeHandler
	bookingHandler *bookingHandler.BookingHandler
}

func NewRouter(b *tgbot.BotAPI, acID int64, hH *hikeHandler.HikeHandler, bH *bookingHandler.BookingHandler) *router {
	return &router{
		bot:            b,
		adminChatID:    acID,
		hikeHandler:    hH,
		bookingHandler: bH,
	}
}

func (r *router) Route(ctx context.Context, u tgbot.Update) error {
	// Private messages
	if m := u.Message; m != nil && m.Chat.IsPrivate() {
		if r.isAdmin(m.From.ID) {
			return r.routeMessage(ctx, m)
		}
	}

	// Private callbacks
	if q := u.CallbackQuery; q != nil && q.Message != nil && q.Message.Chat.IsPrivate() {
		if r.isAdmin(q.From.ID) {
			return r.routeCallback(ctx, q)
		}
	}

	return nil
}

func (r *router) routeMessage(ctx context.Context, m *tgbot.Message) error {
	if m.Text == "⬅️ Назад" {
		r.hikeHandler.ResetFSM(m.From.ID)
		return r.showMainMenu(m.Chat.ID)
	}

	if r.hikeHandler.InProgressFSM(m.From.ID) {
		return r.hikeHandler.HandleFSM(ctx, m)
	}

	switch m.Text {
	case "🏔 Хайки", "➕ Создать хайк", "📋 Список хайков":
		return r.routeHikeMessage(ctx, m)

	case "📥 Заявки", "📋 Список заявок", "📊 Статистика заявок":
		return r.routeBookingMessage(ctx, m)

	case "❓ Помощь":
		return r.showHelp(m.Chat.ID)
	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) routeHikeMessage(ctx context.Context, m *tgbot.Message) error {
	switch m.Text {
	case "🏔 Хайки":
		return r.hikeHandler.ShowMenu(ctx, m)
	case "➕ Создать хайк":
		return r.hikeHandler.StartCreateHike(ctx, m)
	case "📋 Список хайков":
		return r.hikeHandler.ListHikes(ctx, m)
	}

	return r.showMainMenu(m.Chat.ID)
}

func (r *router) routeBookingMessage(ctx context.Context, m *tgbot.Message) error {
	switch m.Text {
	case "📥 Заявки":
		return r.bookingHandler.ShowMenu(ctx, m)
	case "📋 Список заявок":
		return r.bookingHandler.ListBookings(ctx, m)
	case "📊 Статистика заявок":
		// TODO: return r.bookingHandler.Stat(ctx, m)
		return nil
	}

	return r.showMainMenu(m.Chat.ID)
}

// TODO: убрать из роутера
func (r *router) showHelp(chatID int64) error {
	text := `❓ <b>Помощь для администратора</b>

━━━━━━━━━━━━━━━
🏔 <b>Как создать хайк</b>

1️⃣ Откройте раздел <b>🏔 Хайки</b>  
2️⃣ Нажмите <b>➕ Создать хайк</b>  
3️⃣ Заполните поля:
• Название  
• Описание  
• Даты  
• Цена  
• Дистанция  
• Набор высоты  
• Фото  

4️⃣ Проверьте данные  
5️⃣ Нажмите <b>✅ Подтвердить</b>

После этого хайк появится в клиентском боте

━━━━━━━━━━━━━━━
📋 <b>Работа с хайками</b>

📋 Список хайков — посмотреть все хайки  
Вы можете:
• Опубликовать хайк  
• Скрыть хайк  
• Редактировать позже (в будущем)

━━━━━━━━━━━━━━━
📥 <b>Работа с заявками</b>

Когда клиент бронирует хайк:
• В админ-чате появляется заявка  
• Любой менеджер может взять её в работу  

Статусы заявок:
🟡 В работе — менеджер взял заявку  
🟢 Подтверждена — клиент подтвердил участие  
🏁 Завершена — хайк состоялся  
🔴 Отменена — заявка отменена  

━━━━━━━━━━━━━━━
📋 <b>Как работать с заявкой</b>

1️⃣ Откройте <b>📋 Список заявок</b>  
2️⃣ Выберите заявку  
3️⃣ Нажмите нужное действие:
• ✅ Подтвердить  
• ❌ Отменить  
• 🏁 Завершить  

━━━━━━━━━━━━━━━
💡 <b>Важно</b>

• Новые заявки приходят автоматически  
• Один менеджер — одна заявка  
• После взятия заявки другие менеджеры её не обрабатывают  

Если возникли проблемы — напишите разработчику 😄`

	msg := tgbot.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	_, err := r.bot.Send(msg)
	return err
}

func (r *router) showMainMenu(chatID int64) error {
	msg := tgbot.NewMessage(chatID, "Выберите раздел")
	msg.ReplyMarkup = common.MainMenu()

	_, err := r.bot.Send(msg)
	return err
}

func (r *router) routeCallback(ctx context.Context, q *tgbot.CallbackQuery) error {
	switch {
	case strings.HasPrefix(q.Data, "hike:"):
		return nil
		// TODO: return r.hikeHandler.HandleCallback(ctx, q)

	case strings.HasPrefix(q.Data, "booking:"):
		return r.bookingHandler.HandleCallback(ctx, q)
	}

	return nil
}

func (r *router) isAdmin(userID int64) bool {
	m, err := r.bot.GetChatMember(tgbot.GetChatMemberConfig{
		ChatConfigWithUser: tgbot.ChatConfigWithUser{
			ChatID: r.adminChatID,
			UserID: userID,
		},
	})
	if err != nil {
		return false
	}

	return m.Status != "left" && m.Status != "kicked"
}
