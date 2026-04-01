package repository

import (
	"context"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/booking/service"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	"github.com/boris-guzeev/aktiv-hike-bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	queries *admin.Queries
}

func New(q *admin.Queries) service.Repository {
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

	return service.Booking{
		ID:             rawBooking.ID,
		HikeID:         rawBooking.HikeID,
		UserID:         rawBooking.UserID,
		Status:         service.BookingStatus(rawBooking.Status),
		TakenByAdminID: takenByAdminID,
	}, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int32, newStatus service.BookingStatus) (service.Booking, error) {
	rawBooking, err := r.queries.UpdateBookingStatus(ctx, admin.UpdateBookingStatusParams{
		ID:        id,
		NewStatus: string(newStatus),
	})
	if err != nil {
		return service.Booking{}, logger.WrapError(err)
	}

	var takenByAdminID *int32
	if rawBooking.TakenByAdminID.Valid {
		takenByAdminID = &rawBooking.TakenByAdminID.Int32
	}

	return service.Booking{
		ID:             rawBooking.ID,
		HikeID:         rawBooking.HikeID,
		UserID:         rawBooking.UserID,
		Status:         service.BookingStatus(rawBooking.Status),
		TakenByAdminID: takenByAdminID,
	}, nil
}

func (r *repository) ListAdminBookings(ctx context.Context, adminID int32) ([]service.Booking, error) {
	rows, err := r.queries.ListAdminBookings(ctx, pgtype.Int4{Int32: adminID, Valid: true})
	if err != nil {
		return nil, logger.WrapError(err)
	}

	bookings := make([]service.Booking, 0, len(rows))
	for _, row := range rows {
		var takenAt *time.Time
		if row.TakenAt.Valid {
			t := row.TakenAt.Time
			takenAt = &t
		}

		bookings = append(bookings, service.Booking{
			ID:        row.ID,
			HikeID:    row.HikeID,
			HikeTitle: row.HikeTitle,
			UserID:    row.UserID,
			UserName:  row.UserName,
			UserTgID:  row.UserTgID,
			Status:    service.BookingStatus(row.Status),
			TakenAt:   takenAt,
			CreatedAt: row.CreatedAt,
		})
	}

	return bookings, nil
}
