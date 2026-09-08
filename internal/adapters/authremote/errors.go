package authremote

import "errors"

var (
	errUnavailable    = errors.New("workstation auth-agent or SSH connection unavailable")
	errProtocol       = errors.New("invalid auth-agent response")
	errInvalidRequest = errors.New("invalid token request")
)
