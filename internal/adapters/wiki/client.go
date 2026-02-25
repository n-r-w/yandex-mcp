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
}

// Compile-time check that Client implements the tools interface.
var _ wikitools.IWikiAdapter = (*Client)(nil)

// NewClient creates a new Wiki API client.
func NewClient(cfg *config.Config, tokenProvider apihelpers.ITokenProvider) *Client {
	client := &Client{
		apiClient: nil, // set below
	}

	client.apiClient = apihelpers.NewAPIClient(apihelpers.APIClientConfig{
		HTTPClient:          nil, // uses default
		TokenProvider:       tokenProvider,
		BaseURL:             strings.TrimSuffix(cfg.WikiBaseURL, "/"),
		OrgID:               cfg.CloudOrgID,
		ExtraHeaders:        nil,
		ServiceName:         string(domain.ServiceWiki),
		ParseError:          client.parseError,
		HTTPTimeout:         cfg.HTTPTimeout,
		RawResponseMaxBytes: cfg.AttachInlineMaxBytes,
	})

	return client
}

// GetPageBySlug retrieves a page by its slug.
func (c *Client) GetPageBySlug(
	ctx context.Context, slug string, opts domain.WikiGetPageOpts,
) (*domain.WikiPage, error) {
	u, err := url.Parse("/v1/pages")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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
	u, err := url.Parse("/v1/pages/" + url.PathEscape(id))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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
	u, err := url.Parse("/v1/pages/" + url.PathEscape(pageID) + "/resources")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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
	u, err := url.Parse("/v1/pages/" + url.PathEscape(pageID) + "/grids")
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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
	u, err := url.Parse("/v1/grids/" + url.PathEscape(gridID))
	if err != nil {
		return nil, c.apiClient.ErrorLogWrapper(ctx, fmt.Errorf("parse endpoint path: %w", err))
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
