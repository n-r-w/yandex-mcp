// Package wiki provides HTTP client for Yandex Wiki API.
package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"
	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	wikitools "github.com/n-r-w/yandex-mcp/internal/tools/wiki"
)

// Client implements IWikiClient for Yandex Wiki API.
type Client struct {
	apiClient *apihelpers.APIClient
	baseURL   string
}

// Compile-time check that Client implements the tools interface.
var _ wikitools.IWikiAdapter = (*Client)(nil)

// NewClient creates a new Wiki API client.
func NewClient(cfg *config.Config, tokenProvider apihelpers.ITokenProvider) *Client {
	client := &Client{
		apiClient: nil, // set below
		baseURL:   strings.TrimSuffix(cfg.WikiBaseURL, "/"),
	}

	client.apiClient = apihelpers.NewAPIClient(apihelpers.APIClientConfig{
		HTTPClient:    nil, // uses default
		TokenProvider: tokenProvider,
		OrgID:         cfg.CloudOrgID,
		ExtraHeaders:  nil,
		ServiceName:   string(domain.ServiceWiki),
		ParseError:    client.parseError,
		HTTPTimeout:   cfg.HTTPTimeout,
	})

	return client
}

// GetPageBySlug retrieves a page by its slug.
func (c *Client) GetPageBySlug(
	ctx context.Context, slug string, opts domain.WikiGetPageOpts,
) (*domain.WikiPage, error) {
	u, err := url.Parse(c.baseURL + "/v1/pages")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	q.Set("slug", slug)
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	if opts.RevisionID != "" {
		q.Set("revision_id", opts.RevisionID)
	}
	if opts.RaiseOnRedirect {
		q.Set("raise_on_redirect", "true")
	}
	u.RawQuery = q.Encode()

	var page pageDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &page, "GetPageBySlug"); err != nil {
		return nil, err
	}
	return pageToWikiPage(&page), nil
}

// GetPageByID retrieves a page by its ID.
func (c *Client) GetPageByID(ctx context.Context, id string, opts domain.WikiGetPageOpts) (*domain.WikiPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%s", c.baseURL, id))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	if opts.RevisionID != "" {
		q.Set("revision_id", opts.RevisionID)
	}
	if opts.RaiseOnRedirect {
		q.Set("raise_on_redirect", "true")
	}
	u.RawQuery = q.Encode()

	var page pageDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &page, "GetPageByID"); err != nil {
		return nil, err
	}
	return pageToWikiPage(&page), nil
}

// ListPageResources lists resources (attachments, grids) for a page.
func (c *Client) ListPageResources(
	ctx context.Context,
	pageID string,
	opts domain.WikiListResourcesOpts,
) (*domain.WikiResourcesPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%s/resources", c.baseURL, pageID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	if opts.PageSize > 0 {
		pageSize := opts.PageSize
		if pageSize > maxResourcesSize {
			pageSize = maxResourcesSize
		}
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	if opts.OrderBy != "" {
		q.Set("order_by", opts.OrderBy)
	}
	if opts.OrderDirection != "" {
		q.Set("order_direction", opts.OrderDirection)
	}
	if opts.Query != "" {
		q.Set("q", opts.Query)
	}
	if opts.Types != "" {
		q.Set("types", opts.Types)
	}
	if opts.PageIDLegacy != "" {
		q.Set("page_id", opts.PageIDLegacy)
	}
	u.RawQuery = q.Encode()

	var resp resourcesResponseDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &resp, "ListPageResources"); err != nil {
		return nil, err
	}

	rp := &resourcesPageDTO{
		Resources:  resp.Items,
		NextCursor: resp.NextCursor,
		PrevCursor: resp.PrevCursor,
	}
	return resourcesPageToWikiResourcesPage(rp)
}

// ListPageGrids lists dynamic tables (grids) for a page.
func (c *Client) ListPageGrids(
	ctx context.Context,
	pageID string,
	opts domain.WikiListGridsOpts,
) (*domain.WikiGridsPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%s/grids", c.baseURL, pageID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if opts.Cursor != "" {
		q.Set("cursor", opts.Cursor)
	}
	if opts.PageSize > 0 {
		pageSize := opts.PageSize
		if pageSize > maxGridsSize {
			pageSize = maxGridsSize
		}
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	if opts.OrderBy != "" {
		q.Set("order_by", opts.OrderBy)
	}
	if opts.OrderDirection != "" {
		q.Set("order_direction", opts.OrderDirection)
	}
	if opts.PageIDLegacy != "" {
		q.Set("page_id", opts.PageIDLegacy)
	}
	u.RawQuery = q.Encode()

	var resp gridsResponseDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &resp, "ListPageGrids"); err != nil {
		return nil, err
	}

	gp := &gridsPageDTO{
		Grids:      resp.Items,
		NextCursor: resp.NextCursor,
		PrevCursor: resp.PrevCursor,
	}
	return gridsPageToWikiGridsPage(gp), nil
}

// GetGridByID retrieves a dynamic table by its ID.
func (c *Client) GetGridByID(
	ctx context.Context,
	gridID string,
	opts domain.WikiGetGridOpts,
) (*domain.WikiGrid, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/grids/%s", c.baseURL, gridID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	if opts.Filter != "" {
		q.Set("filter", opts.Filter)
	}
	if opts.OnlyCols != "" {
		q.Set("only_cols", opts.OnlyCols)
	}
	if opts.OnlyRows != "" {
		q.Set("only_rows", opts.OnlyRows)
	}
	if opts.Revision != "" {
		q.Set("revision", opts.Revision)
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	u.RawQuery = q.Encode()

	var grid gridDTO
	if _, err := c.apiClient.DoGET(ctx, u.String(), &grid, "GetGridByID"); err != nil {
		return nil, err
	}
	return gridToWikiGrid(&grid), nil
}

// CreatePage creates a new wiki page.
func (c *Client) CreatePage(
	ctx context.Context, req *domain.WikiPageCreateRequest,
) (*domain.WikiPageCreateResponse, error) {
	baseURL := c.baseURL + "/v1/pages"

	body := createPageRequestDTO{ //nolint:exhaustruct // CloudPage set conditionally
		Slug:       req.Slug,
		Title:      req.Title,
		Content:    req.Content,
		PageType:   req.PageType,
		GridFormat: req.GridFormat,
	}

	if req.CloudPage != nil {
		body.CloudPage = &cloudPageRequestDTO{
			Method:  req.CloudPage.Method,
			Doctype: req.CloudPage.Doctype,
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if req.IsSilent {
		q.Set("is_silent", "true")
	}
	if len(req.Fields) > 0 {
		q.Set("fields", strings.Join(req.Fields, ","))
	}
	u.RawQuery = q.Encode()

	var page pageDTO
	if _, err := c.apiClient.DoPOST(ctx, u.String(), body, &page, "CreatePage"); err != nil {
		return nil, err
	}

	return &domain.WikiPageCreateResponse{
		Page: *pageToWikiPage(&page),
	}, nil
}

// UpdatePage updates an existing wiki page.
func (c *Client) UpdatePage(
	ctx context.Context, req *domain.WikiPageUpdateRequest,
) (*domain.WikiPageUpdateResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/pages/%s", c.baseURL, req.PageID)

	body := updatePageRequestDTO{ //nolint:exhaustruct // Redirect set conditionally
		Title:   req.Title,
		Content: req.Content,
	}

	if req.Redirect != nil {
		body.Redirect = &redirectRequestDTO{
			Page: &pageIdentityRequestDTO{
				ID:   apihelpers.StringIDFromPointer(req.Redirect.PageID),
				Slug: req.Redirect.Slug,
			},
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if req.AllowMerge {
		q.Set("allow_merge", "true")
	}
	if req.IsSilent {
		q.Set("is_silent", "true")
	}
	if len(req.Fields) > 0 {
		q.Set("fields", strings.Join(req.Fields, ","))
	}
	u.RawQuery = q.Encode()

	var page pageDTO
	if _, err := c.apiClient.DoPATCH(ctx, u.String(), body, &page, "UpdatePage"); err != nil {
		return nil, err
	}

	return &domain.WikiPageUpdateResponse{
		Page: *pageToWikiPage(&page),
	}, nil
}

// AppendPage appends content to an existing wiki page.
func (c *Client) AppendPage(
	ctx context.Context, req *domain.WikiPageAppendRequest,
) (*domain.WikiPageAppendResponse, error) {
	baseURL := fmt.Sprintf("%s/v1/pages/%s/append-content", c.baseURL, req.PageID)

	body := appendPageRequestDTO{ //nolint:exhaustruct // Body, Section, Anchor set conditionally
		Content: req.Content,
	}

	if req.Body != nil {
		body.Body = &BodyLocationRequest{
			Location: req.Body.Location,
		}
	}

	if req.Section != nil {
		body.Section = &sectionRequestDTO{
			ID:       apihelpers.StringID(req.Section.ID),
			Location: req.Section.Location,
		}
	}

	if req.Anchor != nil {
		body.Anchor = &anchorRequestDTO{
			Name:     req.Anchor.Name,
			Fallback: req.Anchor.Fallback,
			Regex:    req.Anchor.Regex,
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if req.IsSilent {
		q.Set("is_silent", "true")
	}
	if len(req.Fields) > 0 {
		q.Set("fields", strings.Join(req.Fields, ","))
	}
	u.RawQuery = q.Encode()

	var page pageDTO
	if _, err := c.apiClient.DoPOST(ctx, u.String(), body, &page, "AppendPage"); err != nil {
		return nil, err
	}

	return &domain.WikiPageAppendResponse{
		Page: *pageToWikiPage(&page),
	}, nil
}

// CreateGrid creates a new grid on a page.
func (c *Client) CreateGrid(
	ctx context.Context, req *domain.WikiGridCreateRequest,
) (*domain.WikiGridCreateResponse, error) {
	u := fmt.Sprintf("%s/v1/pages/%s/grids", c.baseURL, req.PageID)

	if len(req.Fields) > 0 {
		u += "?fields=" + strings.Join(req.Fields, ",")
	}

	columns := make([]columnCreateDTO, 0, len(req.Columns))
	for _, col := range req.Columns {
		columns = append(columns, columnCreateDTO{
			Slug:  col.Slug,
			Title: col.Title,
			Type:  col.Type,
		})
	}

	body := createGridRequestDTO{
		Title:   req.Title,
		Columns: columns,
	}

	var grid gridDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &grid, "CreateGrid"); err != nil {
		return nil, err
	}

	return &domain.WikiGridCreateResponse{
		Grid: *gridToWikiGrid(&grid),
	}, nil
}

// UpdateGridCells updates cells in a grid.
func (c *Client) UpdateGridCells(
	ctx context.Context, req *domain.WikiGridCellsUpdateRequest,
) (*domain.WikiGridCellsUpdateResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/cells", c.baseURL, req.GridID)

	cells := make([]cellUpdateDTO, 0, len(req.Cells))
	for _, cell := range req.Cells {
		cells = append(cells, cellUpdateDTO{
			RowID:      apihelpers.StringID(cell.RowID),
			ColumnSlug: cell.ColumnSlug,
			Value:      cell.Value,
		})
	}

	body := updateGridCellsRequestDTO{
		Cells:    cells,
		Revision: req.Revision,
	}

	var grid gridDTO
	if _, err := c.apiClient.DoPATCH(ctx, u, body, &grid, "UpdateGridCells"); err != nil {
		return nil, err
	}

	return &domain.WikiGridCellsUpdateResponse{
		Grid: *gridToWikiGrid(&grid),
	}, nil
}

// DeletePage deletes a wiki page by its ID.
func (c *Client) DeletePage(ctx context.Context, pageID string) (*domain.WikiPageDeleteResponse, error) {
	u := fmt.Sprintf("%s/v1/pages/%s", c.baseURL, pageID)

	var resp deletePageResponseDTO
	if _, err := c.apiClient.DoRequest(ctx, http.MethodDelete, u, nil, &resp, "DeletePage"); err != nil {
		return nil, err
	}

	return &domain.WikiPageDeleteResponse{
		RecoveryToken: resp.RecoveryToken,
	}, nil
}

// ClonePage clones a wiki page to a new location.
func (c *Client) ClonePage(
	ctx context.Context, req domain.WikiPageCloneRequest,
) (*domain.WikiCloneOperationResponse, error) {
	u := fmt.Sprintf("%s/v1/pages/%s/clone", c.baseURL, req.PageID)

	body := clonePageRequestDTO{
		Target:      req.Target,
		Title:       req.Title,
		SubscribeMe: req.SubscribeMe,
	}

	var resp cloneOperationResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "ClonePage"); err != nil {
		return nil, err
	}

	return &domain.WikiCloneOperationResponse{
		OperationID:   resp.Operation.ID.String(),
		OperationType: resp.Operation.Type,
		DryRun:        resp.DryRun,
		StatusURL:     resp.StatusURL,
	}, nil
}

// DeleteGrid deletes a wiki grid by its ID.
func (c *Client) DeleteGrid(ctx context.Context, gridID string) error {
	u := fmt.Sprintf("%s/v1/grids/%s", c.baseURL, gridID)

	if _, err := c.apiClient.DoDELETE(ctx, u, "DeleteGrid"); err != nil {
		return err
	}

	return nil
}

// CloneGrid clones a wiki grid to a new location.
func (c *Client) CloneGrid(
	ctx context.Context, req domain.WikiGridCloneRequest,
) (*domain.WikiCloneOperationResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/clone", c.baseURL, req.GridID)

	body := cloneGridRequestDTO{
		Target:   req.Target,
		Title:    req.Title,
		WithData: req.WithData,
	}

	var resp cloneOperationResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "CloneGrid"); err != nil {
		return nil, err
	}

	return &domain.WikiCloneOperationResponse{
		OperationID:   resp.Operation.ID.String(),
		OperationType: resp.Operation.Type,
		DryRun:        resp.DryRun,
		StatusURL:     resp.StatusURL,
	}, nil
}

// AddGridRows adds rows to a grid.
func (c *Client) AddGridRows(
	ctx context.Context, req domain.WikiGridRowsAddRequest,
) (*domain.WikiGridRowsAddResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/rows", c.baseURL, req.GridID)

	body := addGridRowsRequestDTO{
		Rows:       req.Rows,
		AfterRowID: apihelpers.StringID(req.AfterRowID),
		Position:   req.Position,
		Revision:   req.Revision,
	}

	var resp addGridRowsResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "AddGridRows"); err != nil {
		return nil, err
	}

	results := make([]domain.WikiGridRowResult, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = domain.WikiGridRowResult{
			ID:     r.ID.String(),
			Row:    r.Row,
			Color:  r.Color,
			Pinned: r.Pinned,
		}
	}

	return &domain.WikiGridRowsAddResponse{
		Revision: resp.Revision,
		Results:  results,
	}, nil
}

// DeleteGridRows deletes rows from a grid.
func (c *Client) DeleteGridRows(
	ctx context.Context, req domain.WikiGridRowsDeleteRequest,
) (*domain.WikiRevisionResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/rows", c.baseURL, req.GridID)

	body := deleteGridRowsRequestDTO{
		RowIDs:   apihelpers.StringsToStringIDs(req.RowIDs),
		Revision: req.Revision,
	}

	var resp revisionResponseDTO
	if _, err := c.apiClient.DoRequest(ctx, http.MethodDelete, u, body, &resp, "DeleteGridRows"); err != nil {
		return nil, err
	}

	return &domain.WikiRevisionResponse{
		Revision: resp.Revision,
	}, nil
}

// MoveGridRows moves rows within a grid.
func (c *Client) MoveGridRows(
	ctx context.Context, req domain.WikiGridRowsMoveRequest,
) (*domain.WikiRevisionResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/rows/move", c.baseURL, req.GridID)

	body := moveGridRowsRequestDTO{
		RowID:      apihelpers.StringID(req.RowID),
		AfterRowID: apihelpers.StringID(req.AfterRowID),
		Position:   req.Position,
		RowsCount:  req.RowsCount,
		Revision:   req.Revision,
	}

	var resp revisionResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "MoveGridRows"); err != nil {
		return nil, err
	}

	return &domain.WikiRevisionResponse{
		Revision: resp.Revision,
	}, nil
}

// AddGridColumns adds columns to a grid.
func (c *Client) AddGridColumns(
	ctx context.Context, req domain.WikiGridColumnsAddRequest,
) (*domain.WikiRevisionResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/columns", c.baseURL, req.GridID)

	columns := make([]newColumnSchemaReqDTO, len(req.Columns))
	for i, col := range req.Columns {
		columns[i] = newColumnSchemaReqDTO{
			Slug:          col.Slug,
			Title:         col.Title,
			Type:          col.Type,
			Required:      col.Required,
			Description:   col.Description,
			Color:         col.Color,
			Format:        col.Format,
			SelectOptions: col.SelectOptions,
			Multiple:      col.Multiple,
			MarkRows:      col.MarkRows,
			TicketField:   col.TicketField,
			Width:         col.Width,
			WidthUnits:    col.WidthUnits,
			Pinned:        col.Pinned,
		}
	}

	body := addGridColumnsRequestDTO{
		Columns:  columns,
		Position: req.Position,
		Revision: req.Revision,
	}

	var resp revisionResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "AddGridColumns"); err != nil {
		return nil, err
	}

	return &domain.WikiRevisionResponse{
		Revision: resp.Revision,
	}, nil
}

// DeleteGridColumns deletes columns from a grid.
func (c *Client) DeleteGridColumns(
	ctx context.Context, req domain.WikiGridColumnsDeleteRequest,
) (*domain.WikiRevisionResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/columns", c.baseURL, req.GridID)

	body := deleteGridColumnsRequestDTO{
		ColumnSlugs: req.ColumnSlugs,
		Revision:    req.Revision,
	}

	var resp revisionResponseDTO
	if _, err := c.apiClient.DoRequest(ctx, http.MethodDelete, u, body, &resp, "DeleteGridColumns"); err != nil {
		return nil, err
	}

	return &domain.WikiRevisionResponse{
		Revision: resp.Revision,
	}, nil
}

// MoveGridColumns moves columns within a grid.
func (c *Client) MoveGridColumns(
	ctx context.Context, req domain.WikiGridColumnsMoveRequest,
) (*domain.WikiRevisionResponse, error) {
	u := fmt.Sprintf("%s/v1/grids/%s/columns/move", c.baseURL, req.GridID)

	body := moveGridColumnsRequestDTO{
		ColumnSlug:   req.ColumnSlug,
		Position:     req.Position,
		ColumnsCount: req.ColumnsCount,
		Revision:     req.Revision,
	}

	var resp revisionResponseDTO
	if _, err := c.apiClient.DoPOST(ctx, u, body, &resp, "MoveGridColumns"); err != nil {
		return nil, err
	}

	return &domain.WikiRevisionResponse{
		Revision: resp.Revision,
	}, nil
}

// parseError converts an HTTP error response into a domain.UpstreamError.
func (c *Client) parseError(ctx context.Context, statusCode int, body []byte, operation string) error {
	var errResp errorResponseDTO
	var code, message string

	// Attempt to parse structured error
	if err := json.Unmarshal(body, &errResp); err == nil {
		code = errResp.ErrorCode
		message = errResp.DebugMessage
	}

	if message == "" {
		message = http.StatusText(statusCode)
	}

	err := domain.NewUpstreamError(
		domain.ServiceWiki,
		operation,
		statusCode,
		code,
		message,
		string(body),
	)

	return c.apiClient.ErrorLogWrapper(ctx, err)
}
