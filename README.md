# 🚌 Sistema de Venta de Boletos de Buses con Módulo Administrativo

Plataforma integral y robusta para la gestión de flotas, rutas, viajes y venta de boletos de transporte intermunicipal, diseñada siguiendo principios de **Clean Architecture** (Arquitectura Limpia), alta concurrencia y despliegue contenerizado con Docker.

---

## 🛠️ Stack Tecnológico

- **Backend:** Go (Golang) 1.22+ con el framework web [Gin Gonic](https://gin-gonic.com/) y el driver nativo de alto rendimiento [pgx/v5](https://github.com/jackc/pgx) para PostgreSQL.
- **Frontend:** SPA limpia, reactiva e interactiva con JavaScript nativo moderno y CSS Grid para la representación en tiempo real del bus y selección de asientos.
- **Base de Datos:** PostgreSQL 15+ estructurada con restricciones de integridad referencial, índices y control estricto de concurrencia.
- **Infraestructura:** Docker & Docker Compose para orquestar la API backend y el clúster de base de datos.
- **Control de Versiones & CI:** Git y GitHub.

---

## 🏛️ Arquitectura del Sistema (Clean Architecture)

El backend está desacoplado en capas con responsabilidades delimitadas para asegurar mantenibilidad, testabilidad y extensibilidad:

```
├── cmd/
│   └── api/
│       └── main.go          # Punto de entrada de la aplicación, inyección de dependencias
├── internal/
│   ├── models/              # Entidades del dominio y estructuras de datos
│   ├── admin/               # Módulo administrativo (Ciudades, Rutas, Tipos de Bus, Buses)
│   │   ├── handlers/        # Capa HTTP (Controladores Gin)
│   │   ├── services/        # Lógica de negocio y casos de uso
│   │   └── repositories/    # Persistencia e interacción SQL con PostgreSQL
│   ├── tickets/             # Módulo de viajes y venta de boletos
│   │   ├── handlers/        # Controladores de búsqueda y reserva
│   │   ├── services/        # Lógica de reserva transaccional
│   │   └── repositories/    # Consultas con bloqueo pesimista (FOR UPDATE)
│   └── platform/
│       ├── database/        # Conexión pgx, migraciones y healthcheck
│       └── config/          # Carga de variables de entorno
├── web/                     # Frontend SPA (Panel Admin + Portal de Venta)
├── docker-compose.yml       # Orquestación de servicios locales
└── Dockerfile               # Multi-stage build optimizado para el binario Go
```

---

## 🗄️ Modelo Relacional de Datos

1. **`ciudades`**: `id`, `nombre`, `terminal`.
2. **`rutas`**: `id`, `ciudad_origen_id`, `ciudad_destino_id`, `distancia_km`, `duracion_estimada_minutos`.
3. **`tipos_bus`**: `id`, `nombre` (ej: "Ejecutivo", "VIP"), `capacidad_sillas`.
4. **`buses`**: `id`, `placa`, `tipo_bus_id`, `numero_interno`.
5. **`viajes`**: `id`, `ruta_id`, `bus_id`, `fecha_hora_salida`, `precio_boleto`.
6. **`sillas`**: `id`, `bus_id`, `numero_silla`, `fila`, `columna`.
7. **`boletos`**: `id`, `viaje_id`, `silla_id`, `nombre_pasajero`, `documento`, `estado` (*"RESERVADO"*, *"PAGADO"*), `creado_en`.

---

## ⚡ Concurrencia y Prevención de Race Conditions

Para garantizar que **dos pasajeros no puedan reservar o pagar la misma silla simultáneamente**:
- Se implementan transacciones ACID a nivel de PostgreSQL con aislamiento adecuado.
- Se aplica bloqueo pesimista explícito mediante **`SELECT ... FOR UPDATE`** sobre el registro del asiento y el estado de disponibilidad del viaje durante el proceso de compra.
- Cualquier intento concurrente sobre la misma silla en el mismo viaje queda en espera del bloqueo y es rechazado inmediatamente si la reserva previa se confirma con éxito.

---

## 🚀 Despliegue Rápido (Local con Docker)

### Requisitos previos
- Docker & Docker Compose
- Go 1.22+ (opcional para desarrollo local fuera del contenedor)

### Pasos
1. Clona el repositorio:
   ```bash
   git clone <URL_DEL_REPOSITORIO>
   cd sistema-venta-boletos-buses
   ```
2. Configura las variables de entorno:
   ```bash
   cp .env.example .env
   ```
3. Levanta los contenedores:
   ```bash
   docker compose up --build
   ```
4. Accede a la plataforma:
   - **Frontend Web:** `http://localhost:8080`
   - **API Healthcheck:** `http://localhost:8080/health`

---

## 🧪 Pruebas Unitarias

Ejecuta el conjunto de pruebas unitarias con cobertura:
```bash
go test -v -cover ./internal/...
```

---

## 📄 Licencia

Este proyecto está bajo la licencia MIT.
