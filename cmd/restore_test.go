package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRestoreTest(t *testing.T) (*db.DB, string) {
	t.Helper()
	d, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	trashDir := filepath.Join(t.TempDir(), "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	origOpenDB := restoreOpenDB
	origTrashDir := restoreTrashDir
	origToFlag := restoreToFlag
	t.Cleanup(func() {
		restoreOpenDB = origOpenDB
		restoreTrashDir = origTrashDir
		restoreToFlag = origToFlag
		flagDryRun = false
	})

	restoreOpenDB = func() (*db.DB, error) { return d, nil }
	restoreTrashDir = func() (string, error) { return trashDir, nil }

	return d, trashDir
}

func TestRestoreCmd_Success(t *testing.T) {
	d, trashDir := setupRestoreTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "restore-me.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("data"), 0644))
	item, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runRestore(cmd, []string{item.ID})
	require.NoError(t, err)

	// File should be restored
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "data", string(data))
}

func TestRestoreCmd_WithToFlag(t *testing.T) {
	d, trashDir := setupRestoreTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "to-flag.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("override"), 0644))
	item, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	newDest := filepath.Join(t.TempDir(), "new-location.txt")
	restoreToFlag = newDest

	cmd := &cobra.Command{Use: "test"}
	err = runRestore(cmd, []string{item.ID})
	require.NoError(t, err)

	data, err := os.ReadFile(newDest)
	require.NoError(t, err)
	assert.Equal(t, "override", string(data))
}

func TestRestoreCmd_NotFound(t *testing.T) {
	d, _ := setupRestoreTest(t)
	defer d.Close()

	cmd := &cobra.Command{Use: "test"}
	err := runRestore(cmd, []string{"nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no item found")
}

func TestRestoreCmd_DryRun(t *testing.T) {
	d, trashDir := setupRestoreTest(t)
	defer d.Close()
	flagDryRun = true

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "dry.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("dry"), 0644))
	item, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runRestore(cmd, []string{item.ID})
	assert.NoError(t, err)

	// File should NOT be restored (dry run)
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))
}
