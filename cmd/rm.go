package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
)

// For testing
var (
	rmOpenDB   = func() (*db.DB, error) { return db.Open() }
	rmTrashDir = defaultTrashDir
)

func defaultTrashDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "amadla", "trash"), nil
}

// RmCmd removes files/directories by moving them to trash.
var RmCmd = &cobra.Command{
	Use:   "rm <path>...",
	Short: "Move files or directories to trash",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRm,
}

func runRm(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()

	d, err := rmOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := rmTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)

	for _, path := range args {
		if flagDryRun {
			out.Info("Would trash: %s", path)
			continue
		}

		item, err := svc.Trash(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error trashing %s: %v\n", path, err)
			continue
		}

		out.Result(item, func(w io.Writer) {
			fmt.Fprintf(w, "Trashed: %s → %s\n", item.OriginalPath, item.ID)
		})

		out.Verbose("  Type: %s, Size: %d bytes", item.Type, item.SizeBytes)
	}

	return nil
}
