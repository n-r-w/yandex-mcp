package wiki

import "github.com/n-r-w/yandex-mcp/internal/adapters/apihelpers"

// pageDTO represents a Wiki page response.
type pageDTO struct {
	ID         apihelpers.StringID `json:"id"`
	PageType   string              `json:"page_type"`
	Slug       string              `json:"slug"`
	Title      string              `json:"title"`
	Content    string              `json:"content,omitempty"`
	Attributes *attributesDTO      `json:"attributes,omitempty"`
	Redirect   *redirectDTO        `json:"redirect,omitempty"`
}

// attributesDTO contains page metadata.
type attributesDTO struct {
	CommentsCount   int    `json:"comments_count"`
	CommentsEnabled bool   `json:"comments_enabled"`
	CreatedAt       string `json:"created_at"`
	IsReadonly      bool   `json:"is_readonly"`
	Lang            string `json:"lang"`
	ModifiedAt      string `json:"modified_at"`
	IsCollaborative bool   `json:"is_collaborative"`
	IsDraft         bool   `json:"is_draft"`
}

// redirectDTO represents page redirect info.
type redirectDTO struct {
	PageID         apihelpers.StringID `json:"page_id"`
	RedirectTarget *redirectTargetDTO  `json:"redirect_target,omitempty"`
}

// redirectTargetDTO represents the target page of a redirect.
type redirectTargetDTO struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	PageType string `json:"page_type"`
}

// resourceDTO represents a page resource (attachment, grid, or sharepoint resource).
type resourceDTO struct {
	Type string `json:"type"`
	Item any    `json:"item"`
}

// attachmentDTO represents a file attachment.
type attachmentDTO struct {
	ID          apihelpers.StringID `json:"id"`
	Name        string              `json:"name"`
	Size        int64               `json:"size"`
	Mimetype    string              `json:"mimetype"`
	DownloadURL string              `json:"download_url"`
	CreatedAt   string              `json:"created_at"`
	HasPreview  bool                `json:"has_preview"`
}

// pageGridSummaryDTO represents a grid summary in page resources.
type pageGridSummaryDTO struct {
	ID        apihelpers.StringID `json:"id"`
	Title     string              `json:"title"`
	CreatedAt string              `json:"created_at"`
}

// sharepointResourceDTO represents a SharePoint/MS365 document.
type sharepointResourceDTO struct {
	ID        apihelpers.StringID `json:"id"`
	Title     string              `json:"title"`
	Doctype   string              `json:"doctype"`
	CreatedAt string              `json:"created_at"`
}

// gridDTO represents a dynamic table (grid) with full details.
type gridDTO struct {
	ID             apihelpers.StringID `json:"id"`
	Title          string              `json:"title"`
	Structure      []columnDTO         `json:"structure"`
	Rows           []gridRowDTO        `json:"rows"`
	Revision       string              `json:"revision"`
	CreatedAt      string              `json:"created_at"`
	RichTextFormat string              `json:"rich_text_format"`
	Attributes     *attributesDTO      `json:"attributes,omitempty"`
}

// columnDTO represents a grid column definition.
type columnDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// gridRowDTO represents a row in a grid.
type gridRowDTO struct {
	ID    apihelpers.StringID `json:"id"`
	Cells map[string]any      `json:"cells"`
}

// resourcesPageDTO represents a paginated list of resources.
type resourcesPageDTO struct {
	Resources  []resourceDTO `json:"resources"`
	NextCursor string        `json:"next_cursor,omitempty"`
	PrevCursor string        `json:"prev_cursor,omitempty"`
}

// gridsPageDTO represents a paginated list of grids.
type gridsPageDTO struct {
	Grids      []pageGridSummaryDTO `json:"grids"`
	NextCursor string               `json:"next_cursor,omitempty"`
	PrevCursor string               `json:"prev_cursor,omitempty"`
}

// errorResponseDTO represents the Wiki API error format.
type errorResponseDTO struct {
	DebugMessage string `json:"debug_message"`
	ErrorCode    string `json:"error_code"`
}

// resourcesResponseDTO represents the raw resources list response.
type resourcesResponseDTO struct {
	Items      []resourceDTO `json:"items"`
	NextCursor string        `json:"next_cursor"`
	PrevCursor string        `json:"prev_cursor"`
}

// gridsResponseDTO represents the raw grids list response.
type gridsResponseDTO struct {
	Items      []pageGridSummaryDTO `json:"items"`
	NextCursor string               `json:"next_cursor"`
	PrevCursor string               `json:"prev_cursor"`
}
