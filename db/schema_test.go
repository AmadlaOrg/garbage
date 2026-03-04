package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate_CreatesTable(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	err = d.Migrate()
	require.NoError(t, err)

	// Verify table exists by querying it
	rows, err := d.Conn().Query("SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items")
	require.NoError(t, err)
	defer rows.Close()
}

func TestMigrate_CreatesIndexes(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	err = d.Migrate()
	require.NoError(t, err)

	// Check indexes exist
	indexes := []string{
		"idx_trash_items_name",
		"idx_trash_items_trashed_at",
		"idx_trash_items_item_type",
	}
	for _, idx := range indexes {
		var name string
		err := d.Conn().QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		assert.NoError(t, err, "index %s should exist", idx)
		assert.Equal(t, idx, name)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	require.NoError(t, d.Migrate())
	require.NoError(t, d.Migrate()) // second call should not error
}

func TestMigrate_InsertAndQuery(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	require.NoError(t, d.Migrate())

	_, err = d.Conn().Exec(
		"INSERT INTO trash_items (id, name, original_path, trash_path, item_type, size_bytes, trashed_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"abc-123", "test.txt", "/tmp/test.txt", "/trash/abc-123/test.txt", "file", 42, "2026-01-01T00:00:00Z",
	)
	require.NoError(t, err)

	var name string
	err = d.Conn().QueryRow("SELECT name FROM trash_items WHERE id=?", "abc-123").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "test.txt", name)
}
