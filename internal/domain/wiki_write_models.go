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
	PageID     string
	Title      string
	Content    string
	AllowMerge bool
	IsSilent   bool
	Fields     []string
	Redirect   *WikiRedirectInput
}

// WikiPageAppendRequest represents a request to append content to an existing wiki page.
type WikiPageAppendRequest struct {
	PageID   string
	Content  string
	IsSilent bool
	Fields   []string
	Body     *WikiBodyLocation
	Section  *WikiSectionLocation
	Anchor   *WikiAnchorLocation
}

// WikiGridCreateRequest represents a request to create a new grid.
type WikiGridCreateRequest struct {
	PageID  string
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
	RowID      string
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
	PageID *string
	Slug   *string
}

// WikiBodyLocation represents body location targeting for content append.
type WikiBodyLocation struct {
	Location string
}

// WikiSectionLocation represents section location targeting for content append.
type WikiSectionLocation struct {
	ID       string
	Location string
}

// WikiAnchorLocation represents anchor location targeting for content append.
type WikiAnchorLocation struct {
	Name     string
	Fallback bool
	Regex    bool
}

// WikiPageDeleteRequest represents a request to delete a wiki page.
type WikiPageDeleteRequest struct {
	PageID string
}

// WikiPageDeleteResponse represents the response after deleting a wiki page.
type WikiPageDeleteResponse struct {
	RecoveryToken string
}

// WikiPageCloneRequest represents a request to clone a wiki page.
type WikiPageCloneRequest struct {
	PageID      string
	Target      string
	Title       string
	SubscribeMe bool
}

// WikiCloneOperationResponse represents the response for async clone operations.
type WikiCloneOperationResponse struct {
	OperationID   string
	OperationType string
	DryRun        bool
	StatusURL     string
}

// WikiGridCloneRequest represents a request to clone a wiki grid.
type WikiGridCloneRequest struct {
	GridID   string
	Target   string
	Title    string
	WithData bool
}

// WikiGridRowsAddRequest represents a request to add rows to a grid.
type WikiGridRowsAddRequest struct {
	GridID     string
	Rows       []map[string]any
	AfterRowID string
	Position   *int
	Revision   string
}

// WikiGridRowsAddResponse represents the response from adding rows to a grid.
type WikiGridRowsAddResponse struct {
	Revision string
	Results  []WikiGridRowResult
}

// WikiGridRowResult represents a row result from grid row operations.
type WikiGridRowResult struct {
	ID     string
	Row    []any
	Color  string
	Pinned bool
}

// WikiGridRowsDeleteRequest represents a request to delete rows from a grid.
type WikiGridRowsDeleteRequest struct {
	GridID   string
	RowIDs   []string
	Revision string
}

// WikiRevisionResponse represents a response containing only revision info.
type WikiRevisionResponse struct {
	Revision string
}

// WikiGridRowsMoveRequest represents a request to move rows in a grid.
type WikiGridRowsMoveRequest struct {
	GridID     string
	RowID      string
	AfterRowID string
	Position   *int
	RowsCount  *int
	Revision   string
}

// WikiGridColumnsAddRequest represents a request to add columns to a grid.
type WikiGridColumnsAddRequest struct {
	GridID   string
	Columns  []WikiNewColumnDefinition
	Position *int
	Revision string
}

// WikiNewColumnDefinition represents a column definition for grid column creation.
type WikiNewColumnDefinition struct {
	Slug          string
	Title         string
	Type          string
	Required      bool
	Description   string
	Color         string
	Format        string
	SelectOptions []string
	Multiple      bool
	MarkRows      bool
	TicketField   string
	Width         *int
	WidthUnits    string
	Pinned        string
}

// WikiGridColumnsDeleteRequest represents a request to delete columns from a grid.
type WikiGridColumnsDeleteRequest struct {
	GridID      string
	ColumnSlugs []string
	Revision    string
}

// WikiGridColumnsMoveRequest represents a request to move columns in a grid.
type WikiGridColumnsMoveRequest struct {
	GridID       string
	ColumnSlug   string
	Position     int
	ColumnsCount *int
	Revision     string
}
