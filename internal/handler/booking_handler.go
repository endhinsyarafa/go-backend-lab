package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latihan/go-backend-lab/internal/model"
	"github.com/latihan/go-backend-lab/internal/service"
)

type BookingHandler struct {
	Service *service.BookingService
}

func NewBookingHandler(svc *service.BookingService) *BookingHandler {
	return &BookingHandler{Service: svc}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req model.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  "INVALID_REQUEST",
			"error": "room_id, title, start_time, and end_time are required",
		})
		return
	}

	userIDFloat, _ := c.MustGet("user_id").(float64)
	userID := int(userIDFloat)

	booking, err := h.Service.CreateBooking(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTime):
			c.JSON(http.StatusBadRequest, gin.H{
				"error_code": "INVALID_TIME",
				"message":    "Waktu mulai harus di masa depan dan waktu selesai harus setelah waktu mulai",
			})
		case errors.Is(err, service.ErrRoomNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error_code": "ROOM_NOT_FOUND",
				"message":    "Ruangan tidak ditemukan",
			})
		case errors.Is(err, service.ErrRoomInactive):
			c.JSON(http.StatusBadRequest, gin.H{
				"error_code": "ROOM_INACTIVE",
				"message":    "Ruangan tidak aktif",
			})
		case errors.Is(err, service.ErrBookingConflict):
			c.JSON(http.StatusConflict, gin.H{
				"error_code": "ROOM_CONFLICT",
				"message":    "Ruangan sudah dipesan pada rentang waktu tersebut",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  "INTERNAL_ERROR",
				"error": "failed to create booking",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingID, err := strconv.Atoi(c.Param("id"))
	if err != nil || bookingID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	userIDFloat, _ := c.MustGet("user_id").(float64)
	userID := int(userIDFloat)
	role, _ := c.MustGet("role").(string)

	booking, err := h.Service.CancelBooking(bookingID, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbiddenCancel):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrAlreadyCancelled):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel booking"})
		}
		return
	}

	c.JSON(http.StatusOK, booking)
}
