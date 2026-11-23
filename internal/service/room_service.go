package service

import (
	"errors"
	"time"

	"github.com/latihan/go-backend-lab/internal/model"
	"github.com/latihan/go-backend-lab/internal/repository"
)

var ErrInvalidTimeRange = errors.New("end must be after start")

type RoomService struct {
	Repo *repository.RoomRepository
}

func NewRoomService(repo *repository.RoomRepository) *RoomService {
	return &RoomService{Repo: repo}
}

func (s *RoomService) CreateRoom(req model.CreateRoomRequest) (*model.Room, error) {
	room := &model.Room{
		Name:     req.Name,
		Capacity: req.Capacity,
	}
	if err := s.Repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetAvailableRooms(startTime, endTime time.Time) ([]model.AvailableRoomResponse, error) {
	if !endTime.After(startTime) {
		return nil, ErrInvalidTimeRange
	}
	return s.Repo.FindAvailable(startTime.UTC(), endTime.UTC())
}
