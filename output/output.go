package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Mode represents the output verbosity mode
type Mode int

const (
	ModeNormal  Mode = iota
	ModeQuiet
	ModeVerbose
	ModeJSON
)

// Writer wraps an io.Writer with mode-aware output methods
type Writer struct {
	w    io.Writer
	mode Mode
}

// NewWriter creates a new mode-aware Writer
func NewWriter(w io.Writer, mode Mode) *Writer {
	return &Writer{w: w, mode: mode}
}

// Info prints informational messages (normal + verbose modes only)
func (o *Writer) Info(format string, args ...any) {
	if o.mode == ModeQuiet || o.mode == ModeJSON {
		return
	}
	fmt.Fprintf(o.w, format+"\n", args...)
}

// Verbose prints messages only in verbose mode
func (o *Writer) Verbose(format string, args ...any) {
	if o.mode != ModeVerbose {
		return
	}
	fmt.Fprintf(o.w, format+"\n", args...)
}

// Result renders structured data: JSON object in JSON mode, or textFn callback otherwise
func (o *Writer) Result(data any, textFn func(w io.Writer)) {
	if o.mode == ModeJSON {
		enc := json.NewEncoder(o.w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(data)
		return
	}
	if o.mode != ModeQuiet {
		textFn(o.w)
	}
}

// IsQuiet returns true if the writer is in quiet mode
func (o *Writer) IsQuiet() bool {
	return o.mode == ModeQuiet
}

// GetMode returns the current output mode
func (o *Writer) GetMode() Mode {
	return o.mode
}
