package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// For testing
var (
	userHomeDir = os.UserHomeDir
	osMkdirAll  = os.MkdirAll
)

// DB wraps a sql.DB connection to the shared Amadla SQLite database.
type DB struct {
	conn *sql.DB
	path string
}

// defaultDBPath returns ~/.local/amadla/db.sqlite
func defaultDBPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "amadla", "db.sqlite"), nil
}

// Open opens the shared Amadla database at the default path.
func Open() (*DB, error) {
	path, err := defaultDBPath()
	if err != nil {
		return nil, err
	}
	return OpenAt(path)
}

// OpenAt opens a SQLite database at the given path, creating parent directories as needed.
func OpenAt(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := osMkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}

	return &DB{conn: conn, path: path}, nil
}

// OpenMemory opens an in-memory SQLite database (for testing).
func OpenMemory() (*DB, error) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	return &DB{conn: conn, path: ":memory:"}, nil
}

// Conn returns the underlying *sql.DB connection.
func (d *DB) Conn() *sql.DB {
	return d.conn
}

// Path returns the database file path.
func (d *DB) Path() string {
	return d.path
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}
