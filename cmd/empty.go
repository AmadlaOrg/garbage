package cmd

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/AmadlaOrg/garbage/db"
	"github.com/AmadlaOrg/garbage/trash"
	"github.com/spf13/cobra"
)

// For testing
var (
	emptyOpenDB   = func() (*db.DB, error) { return db.Open() }
	emptyTrashDir = defaultTrashDir
)

var (
	emptyOlderThanFlag string
	emptyForceFlag     bool
)

// EmptyCmd permanently deletes trashed items.
var EmptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Permanently delete trashed items",
	Args:  cobra.NoArgs,
	RunE:  runEmpty,
}

func init() {
	EmptyCmd.Flags().StringVar(&emptyOlderThanFlag, "older-than", "", "Only delete items older than duration (e.g. 30d, 24h)")
	EmptyCmd.Flags().BoolVar(&emptyForceFlag, "force", false, "Skip confirmation prompt")
}

func runEmpty(cmd *cobra.Command, args []string) error {
	out := StdoutWriter()

	d, err := emptyOpenDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	if err := d.Migrate(); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	trashDir, err := emptyTrashDir()
	if err != nil {
		return fmt.Errorf("determining trash dir: %w", err)
	}

	svc := trash.NewService(d.Conn(), trashDir)

	if emptyOlderThanFlag != "" {
		dur, err := parseDuration(emptyOlderThanFlag)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", emptyOlderThanFlag, err)
		}

		if flagDryRun {
			items, err := svc.List(trash.ListFilter{OlderThan: dur})
			if err != nil {
				return err
			}
			out.Info("Would permanently delete %d item(s) older than %s", len(items), emptyOlderThanFlag)
			return nil
		}

		if !emptyForceFlag {
			out.Info("Use --force to confirm permanent deletion")
			return nil
		}

		count, bytes, err := svc.DeleteOlderThan(dur)
		if err != nil {
			return fmt.Errorf("deleting items: %w", err)
		}

		result := map[string]any{"deleted": count, "bytes_freed": bytes}
		out.Result(result, func(w io.Writer) {
			fmt.Fprintf(w, "Permanently deleted %d item(s), freed %s\n", count, formatSize(bytes))
		})

		return nil
	}

	// Delete all
	if flagDryRun {
		count, bytes, err := svc.Stats()
		if err != nil {
			return err
		}
		out.Info("Would permanently delete %d item(s) (%s)", count, formatSize(bytes))
		return nil
	}

	if !emptyForceFlag {
		count, bytes, err := svc.Stats()
		if err != nil {
			return err
		}
		if count == 0 {
			out.Info("Trash is already empty.")
			return nil
		}
		out.Info("Trash contains %d item(s) (%s). Use --force to confirm permanent deletion.", count, formatSize(bytes))
		return nil
	}

	count, bytes, err := svc.DeleteAll()
	if err != nil {
		return fmt.Errorf("emptying trash: %w", err)
	}

	result := map[string]any{"deleted": count, "bytes_freed": bytes}
	out.Result(result, func(w io.Writer) {
		fmt.Fprintf(w, "Permanently deleted %d item(s), freed %s\n", count, formatSize(bytes))
	})

	return nil
}

// parseDuration parses durations like "30d", "24h", "2h30m", or standard Go durations.
func parseDuration(s string) (time.Duration, error) {
	// Handle day suffix (not supported by time.ParseDuration)
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid day count: %s", numStr)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}
