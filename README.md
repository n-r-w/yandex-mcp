# Yandex MCP Server

An MCP (Model Context Protocol) server that lets MCP-capable clients work with:

- Yandex Tracker (issues, queues, transitions, comments)
- Yandex Wiki (pages, attachments/resources, dynamic tables)

The project is not an official MCP from Yandex.

## Development Status

- This project is in early development.
- Read operations have been tested against production instances of Yandex Tracker and Yandex Wiki.
- The server operates in read-only mode.

## Tools

For the full tool list and a parameter overview, see:

- [docs/tracker-tools.md](docs/tracker-tools.md)
- [docs/wiki-tools.md](docs/wiki-tools.md)

Exact JSON schemas (including validation rules) are also available via MCP tool introspection at runtime.

### Yandex Wiki tools

- `wiki_page_get` — Retrieves a Yandex Wiki page by its slug (URL path)
- `wiki_page_get_by_id` — Retrieves a Yandex Wiki page by its numeric ID
- `wiki_page_resources_list` — Lists resources (attachments, grids) for a Yandex Wiki page
- `wiki_page_grids_list` — Lists dynamic tables (grids) for a Yandex Wiki page
- `wiki_grid_get` — Retrieves a Yandex Wiki dynamic table (grid) by its ID

### Yandex Tracker tools

- `tracker_issue_get` — Retrieves a Yandex Tracker issue by its ID or key
- `tracker_issue_search` — Searches Yandex Tracker issues using filter or query
- `tracker_issue_count` — Counts Yandex Tracker issues matching filter or query
- `tracker_issue_transitions_list` — Lists available status transitions for a Yandex Tracker issue
- `tracker_queues_list` — Lists Yandex Tracker queues
- `tracker_issue_comments_list` — Lists comments for a Yandex Tracker issue
- `tracker_issue_attachments_list` — Lists attachments for a Yandex Tracker issue
- `tracker_queue_get` — Retrieves a Yandex Tracker queue by its key
- `tracker_user_current` — Retrieves the current Yandex Tracker user
- `tracker_users_list` — Lists Yandex Tracker users
- `tracker_user_get` — Retrieves a Yandex Tracker user by ID
- `tracker_issue_links_list` — Lists links for a Yandex Tracker issue
- `tracker_issue_changelog` — Retrieves the changelog for a Yandex Tracker issue
- `tracker_project_comments_list` — Lists comments for a Yandex Tracker project entity

## Installation

### Binary Releases

Pre-compiled binaries are available for multiple platforms:

- **Linux (AMD64)**: `yandex-mcp-v*-linux-amd64.tar.gz`
- **macOS (Intel)**: `yandex-mcp-v*-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `yandex-mcp-v*-darwin-arm64.tar.gz`
- **Windows (AMD64)**: `yandex-mcp-v*-windows-amd64.zip`

Download the latest release from [GitHub Releases](https://github.com/n-r-w/yandex-mcp/releases).

### Build from Source

```bash
go build -o yandex-mcp ./cmd/yandex-mcp
```

or use Task:

```bash
task build
```

### macOS Installation Notes

macOS may block execution of downloaded binaries by default due to security settings. To allow the executable to run:

1. **First execution attempt**: Run the executable from terminal
   ```bash
   ./yandex-mcp --version
   ```
   This will show a security warning. Press **Done**.

2. **Allow execution via System Settings**:
   - Open **System Settings** → **Privacy & Security** → **Security**
   - Find the message about the blocked executable
   - Click **"Allow Anyway"**

3. **Second execution**: Run the executable again
   ```bash
   ./yandex-mcp --version
   ```

4. **Confirm execution**: A dialog will appear asking for confirmation
   - Click **"Open Anyway"** and enter your password if prompted
   - The executable will now be allowed to run

After these steps, the executable will be permanently allowed to run on your system.

## Environment variables

- `YANDEX_CLOUD_ORG_ID` (required)
  * Yandex Cloud Organization ID.
  * Used to set the organization header required by Yandex APIs.
  * Run `yc organization-manager organization list` to get your organization ID.

- `YANDEX_WIKI_BASE_URL` (optional, default: `https://api.wiki.yandex.net`)
  * Base URL for Yandex Wiki API.
  * Must be an `https://` URL.

- `YANDEX_TRACKER_BASE_URL` (optional, default: `https://api.tracker.yandex.net`)
  * Base URL for Yandex Tracker API.
  * Must be an `https://` URL.

- `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (optional, default: `10`)
  * IAM token refresh period in **hours**.
  * The server caches the token and refreshes it when the cached token is older than this period.
  * IAM tokens are valid for **no more than 12 hours**; this refresh period should not exceed `12`.

- `YANDEX_API_HTTP_TIMEOUT` (optional, default: `30s`)
  * HTTP timeout for Yandex API requests.
  * Accepts Go duration strings (e.g., `30s`, `1m`).

## Authentication

The project supports IAM token authentication via the Yandex Cloud CLI (`yc`) only.

**IAM token acquisition (`yc` prerequisites)**

Installation: https://yandex.cloud/en/docs/cli/operations/install-cli

This server obtains IAM tokens by running:
- `yc iam create-token`

That means:

- You must have the **Yandex Cloud CLI** (`yc`) installed and available in `PATH`.
- You must have an initialized/authenticated `yc` profile (typically via `yc init`).

Notes:

- Yandex IAM tokens are valid for **no more than 12 hours**, so long-running use requires periodic refresh.
- The server refreshes the token periodically based on `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (by default every **10 hours**; you can set it to `12` to refresh roughly every 12 hours).
- When the refresh happens, the server calls `yc iam create-token` again. If your `yc` session/profile requires interactive authentication, `yc` may open your **default browser** and ask you to log in.

Official references:

- Tracker IAM token auth + lifetime: https://yandex.ru/support/tracker/en/concepts/access#iam-token
- Wiki IAM token auth + lifetime: https://yandex.ru/support/wiki/en/api-ref/access#iam-token

## Client configuration examples

### Claude Code

```bash
claude mcp add -s user -e YANDEX_CLOUD_ORG_ID={yandex organization id} --transport stdio yandex /path/to/yandex-mcp
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