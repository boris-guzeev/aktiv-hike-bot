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

var bookingStatuses = map[BookingStatus]string{
	StatusNew:        "новая",
	StatusInProgress: "в работе",
	StatusConfirmed:  "подтверждена",
	StatusCompleted:  "завершена",
	StatusCanceled:   "отменена",
}

func (b BookingStatus) String() string {
	return bookingStatuses[b]
}

var (
	ErrBookingAlreadyExists = errors.New("booking already exists")
	ErrBookingAlreadyTaken  = errors.New("booking already taken")
)

type Booking struct {
	ID             int32
	HikeID         int32
	UserID         int32
	Status         BookingStatus
	TakenByAdminID *int32
	TakenAt        *time.Time
}

type Repository interface {
	Create(ctx context.Context, booking Booking) (int32, error)
	TakeInProgress(ctx context.Context, bookingID, adminID int32) (int32, error)
}

type Service interface {
	Create(ctx context.Context, hikeID, userID int32) (int32, error)
	TakeInProgress(ctx context.Context, bookingID, adminID int32) (int32, error)
}

type service struct {
	repo Repository
}

func New(r Repository) Service {
	return &service{repo: r}
}

func (s *service) Create(ctx context.Context, hikeID, userID int32) (int32, error) {
	booking := Booking{
		HikeID: hikeID,
		UserID: userID,
		Status: StatusNew,
	}
	return s.repo.Create(ctx, booking)
}

func (s *service) TakeInProgress(ctx context.Context, bookingID, adminID int32) (int32, error) {
	return s.repo.TakeInProgress(ctx, bookingID, adminID)
}
