package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/repositories"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/services"
)

// TicketHandler handles HTTP endpoints for searching trips, viewing seat maps, and booking tickets
type TicketHandler struct {
	service services.TicketService
}

// NewTicketHandler instantiates TicketHandler
func NewTicketHandler(service services.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

// RegisterRoutes registers public ticketing routes
func (h *TicketHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Trips search and inspection
	rg.GET("/trips/search", h.SearchTrips)
	rg.GET("/trips/:id", h.GetTripByID)
	rg.GET("/trips/:id/seats", h.GetSeatMap)
	rg.GET("/trips/:id/tickets", h.GetTicketsByTrip)

	// Booking and tickets
	rg.POST("/tickets/book", h.BookSeat)
	rg.POST("/tickets/book-multiple", h.BookMultipleSeats)
	rg.GET("/tickets/:id", h.GetTicketByID)
}

// SearchTrips handles trip search by origin, destination and optional date
func (h *TicketHandler) SearchTrips(c *gin.Context) {
	originID, err := strconv.Atoi(c.Query("origin_id"))
	if err != nil || originID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "origin_id query parameter is required and must be a positive integer"})
		return
	}

	destID, err := strconv.Atoi(c.Query("destination_id"))
	if err != nil || destID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "destination_id query parameter is required and must be a positive integer"})
		return
	}

	date := c.Query("date")

	trips, err := h.service.SearchTrips(c.Request.Context(), originID, destID, date)
	if err != nil {
		if errors.Is(err, services.ErrIdenticalSearchCities) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": trips})
}

// GetTripByID returns detailed trip information
func (h *TicketHandler) GetTripByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	trip, err := h.service.GetTripByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrTripNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": trip})
}

// GetSeatMap returns real-time seat availability for the interactive bus layout
func (h *TicketHandler) GetSeatMap(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	seats, err := h.service.GetSeatMap(c.Request.Context(), tripID)
	if err != nil {
		if errors.Is(err, repositories.ErrTripNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": seats})
}

// BookSeat handles concurrent seat purchasing with race condition protection
func (h *TicketHandler) BookSeat(c *gin.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.service.BookSeat(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repositories.ErrSeatAlreadyOccupied) {
			// HTTP 409 Conflict: another user reserved/bought the seat first!
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "seat is already occupied for this trip",
				"code":    "SEAT_UNAVAILABLE",
			})
			return
		}
		if errors.Is(err, repositories.ErrSeatNotFound) || errors.Is(err, repositories.ErrTripNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrInvalidPassengerName) || errors.Is(err, services.ErrInvalidDocument) || errors.Is(err, services.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "ticket booked successfully",
		"data":    ticket,
	})
}

// BookMultipleSeats handles atomic multi-seat ticket booking with race condition protection
func (h *TicketHandler) BookMultipleSeats(c *gin.Context) {
	var req models.MultiBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.BookMultipleSeats(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repositories.ErrSeatAlreadyOccupied) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   err.Error(),
				"code":    "SEAT_UNAVAILABLE",
			})
			return
		}
		if errors.Is(err, repositories.ErrSeatNotFound) || errors.Is(err, repositories.ErrTripNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrInvalidPassengerName) || errors.Is(err, services.ErrInvalidDocument) || errors.Is(err, services.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "all tickets booked successfully",
		"data":    response,
	})
}


// GetTicketByID returns ticket confirmation
func (h *TicketHandler) GetTicketByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	ticket, err := h.service.GetTicketByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": ticket})
}

// GetTicketsByTrip returns all tickets for a trip
func (h *TicketHandler) GetTicketsByTrip(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	tickets, err := h.service.GetTicketsByTrip(c.Request.Context(), tripID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": tickets})
}
