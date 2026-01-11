//nolint:lll
package wiki

// Input DTOs for wiki tools.

// getPageBySlugInputDTO is the input for wiki_page_get tool.
type getPageBySlugInputDTO struct {
	Slug            string   `json:"slug" jsonschema:"Page slug (URL path). Required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      string   `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID (string)"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// getPageByIDInputDTO is the input for wiki_page_get_by_id tool.
type getPageByIDInputDTO struct {
	PageID          string   `json:"page_id" jsonschema:"Page ID (string). Required"`
	Fields          []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, breadcrumbs, content, redirect"`
	RevisionID      string   `json:"revision_id,omitempty" jsonschema:"Fetch specific page revision by ID (string)"`
	RaiseOnRedirect bool     `json:"raise_on_redirect,omitempty" jsonschema:"Return error if page redirects instead of following redirect"`
}

// listResourcesInputDTO is the input for wiki_page_resources_list tool.
type listResourcesInputDTO struct {
	PageID         string `json:"page_id" jsonschema:"Page ID (string) to list resources for. Required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page. Valid range: 1-50. Default: 25"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: name_title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc, desc. Default: asc"`
	Q              string `json:"q,omitempty" jsonschema:"Filter resources by title. Maximum: 255 chars"`
	Types          string `json:"types,omitempty" jsonschema:"Resource types filter. Possible values: attachment, sharepoint_resource, grid. Can be comma-separated for multiple types"`
}

// listGridsInputDTO is the input for wiki_page_grids_list tool.
type listGridsInputDTO struct {
	PageID         string `json:"page_id" jsonschema:"Page ID (string) to list grids for. Required"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"Pagination cursor for subsequent requests"`
	PageSize       int    `json:"page_size,omitempty" jsonschema:"Number of items per page. Valid range: 1-50. Default: 25"`
	OrderBy        string `json:"order_by,omitempty" jsonschema:"Field to order by. Possible values: title, created_at"`
	OrderDirection string `json:"order_direction,omitempty" jsonschema:"Order direction. Possible values: asc, desc. Default: asc"`
}

// getGridInputDTO is the input for wiki_grid_get tool.
type getGridInputDTO struct {
	GridID   string   `json:"grid_id" jsonschema:"Grid ID (UUID string). Required"`
	Fields   []string `json:"fields,omitempty" jsonschema:"Additional fields to include in the response. Allowed values: attributes, user_permissions"`
	Filter   string   `json:"filter,omitempty" jsonschema:"Row filter expression to filter grid rows. Syntax: [column_slug] operator value. Operators: ~ (contains), <, >, <=, >=, =, !. Logical: AND, OR, (). Example: [slug] ~ wiki AND [slug2]<32"`
	OnlyCols string   `json:"only_cols,omitempty" jsonschema:"Return only specified columns (comma-separated column slugs)"`
	OnlyRows string   `json:"only_rows,omitempty" jsonschema:"Return only specified rows (comma-separated row IDs)"`
	Revision string   `json:"revision,omitempty" jsonschema:"Grid revision number for optimistic locking and historical versions"`
	Sort     string   `json:"sort,omitempty" jsonschema:"Sort expression to order rows by column"`
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
