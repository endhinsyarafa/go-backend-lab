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

	router.StaticFile("/swagger.yaml", "./swagger.yaml")

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	router.GET("/swagger/index.html", func(c *gin.Context) {
		html := `<!DOCTYPE html>
			<html lang="en">
			<head>
			<meta charset="UTF-8">
			<title>Meeting Room Booking API - Swagger UI</title>
			<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
			</head>
			<body>
			<div id="swagger-ui"></div>
			<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
			<script>
				window.onload = () => {
				SwaggerUIBundle({
					url: '/swagger.yaml',
					dom_id: '#swagger-ui',
				});
				};
			</script>
			</body>
			</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/login", authHandler.Login)

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.POST("/rooms", middleware.RequireRole("ADMIN"), roomHandler.CreateRoom)
			protected.GET("/rooms/available", roomHandler.GetAvailableRooms)
			protected.POST("/bookings", bookingHandler.CreateBooking)
			protected.PATCH("/bookings/:id/cancel", bookingHandler.CancelBooking)
		}
	}

	log.Printf("Server running on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
