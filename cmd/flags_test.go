package cmd

import (
	"testing"

	"github.com/AmadlaOrg/garbage/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRegisterGlobalFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	RegisterGlobalFlags(rootCmd)

	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("json"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("quiet"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("dry-run"))
}

func TestStdoutWriter_Normal(t *testing.T) {
	flagJSON = false
	flagQuiet = false
	flagVerbose = false
	defer func() { flagJSON = false; flagQuiet = false; flagVerbose = false }()

	w := StdoutWriter()
	assert.Equal(t, output.ModeNormal, w.GetMode())
}

func TestStdoutWriter_JSON(t *testing.T) {
	flagJSON = true
	flagQuiet = false
	flagVerbose = false
	defer func() { flagJSON = false }()

	w := StdoutWriter()
	assert.Equal(t, output.ModeJSON, w.GetMode())
}

func TestStdoutWriter_Quiet(t *testing.T) {
	flagJSON = false
	flagQuiet = true
	flagVerbose = false
	defer func() { flagQuiet = false }()

	w := StdoutWriter()
	assert.Equal(t, output.ModeQuiet, w.GetMode())
}

func TestStdoutWriter_Verbose(t *testing.T) {
	flagJSON = false
	flagQuiet = false
	flagVerbose = true
	defer func() { flagVerbose = false }()

	w := StdoutWriter()
	assert.Equal(t, output.ModeVerbose, w.GetMode())
}

func TestStdoutWriter_JSONPrecedence(t *testing.T) {
	flagJSON = true
	flagQuiet = true
	flagVerbose = true
	defer func() { flagJSON = false; flagQuiet = false; flagVerbose = false }()

	w := StdoutWriter()
	assert.Equal(t, output.ModeJSON, w.GetMode(), "JSON should take precedence")
}
