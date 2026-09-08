package ytoken

import "context"

//go:generate go tool mockgen -source=interfaces.go -destination=mock_interfaces.go -package=ytoken

// ITokenSource acquires an IAM token for a profile without caching it.
type ITokenSource interface {
	Acquire(ctx context.Context, profile string) (string, error)
}
