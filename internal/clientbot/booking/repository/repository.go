package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/client"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/clientbot/booking/service"
)

type repository struct {
	queries *client.Queries
}

func New(q *client.Queries) service.Repository {
	return &repository{queries: q}
}

func (r *repository) GetByID(ctx context.Context, id int32) (service.Booking, error) {
	rawBooking, err := r.queries.GetBookingByID(ctx, id)
	if err != nil {
		return service.Booking{}, logger.WrapError(err)
	}

	var takenByAdminID *int32
	if rawBooking.TakenByAdminID.Valid {
		takenByAdminID = &rawBooking.TakenByAdminID.Int32
	}

	var takenAt *time.Time
	if rawBooking.TakenAt.Valid {
		takenAt = &rawBooking.TakenAt.Time
	}

	return service.Booking{
		ID:             rawBooking.ID,
		HikeID:         rawBooking.HikeID,
		UserID:         rawBooking.UserID,
		Status:         service.BookingStatus(rawBooking.Status),
		TakenByAdminID: takenByAdminID,
		TakenAt:        takenAt,
	}, nil
}

func (r *repository) Create(ctx context.Context, booking service.Booking) (int32, error) {
	id, err := r.queries.CreateBooking(ctx, client.CreateBookingParams{
		HikeID: booking.HikeID,
		UserID: booking.UserID,
		Status: string(service.StatusNew),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, logger.WrapError(service.ErrBookingAlreadyExists)
		}
		return 0, logger.WrapError(err)
	}

	return id, nil
}

func (r *repository) TakeInProgress(ctx context.Context, bookingID, adminID int32) (int32, error) {
	inProgressBookingID, err := r.queries.TakeBookingInProgress(ctx, client.TakeBookingInProgressParams{
		ID:             bookingID,
		Status:         string(service.StatusInProgress),
		TakenByAdminID: toPgInt4(adminID),
		ExpectedStatus: string(service.StatusNew),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, logger.WrapError(service.ErrBookingAlreadyTaken)
		}
		return 0, logger.WrapError(err)
	}

	return inProgressBookingID, nil
}

// TODO: вынести отдельно
func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}
