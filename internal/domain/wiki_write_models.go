package domain

// WikiPageCreateRequest represents a request to create a new wiki page.
type WikiPageCreateRequest struct {
	Slug       string
	Title      string
	Content    string
	PageType   string
	IsSilent   bool
	Fields     []string
	CloudPage  *WikiCloudPageInput
	GridFormat string
}

// WikiPageUpdateRequest represents a request to update an existing wiki page.
type WikiPageUpdateRequest struct {
	PageID     int
	Title      string
	Content    string
	AllowMerge bool
	IsSilent   bool
	Fields     []string
	Redirect   *WikiRedirectInput
}

// WikiPageAppendRequest represents a request to append content to an existing wiki page.
type WikiPageAppendRequest struct {
	PageID   int
	Content  string
	IsSilent bool
	Fields   []string
	Body     *WikiBodyLocation
	Section  *WikiSectionLocation
	Anchor   *WikiAnchorLocation
}

// WikiGridCreateRequest represents a request to create a new grid.
type WikiGridCreateRequest struct {
	PageID  int64
	Title   string
	Columns []WikiColumnDefinition
	Fields  []string
}

// WikiColumnDefinition represents a column definition for grid creation.
type WikiColumnDefinition struct {
	Slug  string
	Title string
	Type  string
}

// WikiGridCellsUpdateRequest represents a request to update grid cells.
type WikiGridCellsUpdateRequest struct {
	GridID   string
	Cells    []WikiCellUpdate
	Revision string
}

// WikiCellUpdate represents a single cell update.
type WikiCellUpdate struct {
	RowID      int
	ColumnSlug string
	Value      string
}

// WikiPageCreateResponse represents the response from page creation.
type WikiPageCreateResponse struct {
	Page WikiPage
}

// WikiPageUpdateResponse represents the response from page update.
type WikiPageUpdateResponse struct {
	Page WikiPage
}

// WikiPageAppendResponse represents the response from page append.
type WikiPageAppendResponse struct {
	Page WikiPage
}

// WikiGridCreateResponse represents the response from grid creation.
type WikiGridCreateResponse struct {
	Grid WikiGrid
}

// WikiGridCellsUpdateResponse represents the response from grid cells update.
type WikiGridCellsUpdateResponse struct {
	Grid WikiGrid
}

// WikiCloudPageInput represents cloud page options for page creation.
type WikiCloudPageInput struct {
	Method  string
	Doctype string
}

// WikiRedirectInput represents redirect options for page update.
type WikiRedirectInput struct {
	PageID *int64
	Slug   *string
}

// WikiBodyLocation represents body location targeting for content append.
type WikiBodyLocation struct {
	Location string
}

// WikiSectionLocation represents section location targeting for content append.
type WikiSectionLocation struct {
	ID       int
	Location string
}

// WikiAnchorLocation represents anchor location targeting for content append.
type WikiAnchorLocation struct {
	Name     string
	Fallback bool
	Regex    bool
}
