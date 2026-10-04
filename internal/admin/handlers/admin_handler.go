package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/repositories"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/services"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/models"
)

// AdminHandler handles HTTP requests for admin management
type AdminHandler struct {
	service services.AdminService
}

// NewAdminHandler instantiates AdminHandler with its service dependency
func NewAdminHandler(service services.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// RegisterRoutes registers all administration endpoints on the provided router group
func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	{
		// Cities
		admin.POST("/cities", h.CreateCity)
		admin.GET("/cities", h.GetAllCities)
		admin.GET("/cities/:id", h.GetCityByID)
		admin.PUT("/cities/:id", h.UpdateCity)
		admin.DELETE("/cities/:id", h.DeleteCity)

		// Routes
		admin.POST("/routes", h.CreateRoute)
		admin.GET("/routes", h.GetAllRoutes)
		admin.GET("/routes/:id", h.GetRouteByID)
		admin.PUT("/routes/:id", h.UpdateRoute)
		admin.DELETE("/routes/:id", h.DeleteRoute)

		// Bus Types
		admin.POST("/bus-types", h.CreateBusType)
		admin.GET("/bus-types", h.GetAllBusTypes)
		admin.GET("/bus-types/:id", h.GetBusTypeByID)
		admin.PUT("/bus-types/:id", h.UpdateBusType)
		admin.DELETE("/bus-types/:id", h.DeleteBusType)

		// Buses
		admin.POST("/buses", h.CreateBus)
		admin.GET("/buses", h.GetAllBuses)
		admin.GET("/buses/:id", h.GetBusByID)
		admin.PUT("/buses/:id", h.UpdateBus)
		admin.DELETE("/buses/:id", h.DeleteBus)

		// Trips
		admin.POST("/trips", h.CreateTrip)
		admin.GET("/trips", h.GetAllTrips)
		admin.GET("/trips/:id", h.GetTripByID)
		admin.DELETE("/trips/:id", h.DeleteTrip)
	}
}

// ==========================================
// CITIES HANDLERS
// ==========================================

func (h *AdminHandler) CreateCity(c *gin.Context) {
	var city models.City
	if err := c.ShouldBindJSON(&city); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateCity(c.Request.Context(), &city); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": city})
}

func (h *AdminHandler) GetAllCities(c *gin.Context) {
	cities, err := h.service.GetAllCities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cities})
}

func (h *AdminHandler) GetCityByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	city, err := h.service.GetCityByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": city})
}

func (h *AdminHandler) UpdateCity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	var city models.City
	if err := c.ShouldBindJSON(&city); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	city.ID = id

	if err := h.service.UpdateCity(c.Request.Context(), &city); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": city})
}

func (h *AdminHandler) DeleteCity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid city id"})
		return
	}

	if err := h.service.DeleteCity(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "city deleted successfully"})
}

// ==========================================
// ROUTES HANDLERS
// ==========================================

func (h *AdminHandler) CreateRoute(c *gin.Context) {
	var route models.Route
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateRoute(c.Request.Context(), &route); err != nil {
		if errors.Is(err, services.ErrIdenticalCities) || errors.Is(err, services.ErrDuplicateRoute) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": route})
}

func (h *AdminHandler) GetAllRoutes(c *gin.Context) {
	routes, err := h.service.GetAllRoutes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": routes})
}

func (h *AdminHandler) GetRouteByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	route, err := h.service.GetRouteByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": route})
}

func (h *AdminHandler) UpdateRoute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	var route models.Route
	if err := c.ShouldBindJSON(&route); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	route.ID = id

	if err := h.service.UpdateRoute(c.Request.Context(), &route); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		if errors.Is(err, services.ErrIdenticalCities) || errors.Is(err, services.ErrDuplicateRoute) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": route})
}

func (h *AdminHandler) DeleteRoute(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid route id"})
		return
	}

	if err := h.service.DeleteRoute(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "route deleted successfully"})
}

// ==========================================
// BUS TYPES HANDLERS
// ==========================================

func (h *AdminHandler) CreateBusType(c *gin.Context) {
	var busType models.BusType
	if err := c.ShouldBindJSON(&busType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateBusType(c.Request.Context(), &busType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": busType})
}

func (h *AdminHandler) GetAllBusTypes(c *gin.Context) {
	busTypes, err := h.service.GetAllBusTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": busTypes})
}

func (h *AdminHandler) GetBusTypeByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus type id"})
		return
	}

	busType, err := h.service.GetBusTypeByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": busType})
}

func (h *AdminHandler) UpdateBusType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus type id"})
		return
	}

	var busType models.BusType
	if err := c.ShouldBindJSON(&busType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	busType.ID = id

	if err := h.service.UpdateBusType(c.Request.Context(), &busType); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": busType})
}

func (h *AdminHandler) DeleteBusType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus type id"})
		return
	}

	if err := h.service.DeleteBusType(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "bus type deleted successfully"})
}

// ==========================================
// BUSES HANDLERS
// ==========================================

func (h *AdminHandler) CreateBus(c *gin.Context) {
	var bus models.Bus
	if err := c.ShouldBindJSON(&bus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateBus(c.Request.Context(), &bus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": bus})
}

func (h *AdminHandler) GetAllBuses(c *gin.Context) {
	buses, err := h.service.GetAllBuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": buses})
}

func (h *AdminHandler) GetBusByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus id"})
		return
	}

	bus, err := h.service.GetBusByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": bus})
}

func (h *AdminHandler) UpdateBus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus id"})
		return
	}

	var bus models.Bus
	if err := c.ShouldBindJSON(&bus); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bus.ID = id

	if err := h.service.UpdateBus(c.Request.Context(), &bus); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": bus})
}

func (h *AdminHandler) DeleteBus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus id"})
		return
	}

	if err := h.service.DeleteBus(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bus not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "bus deleted successfully"})
}

// ==========================================
// TRIPS HANDLERS
// ==========================================

func (h *AdminHandler) CreateTrip(c *gin.Context) {
	var trip models.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateTrip(c.Request.Context(), &trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": trip})
}

func (h *AdminHandler) GetAllTrips(c *gin.Context) {
	trips, err := h.service.GetAllTrips(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": trips})
}

func (h *AdminHandler) GetTripByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	trip, err := h.service.GetTripByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": trip})
}

func (h *AdminHandler) DeleteTrip(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trip id"})
		return
	}

	if err := h.service.DeleteTrip(c.Request.Context(), id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "trip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "trip deleted successfully"})
}
