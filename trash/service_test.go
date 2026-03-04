package trash

import (
	"testing"

	gdb "github.com/AmadlaOrg/garbage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	d, err := gdb.OpenMemory()
	require.NoError(t, err)
	defer d.Close()

	svc := NewService(d.Conn(), t.TempDir())
	assert.NotNil(t, svc)
}
