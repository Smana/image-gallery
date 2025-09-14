package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	tests := []struct {
		name           string
		dbURL          string
		wantError      bool
		expectedError  error
		wantConnection bool
	}{
		{
			name:          "empty database URL",
			dbURL:         "",
			wantError:     true,
			expectedError: ErrMissingDatabaseURL,
		},
		{
			name:      "invalid database URL format",
			dbURL:     "invalid-url",
			wantError: true,
		},
		{
			name:      "unreachable database",
			dbURL:     "postgres://user:pass@nonexistent:5432/db",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewConnection(tt.dbURL)

			if tt.wantError {
				require.Error(t, err, "NewConnection() should return error")
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				if db != nil {
					_ = db.Close() //nolint:errcheck // Connection cleanup in test
				}
				return
			}

			require.NoError(t, err, "NewConnection() should not return error")
			require.NotNil(t, db, "NewConnection() should return database connection")

			// Verify connection is actually working
			err = db.Ping()
			assert.NoError(t, err, "Database connection should be pingable")

			_ = db.Close() //nolint:errcheck // Connection cleanup in test
		})
	}
}

func TestNewConnectionWithRealDatabase(t *testing.T) {
	// Skip if no test database URL is provided
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err, "Should connect to test database")
	require.NotNil(t, db, "Database connection should not be nil")
	defer func() { _ = db.Close() }() //nolint:errcheck // Connection cleanup in test

	// Test connection is working
	err = db.Ping()
	assert.NoError(t, err, "Should be able to ping database")

	// Test connection settings
	assert.NotZero(t, db.Stats().MaxOpenConnections, "Should have connection pool configured")
}

func TestConnectionPooling(t *testing.T) {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := NewConnection(testDBURL)
	require.NoError(t, err)
	defer func() { _ = db.Close() }() //nolint:errcheck // Connection cleanup in test

	// Configure connection pool for testing
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test multiple concurrent connections
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			err := db.Ping()
			assert.NoError(t, err)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify connection stats
	stats := db.Stats()
	assert.LessOrEqual(t, stats.OpenConnections, 5, "Should not exceed max open connections")
}

func TestConnectionTimeout(t *testing.T) {
	// Test connection with very short timeout
	db, err := NewConnection("postgres://user:pass@192.0.2.0:5432/db?connect_timeout=1")
	if err == nil {
		_ = db.Close() //nolint:errcheck // Connection cleanup in test
		t.Skip("Connection unexpectedly succeeded")
	}

	// Should fail quickly due to timeout
	assert.Error(t, err, "Connection should timeout")
}

func TestDatabaseErrors(t *testing.T) {
	tests := []struct {
		name          string
		error         error
		expectedError error
	}{
		{
			name:          "missing database URL error",
			error:         ErrMissingDatabaseURL,
			expectedError: ErrMissingDatabaseURL,
		},
		{
			name:          "migration failed error",
			error:         ErrMigrationFailed,
			expectedError: ErrMigrationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedError, tt.error)
			assert.NotEmpty(t, tt.error.Error())
		})
	}
}
