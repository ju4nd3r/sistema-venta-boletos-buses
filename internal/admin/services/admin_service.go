package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/repositories"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
)

var (
	ErrIdenticalCities = errors.New("origin and destination cities cannot be identical")
	ErrDuplicateRoute  = errors.New("a route between these cities already exists")
	ErrInvalidCapacity = errors.New("seat capacity must be greater than zero")
	ErrInvalidDistance = errors.New("distance must be greater than zero")
	ErrInvalidDuration = errors.New("estimated duration must be greater than zero")
	ErrInvalidPrice    = errors.New("ticket price cannot be negative")
)

// AdminService defines business operations for the administrative panel
type AdminService interface {
	// City operations
	CreateCity(ctx context.Context, city *models.City) error
	GetAllCities(ctx context.Context) ([]models.City, error)
	GetCityByID(ctx context.Context, id int) (*models.City, error)
	UpdateCity(ctx context.Context, city *models.City) error
	DeleteCity(ctx context.Context, id int) error

	// Route operations
	CreateRoute(ctx context.Context, route *models.Route) error
	GetAllRoutes(ctx context.Context) ([]models.Route, error)
	GetRouteByID(ctx context.Context, id int) (*models.Route, error)
	UpdateRoute(ctx context.Context, route *models.Route) error
	DeleteRoute(ctx context.Context, id int) error

	// BusType operations
	CreateBusType(ctx context.Context, busType *models.BusType) error
	GetAllBusTypes(ctx context.Context) ([]models.BusType, error)
	GetBusTypeByID(ctx context.Context, id int) (*models.BusType, error)
	UpdateBusType(ctx context.Context, busType *models.BusType) error
	DeleteBusType(ctx context.Context, id int) error

	// Bus operations
	CreateBus(ctx context.Context, bus *models.Bus) error
	GetAllBuses(ctx context.Context) ([]models.Bus, error)
	GetBusByID(ctx context.Context, id int) (*models.Bus, error)
	UpdateBus(ctx context.Context, bus *models.Bus) error
	DeleteBus(ctx context.Context, id int) error

	// Trip operations
	CreateTrip(ctx context.Context, trip *models.Trip) error
	GetAllTrips(ctx context.Context) ([]models.Trip, error)
	GetTripByID(ctx context.Context, id int) (*models.Trip, error)
	DeleteTrip(ctx context.Context, id int) error
}

type adminService struct {
	repo repositories.AdminRepository
}

// NewAdminService creates an implementation of AdminService
func NewAdminService(repo repositories.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

// ==========================================
// CITIES
// ==========================================

func (s *adminService) CreateCity(ctx context.Context, city *models.City) error {
	city.Name = strings.TrimSpace(city.Name)
	city.Terminal = strings.TrimSpace(city.Terminal)
	if city.Name == "" || city.Terminal == "" {
		return errors.New("city name and terminal are required")
	}
	return s.repo.CreateCity(ctx, city)
}

func (s *adminService) GetAllCities(ctx context.Context) ([]models.City, error) {
	return s.repo.GetAllCities(ctx)
}

func (s *adminService) GetCityByID(ctx context.Context, id int) (*models.City, error) {
	return s.repo.GetCityByID(ctx, id)
}

func (s *adminService) UpdateCity(ctx context.Context, city *models.City) error {
	city.Name = strings.TrimSpace(city.Name)
	city.Terminal = strings.TrimSpace(city.Terminal)
	if city.Name == "" || city.Terminal == "" {
		return errors.New("city name and terminal are required")
	}
	return s.repo.UpdateCity(ctx, city)
}

func (s *adminService) DeleteCity(ctx context.Context, id int) error {
	return s.repo.DeleteCity(ctx, id)
}

// ==========================================
// ROUTES
// ==========================================

func (s *adminService) CreateRoute(ctx context.Context, route *models.Route) error {
	if route.OriginCityID == route.DestinationCityID {
		return ErrIdenticalCities
	}
	if route.DistanceKm <= 0 {
		return ErrInvalidDistance
	}
	if route.EstimatedDurationMinutes <= 0 {
		return ErrInvalidDuration
	}

	// Verify both cities exist
	if _, err := s.repo.GetCityByID(ctx, route.OriginCityID); err != nil {
		return fmt.Errorf("origin city not found: %w", err)
	}
	if _, err := s.repo.GetCityByID(ctx, route.DestinationCityID); err != nil {
		return fmt.Errorf("destination city not found: %w", err)
	}

	// Check duplicate route
	existing, err := s.repo.GetRouteByCities(ctx, route.OriginCityID, route.DestinationCityID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrDuplicateRoute
	}

	return s.repo.CreateRoute(ctx, route)
}

func (s *adminService) GetAllRoutes(ctx context.Context) ([]models.Route, error) {
	return s.repo.GetAllRoutes(ctx)
}

func (s *adminService) GetRouteByID(ctx context.Context, id int) (*models.Route, error) {
	return s.repo.GetRouteByID(ctx, id)
}

func (s *adminService) UpdateRoute(ctx context.Context, route *models.Route) error {
	if route.OriginCityID == route.DestinationCityID {
		return ErrIdenticalCities
	}
	if route.DistanceKm <= 0 {
		return ErrInvalidDistance
	}
	if route.EstimatedDurationMinutes <= 0 {
		return ErrInvalidDuration
	}

	// Check if updating creates duplicate with another existing route
	existing, err := s.repo.GetRouteByCities(ctx, route.OriginCityID, route.DestinationCityID)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != route.ID {
		return ErrDuplicateRoute
	}

	return s.repo.UpdateRoute(ctx, route)
}

func (s *adminService) DeleteRoute(ctx context.Context, id int) error {
	return s.repo.DeleteRoute(ctx, id)
}

// ==========================================
// BUS TYPES
// ==========================================

func (s *adminService) CreateBusType(ctx context.Context, busType *models.BusType) error {
	busType.Name = strings.TrimSpace(busType.Name)
	if busType.Name == "" {
		return errors.New("bus type name is required")
	}
	if busType.SeatCapacity <= 0 {
		return ErrInvalidCapacity
	}
	return s.repo.CreateBusType(ctx, busType)
}

func (s *adminService) GetAllBusTypes(ctx context.Context) ([]models.BusType, error) {
	return s.repo.GetAllBusTypes(ctx)
}

func (s *adminService) GetBusTypeByID(ctx context.Context, id int) (*models.BusType, error) {
	return s.repo.GetBusTypeByID(ctx, id)
}

func (s *adminService) UpdateBusType(ctx context.Context, busType *models.BusType) error {
	busType.Name = strings.TrimSpace(busType.Name)
	if busType.Name == "" {
		return errors.New("bus type name is required")
	}
	if busType.SeatCapacity <= 0 {
		return ErrInvalidCapacity
	}
	return s.repo.UpdateBusType(ctx, busType)
}

func (s *adminService) DeleteBusType(ctx context.Context, id int) error {
	return s.repo.DeleteBusType(ctx, id)
}

// ==========================================
// BUSES & SEAT AUTOGENERATION
// ==========================================

func (s *adminService) CreateBus(ctx context.Context, bus *models.Bus) error {
	bus.Plate = strings.ToUpper(strings.TrimSpace(bus.Plate))
	bus.InternalNumber = strings.TrimSpace(bus.InternalNumber)

	if bus.Plate == "" || bus.InternalNumber == "" {
		return errors.New("plate and internal number are required")
	}

	// 1. Fetch BusType to know capacity
	busType, err := s.repo.GetBusTypeByID(ctx, bus.BusTypeID)
	if err != nil {
		return fmt.Errorf("invalid bus type: %w", err)
	}

	if busType.SeatCapacity <= 0 {
		return ErrInvalidCapacity
	}

	// 2. Auto-generate seats in standard 4-column bus grid layout
	// Columns: 1 (Window Left), 2 (Aisle Left), 3 (Aisle Right), 4 (Window Right)
	seats := make([]models.Seat, busType.SeatCapacity)
	for i := 0; i < busType.SeatCapacity; i++ {
		seatNum := i + 1
		row := (i / 4) + 1
		col := (i % 4) + 1

		seats[i] = models.Seat{
			SeatNumber: seatNum,
			Row:        row,
			Column:     col,
		}
	}

	// 3. Atomically persist bus and all seats
	return s.repo.CreateBusWithSeats(ctx, bus, seats)
}

func (s *adminService) GetAllBuses(ctx context.Context) ([]models.Bus, error) {
	return s.repo.GetAllBuses(ctx)
}

func (s *adminService) GetBusByID(ctx context.Context, id int) (*models.Bus, error) {
	return s.repo.GetBusByID(ctx, id)
}

func (s *adminService) UpdateBus(ctx context.Context, bus *models.Bus) error {
	bus.Plate = strings.ToUpper(strings.TrimSpace(bus.Plate))
	bus.InternalNumber = strings.TrimSpace(bus.InternalNumber)

	if bus.Plate == "" || bus.InternalNumber == "" {
		return errors.New("plate and internal number are required")
	}

	return s.repo.UpdateBus(ctx, bus)
}

func (s *adminService) DeleteBus(ctx context.Context, id int) error {
	return s.repo.DeleteBus(ctx, id)
}

// ==========================================
// TRIPS
// ==========================================

func (s *adminService) CreateTrip(ctx context.Context, trip *models.Trip) error {
	if trip.TicketPrice < 0 {
		return ErrInvalidPrice
	}
	if trip.DepartureTime.IsZero() {
		return errors.New("departure time is required")
	}

	// Verify route exists
	if _, err := s.repo.GetRouteByID(ctx, trip.RouteID); err != nil {
		return fmt.Errorf("route not found: %w", err)
	}

	// Verify bus exists
	if _, err := s.repo.GetBusByID(ctx, trip.BusID); err != nil {
		return fmt.Errorf("bus not found: %w", err)
	}

	return s.repo.CreateTrip(ctx, trip)
}

func (s *adminService) GetAllTrips(ctx context.Context) ([]models.Trip, error) {
	return s.repo.GetAllTrips(ctx)
}

func (s *adminService) GetTripByID(ctx context.Context, id int) (*models.Trip, error) {
	return s.repo.GetTripByID(ctx, id)
}

func (s *adminService) DeleteTrip(ctx context.Context, id int) error {
	return s.repo.DeleteTrip(ctx, id)
}
