package ytoken

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommandExecutor_ExecuteWrapsExecutionError verifies command execution failures are wrapped with context.
func TestCommandExecutor_ExecuteWrapsExecutionError(t *testing.T) {
	t.Setenv("PATH", "")
	exec := newCommandExecutor("")

	_, err := exec.Execute(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command execution failed")
}

// TestCommandExecutor_ExecuteUsesConfiguredProfile verifies IAM token retrieval uses the configured yc profile.
func TestCommandExecutor_ExecuteUsesConfiguredProfile(t *testing.T) {
	binDir := t.TempDir()
	argsPath := filepath.Join(t.TempDir(), "args.txt")
	ycPath := filepath.Join(binDir, ycCommandName)
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"" + argsPath + "\"\nprintf 'test-token\\n'\n"
	require.NoError(t, os.WriteFile(ycPath, []byte(script), 0o700))
	t.Setenv("PATH", binDir)

	exec := newCommandExecutor("gamepult")

	output, err := exec.Execute(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "test-token", strings.TrimSpace(string(output)))

	args, err := os.ReadFile(argsPath)
	require.NoError(t, err)
	assert.Equal(t, "iam create-token --profile gamepult", strings.TrimSpace(string(args)))
}
