package authagent

import "context"

//go:generate go tool mockgen -source=interfaces.go -destination=mock_interfaces.go -package=authagent

// ITokenSource obtains a workstation token for an allowed profile.
type ITokenSource interface {
	Acquire(ctx context.Context, profile string) (string, error)
}
