package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/parser"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/hike"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *HikeHandler) InProgressFSM(userID int64) bool {
	return h.fsm.State(userID) != fsm.StateIdle
}

func (h *HikeHandler) ResetFSM(userID int64) {
	h.fsm.Reset(userID)
}

func (h *HikeHandler) HandleFSM(ctx context.Context, m *tgbot.Message) error {
	switch h.fsm.State(m.From.ID) {
	case fsm.StateCreateTitleRU,
		fsm.StateCreateDescRU,
		fsm.StateCreatePrice,
		fsm.StateCreateDistanceKm,
		fsm.StateCreateElevationGain,
		fsm.StateCreateDates,
		fsm.StateCreatePhoto,
		fsm.StateConfirm:
		return h.HandleCreateHike(ctx, m)

	case fsm.StateSelectHikeID:
		return h.HandleSelectHike(ctx, m)

	case fsm.StateSelectedHikeAction, fsm.StateConfirmPublishHike, fsm.StateConfirmHideHike:
		return h.HandlePublishHike(ctx, m)

	default:
		h.fsm.Reset(m.From.ID)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Состояние сброшено."))
		return err
	}
}

func (h *HikeHandler) sendCreateStep(chatID int64, text string) error {
	msg := tgbot.NewMessage(chatID, text)
	msg.ReplyMarkup = hikeUI.CreateHikeKeyboard()

	_, err := h.bot.Send(msg)
	return err
}

func (h *HikeHandler) ShowMenu(ctx context.Context, m *tgbot.Message) error {
	msg := tgbot.NewMessage(m.Chat.ID, "Раздел хайков")
	msg.ReplyMarkup = hikeUI.HikeMenu()

	_, err := h.bot.Send(msg)
	return err
}

func (h *HikeHandler) StartCreateHike(ctx context.Context, m *tgbot.Message) error {
	h.fsm.Reset(m.From.ID)
	h.fsm.Set(m.From.ID, fsm.StateCreateTitleRU)
	return h.sendCreateStep(m.Chat.ID, "Введите название RU:")
}

func (h *HikeHandler) HandleSelectHike(ctx context.Context, m *tgbot.Message) error {
	txt := strings.TrimSpace(m.Text)

	if txt == "⬅️ Назад" {
		h.fsm.Reset(m.From.ID)
		return h.ShowMenu(ctx, m)
	}

	hikeID, err := strconv.Atoi(txt)
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите корректный ID хайка."))
		return nil
	}

	hike, err := h.service.GetHike(ctx, int32(hikeID))
	if err != nil {
		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, fmt.Sprintf("Хайк с ID %d не найден.", hikeID)))
		return nil
	}

	h.fsm.Put(m.From.ID, "selected_hike_id", fmt.Sprintf("%d", hike.ID))
	h.fsm.Put(m.From.ID, "selected_hike_title", hike.TitleRu)
	h.fsm.Put(m.From.ID, "selected_hike_is_published", strconv.FormatBool(hike.IsPublished))
	h.fsm.Set(m.From.ID, fsm.StateSelectedHikeAction)

	msg := tgbot.NewMessage(m.Chat.ID, fmt.Sprintf("Выбран хайк: %s", hike.TitleRu))
	msg.ReplyMarkup = hikeUI.SelectedHikeActionsKeyboard(hike.IsPublished)

	_, err = h.bot.Send(msg)
	return err
}

func (h *HikeHandler) HandlePublishHike(ctx context.Context, m *tgbot.Message) error {
	txt := strings.TrimSpace(m.Text)

	switch h.fsm.State(m.From.ID) {
	case fsm.StateSelectedHikeAction:
		data := h.fsm.Data(m.From.ID)
		title := data["selected_hike_title"]
		isPublished, _ := strconv.ParseBool(data["selected_hike_is_published"])

		switch txt {
		case "📢 Опубликовать хайк":
			if isPublished {
				_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Этот хайк уже опубликован."))
				return err
			}

			h.fsm.Set(m.From.ID, fsm.StateConfirmPublishHike)

			msg := tgbot.NewMessage(
				m.Chat.ID,
				fmt.Sprintf("Вы действительно хотите опубликовать хайк?\n\n%s", title),
			)
			msg.ReplyMarkup = hikeUI.PublishConfirmKeyboard()

			_, err := h.bot.Send(msg)
			return err

		case "🙈 Скрыть хайк":
			if !isPublished {
				_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Этот хайк уже скрыт."))
				return err
			}

			h.fsm.Set(m.From.ID, fsm.StateConfirmHideHike)

			msg := tgbot.NewMessage(
				m.Chat.ID,
				fmt.Sprintf("Вы действительно хотите скрыть хайк?\n\n%s", title),
			)
			msg.ReplyMarkup = hikeUI.HideConfirmKeyboard()

			_, err := h.bot.Send(msg)
			return err

		case "🧾 Карточка хайка":
			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Карточка хайка пока в разработке."))
			return err

		case "⬅️ Назад":
			h.fsm.Set(m.From.ID, fsm.StateSelectHikeID)
			_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Введите ID хайка."))
			return err

		default:
			msg := tgbot.NewMessage(m.Chat.ID, "Выберите действие с помощью кнопок ниже.")
			msg.ReplyMarkup = hikeUI.SelectedHikeActionsKeyboard(isPublished)

			_, err := h.bot.Send(msg)
			return err
		}

	case fsm.StateConfirmPublishHike:
		switch txt {
		case "✅ Да, опубликовать":
			return h.confirmPublishHike(ctx, m)

		case "❌ Отмена":
			return h.backToSelectedHikeActions(m)

		default:
			msg := tgbot.NewMessage(m.Chat.ID, "Подтвердите публикацию или отмените действие.")
			msg.ReplyMarkup = hikeUI.PublishConfirmKeyboard()

			_, err := h.bot.Send(msg)
			return err
		}

	case fsm.StateConfirmHideHike:
		switch txt {
		case "✅ Да, скрыть":
			return h.confirmHideHike(ctx, m)

		case "❌ Отмена":
			return h.backToSelectedHikeActions(m)

		default:
			msg := tgbot.NewMessage(m.Chat.ID, "Подтвердите скрытие или отмените действие.")
			msg.ReplyMarkup = hikeUI.HideConfirmKeyboard()

			_, err := h.bot.Send(msg)
			return err
		}
	}

	h.fsm.Reset(m.From.ID)
	_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Неизвестное состояние. Сбросил сценарий."))
	return err
}

func (h *HikeHandler) confirmPublishHike(ctx context.Context, m *tgbot.Message) error {
	data := h.fsm.Data(m.From.ID)

	hikeID, err := strconv.Atoi(data["selected_hike_id"])
	if err != nil {
		h.fsm.Reset(m.From.ID)
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Некорректный ID хайка. Состояние сброшено."))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	hike, err := h.service.GetHike(ctx, int32(hikeID))
	if err != nil {
		h.fsm.Reset(m.From.ID)
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, fmt.Sprintf("Хайк с ID %d не найден.", hikeID)))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	if hike.IsPublished {
		h.fsm.Reset(m.From.ID)

		msg := tgbot.NewMessage(m.Chat.ID, "Этот хайк уже опубликован.")
		msg.ReplyMarkup = hikeUI.HikeMenu()

		_, err := h.bot.Send(msg)
		return err
	}

	if err := h.service.PublishHike(ctx, int32(hikeID)); err != nil {
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не удалось опубликовать хайк."))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	h.fsm.Reset(m.From.ID)

	msg := tgbot.NewMessage(m.Chat.ID, "Хайк успешно опубликован ✅")
	msg.ReplyMarkup = hikeUI.HikeMenu()

	_, err = h.bot.Send(msg)
	return err
}

func (h *HikeHandler) confirmHideHike(ctx context.Context, m *tgbot.Message) error {
	data := h.fsm.Data(m.From.ID)

	hikeID, err := strconv.Atoi(data["selected_hike_id"])
	if err != nil {
		h.fsm.Reset(m.From.ID)
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Некорректный ID хайка. Состояние сброшено."))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	hike, err := h.service.GetHike(ctx, int32(hikeID))
	if err != nil {
		h.fsm.Reset(m.From.ID)
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, fmt.Sprintf("Хайк с ID %d не найден.", hikeID)))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	if !hike.IsPublished {
		h.fsm.Reset(m.From.ID)

		msg := tgbot.NewMessage(m.Chat.ID, "Этот хайк уже скрыт.")
		msg.ReplyMarkup = hikeUI.HikeMenu()

		_, err := h.bot.Send(msg)
		return err
	}

	if err := h.service.HideHike(ctx, int32(hikeID)); err != nil {
		_, sendErr := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Не удалось скрыть хайк."))
		if sendErr != nil {
			return sendErr
		}
		return err
	}

	h.fsm.Reset(m.From.ID)

	msg := tgbot.NewMessage(m.Chat.ID, "Хайк успешно скрыт 🙈")
	msg.ReplyMarkup = hikeUI.HikeMenu()

	_, err = h.bot.Send(msg)
	return err
}

func (h *HikeHandler) backToSelectedHikeActions(m *tgbot.Message) error {
	data := h.fsm.Data(m.From.ID)
	isPublished, _ := strconv.ParseBool(data["selected_hike_is_published"])

	h.fsm.Set(m.From.ID, fsm.StateSelectedHikeAction)

	msg := tgbot.NewMessage(m.Chat.ID, "Действие отменено. Выберите действие.")
	msg.ReplyMarkup = hikeUI.SelectedHikeActionsKeyboard(isPublished)

	_, err := h.bot.Send(msg)
	return err
}

func (h *HikeHandler) HandleCreateHike(ctx context.Context, m *tgbot.Message) error {
	switch h.fsm.State(m.From.ID) {
	case fsm.StateCreateTitleRU:
		h.fsm.Put(m.From.ID, "title_ru", strings.TrimSpace(m.Text))
		h.fsm.Set(m.From.ID, fsm.StateCreateDescRU)
		return h.sendCreateStep(m.Chat.ID, "Введите описание RU:")

	case fsm.StateCreateDescRU:
		h.fsm.Put(m.From.ID, "description_ru", strings.TrimSpace(m.Text))
		h.fsm.Set(m.From.ID, fsm.StateCreatePrice)
		return h.sendCreateStep(m.Chat.ID, "Введите цену в лари (например: 120):")

	case fsm.StateCreatePrice:
		txt := strings.TrimSpace(m.Text)

		price, err := strconv.Atoi(txt)
		if err != nil || price < 0 {
			_ = h.sendCreateStep(m.Chat.ID, "Введите корректную цену в лари целым числом. Например: 120")
			return nil
		}

		h.fsm.Put(m.From.ID, "price_gel", strconv.Itoa(price))
		h.fsm.Set(m.From.ID, fsm.StateCreateDistanceKm)
		return h.sendCreateStep(m.Chat.ID, "Введите длину маршрута в км (например: 8.5):")

	case fsm.StateCreateDistanceKm:
		txt := strings.TrimSpace(strings.ReplaceAll(m.Text, ",", "."))

		distance, err := strconv.ParseFloat(txt, 64)
		if err != nil || distance < 0 {
			_ = h.sendCreateStep(m.Chat.ID, "Введите корректную длину маршрута. Например: 8.5")
			return nil
		}

		h.fsm.Put(m.From.ID, "distance_km", strconv.FormatFloat(distance, 'f', 2, 64))
		h.fsm.Set(m.From.ID, fsm.StateCreateElevationGain)
		return h.sendCreateStep(m.Chat.ID, "Введите набор высоты в метрах (например: 650):")

	case fsm.StateCreateElevationGain:
		txt := strings.TrimSpace(m.Text)

		elevationGain, err := strconv.Atoi(txt)
		if err != nil || elevationGain < 0 {
			_ = h.sendCreateStep(m.Chat.ID, "Введите корректный набор высоты в метрах. Например: 650")
			return nil
		}

		h.fsm.Put(m.From.ID, "elevation_gain_m", strconv.Itoa(elevationGain))
		h.fsm.Set(m.From.ID, fsm.StateCreateDates)

		examples := "Введите даты начала и завершения хайка (примеры: 10, 10 12, 10-12, 31 3, 03.02-04.02, 15.12 16.12)."
		return h.sendCreateStep(m.Chat.ID, examples)

	case fsm.StateCreateDates:
		loc := h.loc
		start, end, err := parser.ParseHikeDates(m.Text, time.Now().In(loc), loc)
		if err != nil {
			_ = h.sendCreateStep(m.Chat.ID, "Не получилось распознать даты. Попробуйте ещё раз.\nПримеры: 10 · 10 12 · 10-12 · 31 3 · 03.02-04.02 · 15.12 16.12")
			return nil
		}

		h.fsm.Put(m.From.ID, "starts_at", start.Format("02.01.2006 15:04"))
		h.fsm.Put(m.From.ID, "ends_at", end.Format("02.01.2006 15:04"))
		h.fsm.Set(m.From.ID, fsm.StateCreatePhoto)

		return h.sendCreateStep(m.Chat.ID, "Загрузите фото:")

	case fsm.StateCreatePhoto:
		if len(m.Photo) == 0 {
			_ = h.sendCreateStep(m.Chat.ID, "Пожалуйста, отправьте именно фото.")
			return nil
		}

		photo := m.Photo[len(m.Photo)-1]
		h.fsm.Put(m.From.ID, "photo_file_id", photo.FileID)
		h.fsm.Set(m.From.ID, fsm.StateConfirm)

		preview := fmt.Sprintf(
			"Проверьте данные:\n\nНазвание: %s\nОписание: %s\nЦена: %s GEL\nДлина: %s км\nНабор высоты: %s м\nДаты: %s → %s\nФото: добавлено\n\nНапишите 'ok' для сохранения или 'cancel' для отмены.",
			h.fsm.Data(m.From.ID)["title_ru"],
			h.fsm.Data(m.From.ID)["description_ru"],
			h.fsm.Data(m.From.ID)["price_gel"],
			h.fsm.Data(m.From.ID)["distance_km"],
			h.fsm.Data(m.From.ID)["elevation_gain_m"],
			h.fsm.Data(m.From.ID)["starts_at"],
			h.fsm.Data(m.From.ID)["ends_at"],
		)

		msg := tgbot.NewMessage(m.Chat.ID, preview)
		msg.ReplyMarkup = hikeUI.HikeConfirmMenu()
		_, err := h.bot.Send(msg)
		return err

	case fsm.StateConfirm:
		txt := strings.TrimSpace(strings.ToLower(m.Text))

		switch txt {
		case "✅ подтвердить":
			if err := h.saveCreatedHike(ctx, m.From.ID); err != nil {
				_ = h.sendCreateStep(m.Chat.ID, "Ошибка при сохранении хайка :(")
				return err
			}

			h.fsm.Reset(m.From.ID)

			msg := tgbot.NewMessage(m.Chat.ID, "Хайк создан!")
			msg.ReplyMarkup = hikeUI.HikeMenu()

			_, err := h.bot.Send(msg)
			return err

		case "❌ отмена":
			h.fsm.Reset(m.From.ID)

			msg := tgbot.NewMessage(m.Chat.ID, "Создание отменено.")
			msg.ReplyMarkup = hikeUI.HikeMenu()

			_, err := h.bot.Send(msg)
			return err

		default:
			_ = h.sendCreateStep(m.Chat.ID, "Выберите действие кнопкой: ✅ Подтвердить, ❌ Отмена или ⬅️ Назад.")
			return nil
		}

	default:
		h.fsm.Reset(m.From.ID)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Сбросил состояние."))
		return err
	}
}

func (h *HikeHandler) saveCreatedHike(ctx context.Context, userID int64) error {
	data := h.fsm.Data(userID)

	startAt, err := time.ParseInLocation("02.01.2006 15:04", data["starts_at"], h.loc)
	if err != nil {
		return err
	}

	endsAt, err := time.ParseInLocation("02.01.2006 15:04", data["ends_at"], h.loc)
	if err != nil {
		return err
	}

	priceGel, err := strconv.Atoi(data["price_gel"])
	if err != nil {
		return err
	}

	distanceKm, err := strconv.ParseFloat(data["distance_km"], 64)
	if err != nil {
		return err
	}

	elevationGainM, err := strconv.Atoi(data["elevation_gain_m"])
	if err != nil {
		return err
	}

	hike := service.Hike{
		TitleRu:        data["title_ru"],
		DescriptionRu:  data["description_ru"],
		PriceGel:       priceGel,
		DistanceKm:     distanceKm,
		ElevationGainM: elevationGainM,
		StartsAt:       startAt,
		EndsAt:         endsAt,
		PhotoFileID:    data["photo_file_id"],
	}

	createdHikeID, err := h.service.CreateHike(ctx, hike)
	if err != nil {
		return err
	}

	imagePath, err := h.saveImage(ctx, data["photo_file_id"], createdHikeID)
	if err != nil {
		return err
	}

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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.jpg", hikeID))

	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
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

	h.fsm.Set(m.From.ID, fsm.StateSelectHikeID)
	b.WriteString("\nОтправьте ID хайка, чтобы выбрать действие.")

	_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, b.String()))
	return err
}
