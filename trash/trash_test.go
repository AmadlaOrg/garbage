package trash

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	gdb "github.com/AmadlaOrg/garbage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (Service, *gdb.DB, string) {
	t.Helper()
	d, err := gdb.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, d.Migrate())
	trashDir := filepath.Join(t.TempDir(), "trash")
	require.NoError(t, os.MkdirAll(trashDir, 0755))
	svc := NewService(d.Conn(), trashDir)
	return svc, d, trashDir
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile.txt")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func createTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	subDir := filepath.Join(dir, "myproject")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "a.txt"), []byte("aaa"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("bbb"), 0644))
	return subDir
}

func TestTrash_File(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "hello world")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	assert.NotEmpty(t, item.ID)
	assert.Equal(t, "testfile.txt", item.Name)
	assert.Equal(t, ItemTypeFile, item.Type)
	assert.Equal(t, int64(11), item.SizeBytes)
	assert.NotEmpty(t, item.TrashedAt)

	// Original file should be gone
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))

	// Trash file should exist
	_, err = os.Stat(item.TrashPath)
	assert.NoError(t, err)
}

func TestTrash_Directory(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempDir(t)
	item, err := svc.Trash(path)
	require.NoError(t, err)

	assert.Equal(t, "myproject", item.Name)
	assert.Equal(t, ItemTypeDirectory, item.Type)
	assert.Equal(t, int64(6), item.SizeBytes) // 3 + 3

	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestTrash_NonexistentPath(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	_, err := svc.Trash("/nonexistent/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stat")
}

func TestRestore_ByID(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "restore me")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	restored, err := svc.Restore(item.ID, "")
	require.NoError(t, err)
	assert.Equal(t, path, restored.OriginalPath)

	// File should be back
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "restore me", string(data))
}

func TestRestore_ToOverride(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "override dest")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	newDest := filepath.Join(t.TempDir(), "restored.txt")
	restored, err := svc.Restore(item.ID, newDest)
	require.NoError(t, err)
	assert.Equal(t, newDest, restored.OriginalPath)

	data, err := os.ReadFile(newDest)
	require.NoError(t, err)
	assert.Equal(t, "override dest", string(data))
}

func TestRestore_DestinationExists(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "data")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	// Create something at the original path
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte("blocker"), 0644))

	_, err = svc.Restore(item.ID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination already exists")
}

func TestRestore_NotFound(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	_, err := svc.Restore("nonexistent-id", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no item found")
}

func TestList_Empty(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	items, err := svc.List(ListFilter{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestList_All(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path1 := createTempFile(t, "a")
	path2 := createTempFile(t, "bb")
	_, err := svc.Trash(path1)
	require.NoError(t, err)
	_, err = svc.Trash(path2)
	require.NoError(t, err)

	items, err := svc.List(ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestList_FilterByType(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	filePath := createTempFile(t, "a")
	dirPath := createTempDir(t)
	_, err := svc.Trash(filePath)
	require.NoError(t, err)
	_, err = svc.Trash(dirPath)
	require.NoError(t, err)

	files, err := svc.List(ListFilter{Type: ItemTypeFile})
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, ItemTypeFile, files[0].Type)

	dirs, err := svc.List(ListFilter{Type: ItemTypeDirectory})
	require.NoError(t, err)
	assert.Len(t, dirs, 1)
	assert.Equal(t, ItemTypeDirectory, dirs[0].Type)
}

func TestFind_ByPartialID(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "find me")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	// Use first 8 chars of the UUID
	prefix := item.ID[:8]
	found, err := svc.Find(prefix)
	require.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, item.ID, found[0].ID)
}

func TestFind_ByName(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "named")
	_, err := svc.Trash(path)
	require.NoError(t, err)

	found, err := svc.Find("testfile.txt")
	require.NoError(t, err)
	assert.Len(t, found, 1)
	assert.Equal(t, "testfile.txt", found[0].Name)
}

func TestGet(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "get me")
	trashed, err := svc.Trash(path)
	require.NoError(t, err)

	item, err := svc.Get(trashed.ID)
	require.NoError(t, err)
	assert.Equal(t, trashed.ID, item.ID)
	assert.Equal(t, "testfile.txt", item.Name)
}

func TestGet_NotFound(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	_, err := svc.Get("nonexistent")
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path := createTempFile(t, "delete me")
	item, err := svc.Trash(path)
	require.NoError(t, err)

	err = svc.Delete(item.ID)
	require.NoError(t, err)

	// Should be gone from DB
	_, err = svc.Get(item.ID)
	assert.Error(t, err)

	// Trash file should be gone
	_, err = os.Stat(item.TrashPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteAll(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	path1 := createTempFile(t, "aaa")
	path2 := createTempFile(t, "bbbb")
	_, err := svc.Trash(path1)
	require.NoError(t, err)
	_, err = svc.Trash(path2)
	require.NoError(t, err)

	count, bytes, err := svc.DeleteAll()
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, int64(7), bytes)

	items, err := svc.List(ListFilter{})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestDeleteOlderThan(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	// Override timeNow to create an "old" item
	origTimeNow := timeNow
	defer func() { timeNow = origTimeNow }()

	timeNow = func() time.Time { return time.Now().Add(-48 * time.Hour) }
	path1 := createTempFile(t, "old")
	_, err := svc.Trash(path1)
	require.NoError(t, err)

	timeNow = origTimeNow
	path2 := createTempFile(t, "new")
	_, err = svc.Trash(path2)
	require.NoError(t, err)

	count, _, err := svc.DeleteOlderThan(24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	items, err := svc.List(ListFilter{})
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "testfile.txt", items[0].Name)
}

func TestStats(t *testing.T) {
	svc, d, _ := setupTest(t)
	defer d.Close()

	count, bytes, err := svc.Stats()
	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, int64(0), bytes)

	path := createTempFile(t, "12345")
	_, err = svc.Trash(path)
	require.NoError(t, err)

	count, bytes, err = svc.Stats()
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(5), bytes)
}

func TestDirSize(t *testing.T) {
	dir := createTempDir(t)
	size, err := dirSize(dir)
	require.NoError(t, err)
	assert.Equal(t, int64(6), size) // "aaa" + "bbb"
}
