package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenMemory(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	assert.NotNil(t, d.Conn())
	assert.Equal(t, ":memory:", d.Path())
}

func TestOpenAt_CreatesDirectoryAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "sub", "dir", "test.sqlite")

	d, err := OpenAt(dbPath)
	require.NoError(t, err)
	defer d.Close()

	assert.Equal(t, dbPath, d.Path())

	// Verify file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestOpenAt_WALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "wal-test.sqlite")

	d, err := OpenAt(dbPath)
	require.NoError(t, err)
	defer d.Close()

	var journalMode string
	err = d.Conn().QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode)
}

func TestClose(t *testing.T) {
	d, err := OpenMemory()
	require.NoError(t, err)

	err = d.Close()
	assert.NoError(t, err)

	// After close, operations should fail
	err = d.Conn().Ping()
	assert.Error(t, err)
}

func TestDefaultDBPath(t *testing.T) {
	path, err := defaultDBPath()
	require.NoError(t, err)
	assert.Contains(t, path, filepath.Join(".local", "amadla", "db.sqlite"))
}

func TestOpen_UsesDefaultPath(t *testing.T) {
	// Just verify Open() works without error (creates real file in home dir)
	d, err := Open()
	require.NoError(t, err)
	defer d.Close()

	expected, _ := defaultDBPath()
	assert.Equal(t, expected, d.Path())
}
