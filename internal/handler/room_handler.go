package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/latihan/go-backend-lab/internal/model"
	"github.com/latihan/go-backend-lab/internal/service"
)

type RoomHandler struct {
	Service *service.RoomService
}

func NewRoomHandler(svc *service.RoomService) *RoomHandler {
	return &RoomHandler{Service: svc}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req model.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.Service.CreateRoom(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func (h *RoomHandler) GetAvailableRooms(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query parameters are required (ISO-8601 UTC)"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start format, use ISO-8601 UTC (e.g. 2026-11-24T09:00:00Z)"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end format, use ISO-8601 UTC (e.g. 2026-11-24T11:00:00Z)"})
		return
	}

	rooms, err := h.Service.GetAvailableRooms(startTime, endTime)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch available rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rooms,
	})
}
