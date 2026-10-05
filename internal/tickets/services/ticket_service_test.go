package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/repositories"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/services"
)

// MockTicketRepository implements repositories.TicketRepository for unit testing
type MockTicketRepository struct {
	SearchTripsFunc        func(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error)
	GetTripByIDFunc        func(ctx context.Context, tripID int) (*models.Trip, error)
	GetSeatMapByTripIDFunc        func(ctx context.Context, tripID int) ([]models.SeatMapStatus, error)
	BookSeatWithLockFunc          func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error)
	BookMultipleSeatsWithLockFunc func(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error)
	GetTicketByIDFunc             func(ctx context.Context, ticketID int) (*models.Ticket, error)
	GetTicketsByTripIDFunc func(ctx context.Context, tripID int) ([]models.Ticket, error)
}

func (m *MockTicketRepository) SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error) {
	if m.SearchTripsFunc != nil {
		return m.SearchTripsFunc(ctx, originCityID, destCityID, date)
	}
	return nil, nil
}

func (m *MockTicketRepository) GetTripByID(ctx context.Context, tripID int) (*models.Trip, error) {
	if m.GetTripByIDFunc != nil {
		return m.GetTripByIDFunc(ctx, tripID)
	}
	return &models.Trip{ID: tripID, TicketPrice: 85000}, nil
}

func (m *MockTicketRepository) GetSeatMapByTripID(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
	if m.GetSeatMapByTripIDFunc != nil {
		return m.GetSeatMapByTripIDFunc(ctx, tripID)
	}
	return nil, nil
}

func (m *MockTicketRepository) BookSeatWithLock(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
	if m.BookSeatWithLockFunc != nil {
		return m.BookSeatWithLockFunc(ctx, req)
	}
	return &models.Ticket{
		ID:            101,
		TripID:        req.TripID,
		SeatID:        req.SeatID,
		PassengerName: req.PassengerName,
		DocumentID:    req.DocumentID,
		Status:        req.Status,
		CreatedAt:     time.Now(),
	}, nil
}

func (m *MockTicketRepository) BookMultipleSeatsWithLock(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error) {
	if m.BookMultipleSeatsWithLockFunc != nil {
		return m.BookMultipleSeatsWithLockFunc(ctx, req)
	}
	tickets := make([]models.Ticket, len(req.Passengers))
	for i, p := range req.Passengers {
		tickets[i] = models.Ticket{
			ID:            i + 1,
			TripID:        req.TripID,
			SeatID:        p.SeatID,
			PassengerName: p.PassengerName,
			DocumentID:    p.DocumentID,
			Status:        req.Status,
		}
	}
	return &models.MultiBookingResponse{
		TripID:         req.TripID,
		Tickets:        tickets,
		TotalAmount:    float64(len(tickets)) * 85000,
		PassengerCount: len(tickets),
	}, nil
}


func (m *MockTicketRepository) GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error) {
	if m.GetTicketByIDFunc != nil {
		return m.GetTicketByIDFunc(ctx, ticketID)
	}
	return &models.Ticket{
		ID:            ticketID,
		TripID:        1,
		SeatID:        2,
		PassengerName: "Test Passenger",
		DocumentID:    "CC123456",
		Status:        models.TicketStatusPaid,
	}, nil
}

func (m *MockTicketRepository) GetTicketsByTripID(ctx context.Context, tripID int) ([]models.Ticket, error) {
	if m.GetTicketsByTripIDFunc != nil {
		return m.GetTicketsByTripIDFunc(ctx, tripID)
	}
	return nil, nil
}

// ==========================================
// UNIT TESTS
// ==========================================

func TestBookSeat_Success(t *testing.T) {
	mockRepo := &MockTicketRepository{
		BookSeatWithLockFunc: func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
			return &models.Ticket{
				ID:            77,
				TripID:        req.TripID,
				SeatID:        req.SeatID,
				PassengerName: req.PassengerName,
				DocumentID:    req.DocumentID,
				Status:        req.Status,
			}, nil
		},
		GetTicketByIDFunc: func(ctx context.Context, ticketID int) (*models.Ticket, error) {
			return &models.Ticket{
				ID:            ticketID,
				TripID:        1,
				SeatID:        3,
				PassengerName: "Andrea Gomez",
				DocumentID:    "CC-998877",
				Status:        models.TicketStatusPaid,
			}, nil
		},
	}

	service := services.NewTicketService(mockRepo)

	req := &models.BookingRequest{
		TripID:        1,
		SeatID:        3,
		PassengerName: "Andrea Gomez",
		DocumentID:    "CC-998877",
		Status:        models.TicketStatusPaid,
	}

	ticket, err := service.BookSeat(context.Background(), req)
	if err != nil {
		t.Fatalf("expected successful booking, got error: %v", err)
	}

	if ticket.ID != 77 && ticket.ID != 1 {
		t.Errorf("expected valid ticket ID, got: %d", ticket.ID)
	}
	if ticket.PassengerName != "Andrea Gomez" {
		t.Errorf("expected passenger 'Andrea Gomez', got: %s", ticket.PassengerName)
	}
}

func TestBookSeat_SeatAlreadyOccupied(t *testing.T) {
	mockRepo := &MockTicketRepository{
		BookSeatWithLockFunc: func(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
			// Simulates database rejection due to seat already taken
			return nil, repositories.ErrSeatAlreadyOccupied
		},
	}

	service := services.NewTicketService(mockRepo)

	req := &models.BookingRequest{
		TripID:        1,
		SeatID:        2,
		PassengerName: "Late Passenger",
		DocumentID:    "CC-000000",
		Status:        models.TicketStatusPaid,
	}

	ticket, err := service.BookSeat(context.Background(), req)
	if ticket != nil {
		t.Errorf("expected nil ticket, got: %+v", ticket)
	}
	if !errors.Is(err, repositories.ErrSeatAlreadyOccupied) {
		t.Fatalf("expected ErrSeatAlreadyOccupied, got: %v", err)
	}
}

func TestBookSeat_ValidationFailures(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	service := services.NewTicketService(mockRepo)

	// Missing passenger name
	req1 := &models.BookingRequest{
		TripID:        1,
		SeatID:        1,
		PassengerName: "  ",
		DocumentID:    "CC123",
	}
	_, err1 := service.BookSeat(context.Background(), req1)
	if !errors.Is(err1, services.ErrInvalidPassengerName) {
		t.Errorf("expected ErrInvalidPassengerName, got: %v", err1)
	}

	// Missing document
	req2 := &models.BookingRequest{
		TripID:        1,
		SeatID:        1,
		PassengerName: "John Doe",
		DocumentID:    "",
	}
	_, err2 := service.BookSeat(context.Background(), req2)
	if !errors.Is(err2, services.ErrInvalidDocument) {
		t.Errorf("expected ErrInvalidDocument, got: %v", err2)
	}

	// Invalid status
	req3 := &models.BookingRequest{
		TripID:        1,
		SeatID:        1,
		PassengerName: "John Doe",
		DocumentID:    "CC123",
		Status:        "UNKNOWN_STATUS",
	}
	_, err3 := service.BookSeat(context.Background(), req3)
	if !errors.Is(err3, services.ErrInvalidStatus) {
		t.Errorf("expected ErrInvalidStatus, got: %v", err3)
	}
}

func TestSearchTrips_IdenticalCities(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	service := services.NewTicketService(mockRepo)

	_, err := service.SearchTrips(context.Background(), 1, 1, "2026-10-10")
	if !errors.Is(err, services.ErrIdenticalSearchCities) {
		t.Fatalf("expected ErrIdenticalSearchCities, got: %v", err)
	}
}

func TestGetSeatMap_Success(t *testing.T) {
	paidStatus := "PAGADO"
	passenger := "Carlos"

	mockRepo := &MockTicketRepository{
		GetTripByIDFunc: func(ctx context.Context, tripID int) (*models.Trip, error) {
			return &models.Trip{ID: tripID}, nil
		},
		GetSeatMapByTripIDFunc: func(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
			return []models.SeatMapStatus{
				{SeatID: 1, SeatNumber: 1, Row: 1, Column: 1, Available: false, TicketStatus: &paidStatus, PassengerName: &passenger},
				{SeatID: 2, SeatNumber: 2, Row: 1, Column: 2, Available: true},
			}, nil
		},
	}

	service := services.NewTicketService(mockRepo)
	seats, err := service.GetSeatMap(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(seats) != 2 {
		t.Fatalf("expected 2 seats, got: %d", len(seats))
	}
	if seats[0].Available != false || seats[1].Available != true {
		t.Errorf("seat availability mismatch in returned seat map")
	}
}

func TestGetSeatMap_TripNotFound(t *testing.T) {
	mockRepo := &MockTicketRepository{
		GetTripByIDFunc: func(ctx context.Context, tripID int) (*models.Trip, error) {
			return nil, repositories.ErrTripNotFound
		},
	}

	service := services.NewTicketService(mockRepo)
	_, err := service.GetSeatMap(context.Background(), 999)
	if !errors.Is(err, repositories.ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound, got: %v", err)
	}
}

func TestBookMultipleSeats_Success(t *testing.T) {
	mockRepo := &MockTicketRepository{
		BookMultipleSeatsWithLockFunc: func(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error) {
			tickets := make([]models.Ticket, len(req.Passengers))
			for i, p := range req.Passengers {
				tickets[i] = models.Ticket{
					ID:            i + 1,
					TripID:        req.TripID,
					SeatID:        p.SeatID,
					PassengerName: p.PassengerName,
					DocumentID:    p.DocumentID,
					Status:        req.Status,
				}
			}
			return &models.MultiBookingResponse{
				TripID:         req.TripID,
				Tickets:        tickets,
				TotalAmount:    255000,
				PassengerCount: len(tickets),
			}, nil
		},
	}

	service := services.NewTicketService(mockRepo)
	req := &models.MultiBookingRequest{
		TripID: 1,
		Passengers: []models.PassengerItem{
			{SeatID: 1, PassengerName: "Adult 1", DocumentID: "DOC-1", PassengerType: "ADULT"},
			{SeatID: 2, PassengerName: "Adult 2", DocumentID: "DOC-2", PassengerType: "ADULT"},
			{SeatID: 3, PassengerName: "Child 1", DocumentID: "DOC-3", PassengerType: "CHILD"},
		},
		Status: "PAGADO",
	}

	resp, err := service.BookMultipleSeats(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.PassengerCount != 3 {
		t.Errorf("expected 3 passengers, got: %d", resp.PassengerCount)
	}
	if len(resp.Tickets) != 3 {
		t.Errorf("expected 3 tickets, got: %d", len(resp.Tickets))
	}
}

func TestBookMultipleSeats_BatchConflict(t *testing.T) {
	mockRepo := &MockTicketRepository{
		BookMultipleSeatsWithLockFunc: func(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error) {
			// Simulates rejection when even 1 seat is taken
			return nil, repositories.ErrSeatAlreadyOccupied
		},
	}

	service := services.NewTicketService(mockRepo)
	req := &models.MultiBookingRequest{
		TripID: 1,
		Passengers: []models.PassengerItem{
			{SeatID: 1, PassengerName: "Adult 1", DocumentID: "DOC-1"},
			{SeatID: 2, PassengerName: "Adult 2", DocumentID: "DOC-2"},
		},
	}

	_, err := service.BookMultipleSeats(context.Background(), req)
	if !errors.Is(err, repositories.ErrSeatAlreadyOccupied) {
		t.Fatalf("expected ErrSeatAlreadyOccupied, got: %v", err)
	}
}

func TestBookMultipleSeats_DuplicateSeatInRequest(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	service := services.NewTicketService(mockRepo)

	req := &models.MultiBookingRequest{
		TripID: 1,
		Passengers: []models.PassengerItem{
			{SeatID: 5, PassengerName: "Adult 1", DocumentID: "DOC-1"},
			{SeatID: 5, PassengerName: "Adult 2", DocumentID: "DOC-2"}, // Duplicate seat ID in same batch
		},
	}

	_, err := service.BookMultipleSeats(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for duplicate seat in same request, got nil")
	}
}

func TestBookMultipleSeats_EmptyPassengers(t *testing.T) {
	mockRepo := &MockTicketRepository{}
	service := services.NewTicketService(mockRepo)

	req := &models.MultiBookingRequest{
		TripID:     1,
		Passengers: []models.PassengerItem{},
	}

	_, err := service.BookMultipleSeats(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for empty passengers, got nil")
	}
}

