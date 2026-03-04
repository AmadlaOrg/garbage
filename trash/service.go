package trash

import (
	"database/sql"
	"time"
)

// Service defines operations for managing trashed items.
type Service interface {
	Trash(path string) (*Item, error)
	Restore(id string, toOverride string) (*Item, error)
	List(filter ListFilter) ([]Item, error)
	Find(idOrName string) ([]Item, error)
	Get(id string) (*Item, error)
	Delete(id string) error
	DeleteAll() (int, int64, error)
	DeleteOlderThan(d time.Duration) (int, int64, error)
	Stats() (count int, totalBytes int64, err error)
}

// NewService creates a new trash service backed by the given database connection and trash directory.
func NewService(conn *sql.DB, trashDir string) Service {
	return &TrashService{
		conn:     conn,
		trashDir: trashDir,
	}
}
