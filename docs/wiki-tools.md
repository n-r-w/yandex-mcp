# Yandex Wiki MCP tools

This document lists the MCP tools implemented under `internal/tools/wiki/`.

Tool names and one-line descriptions are taken from tool registration in `internal/tools/wiki/service.go`.
Input/output schemas are derived from the MCP handler layer DTOs in `internal/tools/wiki/dto.go`.

## Write operations gating

By default, the server registers only read-only Wiki tools.

To enable Wiki write tools, start the server with:

- `--wiki-write` (default: false)

Write-gated tools in this document are explicitly marked.

## Conventions

- Types are described using JSON-compatible terms (string, number/integer, boolean, array, object).
- “Required” means the tool validates the parameter as required (and/or marks it required in the schema).
- Timestamp fields are strings as returned by the upstream Yandex Wiki API.

## wiki_page_get

Retrieves a Yandex Wiki page by its slug (URL path).

### Input

- `slug` (string, required): Page slug (URL path).
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `revision_id` (integer, optional): Fetch a specific page revision by ID.
- `raise_on_redirect` (boolean, optional): Return an error if the page redirects instead of following the redirect.

### Output

Returns `PageOutput`:

- `id` (integer)
- `page_type` (string)
- `slug` (string)
- `title` (string)
- `content` (string, optional)
- `attributes` (object, optional): `AttributesOutput`
  - `comments_count` (integer)
  - `comments_enabled` (boolean)
  - `created_at` (string)
  - `is_readonly` (boolean)
  - `lang` (string)
  - `modified_at` (string)
  - `is_collaborative` (boolean)
  - `is_draft` (boolean)
- `redirect` (object, optional): `RedirectOutput`
  - `page_id` (integer)
  - `slug` (string)

## wiki_page_get_by_id

Retrieves a Yandex Wiki page by its numeric ID.

### Input

- `page_id` (integer, required): Page ID. Must be positive.
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `revision_id` (integer, optional): Fetch a specific page revision by ID.
- `raise_on_redirect` (boolean, optional): Return an error if the page redirects instead of following the redirect.

### Output

Returns `PageOutput` (same shape as `wiki_page_get`).

## wiki_page_resources_list

Lists resources (attachments, grids) for a Yandex Wiki page.

### Input

- `page_id` (integer, required): Page ID to list resources for. Must be positive.
- `cursor` (string, optional): Pagination cursor for subsequent requests.
- `page_size` (integer, optional): Number of items per page.
  - Tool validation: must be non-negative and must not exceed 50.
- `order_by` (string, optional): Field to order by.
  - Allowed values: `name_title`, `created_at`
- `order_direction` (string, optional): Order direction.
  - Allowed values: `asc`, `desc`
- `q` (string, optional): Filter resources by title.
- `types` (string, optional): Resource types filter.
  - Allowed values: `attachment`, `sharepoint_resource`, `grid`
  - Multiple values can be comma-separated.
- `page_id_legacy` (integer, optional): Legacy page number for backward-compatibility pagination.

### Output

Returns `ResourcesListOutput`:

- `resources` (array of object): array of `ResourceOutput`
  - `type` (string)
  - `item` (object): shape depends on `type`
    - If `type` is `attachment`, `item` is `AttachmentOutput`
      - `id` (integer)
      - `name` (string)
      - `size` (integer)
      - `mimetype` (string)
      - `download_url` (string)
      - `created_at` (string)
      - `has_preview` (boolean)
    - If `type` is `sharepoint_resource`, `item` is `SharepointResourceOutput`
      - `id` (integer)
      - `title` (string)
      - `doctype` (string)
      - `created_at` (string)
    - If `type` is `grid`, `item` is `GridResourceOutput`
      - `id` (string)
      - `title` (string)
      - `created_at` (string)
- `next_cursor` (string, optional)
- `prev_cursor` (string, optional)

## wiki_page_grids_list

Lists dynamic tables (grids) for a Yandex Wiki page.

### Input

- `page_id` (integer, required): Page ID to list grids for. Must be positive.
- `cursor` (string, optional): Pagination cursor for subsequent requests.
- `page_size` (integer, optional): Number of items per page.
  - Tool validation: must be non-negative and must not exceed 50.
- `order_by` (string, optional): Field to order by.
  - Allowed values: `title`, `created_at`
- `order_direction` (string, optional): Order direction.
  - Allowed values: `asc`, `desc`
- `page_id_legacy` (integer, optional): Legacy page number for backward-compatibility pagination.

### Output

Returns `GridsListOutput`:

- `grids` (array of object): array of `GridSummaryOutput`
  - `id` (string)
  - `title` (string)
  - `created_at` (string)
- `next_cursor` (string, optional)
- `prev_cursor` (string, optional)

## wiki_grid_get

Retrieves a Yandex Wiki dynamic table (grid) by its ID.

### Input

- `grid_id` (string, required): Grid ID (UUID string).
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `user_permissions`
- `filter` (string, optional): Row filter expression to filter grid rows.
  - Syntax: `[column_slug] operator value`
  - Operators: `~` (contains), `<`, `>`, `<=`, `>=`, `=`, `!`
  - Logical: `AND`, `OR`, `(`, `)`
- `only_cols` (string, optional): Return only specified columns (comma-separated column slugs).
- `only_rows` (string, optional): Return only specified rows (comma-separated row IDs).
- `revision` (integer, optional): Grid revision number for optimistic locking and historical versions.
- `sort` (string, optional): Sort expression to order rows by column.

### Output

Returns `GridOutput`:

- `id` (string)
- `title` (string)
- `structure` (array of object, optional): array of `ColumnOutput`
  - `slug` (string)
  - `title` (string)
  - `type` (string)
- `rows` (array of object, optional): array of `GridRowOutput`
  - `id` (string)
  - `cells` (object): map from column slug to cell value
- `revision` (string)
- `created_at` (string)
- `rich_text_format` (string)
- `attributes` (object, optional): `AttributesOutput` (same shape as in `PageOutput`)

## wiki_page_create

Creates a new Yandex Wiki page.

This tool is write-gated and requires `--wiki-write`.

### Input

- `slug` (string, required): Page slug (URL path).
- `title` (string, required): Page title.
- `page_type` (string, required): Page type.
  - Allowed values: `page`, `grid`, `cloud_page`, `wysiwyg`, `template`
- `content` (string, optional): Page content in wikitext format.
- `is_silent` (boolean, optional): Suppress notifications for this operation.
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `cloud_page` (object, optional): Cloud page options for `cloud_page` type.
  - `method` (string, required): Method for creating cloud page.
    - Allowed values: `empty_doc`, `from_url`, `upload_doc`, `finalize_upload`, `upload_onprem`
  - `doctype` (string, required): Document type.
    - Allowed values: `docx`, `pptx`, `xlsx`
- `grid_format` (string, optional): Text format for grid columns.
  - Allowed values: `yfm`, `wom`, `plain`

### Output

Returns `PageOutput`.

## wiki_page_update

Updates an existing Yandex Wiki page.

This tool is write-gated and requires `--wiki-write`.

### Input

- `page_id` (integer, required): Page ID. Must be positive.
- `title` (string, optional): Page title.
- `content` (string, optional): Page content in wikitext format.
- `allow_merge` (boolean, optional): Enable 3-way merge for concurrent edits.
- `is_silent` (boolean, optional): Suppress notifications for this operation.
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `redirect` (object, optional): Set or remove page redirect.
  - `page_id` (integer or null, optional): Target page ID for redirect. Set to null to remove redirect.
  - `slug` (string or null, optional): Target page slug for redirect. If both `page_id` and `slug` are provided, `page_id` is used.

Note: the tool requires at least one of `title`, `content`, or `redirect` to be provided.

### Output

Returns `PageOutput`.

## wiki_page_append_content

Appends content to an existing Yandex Wiki page.

This tool is write-gated and requires `--wiki-write`.

### Input

- `page_id` (integer, required): Page ID. Must be positive.
- `content` (string, required): Content to append in wikitext format.
- `is_silent` (boolean, optional): Suppress notifications for this operation.
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `body` (object, optional): Append to top or bottom of page body.
  - `location` (string, required): Append location within body.
    - Allowed values: `top`, `bottom`
- `section` (object, optional): Append to top or bottom of specific section.
  - `id` (integer, required): Section ID.
  - `location` (string, required): Append location within section.
    - Allowed values: `top`, `bottom`
- `anchor` (object, optional): Append relative to named anchor.
  - `name` (string, required): Anchor name.
  - `fallback` (boolean, optional): Fall back to default behavior if anchor is not found.
  - `regex` (boolean, optional): Treat anchor name as regular expression.

### Output

Returns `PageOutput`.

## wiki_grid_create

Creates a new Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `page` (object, required): Page where the grid will be created.
  - `id` (integer, optional): Page ID.
  - `slug` (string, optional): Page slug (URL path).
  - Note: the tool requires at least one of `page.id` or `page.slug`.
- `title` (string, required): Grid title.
- `columns` (array of object, required): Grid columns definition.
  - Each item is `ColumnInputCreate`:
    - `slug` (string, required): Column slug (ID).
    - `title` (string, required): Column title.
    - `type` (string, optional): Column type.
      - Allowed values: `string`, `number`, `date`, `select`, `staff`, `checkbox`, `ticket`, `ticket_field`
- `fields` (string, optional): Additional fields to include.
  - Allowed values: `attributes`, `user_permissions`
  - Format: comma-separated values.

### Output

Returns `GridOutput`.

## wiki_grid_update_cells

Updates cells in a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string).
- `cells` (array of object, required): Array of cell updates.
  - Each item is `CellUpdateInput`:
    - `row_id` (integer, required): Row ID. Must be positive.
    - `column_slug` (string, required): Column slug.
    - `value` (string, required): Cell value.
      - Note: the schema accepts any JSON value, but the tool currently validates that `value` is a string.
- `revision` (string, optional): Grid revision for optimistic locking.

### Output

Returns `GridOutput`.

## wiki_page_delete

Deletes a Yandex Wiki page.

This tool is write-gated and requires `--wiki-write`.

### Input

- `page_id` (integer, required): Page ID to delete. Must be positive.

### Output

Returns `DeletePageOutput`:

- `recovery_token` (string): Recovery token for potential page restoration.

## wiki_page_clone

Clones a Yandex Wiki page to a new location (async operation).

This tool is write-gated and requires `--wiki-write`.

### Input

- `page_id` (integer, required): Source page ID to clone. Must be positive.
- `target` (string, required): Target page slug where clone will be created.
- `title` (string, optional): New page title after cloning.
- `subscribe_me` (boolean, optional): Subscribe to changes on the cloned page (default: false).

Note: clone is asynchronous; the output contains a `status_url` for polling operation status.

### Output

Returns `CloneOperationOutput`:

- `operation_id` (string): Async operation ID.
- `operation_type` (string): Operation type identifier.
- `dry_run` (boolean): Whether the operation was executed as a dry run.
- `status_url` (string): URL for polling operation status.

## wiki_grid_delete

Deletes a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to delete.

### Output

No output. Success is indicated by the absence of an error.

## wiki_grid_clone

Clones a Yandex Wiki grid to a new location (async operation).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Source grid ID (UUID string) to clone.
- `target` (string, required): Target page slug where the grid will be copied; the page is created if it does not exist.
- `title` (string, optional): New grid title after copying (1-255 chars).
- `with_data` (boolean, optional): Copy grid rows (default: false).

Note: clone is asynchronous; the output contains a `status_url` for polling operation status.

### Output

Returns `CloneOperationOutput` (same shape as `wiki_page_clone`).

## wiki_grid_rows_add

Adds rows to a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to add rows to.
- `rows` (array of object, required): Array of row objects; each object is a mapping of `column_slug` to value.
  - Tool validation: must contain at least one element.
- `after_row_id` (string, optional): Insert rows after this row ID.
- `position` (integer, optional): Absolute insertion position (0-based).
- `revision` (string, optional): Current revision for optimistic locking.

### Output

Returns `AddGridRowsOutput`:

- `revision` (string): New grid revision.
- `results` (array of object): Array of row result items.

Row result item:

- `id` (string): Row ID.
- `row` (array): Cell values.
- `color` (string, optional): Row color value returned by the API.
- `pinned` (boolean, optional): Row pinned status.

## wiki_grid_rows_delete

Deletes rows from a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to delete rows from.
- `row_ids` (array of string, required): Row IDs to delete.
  - Tool validation: must contain at least one element.
- `revision` (string, optional): Current revision for optimistic locking.

### Output

Returns `RevisionOutput`:

- `revision` (string): New grid revision.

## wiki_grid_rows_move

Moves rows within a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to move rows in.
- `row_id` (string, required): Starting row ID to move.
- `after_row_id` (string, optional): Move rows to after this row ID.
- `position` (integer, optional): Move to absolute position (0-based).
- `rows_count` (integer, optional): Number of consecutive rows to move starting from `row_id` (must be greater than 0 if provided).
- `revision` (string, optional): Current revision for optimistic locking.

### Output

Returns `RevisionOutput` (same shape as `wiki_grid_rows_delete`).

## wiki_grid_columns_add

Adds columns to a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to add columns to.
- `columns` (array of object, required): Array of column definitions.
  - Tool validation: must contain at least one element.
- `position` (integer, optional): Insertion position (0-based).
- `revision` (string, optional): Current revision for optimistic locking.

Each column definition:

- `slug` (string, required): Column identifier (alphanumeric underscores).
- `title` (string, required): Column display title (1-255 chars).
- `type` (string, required): Column type.
  - Allowed values: `string`, `number`, `date`, `select`, `staff`, `checkbox`, `ticket`, `ticket_field`
- `required` (boolean, required): Whether column value is required.
- `description` (string, optional): Column description (max 1000 chars).
- `color` (string, optional): Column header color.
  - Allowed values: `blue`, `yellow`, `pink`, `red`, `green`, `mint`, `grey`, `orange`, `magenta`, `purple`, `copper`, `ocean`
- `format` (string, optional): Text format for string columns only.
  - Allowed values: `yfm`, `wom`, `plain`
- `select_options` (array of string, optional): Options for select column type.
- `multiple` (boolean, optional): Enable multiple selection for `select` and `staff` column types.
- `mark_rows` (boolean, optional): For checkbox columns: mark row as completed in UI.
- `ticket_field` (string, optional): Tracker field for `ticket_field` column type.
  - Allowed values: `assignee`, `components`, `created_at`, `deadline`, `description`, `end`, `estimation`, `fixversions`, `followers`, `last_comment_updated_at`, `original_estimation`, `parent`, `pending_reply_from`, `priority`, `project`, `queue`, `reporter`, `resolution`, `resolved_at`, `sprint`, `start`, `status`, `status_start_time`, `status_type`, `storypoints`, `subject`, `tags`, `type`, `updated_at`, `votes`
- `width` (integer, optional): Column width value.
- `width_units` (string, optional): Column width units.
  - Allowed values: `%`, `px`
- `pinned` (string, optional): Pin column position.
  - Allowed values: `left`, `right`

### Output

Returns `RevisionOutput` (same shape as `wiki_grid_rows_delete`).

## wiki_grid_columns_delete

Deletes columns from a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to delete columns from.
- `column_slugs` (array of string, required): Column slugs to delete.
  - Tool validation: must contain at least one element.
- `revision` (string, optional): Current revision for optimistic locking.

### Output

Returns `RevisionOutput` (same shape as `wiki_grid_rows_delete`).

## wiki_grid_columns_move

Moves columns within a Yandex Wiki dynamic table (grid).

This tool is write-gated and requires `--wiki-write`.

### Input

- `grid_id` (string, required): Grid ID (UUID string) to move columns in.
- `column_slug` (string, required): Starting column slug to move.
- `position` (integer, required): Destination position (0-based).
- `columns_count` (integer, optional): Number of consecutive columns to move (must be greater than 0 if provided).
- `revision` (string, optional): Current revision for optimistic locking.

### Output

Returns `RevisionOutput` (same shape as `wiki_grid_rows_delete`).

## Planned / Requires Additional Research

Items in this section are not implemented and are not available as MCP tools yet.

- Update grid settings beyond cell values (for example, changing default sort for a grid).
  - Requires: upstream schema for the sort configuration object and how it interacts with `wiki_grid_get` and revisions.
  - Tentative tool name (not callable): `wiki_grid_update`
