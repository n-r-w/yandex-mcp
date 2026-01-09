// Package tracker provides HTTP client for Yandex Tracker API.
package tracker

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=tracker

// ITokenProvider defines the interface for obtaining IAM tokens.
type ITokenProvider interface {
	// Token returns a valid IAM token, refreshing if needed.
	Token(ctx context.Context) (string, error)

	// ForceRefresh forces a token refresh and returns the new token.
	ForceRefresh(ctx context.Context) (string, error)
}
