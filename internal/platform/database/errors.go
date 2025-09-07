package database

import "errors"

var (
	ErrMissingDatabaseURL = errors.New("database URL is required")
	ErrMigrationFailed    = errors.New("migration failed")
)
