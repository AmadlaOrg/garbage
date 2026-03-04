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
	restoreOpenDB   = func() (*db.DB, error) { return db.Open() }
	restoreTrashDir = defaultTrashDir
)

var restoreToFlag string

// RestoreCmd restores a trashed item to its original or specified location.
var RestoreCmd = &cobra.Command{
	Use:   "restore <id|name>",
	Short: "Restore a trashed item",
	Args:  cobra.ExactArgs(1),
	RunE:  runRestore,
}

func init() {
	RestoreCmd.Flags().StringVar(&restoreToFlag, "to", "", "Restore to a specific path instead of the original")
}

func runRestore(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()
	idOrName := args[0]

	d, err := restoreOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := restoreTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)

	if flagDryRun {
		items, err := svc.Find(idOrName)
		if err != nil {
			return fmt.Errorf("finding item: %w", err)
		}
		if len(items) == 0 {
			return fmt.Errorf("no item found for %q", idOrName)
		}
		for _, item := range items {
			dest := item.OriginalPath
			if restoreToFlag != "" {
				dest = restoreToFlag
			}
			out.Info("Would restore: %s → %s", item.Name, dest)
		}
		return nil
	}

	item, err := svc.Restore(idOrName, restoreToFlag)
	if err != nil {
		return fmt.Errorf("restoring: %w", err)
	}

	out.Result(item, func(w io.Writer) {
		fmt.Fprintf(w, "Restored: %s → %s\n", item.Name, item.OriginalPath)
	})

	return nil
}
