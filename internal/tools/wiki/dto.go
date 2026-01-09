//nolint:lll
package wiki

// Input DTOs for wiki tools.

// GetPageBySlugInput is the input for wiki_page_get tool.
type GetPageBySlugInput struct {
	Slug            string   `json:"slug" jsonschema:"Page slug (URL path),required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      int      `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// GetPageByIDInput is the input for wiki_page_get_by_id tool.
type GetPageByIDInput struct {
	PageID          int64    `json:"page_id" jsonschema:"Page ID,required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      int      `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// ListResourcesInput is the input for wiki_page_resources_list tool.
type ListResourcesInput struct {
	PageID         int64  `json:"page_id" jsonschema:"Page ID to list resources for,required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page (default: 25, min: 1, max: 50)"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: name_title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc (default), desc"`
	Q              string `json:"q,omitempty" jsonschema:"Filter resources by title (max 255 chars)"`
	Types          string `json:"types,omitempty" jsonschema:"Resource types filter. Possible values: attachment, sharepoint_resource, grid. Can be comma-separated for multiple types"`
	PageIDLegacy   int    `json:"page_id_legacy,omitempty" jsonschema:"Legacy page number for backward-compatibility pagination (default: 1)"`
}

// ListGridsInput is the input for wiki_page_grids_list tool.
type ListGridsInput struct {
	PageID         int64  `json:"page_id" jsonschema:"Page ID to list grids for,required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page (default: 25, min: 1, max: 50)"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc (default), desc"`
	PageIDLegacy   int    `json:"page_id_legacy,omitempty" jsonschema:"Legacy page number for backward-compatibility pagination (default: 1)"`
}

// GetGridInput is the input for wiki_grid_get tool.
type GetGridInput struct {
	GridID   string   `json:"grid_id" jsonschema:"Grid ID (UUID string),required"`
	Fields   []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, user_permissions"`
	Filter   string   `json:"filter,omitempty" jsonschema:"Row filter expression to filter grid rows. Syntax: [column_slug] operator value. Operators: ~ (contains), <, >, <=, >=, =, !. Logical: AND, OR, (). Example: [slug] ~ wiki AND [slug2]<32"`
	OnlyCols string   `json:"only_cols,omitempty" jsonschema:"Return only specified columns (comma-separated column slugs)"`
	OnlyRows string   `json:"only_rows,omitempty" jsonschema:"Return only specified rows (comma-separated row IDs)"`
	Revision int      `json:"revision,omitempty" jsonschema:"Grid revision number for optimistic locking and historical versions"`
	Sort     string   `json:"sort,omitempty" jsonschema:"Sort expression to order rows by column"`
}

// Write tool input DTOs.

// CreatePageInput is the input for wiki_page_create tool.
type CreatePageInput struct {
	Slug       string          `json:"slug" jsonschema:"Page slug (URL path),required"`
	Title      string          `json:"title" jsonschema:"Page title,required"`
	PageType   string          `json:"page_type" jsonschema:"Page type. Possible values: page, grid, cloud_page, wysiwyg, template,required"`
	Content    string          `json:"content,omitempty" jsonschema:"Page content in wikitext format"`
	IsSilent   bool            `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields     []string        `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	CloudPage  *CloudPageInput `json:"cloud_page,omitempty" jsonschema:"Cloud page options for cloud_page type"`
	GridFormat string          `json:"grid_format,omitempty" jsonschema:"Text format for grid columns. Possible values: yfm, wom, plain"`
}

// UpdatePageInput is the input for wiki_page_update tool.
type UpdatePageInput struct {
	PageID     int            `json:"page_id" jsonschema:"Page ID,required"`
	Title      string         `json:"title,omitempty" jsonschema:"Page title"`
	Content    string         `json:"content,omitempty" jsonschema:"Page content in wikitext format"`
	AllowMerge bool           `json:"allow_merge,omitempty" jsonschema:"Enable 3-way merge for concurrent edits"`
	IsSilent   bool           `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields     []string       `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	Redirect   *RedirectInput `json:"redirect,omitempty" jsonschema:"Set or remove page redirect"`
}

// AppendPageInput is the input for wiki_page_append_content tool.
type AppendPageInput struct {
	PageID   int              `json:"page_id" jsonschema:"Page ID,required"`
	Content  string           `json:"content" jsonschema:"Content to append in wikitext format,required"`
	IsSilent bool             `json:"is_silent,omitempty" jsonschema:"Suppress notifications for this operation"`
	Fields   []string         `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	Body     *BodyLocation    `json:"body,omitempty" jsonschema:"Append to top or bottom of page body"`
	Section  *SectionLocation `json:"section,omitempty" jsonschema:"Append to top or bottom of specific section"`
	Anchor   *AnchorLocation  `json:"anchor,omitempty" jsonschema:"Append relative to named anchor"`
}

// PageInput represents page identification (by ID or slug).
type PageInput struct {
	ID   int64  `json:"id,omitempty" jsonschema:"Page ID"`
	Slug string `json:"slug,omitempty" jsonschema:"Page slug (URL path)"`
}

// CreateGridInput is the input for wiki_grid_create tool.
type CreateGridInput struct {
	Page    PageInput           `json:"page" jsonschema:"Page where grid will be created (provide id or slug),required"`
	Title   string              `json:"title" jsonschema:"Grid title,required"`
	Columns []ColumnInputCreate `json:"columns" jsonschema:"Grid columns definition,required"`
	Fields  string              `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, user_permissions"`
}

// ColumnInputCreate defines a column for grid creation.
type ColumnInputCreate struct {
	Slug  string `json:"slug" jsonschema:"Column slug (ID),required"`
	Title string `json:"title" jsonschema:"Column title,required"`
	Type  string `json:"type" jsonschema:"Column type. Possible values: string, number, date, select, staff, checkbox, ticket, ticket_field"`
}

// UpdateGridCellsInput is the input for wiki_grid_update_cells tool.
type UpdateGridCellsInput struct {
	GridID   string            `json:"grid_id" jsonschema:"Grid ID (UUID string),required"`
	Cells    []CellUpdateInput `json:"cells" jsonschema:"Array of cell updates,required"`
	Revision string            `json:"revision,omitempty" jsonschema:"Grid revision for optimistic locking"`
}

// CellUpdateInput represents a single cell update.
type CellUpdateInput struct {
	RowID      int    `json:"row_id" jsonschema:"Row ID,required"`
	ColumnSlug string `json:"column_slug" jsonschema:"Column slug,required"`
	Value      any    `json:"value" jsonschema:"Cell value (string only in v1),required"`
}

// CloudPageInput represents cloud page options for page creation.
type CloudPageInput struct {
	Method  string `json:"method" jsonschema:"Method for creating cloud page. Possible values: empty_doc, from_url, upload_doc, finalize_upload, upload_onprem,required"`
	Doctype string `json:"doctype" jsonschema:"Document type. Possible values: docx, pptx, xlsx,required"`
}

// RedirectInput represents redirect options for page update.
type RedirectInput struct {
	PageID *int64  `json:"page_id,omitempty" jsonschema:"Target page ID for redirect. Set to null to remove redirect"`
	Slug   *string `json:"slug,omitempty" jsonschema:"Target page slug for redirect. If both page_id and slug provided, page_id is used"`
}

// BodyLocation represents body location targeting for content append.
type BodyLocation struct {
	Location string `json:"location" jsonschema:"Append location within body. Possible values: top, bottom,required"`
}

// SectionLocation represents section location targeting for content append.
type SectionLocation struct {
	ID       int    `json:"id" jsonschema:"Section ID,required"`
	Location string `json:"location" jsonschema:"Append location within section. Possible values: top, bottom,required"`
}

// AnchorLocation represents anchor location targeting for content append.
type AnchorLocation struct {
	Name     string `json:"name" jsonschema:"Anchor name,required"`
	Fallback bool   `json:"fallback,omitempty" jsonschema:"Fall back to default behavior if anchor not found"`
	Regex    bool   `json:"regex,omitempty" jsonschema:"Treat anchor name as regular expression"`
}

// Output DTOs for wiki tools.

// PageOutput is the output for page retrieval tools.
type PageOutput struct {
	ID         int64             `json:"id"`
	PageType   string            `json:"page_type"`
	Slug       string            `json:"slug"`
	Title      string            `json:"title"`
	Content    string            `json:"content,omitempty"`
	Attributes *AttributesOutput `json:"attributes,omitempty"`
	Redirect   *RedirectOutput   `json:"redirect,omitempty"`
}

// AttributesOutput contains page attributes.
type AttributesOutput struct {
	CommentsCount   int    `json:"comments_count"`
	CommentsEnabled bool   `json:"comments_enabled"`
	CreatedAt       string `json:"created_at"`
	IsReadonly      bool   `json:"is_readonly"`
	Lang            string `json:"lang"`
	ModifiedAt      string `json:"modified_at"`
	IsCollaborative bool   `json:"is_collaborative"`
	IsDraft         bool   `json:"is_draft"`
}

// RedirectOutput contains page redirect info.
type RedirectOutput struct {
	PageID int64  `json:"page_id"`
	Slug   string `json:"slug"`
}

// ResourcesListOutput is the output for wiki_page_resources_list tool.
type ResourcesListOutput struct {
	Resources  []ResourceOutput `json:"resources"`
	NextCursor string           `json:"next_cursor,omitempty"`
	PrevCursor string           `json:"prev_cursor,omitempty"`
}

// ResourceOutput represents a page resource.
type ResourceOutput struct {
	Type string `json:"type"`
	Item any    `json:"item"`
}

// AttachmentOutput represents a file attachment for serialization.
type AttachmentOutput struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Mimetype    string `json:"mimetype"`
	DownloadURL string `json:"download_url"`
	CreatedAt   string `json:"created_at"`
	HasPreview  bool   `json:"has_preview"`
}

// SharepointResourceOutput represents a SharePoint/MS365 document for serialization.
type SharepointResourceOutput struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Doctype   string `json:"doctype"`
	CreatedAt string `json:"created_at"`
}

// GridResourceOutput represents a grid resource item for serialization.
type GridResourceOutput struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// GridsListOutput is the output for wiki_page_grids_list tool.
type GridsListOutput struct {
	Grids      []GridSummaryOutput `json:"grids"`
	NextCursor string              `json:"next_cursor,omitempty"`
	PrevCursor string              `json:"prev_cursor,omitempty"`
}

// GridSummaryOutput represents a grid summary.
type GridSummaryOutput struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// GridOutput is the output for wiki_grid_get tool.
type GridOutput struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Structure   []ColumnOutput    `json:"structure,omitempty"`
	Rows        []GridRowOutput   `json:"rows,omitempty"`
	Revision    string            `json:"revision"`
	CreatedAt   string            `json:"created_at"`
	RichTextFmt string            `json:"rich_text_format"`
	Attributes  *AttributesOutput `json:"attributes,omitempty"`
}

// ColumnOutput represents a grid column.
type ColumnOutput struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// GridRowOutput represents a grid row.
type GridRowOutput struct {
	ID    string         `json:"id"`
	Cells map[string]any `json:"cells"`
}
