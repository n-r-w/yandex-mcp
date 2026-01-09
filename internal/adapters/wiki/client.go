// Package wiki provides HTTP client for Yandex Wiki API.
package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	wikitools "github.com/n-r-w/yandex-mcp/internal/tools/wiki"
)

// Client implements IWikiClient for Yandex Wiki API.
type Client struct {
	httpClient    *http.Client
	tokenProvider ITokenProvider
	baseURL       string
	orgID         string
}

// Compile-time check that Client implements the tools interface.
var _ wikitools.IWikiAdapter = (*Client)(nil)

// NewClient creates a new Wiki API client.
func NewClient(cfg *config.Config, tokenProvider ITokenProvider) *Client {
	return &Client{
		httpClient: &http.Client{ //nolint:exhaustruct // optional fields use defaults
			Timeout: defaultTimeout,
		},
		tokenProvider: tokenProvider,
		baseURL:       strings.TrimSuffix(cfg.WikiBaseURL, "/"),
		orgID:         cfg.CloudOrgID,
	}
}

// GetPageBySlug retrieves a page by its slug.
func (c *Client) GetPageBySlug(
	ctx context.Context, slug string, opts domain.WikiGetPageOpts,
) (*domain.WikiPage, error) {
	u, err := url.Parse(c.baseURL + "/v1/pages")
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	q.Set("slug", slug)
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	if opts.RevisionID > 0 {
		q.Set("revision_id", strconv.Itoa(opts.RevisionID))
	}
	if opts.RaiseOnRedirect {
		q.Set("raise_on_redirect", "true")
	}
	u.RawQuery = q.Encode()

	var page Page
	if err := c.doGET(ctx, u.String(), &page, "GetPageBySlug"); err != nil {
		return nil, err
	}
	return pageToWikiPage(&page), nil
}

// GetPageByID retrieves a page by its ID.
func (c *Client) GetPageByID(ctx context.Context, id int64, opts domain.WikiGetPageOpts) (*domain.WikiPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%d", c.baseURL, id))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	if opts.RevisionID > 0 {
		q.Set("revision_id", strconv.Itoa(opts.RevisionID))
	}
	if opts.RaiseOnRedirect {
		q.Set("raise_on_redirect", "true")
	}
	u.RawQuery = q.Encode()

	var page Page
	if err := c.doGET(ctx, u.String(), &page, "GetPageByID"); err != nil {
		return nil, err
	}
	return pageToWikiPage(&page), nil
}

// ListPageResources lists resources (attachments, grids) for a page.
func (c *Client) ListPageResources(
	ctx context.Context,
	pageID int64,
	opts domain.WikiListResourcesOpts,
) (*domain.WikiResourcesPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%d/resources", c.baseURL, pageID))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	if opts.PageIDLegacy > 0 {
		q.Set("page_id", strconv.Itoa(opts.PageIDLegacy))
	}
	u.RawQuery = q.Encode()

	var resp resourcesResponse
	if err := c.doGET(ctx, u.String(), &resp, "ListPageResources"); err != nil {
		return nil, err
	}

	rp := &ResourcesPage{
		Resources:  resp.Items,
		NextCursor: resp.NextCursor,
		PrevCursor: resp.PrevCursor,
	}
	return resourcesPageToWikiResourcesPage(rp)
}

// ListPageGrids lists dynamic tables (grids) for a page.
func (c *Client) ListPageGrids(
	ctx context.Context,
	pageID int64,
	opts domain.WikiListGridsOpts,
) (*domain.WikiGridsPage, error) {
	u, err := url.Parse(fmt.Sprintf("%s/v1/pages/%d/grids", c.baseURL, pageID))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	if opts.PageIDLegacy > 0 {
		q.Set("page_id", strconv.Itoa(opts.PageIDLegacy))
	}
	u.RawQuery = q.Encode()

	var resp gridsResponse
	if err := c.doGET(ctx, u.String(), &resp, "ListPageGrids"); err != nil {
		return nil, err
	}

	gp := &GridsPage{
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
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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
	if opts.Revision > 0 {
		q.Set("revision", strconv.Itoa(opts.Revision))
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	u.RawQuery = q.Encode()

	var grid Grid
	if err := c.doGET(ctx, u.String(), &grid, "GetGridByID"); err != nil {
		return nil, err
	}
	return gridToWikiGrid(&grid), nil
}

// CreatePage creates a new wiki page.
func (c *Client) CreatePage(
	ctx context.Context, req *domain.WikiPageCreateRequest,
) (*domain.WikiPageCreateResponse, error) {
	baseURL := c.baseURL + "/v1/pages"

	body := CreatePageRequest{ //nolint:exhaustruct // CloudPage set conditionally
		Slug:       req.Slug,
		Title:      req.Title,
		Content:    req.Content,
		PageType:   req.PageType,
		GridFormat: req.GridFormat,
	}

	if req.CloudPage != nil {
		body.CloudPage = &CloudPageRequest{
			Method:  req.CloudPage.Method,
			Doctype: req.CloudPage.Doctype,
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if req.IsSilent {
		q.Set("is_silent", "true")
	}
	if len(req.Fields) > 0 {
		q.Set("fields", strings.Join(req.Fields, ","))
	}
	u.RawQuery = q.Encode()

	var page Page
	if err := c.doPOST(ctx, u.String(), body, &page, "CreatePage"); err != nil {
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
	baseURL := fmt.Sprintf("%s/v1/pages/%d", c.baseURL, req.PageID)

	body := UpdatePageRequest{ //nolint:exhaustruct // Redirect set conditionally
		Title:   req.Title,
		Content: req.Content,
	}

	if req.Redirect != nil {
		body.Redirect = &RedirectRequest{
			Page: &PageIdentityRequest{
				ID:   req.Redirect.PageID,
				Slug: req.Redirect.Slug,
			},
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
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

	var page Page
	if err := c.doPATCH(ctx, u.String(), body, &page, "UpdatePage"); err != nil {
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
	baseURL := fmt.Sprintf("%s/v1/pages/%d/append-content", c.baseURL, req.PageID)

	body := AppendPageRequest{ //nolint:exhaustruct // Body, Section, Anchor set conditionally
		Content: req.Content,
	}

	if req.Body != nil {
		body.Body = &BodyLocationRequest{
			Location: req.Body.Location,
		}
	}

	if req.Section != nil {
		body.Section = &SectionRequest{
			ID:       req.Section.ID,
			Location: req.Section.Location,
		}
	}

	if req.Anchor != nil {
		body.Anchor = &AnchorRequest{
			Name:     req.Anchor.Name,
			Fallback: req.Anchor.Fallback,
			Regex:    req.Anchor.Regex,
		}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("parse base URL: %w", err))
	}

	q := u.Query()
	if req.IsSilent {
		q.Set("is_silent", "true")
	}
	if len(req.Fields) > 0 {
		q.Set("fields", strings.Join(req.Fields, ","))
	}
	u.RawQuery = q.Encode()

	var page Page
	if err := c.doPOST(ctx, u.String(), body, &page, "AppendPage"); err != nil {
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
	u := fmt.Sprintf("%s/v1/pages/%d/grids", c.baseURL, req.PageID)

	if len(req.Fields) > 0 {
		u += "?fields=" + strings.Join(req.Fields, ",")
	}

	columns := make([]ColumnCreate, 0, len(req.Columns))
	for _, col := range req.Columns {
		columns = append(columns, ColumnCreate{
			Slug:  col.Slug,
			Title: col.Title,
			Type:  col.Type,
		})
	}

	body := CreateGridRequest{
		Title:   req.Title,
		Columns: columns,
	}

	var grid Grid
	if err := c.doPOST(ctx, u, body, &grid, "CreateGrid"); err != nil {
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

	cells := make([]CellUpdate, 0, len(req.Cells))
	for _, cell := range req.Cells {
		cells = append(cells, CellUpdate{
			RowID:      strconv.Itoa(cell.RowID),
			ColumnSlug: cell.ColumnSlug,
			Value:      cell.Value,
		})
	}

	body := UpdateGridCellsRequest{
		Cells:    cells,
		Revision: req.Revision,
	}

	var grid Grid
	if err := c.doPATCH(ctx, u, body, &grid, "UpdateGridCells"); err != nil {
		return nil, err
	}

	return &domain.WikiGridCellsUpdateResponse{
		Grid: *gridToWikiGrid(&grid),
	}, nil
}

// doPOST performs an HTTP POST request with retry on 401.
func (c *Client) doPOST(ctx context.Context, urlStr string, body any, result any, operation string) error {
	resp, err := c.executeRequest(ctx, http.MethodPost, urlStr, body, false)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		resp, err = c.executeRequest(ctx, http.MethodPost, urlStr, body, true)
		if err != nil {
			return err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return c.handleResponse(resp, result, operation)
}

// doPATCH performs an HTTP PATCH request with retry on 401.
func (c *Client) doPATCH(ctx context.Context, urlStr string, body any, result any, operation string) error {
	resp, err := c.executeRequest(ctx, http.MethodPatch, urlStr, body, false)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		resp, err = c.executeRequest(ctx, http.MethodPatch, urlStr, body, true)
		if err != nil {
			return err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return c.handleResponse(resp, result, operation)
}

// executeRequest performs a request with token injection and body encoding.
func (c *Client) executeRequest(
	ctx context.Context, method, urlStr string, body any, forceRefresh bool,
) (*http.Response, error) {
	var token string
	var err error

	if forceRefresh {
		token, err = c.tokenProvider.ForceRefresh(ctx)
	} else {
		token, err = c.tokenProvider.Token(ctx)
	}
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("get token: %w", err))
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("encode body: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("create request: %w", err))
	}

	req.Header.Set(headerAuthorization, "Bearer "+token)
	req.Header.Set(headerCloudOrgID, c.orgID)
	req.Header.Set(headerContentType, contentTypeJSON)

	return c.httpClient.Do(req)
}

// doGET executes a GET request with token injection and 401 retry logic.
func (c *Client) doGET(ctx context.Context, urlStr string, result any, operation string) error {
	resp, err := c.executeGET(ctx, urlStr, false)
	if err != nil {
		return err
	}

	// On 401, force token refresh and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		resp, err = c.executeGET(ctx, urlStr, true)
		if err != nil {
			return err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	return c.handleResponse(resp, result, operation)
}

// executeGET performs a single GET request with token injection.
func (c *Client) executeGET(ctx context.Context, urlStr string, forceRefresh bool) (*http.Response, error) {
	var token string
	var err error

	if forceRefresh {
		token, err = c.tokenProvider.ForceRefresh(ctx)
	} else {
		token, err = c.tokenProvider.Token(ctx)
	}
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("get token: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, errorLogWrapper(ctx, fmt.Errorf("create request: %w", err))
	}

	req.Header.Set(headerAuthorization, "Bearer "+token)
	req.Header.Set(headerCloudOrgID, c.orgID)
	req.Header.Set(headerContentType, contentTypeJSON)

	return c.httpClient.Do(req)
}

// handleResponse processes the HTTP response and decodes the result.
func (c *Client) handleResponse(resp *http.Response, result any, operation string) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorLogWrapper(context.Background(), fmt.Errorf("read response body: %w", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseError(resp.StatusCode, bodyBytes, operation)
	}

	if result != nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return errorLogWrapper(context.Background(), fmt.Errorf("decode response: %w", err))
		}
	}

	return nil
}

// parseError converts an HTTP error response into a domain.UpstreamError.
func (c *Client) parseError(statusCode int, body []byte, operation string) error {
	var errResp errorResponse
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

	return errorLogWrapper(context.Background(), err)
}

func errorLogWrapper(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	slog.ErrorContext(ctx, "wiki adapter error", "error", err)
	return err
}
