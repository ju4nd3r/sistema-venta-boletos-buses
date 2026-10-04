package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
)

var (
	ErrSeatAlreadyOccupied = errors.New("seat is already booked or paid for this trip")
	ErrSeatNotFound        = errors.New("seat does not exist or does not belong to the trip's bus")
	ErrTripNotFound        = errors.New("trip not found")
)

// TicketRepository defines contracts for public queries and concurrent ticket booking
type TicketRepository interface {
	SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error)
	GetTripByID(ctx context.Context, tripID int) (*models.Trip, error)
	GetSeatMapByTripID(ctx context.Context, tripID int) ([]models.SeatMapStatus, error)
	BookSeatWithLock(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error)
	GetTicketsByTripID(ctx context.Context, tripID int) ([]models.Ticket, error)
}

type postgresTicketRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTicketRepository instantiates TicketRepository backed by pgxpool
func NewPostgresTicketRepository(pool *pgxpool.Pool) TicketRepository {
	return &postgresTicketRepository{pool: pool}
}

// SearchTrips searches for scheduled trips filtering by origin, destination and optional date (YYYY-MM-DD)
func (r *postgresTicketRepository) SearchTrips(ctx context.Context, originCityID, destCityID int, date string) ([]models.Trip, error) {
	var query strings.Builder
	query.WriteString(`
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
		WHERE r.ciudad_origen_id = $1 AND r.ciudad_destino_id = $2
	`)

	params := []interface{}{originCityID, destCityID}
	if strings.TrimSpace(date) != "" {
		params = append(params, date)
		query.WriteString(fmt.Sprintf(" AND DATE(v.fecha_hora_salida) = $%d", len(params)))
	}

	query.WriteString(" ORDER BY v.fecha_hora_salida ASC")

	rows, err := r.pool.Query(ctx, query.String(), params...)
	if err != nil {
		return nil, fmt.Errorf("failed executing trip search: %w", err)
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
			return nil, fmt.Errorf("failed scanning search trip row: %w", err)
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

// GetTripByID returns trip details by ID
func (r *postgresTicketRepository) GetTripByID(ctx context.Context, tripID int) (*models.Trip, error) {
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

	err := r.pool.QueryRow(ctx, query, tripID).Scan(
		&v.ID, &v.RouteID, &v.BusID, &v.DepartureTime, &v.TicketPrice, &v.CreatedAt,
		&rt.DistanceKm, &rt.EstimatedDurationMinutes,
		&co.ID, &co.Name, &co.Terminal,
		&cd.ID, &cd.Name, &cd.Terminal,
		&b.Plate, &b.InternalNumber,
		&tb.Name, &tb.SeatCapacity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("failed fetching trip %d: %w", tripID, err)
	}
	rt.OriginCity = &co
	rt.DestinationCity = &cd
	v.Route = &rt
	b.BusType = &tb
	v.Bus = &b
	return &v, nil
}

// GetSeatMapByTripID returns a complete layout map of all bus seats, marking occupancy
func (r *postgresTicketRepository) GetSeatMapByTripID(ctx context.Context, tripID int) ([]models.SeatMapStatus, error) {
	query := `
		SELECT 
			s.id, s.numero_silla, s.fila, s.columna,
			CASE WHEN b.id IS NOT NULL AND b.estado IN ('RESERVADO', 'PAGADO') THEN FALSE ELSE TRUE END AS disponible,
			b.estado,
			b.nombre_pasajero
		FROM viajes v
		JOIN sillas s ON s.bus_id = v.bus_id
		LEFT JOIN boletos b ON b.viaje_id = v.id AND b.silla_id = s.id AND b.estado IN ('RESERVADO', 'PAGADO')
		WHERE v.id = $1
		ORDER BY s.numero_silla ASC`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching seat map for trip %d: %w", tripID, err)
	}
	defer rows.Close()

	seatMap := make([]models.SeatMapStatus, 0)
	for rows.Next() {
		var sm models.SeatMapStatus
		if err := rows.Scan(&sm.SeatID, &sm.SeatNumber, &sm.Row, &sm.Column, &sm.Available, &sm.TicketStatus, &sm.PassengerName); err != nil {
			return nil, fmt.Errorf("failed scanning seat map row: %w", err)
		}
		seatMap = append(seatMap, sm)
	}

	return seatMap, rows.Err()
}

// BookSeatWithLock executes pessimistic locking using SELECT ... FOR UPDATE within an ACID transaction
// preventing concurrent race conditions when booking seats.
func (r *postgresTicketRepository) BookSeatWithLock(ctx context.Context, req *models.BookingRequest) (*models.Ticket, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed beginning booking transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Verify trip exists and retrieve bus ID
	var busID int
	queryTrip := `SELECT bus_id FROM viajes WHERE id = $1`
	err = tx.QueryRow(ctx, queryTrip, req.TripID).Scan(&busID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("failed verifying trip in booking transaction: %w", err)
	}

	// Step 2: Pessimistically lock the seat row for update
	// Any concurrent transaction attempting to purchase this seat will be blocked here until this tx finishes
	var lockedSeatID int
	queryLockSeat := `SELECT id FROM sillas WHERE id = $1 AND bus_id = $2 FOR UPDATE`
	err = tx.QueryRow(ctx, queryLockSeat, req.SeatID, busID).Scan(&lockedSeatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSeatNotFound
		}
		return nil, fmt.Errorf("failed acquiring lock on seat %d: %w", req.SeatID, err)
	}

	// Step 3: Check if ticket already exists and is active for this trip and seat
	var existingTicketID int
	var existingStatus string
	queryCheckOccupied := `
		SELECT id, estado 
		FROM boletos 
		WHERE viaje_id = $1 AND silla_id = $2 AND estado IN ('RESERVADO', 'PAGADO')
		FOR UPDATE`
	err = tx.QueryRow(ctx, queryCheckOccupied, req.TripID, req.SeatID).Scan(&existingTicketID, &existingStatus)
	if err == nil {
		// Seat is already occupied!
		return nil, ErrSeatAlreadyOccupied
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error verifying existing ticket: %w", err)
	}

	// Step 4: Insert the ticket record
	insertTicketQuery := `
		INSERT INTO boletos (viaje_id, silla_id, nombre_pasajero, documento, estado)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, creado_en`

	ticket := &models.Ticket{
		TripID:        req.TripID,
		SeatID:        req.SeatID,
		PassengerName: req.PassengerName,
		DocumentID:    req.DocumentID,
		Status:        req.Status,
	}

	err = tx.QueryRow(ctx, insertTicketQuery,
		ticket.TripID,
		ticket.SeatID,
		ticket.PassengerName,
		ticket.DocumentID,
		ticket.Status,
	).Scan(&ticket.ID, &ticket.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed inserting booked ticket: %w", err)
	}

	// Step 5: Commit transaction releasing the lock
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed committing ticket booking transaction: %w", err)
	}

	return ticket, nil
}

// GetTicketByID fetches ticket details including related trip and seat
func (r *postgresTicketRepository) GetTicketByID(ctx context.Context, ticketID int) (*models.Ticket, error) {
	query := `
		SELECT 
			b.id, b.viaje_id, b.silla_id, b.nombre_pasajero, b.documento, b.estado, b.creado_en,
			s.bus_id, s.numero_silla, s.fila, s.columna,
			v.fecha_hora_salida, v.precio_boleto
		FROM boletos b
		JOIN sillas s ON b.silla_id = s.id
		JOIN viajes v ON b.viaje_id = v.id
		WHERE b.id = $1`

	var t models.Ticket
	var s models.Seat
	var v models.Trip

	err := r.pool.QueryRow(ctx, query, ticketID).Scan(
		&t.ID, &t.TripID, &t.SeatID, &t.PassengerName, &t.DocumentID, &t.Status, &t.CreatedAt,
		&s.BusID, &s.SeatNumber, &s.Row, &s.Column,
		&v.DepartureTime, &v.TicketPrice,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("ticket not found")
		}
		return nil, fmt.Errorf("failed fetching ticket %d: %w", ticketID, err)
	}
	s.ID = t.SeatID
	v.ID = t.TripID
	t.Seat = &s
	t.Trip = &v

	return &t, nil
}

// GetTicketsByTripID retrieves all tickets associated with a given trip
func (r *postgresTicketRepository) GetTicketsByTripID(ctx context.Context, tripID int) ([]models.Ticket, error) {
	query := `
		SELECT 
			b.id, b.viaje_id, b.silla_id, b.nombre_pasajero, b.documento, b.estado, b.creado_en,
			s.numero_silla, s.fila, s.columna
		FROM boletos b
		JOIN sillas s ON b.silla_id = s.id
		WHERE b.viaje_id = $1
		ORDER BY s.numero_silla ASC`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed fetching tickets for trip %d: %w", tripID, err)
	}
	defer rows.Close()

	tickets := make([]models.Ticket, 0)
	for rows.Next() {
		var t models.Ticket
		var s models.Seat
		err := rows.Scan(
			&t.ID, &t.TripID, &t.SeatID, &t.PassengerName, &t.DocumentID, &t.Status, &t.CreatedAt,
			&s.SeatNumber, &s.Row, &s.Column,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning ticket row: %w", err)
		}
		s.ID = t.SeatID
		t.Seat = &s
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}
