# Yandex MCP Server

An MCP (Model Context Protocol) server that lets MCP-capable clients work with:

- Yandex Tracker (issues, queues, transitions, comments)
- Yandex Wiki (pages, attachments/resources, dynamic tables)

The project is not an official MCP from Yandex.

## Development Status

- This project is in early development. 
- Read operations are tested and working on production Yandex Tracker and Wiki instances.
- Write operations are not tested on real production environments, i'm not that brave :)

## Tools

For full parameter and schema documentation, see:

- [docs/tracker-tools.md](docs/tracker-tools.md)
- [docs/wiki-tools.md](docs/wiki-tools.md)

### Yandex Wiki tools

Enabled by default (read-only):

- `wiki_page_get` — Retrieves a Yandex Wiki page by its slug (URL path)
- `wiki_page_get_by_id` — Retrieves a Yandex Wiki page by its numeric ID
- `wiki_page_resources_list` — Lists resources (attachments, grids) for a Yandex Wiki page
- `wiki_page_grids_list` — Lists dynamic tables (grids) for a Yandex Wiki page
- `wiki_grid_get` — Retrieves a Yandex Wiki dynamic table (grid) by its ID

Require `--wiki-write` (write operations):

- `wiki_page_create` — Creates a new Yandex Wiki page
- `wiki_page_update` — Updates an existing Yandex Wiki page
- `wiki_page_append_content` — Appends content to an existing Yandex Wiki page
- `wiki_grid_create` — Creates a new Yandex Wiki dynamic table (grid)
- `wiki_grid_update_cells` — Updates cells in a Yandex Wiki dynamic table (grid)

### Yandex Tracker tools

Enabled by default (read-only):

- `tracker_issue_get` — Retrieves a Yandex Tracker issue by its ID or key
- `tracker_issue_search` — Searches Yandex Tracker issues using filter or query
- `tracker_issue_count` — Counts Yandex Tracker issues matching filter or query
- `tracker_issue_transitions_list` — Lists available status transitions for a Yandex Tracker issue
- `tracker_queues_list` — Lists Yandex Tracker queues
- `tracker_issue_comments_list` — Lists comments for a Yandex Tracker issue

Require `--tracker-write` (write operations):

- `tracker_issue_create` — Creates a new Yandex Tracker issue
- `tracker_issue_update` — Updates an existing Yandex Tracker issue
- `tracker_issue_transition_execute` — Executes a status transition on a Yandex Tracker issue
- `tracker_issue_comment_add` — Adds a comment to a Yandex Tracker issue

## Client configuration examples

### Claude Code

```bash
claude mcp add -s user -e YANDEX_CLOUD_ORG_ID={yandex organization id} --transport stdio standards /path/to/yandex-mcp
```

### VS Code, RooCode, etc.

```json
"yandex": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_CLOUD_ORG_ID": "yandex organization id"
  }
}
```

Notes:

- The `command` must point to the built executable (for this repo, `task build` produces `bin/yandex-mcp`).
- The server communicates over stdio; clients should use a stdio transport.

## CLI arguments

- `--wiki-write` (default: `false`) — enable write operations for Yandex Wiki tools
- `--tracker-write` (default: `false`) — enable write operations for Yandex Tracker tools

## Environment variables

- `YANDEX_CLOUD_ORG_ID` (required)
  - Yandex Cloud Organization ID.
  - Used to set the organization header required by Yandex APIs.

- `YANDEX_WIKI_BASE_URL` (optional, default: `https://api.wiki.yandex.net`)
  - Base URL for Yandex Wiki API.
  - Must be an `https://` URL.

- `YANDEX_TRACKER_BASE_URL` (optional, default: `https://api.tracker.yandex.net`)
  - Base URL for Yandex Tracker API.
  - Must be an `https://` URL.

- `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (optional, default: `10`)
  - IAM token refresh period in **hours**.
  - The server caches the token and refreshes it when the cached token is older than this period.
  - IAM tokens are valid for **no more than 12 hours**, so set this value to `12` or lower.

## Authentication

The project supports only one authentication method - via IAM token and the Yandex Cloud CLI (`yc`).

**IAM token acquisition (`yc` prerequisites)**

Installation: https://yandex.cloud/en/docs/cli/operations/install-cli

This server obtains IAM tokens by running:
- `yc iam create-token`

That means:

- You must have the **Yandex Cloud CLI** (`yc`) installed and available in `PATH`.
- You must have an initialized/authenticated `yc` profile (typically via `yc init`).

Important behavior to be aware of:

- Yandex IAM tokens are valid for **no more than 12 hours**, so you should expect token refresh to happen **at least once every 12 hours**.
- The server refreshes the token periodically based on `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (by default every **10 hours**; you can set it to `12` to refresh roughly every 12 hours).
- When the refresh happens, the server calls `yc iam create-token` again. If your `yc` session/profile requires interactive authentication, `yc` may open your **default browser** and ask you to log in.

Official references:

- Tracker IAM token auth + lifetime: https://yandex.ru/support/tracker/en/concepts/access#iam-token
- Wiki IAM token auth + lifetime: https://yandex.ru/support/wiki/en/api-ref/access#iam-token

## Yandex API reference (official)

Yandex Tracker:

- API overview: https://yandex.ru/support/tracker/en/about-api
- API access (OAuth / IAM): https://yandex.ru/support/tracker/en/concepts/access
- Common request format: https://yandex.ru/support/tracker/en/common-format
- Error codes: https://yandex.ru/support/tracker/en/error-codes

Yandex Wiki:

- API overview: https://yandex.ru/support/wiki/en/api-ref/about
- API access (OAuth / IAM): https://yandex.ru/support/wiki/en/api-ref/access
- API reference index: https://yandex.ru/support/wiki/en/api-ref/

IAM token (Yandex Cloud):

- Tracker: IAM token section (mentions 12-hour max lifetime): https://yandex.ru/support/tracker/en/concepts/access#iam-token
- Wiki: IAM token section (mentions 12-hour max lifetime): https://yandex.ru/support/wiki/en/api-ref/access#iam-token