package cmd

import (
	"fmt"
	"io"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// For testing
var (
	listOpenDB    = func() (*db.DB, error) { return db.Open() }
	listTrashDir  = defaultTrashDir
)

var listTypeFlag string

// ListCmd lists trashed items.
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List trashed items",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	ListCmd.Flags().StringVar(&listTypeFlag, "type", "", "Filter by type (file, directory)")
}

func runList(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()

	d, err := listOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := listTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)

	filter := trash.ListFilter{}
	if listTypeFlag != "" {
		filter.Type = trash.ItemType(listTypeFlag)
	}

	items, err := svc.List(filter)
	if err != nil {
		return fmt.Errorf("listing items: %w", err)
	}

	if len(items) == 0 {
		out.Info("Trash is empty.")
		return nil
	}

	out.Result(items, func(w io.Writer) {
		table := tablewriter.NewWriter(w)
		table.SetHeader([]string{"ID", "Name", "Type", "Original Path", "Size", "Trashed At"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		for _, item := range items {
			table.Append([]string{
				shortID(item.ID),
				item.Name,
				string(item.Type),
				item.OriginalPath,
				formatSize(item.SizeBytes),
				item.TrashedAt,
			})
		}
		table.Render()
	})

	return nil
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
