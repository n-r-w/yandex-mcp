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
	PageID apihelpers.StringID `json:"page_id"`
	Slug   string              `json:"slug"`
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

// Write operation request DTOs.

// createPageRequestDTO is the request body for page creation.
type createPageRequestDTO struct {
	Slug       string               `json:"slug"`
	Title      string               `json:"title"`
	Content    string               `json:"body,omitempty"`
	PageType   string               `json:"page_type"`
	CloudPage  *cloudPageRequestDTO `json:"cloud_page,omitempty"`
	GridFormat string               `json:"grid_format,omitempty"`
}

// updatePageRequestDTO is the request body for page update.
type updatePageRequestDTO struct {
	Title    string              `json:"title,omitempty"`
	Content  string              `json:"body,omitempty"`
	Redirect *redirectRequestDTO `json:"redirect,omitempty"`
}

// appendPageRequestDTO is the request body for appending content.
type appendPageRequestDTO struct {
	Content string               `json:"body"`
	Body    *BodyLocationRequest `json:"body_location,omitempty"`
	Section *sectionRequestDTO   `json:"section,omitempty"`
	Anchor  *anchorRequestDTO    `json:"anchor,omitempty"`
}

// createGridRequestDTO is the request body for grid creation.
type createGridRequestDTO struct {
	Title   string            `json:"title"`
	Columns []columnCreateDTO `json:"columns"`
}

// columnCreateDTO represents a column definition for grid creation.
type columnCreateDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type,omitempty"`
}

// updateGridCellsRequestDTO is the request body for updating grid cells.
type updateGridCellsRequestDTO struct {
	Cells    []cellUpdateDTO `json:"data"`
	Revision string          `json:"revision,omitempty"`
}

// cellUpdateDTO represents a single cell update.
type cellUpdateDTO struct {
	RowID      apihelpers.StringID `json:"row_id"`
	ColumnSlug string              `json:"column_id"`
	Value      string              `json:"value"`
}

// cloudPageRequestDTO represents cloud page options for page creation.
type cloudPageRequestDTO struct {
	Method  string `json:"method"`
	Doctype string `json:"doctype"`
}

// redirectRequestDTO represents redirect options for page update.
type redirectRequestDTO struct {
	Page *pageIdentityRequestDTO `json:"page"`
}

// pageIdentityRequestDTO identifies a page by ID or slug.
type pageIdentityRequestDTO struct {
	ID   *apihelpers.StringID `json:"id,omitempty"`
	Slug *string              `json:"slug,omitempty"`
}

// BodyLocationRequest represents body location targeting for content append.
type BodyLocationRequest struct {
	Location string `json:"location"`
}

// sectionRequestDTO represents section location targeting for content append.
type sectionRequestDTO struct {
	ID       apihelpers.StringID `json:"id"`
	Location string              `json:"location"`
}

// anchorRequestDTO represents anchor location targeting for content append.
type anchorRequestDTO struct {
	Name     string `json:"name"`
	Fallback bool   `json:"fallback,omitempty"`
	Regex    bool   `json:"regex,omitempty"`
}

// deletePageResponseDTO is the response from the delete page endpoint.
type deletePageResponseDTO struct {
	RecoveryToken string `json:"recovery_token"`
}

// clonePageRequestDTO is the request body for cloning a wiki page.
type clonePageRequestDTO struct {
	Target      string `json:"target"`
	Title       string `json:"title,omitempty"`
	SubscribeMe bool   `json:"subscribe_me,omitempty"`
}

// cloneGridRequestDTO is the request body for cloning a wiki grid.
type cloneGridRequestDTO struct {
	Target   string `json:"target"`
	Title    string `json:"title,omitempty"`
	WithData bool   `json:"with_data,omitempty"`
}

// cloneOperationResponseDTO is the response from clone operations (page or grid).
type cloneOperationResponseDTO struct {
	Operation operationIdentityDTO `json:"operation"`
	DryRun    bool                 `json:"dry_run"`
	StatusURL string               `json:"status_url"`
}

// operationIdentityDTO contains the identity of an async operation.
type operationIdentityDTO struct {
	ID   apihelpers.StringID `json:"id"`
	Type string              `json:"type"`
}

// addGridRowsRequestDTO is the request body for adding rows to a grid.
type addGridRowsRequestDTO struct {
	Rows       []map[string]any    `json:"rows"`
	AfterRowID apihelpers.StringID `json:"after_row_id,omitempty"`
	Position   *int                `json:"position,omitempty"`
	Revision   string              `json:"revision,omitempty"`
}

// addGridRowsResponseDTO is the response from adding rows to a grid.
type addGridRowsResponseDTO struct {
	Revision string                 `json:"revision"`
	Results  []gridRowSchemaRespDTO `json:"results"`
}

// gridRowSchemaRespDTO represents a row in grid row operation responses.
type gridRowSchemaRespDTO struct {
	ID     apihelpers.StringID `json:"id"`
	Row    []any               `json:"row"`
	Color  string              `json:"color,omitempty"`
	Pinned bool                `json:"pinned,omitempty"`
}

// deleteGridRowsRequestDTO is the request body for deleting rows from a grid.
type deleteGridRowsRequestDTO struct {
	RowIDs   []apihelpers.StringID `json:"row_ids"`
	Revision string                `json:"revision,omitempty"`
}

// revisionResponseDTO is a response containing only revision info.
type revisionResponseDTO struct {
	Revision string `json:"revision"`
}

// moveGridRowsRequestDTO is the request body for moving rows in a grid.
type moveGridRowsRequestDTO struct {
	RowID      apihelpers.StringID `json:"row_id"`
	AfterRowID apihelpers.StringID `json:"after_row_id,omitempty"`
	Position   *int                `json:"position,omitempty"`
	RowsCount  *int                `json:"rows_count,omitempty"`
	Revision   string              `json:"revision,omitempty"`
}

// addGridColumnsRequestDTO is the request body for adding columns to a grid.
type addGridColumnsRequestDTO struct {
	Columns  []newColumnSchemaReqDTO `json:"columns"`
	Position *int                    `json:"position,omitempty"`
	Revision string                  `json:"revision,omitempty"`
}

// newColumnSchemaReqDTO represents a column definition for column creation.
type newColumnSchemaReqDTO struct {
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Type          string   `json:"type"`
	Required      bool     `json:"required"`
	Description   string   `json:"description,omitempty"`
	Color         string   `json:"color,omitempty"`
	Format        string   `json:"format,omitempty"`
	SelectOptions []string `json:"select_options,omitempty"`
	Multiple      bool     `json:"multiple,omitempty"`
	MarkRows      bool     `json:"mark_rows,omitempty"`
	TicketField   string   `json:"ticket_field,omitempty"`
	Width         *int     `json:"width,omitempty"`
	WidthUnits    string   `json:"width_units,omitempty"`
	Pinned        string   `json:"pinned,omitempty"`
}

// deleteGridColumnsRequestDTO is the request body for deleting columns from a grid.
type deleteGridColumnsRequestDTO struct {
	ColumnSlugs []string `json:"column_slugs"`
	Revision    string   `json:"revision,omitempty"`
}

// moveGridColumnsRequestDTO is the request body for moving columns in a grid.
type moveGridColumnsRequestDTO struct {
	ColumnSlug   string `json:"column_slug"`
	Position     int    `json:"position"`
	ColumnsCount *int   `json:"columns_count,omitempty"`
	Revision     string `json:"revision,omitempty"`
}
