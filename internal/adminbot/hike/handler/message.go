package handler

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/fsm"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/parser"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/common"
	hikeUI "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/ui/hike"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
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
		fsm.StateCreatePreviewRU,
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
		return h.HandleSelectedHike(ctx, m)

	case fsm.StateViewDetailsHike, fsm.StateEditHikeTitleRU, fsm.StateEditHikePreviewRU, fsm.StateEditHikeDescriptionRU:
		return h.HandleHikeDetailsFlow(ctx, m)

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

func (h *HikeHandler) HandleSelectedHike(ctx context.Context, m *tgbot.Message) error {
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
			err := h.showHikeDetails(ctx, m)
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

func (h *HikeHandler) HandleHikeDetailsFlow(ctx context.Context, m *tgbot.Message) error {
	switch h.fsm.State(m.From.ID) {
	case fsm.StateViewDetailsHike:
		return h.handleHikeDetailsActions(ctx, m)

	case fsm.StateEditHikeTitleRU:
		return h.handleEditTitleRu(ctx, m)

	case fsm.StateEditHikePreviewRU:
		return h.handleEditPreviewRu(ctx, m)

	case fsm.StateEditHikeDescriptionRU:
		return h.handleEditDescriptionRu(ctx, m)
	}

	h.fsm.Reset(m.From.ID)
	_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Неизвестное состояние. Сценарий сброшен."))
	return err
}

func (h *HikeHandler) handleHikeDetailsActions(ctx context.Context, m *tgbot.Message) error {
	data := h.fsm.Data(m.From.ID)
	strID := data["selected_hike_id"]
	hikeID64, err := strconv.ParseInt(strID, 10, 32)
	if err != nil {
		return logger.WrapError(err)
	}
	hikeID := int32(hikeID64)

	hike, err := h.service.GetHike(ctx, hikeID)
	if err != nil {
		return logger.WrapError(err)
	}

	switch m.Text {
	case hikeUI.ButtonEditTitleRu:
		h.fsm.Set(m.From.ID, fsm.StateEditHikeTitleRU)

		// Current TitleRu Notice
		msg := tgbot.NewMessage(m.Chat.ID, "Текущее название RU одним сообщением:")
		msg.ReplyMarkup = common.OnlyBackKeyboard()
		_, err := h.bot.Send(msg)
		if err != nil {
			return logger.WrapError(err)
		}

		// Current TitleRu Message
		_, err = h.bot.Send(
			tgbot.NewMessage(m.Chat.ID, hike.TitleRu),
		)
		if err != nil {
			return logger.WrapError(err)
		}

		// New TitleRu Request
		_, err = h.bot.Send(
			tgbot.NewMessage(m.Chat.ID, "Введите новое название RU:"),
		)
		return logger.WrapError(err)

	case hikeUI.ButtonEditPreviewRu:
		h.fsm.Set(m.From.ID, fsm.StateEditHikePreviewRU)

		// Current PreviewRu Notice
		msg := tgbot.NewMessage(m.Chat.ID, "Текущее превью RU одним сообщением:")
		msg.ReplyMarkup = common.OnlyBackKeyboard()
		if len(strings.TrimSpace(hike.PreviewRu)) != 0 {
			_, err := h.bot.Send(msg)
			if err != nil {
				return logger.WrapError(err)
			}

			// Current PreviewRu Message
			_, err = h.bot.Send(
				tgbot.NewMessage(m.Chat.ID, hike.PreviewRu),
			)
			if err != nil {
				return logger.WrapError(err)
			}
		}
		// New PreviewRu Request
		_, err = h.bot.Send(
			tgbot.NewMessage(m.Chat.ID, "Введите новое превью RU:"),
		)
		return logger.WrapError(err)

	case hikeUI.ButtonEditDescriptionRu:
		h.fsm.Set(m.From.ID, fsm.StateEditHikeDescriptionRU)

		// Current DescriptionRu Notice
		msg := tgbot.NewMessage(m.Chat.ID, "Текущее описание RU одним сообщением:")
		msg.ReplyMarkup = common.OnlyBackKeyboard()
		if len(strings.TrimSpace(hike.DescriptionRu)) != 0 {
			_, err := h.bot.Send(msg)
			if err != nil {
				return logger.WrapError(err)
			}

			// Current DescriptionRu Message
			_, err = h.bot.Send(
				tgbot.NewMessage(m.Chat.ID, hike.DescriptionRu),
			)
			if err != nil {
				return logger.WrapError(err)
			}
		}
		// New DescriptionRu Request
		_, err = h.bot.Send(
			tgbot.NewMessage(m.Chat.ID, "Введите новое описание RU:"),
		)
		return logger.WrapError(err)

	case common.ButtonBack:
		h.fsm.Set(m.From.ID, fsm.StateSelectedHikeAction)

		msg := tgbot.NewMessage(m.Chat.ID, "Выберите действие")
		data := h.fsm.Data(m.From.ID)
		isPublished, _ := strconv.ParseBool(data["selected_hike_is_published"])
		msg.ReplyMarkup = hikeUI.SelectedHikeActionsKeyboard(isPublished)

		_, err := h.bot.Send(msg)
		return logger.WrapError(err)
	}

	return nil
}

func (h *HikeHandler) handleEditTitleRu(ctx context.Context, m *tgbot.Message) error {
	title := strings.TrimSpace(m.Text)
	if title == "" {
		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Название не может быть пустым."))
		return nil
	}

	data := h.fsm.Data(m.From.ID)

	hikeID, err := strconv.Atoi(data["selected_hike_id"])
	if err != nil {
		return logger.WrapError(err)
	}

	if err := h.service.UpdateTitleRu(ctx, int32(hikeID), title); err != nil {
		return err
	}

	h.fsm.Put(m.From.ID, "selected_hike_title", title)
	h.fsm.Set(m.From.ID, fsm.StateViewDetailsHike)

	_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Название обновлено ✅"))
	if err != nil {
		return err
	}

	return h.showHikeDetails(ctx, m)
}

func (h *HikeHandler) handleEditPreviewRu(ctx context.Context, m *tgbot.Message) error {
	preview := strings.TrimSpace(m.Text)
	if preview == "" {
		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Превью не может быть пустым."))
		return nil
	}

	data := h.fsm.Data(m.From.ID)

	hikeID, err := strconv.Atoi(data["selected_hike_id"])
	if err != nil {
		return logger.WrapError(err)
	}

	if err := h.service.UpdatePreviewRu(ctx, int32(hikeID), preview); err != nil {
		return err
	}

	h.fsm.Set(m.From.ID, fsm.StateViewDetailsHike)

	_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Превью обновлено ✅"))
	if err != nil {
		return err
	}

	return h.showHikeDetails(ctx, m)
}

func (h *HikeHandler) handleEditDescriptionRu(ctx context.Context, m *tgbot.Message) error {
	description := strings.TrimSpace(m.Text)
	if description == "" {
		_, _ = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Описание не может быть пустым."))
		return nil
	}

	data := h.fsm.Data(m.From.ID)

	hikeID, err := strconv.Atoi(data["selected_hike_id"])
	if err != nil {
		return logger.WrapError(err)
	}

	if err := h.service.UpdateDescriptionRu(ctx, int32(hikeID), description); err != nil {
		return err
	}

	h.fsm.Set(m.From.ID, fsm.StateViewDetailsHike)

	_, err = h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Описание обновлено ✅"))
	if err != nil {
		return err
	}

	return h.showHikeDetails(ctx, m)
}

func (h *HikeHandler) showHikeDetails(ctx context.Context, m *tgbot.Message) error {
	data := h.fsm.Data(m.From.ID)

	strID := data["selected_hike_id"]
	id, err := strconv.ParseInt(strID, 10, 32)
	if err != nil {
		return logger.WrapError(err)
	}

	hike, err := h.service.GetHike(ctx, int32(id))
	if err != nil {
		return logger.WrapError(err)
	}

	h.fsm.Set(m.From.ID, fsm.StateViewDetailsHike)

	card := buildHikeDetailsMessage(hike)

	msg := tgbot.NewMessage(m.Chat.ID, card)
	msg.ParseMode = tgbot.ModeHTML
	msg.ReplyMarkup = hikeUI.HikeDetailsKeyboard()

	if _, err := h.bot.Send(msg); err != nil {
		return logger.WrapError(err)
	}

	return nil
}

func buildHikeDetailsMessage(hike service.Hike) string {
	var b strings.Builder

	b.WriteString("<b>🧾 Карточка хайка</b>\n\n")
	b.WriteString(fmt.Sprintf("<b>ID:</b> %d\n", hike.ID))
	b.WriteString(fmt.Sprintf("<b>Название:</b> %s\n", html.EscapeString(hike.TitleRu)))
	b.WriteString(
		fmt.Sprintf(
			"<b>Опубликован:</b> %s\n", map[bool]string{
				true:  "да",
				false: "нет",
			}[hike.IsPublished],
		),
	)
	b.WriteString(fmt.Sprintf("<b>Дата начала:</b> %s\n", common.Format(hike.StartsAt)))
	b.WriteString(fmt.Sprintf("<b>Дата окончания:</b> %s\n", common.Format(hike.EndsAt)))

	if hike.PriceGel != 0 {
		b.WriteString(fmt.Sprintf("<b>Цена:</b> %d GEL\n", hike.PriceGel))
	}

	if hike.DistanceKm != 0 {
		b.WriteString(fmt.Sprintf("<b>Длина:</b> %.2f км\n", hike.DistanceKm))
	}

	if hike.ElevationGainM != 0 {
		b.WriteString(fmt.Sprintf("<b>Набор высоты:</b> %d м\n", hike.ElevationGainM))
	}

	if strings.TrimSpace(hike.PreviewRu) == "" {
		b.WriteString("<b>Превью:</b> —\n")
	} else {
		b.WriteString("<b>Превью:</b> заполнено\n")
	}

	if strings.TrimSpace(hike.DescriptionRu) == "" {
		b.WriteString("<b>Описание RU:</b> —\n")
	} else {
		b.WriteString("<b>Описание RU:</b> заполнено\n")
	}

	return b.String()
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
		h.fsm.Set(m.From.ID, fsm.StateCreatePreviewRU)
		return h.sendCreateStep(m.Chat.ID, "Введите превью RU (1024 символа):")

	case fsm.StateCreatePreviewRU:
		preview := strings.TrimSpace(m.Text)
		if preview == "" {
			_ = h.sendCreateStep(m.Chat.ID, "Введите краткое описание для превью.")
			return nil
		}

		count := utf8.RuneCountInString(preview)
		if count > 1024 {
			_ = h.sendCreateStep(
				m.Chat.ID,
				fmt.Sprintf(
					"Превью слишком длинное: %d символов из 1024 допустимых. Сократите текст.",
					count,
				),
			)
			return nil
		}

		h.fsm.Put(m.From.ID, "preview_ru", preview)
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

		clientCaptionLen := countClientCaption(h.fsm.Data(m.From.ID))
		if clientCaptionLen > 1024 {
			_ = h.sendCreateStep(
				m.Chat.ID,
				fmt.Sprintf(
					"Итоговый Telegram caption слишком длинный: %d символов из 1024 допустимых. Сократите превью.",
					clientCaptionLen,
				),
			)
			return nil
		}
		h.fsm.Set(m.From.ID, fsm.StateConfirm)

		// TODO: вынести формирование Caption в общий пакет
		preview := fmt.Sprintf(
			"📋 <b>Проверьте данные хайка</b>\n\n"+
				"🏔 Название: %s\n"+
				"🔎 Превью: %s\n"+
				"📝 Описание: %s\n"+
				"💰 Цена: %s GEL\n"+
				"📏 Длина: %s км\n"+
				"⛰ Набор высоты: %s м\n"+
				"🗓 Даты: %s → %s\n"+
				"📷 Фото: добавлено\n\n"+
				"📐 Общий Telegram caption: %d / 1024\n\n"+
				"Выберите действие ниже:",
			h.fsm.Data(m.From.ID)["title_ru"],
			h.fsm.Data(m.From.ID)["preview_ru"],
			h.fsm.Data(m.From.ID)["description_ru"],
			h.fsm.Data(m.From.ID)["price_gel"],
			h.fsm.Data(m.From.ID)["distance_km"],
			h.fsm.Data(m.From.ID)["elevation_gain_m"],
			h.fsm.Data(m.From.ID)["starts_at"],
			h.fsm.Data(m.From.ID)["ends_at"],
			clientCaptionLen,
		)

		msg := tgbot.NewMessage(m.Chat.ID, preview)
		msg.ReplyMarkup = hikeUI.HikeConfirmMenu()
		msg.ParseMode = tgbot.ModeHTML
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
			msg := tgbot.NewMessage(m.Chat.ID, "Выберите действие кнопкой: ✅ Подтвердить, ❌ Отмена или ⬅️ Назад.")
			msg.ReplyMarkup = hikeUI.HikeConfirmMenu()
			_, _ = h.bot.Send(msg)
			return nil
		}

	default:
		h.fsm.Reset(m.From.ID)
		_, err := h.bot.Send(tgbot.NewMessage(m.Chat.ID, "Сбросил состояние."))
		return err
	}
}

func countClientCaption(data map[string]string) int {
	return utf8.RuneCountInString(fmt.Sprintf(
		"🏔 <b>%s</b>\n\n"+
			"%s\n\n"+
			"💰 %s GEL\n"+
			"📏 %s км\n"+
			"⛰ %s м\n"+
			"🗓 %s → %s",
		data["title_ru"],
		data["preview_ru"],
		data["price_gel"],
		data["distance_km"],
		data["elevation_gain_m"],
		data["starts_at"],
		data["ends_at"],
	))
}

func (h *HikeHandler) saveCreatedHike(ctx context.Context, userID int64) error {
	data := h.fsm.Data(userID)

	startAt, err := time.ParseInLocation("02.01.2006 15:04", data["starts_at"], h.loc)
	if err != nil {
		return logger.WrapError(err)
	}

	endsAt, err := time.ParseInLocation("02.01.2006 15:04", data["ends_at"], h.loc)
	if err != nil {
		return logger.WrapError(err)
	}

	priceGel, err := strconv.Atoi(data["price_gel"])
	if err != nil {
		return logger.WrapError(err)
	}

	distanceKm, err := strconv.ParseFloat(data["distance_km"], 64)
	if err != nil {
		return logger.WrapError(err)
	}

	elevationGainM, err := strconv.Atoi(data["elevation_gain_m"])
	if err != nil {
		return logger.WrapError(err)
	}

	previewRu := strings.TrimSpace(h.fsm.Data(userID)["preview_ru"])
	if previewRu == "" {
		return logger.WrapError(errors.New("preview_ru is empty"))
	}

	if countClientCaption(data) > 1024 {
		return logger.WrapError(errors.New("client caption is too long"))
	}

	hike := service.Hike{
		TitleRu:        data["title_ru"],
		PreviewRu:      previewRu,
		DescriptionRu:  data["description_ru"],
		PriceGel:       int32(priceGel),
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
		return "", logger.WrapError(err)
	}

	url := file.Link(h.bot.Token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", logger.WrapError(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", logger.WrapError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", logger.WrapError(fmt.Errorf("unexpected http-status code: %d", resp.StatusCode))
	}

	dir := filepath.Join(h.storageRoot, "hikes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", logger.WrapError(err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%d.jpg", hikeID))

	out, err := os.Create(path)
	if err != nil {
		return "", logger.WrapError(err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return "", logger.WrapError(err)
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
