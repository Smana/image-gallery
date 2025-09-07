package database

import (
	"testing"
)

func TestNewConnection(t *testing.T) {
	tests := []struct {
		name      string
		dbURL     string
		wantError bool
	}{
		{
			name:      "empty database URL",
			dbURL:     "",
			wantError: true,
		},
		{
			name:      "invalid database URL",
			dbURL:     "invalid-url",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewConnection(tt.dbURL)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewConnection() expected error, got nil")
				}
				if db != nil {
					db.Close()
				}
				return
			}

			if err != nil {
				t.Errorf("NewConnection() error = %v", err)
				return
			}

			if db == nil {
				t.Errorf("NewConnection() returned nil database")
				return
			}

			db.Close()
		})
	}
}
