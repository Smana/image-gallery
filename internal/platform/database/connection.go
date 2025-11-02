package database

import (
	"database/sql"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewConnection(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		return nil, ErrMissingDatabaseURL
	}

	// Open database connection with OpenTelemetry instrumentation
	// This automatically traces all SQL queries with proper semantic conventions
	db, err := otelsql.Open("postgres", dbURL,
		otelsql.WithAttributes(
			semconv.DBSystemPostgreSQL,
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	// Register DB stats metrics for connection pool monitoring
	if err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	)); err != nil {
		_ = db.Close() //nolint:errcheck // Connection cleanup in error path
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close() //nolint:errcheck // Connection cleanup in error path
		return nil, err
	}

	// Configure connection pool limits to prevent resource exhaustion
	// These settings are critical for high-concurrency scenarios
	db.SetMaxOpenConns(25)          // Limit total open connections (default: unlimited)
	db.SetMaxIdleConns(10)          // Keep some connections ready (default: 2)
	db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections periodically
	db.SetConnMaxIdleTime(2 * time.Minute) // Close idle connections after 2min

	return db, nil
}
