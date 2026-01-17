package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgxpool with tenant-aware context
type DB struct {
	Pool *pgxpool.Pool
}

// Config holds database configuration
type Config struct {
	URL          string
	MaxConns     int32
	MinConns     int32
	MaxConnLife  time.Duration
	MaxConnIdle  time.Duration
	HealthCheck  time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() Config {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"
	}
	return Config{
		URL:          url,
		MaxConns:     25,
		MinConns:     5,
		MaxConnLife:  time.Hour,
		MaxConnIdle:  30 * time.Minute,
		HealthCheck:  time.Minute,
	}
}

// New creates a new database connection pool
func New(ctx context.Context, cfg Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLife
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdle
	poolConfig.HealthCheckPeriod = cfg.HealthCheck

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// WithTenant returns a connection with tenant context set for RLS
func (db *DB) WithTenant(ctx context.Context, tenantID string) (pgx.Tx, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	// Set tenant context for RLS policies
	_, err = tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("set tenant: %w", err)
	}

	return tx, nil
}

// ExecWithTenant executes a query with tenant context
func (db *DB) ExecWithTenant(ctx context.Context, tenantID string, sql string, args ...interface{}) error {
	tx, err := db.WithTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return tx.Commit(ctx)
}

// QueryWithTenant executes a query with tenant context and returns rows
func (db *DB) QueryWithTenant(ctx context.Context, tenantID string, sql string, args ...interface{}) (pgx.Rows, func(), error) {
	tx, err := db.WithTenant(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		tx.Rollback(ctx)
		return nil, nil, fmt.Errorf("query: %w", err)
	}

	cleanup := func() {
		rows.Close()
		tx.Commit(ctx)
	}

	return rows, cleanup, nil
}

// Health checks database health
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
