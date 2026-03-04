package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRmTestSimple(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	trashDir := filepath.Join(tmpDir, "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	// Pre-create and migrate
	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	d.Close()

	origOpenDB := rmOpenDB
	origTrashDir := rmTrashDir
	t.Cleanup(func() {
		rmOpenDB = origOpenDB
		rmTrashDir = origTrashDir
		flagDryRun = false
		flagJSON = false
		flagQuiet = false
		flagVerbose = false
	})

	rmOpenDB = func() (*db.DB, error) { return db.OpenAt(dbPath) }
	rmTrashDir = func() (string, error) { return trashDir, nil }

	return dbPath, trashDir
}

func TestRmCmd_TrashFile(t *testing.T) {
	dbPath, trashDir := setupRmTestSimple(t)

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("hello"), 0644))

	var buf bytes.Buffer
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := runRm(cmd, []string{tmpFile})
	require.NoError(t, err)

	// File should be gone from original location
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))

	// Reopen DB to verify
	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	defer d.Close()
	svc := trash.NewService(d.Conn(), trashDir)
	items, err := svc.List(trash.ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestRmCmd_DryRun(t *testing.T) {
	_, _ = setupRmTestSimple(t)
	flagDryRun = true

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("hello"), 0644))

	cmd := &cobra.Command{Use: "test"}
	err := runRm(cmd, []string{tmpFile})
	require.NoError(t, err)

	// File should still exist (dry run)
	_, err = os.Stat(tmpFile)
	assert.NoError(t, err)
}

func TestRmCmd_NonexistentFile(t *testing.T) {
	_, _ = setupRmTestSimple(t)

	cmd := &cobra.Command{Use: "test"}
	// Should not return error (logs to stderr per-file), but shouldn't panic
	err := runRm(cmd, []string{"/nonexistent/file.txt"})
	assert.NoError(t, err)
}
