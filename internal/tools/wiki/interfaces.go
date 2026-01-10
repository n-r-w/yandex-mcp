// Package wiki provides MCP tool handlers for Yandex Wiki operations.
package wiki

import (
	"context"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=wiki

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
	CreatePage(ctx context.Context, req *domain.WikiPageCreateRequest) (*domain.WikiPageCreateResponse, error)
	UpdatePage(ctx context.Context, req *domain.WikiPageUpdateRequest) (*domain.WikiPageUpdateResponse, error)
	AppendPage(ctx context.Context, req *domain.WikiPageAppendRequest) (*domain.WikiPageAppendResponse, error)
	CreateGrid(ctx context.Context, req *domain.WikiGridCreateRequest) (*domain.WikiGridCreateResponse, error)
	UpdateGridCells(
		ctx context.Context, req *domain.WikiGridCellsUpdateRequest,
	) (*domain.WikiGridCellsUpdateResponse, error)
	DeletePage(ctx context.Context, pageID string) (*domain.WikiPageDeleteResponse, error)
	ClonePage(ctx context.Context, req domain.WikiPageCloneRequest) (*domain.WikiCloneOperationResponse, error)
	DeleteGrid(ctx context.Context, gridID string) error
	CloneGrid(ctx context.Context, req domain.WikiGridCloneRequest) (*domain.WikiCloneOperationResponse, error)
	AddGridRows(ctx context.Context, req domain.WikiGridRowsAddRequest) (*domain.WikiGridRowsAddResponse, error)
	DeleteGridRows(ctx context.Context, req domain.WikiGridRowsDeleteRequest) (*domain.WikiRevisionResponse, error)
	MoveGridRows(ctx context.Context, req domain.WikiGridRowsMoveRequest) (*domain.WikiRevisionResponse, error)
	AddGridColumns(ctx context.Context, req domain.WikiGridColumnsAddRequest) (*domain.WikiRevisionResponse, error)
	DeleteGridColumns(ctx context.Context, req domain.WikiGridColumnsDeleteRequest) (*domain.WikiRevisionResponse, error)
	MoveGridColumns(ctx context.Context, req domain.WikiGridColumnsMoveRequest) (*domain.WikiRevisionResponse, error)
}
