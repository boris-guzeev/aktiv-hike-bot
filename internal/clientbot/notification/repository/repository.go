package notification

import (
	"context"
	"errors"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/notification/worker"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	log     logger.Logger
	queries *client.Queries
}

func New(l logger.Logger, q *client.Queries) worker.Repository {
	return &repository{
		log:     l,
		queries: q,
	}
}

func (r *repository) ClaimPending(ctx context.Context, limit int32) ([]worker.Notification, error) {
	rows, err := r.queries.ClaimPendingNotifications(ctx, limit)

	if err != nil {
		return nil, err
	}
	notifications := make([]worker.Notification, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, worker.Notification{
			ID:        row.ID,
			TgUserID:  row.RecipientTgUserID,
			Type:      row.Type,
			Payload:   row.Payload,
			Attempts:  row.Attempts,
			CreatedAt: row.CreatedAt,
		})
	}
	return notifications, nil
}

var ErrNotificationNotProcessing = errors.New("notification is not processing")

func (r *repository) MarkSent(ctx context.Context, id int64) error {
	affected, err := r.queries.MarkNotificationSent(ctx, id)
	if err != nil {
		return logger.WrapError(err)
	}

	if affected == 0 {
		return logger.WrapError(ErrNotificationNotProcessing)
	}

	return nil
}

func (r *repository) MarkFailed(ctx context.Context, id int64, message string) error {
	affected, err := r.queries.MarkNotificationFailed(
		ctx,
		client.MarkNotificationFailedParams{
			ID:        id,
			LastError: pgtype.Text{String: message, Valid: true},
		},
	)
	if err != nil {
		return logger.WrapError(err)
	}

	if affected == 0 {
		return logger.WrapError(ErrNotificationNotProcessing)
	}

	return nil
}

func (r *repository) ReleaseForRetry(ctx context.Context, id int64, message string, nextRetryAt time.Time) error {
	affected, err := r.queries.ReleaseNotificationForRetry(
		ctx,
		client.ReleaseNotificationForRetryParams{
			ID:          id,
			LastError:   pgtype.Text{String: message, Valid: true},
			NextRetryAt: pgtype.Timestamptz{Time: nextRetryAt, Valid: true},
		},
	)
	if err != nil {
		return err
	}

	if affected == 0 {
		return logger.WrapError(ErrNotificationNotProcessing)
	}

	return nil
}
