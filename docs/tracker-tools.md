# Yandex Tracker MCP tools

This document lists the MCP tools implemented under `internal/tools/tracker/`.

Tool names and one-line descriptions are taken from tool registration in `internal/tools/tracker/service.go`.
Input/output schemas are derived from the MCP handler layer DTOs in `internal/tools/tracker/dto.go`.

## Conventions

- Types are described using JSON-compatible terms (string, number/integer, boolean, array, object).
- “Required” means the tool validates the parameter as required (and/or marks it required in the schema).
- Timestamp fields are strings as returned by the upstream Yandex Tracker API.

## tracker_issue_get

Retrieves a Yandex Tracker issue by its ID or key.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `expand` (string, optional): Additional fields to include.
  - Allowed values: `attachments`

### Output

Returns `IssueOutput`:

- `self` (string)
- `id` (string)
- `key` (string)
- `version` (integer)
- `summary` (string)
- `description` (string, optional)
- `status_start_time` (string, optional)
- `created_at` (string, optional)
- `updated_at` (string, optional)
- `resolved_at` (string, optional)
- `status` (object, optional): `StatusOutput`
  - `self` (string)
  - `id` (string)
  - `key` (string)
  - `display` (string)
- `type` (object, optional): `TypeOutput`
  - `self` (string)
  - `id` (string)
  - `key` (string)
  - `display` (string)
- `priority` (object, optional): `PriorityOutput`
  - `self` (string)
  - `id` (string)
  - `key` (string)
  - `display` (string)
- `queue` (object, optional): `QueueOutput`
  - `self` (string)
  - `id` (string)
  - `key` (string)
  - `display` (string, optional)
  - `name` (string, optional)
  - `version` (integer, optional)
  - `lead` (object, optional): `UserOutput`
  - `assign_auto` (boolean, optional)
  - `allow_externals` (boolean, optional)
  - `deny_voting` (boolean, optional)
- `assignee` (object, optional): `UserOutput`
- `created_by` (object, optional): `UserOutput`
- `updated_by` (object, optional): `UserOutput`
- `votes` (integer, optional)
- `favorite` (boolean, optional)

`UserOutput`:

- `self` (string)
- `id` (string)
- `uid` (string, optional)
- `login` (string, optional)
- `display` (string, optional)
- `first_name` (string, optional)
- `last_name` (string, optional)
- `email` (string, optional)
- `cloud_uid` (string, optional)
- `passport_uid` (string, optional)

## tracker_issue_search

Searches Yandex Tracker issues using filter or query.

### Input

- `filter` (object, optional): Field-based filter object with key-value pairs.
  - Note: the tool requires all filter values to be strings.
- `query` (string, optional): Query language filter string (Yandex Tracker query syntax).
- `order` (string, optional): Sorting direction and field.
  - Format: `+<field_key>` or `-<field_key>`
  - Note: only used together with `filter`, not with `query`.
- `expand` (string, optional): Additional fields to include.
  - Allowed values: `transitions`, `attachments`
- `per_page` (integer, optional): Number of results per page (standard pagination).
  - Tool validation: Valid range: 1-50
- `page` (integer, optional): Page number (standard pagination).
  - Tool validation: must be non-negative.
- `scroll_type` (string, optional): Scroll type for large result sets.
  - Allowed values: `sorted`, `unsorted`
  - Note: used only in the first request of a scroll sequence.
- `per_scroll` (integer, optional): Max issues per scroll response.
  - Tool validation: Valid range: 1-1000
- `scroll_ttl_millis` (integer, optional): Scroll context lifetime in milliseconds.
  - Tool validation: Default: 60000, maximum: 600000
- `scroll_id` (string, optional): Scroll page ID for 2nd and subsequent scroll requests.

### Output

Returns `SearchIssuesOutput`:

- `issues` (array of object): array of `IssueOutput`
- `total_count` (integer)
- `total_pages` (integer)
- `scroll_id` (string, optional)
- `scroll_token` (string, optional)
- `next_link` (string, optional)

## tracker_issue_count

Counts Yandex Tracker issues matching filter or query.

### Input

- `filter` (object, optional): Field-based filter object.
  - Note: the tool requires all filter values to be strings.
- `query` (string, optional): Query language filter string.

### Output

Returns `CountIssuesOutput`:

- `count` (integer)

## tracker_issue_transitions_list

Lists available status transitions for a Yandex Tracker issue.

### Input

- `issue_id_or_key` (string, required): Issue ID or key.

### Output

Returns `TransitionsListOutput`:

- `transitions` (array of object): array of `TransitionOutput`
  - `id` (string)
  - `display` (string)
  - `self` (string)
  - `to` (object, optional): `StatusOutput`

## tracker_queues_list

Lists Yandex Tracker queues.

### Input

- `expand` (string, optional): Additional fields to include.
  - Allowed values: `projects`, `components`, `versions`, `types`, `team`, `workflows`, `all`
- `per_page` (integer, optional): Number of queues per page.
  - Tool validation: must be non-negative.
- `page` (integer, optional): Page number.
  - Tool validation: must be non-negative.

### Output

Returns `QueuesListOutput`:

- `queues` (array of object): array of `QueueOutput`
- `total_count` (integer)
- `total_pages` (integer)

`QueueOutput`:

- `self` (string)
- `id` (string)
- `key` (string)
- `display` (string, optional)
- `name` (string, optional)
- `version` (integer, optional)
- `lead` (object, optional): `UserOutput`
- `assign_auto` (boolean, optional)
- `allow_externals` (boolean, optional)
- `deny_voting` (boolean, optional)

## tracker_issue_comments_list

Lists comments for a Yandex Tracker issue.

### Input

- `issue_id_or_key` (string, required): Issue ID or key.
- `expand` (string, optional): Additional fields to include.
  - Allowed values: `attachments`, `html`, `all`
- `per_page` (integer, optional): Number of comments per page.
  - Tool validation: must be non-negative.
- `id` (string, optional): Comment id value after which the requested page will begin (for pagination).

### Output

Returns `CommentsListOutput`:

- `comments` (array of object): array of `CommentOutput`
- `next_link` (string, optional)

`CommentOutput`:

- `id` (string)
- `long_id` (string)
- `self` (string)
- `text` (string)
- `version` (integer)
- `type` (string, optional)
- `transport` (string, optional)
- `created_at` (string, optional)
- `updated_at` (string, optional)
- `created_by` (object, optional): `UserOutput`
- `updated_by` (object, optional): `UserOutput`







## tracker_issue_attachments_list

Lists attachments for a Yandex Tracker issue.

This tool is read-only.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).

### Output

Returns `AttachmentsListOutput`:

- `attachments` (array of object): array of `AttachmentOutput`

`AttachmentOutput`:

- `id` (string)
- `name` (string)
- `content_url` (string)
- `thumbnail_url` (string, optional)
- `mimetype` (string, optional)
- `size` (integer)
- `created_at` (string, optional)
- `created_by` (object, optional): `UserOutput`
- `metadata` (object, optional): `AttachmentMetadataOutput`

`AttachmentMetadataOutput`:

- `size` (string, optional)

## tracker_issue_attachment_get

Downloads a file attached to a Yandex Tracker issue. Either saves it to the local workspace or returns text content.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `attachment_id` (string, required): Attachment ID (for example, `4159`).
- `file_name` (string, required): Attachment file name (for example, `attachment.txt`).
- `save_path` (string, optional): Absolute path to save the attachment (for example, `/Users/me/attachments/attachment.txt`).
- `get_content` (boolean, optional): Return text content in output when `true`.
- `override` (boolean, optional): Overwrite existing file if `true` (default: `false`).

Requirement:

- Exactly one of `save_path` or `get_content` must be provided.

Notes:

- The extension is validated against an allowlist. By default: txt, json, jsonc, yaml, yml, md, pdf, doc, docx, rtf, odt, xls, xlsx, ods, csv, tsv, ppt, pptx, odp, jpg, jpeg, png, tiff, tif, gif, bmp, webp, zip, 7z, tar, tgz, tar.gz, gz, bz2, xz, rar.
- The extension is validated using the `save_path` file name.
- The allowlist can be replaced via `YANDEX_MCP_ATTACH_EXT` (comma-separated, without dots).
- By default, `save_path` must be inside the user home directory, must not point to the home root, and must not be within a hidden top-level home subdirectory (for example, `~/.ssh`).
- The directory restriction can be fully replaced via `YANDEX_MCP_ATTACH_DIR` (comma-separated absolute paths). When it is set, only the provided directories (and their subdirectories) are allowed.
- When `save_path` is used, the attachment is streamed to disk and not fully loaded into memory.
- When `get_content` is used, the file extension is validated against the text allowlist. By default: txt, json, jsonc, yaml, yml, md, csv, tsv, rtf.
- The extension for `get_content` is validated using the `file_name` value.
- The text allowlist can be replaced via `YANDEX_MCP_ATTACH_VIEW_EXT` (comma-separated, without dots).
- When `get_content` is used, inline content is limited by `YANDEX_MCP_ATTACH_INLINE_MAX_BYTES` (default: 10485760). Larger payloads are rejected.
- Paths are validated after resolving symlinks; the resolved path must remain within the allowed directory scope.

### Output

Returns `AttachmentContentOutput`:

- `file_name` (string, optional)
- `content_type` (string, optional)
- `saved_path` (string, optional): Path where the attachment was saved (relative to workspace).
- `content` (string, optional): Attachment text content (raw bytes interpreted as UTF-8) when `get_content` is `true`.
- `size` (integer): Attachment size in bytes.

## tracker_issue_attachment_preview_get

Downloads a thumbnail for an attachment in a Yandex Tracker issue and saves it to the local workspace.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `attachment_id` (string, required): Attachment ID (for example, `4159`).
- `save_path` (string, required): Absolute path to save the attachment preview (for example, `/Users/me/attachments/preview.png`).
- `override` (boolean, optional): Overwrite existing file if `true` (default: `false`).

Notes:

- The same extension and directory rules as `tracker_issue_attachment_get` apply to the preview.

### Output

Returns `AttachmentContentOutput` (same shape as `tracker_issue_attachment_get`).


## tracker_queue_get

Gets a Yandex Tracker queue by ID or key.

### Input

- `queue_id_or_key` (string, required): Queue ID or key (for example, `MYQUEUE`).
- `expand` (string, optional): Additional fields to include in the response.
  - Allowed values: `projects`, `components`, `versions`, `types`, `team`, `workflows`, `all`

### Output

Returns `QueueDetailOutput`:

- `self` (string)
- `id` (string)
- `key` (string)
- `display` (string, optional)
- `name` (string, optional)
- `description` (string, optional)
- `version` (integer, optional)
- `lead` (object, optional): `UserOutput`
- `assign_auto` (boolean, optional)
- `allow_externals` (boolean, optional)
- `deny_voting` (boolean, optional)
- `default_type` (object, optional): `TypeOutput`
- `default_priority` (object, optional): `PriorityOutput`




## tracker_user_current

Gets the current authenticated Yandex Tracker user.

### Input

No input.

### Output

Returns `UserDetailOutput`:

- `self` (string)
- `id` (string)
- `uid` (string, optional)
- `tracker_uid` (string, optional)
- `login` (string, optional)
- `display` (string, optional)
- `first_name` (string, optional)
- `last_name` (string, optional)
- `email` (string, optional)
- `cloud_uid` (string, optional)
- `passport_uid` (string, optional)
- `has_license` (boolean, optional)
- `dismissed` (boolean, optional)
- `external` (boolean, optional)

## tracker_users_list

Lists Yandex Tracker users.

### Input

- `per_page` (integer, optional): Number of users per page (default: 50).
  - Tool validation: must be non-negative.
- `page` (integer, optional): Page number (default: 1).
  - Tool validation: must be non-negative.

### Output

Returns `UsersListOutput`:

- `users` (array of object): array of `UserDetailOutput`
- `total_count` (integer, optional)
- `total_pages` (integer, optional)

## tracker_user_get

Gets a Yandex Tracker user by ID or login.

### Input

- `user_id` (string, required): User login or ID.

### Output

Returns `UserDetailOutput` (same shape as in `tracker_user_current`).

## tracker_issue_links_list

Lists all links for a Yandex Tracker issue.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).

### Output

Returns `LinksListOutput`:

- `links` (array of object): array of `LinkOutput`

`LinkOutput`:

- `id` (string)
- `self` (string)
- `type` (object, optional): `LinkTypeOutput`
- `direction` (string, optional)
  - Documented values: `inward`, `outward`
- `object` (object, optional): `LinkedIssueOutput`
- `created_by` (object, optional): `UserOutput`
- `updated_by` (object, optional): `UserOutput`
- `created_at` (string, optional)
- `updated_at` (string, optional)

`LinkTypeOutput`:

- `id` (string)
- `inward` (string, optional)
- `outward` (string, optional)

`LinkedIssueOutput`:

- `self` (string)
- `id` (string)
- `key` (string)
- `display` (string, optional)



## tracker_issue_changelog

Gets the changelog for a Yandex Tracker issue.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `per_page` (integer, optional): Number of changelog entries per page (default: 50).
  - Tool validation: must be non-negative.

### Output

Returns `ChangelogOutput`:

- `entries` (array of object): array of `ChangelogEntryOutput`

`ChangelogEntryOutput`:

- `id` (string)
- `self` (string)
- `issue` (object, optional): `LinkedIssueOutput`
- `updated_at` (string, optional)
- `updated_by` (object, optional): `UserOutput`
- `type` (string, optional)
  - Documented values: `IssueCreated`, `IssueUpdated`, `IssueWorkflow`
- `transport` (string, optional)
- `fields` (array of object, optional): array of `ChangelogFieldOutput`

`ChangelogFieldOutput`:

- `field` (string)
- `from` (any, optional)
- `to` (any, optional)


## tracker_project_comments_list

Lists comments for a Yandex Tracker project entity.

### Input

- `project_id` (string, required): Project ID or short ID.
- `expand` (string, optional): Additional fields to include.
  - Allowed values: `all`, `html`, `attachments`, `reactions`

### Output

Returns `ProjectCommentsListOutput`:

- `comments` (array of object): array of `ProjectCommentOutput`

`ProjectCommentOutput`:

- `id` (string)
- `long_id` (string, optional)
- `self` (string)
- `text` (string, optional)
- `created_at` (string, optional)
- `updated_at` (string, optional)
- `created_by` (object, optional): `UserOutput`
- `updated_by` (object, optional): `UserOutput`
