package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/hike"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/parser"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *HikeHandler) InProgress(userID int64) bool {
	return h.fsm.State(userID) != fsm.StateIdle
}

func (h *HikeHandler) ShowMenu(ctx context.Context, m *tgbot.Message) error {
	msgConfig := tgbot.NewMessage(m.Chat.ID, "Раздел хайков")
	msgConfig.ReplyMarkup = hikeUI.HikeMenu()

	_, err := h.bot.Send(msgConfig)
	if err != nil {
		return err
	}
	return nil
}

func (h *HikeHandler) StartCreateHike(ctx context.Context, m *tgbot.Message) error {
	h.fsm.Reset(m.From.ID)
	h.fsm.Set(m.From.ID, fsm.StateCreateTitleRU)
	_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите название RU:"))
	return err
}

func (h *HikeHandler) HandleCreateHike(ctx context.Context, m *tgbot.Message) error {
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
		examples := "Введите даты начала и завершения хайка (примеры: 10, 10 12, 10-12, 31 3, 03.02-04.02, 15.12 16.12)."
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, examples))
		return err

	case fsm.StateCreateDates:
		loc := h.loc
		start, end, err := parser.ParseHikeDates(m.Text, time.Now().In(loc), loc)
		if err != nil {
			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не получилось распознать даты. Попробуйте ещё раз.\nПримеры: 10 · 10 12 · 10-12 · 31 3 · 03.02-04.02 · 15.12 16.12"))
			return nil
		}

		h.fsm.Put(m.From.ID, "starts_at", start.Format("02.01.2006 15:04"))
		h.fsm.Put(m.From.ID, "ends_at", end.Format("02.01.2006 15:04"))

		h.fsm.Set(m.From.ID, fsm.StateCreatePhoto)

		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Загрузите фото:"))
		return err

	case fsm.StateCreatePhoto:
		if len(m.Photo) == 0 {
			_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Пожалуйста, отправьте именно фото."))
			return nil
		}
		photo := m.Photo[len(m.Photo)-1]
		h.fsm.Put(m.From.ID, "photo_file_id", photo.FileID)

		h.fsm.Set(m.From.ID, fsm.StateConfirm)

		preview := fmt.Sprintf(
			"Проверьте данные:\n\nНазвание: %s\nОписание: %s\nДаты: %s → %s\nФото: добавлено\n\nОтправьте 'ok' для сохранения или 'cancel' для отмены или нажмите Подтвердить/Отмена на клавиатуре.",
			h.fsm.Data(m.From.ID)["title_ru"],
			h.fsm.Data(m.From.ID)["description_ru"],
			h.fsm.Data(m.From.ID)["starts_at"],
			h.fsm.Data(m.From.ID)["ends_at"],
		)

		msg := tgbot.NewMessage(m.Chat.ID, preview)
		msg.ReplyMarkup = hikeUI.ConfirmKeyboard()

		_, err := h.bot.Send(msg)
		return err

	case fsm.StateConfirm:
		txt := strings.TrimSpace(strings.ToLower(m.Text))
		if txt == "cancel" {
			h.fsm.Reset(m.From.ID)
			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Создание отменено."))
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
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Сбросил состояние."))
		return err
	}
}

func (h *HikeHandler) saveCreatedHike(ctx context.Context, userID int64) error {
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
	hike := service.Hike{
		TitleRu:       data["title_ru"],
		DescriptionRu: data["description_ru"],
		StartsAt:      startAt,
		EndsAt:        endsAt,
		PhotoFileID:   data["photo_file_id"],
	}

	// Create Hike
	createdHikeID, err := h.service.CreateHike(ctx, hike)
	if err != nil {
		return err
	}
	imagePath, err := h.saveImage(ctx, data["photo_file_id"], createdHikeID)
	if err != nil {
		return err
	}

	// Save image path
	if err := h.service.UpdateImagePath(ctx, createdHikeID, imagePath); err != nil {
		return err
	}

	return nil
}

func (h *HikeHandler) saveImage(ctx context.Context, fileID string, hikeID int32) (string, error) {
	file, err := h.bot.GetFile(tgbot.FileConfig{FileID: fileID})
	if err != nil {
		return "", err
	}

	url := file.Link(h.bot.Token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected http-status code: %d", resp.StatusCode)
	}

	dir := filepath.Join(h.storageRoot, "hikes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.jpg", hikeID))

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("hikes/%d.jpg", hikeID), nil
}

func (h *HikeHandler) ListHikes(ctx context.Context, m *tgbot.Message) error {
	hikes, err := h.service.ListHikes(ctx, 1, 20)
	if err != nil {
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не удалось загрузить список хайков."))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	if len(hikes) == 0 {
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Список хайков пуст."))
		return err
	}

	var b strings.Builder
	b.WriteString("🏔 Список хайков\n\n")

	for _, hike := range hikes {
		status := "📝"
		if hike.IsPublished {
			status = "✅"
		}

		line := fmt.Sprintf(
			"#%d · %s · %s · %s\n\n",
			hike.ID,
			hike.TitleRu,
			hike.StartsAt.In(h.loc).Format("02.01.2006"),
			status,
		)
		b.WriteString(line)
	}
	b.WriteString("\nОтправьте ID хайка, чтобы открыть карточку.")

	_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, b.String()))
	return err
}
