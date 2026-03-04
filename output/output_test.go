package output

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeNormal)
	assert.NotNil(t, w)
	assert.Equal(t, ModeNormal, w.GetMode())
}

func TestInfo_NormalMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeNormal)
	w.Info("hello %s", "world")
	assert.Equal(t, "hello world\n", buf.String())
}

func TestInfo_VerboseMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeVerbose)
	w.Info("hello %s", "world")
	assert.Equal(t, "hello world\n", buf.String())
}

func TestInfo_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeQuiet)
	w.Info("hello %s", "world")
	assert.Empty(t, buf.String())
}

func TestInfo_JSONMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeJSON)
	w.Info("hello %s", "world")
	assert.Empty(t, buf.String())
}

func TestVerbose_VerboseMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeVerbose)
	w.Verbose("debug: %d", 42)
	assert.Equal(t, "debug: 42\n", buf.String())
}

func TestVerbose_NormalMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeNormal)
	w.Verbose("debug: %d", 42)
	assert.Empty(t, buf.String())
}

func TestVerbose_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeQuiet)
	w.Verbose("debug: %d", 42)
	assert.Empty(t, buf.String())
}

func TestVerbose_JSONMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeJSON)
	w.Verbose("debug: %d", 42)
	assert.Empty(t, buf.String())
}

func TestResult_NormalMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeNormal)

	data := map[string]string{"name": "curl"}
	w.Result(data, func(w io.Writer) {
		_, _ = io.WriteString(w, "Name: curl\n")
	})
	assert.Equal(t, "Name: curl\n", buf.String())
}

func TestResult_JSONMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeJSON)

	data := map[string]string{"name": "curl"}
	w.Result(data, func(w io.Writer) {
		_, _ = io.WriteString(w, "Name: curl\n")
	})
	assert.Contains(t, buf.String(), `"name": "curl"`)
	assert.NotContains(t, buf.String(), "Name: curl")
}

func TestResult_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeQuiet)

	data := map[string]string{"name": "curl"}
	w.Result(data, func(w io.Writer) {
		_, _ = io.WriteString(w, "Name: curl\n")
	})
	assert.Empty(t, buf.String())
}

func TestResult_VerboseMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ModeVerbose)

	data := map[string]string{"name": "curl"}
	w.Result(data, func(w io.Writer) {
		_, _ = io.WriteString(w, "Name: curl\n")
	})
	assert.Equal(t, "Name: curl\n", buf.String())
}

func TestIsQuiet(t *testing.T) {
	tests := []struct {
		mode Mode
		want bool
	}{
		{ModeNormal, false},
		{ModeVerbose, false},
		{ModeQuiet, true},
		{ModeJSON, false},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		w := NewWriter(&buf, tt.mode)
		assert.Equal(t, tt.want, w.IsQuiet())
	}
}

func TestGetMode(t *testing.T) {
	modes := []Mode{ModeNormal, ModeQuiet, ModeVerbose, ModeJSON}
	for _, m := range modes {
		var buf bytes.Buffer
		w := NewWriter(&buf, m)
		assert.Equal(t, m, w.GetMode())
	}
}
