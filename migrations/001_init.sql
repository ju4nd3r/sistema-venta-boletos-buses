-- Migración inicial: Creación de tablas para el Sistema de Venta de Boletos de Buses

CREATE TABLE IF NOT EXISTS ciudades (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    terminal VARCHAR(150) NOT NULL,
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rutas (
    id SERIAL PRIMARY KEY,
    ciudad_origen_id INT NOT NULL REFERENCES ciudades(id) ON DELETE RESTRICT,
    ciudad_destino_id INT NOT NULL REFERENCES ciudades(id) ON DELETE RESTRICT,
    distancia_km NUMERIC(8,2) NOT NULL CHECK (distancia_km > 0),
    duracion_estimada_minutos INT NOT NULL CHECK (duracion_estimada_minutos > 0),
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_ruta_origen_destino UNIQUE (ciudad_origen_id, ciudad_destino_id),
    CONSTRAINT chk_rutas_distintas CHECK (ciudad_origen_id <> ciudad_destino_id)
);

CREATE TABLE IF NOT EXISTS tipos_bus (
    id SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    capacidad_sillas INT NOT NULL CHECK (capacidad_sillas > 0),
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS buses (
    id SERIAL PRIMARY KEY,
    placa VARCHAR(20) NOT NULL UNIQUE,
    tipo_bus_id INT NOT NULL REFERENCES tipos_bus(id) ON DELETE RESTRICT,
    numero_interno VARCHAR(20) NOT NULL UNIQUE,
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sillas (
    id SERIAL PRIMARY KEY,
    bus_id INT NOT NULL REFERENCES buses(id) ON DELETE CASCADE,
    numero_silla INT NOT NULL,
    fila INT NOT NULL CHECK (fila > 0),
    columna INT NOT NULL CHECK (columna > 0),
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_bus_silla UNIQUE (bus_id, numero_silla)
);

CREATE TABLE IF NOT EXISTS viajes (
    id SERIAL PRIMARY KEY,
    ruta_id INT NOT NULL REFERENCES rutas(id) ON DELETE RESTRICT,
    bus_id INT NOT NULL REFERENCES buses(id) ON DELETE RESTRICT,
    fecha_hora_salida TIMESTAMP WITH TIME ZONE NOT NULL,
    precio_boleto NUMERIC(10,2) NOT NULL CHECK (precio_boleto >= 0),
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS boletos (
    id SERIAL PRIMARY KEY,
    viaje_id INT NOT NULL REFERENCES viajes(id) ON DELETE RESTRICT,
    silla_id INT NOT NULL REFERENCES sillas(id) ON DELETE RESTRICT,
    nombre_pasajero VARCHAR(150) NOT NULL,
    documento VARCHAR(50) NOT NULL,
    estado VARCHAR(20) NOT NULL CHECK (estado IN ('RESERVADO', 'PAGADO', 'CANCELADO')),
    creado_en TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_viaje_silla UNIQUE (viaje_id, silla_id)
);

-- Índices para optimizar consultas de ventas y concurrencia
CREATE INDEX IF NOT EXISTS idx_rutas_ciudades ON rutas(ciudad_origen_id, ciudad_destino_id);
CREATE INDEX IF NOT EXISTS idx_viajes_bus ON viajes(bus_id);
CREATE INDEX IF NOT EXISTS idx_viajes_fecha ON viajes(ruta_id, fecha_hora_salida);
CREATE INDEX IF NOT EXISTS idx_sillas_bus ON sillas(bus_id);
CREATE INDEX IF NOT EXISTS idx_boletos_viaje ON boletos(viaje_id);
CREATE INDEX IF NOT EXISTS idx_boletos_silla ON boletos(silla_id);
