# Yandex Tracker MCP tools

This document lists the MCP tools implemented under `internal/tools/tracker/`.

Tool names and one-line descriptions are taken from tool registration in `internal/tools/tracker/service.go`.
Input/output schemas are derived from the MCP handler layer DTOs in `internal/tools/tracker/dto.go`.

## Write operations gating

By default, the server registers only read-only Tracker tools.

To enable Tracker write tools, start the server with:

- `--tracker-write` (default: false)

Write-gated tools in this document are explicitly marked.

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
  - Tool validation: must be non-negative.
- `page` (integer, optional): Page number (standard pagination).
  - Tool validation: must be non-negative.
- `scroll_type` (string, optional): Scroll type for large result sets.
  - Allowed values: `sorted`, `unsorted`
  - Note: used only in the first request of a scroll sequence.
- `per_scroll` (integer, optional): Max issues per scroll response.
  - Tool validation: must be non-negative and must not exceed 1000.
- `scroll_ttl_millis` (integer, optional): Scroll context lifetime in milliseconds.
  - Tool validation: must be non-negative.
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
  - Allowed values: `projects`, `components`, `versions`, `types`, `team`, `workflows`
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

## tracker_issue_create

Creates a new Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `queue` (string, required): Queue key.
- `summary` (string, required): Issue summary.
- `description` (string, optional): Issue description.
- `type` (string, optional): Issue type key.
- `priority` (string, optional): Priority key.
- `assignee` (string, optional): Assignee login.
- `tags` (array of string, optional): Issue tags.
- `parent` (string, optional): Parent issue key.
- `attachment_ids` (array of string, optional): Attachment IDs to link.
- `sprint` (array of string, optional): Sprint IDs to add issue to.

### Output

Returns `IssueOutput`.

## tracker_issue_update

Updates an existing Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key.
- `summary` (string, optional): Issue summary.
- `description` (string, optional): Issue description.
- `type` (string, optional): Issue type key.
- `priority` (string, optional): Priority key.
- `assignee` (string, optional): Assignee login.
- `project_primary` (integer, optional): Primary project ID.
- `project_secondary_add` (array of integer, optional): Secondary project IDs to add.
- `sprint` (array of string, optional): Sprint IDs or keys to assign.
- `version` (integer, optional): Issue version for optimistic locking.

Note: the tool requires at least one of `summary`, `description`, `type`, `priority`, `assignee`, `project_primary`, `project_secondary_add`, `sprint` to be provided.

### Output

Returns `IssueOutput`.

## tracker_issue_transition_execute

Executes a status transition on a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key.
- `transition_id` (string, required): Transition ID.
- `comment` (string, optional): Comment to add during transition.
- `fields` (object, optional): Additional fields to set during transition.

### Output

Returns `TransitionsListOutput`.

## tracker_issue_comment_add

Adds a comment to a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key.
- `text` (string, required): Comment text.
- `attachment_ids` (array of string, optional): Attachment IDs to link.
- `markup_type` (string, optional): Text markup type. Use `md` for YFM markup.
- `summonees` (array of string, optional): User logins to summon.
- `maillist_summonees` (array of string, optional): Mailing list addresses to summon.
- `is_add_to_followers` (boolean, optional): Add summoned users to followers.

### Output

Returns `CommentOutput` (same shape as in `tracker_issue_comments_list`).

## tracker_issue_comment_update

Updates an existing comment on a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `comment_id` (string, required): Comment ID.
- `text` (string, required): Comment text.
- `attachment_ids` (array of string, optional): Attachment IDs to link.
- `markup_type` (string, optional): Text markup type. Use `md` for YFM markup.
- `summonees` (array of string, optional): User logins to summon.
- `maillist_summonees` (array of string, optional): Mailing list addresses to summon.

### Output

Returns `CommentOutput` (same shape as in `tracker_issue_comments_list`).

## tracker_issue_comment_delete

Deletes a comment from a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `comment_id` (string, required): Comment ID.

### Output

Returns `DeleteCommentOutput`:

- `success` (boolean)

## tracker_issue_attachments_list

Lists attachments for a Yandex Tracker issue.

This tool is read-only (available without `--tracker-write`).

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

## tracker_issue_attachment_delete

Deletes an attachment from a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `file_id` (string, required): Attachment file ID.

### Output

Returns `DeleteAttachmentOutput`:

- `success` (boolean)


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

## tracker_queue_create

Creates a new Yandex Tracker queue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `key` (string, required): Queue key (for example, `MYQUEUE`).
- `name` (string, required): Queue name.
- `lead` (string, required): Queue lead login or user ID.
- `default_type` (string, required): Default issue type key or ID.
- `default_priority` (string, required): Default priority key or ID.

### Output

Returns `QueueDetailOutput`.

## tracker_queue_delete

Deletes a Yandex Tracker queue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `queue_id_or_key` (string, required): Queue ID or key (for example, `MYQUEUE`).

### Output

Returns `DeleteQueueOutput`:

- `success` (boolean)

## tracker_queue_restore

Restores a deleted Yandex Tracker queue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `queue_id_or_key` (string, required): Queue ID or key (for example, `MYQUEUE`).

### Output

Returns `QueueDetailOutput`.

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

## tracker_issue_link_create

Creates a link between Yandex Tracker issues.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `relationship` (string, required): Link type ID (for example, `relates`, `depends`, `duplicates`).
- `target_issue` (string, required): Target issue ID or key to link to.

### Output

Returns `LinkOutput` (same shape as in `tracker_issue_links_list`).

## tracker_issue_link_delete

Deletes a link from a Yandex Tracker issue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `link_id` (string, required): Link ID to delete.

### Output

Returns `DeleteLinkOutput`:

- `success` (boolean)

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

## tracker_issue_move

Moves a Yandex Tracker issue to another queue.

This tool is write-gated and requires `--tracker-write`.

### Input

- `issue_id_or_key` (string, required): Issue ID or key (for example, `TEST-1`).
- `queue` (string, required): Target queue key (for example, `NEWQUEUE`).
- `initial_status` (boolean, optional): Reset issue status to initial value when moving.

### Output

Returns `IssueOutput`.

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

## Planned / Requires Additional Research

Items in this section are not implemented and are not available as MCP tools yet.

- Tracker attachment upload via multipart/form-data (upload a file, then link it to an issue comment or issue attachments).
  - Requires: MCP-compatible file transfer approach, request/response schema confirmation, limits, and error mapping.
  - Tentative tool names (not callable): `tracker_attachment_upload`, `tracker_issue_attachment_upload`
- Tracker agile boards and sprints management (read/write operations on boards, board columns, board sprints, sprint details).
  - Requires: request/response schemas, pagination behavior, optimistic locking headers (If-Match / ETag), and deletion semantics.
  - Tentative tool names (not callable): `tracker_boards_list`, `tracker_board_get`, `tracker_board_create`, `tracker_board_update`, `tracker_board_delete`, `tracker_board_columns_list`, `tracker_board_sprints_list`, `tracker_sprint_get`, `tracker_sprint_create`
