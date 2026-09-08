package authremote

import (
	"fmt"
	"net/http"
	"strings"
)

type tokenRequest struct {
	Profile string `json:"profile"`
}

type tokenResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

// result returns the token or the original diagnostic with its HTTP status.
func (r tokenResponse) result(status int) (string, error) {
	if status == http.StatusOK {
		if r.Token == "" || r.Message != "" || strings.ContainsAny(r.Token, " \t\r\n") {
			return "", errProtocol
		}
		return r.Token, nil
	}
	if r.Token != "" || r.Message == "" {
		return "", errProtocol
	}
	return "", fmt.Errorf("auth-agent HTTP %d: %s", status, r.Message)
}
