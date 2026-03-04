package db

const createTrashItemsTable = `
CREATE TABLE IF NOT EXISTS trash_items (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    original_path TEXT NOT NULL,
    trash_path    TEXT NOT NULL,
    item_type     TEXT NOT NULL DEFAULT 'file',
    size_bytes    INTEGER NOT NULL DEFAULT 0,
    trashed_at    TEXT NOT NULL,
    metadata      TEXT
);`

const createIndexName = `CREATE INDEX IF NOT EXISTS idx_trash_items_name ON trash_items(name);`
const createIndexTrashedAt = `CREATE INDEX IF NOT EXISTS idx_trash_items_trashed_at ON trash_items(trashed_at);`
const createIndexItemType = `CREATE INDEX IF NOT EXISTS idx_trash_items_item_type ON trash_items(item_type);`

// Migrate creates the trash_items table and indexes if they don't exist.
func (d *DB) Migrate() error {
	stmts := []string{
		createTrashItemsTable,
		createIndexName,
		createIndexTrashedAt,
		createIndexItemType,
	}
	for _, stmt := range stmts {
		if _, err := d.conn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
