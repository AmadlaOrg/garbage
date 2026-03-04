package cmd

import (
	"fmt"
	"io"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
)

// For testing
var (
	settingsOpenDB   = func() (*db.DB, error) { return db.Open() }
	settingsTrashDir = defaultTrashDir
)

// SettingsCmd displays garbage configuration and stats.
var SettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show garbage configuration and statistics",
	Args:  cobra.NoArgs,
	RunE:  runSettings,
}

func runSettings(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()

	d, err := settingsOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := settingsTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)
	count, totalBytes, err := svc.Stats()
	if err != nil {
		return fmt.Errorf("getting stats: %w", err)
	}

	data := map[string]any{
		"trash_dir":   trashDir,
		"db_path":     d.Path(),
		"item_count":  count,
		"total_bytes": totalBytes,
		"total_size":  formatSize(totalBytes),
	}

	out.Result(data, func(w io.Writer) {
		fmt.Fprintf(w, "Trash Directory: %s\n", trashDir)
		fmt.Fprintf(w, "Database Path:   %s\n", d.Path())
		fmt.Fprintf(w, "Item Count:      %d\n", count)
		fmt.Fprintf(w, "Total Size:      %s\n", formatSize(totalBytes))
	})

	return nil
}
