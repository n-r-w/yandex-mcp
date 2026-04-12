package wiki

import (
	"context"
	"errors"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// normalizePageLookup keeps shared page lookup inputs consistent before adapter calls.
func normalizePageLookup(
	fields []string,
	value string,
	revisionID string,
	raiseOnRedirect bool,
) (string, domain.WikiGetPageOpts) {
	fields = helpers.TrimStrings(fields)
	helpers.TrimStringFields(&value, &revisionID)

	return value, domain.WikiGetPageOpts{
		Fields:          fields,
		RevisionID:      revisionID,
		RaiseOnRedirect: raiseOnRedirect,
	}
}

// getPageBySlug retrieves a Wiki page by its slug.
func (r *Registrator) getPageBySlug(ctx context.Context, input getPageBySlugInputDTO) (*pageOutputDTO, error) {
	slug, opts := normalizePageLookup(input.Fields, input.Slug, input.RevisionID, input.RaiseOnRedirect)
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	page, err := r.adapter.GetPageBySlug(ctx, slug, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(page), nil
}

// getPageByID retrieves a Wiki page by its ID.
func (r *Registrator) getPageByID(ctx context.Context, input getPageByIDInputDTO) (*pageOutputDTO, error) {
	pageID, opts := normalizePageLookup(input.Fields, input.PageID, input.RevisionID, input.RaiseOnRedirect)
	if pageID == "" {
		return nil, errors.New("page_id is required")
	}

	page, err := r.adapter.GetPageByID(ctx, pageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(page), nil
}

// listResources lists resources (attachments, grids) for a page.
func (r *Registrator) listResources(ctx context.Context, input listResourcesInputDTO) (*resourcesListOutputDTO, error) {
	helpers.TrimStringFields(
		&input.PageID,
		&input.Cursor,
		&input.OrderBy,
		&input.OrderDirection,
		&input.Q,
		&input.Types,
	)

	if input.PageID == "" {
		return nil, errors.New("page_id is required")
	}

	if input.PageSize < 0 {
		return nil, errors.New("page_size must be non-negative")
	}

	if input.PageSize > maxPageSize {
		return nil, fmt.Errorf("page_size must not exceed %d", maxPageSize)
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
	helpers.TrimStringFields(&input.PageID, &input.Cursor, &input.OrderBy, &input.OrderDirection)

	if input.PageID == "" {
		return nil, errors.New("page_id is required")
	}

	if input.PageSize < 0 {
		return nil, errors.New("page_size must be non-negative")
	}

	if input.PageSize > maxPageSize {
		return nil, fmt.Errorf("page_size must not exceed %d", maxPageSize)
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

// normalizeDescendantsOpts validates and builds shared options for descendants listing.
func normalizeDescendantsOpts(
	actuality, cursor string, pageSize int,
) (domain.WikiListDescendantsOpts, error) {
	if pageSize < 0 {
		return domain.WikiListDescendantsOpts{}, errors.New("page_size must be non-negative")
	}

	if pageSize > maxDescendantsPageSize {
		return domain.WikiListDescendantsOpts{}, fmt.Errorf("page_size must not exceed %d", maxDescendantsPageSize)
	}

	return domain.WikiListDescendantsOpts{
		Actuality: actuality,
		Cursor:    cursor,
		PageSize:  pageSize,
	}, nil
}

// listDescendants lists subpages of a Wiki page by its slug. Empty slug lists from root.
func (r *Registrator) listDescendants(
	ctx context.Context, input listDescendantsInputDTO,
) (*descendantsListOutputDTO, error) {
	helpers.TrimStringFields(&input.Slug, &input.Actuality, &input.Cursor)

	opts, err := normalizeDescendantsOpts(input.Actuality, input.Cursor, input.PageSize)
	if err != nil {
		return nil, err
	}

	result, err := r.adapter.ListDescendantsBySlug(ctx, input.Slug, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapDescendantsPageToOutput(result), nil
}

// listDescendantsByID lists subpages of a Wiki page by its ID.
func (r *Registrator) listDescendantsByID(
	ctx context.Context, input listDescendantsByIDInputDTO,
) (*descendantsListOutputDTO, error) {
	helpers.TrimStringFields(&input.PageID, &input.Actuality, &input.Cursor)

	if input.PageID == "" {
		return nil, errors.New("page_id is required")
	}

	opts, err := normalizeDescendantsOpts(input.Actuality, input.Cursor, input.PageSize)
	if err != nil {
		return nil, err
	}

	result, err := r.adapter.ListDescendantsByID(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapDescendantsPageToOutput(result), nil
}

// getGrid retrieves a dynamic table by its ID.
func (r *Registrator) getGrid(ctx context.Context, input getGridInputDTO) (*gridOutputDTO, error) {
	input.Fields = helpers.TrimStrings(input.Fields)
	helpers.TrimStringFields(
		&input.GridID,
		&input.Filter,
		&input.OnlyCols,
		&input.OnlyRows,
		&input.Revision,
		&input.Sort,
	)

	if input.GridID == "" {
		return nil, errors.New("grid_id is required")
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
