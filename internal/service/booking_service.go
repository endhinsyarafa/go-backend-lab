package service

import (
	"errors"
	"strings"
	"time"

	"github.com/latihan/go-backend-lab/internal/model"
	"github.com/latihan/go-backend-lab/internal/repository"
)

var (
	ErrInvalidTime      = errors.New("start_time must be in the future and end_time must be after start_time")
	ErrRoomNotFound     = repository.ErrRoomNotFound
	ErrRoomInactive     = repository.ErrRoomInactive
	ErrBookingConflict  = repository.ErrBookingConflict
	ErrBookingNotFound  = errors.New("booking not found")
	ErrForbiddenCancel  = errors.New("you are not allowed to cancel this booking")
	ErrAlreadyCancelled = errors.New("booking is already cancelled")
)

type BookingService struct {
	Repo *repository.BookingRepository
}

func NewBookingService(repo *repository.BookingRepository) *BookingService {
	return &BookingService{Repo: repo}
}

func (s *BookingService) CreateBooking(userID int, req model.CreateBookingRequest) (*model.Booking, error) {
	now := time.Now().UTC()
	start := req.StartTime.UTC()
	end := req.EndTime.UTC()

	if start.Before(now) || !end.After(start) {
		return nil, ErrInvalidTime
	}

	booking := &model.Booking{
		RoomID:    req.RoomID,
		UserID:    userID,
		Title:     strings.TrimSpace(req.Title),
		StartTime: start,
		EndTime:   end,
	}

	if err := s.Repo.CreateWithLock(booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) CancelBooking(bookingID, userID int, role string) (*model.Booking, error) {
	booking, err := s.Repo.FindByID(bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}

	if booking.UserID != userID && role != "ADMIN" {
		return nil, ErrForbiddenCancel
	}

	if booking.Status == "CANCELLED" {
		return nil, ErrAlreadyCancelled
	}

	if err := s.Repo.Cancel(bookingID); err != nil {
		return nil, err
	}
	booking.Status = "CANCELLED"
	return booking, nil
}
