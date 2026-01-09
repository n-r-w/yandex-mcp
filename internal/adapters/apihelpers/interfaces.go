// Package apihelpers provides shared HTTP request helpers for Yandex API adapters.
package apihelpers

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=apihelpers

// ITokenProvider provides IAM tokens for API authentication.
type ITokenProvider interface {
	// Token returns a valid IAM token, refreshing if needed.
	Token(ctx context.Context) (string, error)

	// ForceRefresh forces a token refresh and returns the new token.
	ForceRefresh(ctx context.Context) (string, error)
}
