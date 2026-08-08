package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
)

const pollingInterval = 5 * time.Second

const bookingStatusChanged = "booking_status_changed"

type bookingStatusChangedPayload struct {
	BookingID     int32    `json:"booking_id"`
	HikeTitle     string   `json:"hike_title"`
	NewStatus     string   `json:"new_status"`
	DistanceKm    *float64 `json:"distance_km"`
	ElevationGain *int32   `json:"elevation_gain_m"`
}

type Notification struct {
	ID        int64
	TgUserID  int64
	Type      string
	Payload   json.RawMessage
	Attempts  int32
	CreatedAt time.Time
}

type Repository interface {
	ClaimPending(ctx context.Context, limit int32) ([]Notification, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, message string) error
	ReleaseForRetry(ctx context.Context, id int64, message string, nextRetryAt time.Time) error
}

type Worker struct {
	log  logger.Logger
	bot  *tgbot.BotAPI
	repo Repository
}

func New(l logger.Logger, b *tgbot.BotAPI, r Repository) *Worker {
	return &Worker{
		log:  l,
		bot:  b,
		repo: r,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()
	w.log.Info("notification worker started")
	for {
		select {
		case <-ctx.Done():
			w.log.Info("notification worker stopped")
			return
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.log.StructuredError(
					"notification batch processing error",
					err,
				)
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) error {
	notifications, err := w.repo.ClaimPending(ctx, 10)
	if err != nil {
		return err
	}
	for _, notification := range notifications {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := w.processOne(ctx, notification); err != nil {
			w.log.
				WithField("notification_id", notification.ID).
				StructuredError("notification processing error", err)
		}
	}
	return nil
}

func (w *Worker) processOne(ctx context.Context, notification Notification) error {
	text, err := notificationText(notification)
	if err != nil {
		if markErr := w.repo.MarkFailed(ctx, notification.ID, err.Error()); markErr != nil {
			return errors.Join(err, markErr)
		}
		return err
	}

	message := tgbot.NewMessage(notification.TgUserID, text)
	_, err = w.bot.Send(message)
	if err != nil {
		return w.handleSendError(ctx, notification, err)
	}
	if err := w.repo.MarkSent(ctx, notification.ID); err != nil {
		return err
	}
	return nil
}

func notificationText(notification Notification) (string, error) {
	if notification.Type != bookingStatusChanged {
		return "", fmt.Errorf("unsupported notification type %q", notification.Type)
	}

	var payload bookingStatusChangedPayload
	if err := json.Unmarshal(notification.Payload, &payload); err != nil {
		return "", fmt.Errorf("decode notification payload: %w", err)
	}

	switch payload.NewStatus {
	case "confirmed":
		return fmt.Sprintf("Ваша заявка на участие в хайке %q подтверждена ✅", payload.HikeTitle), nil
	case "canceled":
		return fmt.Sprintf("Ваша заявка на участие в хайке %q отменена.", payload.HikeTitle), nil
	case "completed":
		return completedText(payload), nil
	default:
		return "", fmt.Errorf("unsupported booking status %q", payload.NewStatus)
	}
}

func completedText(payload bookingStatusChangedPayload) string {
	result := fmt.Sprintf("Поздравляем! Вы завершили хайк %q 🎉", payload.HikeTitle)

	if payload.DistanceKm != nil {
		result += fmt.Sprintf("\n🥾 Пройдено: %g км", *payload.DistanceKm)
	}
	if payload.ElevationGain != nil {
		result += fmt.Sprintf("\n⛰ Набор высоты: %d м", *payload.ElevationGain)
	}

	return result + "\n\nОтличное достижение! Так держать 🙌"
}

func (w *Worker) handleSendError(ctx context.Context, notification Notification, sendErr error) error {
	const maxAttempts = 5
	if notification.Attempts >= maxAttempts {
		if err := w.repo.MarkFailed(
			ctx,
			notification.ID,
			sendErr.Error(),
		); err != nil {
			return errors.Join(sendErr, err)
		}
		return sendErr
	}
	nextRetryAt := time.Now().Add(
		retryDelay(int(notification.Attempts)),
	)
	if err := w.repo.ReleaseForRetry(
		ctx,
		notification.ID,
		sendErr.Error(),
		nextRetryAt,
	); err != nil {
		return errors.Join(sendErr, err)
	}
	return sendErr
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return 5 * time.Second * time.Duration(1<<(attempt-1))
}
