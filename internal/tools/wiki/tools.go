package wiki

import (
	"context"
	"errors"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// getPageBySlug retrieves a Wiki page by its slug.
func (r *Registrator) getPageBySlug(ctx context.Context, input getPageBySlugInputDTO) (*pageOutputDTO, error) {
	if input.Slug == "" {
		return nil, r.logError(ctx, errors.New("slug is required"))
	}

	opts := domain.WikiGetPageOpts{
		Fields:          input.Fields,
		RevisionID:      input.RevisionID,
		RaiseOnRedirect: input.RaiseOnRedirect,
	}

	page, err := r.adapter.GetPageBySlug(ctx, input.Slug, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(page), nil
}

// getPageByID retrieves a Wiki page by its ID.
func (r *Registrator) getPageByID(ctx context.Context, input getPageByIDInputDTO) (*pageOutputDTO, error) {
	opts := domain.WikiGetPageOpts{
		Fields:          input.Fields,
		RevisionID:      input.RevisionID,
		RaiseOnRedirect: input.RaiseOnRedirect,
	}

	page, err := r.adapter.GetPageByID(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(page), nil
}

// listResources lists resources (attachments, grids) for a page.
func (r *Registrator) listResources(ctx context.Context, input listResourcesInputDTO) (*resourcesListOutputDTO, error) {
	if input.PageSize < 0 {
		return nil, r.logError(ctx, errors.New("page_size must be non-negative"))
	}

	if input.PageSize > maxPageSize {
		return nil, r.logError(ctx, fmt.Errorf("page_size must not exceed %d", maxPageSize))
	}

	opts := domain.WikiListResourcesOpts{
		Cursor:         input.Cursor,
		PageSize:       input.PageSize,
		OrderBy:        input.OrderBy,
		OrderDirection: input.OrderDirection,
		Query:          input.Q,
		Types:          input.Types,
	}

	result, err := r.adapter.ListPageResources(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapResourcesPageToOutput(result), nil
}

// listGrids lists dynamic tables (grids) for a page.
func (r *Registrator) listGrids(ctx context.Context, input listGridsInputDTO) (*gridsListOutputDTO, error) {
	if input.PageSize < 0 {
		return nil, r.logError(ctx, errors.New("page_size must be non-negative"))
	}

	if input.PageSize > maxPageSize {
		return nil, r.logError(ctx, fmt.Errorf("page_size must not exceed %d", maxPageSize))
	}

	opts := domain.WikiListGridsOpts{
		Cursor:         input.Cursor,
		PageSize:       input.PageSize,
		OrderBy:        input.OrderBy,
		OrderDirection: input.OrderDirection,
	}

	result, err := r.adapter.ListPageGrids(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapGridsPageToOutput(result), nil
}

// getGrid retrieves a dynamic table by its ID.
func (r *Registrator) getGrid(ctx context.Context, input getGridInputDTO) (*gridOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	opts := domain.WikiGetGridOpts{
		Fields:   input.Fields,
		Filter:   input.Filter,
		OnlyCols: input.OnlyCols,
		OnlyRows: input.OnlyRows,
		Revision: input.Revision,
		Sort:     input.Sort,
	}

	grid, err := r.adapter.GetGridByID(ctx, input.GridID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapGridToOutput(grid), nil
}
