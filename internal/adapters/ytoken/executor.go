package ytoken

import (
	"context"
	"fmt"
	"os/exec"
)

// commandExecutor implements ICommandExecutor using os/exec.
type commandExecutor struct{}

var _ ICommandExecutor = (*commandExecutor)(nil)

// newCommandExecutor creates a new commandExecutor.
func newCommandExecutor() *commandExecutor {
	return &commandExecutor{}
}

// Execute runs a command and returns its stdout output.
func (e *commandExecutor) Execute(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, ycCommandName, ycCommandArgIAM, ycCommandArgCreateToken)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}
	return output, nil
}
