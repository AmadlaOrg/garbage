package main

import (
	"github.com/AmadlaOrg/LibraryFramework/cli"
	"github.com/AmadlaOrg/garbage/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cli.New(
		"garbage",
		"Garbage",
		"1.0.0",
		func(rootCmd *cobra.Command) {
			cmd.RegisterGlobalFlags(rootCmd)
			rootCmd.AddCommand(cmd.RmCmd)
			rootCmd.AddCommand(cmd.ListCmd)
			rootCmd.AddCommand(cmd.RestoreCmd)
			rootCmd.AddCommand(cmd.EmptyCmd)
			rootCmd.AddCommand(cmd.InfoCmd)
			rootCmd.AddCommand(cmd.SettingsCmd)
		})
}
