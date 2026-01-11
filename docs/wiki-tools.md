# Yandex Wiki MCP tools

This document lists the MCP tools implemented under `internal/tools/wiki/`.

Tool names and one-line descriptions are taken from tool registration in `internal/tools/wiki/service.go`.
Input/output schemas are derived from the MCP handler layer DTOs in `internal/tools/wiki/dto.go`.

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
- `revision_id` (string, optional): Fetch a specific page revision by ID.
- `raise_on_redirect` (boolean, optional): Return an error if the page redirects instead of following the redirect.

### Output

Returns `PageOutput`:

- `id` (string)
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
  - `page_id` (string)
  - `slug` (string)

## wiki_page_get_by_id

Retrieves a Yandex Wiki page by its numeric ID.

### Input

- `page_id` (string, required): Page ID. Must be positive.
- `fields` (array of string, optional): Additional fields to include.
  - Allowed values: `attributes`, `breadcrumbs`, `content`, `redirect`
- `revision_id` (string, optional): Fetch a specific page revision by ID.
- `raise_on_redirect` (boolean, optional): Return an error if the page redirects instead of following the redirect.

### Output

Returns `PageOutput` (same shape as `wiki_page_get`).

## wiki_page_resources_list

Lists resources (attachments, grids) for a Yandex Wiki page.

### Input

- `page_id` (string, required): Page ID to list resources for. Must be positive.
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

### Output

Returns `ResourcesListOutput`:

- `resources` (array of object): array of `ResourceOutput`
  - `type` (string)
  - `item` (object): shape depends on `type`
    - If `type` is `attachment`, `item` is `AttachmentOutput`
      - `id` (string)
      - `name` (string)
      - `size` (integer)
      - `mimetype` (string)
      - `download_url` (string)
      - `created_at` (string)
      - `has_preview` (boolean)
    - If `type` is `sharepoint_resource`, `item` is `SharepointResourceOutput`
      - `id` (string)
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

- `page_id` (string, required): Page ID to list grids for. Must be positive.
- `cursor` (string, optional): Pagination cursor for subsequent requests.
- `page_size` (integer, optional): Number of items per page.
  - Tool validation: must be non-negative and must not exceed 50.
- `order_by` (string, optional): Field to order by.
  - Allowed values: `title`, `created_at`
- `order_direction` (string, optional): Order direction.
  - Allowed values: `asc`, `desc`

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
- `revision` (string, optional): Grid revision number for optimistic locking and historical versions.
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

