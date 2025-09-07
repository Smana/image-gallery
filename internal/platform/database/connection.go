package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewConnection(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		return nil, ErrMissingDatabaseURL
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}