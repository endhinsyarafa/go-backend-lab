package service_test

import (
	"testing"
	"time"

	"github.com/latihan/go-backend-lab/internal/model"
	"github.com/latihan/go-backend-lab/internal/service"
)

func TestBookingService_InvalidTime(t *testing.T) {
	svc := service.NewBookingService(nil)

	now := time.Now().UTC()

	tests := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
	}{
		{
			name:      "Start time di masa lalu",
			startTime: now.Add(-2 * time.Hour),
			endTime:   now.Add(-1 * time.Hour),
		},
		{
			name:      "End time sama dengan Start time",
			startTime: now.Add(2 * time.Hour),
			endTime:   now.Add(2 * time.Hour),
		},
		{
			name:      "End time sebelum Start time",
			startTime: now.Add(3 * time.Hour),
			endTime:   now.Add(2 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := model.CreateBookingRequest{
				RoomID:    1,
				Title:     "Test Invalid Time",
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			_, err := svc.CreateBooking(1, req)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}
		})
	}
}
