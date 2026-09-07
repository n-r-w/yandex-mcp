// Package wiki provides MCP tool handlers for Yandex Wiki operations.
package wiki

import (
	"context"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

//go:generate go tool mockgen -source=interfaces.go -destination=mock_interfaces.go -package=wiki

// IWikiAdapter defines the interface for Wiki adapter operations consumed by tools.
type IWikiAdapter interface {
	GetPageBySlug(ctx context.Context, slug string, opts domain.WikiGetPageOpts) (*domain.WikiPage, error)
	GetPageByID(ctx context.Context, id string, opts domain.WikiGetPageOpts) (*domain.WikiPage, error)
	ListPageResources(
		ctx context.Context, pageID string, opts domain.WikiListResourcesOpts,
	) (*domain.WikiResourcesPage, error)
	ListPageGrids(
		ctx context.Context, pageID string, opts domain.WikiListGridsOpts,
	) (*domain.WikiGridsPage, error)
	GetGridByID(ctx context.Context, gridID string, opts domain.WikiGetGridOpts) (*domain.WikiGrid, error)
	ListDescendantsBySlug(
		ctx context.Context, slug string, opts domain.WikiListDescendantsOpts,
	) (*domain.WikiDescendantsPage, error)
	ListDescendantsByID(
		ctx context.Context, id string, opts domain.WikiListDescendantsOpts,
	) (*domain.WikiDescendantsPage, error)
}
