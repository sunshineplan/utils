package database

import "database/sql"

// Database is the interface that wraps the basic database operation method.
type Database interface {
	Open() (*sql.DB, error)
	Backup(file string) error
	Restore(file string) error
}
