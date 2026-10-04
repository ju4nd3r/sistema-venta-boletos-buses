package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	adminHandlers "github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/handlers"
	adminRepos "github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/repositories"
	adminServices "github.com/ju4nd3r/sistema-venta-boletos-buses/internal/admin/services"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/platform/config"
	"github.com/ju4nd3r/sistema-venta-boletos-buses/internal/platform/database"
)

func main() {
	// 1. Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Critical error loading configuration: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	// 2. Connect to PostgreSQL with retry loop for container readiness
	var pool *pgxpool.Pool
	for attempts := 1; attempts <= 10; attempts++ {
		pool, err = database.NewConnection(cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for database (attempt %d/10): %v", attempts, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("Warning: Failed to connect to database at startup: %v. Running in degraded state for healthcheck.", err)
	} else {
		defer pool.Close()

		// Optional automatic migration verification
		migrationPath := "migrations/001_init.sql"
		if _, statErr := os.Stat(migrationPath); statErr == nil {
			if migErr := database.RunMigrations(pool, migrationPath); migErr != nil {
				log.Printf("Migration notice: %v", migErr)
			}
		}
	}

	// 3. Initialize HTTP Router
	router := gin.Default()

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 4. Healthcheck Endpoint
	router.GET("/health", func(c *gin.Context) {
		dbStatus := "down"
		if pool != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := pool.Ping(ctx); err == nil {
				dbStatus = "up"
			}
		}

		status := http.StatusOK
		if dbStatus != "up" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"database":  dbStatus,
			"service":   "bus-ticket-system-api",
		})
	})

	// 5. Initialize Modules
	if pool != nil {
		adminRepo := adminRepos.NewPostgresAdminRepository(pool)
		adminService := adminServices.NewAdminService(adminRepo)
		adminHandler := adminHandlers.NewAdminHandler(adminService)

		api := router.Group("/api")
		adminHandler.RegisterRoutes(api)
	}

	// 6. Serve Frontend static assets if available
	if _, err := os.Stat("web"); err == nil {
		router.Static("/static", "./web")
		router.StaticFile("/", "./web/index.html")
	}

	// 6. Start HTTP server
	addr := ":" + cfg.Port
	log.Printf("Server listening on port %s", cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
