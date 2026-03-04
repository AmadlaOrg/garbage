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

func setupListTest(t *testing.T) (*db.DB, string) {
	t.Helper()
	d, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	trashDir := filepath.Join(t.TempDir(), "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	origOpenDB := listOpenDB
	origTrashDir := listTrashDir
	origTypeFlag := listTypeFlag
	t.Cleanup(func() {
		listOpenDB = origOpenDB
		listTrashDir = origTrashDir
		listTypeFlag = origTypeFlag
		flagJSON = false
		flagQuiet = false
	})

	listOpenDB = func() (*db.DB, error) { return d, nil }
	listTrashDir = func() (string, error) { return trashDir, nil }

	return d, trashDir
}

func TestListCmd_Empty(t *testing.T) {
	d, _ := setupListTest(t)
	defer d.Close()

	cmd := &cobra.Command{Use: "test"}
	err := runList(cmd, nil)
	assert.NoError(t, err)
}

func TestListCmd_WithItems(t *testing.T) {
	d, trashDir := setupListTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "hello.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("data"), 0644))
	_, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	cmd := &cobra.Command{Use: "test"}
	err = runList(cmd, nil)
	assert.NoError(t, err)
}

func TestListCmd_FilterByType(t *testing.T) {
	d, trashDir := setupListTest(t)
	defer d.Close()

	svc := trash.NewService(d.Conn(), trashDir)
	tmpFile := filepath.Join(t.TempDir(), "hello.txt")
	require.NoError(t, os.WriteFile(tmpFile, []byte("data"), 0644))
	_, err := svc.Trash(tmpFile)
	require.NoError(t, err)

	listTypeFlag = "directory"
	cmd := &cobra.Command{Use: "test"}
	err = runList(cmd, nil)
	assert.NoError(t, err)
}

func TestShortID(t *testing.T) {
	assert.Equal(t, "abcdefgh", shortID("abcdefgh-1234-5678-9abc-def012345678"))
	assert.Equal(t, "short", shortID("short"))
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, formatSize(tt.bytes))
	}
}
