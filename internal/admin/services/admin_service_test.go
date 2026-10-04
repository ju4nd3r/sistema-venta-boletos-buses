package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/repositories"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/services"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
)

// MockAdminRepository implements repositories.AdminRepository for unit testing
type MockAdminRepository struct {
	CreateCityFunc         func(ctx context.Context, city *models.City) error
	GetAllCitiesFunc       func(ctx context.Context) ([]models.City, error)
	GetCityByIDFunc        func(ctx context.Context, id int) (*models.City, error)
	UpdateCityFunc         func(ctx context.Context, city *models.City) error
	DeleteCityFunc         func(ctx context.Context, id int) error

	CreateRouteFunc        func(ctx context.Context, route *models.Route) error
	GetAllRoutesFunc       func(ctx context.Context) ([]models.Route, error)
	GetRouteByIDFunc       func(ctx context.Context, id int) (*models.Route, error)
	GetRouteByCitiesFunc   func(ctx context.Context, originID, destID int) (*models.Route, error)
	UpdateRouteFunc        func(ctx context.Context, route *models.Route) error
	DeleteRouteFunc        func(ctx context.Context, id int) error

	CreateBusTypeFunc      func(ctx context.Context, busType *models.BusType) error
	GetAllBusTypesFunc     func(ctx context.Context) ([]models.BusType, error)
	GetBusTypeByIDFunc     func(ctx context.Context, id int) (*models.BusType, error)
	UpdateBusTypeFunc      func(ctx context.Context, busType *models.BusType) error
	DeleteBusTypeFunc      func(ctx context.Context, id int) error

	CreateBusWithSeatsFunc func(ctx context.Context, bus *models.Bus, seats []models.Seat) error
	GetAllBusesFunc        func(ctx context.Context) ([]models.Bus, error)
	GetBusByIDFunc         func(ctx context.Context, id int) (*models.Bus, error)
	UpdateBusFunc          func(ctx context.Context, bus *models.Bus) error
	DeleteBusFunc          func(ctx context.Context, id int) error

	CreateTripFunc         func(ctx context.Context, trip *models.Trip) error
	GetAllTripsFunc        func(ctx context.Context) ([]models.Trip, error)
	GetTripByIDFunc        func(ctx context.Context, id int) (*models.Trip, error)
	DeleteTripFunc         func(ctx context.Context, id int) error
}

func (m *MockAdminRepository) CreateCity(ctx context.Context, city *models.City) error {
	if m.CreateCityFunc != nil {
		return m.CreateCityFunc(ctx, city)
	}
	return nil
}
func (m *MockAdminRepository) GetAllCities(ctx context.Context) ([]models.City, error) {
	if m.GetAllCitiesFunc != nil {
		return m.GetAllCitiesFunc(ctx)
	}
	return nil, nil
}
func (m *MockAdminRepository) GetCityByID(ctx context.Context, id int) (*models.City, error) {
	if m.GetCityByIDFunc != nil {
		return m.GetCityByIDFunc(ctx, id)
	}
	return &models.City{ID: id, Name: "City", Terminal: "Terminal"}, nil
}
func (m *MockAdminRepository) UpdateCity(ctx context.Context, city *models.City) error {
	if m.UpdateCityFunc != nil {
		return m.UpdateCityFunc(ctx, city)
	}
	return nil
}
func (m *MockAdminRepository) DeleteCity(ctx context.Context, id int) error {
	if m.DeleteCityFunc != nil {
		return m.DeleteCityFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminRepository) CreateRoute(ctx context.Context, route *models.Route) error {
	if m.CreateRouteFunc != nil {
		return m.CreateRouteFunc(ctx, route)
	}
	route.ID = 1
	return nil
}
func (m *MockAdminRepository) GetAllRoutes(ctx context.Context) ([]models.Route, error) {
	if m.GetAllRoutesFunc != nil {
		return m.GetAllRoutesFunc(ctx)
	}
	return nil, nil
}
func (m *MockAdminRepository) GetRouteByID(ctx context.Context, id int) (*models.Route, error) {
	if m.GetRouteByIDFunc != nil {
		return m.GetRouteByIDFunc(ctx, id)
	}
	return &models.Route{ID: id}, nil
}
func (m *MockAdminRepository) GetRouteByCities(ctx context.Context, originID, destID int) (*models.Route, error) {
	if m.GetRouteByCitiesFunc != nil {
		return m.GetRouteByCitiesFunc(ctx, originID, destID)
	}
	return nil, nil
}
func (m *MockAdminRepository) UpdateRoute(ctx context.Context, route *models.Route) error {
	if m.UpdateRouteFunc != nil {
		return m.UpdateRouteFunc(ctx, route)
	}
	return nil
}
func (m *MockAdminRepository) DeleteRoute(ctx context.Context, id int) error {
	if m.DeleteRouteFunc != nil {
		return m.DeleteRouteFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminRepository) CreateBusType(ctx context.Context, busType *models.BusType) error {
	if m.CreateBusTypeFunc != nil {
		return m.CreateBusTypeFunc(ctx, busType)
	}
	busType.ID = 1
	return nil
}
func (m *MockAdminRepository) GetAllBusTypes(ctx context.Context) ([]models.BusType, error) {
	if m.GetAllBusTypesFunc != nil {
		return m.GetAllBusTypesFunc(ctx)
	}
	return nil, nil
}
func (m *MockAdminRepository) GetBusTypeByID(ctx context.Context, id int) (*models.BusType, error) {
	if m.GetBusTypeByIDFunc != nil {
		return m.GetBusTypeByIDFunc(ctx, id)
	}
	return &models.BusType{ID: id, Name: "Ejecutivo", SeatCapacity: 8}, nil
}
func (m *MockAdminRepository) UpdateBusType(ctx context.Context, busType *models.BusType) error {
	if m.UpdateBusTypeFunc != nil {
		return m.UpdateBusTypeFunc(ctx, busType)
	}
	return nil
}
func (m *MockAdminRepository) DeleteBusType(ctx context.Context, id int) error {
	if m.DeleteBusTypeFunc != nil {
		return m.DeleteBusTypeFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminRepository) CreateBusWithSeats(ctx context.Context, bus *models.Bus, seats []models.Seat) error {
	if m.CreateBusWithSeatsFunc != nil {
		return m.CreateBusWithSeatsFunc(ctx, bus, seats)
	}
	bus.ID = 1
	bus.Seats = seats
	return nil
}
func (m *MockAdminRepository) GetAllBuses(ctx context.Context) ([]models.Bus, error) {
	if m.GetAllBusesFunc != nil {
		return m.GetAllBusesFunc(ctx)
	}
	return nil, nil
}
func (m *MockAdminRepository) GetBusByID(ctx context.Context, id int) (*models.Bus, error) {
	if m.GetBusByIDFunc != nil {
		return m.GetBusByIDFunc(ctx, id)
	}
	return &models.Bus{ID: id}, nil
}
func (m *MockAdminRepository) UpdateBus(ctx context.Context, bus *models.Bus) error {
	if m.UpdateBusFunc != nil {
		return m.UpdateBusFunc(ctx, bus)
	}
	return nil
}
func (m *MockAdminRepository) DeleteBus(ctx context.Context, id int) error {
	if m.DeleteBusFunc != nil {
		return m.DeleteBusFunc(ctx, id)
	}
	return nil
}

func (m *MockAdminRepository) CreateTrip(ctx context.Context, trip *models.Trip) error {
	if m.CreateTripFunc != nil {
		return m.CreateTripFunc(ctx, trip)
	}
	trip.ID = 1
	return nil
}
func (m *MockAdminRepository) GetAllTrips(ctx context.Context) ([]models.Trip, error) {
	if m.GetAllTripsFunc != nil {
		return m.GetAllTripsFunc(ctx)
	}
	return nil, nil
}
func (m *MockAdminRepository) GetTripByID(ctx context.Context, id int) (*models.Trip, error) {
	if m.GetTripByIDFunc != nil {
		return m.GetTripByIDFunc(ctx, id)
	}
	return &models.Trip{ID: id}, nil
}
func (m *MockAdminRepository) DeleteTrip(ctx context.Context, id int) error {
	if m.DeleteTripFunc != nil {
		return m.DeleteTripFunc(ctx, id)
	}
	return nil
}

// ==========================================
// UNIT TESTS
// ==========================================

func TestCreateRoute_Success(t *testing.T) {
	mockRepo := &MockAdminRepository{
		GetCityByIDFunc: func(ctx context.Context, id int) (*models.City, error) {
			return &models.City{ID: id, Name: "City", Terminal: "Terminal"}, nil
		},
		GetRouteByCitiesFunc: func(ctx context.Context, originID, destID int) (*models.Route, error) {
			return nil, nil // No duplicate
		},
		CreateRouteFunc: func(ctx context.Context, route *models.Route) error {
			route.ID = 10
			return nil
		},
	}

	service := services.NewAdminService(mockRepo)
	route := &models.Route{
		OriginCityID:             1,
		DestinationCityID:        2,
		DistanceKm:               400.0,
		EstimatedDurationMinutes: 360,
	}

	err := service.CreateRoute(context.Background(), route)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if route.ID != 10 {
		t.Errorf("expected route ID 10, got: %d", route.ID)
	}
}

func TestCreateRoute_IdenticalCities(t *testing.T) {
	mockRepo := &MockAdminRepository{}
	service := services.NewAdminService(mockRepo)

	route := &models.Route{
		OriginCityID:             1,
		DestinationCityID:        1, // Identical
		DistanceKm:               100.0,
		EstimatedDurationMinutes: 120,
	}

	err := service.CreateRoute(context.Background(), route)
	if !errors.Is(err, services.ErrIdenticalCities) {
		t.Fatalf("expected ErrIdenticalCities, got: %v", err)
	}
}

func TestCreateRoute_DuplicateRoute(t *testing.T) {
	mockRepo := &MockAdminRepository{
		GetCityByIDFunc: func(ctx context.Context, id int) (*models.City, error) {
			return &models.City{ID: id, Name: "City", Terminal: "Terminal"}, nil
		},
		GetRouteByCitiesFunc: func(ctx context.Context, originID, destID int) (*models.Route, error) {
			// Simulates existing duplicate route
			return &models.Route{ID: 5, OriginCityID: originID, DestinationCityID: destID}, nil
		},
	}

	service := services.NewAdminService(mockRepo)
	route := &models.Route{
		OriginCityID:             1,
		DestinationCityID:        2,
		DistanceKm:               300.0,
		EstimatedDurationMinutes: 240,
	}

	err := service.CreateRoute(context.Background(), route)
	if !errors.Is(err, services.ErrDuplicateRoute) {
		t.Fatalf("expected ErrDuplicateRoute, got: %v", err)
	}
}

func TestCreateRoute_InvalidDistanceOrDuration(t *testing.T) {
	mockRepo := &MockAdminRepository{}
	service := services.NewAdminService(mockRepo)

	// Negative distance
	err1 := service.CreateRoute(context.Background(), &models.Route{
		OriginCityID:             1,
		DestinationCityID:        2,
		DistanceKm:               -10.0,
		EstimatedDurationMinutes: 100,
	})
	if !errors.Is(err1, services.ErrInvalidDistance) {
		t.Errorf("expected ErrInvalidDistance, got: %v", err1)
	}

	// Zero duration
	err2 := service.CreateRoute(context.Background(), &models.Route{
		OriginCityID:             1,
		DestinationCityID:        2,
		DistanceKm:               100.0,
		EstimatedDurationMinutes: 0,
	})
	if !errors.Is(err2, services.ErrInvalidDuration) {
		t.Errorf("expected ErrInvalidDuration, got: %v", err2)
	}
}

func TestCreateBus_AutogeneratesSeatsGridCorrectly(t *testing.T) {
	var capturedSeats []models.Seat

	mockRepo := &MockAdminRepository{
		GetBusTypeByIDFunc: func(ctx context.Context, id int) (*models.BusType, error) {
			return &models.BusType{
				ID:           id,
				Name:         "Ejecutivo",
				SeatCapacity: 8, // 8 seats -> exactly 2 rows of 4 columns
			}, nil
		},
		CreateBusWithSeatsFunc: func(ctx context.Context, bus *models.Bus, seats []models.Seat) error {
			bus.ID = 1
			capturedSeats = seats
			return nil
		},
	}

	service := services.NewAdminService(mockRepo)
	bus := &models.Bus{
		Plate:          "XYZ-123",
		BusTypeID:      1,
		InternalNumber: "001",
	}

	err := service.CreateBus(context.Background(), bus)
	if err != nil {
		t.Fatalf("expected no error creating bus, got: %v", err)
	}

	if len(capturedSeats) != 8 {
		t.Fatalf("expected 8 generated seats, got: %d", len(capturedSeats))
	}

	// Verify grid coordinates (row, column, seat number)
	expectedGrid := []struct {
		SeatNumber int
		Row        int
		Col        int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{3, 1, 3},
		{4, 1, 4},
		{5, 2, 1},
		{6, 2, 2},
		{7, 2, 3},
		{8, 2, 4},
	}

	for i, exp := range expectedGrid {
		actual := capturedSeats[i]
		if actual.SeatNumber != exp.SeatNumber || actual.Row != exp.Row || actual.Column != exp.Col {
			t.Errorf("seat %d mismatch: expected (num=%d, row=%d, col=%d), got (num=%d, row=%d, col=%d)",
				i, exp.SeatNumber, exp.Row, exp.Col, actual.SeatNumber, actual.Row, actual.Column)
		}
	}
}

func TestCreateBus_InvalidBusType(t *testing.T) {
	mockRepo := &MockAdminRepository{
		GetBusTypeByIDFunc: func(ctx context.Context, id int) (*models.BusType, error) {
			return nil, repositories.ErrNotFound
		},
	}

	service := services.NewAdminService(mockRepo)
	bus := &models.Bus{
		Plate:          "XYZ-123",
		BusTypeID:      999, // Non-existent
		InternalNumber: "001",
	}

	err := service.CreateBus(context.Background(), bus)
	if err == nil {
		t.Fatalf("expected error for non-existent bus type, got nil")
	}
}

func TestCreateCity_Validation(t *testing.T) {
	mockRepo := &MockAdminRepository{}
	service := services.NewAdminService(mockRepo)

	// Empty name
	err1 := service.CreateCity(context.Background(), &models.City{Name: "", Terminal: "Terminal 1"})
	if err1 == nil {
		t.Errorf("expected error for empty city name, got nil")
	}

	// Empty terminal
	err2 := service.CreateCity(context.Background(), &models.City{Name: "Cali", Terminal: "   "})
	if err2 == nil {
		t.Errorf("expected error for empty terminal, got nil")
	}
}

func TestCreateTrip_InvalidPrice(t *testing.T) {
	mockRepo := &MockAdminRepository{}
	service := services.NewAdminService(mockRepo)

	trip := &models.Trip{
		RouteID:         1,
		BusID:           1,
		DepartureTime:   time.Now().Add(24 * time.Hour),
		TicketPrice:     -500, // Negative price
	}

	err := service.CreateTrip(context.Background(), trip)
	if !errors.Is(err, services.ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got: %v", err)
	}
}
