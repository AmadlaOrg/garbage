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

func setupInfoTest(t *testing.T) (*db.DB, string) {
	t.Helper()
	d, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	trashDir := filepath.Join(t.TempDir(), "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	origOpenDB := infoOpenDB
	origTrashDir := infoTrashDir
	t.Cleanup(func() {
		infoOpenDB = origOpenDB
		infoTrashDir = origTrashDir
	})

	infoOpenDB = func() (*db.DB, error) { return d, nil }
	infoTrashDir = func() (string, error) { return trashDir, nil }

	return d, trashDir
}

func TestInfoCmd_Found(t *testing.T) {
	d, trashDir := setupInfoTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "info-test.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("info data"), 0644))
	item, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runInfo(cmd, []string{item.ID})
	assert.NoError(t, err)
}

func TestInfoCmd_PartialID(t *testing.T) {
	d, trashDir := setupInfoTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "partial.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("partial"), 0644))
	item, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runInfo(cmd, []string{item.ID[:8]})
	assert.NoError(t, err)
}

func TestInfoCmd_NotFound(t *testing.T) {
	d, _ := setupInfoTest(t)
	defer d.Close()

	cmd := &cobra.Command{Use: "test"}
	err := runInfo(cmd, []string{"nonexistent-id"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no item found")
}
