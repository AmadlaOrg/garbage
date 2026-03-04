package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEmptyTest(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	trashDir := filepath.Join(tmpDir, "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))

	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	d.Close()

	origOpenDB := emptyOpenDB
	origTrashDir := emptyTrashDir
	origOlderThan := emptyOlderThanFlag
	origForce := emptyForceFlag
	t.Cleanup(func() {
		emptyOpenDB = origOpenDB
		emptyTrashDir = origTrashDir
		emptyOlderThanFlag = origOlderThan
		emptyForceFlag = origForce
		flagDryRun = false
	})

	emptyOpenDB = func() (*db.DB, error) { return db.OpenAt(dbPath) }
	emptyTrashDir = func() (string, error) { return trashDir, nil }

	return dbPath, trashDir
}

func TestEmptyCmd_AllWithForce(t *testing.T) {
	dbPath, trashDir := setupEmptyTest(t)
	emptyForceFlag = true

	// Seed items using a separate connection
	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	svc := trash.NewService(d.Conn(), trashDir)
	f1 := filepath.Join(t.TempDir(), "a.txt")
	f2 := filepath.Join(t.TempDir(), "b.txt")
	require.NoError(t, os.WriteFile(f1, []byte("aaa"), 0644))
	require.NoError(t, os.WriteFile(f2, []byte("bbb"), 0644))
	_, err = svc.Trash(f1)
	require.NoError(t, err)
	_, err = svc.Trash(f2)
	require.NoError(t, err)
	d.Close()

	cmd := &cobra.Command{Use: "test"}
	err = runEmpty(cmd, nil)
	require.NoError(t, err)

	// Reopen to verify
	d2, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	defer d2.Close()
	svc2 := trash.NewService(d2.Conn(), trashDir)
	items, err := svc2.List(trash.ListFilter{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestEmptyCmd_WithoutForce(t *testing.T) {
	dbPath, trashDir := setupEmptyTest(t)
	emptyForceFlag = false

	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	svc := trash.NewService(d.Conn(), trashDir)
	f := filepath.Join(t.TempDir(), "c.txt")
	require.NoError(t, os.WriteFile(f, []byte("ccc"), 0644))
	_, err = svc.Trash(f)
	require.NoError(t, err)
	d.Close()

	cmd := &cobra.Command{Use: "test"}
	err = runEmpty(cmd, nil)
	require.NoError(t, err)

	// Reopen to verify items still exist
	d2, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	defer d2.Close()
	svc2 := trash.NewService(d2.Conn(), trashDir)
	items, err := svc2.List(trash.ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestEmptyCmd_DryRun(t *testing.T) {
	dbPath, trashDir := setupEmptyTest(t)
	flagDryRun = true

	d, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	svc := trash.NewService(d.Conn(), trashDir)
	f := filepath.Join(t.TempDir(), "d.txt")
	require.NoError(t, os.WriteFile(f, []byte("ddd"), 0644))
	_, err = svc.Trash(f)
	require.NoError(t, err)
	d.Close()

	cmd := &cobra.Command{Use: "test"}
	err = runEmpty(cmd, nil)
	require.NoError(t, err)

	// Reopen to verify items still exist
	d2, err := db.OpenAt(dbPath)
	require.NoError(t, err)
	defer d2.Close()
	svc2 := trash.NewService(d2.Conn(), trashDir)
	items, err := svc2.List(trash.ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestEmptyCmd_EmptyTrash(t *testing.T) {
	_, _ = setupEmptyTest(t)

	cmd := &cobra.Command{Use: "test"}
	err := runEmpty(cmd, nil)
	assert.NoError(t, err)
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		err   bool
	}{
		{"30d", 30 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"invalid", 0, true},
		{"xd", 0, true},
	}
	for _, tt := range tests {
		d, err := parseDuration(tt.input)
		if tt.err {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
			assert.Equal(t, tt.want, d, "input: %s", tt.input)
		}
	}
}
