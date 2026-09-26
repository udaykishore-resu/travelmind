package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL connection pool
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresConnection creates a new PostgreSQL connection pool
func NewPostgresConnection(dsn string) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	// Set connection pool parameters
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 2 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Ping checks the database connection
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Query executes a query and returns rows
func (db *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a command
func (db *PostgresDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

// Begin starts a new transaction
func (db *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

// BeginTx starts a new transaction with options
func (db *PostgresDB) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return db.pool.BeginTx(ctx, txOptions)
}

// Close closes the connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// GetPool returns the underlying connection pool
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Health returns database health information
type HealthInfo struct {
	Connected            bool
	PoolSize             int32
	AvailableConnections int32
	WaitCount            int64
	IdleConnections      int32
	TotalConnections     int32
}

// GetHealth returns current database health status
func (db *PostgresDB) GetHealth(ctx context.Context) (*HealthInfo, error) {
	if err := db.Ping(ctx); err != nil {
		return &HealthInfo{Connected: false}, err
	}

	stat := db.pool.Stat()
	return &HealthInfo{
		Connected:            true,
		PoolSize:             stat.MaxConns(),
		AvailableConnections: stat.MaxConns() - stat.AcquiredConns(),
		WaitCount:            stat.EmptyAcquireCount(),
		IdleConnections:      stat.IdleConns(),
		TotalConnections:     stat.TotalConns(),
	}, nil
}

// RunMigrations runs database migrations (if using sql-migrate)
func (db *PostgresDB) RunMigrations(migrationPath string) error {
	// Implementation would depend on migration tool
	// This is a placeholder for integration with sql-migrate or similar
	return nil
}
