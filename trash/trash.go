package trash

import (
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// For testing
var (
	osStat      = os.Stat
	osRename    = os.Rename
	osRemove    = os.Remove
	osRemoveAll = os.RemoveAll
	osMkdirAll  = os.MkdirAll
	osOpen      = os.Open
	osCreate    = os.Create
	uuidNew     = uuid.NewString
	timeNow     = time.Now
)

// TrashService implements Service using a SQLite database and filesystem trash directory.
type TrashService struct {
	conn     *sql.DB
	trashDir string
}

// Trash moves a file or directory to the trash and records it in the database.
func (s *TrashService) Trash(path string) (*Item, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := osStat(absPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", absPath, err)
	}

	id := uuidNew()
	itemDir := filepath.Join(s.trashDir, id)
	if err := osMkdirAll(itemDir, 0755); err != nil {
		return nil, fmt.Errorf("creating trash dir: %w", err)
	}

	name := filepath.Base(absPath)
	trashPath := filepath.Join(itemDir, name)

	var itemType ItemType
	var sizeBytes int64
	if info.IsDir() {
		itemType = ItemTypeDirectory
		sizeBytes, _ = dirSize(absPath)
	} else {
		itemType = ItemTypeFile
		sizeBytes = info.Size()
	}

	if err := moveItem(absPath, trashPath); err != nil {
		osRemoveAll(itemDir)
		return nil, fmt.Errorf("moving to trash: %w", err)
	}

	now := timeNow().UTC().Format(time.RFC3339)
	_, err = s.conn.Exec(
		"INSERT INTO trash_items (id, name, original_path, trash_path, item_type, size_bytes, trashed_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id, name, absPath, trashPath, string(itemType), sizeBytes, now,
	)
	if err != nil {
		// Try to move back on DB failure
		moveItem(trashPath, absPath)
		osRemoveAll(itemDir)
		return nil, fmt.Errorf("recording in database: %w", err)
	}

	return &Item{
		ID:           id,
		Name:         name,
		OriginalPath: absPath,
		TrashPath:    trashPath,
		Type:         itemType,
		SizeBytes:    sizeBytes,
		TrashedAt:    now,
	}, nil
}

// Restore moves a trashed item back to its original location (or toOverride if set) and removes the DB record.
func (s *TrashService) Restore(id string, toOverride string) (*Item, error) {
	items, err := s.Find(id)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no item found for %q", id)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("ambiguous: %d items match %q — use the full ID", len(items), id)
	}

	item := items[0]

	dest := item.OriginalPath
	if toOverride != "" {
		dest, err = filepath.Abs(toOverride)
		if err != nil {
			return nil, fmt.Errorf("resolving destination: %w", err)
		}
	}

	// Check if destination already exists
	if _, err := osStat(dest); err == nil {
		return nil, fmt.Errorf("destination already exists: %s", dest)
	}

	// Ensure parent directory exists
	if err := osMkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("creating parent dir: %w", err)
	}

	if err := moveItem(item.TrashPath, dest); err != nil {
		return nil, fmt.Errorf("restoring from trash: %w", err)
	}

	if _, err := s.conn.Exec("DELETE FROM trash_items WHERE id=?", item.ID); err != nil {
		// Try to move back to trash on DB failure
		moveItem(dest, item.TrashPath)
		return nil, fmt.Errorf("removing from database: %w", err)
	}

	// Clean up the UUID directory in trash
	osRemoveAll(filepath.Dir(item.TrashPath))

	item.OriginalPath = dest
	return &item, nil
}

// List returns trashed items matching the filter.
func (s *TrashService) List(filter ListFilter) ([]Item, error) {
	query := "SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items"
	var conditions []string
	var args []any

	if filter.Type != "" {
		conditions = append(conditions, "item_type=?")
		args = append(args, string(filter.Type))
	}
	if filter.OlderThan > 0 {
		cutoff := timeNow().UTC().Add(-filter.OlderThan).Format(time.RFC3339)
		conditions = append(conditions, "trashed_at < ?")
		args = append(args, cutoff)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY trashed_at DESC"

	rows, err := s.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

// Find looks up items by full ID, partial ID prefix, or name.
func (s *TrashService) Find(idOrName string) ([]Item, error) {
	// Try exact ID first
	row := s.conn.QueryRow(
		"SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items WHERE id=?",
		idOrName,
	)
	item, err := scanItem(row)
	if err == nil {
		return []Item{*item}, nil
	}

	// Try partial ID prefix
	rows, err := s.conn.Query(
		"SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items WHERE id LIKE ?",
		idOrName+"%",
	)
	if err != nil {
		return nil, err
	}
	items, err := scanItems(rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}

	// Try name match
	rows, err = s.conn.Query(
		"SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items WHERE name=? ORDER BY trashed_at DESC",
		idOrName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

// Get retrieves a single item by exact ID.
func (s *TrashService) Get(id string) (*Item, error) {
	row := s.conn.QueryRow(
		"SELECT id, name, original_path, trash_path, item_type, size_bytes, trashed_at, metadata FROM trash_items WHERE id=?",
		id,
	)
	return scanItem(row)
}

// Delete permanently removes a single trashed item from disk and database.
func (s *TrashService) Delete(id string) error {
	item, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	trashItemDir := filepath.Dir(item.TrashPath)
	if err := osRemoveAll(trashItemDir); err != nil {
		return fmt.Errorf("removing from disk: %w", err)
	}

	if _, err := s.conn.Exec("DELETE FROM trash_items WHERE id=?", id); err != nil {
		return fmt.Errorf("removing from database: %w", err)
	}
	return nil
}

// DeleteAll permanently removes all trashed items. Returns count and total bytes freed.
func (s *TrashService) DeleteAll() (int, int64, error) {
	items, err := s.List(ListFilter{})
	if err != nil {
		return 0, 0, err
	}

	var totalBytes int64
	for _, item := range items {
		totalBytes += item.SizeBytes
		trashItemDir := filepath.Dir(item.TrashPath)
		osRemoveAll(trashItemDir)
	}

	_, err = s.conn.Exec("DELETE FROM trash_items")
	if err != nil {
		return 0, 0, fmt.Errorf("clearing database: %w", err)
	}

	return len(items), totalBytes, nil
}

// DeleteOlderThan permanently removes items trashed more than d ago. Returns count and bytes freed.
func (s *TrashService) DeleteOlderThan(d time.Duration) (int, int64, error) {
	items, err := s.List(ListFilter{OlderThan: d})
	if err != nil {
		return 0, 0, err
	}

	var totalBytes int64
	for _, item := range items {
		totalBytes += item.SizeBytes
		trashItemDir := filepath.Dir(item.TrashPath)
		osRemoveAll(trashItemDir)
	}

	cutoff := timeNow().UTC().Add(-d).Format(time.RFC3339)
	_, err = s.conn.Exec("DELETE FROM trash_items WHERE trashed_at < ?", cutoff)
	if err != nil {
		return 0, 0, fmt.Errorf("clearing database: %w", err)
	}

	return len(items), totalBytes, nil
}

// Stats returns the count of trashed items and their total size in bytes.
func (s *TrashService) Stats() (int, int64, error) {
	var count int
	var totalBytes sql.NullInt64
	err := s.conn.QueryRow("SELECT COUNT(*), COALESCE(SUM(size_bytes), 0) FROM trash_items").Scan(&count, &totalBytes)
	if err != nil {
		return 0, 0, err
	}
	return count, totalBytes.Int64, nil
}

// --- Helpers ---

func scanItem(row *sql.Row) (*Item, error) {
	var item Item
	var metadata sql.NullString
	err := row.Scan(&item.ID, &item.Name, &item.OriginalPath, &item.TrashPath, &item.Type, &item.SizeBytes, &item.TrashedAt, &metadata)
	if err != nil {
		return nil, err
	}
	if metadata.Valid {
		item.Metadata = metadata.String
	}
	return &item, nil
}

func scanItems(rows *sql.Rows) ([]Item, error) {
	var items []Item
	for rows.Next() {
		var item Item
		var metadata sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.OriginalPath, &item.TrashPath, &item.Type, &item.SizeBytes, &item.TrashedAt, &metadata); err != nil {
			return nil, err
		}
		if metadata.Valid {
			item.Metadata = metadata.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// dirSize calculates the total size of all files in a directory tree.
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// moveItem tries os.Rename first, falling back to copy+remove for cross-filesystem moves.
func moveItem(src, dst string) error {
	err := osRename(src, dst)
	if err == nil {
		return nil
	}

	// Fallback: copy then remove
	info, err := osStat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := copyDir(src, dst); err != nil {
			return err
		}
	} else {
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}

	return osRemoveAll(src)
}

func copyFile(src, dst string) error {
	in, err := osOpen(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := osCreate(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func copyDir(src, dst string) error {
	srcInfo, err := osStat(src)
	if err != nil {
		return err
	}

	if err := osMkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
