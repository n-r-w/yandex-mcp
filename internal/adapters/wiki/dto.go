package wiki

// Page represents a Wiki page response.
type Page struct {
	ID         int64       `json:"id"`
	PageType   string      `json:"page_type"`
	Slug       string      `json:"slug"`
	Title      string      `json:"title"`
	Content    string      `json:"content,omitempty"`
	Attributes *Attributes `json:"attributes,omitempty"`
	Redirect   *Redirect   `json:"redirect,omitempty"`
}

// Attributes contains page metadata.
type Attributes struct {
	CommentsCount   int    `json:"comments_count"`
	CommentsEnabled bool   `json:"comments_enabled"`
	CreatedAt       string `json:"created_at"`
	IsReadonly      bool   `json:"is_readonly"`
	Lang            string `json:"lang"`
	ModifiedAt      string `json:"modified_at"`
	IsCollaborative bool   `json:"is_collaborative"`
	IsDraft         bool   `json:"is_draft"`
}

// Redirect represents page redirect info.
type Redirect struct {
	PageID int64  `json:"page_id"`
	Slug   string `json:"slug"`
}

// Resource represents a page resource (attachment, grid, or sharepoint resource).
type Resource struct {
	Type string `json:"type"`
	Item any    `json:"item"`
}

// Attachment represents a file attachment.
type Attachment struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Mimetype    string `json:"mimetype"`
	DownloadURL string `json:"download_url"`
	CreatedAt   string `json:"created_at"`
	HasPreview  bool   `json:"has_preview"`
}

// PageGridSummary represents a grid summary in page resources.
type PageGridSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// SharepointResource represents a SharePoint/MS365 document.
type SharepointResource struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Doctype   string `json:"doctype"`
	CreatedAt string `json:"created_at"`
}

// Grid represents a dynamic table (grid) with full details.
type Grid struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Structure      []Column    `json:"structure"`
	Rows           []GridRow   `json:"rows"`
	Revision       string      `json:"revision"`
	CreatedAt      string      `json:"created_at"`
	RichTextFormat string      `json:"rich_text_format"`
	Attributes     *Attributes `json:"attributes,omitempty"`
}

// Column represents a grid column definition.
type Column struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// GridRow represents a row in a grid.
type GridRow struct {
	ID    string         `json:"id"`
	Cells map[string]any `json:"cells"`
}

// ListResourcesOpts contains options for listing page resources.
type ListResourcesOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
	Query          string
	Types          string
}

// ListGridsOpts contains options for listing page grids.
type ListGridsOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
}

// GetGridOpts contains options for getting a grid.
type GetGridOpts struct {
	Fields   []string
	Filter   string
	OnlyCols string
	OnlyRows string
	Revision int
	Sort     string
}

// ResourcesPage represents a paginated list of resources.
type ResourcesPage struct {
	Resources  []Resource `json:"resources"`
	NextCursor string     `json:"next_cursor,omitempty"`
	PrevCursor string     `json:"prev_cursor,omitempty"`
}

// GridsPage represents a paginated list of grids.
type GridsPage struct {
	Grids      []PageGridSummary `json:"grids"`
	NextCursor string            `json:"next_cursor,omitempty"`
	PrevCursor string            `json:"prev_cursor,omitempty"`
}

// errorResponse represents the Wiki API error format.
type errorResponse struct {
	DebugMessage string `json:"debug_message"`
	ErrorCode    string `json:"error_code"`
}

// resourcesResponse represents the raw resources list response.
type resourcesResponse struct {
	Items      []Resource `json:"items"`
	NextCursor string     `json:"next_cursor"`
	PrevCursor string     `json:"prev_cursor"`
}

// gridsResponse represents the raw grids list response.
type gridsResponse struct {
	Items      []PageGridSummary `json:"items"`
	NextCursor string            `json:"next_cursor"`
	PrevCursor string            `json:"prev_cursor"`
}

// Write operation request DTOs.

// CreatePageRequest is the request body for page creation.
type CreatePageRequest struct {
	Slug       string            `json:"slug"`
	Title      string            `json:"title"`
	Content    string            `json:"body,omitempty"`
	PageType   string            `json:"page_type"`
	CloudPage  *CloudPageRequest `json:"cloud_page,omitempty"`
	GridFormat string            `json:"grid_format,omitempty"`
}

// UpdatePageRequest is the request body for page update.
type UpdatePageRequest struct {
	Title    string           `json:"title,omitempty"`
	Content  string           `json:"body,omitempty"`
	Redirect *RedirectRequest `json:"redirect,omitempty"`
}

// AppendPageRequest is the request body for appending content.
type AppendPageRequest struct {
	Content string               `json:"body"`
	Body    *BodyLocationRequest `json:"body_location,omitempty"`
	Section *SectionRequest      `json:"section,omitempty"`
	Anchor  *AnchorRequest       `json:"anchor,omitempty"`
}

// CreateGridRequest is the request body for grid creation.
type CreateGridRequest struct {
	Title   string         `json:"title"`
	Columns []ColumnCreate `json:"columns"`
}

// ColumnCreate represents a column definition for grid creation.
type ColumnCreate struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type,omitempty"`
}

// UpdateGridCellsRequest is the request body for updating grid cells.
type UpdateGridCellsRequest struct {
	Cells    []CellUpdate `json:"data"`
	Revision string       `json:"revision,omitempty"`
}

// CellUpdate represents a single cell update.
type CellUpdate struct {
	RowID      string `json:"row_id"`
	ColumnSlug string `json:"column_id"`
	Value      string `json:"value"`
}

// CloudPageRequest represents cloud page options for page creation.
type CloudPageRequest struct {
	Method  string `json:"method"`
	Doctype string `json:"doctype"`
}

// RedirectRequest represents redirect options for page update.
type RedirectRequest struct {
	Page *PageIdentityRequest `json:"page"`
}

// PageIdentityRequest identifies a page by ID or slug.
type PageIdentityRequest struct {
	ID   *int64  `json:"id,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

// BodyLocationRequest represents body location targeting for content append.
type BodyLocationRequest struct {
	Location string `json:"location"`
}

// SectionRequest represents section location targeting for content append.
type SectionRequest struct {
	ID       int    `json:"id"`
	Location string `json:"location"`
}

// AnchorRequest represents anchor location targeting for content append.
type AnchorRequest struct {
	Name     string `json:"name"`
	Fallback bool   `json:"fallback,omitempty"`
	Regex    bool   `json:"regex,omitempty"`
}
