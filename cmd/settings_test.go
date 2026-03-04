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

func setupSettingsTest(t *testing.T) (*db.DB, string) {
	t.Helper()
	d, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	trashDir := filepath.Join(t.TempDir(), "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	origOpenDB := settingsOpenDB
	origTrashDir := settingsTrashDir
	t.Cleanup(func() {
		settingsOpenDB = origOpenDB
		settingsTrashDir = origTrashDir
	})

	settingsOpenDB = func() (*db.DB, error) { return d, nil }
	settingsTrashDir = func() (string, error) { return trashDir, nil }

	return d, trashDir
}

func TestSettingsCmd_Empty(t *testing.T) {
	d, _ := setupSettingsTest(t)
	defer d.Close()

	cmd := &cobra.Command{Use: "test"}
	err := runSettings(cmd, nil)
	assert.NoError(t, err)
}

func TestSettingsCmd_WithItems(t *testing.T) {
	d, trashDir := setupSettingsTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	f := filepath.Join(t.TempDir(), "settings-test.txt")
	require.NoError(t, os.WriteFile(f, []byte("12345"), 0644))
	_, err := svc.Trash(f)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runSettings(cmd, nil)
	assert.NoError(t, err)
}
