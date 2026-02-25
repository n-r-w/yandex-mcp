// Package apihelpers provides shared HTTP request helpers for Yandex API adapters.
package apihelpers

import (
	"context"
	"net/http"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=apihelpers

// ITokenProvider provides IAM tokens for API authentication.
type ITokenProvider interface {
	// Token returns a valid IAM token, refreshing if needed.
	Token(ctx context.Context, forceRefresh bool) (string, error)
}

// IHTTPDoer abstracts HTTP request execution for client-level behavior and testability.
type IHTTPDoer interface {
	// Do sends an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)
}
