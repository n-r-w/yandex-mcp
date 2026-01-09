# Yandex Wiki API Capabilities - Raw Extraction

**Related memories:**
- 01-research-segmentation-overview
- 02-research-sources-map-official-docs

**Purpose:** Structured extraction of Yandex Wiki API capabilities from official documentation.

## Base Configuration

### API Base URL
- **URL:** `https://api.wiki.yandex.net`
- **Version:** v1
- **Source:** https://yandex.ru/support/wiki/en/api-ref/access

### Authentication Methods

#### OAuth 2.0
- **Use case:** Both Yandex 360 for Business and Yandex Cloud Organization organizations
- **Header format:** `Authorization: OAuth <OAuth_token>`
- **Org header:** `X-Org-Id` (Yandex 360) or `X-Cloud-Org-Id` (Yandex Cloud)
- **Permissions:**
  - `wiki:read` - read-only access
  - `wiki:write` - full operations (create, delete, edit)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/access#about_OAuth

#### IAM Token
- **Use case:** Yandex Cloud organizations only
- **Header format:** `Authorization: Bearer <IAM_token>`
- **Org header:** `X-Org-Id` or `X-Cloud-Org-Id`
- **Token lifetime:** Maximum 12 hours
- **Cannot use service accounts** - user accounts only
- **Source:** https://yandex.ru/support/wiki/en/api-ref/access#iam-token

### Error Format
All errors follow this structure:
```json
{
    "debug_message": "Human-readable description",
    "details": { /* field-specific errors */ },
    "error_code": "ERROR_CODE"
}
```
Common error codes: `VALIDATION_ERROR`, plus standard HTTP codes (401, 403, 404, etc.)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/

---

## Capabilities by Resource Area

### 1. Pages (Wiki Pages)

#### Get Page by Slug
- **Endpoint:** `GET /v1/pages`
- **Operation:** Read
- **Required parameters:**
  - `slug` (query) - page path/slug (e.g., "users/something/abc")
- **Optional parameters:**
  - `fields` - comma-separated list of additional fields to return (values: `attributes`, `breadcrumbs`, `content`, `redirect`)
  - `raise_on_redirect` (boolean) - error if page redirects
  - `revision_id` (integer) - fetch specific revision
- **Response:** PageDetailsSchema (id, page_type, slug, title, content, attributes, breadcrumbs, redirect)
- **Page types:** page, grid, cloud_page, wysiwyg, template
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details

#### Get Page by ID
- **Endpoint:** `GET /v1/pages/{idx}`
- **Operation:** Read
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Optional parameters:**
  - `fields` - comma-separated list of additional fields to return (values: `attributes`, `breadcrumbs`, `content`, `redirect`)
  - `raise_on_redirect` (boolean) - error if page redirects
  - `revision_id` (integer) - fetch specific revision
- **Response:** PageDetailsSchema:
  - required: `id`, `page_type`, `slug`, `title`
  - optional (only if requested via `fields`): `attributes`, `breadcrumbs`, `content`, `redirect`
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details_by_id

#### Create Page
- **Endpoint:** `POST /v1/pages`
- **Operation:** Write
- **Required body fields:**
  - `page_type` - page type (values: `page`, `grid`, `cloud_page`, `wysiwyg`, `template`)
  - `slug` - string (page path)
  - `title` - string (1-255 chars)
- **Optional body fields:**
  - `content` - string (page content)
  - `cloud_page` - for cloud_page type (MS365 documents) (method values: `empty_doc`, `from_url`, `upload_doc`, `finalize_upload`, `upload_onprem`; doctype values: `docx`, `pptx`, `xlsx`)
  - `grid_format` - text format for grid columns (values: `yfm`, `wom`, `plain`)
- **Optional query parameters:**
  - `fields` - additional fields to return (values: `attributes`, `breadcrumbs`, `content`, `redirect`)
  - `is_silent` (boolean) - suppress notifications
- **Response:** PageFullDetailsSchema:
  - required: `id`, `page_type`, `slug`, `title`
  - optional (only if requested via `fields`): `attributes`, `breadcrumbs`, `content`, `redirect`
- **Limitations:** MS365 upload requires multi-step process (upload_doc -> finalize_upload)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__create_public_page

#### Update Page
- **Endpoint:** `POST /v1/pages/{idx}`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Optional body fields:**
  - `content` - updated page content
  - `title` - updated title (1-255 chars)
  - `redirect` - set or remove redirect:
    - object: `{ "page": PageIdentity }` to set redirect
    - object: `{ "page": null }` to remove redirect
    - PageIdentity: `{ "id": <integer>, "slug": <string> }` (either `id` or `slug`; if both are provided, `id` is used)
- **Optional query parameters:**
  - `allow_merge` (boolean) - enable 3-way merge for concurrent edits (default: false)
  - `fields` - additional fields to return (values: `attributes`, `breadcrumbs`, `content`, `redirect`)
  - `is_silent` (boolean) - suppress notifications
- **Response:** Updated PageDetailsSchema
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__update_public_page_details

#### Delete Page
- **Endpoint:** `DELETE /v1/pages/{idx}`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Response body:**
  - `recovery_token` (UUID) - token for potential restoration
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__delete_page

#### Clone Page
- **Endpoint:** `POST /v1/pages/{idx}/clone`
- **Operation:** Write (async operation)
- **Required parameters:**
  - `idx` (path) - source page ID (integer)
- **Required body fields:**
  - `target` - target page slug
- **Optional body fields:**
  - `title` - new page title after cloning
  - `subscribe_me` (boolean) - subscribe to changes (default: false)
- **Response body:**
  - `operation` - OperationIdentity (`id`, `type`)
  - `dry_run` (boolean)
  - `status_url` (string)
- **Error codes:**
  - `IS_CLOUD_PAGE` - cannot clone cloud pages
  - `SLUG_OCCUPIED` - target slug already exists
  - `SLUG_RESERVED` - target is reserved
  - `FORBIDDEN` - no access
  - `QUOTA_EXCEEDED` - organization page limit reached
  - `CLUSTER_BLOCKED` - target cluster temporarily blocked
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__clone_page

#### Append Content
- **Endpoint:** `POST /v1/pages/{idx}/append-content`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Required body fields:**
  - `content` - content to append (min length depends on type)
- **Optional body fields** (location targeting):
  - `body` - append to top/bottom of page body:
    - object: `{ "location": "top" | "bottom" }`
  - `section` - append to top/bottom of a specific section:
    - object: `{ "id": <integer>, "location": "top" | "bottom" }`
  - `anchor` - append relative to a named anchor:
    - object: `{ "name": <string>, "fallback": <boolean>, "regex": <boolean> }`
    - `fallback` (boolean, default: false)
    - `regex` (boolean, default: false)
- **Optional query parameters:**
  - `fields` - additional fields to return (values: `attributes`, `breadcrumbs`, `content`, `redirect`)
  - `is_silent` (boolean) - suppress notifications
- **Response:** Updated PageDetailsSchema
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__append_content

#### Get Page Grids (Dynamic Tables)
- **Endpoint:** `GET /v1/pages/{idx}/grids`
- **Operation:** Read
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Optional query parameters:**
  - `cursor` - cursor for pagination
  - `order_by` - sort field (values: `title`, `created_at`)
  - `order_direction` - sort direction (values: `asc`, `desc`; default: `asc`)
  - `page_id` (integer) - page number for backward-compatibility pagination (default: 1)
  - `page_size` (integer) - results per page (default: 25, min: 1, max: 50)
- **Response body:**
  - `results` - array of PageGridsSchema (`id` UUID, `title`, `created_at`)
  - `next_cursor`, `prev_cursor` - cursors for navigating pages
  - `has_next`, `page_id` - returned for backward compatibility; if `cursor` is used, `page_id` is always 1 and clients should rely on `next_cursor`
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__page_grids

---

### 2. Dynamic Tables (Grids)

#### Create Grid
- **Endpoint:** `POST /v1/grids`
- **Operation:** Write
- **Required body fields:**
  - `page` - PageIdentity (identify page by `id` or `slug`)
  - `title` (string) - grid title (1-255 chars)
- **Optional query parameters:**
  - `fields` - additional fields to return (values: `attributes`, `user_permissions`)
- **Response body:** GridDetailsSchema:
  - `id` (string UUID4 | integer) - grid ID
  - `created_at` (string, date-time)
  - `title` (string)
  - `page` - PageIdentity (`id` integer, `slug` string)
  - `structure` - GridStructureSchema:
    - `default_sort` (ColumnSortSchema[]) - array of `{ slug, title, direction }`, `direction` is `asc` or `desc`
    - `columns` (ColumnSchema[])
  - `rich_text_format` (TextFormat | null) - enum: `yfm`, `wom`, `plain`
  - `rows` (GridRowSchema[])
  - `revision` (string)
  - `template_id` (integer)
  - `attributes` (GridAttributesSchema) - returned only when `fields` includes `attributes`
  - `user_permissions` (UserPermission[]) - returned only when `fields` includes `user_permissions`
- **Note:** Grid is created as a resource of the specified page
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__create_grid

#### Get Grid
- **Endpoint:** `GET /v1/grids/{idx}`
- **Operation:** Read
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Optional query parameters:**
  - `fields` - additional fields to return (values: `attributes`, `user_permissions`)
  - `filter` - filter rows
  - `only_cols` - return only specified columns (comma-separated slugs)
  - `only_rows` - return only specified rows (comma-separated IDs)
  - `revision` (integer) - load historical version
  - `sort` - sort rows by column
- **Response:** GridDetailsSchema with:
  - Structure (columns, types, formats)
  - Rows (cell values)
  - Column types: string, number, date, select, staff, checkbox, ticket, ticket_field
  - Rich text format: yfm, wom, plain
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__get_grid

#### Update Grid
- **Endpoint:** `POST /v1/grids/{idx}`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Optional body fields:**
  - `title` - new title (1-255 chars)
  - `revision` - current revision (optimistic locking)
  - `default_sort` - array of objects
    - ⚠️ The Update Grid documentation defines this as `object[]` but does not document the object fields.
    - The grid structure (`GridStructureSchema.default_sort`) uses `ColumnSortSchema`:
      - `slug` (string) - column slug
      - `title` (string) - column title
      - `direction` (string) - sort direction (values: `asc`, `desc`)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__update_grid

#### Delete Grid
- **Endpoint:** `DELETE /v1/grids/{idx}`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Response:** 204 No Content
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__delete_grid

#### Clone Grid
- **Endpoint:** `POST /v1/grids/{idx}/clone`
- **Operation:** Write (async operation)
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `target` (string) - target page slug where the grid should be copied; if the page does not exist, it will be created
  - `title` (string, optional) - new grid title after copying (1-255 chars)
  - `with_data` (boolean, optional, default: false) - whether to copy grid rows
- **Response body:**
  - `operation` - OperationIdentity (`id`, `type`)
  - `dry_run` (boolean)
  - `status_url` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__clone_grid

#### Add Rows
- **Endpoint:** `POST /v1/grids/{idx}/rows`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Required body fields:**
  - `rows` - array of objects; each object represents a row as a mapping `column_slug -> value`
    - Allowed value types (per docs): integer, number, boolean, string, string[], UserIdentityExtended[]
- **Optional body fields:**
  - `after_row_id` (string) - insert after row with this ID
  - `position` (integer) - absolute insertion position
  - `revision` (string) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
  - `results` - array of GridRowSchema (`id`, `row`, optional `pinned`, optional `color`)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__add_rows

#### Remove Rows
- **Endpoint:** `DELETE /v1/grids/{idx}/rows`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `row_ids` (string[], min items: 1) - row IDs to delete
  - `revision` (string, optional) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__remove_rows

#### Move Rows
- **Endpoint:** `POST /v1/grids/{idx}/rows/move`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `row_id` (string) - starting row ID to move
  - `after_row_id` (string, optional) - move to after this row ID
  - `position` (integer, optional) - move to absolute position
  - `rows_count` (integer, optional) - number of consecutive rows to move starting from `row_id`
  - `revision` (string, optional) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__move_rows

#### Add Columns
- **Endpoint:** `POST /v1/grids/{idx}/columns`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `columns` (NewColumnSchema[]) - column definitions
  - `position` (integer, optional) - insertion position
  - `revision` (string, optional) - current revision (optimistic locking)
- **NewColumnSchema fields:**
  - required: `slug` (string), `title` (string), `type` (enum: `string`, `number`, `date`, `select`, `staff`, `checkbox`, `ticket`, `ticket_field`), `required` (boolean)
  - optional:
    - `description` (string)
    - `color` (enum: `blue`, `yellow`, `pink`, `red`, `green`, `mint`, `grey`, `orange`, `magenta`, `purple`, `copper`, `ocean`)
    - `format` (only for `string` columns; enum: `yfm`, `wom`, `plain`)
    - `select_options` (string[]; only for `select`)
    - `multiple` (boolean; only for `select` and `staff`)
    - `mark_rows` (boolean; only for `checkbox`)
    - `ticket_field` (only for `ticket_field`)
    - `width` (integer) and `width_units` (enum: `%`, `px`)
    - `pinned` (enum: `left`, `right`)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__add_columns

#### Remove Columns
- **Endpoint:** `DELETE /v1/grids/{idx}/columns`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `column_slugs` (string[]) - column slugs to delete
  - `revision` (string, optional) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__remove_columns

#### Move Columns
- **Endpoint:** `POST /v1/grids/{idx}/columns/move`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `column_slug` (string) - starting column slug to move
  - `position` (integer) - destination position
  - `columns_count` (integer, optional) - number of consecutive columns to move starting from `column_slug`
  - `revision` (string, optional) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__move_columns

#### Update Cells
- **Endpoint:** `POST /v1/grids/{idx}/cells`
- **Operation:** Write
- **Required parameters:**
  - `idx` (path) - grid ID (UUID string)
- **Request body:**
  - `cells` (UpdateCellSchema[]) - list of cell updates:
    - `row_id` (integer) - row identifier
    - `column_slug` (string)
    - `value` - any of:
      - integer
      - number
      - boolean
      - string
      - string[]
      - UserIdentityExtended[] (objects with `uid`, `cloud_uid`, `username`)
  - `revision` (string, optional) - current revision (optimistic locking)
- **Response body:**
  - `revision` (string)
  - `cells` (CellSchema[]) - updated cells:
    - `row_id` (string)
      - ⚠️ Note: request `row_id` is documented as integer, but response `row_id` is documented as string.
    - `column_slug` (string)
    - `value` - can be primitive types and also complex objects (for example TicketSchema, user objects, Tracker enum objects) depending on column type
- **Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__update_cells

---

### 3. Page Resources (Attachments & Grids)

#### Get Resources
- **Endpoint:** `GET /v1/pages/{idx}/resources`
- **Operation:** Read
- **Required parameters:**
  - `idx` (path) - page ID (integer)
- **Optional query parameters:**
  - `cursor` - cursor for pagination
  - `order_by` - sort field (values: `name_title`, `created_at`)
  - `order_direction` - sort direction (values: `asc`, `desc`; default: `asc`)
  - `page_id` (integer) - page number (default: 1)
  - `page_size` (integer) - results per page (default: 25, min: 1, max: 50)
  - `q` (string) - search by title (max 255 chars)
  - `types` (string) - resource types to return, comma-separated (values: `attachment`, `sharepoint_resource`, `grid`)
- **Response body:**
  - `results` - array of Resource objects
  - `next_cursor`, `prev_cursor` - cursors for navigating pages
- **Resource object:**
  - `type` (enum: `attachment`, `grid`, `sharepoint_resource`)
  - `item` - one of:
    - AttachmentSchema (`id`, `name`, `download_url`, `size`, `description`, `user`, `created_at`, `mimetype`, `has_preview`)
    - PageGridsSchema (`id` UUID, `title`, `created_at`)
    - PageSharepointSchema (`id` UUID, `title`, `doctype`, `created_at`)
- **Source:** https://yandex.ru/support/wiki/en/api-ref/pagesresources/pagesresources__resources

---

## Column Type Details (Grids)

Dynamic tables use `ColumnType` and related enums (see schema definitions in the Create Grid API reference).

### ColumnType
- `string`
- `number`
- `date`
- `select`
- `staff`
- `checkbox`
- `ticket`
- `ticket_field`

### TicketField (for `ticket_field` columns)
`TicketField` is an enum of Tracker fields that can be referenced via `ticket_field`, including:
- `assignee`, `components`, `created_at`, `deadline`, `description`, `end`, `estimation`, `fixversions`, `followers`, `last_comment_updated_at`, `original_estimation`, `parent`, `pending_reply_from`, `priority`, `project`, `queue`, `reporter`, `resolution`, `resolved_at`, `sprint`, `start`, `status`, `status_start_time`, `status_type`, `storypoints`, `subject`, `tags`, `type`, `updated_at`, `votes`

### Common column fields
From `ColumnSchema` (responses) and `NewColumnSchema` (create/add-column requests):
- required: `slug` (string), `title` (string), `type` (ColumnType), `required` (boolean)
- optional:
  - `description` (string)
  - `color` (BGColor enum)
  - `format` (TextFormat enum; only for `string` columns)
  - `select_options` (string[]; only for `select`)
  - `multiple` (boolean; only for `select` and `staff`)
  - `mark_rows` (boolean; only for `checkbox`)
  - `ticket_field` (TicketField enum; only for `ticket_field`)
  - `width` (integer) and `width_units` (enum: `%`, `px`)
  - `pinned` (enum: `left`, `right`)

**Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__create_grid

---

## Page Structure Details

### Page type
`PageType` enum values:
- `page`
- `grid`
- `cloud_page`
- `wysiwyg`
- `template`

**Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details_by_id

### Page Attributes (returned when `fields` includes `attributes`)
`PageAttributesSchema` fields:
- `created_at` (string, date-time)
- `modified_at` (string, date-time)
- `lang` (string)
- `is_readonly` (boolean)
- `comments_count` (integer)
- `comments_enabled` (boolean)
- `keywords` (string[])
- `is_collaborative` (boolean)
- `is_draft` (boolean)

**Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details_by_id

### Page Content (returned when `fields` includes `content`)
The `content` field is documented as a union and depends on `page_type`:
- `page`, `wysiwyg`: string
- `cloud_page`: CloudPageContentSchema
  - `embed` (CloudPageEmbeddingSchema): `iframe_src`, `edit_src`
  - `acl_management` (enum: `unknown`, `unmanaged`, `wiki`)
  - `type` (enum: `docx`, `pptx`, `xlsx`)
  - `filename` (string)
  - `error` (string, optional)
- `grid`: legacy grid content schema
  - API reference lists `LegacyGridSchema` as one of the possible `content` shapes
  - ⚠️ The docs also note this is legacy and "for the Wiki frontend it is always null"

**Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details_by_id

### Breadcrumbs (returned when `fields` includes `breadcrumbs`)
`BreadcrumbSchema[]` items contain:
- `id` (integer)
- `title` (string)
- `slug` (string)
- `page_exists` (boolean)

**Source:** https://yandex.ru/support/wiki/en/api-ref/pages/pages__get_page_details_by_id

---

## User Permissions (Page/Grid)

Returned when `fields` includes `user_permissions`:
- Permission enum values: `create_page`, `delete`, `edit`, `view`, `comment`, `change_authors`, `change_acl`, `set_redirect`, `manage_invite`, `view_invite`, `admin`

**Source:** https://yandex.ru/support/wiki/en/api-ref/grids/grids__create_grid

---

## Stated Constraints & Limitations

### Authentication
- Requests must include an authorization header:
  - `Authorization: OAuth <OAuth_token>`
  - `Authorization: Bearer <IAM_token>`
- Requests must include an organization header:
  - `X-Org-Id` for Yandex 360 for Business
  - `X-Cloud-Org-Id` for Yandex Cloud organizations
- Yandex Cloud service accounts cannot be used for Yandex Wiki API authorization (user accounts only)
- IAM token validity: no more than 12 hours; expired token yields `401 Unauthorized`

**Source:** https://yandex.ru/support/wiki/en/api-ref/access

### Pagination
For `GET /v1/pages/{idx}/grids` and `GET /v1/pages/{idx}/resources`:
- `page_size` default: 25
- `page_size` min: 1
- `page_size` max: 50
- cursor pagination uses `next_cursor` and `prev_cursor`
- legacy page-based pagination uses `page_id` (request parameter)
- `/v1/pages/{idx}/grids` additionally returns `has_next` and `page_id` for backward compatibility
  - if `cursor` is used, `page_id` is documented as always `1` and clients should rely on `..._cursor`

**Sources:**
- https://yandex.ru/support/wiki/en/api-ref/pages/pages__page_grids
- https://yandex.ru/support/wiki/en/api-ref/pagesresources/pagesresources__resources

### Page Operations
- Title length: 1-255 characters (create/clone)
- Concurrent edits (Update Page): `allow_merge` query parameter
  - if `allow_merge=true`, concurrent edits are merged with a 3-way merge algorithm; otherwise a conflict is returned
- Clone Page (async) validation error codes include:
  - `IS_CLOUD_PAGE`, `SLUG_OCCUPIED`, `SLUG_RESERVED`, `FORBIDDEN`, `QUOTA_EXCEEDED`, `CLUSTER_BLOCKED`
- MS365 upload flow when creating a `cloud_page` (3-step):
  - step 1: `cloud_page.method=upload_doc` (returns `upload_to` + `upload_session`)
  - step 2: upload file via `PUT` to `upload_to`
  - step 3: `cloud_page.method=finalize_upload` with `upload_session`

**Sources:**
- https://yandex.ru/support/wiki/en/api-ref/pages/pages__create_public_page
- https://yandex.ru/support/wiki/en/api-ref/pages/pages__update_public_page_details
- https://yandex.ru/support/wiki/en/api-ref/pages/pages__clone_page

### Grid Operations
- Most grid endpoints use `idx` as a UUID string (`string<uuid4>` path parameter)
  - ⚠️ The Create Grid response documents `id` as "any of" UUID4 string or integer
- Many write operations accept a `revision` string for optimistic locking (for example: update grid, add/remove/move rows/columns, update cells)

**Sources:**
- https://yandex.ru/support/wiki/en/api-ref/grids/grids__create_grid
- https://yandex.ru/support/wiki/en/api-ref/grids/grids__update_grid

### UNCLEAR / Not Explicitly Documented
The API reference pages above do not explicitly document:
- rate limits per token/org
- maximum page content size
- maximum number of rows/columns per grid
- batch operation limits
