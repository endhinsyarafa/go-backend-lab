package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/latihan/go-backend-lab/internal/config"
	"github.com/latihan/go-backend-lab/internal/handler"
	"github.com/latihan/go-backend-lab/internal/middleware"
	"github.com/latihan/go-backend-lab/internal/repository"
	"github.com/latihan/go-backend-lab/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatal("Gagal membuka koneksi DB:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Gagal ping DB:", err)
	}

	// Dependency Injection
	authHandler := &handler.AuthHandler{
		DB:        db,
		JWTSecret: cfg.JWTSecret,
	}

	roomRepo := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepo)
	roomHandler := handler.NewRoomHandler(roomService)

	bookingRepo := repository.NewBookingRepository(db)
	bookingService := service.NewBookingService(bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		// Public
		v1.POST("/auth/login", authHandler.Login)

		// Protected (ADMIN & USER)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// 1. Room Management (khusus ADMIN)
			protected.POST("/rooms", middleware.RequireRole("ADMIN"), roomHandler.CreateRoom)

			// 2. Room Availability
			protected.GET("/rooms/available", roomHandler.GetAvailableRooms)

			// 3. Booking Management
			protected.POST("/bookings", bookingHandler.CreateBooking)

			// 4. Booking Cancellation
			protected.PATCH("/bookings/:id/cancel", bookingHandler.CancelBooking)
		}
	}

	log.Printf("Server running on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
