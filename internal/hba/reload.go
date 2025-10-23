package hba

import (
	"database/sql"
	"fmt"
)

// Reloader handles PostgreSQL configuration reload
type Reloader struct {
	db *sql.DB
}

// NewReloader creates a new reloader
func NewReloader(db *sql.DB) *Reloader {
	return &Reloader{db: db}
}

// Reload reloads PostgreSQL configuration using pg_reload_conf()
func (r *Reloader) Reload() error {
	_, err := r.db.Exec("SELECT pg_reload_conf()")
	if err != nil {
		return fmt.Errorf("failed to reload PostgreSQL configuration: %w", err)
	}
	return nil
}

// ReloadAndVerify reloads configuration and verifies it was successful
func (r *Reloader) ReloadAndVerify() error {
	var result bool
	err := r.db.QueryRow("SELECT pg_reload_conf()").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to reload PostgreSQL configuration: %w", err)
	}

	if !result {
		return fmt.Errorf("PostgreSQL configuration reload returned false")
	}

	return nil
}
