package repository

import (
	"database/sql"
	"errors"

	"github.com/latihan/go-backend-lab/internal/model"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrRoomInactive    = errors.New("room is inactive")
	ErrBookingConflict = errors.New("room is already booked for the selected time range")
)

type BookingRepository struct {
	DB *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{DB: db}
}

func (r *BookingRepository) CreateWithLock(b *model.Booking) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var isActive bool
	err = tx.QueryRow("SELECT is_active FROM rooms WHERE id = $1 FOR UPDATE", b.RoomID).Scan(&isActive)
	if err == sql.ErrNoRows {
		return ErrRoomNotFound
	}
	if err != nil {
		return err
	}
	if !isActive {
		return ErrRoomInactive
	}

	var conflictCount int
	conflictQuery := `
		SELECT COUNT(1)
		FROM bookings
		WHERE room_id = $1
		  AND status = 'CONFIRMED'
		  AND start_time < $3
		  AND end_time > $2
	`
	if err := tx.QueryRow(conflictQuery, b.RoomID, b.StartTime.UTC(), b.EndTime.UTC()).Scan(&conflictCount); err != nil {
		return err
	}
	if conflictCount > 0 {
		return ErrBookingConflict
	}

	insertQuery := `
		INSERT INTO bookings (room_id, user_id, title, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5, 'CONFIRMED')
		RETURNING id, status
	`
	if err := tx.QueryRow(
		insertQuery,
		b.RoomID,
		b.UserID,
		b.Title,
		b.StartTime.UTC(),
		b.EndTime.UTC(),
	).Scan(&b.ID, &b.Status); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BookingRepository) FindByID(id int) (*model.Booking, error) {
	var b model.Booking
	query := `
		SELECT id, room_id, user_id, title, start_time, end_time, status
		FROM bookings
		WHERE id = $1
	`
	err := r.DB.QueryRow(query, id).Scan(&b.ID, &b.RoomID, &b.UserID, &b.Title, &b.StartTime, &b.EndTime, &b.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func (r *BookingRepository) Cancel(id int) error {
	_, err := r.DB.Exec("UPDATE bookings SET status = 'CANCELLED' WHERE id = $1", id)
	return err
}
