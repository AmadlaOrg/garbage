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
	infoOpenDB   = func() (*db.DB, error) { return db.Open() }
	infoTrashDir = defaultTrashDir
)

// InfoCmd shows details about a single trashed item.
var InfoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show details of a trashed item",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()
	id := args[0]

	d, err := infoOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := infoTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)

	// Support partial ID matching
	items, err := svc.Find(id)
	if err != nil {
		return fmt.Errorf("finding item: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("no item found for %q", id)
	}
	if len(items) > 1 {
		return fmt.Errorf("ambiguous: %d items match %q — use the full ID", len(items), id)
	}

	item := items[0]
	out.Result(item, func(w io.Writer) {
		fmt.Fprintf(w, "ID:            %s\n", item.ID)
		fmt.Fprintf(w, "Name:          %s\n", item.Name)
		fmt.Fprintf(w, "Type:          %s\n", item.Type)
		fmt.Fprintf(w, "Original Path: %s\n", item.OriginalPath)
		fmt.Fprintf(w, "Trash Path:    %s\n", item.TrashPath)
		fmt.Fprintf(w, "Size:          %s (%d bytes)\n", formatSize(item.SizeBytes), item.SizeBytes)
		fmt.Fprintf(w, "Trashed At:    %s\n", item.TrashedAt)
		if item.Metadata != "" {
			fmt.Fprintf(w, "Metadata:      %s\n", item.Metadata)
		}
	})

	return nil
}
