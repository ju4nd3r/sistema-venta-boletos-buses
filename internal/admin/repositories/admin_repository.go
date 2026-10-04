package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrConflict      = errors.New("resource already exists")
	ErrForeignKeyViolation = errors.New("related resource does not exist or cannot be removed")
)

// AdminRepository defines persistence contracts for all administrative entities
type AdminRepository interface {
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
	GetRouteByCities(ctx context.Context, originID, destID int) (*models.Route, error)
	UpdateRoute(ctx context.Context, route *models.Route) error
	DeleteRoute(ctx context.Context, id int) error

	// BusType operations
	CreateBusType(ctx context.Context, busType *models.BusType) error
	GetAllBusTypes(ctx context.Context) ([]models.BusType, error)
	GetBusTypeByID(ctx context.Context, id int) (*models.BusType, error)
	UpdateBusType(ctx context.Context, busType *models.BusType) error
	DeleteBusType(ctx context.Context, id int) error

	// Bus operations
	CreateBusWithSeats(ctx context.Context, bus *models.Bus, seats []models.Seat) error
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

type postgresAdminRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAdminRepository creates a new instance of AdminRepository with pgx connection pool
func NewPostgresAdminRepository(pool *pgxpool.Pool) AdminRepository {
	return &postgresAdminRepository{pool: pool}
}

// ==========================================
// CITIES
// ==========================================

func (r *postgresAdminRepository) CreateCity(ctx context.Context, city *models.City) error {
	query := `
		INSERT INTO ciudades (nombre, terminal)
		VALUES ($1, $2)
		RETURNING id, creado_en`
	err := r.pool.QueryRow(ctx, query, city.Name, city.Terminal).Scan(&city.ID, &city.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed creating city: %w", err)
	}
	return nil
}

func (r *postgresAdminRepository) GetAllCities(ctx context.Context) ([]models.City, error) {
	query := `SELECT id, nombre, terminal, creado_en FROM ciudades ORDER BY nombre ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed fetching cities: %w", err)
	}
	defer rows.Close()

	cities := make([]models.City, 0)
	for rows.Next() {
		var c models.City
		if err := rows.Scan(&c.ID, &c.Name, &c.Terminal, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning city row: %w", err)
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

func (r *postgresAdminRepository) GetCityByID(ctx context.Context, id int) (*models.City, error) {
	query := `SELECT id, nombre, terminal, creado_en FROM ciudades WHERE id = $1`
	var c models.City
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Terminal, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed fetching city by id %d: %w", id, err)
	}
	return &c, nil
}

func (r *postgresAdminRepository) UpdateCity(ctx context.Context, city *models.City) error {
	query := `UPDATE ciudades SET nombre = $1, terminal = $2 WHERE id = $3 RETURNING creado_en`
	err := r.pool.QueryRow(ctx, query, city.Name, city.Terminal, city.ID).Scan(&city.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed updating city %d: %w", city.ID, err)
	}
	return nil
}

func (r *postgresAdminRepository) DeleteCity(ctx context.Context, id int) error {
	query := `DELETE FROM ciudades WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting city %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================
// ROUTES
// ==========================================

func (r *postgresAdminRepository) CreateRoute(ctx context.Context, route *models.Route) error {
	query := `
		INSERT INTO rutas (ciudad_origen_id, ciudad_destino_id, distancia_km, duracion_estimada_minutos)
		VALUES ($1, $2, $3, $4)
		RETURNING id, creado_en`
	err := r.pool.QueryRow(ctx, query,
		route.OriginCityID,
		route.DestinationCityID,
		route.DistanceKm,
		route.EstimatedDurationMinutes,
	).Scan(&route.ID, &route.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed creating route: %w", err)
	}
	return nil
}

func (r *postgresAdminRepository) GetAllRoutes(ctx context.Context) ([]models.Route, error) {
	query := `
		SELECT 
			r.id, r.ciudad_origen_id, r.ciudad_destino_id, r.distancia_km, r.duracion_estimada_minutos, r.creado_en,
			co.id, co.nombre, co.terminal,
			cd.id, cd.nombre, cd.terminal
		FROM rutas r
		JOIN ciudades co ON r.ciudad_origen_id = co.id
		JOIN ciudades cd ON r.ciudad_destino_id = cd.id
		ORDER BY r.id ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed fetching routes: %w", err)
	}
	defer rows.Close()

	routes := make([]models.Route, 0)
	for rows.Next() {
		var rt models.Route
		var o models.City
		var d models.City
		err := rows.Scan(
			&rt.ID, &rt.OriginCityID, &rt.DestinationCityID, &rt.DistanceKm, &rt.EstimatedDurationMinutes, &rt.CreatedAt,
			&o.ID, &o.Name, &o.Terminal,
			&d.ID, &d.Name, &d.Terminal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning route row: %w", err)
		}
		rt.OriginCity = &o
		rt.DestinationCity = &d
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

func (r *postgresAdminRepository) GetRouteByID(ctx context.Context, id int) (*models.Route, error) {
	query := `
		SELECT 
			r.id, r.ciudad_origen_id, r.ciudad_destino_id, r.distancia_km, r.duracion_estimada_minutos, r.creado_en,
			co.id, co.nombre, co.terminal,
			cd.id, cd.nombre, cd.terminal
		FROM rutas r
		JOIN ciudades co ON r.ciudad_origen_id = co.id
		JOIN ciudades cd ON r.ciudad_destino_id = cd.id
		WHERE r.id = $1`
	var rt models.Route
	var o models.City
	var d models.City
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.OriginCityID, &rt.DestinationCityID, &rt.DistanceKm, &rt.EstimatedDurationMinutes, &rt.CreatedAt,
		&o.ID, &o.Name, &o.Terminal,
		&d.ID, &d.Name, &d.Terminal,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed fetching route %d: %w", id, err)
	}
	rt.OriginCity = &o
	rt.DestinationCity = &d
	return &rt, nil
}

func (r *postgresAdminRepository) GetRouteByCities(ctx context.Context, originID, destID int) (*models.Route, error) {
	query := `
		SELECT id, ciudad_origen_id, ciudad_destino_id, distancia_km, duracion_estimada_minutos, creado_en
		FROM rutas
		WHERE ciudad_origen_id = $1 AND ciudad_destino_id = $2`
	var rt models.Route
	err := r.pool.QueryRow(ctx, query, originID, destID).Scan(
		&rt.ID, &rt.OriginCityID, &rt.DestinationCityID, &rt.DistanceKm, &rt.EstimatedDurationMinutes, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // not found, no duplicate
		}
		return nil, fmt.Errorf("failed checking duplicate route: %w", err)
	}
	return &rt, nil
}

func (r *postgresAdminRepository) UpdateRoute(ctx context.Context, route *models.Route) error {
	query := `
		UPDATE rutas
		SET ciudad_origen_id = $1, ciudad_destino_id = $2, distancia_km = $3, duracion_estimada_minutos = $4
		WHERE id = $5
		RETURNING creado_en`
	err := r.pool.QueryRow(ctx, query,
		route.OriginCityID,
		route.DestinationCityID,
		route.DistanceKm,
		route.EstimatedDurationMinutes,
		route.ID,
	).Scan(&route.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed updating route %d: %w", route.ID, err)
	}
	return nil
}

func (r *postgresAdminRepository) DeleteRoute(ctx context.Context, id int) error {
	query := `DELETE FROM rutas WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting route %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================
// BUS TYPES
// ==========================================

func (r *postgresAdminRepository) CreateBusType(ctx context.Context, busType *models.BusType) error {
	query := `
		INSERT INTO tipos_bus (nombre, capacidad_sillas)
		VALUES ($1, $2)
		RETURNING id, creado_en`
	err := r.pool.QueryRow(ctx, query, busType.Name, busType.SeatCapacity).Scan(&busType.ID, &busType.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed creating bus type: %w", err)
	}
	return nil
}

func (r *postgresAdminRepository) GetAllBusTypes(ctx context.Context) ([]models.BusType, error) {
	query := `SELECT id, nombre, capacidad_sillas, creado_en FROM tipos_bus ORDER BY id ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed fetching bus types: %w", err)
	}
	defer rows.Close()

	busTypes := make([]models.BusType, 0)
	for rows.Next() {
		var bt models.BusType
		if err := rows.Scan(&bt.ID, &bt.Name, &bt.SeatCapacity, &bt.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning bus type row: %w", err)
		}
		busTypes = append(busTypes, bt)
	}
	return busTypes, rows.Err()
}

func (r *postgresAdminRepository) GetBusTypeByID(ctx context.Context, id int) (*models.BusType, error) {
	query := `SELECT id, nombre, capacidad_sillas, creado_en FROM tipos_bus WHERE id = $1`
	var bt models.BusType
	err := r.pool.QueryRow(ctx, query, id).Scan(&bt.ID, &bt.Name, &bt.SeatCapacity, &bt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed fetching bus type %d: %w", id, err)
	}
	return &bt, nil
}

func (r *postgresAdminRepository) UpdateBusType(ctx context.Context, busType *models.BusType) error {
	query := `UPDATE tipos_bus SET nombre = $1, capacidad_sillas = $2 WHERE id = $3 RETURNING creado_en`
	err := r.pool.QueryRow(ctx, query, busType.Name, busType.SeatCapacity, busType.ID).Scan(&busType.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed updating bus type %d: %w", busType.ID, err)
	}
	return nil
}

func (r *postgresAdminRepository) DeleteBusType(ctx context.Context, id int) error {
	query := `DELETE FROM tipos_bus WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting bus type %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================
// BUSES & SEATS
// ==========================================

func (r *postgresAdminRepository) CreateBusWithSeats(ctx context.Context, bus *models.Bus, seats []models.Seat) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed initiating bus creation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert Bus
	insertBusQuery := `
		INSERT INTO buses (placa, tipo_bus_id, numero_interno)
		VALUES ($1, $2, $3)
		RETURNING id, creado_en`
	err = tx.QueryRow(ctx, insertBusQuery, bus.Plate, bus.BusTypeID, bus.InternalNumber).Scan(&bus.ID, &bus.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed inserting bus record: %w", err)
	}

	// 2. Insert auto-generated seats in batch
	insertSeatQuery := `
		INSERT INTO sillas (bus_id, numero_silla, fila, columna)
		VALUES ($1, $2, $3, $4)
		RETURNING id, creado_en`

	createdSeats := make([]models.Seat, len(seats))
	for i, s := range seats {
		var seatID int
		err := tx.QueryRow(ctx, insertSeatQuery, bus.ID, s.SeatNumber, s.Row, s.Column).Scan(&seatID, &s.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed inserting seat %d for bus %d: %w", s.SeatNumber, bus.ID, err)
		}
		s.ID = seatID
		s.BusID = bus.ID
		createdSeats[i] = s
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed committing bus and seat creation: %w", err)
	}

	bus.Seats = createdSeats
	return nil
}

func (r *postgresAdminRepository) GetAllBuses(ctx context.Context) ([]models.Bus, error) {
	query := `
		SELECT 
			b.id, b.placa, b.tipo_bus_id, b.numero_interno, b.creado_en,
			tb.id, tb.nombre, tb.capacidad_sillas, tb.creado_en
		FROM buses b
		JOIN tipos_bus tb ON b.tipo_bus_id = tb.id
		ORDER BY b.id ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed fetching buses: %w", err)
	}
	defer rows.Close()

	buses := make([]models.Bus, 0)
	for rows.Next() {
		var b models.Bus
		var tb models.BusType
		err := rows.Scan(
			&b.ID, &b.Plate, &b.BusTypeID, &b.InternalNumber, &b.CreatedAt,
			&tb.ID, &tb.Name, &tb.SeatCapacity, &tb.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning bus row: %w", err)
		}
		b.BusType = &tb
		buses = append(buses, b)
	}
	return buses, rows.Err()
}

func (r *postgresAdminRepository) GetBusByID(ctx context.Context, id int) (*models.Bus, error) {
	queryBus := `
		SELECT 
			b.id, b.placa, b.tipo_bus_id, b.numero_interno, b.creado_en,
			tb.id, tb.nombre, tb.capacidad_sillas, tb.creado_en
		FROM buses b
		JOIN tipos_bus tb ON b.tipo_bus_id = tb.id
		WHERE b.id = $1`
	var b models.Bus
	var tb models.BusType
	err := r.pool.QueryRow(ctx, queryBus, id).Scan(
		&b.ID, &b.Plate, &b.BusTypeID, &b.InternalNumber, &b.CreatedAt,
		&tb.ID, &tb.Name, &tb.SeatCapacity, &tb.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed fetching bus %d: %w", id, err)
	}
	b.BusType = &tb

	// Fetch seats
	querySeats := `
		SELECT id, bus_id, numero_silla, fila, columna, creado_en
		FROM sillas
		WHERE bus_id = $1
		ORDER BY numero_silla ASC`
	seatRows, err := r.pool.Query(ctx, querySeats, id)
	if err != nil {
		return nil, fmt.Errorf("failed fetching seats for bus %d: %w", id, err)
	}
	defer seatRows.Close()

	seats := make([]models.Seat, 0)
	for seatRows.Next() {
		var s models.Seat
		if err := seatRows.Scan(&s.ID, &s.BusID, &s.SeatNumber, &s.Row, &s.Column, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed scanning seat row: %w", err)
		}
		seats = append(seats, s)
	}
	b.Seats = seats

	return &b, seatRows.Err()
}

func (r *postgresAdminRepository) UpdateBus(ctx context.Context, bus *models.Bus) error {
	query := `
		UPDATE buses
		SET placa = $1, tipo_bus_id = $2, numero_interno = $3
		WHERE id = $4
		RETURNING creado_en`
	err := r.pool.QueryRow(ctx, query, bus.Plate, bus.BusTypeID, bus.InternalNumber, bus.ID).Scan(&bus.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed updating bus %d: %w", bus.ID, err)
	}
	return nil
}

func (r *postgresAdminRepository) DeleteBus(ctx context.Context, id int) error {
	query := `DELETE FROM buses WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting bus %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ==========================================
// TRIPS (ADMIN)
// ==========================================

func (r *postgresAdminRepository) CreateTrip(ctx context.Context, trip *models.Trip) error {
	query := `
		INSERT INTO viajes (ruta_id, bus_id, fecha_hora_salida, precio_boleto)
		VALUES ($1, $2, $3, $4)
		RETURNING id, creado_en`
	err := r.pool.QueryRow(ctx, query,
		trip.RouteID,
		trip.BusID,
		trip.DepartureTime,
		trip.TicketPrice,
	).Scan(&trip.ID, &trip.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed creating trip: %w", err)
	}
	return nil
}

func (r *postgresAdminRepository) GetAllTrips(ctx context.Context) ([]models.Trip, error) {
	query := `
		SELECT 
			v.id, v.ruta_id, v.bus_id, v.fecha_hora_salida, v.precio_boleto, v.creado_en,
			r.distancia_km, r.duracion_estimada_minutos,
			co.id, co.nombre, co.terminal,
			cd.id, cd.nombre, cd.terminal,
			b.placa, b.numero_interno,
			tb.nombre, tb.capacidad_sillas
		FROM viajes v
		JOIN rutas r ON v.ruta_id = r.id
		JOIN ciudades co ON r.ciudad_origen_id = co.id
		JOIN ciudades cd ON r.ciudad_destino_id = cd.id
		JOIN buses b ON v.bus_id = b.id
		JOIN tipos_bus tb ON b.tipo_bus_id = tb.id
		ORDER BY v.fecha_hora_salida ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed fetching trips: %w", err)
	}
	defer rows.Close()

	trips := make([]models.Trip, 0)
	for rows.Next() {
		var v models.Trip
		var rt models.Route
		var co, cd models.City
		var b models.Bus
		var tb models.BusType

		err := rows.Scan(
			&v.ID, &v.RouteID, &v.BusID, &v.DepartureTime, &v.TicketPrice, &v.CreatedAt,
			&rt.DistanceKm, &rt.EstimatedDurationMinutes,
			&co.ID, &co.Name, &co.Terminal,
			&cd.ID, &cd.Name, &cd.Terminal,
			&b.Plate, &b.InternalNumber,
			&tb.Name, &tb.SeatCapacity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning trip row: %w", err)
		}
		rt.OriginCity = &co
		rt.DestinationCity = &cd
		v.Route = &rt
		b.BusType = &tb
		v.Bus = &b
		trips = append(trips, v)
	}
	return trips, rows.Err()
}

func (r *postgresAdminRepository) GetTripByID(ctx context.Context, id int) (*models.Trip, error) {
	query := `
		SELECT 
			v.id, v.ruta_id, v.bus_id, v.fecha_hora_salida, v.precio_boleto, v.creado_en,
			r.distancia_km, r.duracion_estimada_minutos,
			co.id, co.nombre, co.terminal,
			cd.id, cd.nombre, cd.terminal,
			b.placa, b.numero_interno,
			tb.nombre, tb.capacidad_sillas
		FROM viajes v
		JOIN rutas r ON v.ruta_id = r.id
		JOIN ciudades co ON r.ciudad_origen_id = co.id
		JOIN ciudades cd ON r.ciudad_destino_id = cd.id
		JOIN buses b ON v.bus_id = b.id
		JOIN tipos_bus tb ON b.tipo_bus_id = tb.id
		WHERE v.id = $1`
	var v models.Trip
	var rt models.Route
	var co, cd models.City
	var b models.Bus
	var tb models.BusType

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.RouteID, &v.BusID, &v.DepartureTime, &v.TicketPrice, &v.CreatedAt,
		&rt.DistanceKm, &rt.EstimatedDurationMinutes,
		&co.ID, &co.Name, &co.Terminal,
		&cd.ID, &cd.Name, &cd.Terminal,
		&b.Plate, &b.InternalNumber,
		&tb.Name, &tb.SeatCapacity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed fetching trip %d: %w", id, err)
	}
	rt.OriginCity = &co
	rt.DestinationCity = &cd
	v.Route = &rt
	b.BusType = &tb
	v.Bus = &b
	return &v, nil
}

func (r *postgresAdminRepository) DeleteTrip(ctx context.Context, id int) error {
	query := `DELETE FROM viajes WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed deleting trip %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
