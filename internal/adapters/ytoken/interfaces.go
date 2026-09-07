package ytoken

import "context"

//go:generate go tool mockgen -source=interfaces.go -destination=mock_interfaces.go -package=ytoken

// ICommandExecutor abstracts shell command execution for testability.
type ICommandExecutor interface {
	// Execute runs a command and returns its stdout output or an error.
	Execute(ctx context.Context) ([]byte, error)
}
