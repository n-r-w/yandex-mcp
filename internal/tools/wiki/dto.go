//nolint:lll
package wiki

// Input DTOs for wiki tools.

// getPageBySlugInputDTO is the input for wiki_page_get tool.
type getPageBySlugInputDTO struct {
	Slug            string   `json:"slug" jsonschema:"Page slug (URL path),required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      string   `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID (string)"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// getPageByIDInputDTO is the input for wiki_page_get_by_id tool.
type getPageByIDInputDTO struct {
	PageID          string   `json:"page_id" jsonschema:"Page ID (string),required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      string   `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID (string)"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// listResourcesInputDTO is the input for wiki_page_resources_list tool.
type listResourcesInputDTO struct {
	PageID         string `json:"page_id" jsonschema:"Page ID (string) to list resources for,required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page (default: 25, min: 1, max: 50)"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: name_title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc (default), desc"`
	Q              string `json:"q,omitempty" jsonschema:"Filter resources by title (max 255 chars)"`
	Types          string `json:"types,omitempty" jsonschema:"Resource types filter. Possible values: attachment, sharepoint_resource, grid. Can be comma-separated for multiple types"`
	PageIDLegacy   string `json:"page_id_legacy,omitempty" jsonschema:"Legacy page ID (string) for backward-compatibility pagination (default: 1)"`
}

// listGridsInputDTO is the input for wiki_page_grids_list tool.
type listGridsInputDTO struct {
	PageID         string `json:"page_id" jsonschema:"Page ID (string) to list grids for,required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page (default: 25, min: 1, max: 50)"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc (default), desc"`
	PageIDLegacy   string `json:"page_id_legacy,omitempty" jsonschema:"Legacy page ID (string) for backward-compatibility pagination (default: 1)"`
}

// getGridInputDTO is the input for wiki_grid_get tool.
type getGridInputDTO struct {
	GridID   string   `json:"grid_id" jsonschema:"Grid ID (UUID string),required"`
	Fields   []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, user_permissions"`
	Filter   string   `json:"filter,omitempty" jsonschema:"Row filter expression to filter grid rows. Syntax: [column_slug] operator value. Operators: ~ (contains), <, >, <=, >=, =, !. Logical: AND, OR, (). Example: [slug] ~ wiki AND [slug2]<32"`
	OnlyCols string   `json:"only_cols,omitempty" jsonschema:"Return only specified columns (comma-separated column slugs)"`
	OnlyRows string   `json:"only_rows,omitempty" jsonschema:"Return only specified rows (comma-separated row IDs)"`
	Revision string   `json:"revision,omitempty" jsonschema:"Grid revision number for optimistic locking and historical versions"`
	Sort     string   `json:"sort,omitempty" jsonschema:"Sort expression to order rows by column"`
}

// Write tool input DTOs.

// createPageInputDTO is the input for wiki_page_create tool.
type createPageInputDTO struct {
	Slug       string             `json:"slug" jsonschema:"Page slug (URL path),required"`
	Title      string             `json:"title" jsonschema:"Page title,required"`
	PageType   string             `json:"page_type" jsonschema:"Page type. Possible values: page, grid, cloud_page, wysiwyg, template,required"`
	Content    string             `json:"content,omitempty" jsonschema:"Page content in wikitext format"`
	IsSilent   bool               `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields     []string           `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	CloudPage  *cloudPageInputDTO `json:"cloud_page,omitempty" jsonschema:"Cloud page options for cloud_page type"`
	GridFormat string             `json:"grid_format,omitempty" jsonschema:"Text format for grid columns. Possible values: yfm, wom, plain"`
}

// updatePageInputDTO is the input for wiki_page_update tool.
type updatePageInputDTO struct {
	PageID     string            `json:"page_id" jsonschema:"Page ID (string),required"`
	Title      string            `json:"title,omitempty" jsonschema:"Page title"`
	Content    string            `json:"content,omitempty" jsonschema:"Page content in wikitext format"`
	AllowMerge bool              `json:"allow_merge,omitempty" jsonschema:"Enable 3-way merge for concurrent edits"`
	IsSilent   bool              `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields     []string          `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	Redirect   *redirectInputDTO `json:"redirect,omitempty" jsonschema:"Set or remove page redirect"`
}

// appendPageInputDTO is the input for wiki_page_append_content tool.
type appendPageInputDTO struct {
	PageID   string              `json:"page_id" jsonschema:"Page ID (string),required"`
	Content  string              `json:"content" jsonschema:"Content to append in wikitext format,required"`
	IsSilent bool                `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields   []string            `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	Body     *bodyLocationDTO    `json:"body,omitempty" jsonschema:"Append to top or bottom of page body"`
	Section  *sectionLocationDTO `json:"section,omitempty" jsonschema:"Append to top or bottom of specific section"`
	Anchor   *anchorLocationDTO  `json:"anchor,omitempty" jsonschema:"Append relative to named anchor"`
}

// pageInputDTO represents page identification (by ID or slug).
type pageInputDTO struct {
	ID   string `json:"id,omitempty" jsonschema:"Page ID (string)"`
	Slug string `json:"slug,omitempty" jsonschema:"Page slug (URL path)"`
}

// createGridInputDTO is the input for wiki_grid_create tool.
type createGridInputDTO struct {
	Page    pageInputDTO           `json:"page" jsonschema:"Page where grid will be created (provide id or slug),required"`
	Title   string                 `json:"title" jsonschema:"Grid title,required"`
	Columns []columnInputCreateDTO `json:"columns" jsonschema:"Grid columns definition,required"`
	Fields  string                 `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, user_permissions"`
}

// columnInputCreateDTO defines a column for grid creation.
type columnInputCreateDTO struct {
	Slug  string `json:"slug" jsonschema:"Column slug (ID string),required"`
	Title string `json:"title" jsonschema:"Column title,required"`
	Type  string `json:"type" jsonschema:"Column type. Possible values: string, number, date, select, staff, checkbox, ticket, ticket_field"`
}

// updateGridCellsInputDTO is the input for wiki_grid_update_cells tool.
type updateGridCellsInputDTO struct {
	GridID   string               `json:"grid_id" jsonschema:"Grid ID (UUID string),required"`
	Cells    []cellUpdateInputDTO `json:"cells" jsonschema:"Array of cell updates,required"`
	Revision string               `json:"revision,omitempty" jsonschema:"Grid revision for optimistic locking"`
}

// cellUpdateInputDTO represents a single cell update.
type cellUpdateInputDTO struct {
	RowID      string `json:"row_id" jsonschema:"Row ID (string),required"`
	ColumnSlug string `json:"column_slug" jsonschema:"Column slug,required"`
	Value      string `json:"value" jsonschema:"Cell value,required"`
}

// cloudPageInputDTO represents cloud page options for page creation.
type cloudPageInputDTO struct {
	Method  string `json:"method" jsonschema:"Method for creating cloud page. Possible values: empty_doc, from_url, upload_doc, finalize_upload, upload_onprem,required"`
	Doctype string `json:"doctype" jsonschema:"Document type. Possible values: docx, pptx, xlsx,required"`
}

// redirectInputDTO represents redirect options for page update.
type redirectInputDTO struct {
	PageID *string `json:"page_id,omitempty" jsonschema:"Target page ID (string) for redirect. Set to null to remove redirect"`
	Slug   *string `json:"slug,omitempty" jsonschema:"Target page slug for redirect. If both page_id and slug provided, page_id is used"`
}

// bodyLocationDTO represents body location targeting for content append.
type bodyLocationDTO struct {
	Location string `json:"location" jsonschema:"Append location within body. Possible values: top, bottom,required"`
}

// sectionLocationDTO represents section location targeting for content append.
type sectionLocationDTO struct {
	ID       string `json:"id" jsonschema:"Section ID (string),required"`
	Location string `json:"location" jsonschema:"Append location within section. Possible values: top, bottom,required"`
}

// anchorLocationDTO represents anchor location targeting for content append.
type anchorLocationDTO struct {
	Name     string `json:"name" jsonschema:"Anchor name,required"`
	Fallback bool   `json:"fallback,omitempty" jsonschema:"Fall back to default behavior if anchor not found"`
	Regex    bool   `json:"regex,omitempty" jsonschema:"Treat anchor name as regular expression"`
}

// Output DTOs for wiki tools.

// pageOutputDTO is the output for page retrieval tools.
type pageOutputDTO struct {
	ID         string               `json:"id"`
	PageType   string               `json:"page_type"`
	Slug       string               `json:"slug"`
	Title      string               `json:"title"`
	Content    string               `json:"content,omitempty"`
	Attributes *attributesOutputDTO `json:"attributes,omitempty"`
	Redirect   *redirectOutputDTO   `json:"redirect,omitempty"`
}

// attributesOutputDTO contains page attributes.
type attributesOutputDTO struct {
	CommentsCount   int    `json:"comments_count"`
	CommentsEnabled bool   `json:"comments_enabled"`
	CreatedAt       string `json:"created_at"`
	IsReadonly      bool   `json:"is_readonly"`
	Lang            string `json:"lang"`
	ModifiedAt      string `json:"modified_at"`
	IsCollaborative bool   `json:"is_collaborative"`
	IsDraft         bool   `json:"is_draft"`
}

// redirectOutputDTO contains page redirect info.
type redirectOutputDTO struct {
	PageID string `json:"page_id"`
	Slug   string `json:"slug"`
}

// resourcesListOutputDTO is the output for wiki_page_resources_list tool.
type resourcesListOutputDTO struct {
	Resources  []resourceOutputDTO `json:"resources"`
	NextCursor string              `json:"next_cursor,omitempty"`
	PrevCursor string              `json:"prev_cursor,omitempty"`
}

// resourceOutputDTO represents a page resource.
type resourceOutputDTO struct {
	Type string `json:"type"`
	Item any    `json:"item"`
}

// attachmentOutputDTO represents a file attachment for serialization.
type attachmentOutputDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Mimetype    string `json:"mimetype"`
	DownloadURL string `json:"download_url"`
	CreatedAt   string `json:"created_at"`
	HasPreview  bool   `json:"has_preview"`
}

// sharepointResourceOutputDTO represents a SharePoint/MS365 document for serialization.
type sharepointResourceOutputDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Doctype   string `json:"doctype"`
	CreatedAt string `json:"created_at"`
}

// gridResourceOutputDTO represents a grid resource item for serialization.
type gridResourceOutputDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// gridsListOutputDTO is the output for wiki_page_grids_list tool.
type gridsListOutputDTO struct {
	Grids      []gridSummaryOutputDTO `json:"grids"`
	NextCursor string                 `json:"next_cursor,omitempty"`
	PrevCursor string                 `json:"prev_cursor,omitempty"`
}

// gridSummaryOutputDTO represents a grid summary.
type gridSummaryOutputDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// gridOutputDTO is the output for wiki_grid_get tool.
type gridOutputDTO struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Structure   []columnOutputDTO    `json:"structure,omitempty"`
	Rows        []gridRowOutputDTO   `json:"rows,omitempty"`
	Revision    string               `json:"revision"`
	CreatedAt   string               `json:"created_at"`
	RichTextFmt string               `json:"rich_text_format"`
	Attributes  *attributesOutputDTO `json:"attributes,omitempty"`
}

// columnOutputDTO represents a grid column.
type columnOutputDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// gridRowOutputDTO represents a grid row.
type gridRowOutputDTO struct {
	ID    string         `json:"id"`
	Cells map[string]any `json:"cells"`
}

// deletePageInputDTO is the input for wiki_page_delete.
type deletePageInputDTO struct {
	PageID string `json:"page_id" jsonschema:"Page ID (string) to delete,required"`
}

// deletePageOutputDTO is the output for wiki_page_delete.
type deletePageOutputDTO struct {
	RecoveryToken string `json:"recovery_token"`
}

// clonePageInputDTO is the input for wiki_page_clone.
type clonePageInputDTO struct {
	PageID      string `json:"page_id" jsonschema:"Source page ID (string) to clone,required"`
	Target      string `json:"target" jsonschema:"Target page slug where clone will be created,required"`
	Title       string `json:"title,omitempty" jsonschema:"New page title after cloning"`
	SubscribeMe bool   `json:"subscribe_me,omitempty" jsonschema:"Subscribe to changes on the cloned page (default: false)"`
}

// cloneOperationOutputDTO is the output for clone operations (page/grid).
type cloneOperationOutputDTO struct {
	OperationID   string `json:"operation_id"`
	OperationType string `json:"operation_type"`
	DryRun        bool   `json:"dry_run"`
	StatusURL     string `json:"status_url"`
}

// deleteGridInputDTO is the input for wiki_grid_delete.
type deleteGridInputDTO struct {
	GridID string `json:"grid_id" jsonschema:"Grid ID (UUID string) to delete,required"`
}

// deleteGridOutputDTO is the output for wiki_grid_delete (empty, 204 No Content).
type deleteGridOutputDTO struct{}

// cloneGridInputDTO is the input for wiki_grid_clone.
type cloneGridInputDTO struct {
	GridID   string `json:"grid_id" jsonschema:"Source grid ID (UUID string) to clone,required"`
	Target   string `json:"target" jsonschema:"Target page slug where grid will be copied; page created if not exists,required"`
	Title    string `json:"title,omitempty" jsonschema:"New grid title after copying (1-255 chars)"`
	WithData bool   `json:"with_data,omitempty" jsonschema:"Copy grid rows (default: false)"`
}

// addGridRowsInputDTO is the input for wiki_grid_rows_add.
type addGridRowsInputDTO struct {
	GridID     string           `json:"grid_id" jsonschema:"Grid ID (UUID string) to add rows to,required"`
	Rows       []map[string]any `json:"rows" jsonschema:"Array of row objects; each object is a mapping of column_slug to value,required"`
	AfterRowID string           `json:"after_row_id,omitempty" jsonschema:"Insert rows after this row ID"`
	Position   *int             `json:"position,omitempty" jsonschema:"Absolute insertion position (0-based)"`
	Revision   string           `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}

// addGridRowsOutputDTO is the output for wiki_grid_rows_add.
type addGridRowsOutputDTO struct {
	Revision string                 `json:"revision"`
	Results  []gridRowResultItemDTO `json:"results"`
}

// gridRowResultItemDTO represents a row result from grid row operations.
type gridRowResultItemDTO struct {
	ID     string `json:"id"`
	Row    []any  `json:"row"`
	Color  string `json:"color,omitempty"`
	Pinned bool   `json:"pinned,omitempty"`
}

// deleteGridRowsInputDTO is the input for wiki_grid_rows_delete.
type deleteGridRowsInputDTO struct {
	GridID   string   `json:"grid_id" jsonschema:"Grid ID (UUID string) to delete rows from,required"`
	RowIDs   []string `json:"row_ids" jsonschema:"Row IDs (strings) to delete (minimum 1),required"`
	Revision string   `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}

// revisionOutputDTO is the output for operations returning only revision.
type revisionOutputDTO struct {
	Revision string `json:"revision"`
}

// moveGridRowsInputDTO is the input for wiki_grid_rows_move.
type moveGridRowsInputDTO struct {
	GridID     string `json:"grid_id" jsonschema:"Grid ID (UUID string) to move rows in,required"`
	RowID      string `json:"row_id" jsonschema:"Starting row ID (string) to move,required"`
	AfterRowID string `json:"after_row_id,omitempty" jsonschema:"Move rows to after this row ID (string)"`
	Position   *int   `json:"position,omitempty" jsonschema:"Move to absolute position (0-based)"`
	RowsCount  *int   `json:"rows_count,omitempty" jsonschema:"Number of consecutive rows to move starting from row_id (exclusive min: 0)"`
	Revision   string `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}

// addGridColumnsInputDTO is the input for wiki_grid_columns_add.
type addGridColumnsInputDTO struct {
	GridID   string                     `json:"grid_id" jsonschema:"Grid ID (UUID string) to add columns to,required"`
	Columns  []columnDefinitionInputDTO `json:"columns" jsonschema:"Array of column definitions,required"`
	Position *int                       `json:"position,omitempty" jsonschema:"Insertion position (0-based)"`
	Revision string                     `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}

// columnDefinitionInputDTO represents a column definition for grid column creation.
type columnDefinitionInputDTO struct {
	Slug          string   `json:"slug" jsonschema:"Column identifier (alphanumeric underscores),required"`
	Title         string   `json:"title" jsonschema:"Column display title (1-255 chars),required"`
	Type          string   `json:"type" jsonschema:"Column type. Allowed values: string number date select staff checkbox ticket ticket_field,required"`
	Required      bool     `json:"required" jsonschema:"Whether column value is required,required"`
	Description   string   `json:"description,omitempty" jsonschema:"Column description (max 1000 chars)"`
	Color         string   `json:"color,omitempty" jsonschema:"Column header color. Allowed values: blue yellow pink red green mint grey orange magenta purple copper ocean"`
	Format        string   `json:"format,omitempty" jsonschema:"Text format for string columns only. Allowed values: yfm wom plain"`
	SelectOptions []string `json:"select_options,omitempty" jsonschema:"Options for select column type"`
	Multiple      bool     `json:"multiple,omitempty" jsonschema:"Enable multiple selection for select and staff column types"`
	MarkRows      bool     `json:"mark_rows,omitempty" jsonschema:"For checkbox columns: mark row as completed in UI"`
	TicketField   string   `json:"ticket_field,omitempty" jsonschema:"Tracker field for ticket_field column type. Allowed values: assignee components created_at deadline description end estimation fixversions followers last_comment_updated_at original_estimation parent pending_reply_from priority project queue reporter resolution resolved_at sprint start status status_start_time status_type storypoints subject tags type updated_at votes"`
	Width         *int     `json:"width,omitempty" jsonschema:"Column width value"`
	WidthUnits    string   `json:"width_units,omitempty" jsonschema:"Column width units. Allowed values: % px"`
	Pinned        string   `json:"pinned,omitempty" jsonschema:"Pin column position. Allowed values: left right"`
}

// deleteGridColumnsInputDTO is the input for wiki_grid_columns_delete.
type deleteGridColumnsInputDTO struct {
	GridID      string   `json:"grid_id" jsonschema:"Grid ID (UUID string) to delete columns from,required"`
	ColumnSlugs []string `json:"column_slugs" jsonschema:"Column slugs to delete,required"`
	Revision    string   `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}

// moveGridColumnsInputDTO is the input for wiki_grid_columns_move.
type moveGridColumnsInputDTO struct {
	GridID       string `json:"grid_id" jsonschema:"Grid ID (UUID string) to move columns in,required"`
	ColumnSlug   string `json:"column_slug" jsonschema:"Starting column slug to move,required"`
	Position     int    `json:"position" jsonschema:"Destination position (0-based),required"`
	ColumnsCount *int   `json:"columns_count,omitempty" jsonschema:"Number of consecutive columns to move (exclusive min: 0)"`
	Revision     string `json:"revision,omitempty" jsonschema:"Current revision for optimistic locking"`
}
