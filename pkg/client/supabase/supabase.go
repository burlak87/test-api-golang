package supabase

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SupabaseClient struct {
	Pool *pgxpool.Pool
}

func NewSupabaseClient() (*SupabaseClient, error) {
	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		log.Fatal("SUPABASE_DB_URL environment variable is not set")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 300000000000

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	var version string
	if err := pool.QueryRow(context.Background(), "SELECT version()").Scan(&version); err != nil {
		pool.Close()
		return nil, err
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