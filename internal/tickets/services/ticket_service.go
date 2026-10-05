package services

import (
	"context"
	"errors"
	"strings"

	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/tickets/repositories"
)

var (
	ErrInvalidPassengerName = errors.New("passenger name is required")
	ErrInvalidDocument      = errors.New("document is required")
	ErrInvalidStatus        = errors.New("ticket status must be RESERVADO or PAGADO")
	ErrIdenticalSearchCities = errors.New("origin and destination cities cannot be identical")
)

// TicketService defines business operations for public search and concurrent ticket booking
type TicketService interface {
	SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error)
	GetTripByID(ctx context.Context, tripID int) (*models.Trip, error)
	GetSeatMap(ctx context.Context, tripID int) ([]models.SeatMapStatus, error)
	BookSeat(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error)
	BookMultipleSeats(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error)
	GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error)
	GetTicketsByTrip(ctx context.Context, tripID int) ([]models.Ticket, error)
}

type ticketService struct {
	repo repositories.TicketRepository
}

// NewTicketService instantiates TicketService with repository dependency
func NewTicketService(repo repositories.TicketRepository) TicketService {
	return &ticketService{repo: repo}
}

func (s *ticketService) SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error) {
	if originCityID <= 0 || destCityID <= 0 {
		return nil, errors.New("valid origin and destination city IDs are required")
	}
	if originCityID == destCityID {
		return nil, ErrIdenticalSearchCities
	}
	return s.repo.SearchTrips(ctx, originCityID, destCityID, strings.TrimSpace(date))
}

func (s *ticketService) GetTripByID(ctx context.Context, tripID int) (*models.Trip, error) {
	if tripID <= 0 {
		return nil, errors.New("invalid trip ID")
	}
	return s.repo.GetTripByID(ctx, tripID)
}

func (s *ticketService) GetSeatMap(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
	if tripID <= 0 {
		return nil, errors.New("invalid trip ID")
	}
	// Verify trip exists
	if _, err := s.repo.GetTripByID(ctx, tripID); err != nil {
		return nil, err
	}
	return s.repo.GetSeatMapByTripID(ctx, tripID)
}

func (s *ticketService) BookSeat(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
	req.PassengerName = strings.TrimSpace(req.PassengerName)
	req.DocumentID = strings.TrimSpace(req.DocumentID)
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))

	if req.TripID <= 0 {
		return nil, errors.New("invalid trip ID")
	}
	if req.SeatID <= 0 {
		return nil, errors.New("invalid seat ID")
	}
	if req.PassengerName == "" {
		return nil, ErrInvalidPassengerName
	}
	if req.DocumentID == "" {
		return nil, ErrInvalidDocument
	}

	if req.Status == "" {
		req.Status = models.TicketStatusPaid
	} else if req.Status != models.TicketStatusReserved && req.Status != models.TicketStatusPaid {
		return nil, ErrInvalidStatus
	}

	// Transactional booking with pessimistic lock
	ticket, err := s.repo.BookSeatWithLock(ctx, req)
	if err != nil {
		return nil, err
	}

	// Load enriched details
	return s.repo.GetTicketByID(ctx, ticket.ID)
}

func (s *ticketService) BookMultipleSeats(ctx context.Context, req *models.MultiBookingRequest) (*models.MultiBookingResponse, error) {
	if req.TripID <= 0 {
		return nil, errors.New("invalid trip ID")
	}
	if len(req.Passengers) == 0 {
		return nil, errors.New("at least one passenger seat must be selected")
	}

	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	if req.Status == "" {
		req.Status = models.TicketStatusPaid
	} else if req.Status != models.TicketStatusReserved && req.Status != models.TicketStatusPaid {
		return nil, ErrInvalidStatus
	}

	seenSeats := make(map[int]bool)
	for i := range req.Passengers {
		p := &req.Passengers[i]
		p.PassengerName = strings.TrimSpace(p.PassengerName)
		p.DocumentID = strings.TrimSpace(p.DocumentID)

		if p.SeatID <= 0 {
			return nil, errors.New("invalid seat ID")
		}
		if seenSeats[p.SeatID] {
			return nil, errors.New("duplicate seat assigned to multiple passengers")
		}
		seenSeats[p.SeatID] = true

		if p.PassengerName == "" {
			return nil, ErrInvalidPassengerName
		}
		if p.DocumentID == "" {
			return nil, ErrInvalidDocument
		}
	}

	return s.repo.BookMultipleSeatsWithLock(ctx, req)
}


func (s *ticketService) GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error) {
	if ticketID <= 0 {
		return nil, errors.New("invalid ticket ID")
	}
	return s.repo.GetTicketByID(ctx, ticketID)
}

func (s *ticketService) GetTicketsByTrip(ctx context.Context, tripID int) ([]models.Ticket, error) {
	if tripID <= 0 {
		return nil, errors.New("invalid trip ID")
	}
	return s.repo.GetTicketsByTripID(ctx, tripID)
}
