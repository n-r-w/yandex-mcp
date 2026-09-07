package authremote

import "errors"

var (
	errUnavailable    = errors.New("workstation auth-agent or SSH connection unavailable")
	errProtocol       = errors.New("invalid auth-agent response")
	errIncompatible   = errors.New("incompatible auth-agent protocol")
	errInvalidRequest = errors.New("invalid token request")
	errForbidden      = errors.New("profile access denied")
	errAuthentication = errors.New("workstation authentication failed")
	errInternal       = errors.New("auth-agent internal error")
)
