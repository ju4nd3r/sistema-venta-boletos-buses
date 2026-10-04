package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewConnection creates a PostgreSQL connection pool using pgxpool
func NewConnection(databaseURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database health check ping failed: %w", err)
	}

	log.Println("Database connection pool established successfully")
	return pool, nil
}

// RunMigrations executes initial SQL migration scripts if necessary
func RunMigrations(pool *pgxpool.Pool, migrationPath string) error {
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("could not read migration file %s: %w", migrationPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed executing migration script: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}
