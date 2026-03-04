package cmd

import (
	"os"

	"github.com/AmadlaOrg/garbage/output"
	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagQuiet   bool
	flagVerbose bool
	flagDryRun  bool
)

// RegisterGlobalFlags adds persistent flags to the root command
func RegisterGlobalFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "Suppress non-error output")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Show detailed output")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be done without making changes")
}

// StdoutWriter creates an output.Writer from the current global flag state
func StdoutWriter() *output.Writer {
	mode := output.ModeNormal
	switch {
	case flagJSON:
		mode = output.ModeJSON
	case flagQuiet:
		mode = output.ModeQuiet
	case flagVerbose:
		mode = output.ModeVerbose
	}
	return output.NewWriter(os.Stdout, mode)
}
