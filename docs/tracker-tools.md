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
- `uid` (integer, optional)
- `login` (string, optional)
- `display` (string, optional)
- `first_name` (string, optional)
- `last_name` (string, optional)
- `email` (string, optional)
- `cloud_uid` (string, optional)
- `passport_uid` (integer, optional)

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
- `id` (integer, optional): Comment numeric id value after which the requested page will begin (for pagination).
  - Tool validation: must be non-negative.

### Output

Returns `CommentsListOutput`:

- `comments` (array of object): array of `CommentOutput`
- `next_link` (string, optional)

`CommentOutput`:

- `id` (integer)
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
- `version` (integer, optional): Issue version for optimistic locking.

Note: the tool requires at least one of `summary`, `description`, `type`, `priority`, `assignee` to be provided.

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
- `markup_type` (string, optional): Text markup type.
  - Allowed values: `plain`, `wiki`, `html`
- `summonees` (array of string, optional): User logins to summon.
- `maillist_summonees` (array of string, optional): Mailing list addresses to summon.
- `is_add_to_followers` (boolean, optional): Add summoned users to followers.

### Output

Returns `CommentOutput` (same shape as in `tracker_issue_comments_list`).
