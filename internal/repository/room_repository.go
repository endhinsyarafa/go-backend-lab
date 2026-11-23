package repository

import (
	"database/sql"
	"time"

	"github.com/latihan/go-backend-lab/internal/model"
)

type RoomRepository struct {
	DB *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{DB: db}
}

func (r *RoomRepository) Create(room *model.Room) error {
	query := `
		INSERT INTO rooms (name, capacity)
		VALUES ($1, $2)
		RETURNING id, is_active
	`
	return r.DB.QueryRow(query, room.Name, room.Capacity).Scan(&room.ID, &room.IsActive)
}

func (r *RoomRepository) FindAvailable(startTime, endTime time.Time) ([]model.AvailableRoomResponse, error) {
	query := `
		SELECT r.id, r.name, r.capacity
		FROM rooms r
		WHERE r.is_active = TRUE
		  AND NOT EXISTS (
			SELECT 1
			FROM bookings b
			WHERE b.room_id = r.id
			  AND b.status = 'CONFIRMED'
			  AND b.start_time < $2
			  AND b.end_time > $1
		  )
		ORDER BY r.id ASC
	`
	rows, err := r.DB.Query(query, startTime.UTC(), endTime.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]model.AvailableRoomResponse, 0)
	for rows.Next() {
		var rm model.AvailableRoomResponse
		if err := rows.Scan(&rm.ID, &rm.Name, &rm.Capacity); err != nil {
			return nil, err
		}
		rooms = append(rooms, rm)
	}
	return rooms, rows.Err()
}
