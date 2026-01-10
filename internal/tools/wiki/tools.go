package wiki

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		PageIDLegacy:   input.PageIDLegacy,
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
		PageIDLegacy:   input.PageIDLegacy,
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

// createPage creates a new wiki page.
func (r *Registrator) createPage(ctx context.Context, input createPageInputDTO) (*pageOutputDTO, error) {
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

// updatePage updates an existing wiki page.
func (r *Registrator) updatePage(ctx context.Context, input updatePageInputDTO) (*pageOutputDTO, error) {
	if input.Title == "" && input.Content == "" && input.Redirect == nil {
		return nil, r.logError(ctx, errors.New("at least one of title, content, or redirect is required"))
	}

	if input.PageID == "" {
		return nil, r.logError(ctx, errors.New("page_id is required"))
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

// appendPage appends content to an existing wiki page.
func (r *Registrator) appendPage(ctx context.Context, input appendPageInputDTO) (*pageOutputDTO, error) {
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

// createGrid creates a new dynamic table (grid).
func (r *Registrator) createGrid(ctx context.Context, input createGridInputDTO) (*gridOutputDTO, error) {
	if input.Page.ID == "" && input.Page.Slug == "" {
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
	if pageID == "" {
		page, err := r.adapter.GetPageBySlug(ctx, input.Page.Slug, domain.WikiGetPageOpts{
			Fields:          nil,
			RevisionID:      "",
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

// updateGridCells updates cells in a dynamic table (grid).
func (r *Registrator) updateGridCells(ctx context.Context, input updateGridCellsInputDTO) (*gridOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.Cells) == 0 {
		return nil, r.logError(ctx, errors.New("at least one cell is required"))
	}

	cells := make([]domain.WikiCellUpdate, 0, len(input.Cells))
	for i, cell := range input.Cells {
		if cell.RowID == "" {
			return nil, r.logError(ctx, fmt.Errorf("cell[%d]: row_id is required", i))
		}

		if cell.ColumnSlug == "" {
			return nil, r.logError(ctx, fmt.Errorf("cell[%d]: column_slug is required", i))
		}

		cells = append(cells, domain.WikiCellUpdate{
			RowID:      cell.RowID,
			ColumnSlug: cell.ColumnSlug,
			Value:      cell.Value,
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

// deletePage deletes a wiki page by its ID.
func (r *Registrator) deletePage(ctx context.Context, input deletePageInputDTO) (*deletePageOutputDTO, error) {
	if input.PageID == "" {
		return nil, r.logError(ctx, errors.New("page_id is required"))
	}

	resp, err := r.adapter.DeletePage(ctx, input.PageID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &deletePageOutputDTO{
		RecoveryToken: resp.RecoveryToken,
	}, nil
}

// clonePage initiates an async clone operation for a wiki page.
func (r *Registrator) clonePage(ctx context.Context, input clonePageInputDTO) (*cloneOperationOutputDTO, error) {
	if input.PageID == "" {
		return nil, r.logError(ctx, errors.New("page_id is required"))
	}

	if input.Target == "" {
		return nil, r.logError(ctx, errors.New("target is required"))
	}

	req := domain.WikiPageCloneRequest{
		PageID:      input.PageID,
		Target:      input.Target,
		Title:       input.Title,
		SubscribeMe: input.SubscribeMe,
	}

	resp, err := r.adapter.ClonePage(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &cloneOperationOutputDTO{
		OperationID:   resp.OperationID,
		OperationType: resp.OperationType,
		DryRun:        resp.DryRun,
		StatusURL:     resp.StatusURL,
	}, nil
}

// deleteGrid deletes a wiki grid by its ID.
func (r *Registrator) deleteGrid(ctx context.Context, input deleteGridInputDTO) (*deleteGridOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	err := r.adapter.DeleteGrid(ctx, input.GridID)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &deleteGridOutputDTO{}, nil
}

// cloneGrid initiates an async clone operation for a wiki grid.
func (r *Registrator) cloneGrid(ctx context.Context, input cloneGridInputDTO) (*cloneOperationOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if input.Target == "" {
		return nil, r.logError(ctx, errors.New("target is required"))
	}

	req := domain.WikiGridCloneRequest{
		GridID:   input.GridID,
		Target:   input.Target,
		Title:    input.Title,
		WithData: input.WithData,
	}

	resp, err := r.adapter.CloneGrid(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &cloneOperationOutputDTO{
		OperationID:   resp.OperationID,
		OperationType: resp.OperationType,
		DryRun:        resp.DryRun,
		StatusURL:     resp.StatusURL,
	}, nil
}

// addGridRows adds rows to a wiki grid.
func (r *Registrator) addGridRows(ctx context.Context, input addGridRowsInputDTO) (*addGridRowsOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.Rows) == 0 {
		return nil, r.logError(ctx, errors.New("rows must contain at least one element"))
	}

	req := domain.WikiGridRowsAddRequest{
		GridID:     input.GridID,
		Rows:       input.Rows,
		AfterRowID: input.AfterRowID,
		Position:   input.Position,
		Revision:   input.Revision,
	}

	resp, err := r.adapter.AddGridRows(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	results := make([]gridRowResultItemDTO, len(resp.Results))
	for i, row := range resp.Results {
		results[i] = gridRowResultItemDTO{
			ID:     row.ID,
			Row:    row.Row,
			Color:  row.Color,
			Pinned: row.Pinned,
		}
	}

	return &addGridRowsOutputDTO{
		Revision: resp.Revision,
		Results:  results,
	}, nil
}

// deleteGridRows deletes rows from a wiki grid.
func (r *Registrator) deleteGridRows(ctx context.Context, input deleteGridRowsInputDTO) (*revisionOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.RowIDs) == 0 {
		return nil, r.logError(ctx, errors.New("row_ids must contain at least one element"))
	}

	req := domain.WikiGridRowsDeleteRequest{
		GridID:   input.GridID,
		RowIDs:   input.RowIDs,
		Revision: input.Revision,
	}

	resp, err := r.adapter.DeleteGridRows(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &revisionOutputDTO{
		Revision: resp.Revision,
	}, nil
}

// moveGridRows moves rows within a wiki grid.
func (r *Registrator) moveGridRows(ctx context.Context, input moveGridRowsInputDTO) (*revisionOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if input.RowID == "" {
		return nil, r.logError(ctx, errors.New("row_id is required"))
	}

	req := domain.WikiGridRowsMoveRequest{
		GridID:     input.GridID,
		RowID:      input.RowID,
		AfterRowID: input.AfterRowID,
		Position:   input.Position,
		RowsCount:  input.RowsCount,
		Revision:   input.Revision,
	}

	resp, err := r.adapter.MoveGridRows(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &revisionOutputDTO{
		Revision: resp.Revision,
	}, nil
}

// addGridColumns adds columns to a wiki grid.
func (r *Registrator) addGridColumns(ctx context.Context, input addGridColumnsInputDTO) (*revisionOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.Columns) == 0 {
		return nil, r.logError(ctx, errors.New("columns must contain at least one element"))
	}

	cols := make([]domain.WikiNewColumnDefinition, len(input.Columns))
	for i, c := range input.Columns {
		cols[i] = domain.WikiNewColumnDefinition{
			Slug:          c.Slug,
			Title:         c.Title,
			Type:          c.Type,
			Required:      c.Required,
			Description:   c.Description,
			Color:         c.Color,
			Format:        c.Format,
			SelectOptions: c.SelectOptions,
			Multiple:      c.Multiple,
			MarkRows:      c.MarkRows,
			TicketField:   c.TicketField,
			Width:         c.Width,
			WidthUnits:    c.WidthUnits,
			Pinned:        c.Pinned,
		}
	}

	req := domain.WikiGridColumnsAddRequest{
		GridID:   input.GridID,
		Columns:  cols,
		Position: input.Position,
		Revision: input.Revision,
	}

	resp, err := r.adapter.AddGridColumns(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &revisionOutputDTO{
		Revision: resp.Revision,
	}, nil
}

// deleteGridColumns deletes columns from a wiki grid.
func (r *Registrator) deleteGridColumns(
	ctx context.Context, input deleteGridColumnsInputDTO,
) (*revisionOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if len(input.ColumnSlugs) == 0 {
		return nil, r.logError(ctx, errors.New("column_slugs must contain at least one element"))
	}

	req := domain.WikiGridColumnsDeleteRequest{
		GridID:      input.GridID,
		ColumnSlugs: input.ColumnSlugs,
		Revision:    input.Revision,
	}

	resp, err := r.adapter.DeleteGridColumns(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &revisionOutputDTO{
		Revision: resp.Revision,
	}, nil
}

// moveGridColumns moves columns within a wiki grid.
func (r *Registrator) moveGridColumns(ctx context.Context, input moveGridColumnsInputDTO) (*revisionOutputDTO, error) {
	if input.GridID == "" {
		return nil, r.logError(ctx, errors.New("grid_id is required"))
	}

	if input.ColumnSlug == "" {
		return nil, r.logError(ctx, errors.New("column_slug is required"))
	}

	req := domain.WikiGridColumnsMoveRequest{
		GridID:       input.GridID,
		ColumnSlug:   input.ColumnSlug,
		Position:     input.Position,
		ColumnsCount: input.ColumnsCount,
		Revision:     input.Revision,
	}

	resp, err := r.adapter.MoveGridColumns(ctx, req)
	if err != nil {
		return nil, helpers.ToSafeError(ctx, domain.ServiceWiki, err)
	}

	return &revisionOutputDTO{
		Revision: resp.Revision,
	}, nil
}
