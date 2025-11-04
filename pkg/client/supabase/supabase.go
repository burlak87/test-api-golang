package supabase

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SupabaseClient struct {
	Pool *pgxpool.Pool
}

func NewSupabaseClient(dbURL string) (*SupabaseClient, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("SUPABASE_CONNECT_DB environment variable is not set")
	}

	if !strings.HasPrefix(dbURL, "postgres://") {
		dbURL = "postgres://" + dbURL
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 300000000000

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	var version string
	if err := pool.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Connected to Supabase:", version)
	return &SupabaseClient{Pool: pool}, nil
}

func (c *SupabaseClient) Close() {
	if c.Pool != nil {
		c.Pool.Close()
		log.Println("Supabase connection pool closed")
	}
}