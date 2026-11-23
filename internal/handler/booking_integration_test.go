package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/latihan/go-backend-lab/internal/config"
	"github.com/latihan/go-backend-lab/internal/handler"
	"github.com/latihan/go-backend-lab/internal/middleware"
	"github.com/latihan/go-backend-lab/internal/repository"
	"github.com/latihan/go-backend-lab/internal/service"
)

func setupTestEnv(t *testing.T) (*gin.Engine, *sql.DB, string, int) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	_ = godotenv.Load("../../.env")

	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		t.Fatalf("gagal membuka koneksi DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("gagal ping DB: %v", err)
	}

	var testRoomID int
	err = db.QueryRow(
		"INSERT INTO rooms (name, capacity, is_active) VALUES ($1, $2, TRUE) RETURNING id",
		fmt.Sprintf("Test Room %d", time.Now().UnixNano()),
		10,
	).Scan(&testRoomID)
	if err != nil {
		t.Fatalf("gagal membuat test room: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM bookings WHERE room_id = $1", testRoomID)
		_, _ = db.Exec("DELETE FROM rooms WHERE id = $1", testRoomID)
		_ = db.Close()
	})

	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  1,
		"role": "ADMIN",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := tokenObj.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("gagal membuat JWT: %v", err)
	}

	bookingRepo := repository.NewBookingRepository(db)
	bookingSvc := service.NewBookingService(bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc)

	r := gin.New()
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.POST("/bookings", bookingHandler.CreateBooking)

	return r, db, tokenStr, testRoomID
}

func TestBookingIntegrationAndConcurrency(t *testing.T) {
	router, db, token, roomID := setupTestEnv(t)

	start1 := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	end1 := start1.Add(2 * time.Hour)

	payload1, _ := json.Marshal(map[string]any{
		"room_id":    roomID,
		"title":      "Integration Test Booking",
		"start_time": start1.Format(time.RFC3339),
		"end_time":   end1.Format(time.RFC3339),
	})

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(payload1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("expected 201 for first booking, got %d: %s", w1.Code, w1.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(payload1))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate booking, got %d: %s", w2.Code, w2.Body.String())
	}

	startConc := start1.Add(24 * time.Hour)
	endConc := startConc.Add(2 * time.Hour)

	payloadConc, _ := json.Marshal(map[string]any{
		"room_id":    roomID,
		"title":      "Concurrency Test Booking",
		"start_time": startConc.Format(time.RFC3339),
		"end_time":   endConc.Format(time.RFC3339),
	})

	const totalReq = 10
	var wg sync.WaitGroup
	results := make(chan int, totalReq)

	for i := 0; i < totalReq; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(payloadConc))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for code := range results {
		switch code {
		case http.StatusCreated:
			successCount++
		case http.StatusConflict:
			conflictCount++
		default:
			t.Errorf("unexpected status code: %d", code)
		}
	}

	if successCount != 1 || conflictCount != 9 {
		t.Fatalf("expected 1 success and 9 conflicts, got %d success and %d conflicts", successCount, conflictCount)
	}

	var confirmedCount int
	err := db.QueryRow("SELECT COUNT(*) FROM bookings WHERE room_id = $1 AND status = 'CONFIRMED'", roomID).Scan(&confirmedCount)
	if err != nil {
		t.Fatalf("gagal menghitung booking CONFIRMED: %v", err)
	}
	if confirmedCount != 2 {
		t.Fatalf("expected 2 CONFIRMED bookings in DB, got %d", confirmedCount)
	}
}
