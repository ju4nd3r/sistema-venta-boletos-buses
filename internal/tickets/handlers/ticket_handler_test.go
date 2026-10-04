package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
	ticketHandlers "github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/handlers"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/repositories"
)

// MockTicketService implements services.TicketService for handler testing
type MockTicketService struct {
	SearchTripsFunc      func(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error)
	GetTripByIDFunc      func(ctx context.Context, tripID int) (*models.Trip, error)
	GetSeatMapFunc       func(ctx context.Context, tripID int) ([]models.SeatMapStatus, error)
	BookSeatFunc         func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error)
	GetTicketByIDFunc    func(ctx context.Context, ticketID int) (*models.Ticket, error)
	GetTicketsByTripFunc func(ctx context.Context, tripID int) ([]models.Ticket, error)
}

func (m *MockTicketService) SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error) {
	if m.SearchTripsFunc != nil {
		return m.SearchTripsFunc(ctx, originCityID, destCityID, date)
	}
	return nil, nil
}
func (m *MockTicketService) GetTripByID(ctx context.Context, tripID int) (*models.Trip, error) {
	if m.GetTripByIDFunc != nil {
		return m.GetTripByIDFunc(ctx, tripID)
	}
	return &models.Trip{ID: tripID}, nil
}
func (m *MockTicketService) GetSeatMap(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
	if m.GetSeatMapFunc != nil {
		return m.GetSeatMapFunc(ctx, tripID)
	}
	return []models.SeatMapStatus{{SeatID: 1, Available: true}}, nil
}
func (m *MockTicketService) BookSeat(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
	if m.BookSeatFunc != nil {
		return m.BookSeatFunc(ctx, req)
	}
	return &models.Ticket{ID: 1, TripID: req.TripID, SeatID: req.SeatID}, nil
}
func (m *MockTicketService) GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error) {
	if m.GetTicketByIDFunc != nil {
		return m.GetTicketByIDFunc(ctx, ticketID)
	}
	return &models.Ticket{ID: ticketID}, nil
}
func (m *MockTicketService) GetTicketsByTrip(ctx context.Context, tripID int) ([]models.Ticket, error) {
	if m.GetTicketsByTripFunc != nil {
		return m.GetTicketsByTripFunc(ctx, tripID)
	}
	return nil, nil
}

func setupTestRouter(handler *ticketHandlers.TicketHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rg := router.Group("/api")
	handler.RegisterRoutes(rg)
	return router
}

func TestHandler_BookSeat_Success(t *testing.T) {
	mockService := &MockTicketService{
		BookSeatFunc: func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
			return &models.Ticket{
				ID:            55,
				TripID:        req.TripID,
				SeatID:        req.SeatID,
				PassengerName: req.PassengerName,
				DocumentID:    req.DocumentID,
				Status:        models.TicketStatusPaid,
			}, nil
		},
	}

	handler := ticketHandlers.NewTicketHandler(mockService)
	router := setupTestRouter(handler)

	reqBody := models.BookingRequest{
		TripID:        1,
		SeatID:        2,
		PassengerName: "Lucia Santos",
		DocumentID:    "CC554433",
		Status:        "PAGADO",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/tickets/book", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 Created, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_BookSeat_ConflictSeatOccupied(t *testing.T) {
	mockService := &MockTicketService{
		BookSeatFunc: func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
			return nil, repositories.ErrSeatAlreadyOccupied
		},
	}

	handler := ticketHandlers.NewTicketHandler(mockService)
	router := setupTestRouter(handler)

	reqBody := models.BookingRequest{
		TripID:        1,
		SeatID:        2,
		PassengerName: "User 2",
		DocumentID:    "CC999",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/tickets/book", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected HTTP 409 Conflict, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetSeatMap_Success(t *testing.T) {
	mockService := &MockTicketService{
		GetSeatMapFunc: func(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
			return []models.SeatMapStatus{
				{SeatID: 1, SeatNumber: 1, Available: true},
				{SeatID: 2, SeatNumber: 2, Available: false},
			}, nil
		},
	}

	handler := ticketHandlers.NewTicketHandler(mockService)
	router := setupTestRouter(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/trips/1/seats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
	}
}
