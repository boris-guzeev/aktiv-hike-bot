package admin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	hikesPageSize = 10
	descMaxLen    = 140 // сколько символов показывать в списке
)

// sendHikesList показывает последние hikesPageSize хайков без фильтров/пагинации.
func (h *Handler) sendHikesList(chatID int64) error {
	ctx := context.Background()

	args := sqlc.ListHikesParams{
		Limit:  int32(hikesPageSize),
		Offset: 0,
	}

	rows, err := h.queries.ListHikes(ctx, args)
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(chatID, "⚠️ Не удалось получить список хайков"))
		return err
	}

	if len(rows) == 0 {
		_, err := h.bot.Send(tgbot.NewMessage(chatID, "Пока нет ни одного хайка. Нажми /newhike, чтобы создать."))
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "🌄 <b>Хайки</b>\n\n")

	for i, r := range rows {
		pub := "🕗 Черновик"
		if r.IsPublished {
			pub = "✅ Опубликован"
		}

		// Заголовок
		title := strings.TrimSpace(r.TitleRu)
		if title == "" && r.TitleEn.Valid {
			title = strings.TrimSpace(r.TitleEn.String)
		}
		if title == "" {
			title = "(без названия)"
		}

		// Краткое описание
		desc := normalizeOneLine(strings.TrimSpace(r.DescriptionRu))
		if desc == "" {
			desc = "(описание не заполнено)"
		}
		desc = truncateRunes(desc, descMaxLen)

		// Даты
		start := r.StartsAt.In(time.Local).Format("02.01 15:04")
		end := r.EndsAt.In(time.Local).Format("02.01 15:04")
		created := r.CreatedAt.In(time.Local).Format("02.01 15:04")

		// Красивый блок для каждого хайка
		fmt.Fprintf(&b, "%d. <b>%s</b>\n", i+1, title)
		fmt.Fprintf(&b, "📝 %s\n", desc)
		fmt.Fprintf(&b, "📅 %s — %s\n", start, end)
		fmt.Fprintf(&b, "📤 %s   •   🕰 создано %s\n\n", pub, created)
	}

	msg := tgbot.NewMessage(chatID, b.String())
	msg.ParseMode = "HTML"
	_, err = h.bot.Send(msg)
	return err
}

// truncateRunes обрезает по количеству рун и добавляет многоточие при обрезке.
func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit-1]) + "…"
}

// normalizeOneLine заменяет переводы строк/табуляции/двойные пробелы на один пробел.
func normalizeOneLine(s string) string {
	spacey := regexp.MustCompile(`[\r\n\t]+`)
	multi := regexp.MustCompile(`\s{2,}`)
	out := spacey.ReplaceAllString(s, " ")
	out = multi.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}