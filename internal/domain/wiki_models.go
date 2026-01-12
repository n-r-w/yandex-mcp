package domain

// WikiPage represents a Yandex Wiki page entity.
type WikiPage struct {
	ID         string
	PageType   string
	Slug       string
	Title      string
	Content    string
	Attributes *WikiAttributes
	Redirect   *WikiRedirect
}

// WikiAttributes represents metadata attributes for a Wiki page.
type WikiAttributes struct {
	CommentsCount   int
	CommentsEnabled bool
	CreatedAt       string
	IsReadonly      bool
	Lang            string
	ModifiedAt      string
	IsCollaborative bool
	IsDraft         bool
}

// WikiRedirect represents redirect information for a Wiki page.
type WikiRedirect struct {
	PageID         string
	RedirectTarget *WikiRedirectTarget
}

// WikiRedirectTarget represents the target page of a redirect.
type WikiRedirectTarget struct {
	ID       string
	Slug     string
	Title    string
	PageType string
}

// WikiResourcesPage represents a paginated list of Wiki resources.
type WikiResourcesPage struct {
	Resources  []WikiResource
	NextCursor string
	PrevCursor string
}

// WikiResource represents a single resource attached to a Wiki page.
type WikiResource struct {
	Type       string
	Attachment *WikiAttachment
	Sharepoint *WikiSharepointResource
	Grid       *WikiGridResource
}

// WikiAttachment represents a file attachment resource.
type WikiAttachment struct {
	ID          string
	Name        string
	Size        int64
	MIMEType    string
	DownloadURL string
	CreatedAt   string
	HasPreview  bool
}

// WikiSharepointResource represents a SharePoint resource linked to a Wiki page.
type WikiSharepointResource struct {
	ID        string
	Title     string
	Doctype   string
	CreatedAt string
}

// WikiGridResource represents a grid resource linked to a Wiki page.
type WikiGridResource struct {
	ID        string
	Title     string
	CreatedAt string
}

// WikiGridsPage represents a paginated list of Wiki grids.
type WikiGridsPage struct {
	Grids      []WikiGridSummary
	NextCursor string
	PrevCursor string
}

// WikiGridSummary represents summary information for a Wiki grid.
type WikiGridSummary struct {
	ID        string
	Title     string
	CreatedAt string
}

// WikiGrid represents a full Wiki grid with structure and data.
type WikiGrid struct {
	ID             string
	Title          string
	Structure      []WikiColumn
	Rows           []WikiGridRow
	Revision       string
	CreatedAt      string
	RichTextFormat string
	Attributes     *WikiAttributes
}

// WikiColumn represents a column definition in a Wiki grid.
type WikiColumn struct {
	Slug  string
	Title string
	Type  string
}

// WikiGridCell represents a single cell value in a Wiki grid row.
// Value holds the string representation of the cell content.
type WikiGridCell struct {
	Value string
}

// WikiGridRow represents a single row in a Wiki grid.
type WikiGridRow struct {
	ID    string
	Cells map[string]WikiGridCell
}
