package ytoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandExecutor_ExecuteWrapsExecutionError verifies command execution failures are wrapped with context.
func TestCommandExecutor_ExecuteWrapsExecutionError(t *testing.T) {
	t.Setenv("PATH", "")
	exec := newCommandExecutor()

	_, err := exec.Execute(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command execution failed")
}
