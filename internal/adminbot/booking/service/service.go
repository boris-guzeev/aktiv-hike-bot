package service

import (
	"context"
	"errors"
	"time"
)

type BookingStatus string

const (
	StatusNew        BookingStatus = "new"
	StatusInProgress BookingStatus = "in_progress"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusCompleted  BookingStatus = "completed"
	StatusCanceled   BookingStatus = "canceled"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrNotYourBooking          = errors.New("not your booking")
	ErrBookingNotUpdated       = errors.New("booking was not updated")
)

type Booking struct {
	ID             int32
	HikeID         int32
	HikeTitle      string
	UserID         int32
	UserName       string
	UserTgID       int64
	Status         BookingStatus
	TakenByAdminID *int32
	TakenAt        *time.Time
	CreatedAt      time.Time
}

type Repository interface {
	GetByID(ctx context.Context, id int32) (Booking, error)
	UpdateStatus(ctx context.Context, id int32, newStatus BookingStatus) error
	ListAdminBookings(ctx context.Context, adminID int32) ([]Booking, error)
}

type Service interface {
	GetByID(ctx context.Context, id int32) (Booking, error)
	UpdateStatus(ctx context.Context, id, adminID int32, newStatus BookingStatus) error
	ListAdminBookings(ctx context.Context, adminID int32) ([]Booking, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) GetByID(ctx context.Context, id int32) (Booking, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) UpdateStatus(ctx context.Context, id, adminID int32, newStatus BookingStatus) error {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if booking.TakenByAdminID == nil || *booking.TakenByAdminID != adminID {
		return ErrNotYourBooking
	}

	if !canTransition(booking.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	return s.repo.UpdateStatus(ctx, id, newStatus)
}

func (s *service) ListAdminBookings(ctx context.Context, adminID int32) ([]Booking, error) {
	return s.repo.ListAdminBookings(ctx, adminID)
}

func canTransition(from, to BookingStatus) bool {
	switch from {

	case StatusInProgress:
		return to == StatusConfirmed ||
			to == StatusCanceled

	case StatusConfirmed:
		return to == StatusCompleted ||
			to == StatusCanceled

	default:
		return false
	}
}
