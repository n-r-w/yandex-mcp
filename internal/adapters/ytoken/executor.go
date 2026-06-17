package ytoken

import (
	"context"
	"fmt"
	"os/exec"
)

// commandExecutor implements ICommandExecutor using os/exec.
type commandExecutor struct {
	cliProfile string
}

var _ ICommandExecutor = (*commandExecutor)(nil)

// newCommandExecutor creates a new commandExecutor.
func newCommandExecutor(cliProfile string) *commandExecutor {
	return &commandExecutor{cliProfile: cliProfile}
}

// Execute runs a command and returns its stdout output.
func (e *commandExecutor) Execute(ctx context.Context) ([]byte, error) {
	args := []string{ycCommandArgIAM, ycCommandArgCreateToken}
	if e.cliProfile != "" {
		args = append(args, ycCommandArgProfile, e.cliProfile)
	}

	cmd := exec.CommandContext(ctx, ycCommandName, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}
	return output, nil
}
