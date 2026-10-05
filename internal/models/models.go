package models

import "time"

// Ticket status constants
const (
	TicketStatusReserved  = "RESERVADO"
	TicketStatusPaid      = "PAGADO"
	TicketStatusCancelled = "CANCELADO"
)

// City represents a travel destination or origin terminal
type City struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"nombre" binding:"required" db:"nombre"`
	Terminal  string    `json:"terminal" binding:"required" db:"terminal"`
	CreatedAt time.Time `json:"creado_en,omitempty" db:"creado_en"`
}

// Route represents the route between two cities
type Route struct {
	ID                       int       `json:"id" db:"id"`
	OriginCityID             int       `json:"ciudad_origen_id" binding:"required" db:"ciudad_origen_id"`
	DestinationCityID        int       `json:"ciudad_destino_id" binding:"required" db:"ciudad_destino_id"`
	DistanceKm               float64   `json:"distancia_km" binding:"required,gt=0" db:"distancia_km"`
	EstimatedDurationMinutes int       `json:"duracion_estimada_minutos" binding:"required,gt=0" db:"duracion_estimada_minutos"`
	CreatedAt                time.Time `json:"creado_en,omitempty" db:"creado_en"`

	// Relational associations
	OriginCity      *City `json:"ciudad_origen,omitempty"`
	DestinationCity *City `json:"ciudad_destino,omitempty"`
}

// BusType defines the category and seat capacity (e.g. Executive, VIP)
type BusType struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"nombre" binding:"required" db:"nombre"`
	SeatCapacity int       `json:"capacidad_sillas" binding:"required,gt=0" db:"capacidad_sillas"`
	CreatedAt    time.Time `json:"creado_en,omitempty" db:"creado_en"`
}

// Bus represents a vehicle assigned to routes and trips
type Bus struct {
	ID             int       `json:"id" db:"id"`
	Plate          string    `json:"placa" binding:"required" db:"placa"`
	BusTypeID      int       `json:"tipo_bus_id" binding:"required" db:"tipo_bus_id"`
	InternalNumber string    `json:"numero_interno" binding:"required" db:"numero_interno"`
	CreatedAt      time.Time `json:"creado_en,omitempty" db:"creado_en"`

	BusType *BusType `json:"tipo_bus,omitempty"`
	Seats   []Seat   `json:"sillas,omitempty"`
}

// Seat represents an individual physical seat on a bus
type Seat struct {
	ID         int       `json:"id" db:"id"`
	BusID      int       `json:"bus_id" db:"bus_id"`
	SeatNumber int       `json:"numero_silla" db:"numero_silla"`
	Row        int       `json:"fila" db:"fila"`
	Column     int       `json:"columna" db:"columna"`
	CreatedAt  time.Time `json:"creado_en,omitempty" db:"creado_en"`
}

// Trip represents a scheduled route journey with assigned bus and ticket price
type Trip struct {
	ID            int       `json:"id" db:"id"`
	RouteID       int       `json:"ruta_id" binding:"required" db:"ruta_id"`
	BusID         int       `json:"bus_id" binding:"required" db:"bus_id"`
	DepartureTime time.Time `json:"fecha_hora_salida" binding:"required" db:"fecha_hora_salida"`
	TicketPrice   float64   `json:"precio_boleto" binding:"required,gte=0" db:"precio_boleto"`
	CreatedAt     time.Time `json:"creado_en,omitempty" db:"creado_en"`

	Route *Route `json:"ruta,omitempty"`
	Bus   *Bus   `json:"bus,omitempty"`
}

// Ticket represents a booked or purchased seat for a specific trip
type Ticket struct {
	ID            int       `json:"id" db:"id"`
	TripID        int       `json:"viaje_id" binding:"required" db:"viaje_id"`
	SeatID        int       `json:"silla_id" binding:"required" db:"silla_id"`
	PassengerName string    `json:"nombre_pasajero" binding:"required" db:"nombre_pasajero"`
	DocumentID    string    `json:"documento" binding:"required" db:"documento"`
	Status        string    `json:"estado" db:"estado"`
	CreatedAt     time.Time `json:"creado_en,omitempty" db:"creado_en"`

	Trip *Trip `json:"viaje,omitempty"`
	Seat *Seat `json:"silla,omitempty"`
}

// SeatMapStatus represents real-time seat availability for a trip
type SeatMapStatus struct {
	SeatID        int     `json:"silla_id"`
	SeatNumber    int     `json:"numero_silla"`
	Row           int     `json:"fila"`
	Column        int     `json:"columna"`
	Available     bool    `json:"disponible"`
	TicketStatus  *string `json:"estado_boleto,omitempty"`
	PassengerName *string `json:"nombre_pasajero,omitempty"`
}

// BookingRequest contains payload required to purchase or reserve a single seat
type BookingRequest struct {
	TripID        int    `json:"viaje_id" binding:"required"`
	SeatID        int    `json:"silla_id" binding:"required"`
	PassengerName string `json:"nombre_pasajero" binding:"required"`
	DocumentID    string `json:"documento" binding:"required"`
	Status        string `json:"estado"` // RESERVADO or PAGADO
}

// PassengerItem represents an individual passenger in a multi-seat reservation
type PassengerItem struct {
	SeatID        int    `json:"silla_id" binding:"required"`
	PassengerName string `json:"nombre_pasajero" binding:"required"`
	DocumentID    string `json:"documento" binding:"required"`
	PassengerType string `json:"tipo_pasajero,omitempty"` // ADULT or CHILD
}

// MultiBookingRequest represents a transaction requesting multiple seat reservations
type MultiBookingRequest struct {
	TripID     int             `json:"viaje_id" binding:"required"`
	Passengers []PassengerItem `json:"pasajeros" binding:"required,min=1"`
	Status     string          `json:"estado"` // RESERVADO or PAGADO
}

// MultiBookingResponse summarizes all tickets booked in a single atomic transaction
type MultiBookingResponse struct {
	TripID         int      `json:"viaje_id"`
	Tickets        []Ticket `json:"boletos"`
	TotalAmount    float64  `json:"total_pagado"`
	PassengerCount int      `json:"total_pasajeros"`
}

