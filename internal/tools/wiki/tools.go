package wiki

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/tools/helpers"
)

// GetPageBySlug retrieves a Wiki page by its slug.
func (r *Registrator) GetPageBySlug(ctx context.Context, input GetPageBySlugInput) (*PageOutput, error) {
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

// GetPageByID retrieves a Wiki page by its ID.
func (r *Registrator) GetPageByID(ctx context.Context, input GetPageByIDInput) (*PageOutput, error) {
	if input.PageID <= 0 {
		return nil, r.logError(ctx, errors.New("page_id must be positive"))
	}

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

// ListResources lists resources (attachments, grids) for a page.
func (r *Registrator) ListResources(ctx context.Context, input ListResourcesInput) (*ResourcesListOutput, error) {
	if input.PageID <= 0 {
		return nil, r.logError(ctx, errors.New("page_id must be positive"))
	}

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
		PageIDLegacy:   input.PageIDLegacy,
	}

	result, err := r.adapter.ListPageResources(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapResourcesPageToOutput(result), nil
}

// ListGrids lists dynamic tables (grids) for a page.
func (r *Registrator) ListGrids(ctx context.Context, input ListGridsInput) (*GridsListOutput, error) {
	if input.PageID <= 0 {
		return nil, r.logError(ctx, errors.New("page_id must be positive"))
	}

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
		PageIDLegacy:   input.PageIDLegacy,
	}

	result, err := r.adapter.ListPageGrids(ctx, input.PageID, opts)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapGridsPageToOutput(result), nil
}

// GetGrid retrieves a dynamic table by its ID.
func (r *Registrator) GetGrid(ctx context.Context, input GetGridInput) (*GridOutput, error) {
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

// CreatePage creates a new wiki page.
func (r *Registrator) CreatePage(ctx context.Context, input CreatePageInput) (*PageOutput, error) {
	if input.Slug == "" {
		return nil, r.logError(ctx, errors.New("slug is required"))
	}

	if input.Title == "" {
		return nil, r.logError(ctx, errors.New("title is required"))
	}

	if input.PageType == "" {
		return nil, r.logError(ctx, errors.New("page_type is required"))
	}

	req := &domain.WikiPageCreateRequest{ //nolint:exhaustruct // CloudPage set conditionally
		Slug:       input.Slug,
		Title:      input.Title,
		Content:    input.Content,
		PageType:   input.PageType,
		IsSilent:   input.IsSilent,
		Fields:     input.Fields,
		GridFormat: input.GridFormat,
	}

	if input.CloudPage != nil {
		req.CloudPage = &domain.WikiCloudPageInput{
			Method:  input.CloudPage.Method,
			Doctype: input.CloudPage.Doctype,
		}
	}

	result, err := r.adapter.CreatePage(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(&result.Page), nil
}

// UpdatePage updates an existing wiki page.
func (r *Registrator) UpdatePage(ctx context.Context, input UpdatePageInput) (*PageOutput, error) {
	if input.PageID <= 0 {
		return nil, r.logError(ctx, errors.New("page_id must be positive"))
	}

	if input.Title == "" && input.Content == "" && input.Redirect == nil {
		return nil, r.logError(ctx, errors.New("at least one of title, content, or redirect is required"))
	}

	req := &domain.WikiPageUpdateRequest{ //nolint:exhaustruct // Redirect set conditionally
		PageID:     input.PageID,
		Title:      input.Title,
		Content:    input.Content,
		AllowMerge: input.AllowMerge,
		IsSilent:   input.IsSilent,
		Fields:     input.Fields,
	}

	if input.Redirect != nil {
		req.Redirect = &domain.WikiRedirectInput{
			PageID: input.Redirect.PageID,
			Slug:   input.Redirect.Slug,
		}
	}

	result, err := r.adapter.UpdatePage(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(&result.Page), nil
}

// AppendPage appends content to an existing wiki page.
func (r *Registrator) AppendPage(ctx context.Context, input AppendPageInput) (*PageOutput, error) {
	if input.PageID <= 0 {
		return nil, r.logError(ctx, errors.New("page_id must be positive"))
	}

	if input.Content == "" {
		return nil, r.logError(ctx, errors.New("content is required"))
	}

	req := &domain.WikiPageAppendRequest{ //nolint:exhaustruct // Body, Section, Anchor set conditionally
		PageID:   input.PageID,
		Content:  input.Content,
		IsSilent: input.IsSilent,
		Fields:   input.Fields,
	}

	if input.Body != nil {
		req.Body = &domain.WikiBodyLocation{
			Location: input.Body.Location,
		}
	}

	if input.Section != nil {
		req.Section = &domain.WikiSectionLocation{
			ID:       input.Section.ID,
			Location: input.Section.Location,
		}
	}

	if input.Anchor != nil {
		req.Anchor = &domain.WikiAnchorLocation{
			Name:     input.Anchor.Name,
			Fallback: input.Anchor.Fallback,
			Regex:    input.Anchor.Regex,
		}
	}

	result, err := r.adapter.AppendPage(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapPageToOutput(&result.Page), nil
}

// CreateGrid creates a new dynamic table (grid).
func (r *Registrator) CreateGrid(ctx context.Context, input CreateGridInput) (*GridOutput, error) {
	if input.Page.ID <= 0 && input.Page.Slug == "" {
		return nil, r.logError(ctx, errors.New("page.id or page.slug is required"))
	}

	if input.Title == "" {
		return nil, r.logError(ctx, errors.New("title is required"))
	}

	if len(input.Columns) == 0 {
		return nil, r.logError(ctx, errors.New("at least one column is required"))
	}

	columns := make([]domain.WikiColumnDefinition, 0, len(input.Columns))
	for _, col := range input.Columns {
		if col.Slug == "" {
			return nil, r.logError(ctx, errors.New("column slug is required"))
		}

		if col.Title == "" {
			return nil, r.logError(ctx, errors.New("column title is required"))
		}

		columns = append(columns, domain.WikiColumnDefinition{
			Slug:  col.Slug,
			Title: col.Title,
			Type:  col.Type,
		})
	}

	// Resolve page ID from slug if not provided directly
	pageID := input.Page.ID
	if pageID <= 0 {
		page, err := r.adapter.GetPageBySlug(ctx, input.Page.Slug, domain.WikiGetPageOpts{
			Fields:          nil,
			RevisionID:      0,
			RaiseOnRedirect: false,
		})
		if err != nil {
			return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
		}
		pageID = page.ID
	}

	fields := splitFields(input.Fields)

	req := &domain.WikiGridCreateRequest{
		PageID:  pageID,
		Title:   input.Title,
		Columns: columns,
		Fields:  fields,
	}

	result, err := r.adapter.CreateGrid(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapGridToOutput(&result.Grid), nil
}

// UpdateGridCells updates cells in a dynamic table (grid).
func (r *Registrator) UpdateGridCells(ctx context.Context, input UpdateGridCellsInput) (*GridOutput, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.Cells) == 0 {
		return nil, r.logError(ctx, errors.New("at least one cell is required"))
	}

	cells := make([]domain.WikiCellUpdate, 0, len(input.Cells))
	for i, cell := range input.Cells {
		if cell.RowID <= 0 {
			return nil, r.logError(ctx, fmt.Errorf("cell[%d]: row_id must be positive", i))
		}

		if cell.ColumnSlug == "" {
			return nil, r.logError(ctx, fmt.Errorf("cell[%d]: column_slug is required", i))
		}

		value, ok := cell.Value.(string)
		if !ok {
			return nil, r.logError(ctx, fmt.Errorf("cell[%d]: value must be a string", i))
		}

		cells = append(cells, domain.WikiCellUpdate{
			RowID:      cell.RowID,
			ColumnSlug: cell.ColumnSlug,
			Value:      value,
		})
	}

	req := &domain.WikiGridCellsUpdateRequest{
		GridID:   input.GridID,
		Cells:    cells,
		Revision: input.Revision,
	}

	result, err := r.adapter.UpdateGridCells(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return mapGridToOutput(&result.Grid), nil
}

// splitFields splits a comma-separated fields string into a slice.
func splitFields(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
