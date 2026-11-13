# Yandex MCP Server

An MCP (Model Context Protocol) server that lets MCP-capable clients work with:

- Yandex Tracker (issues, queues, transitions, comments)
- Yandex Wiki (pages, attachments/resources, dynamic tables)

The server operates in read-only mode. Modification operations are not supported due to the risk of allowing LLMs to perform such operations.

The project is not an official MCP from Yandex.

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
- `tracker_boards_list` — Lists Yandex Tracker boards
- `tracker_board_sprints_list` — Lists sprints for a Yandex Tracker board
- `tracker_issue_comments_list` — Lists comments for a Yandex Tracker issue
- `tracker_issue_attachments_list` — Lists attachments for a Yandex Tracker issue
- `tracker_issue_attachment_get` — Downloads a file attached to a Yandex Tracker issue
- `tracker_issue_attachment_preview_get` — Downloads a thumbnail for a Yandex Tracker issue attachment
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

- **Linux (AMD64)**: `yandex-mcp-*-linux-amd64.tar.gz`
- **macOS (Intel)**: `yandex-mcp-*-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `yandex-mcp-*-darwin-arm64.tar.gz`
- **Windows (AMD64)**: `yandex-mcp-*-windows-amd64.tar.gz`

Download the latest release from [GitHub Releases](https://github.com/n-r-w/yandex-mcp/releases).

### Homebrew

```bash
brew install n-r-w/homebrew-tap/yandex-mcp
```

You can also tap first and install by formula name:

```bash
brew tap n-r-w/tap
brew install yandex-mcp
```

### Go install

```bash
go install github.com/n-r-w/yandex-mcp/cmd/yandex-mcp@latest
```

The binary is installed into `GOBIN` or `$GOPATH/bin`.

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

Set exactly one organization ID: `YANDEX_CLOUD_ORG_ID` or `YANDEX_ORG_ID`.

- `YANDEX_CLOUD_ORG_ID`
  * Yandex Cloud Organization ID, sent as `X-Cloud-Org-Id`.
  * Run `yc organization-manager organization list` to get your organization ID.

- `YANDEX_ORG_ID`
  * Yandex 360 for Business organization ID, sent as `X-Org-Id`.
  * Requires `YANDEX_MCP_TOKEN_TYPE=oauth`: IAM tokens are accepted only by Yandex Cloud organizations.

Token acquisition has two independent settings: where the token is acquired and which token type is acquired.

- `YANDEX_MCP_TOKEN_SOURCE` (optional, default: `local`)
  * `local` acquires the token on the machine that runs MCP.
  * `remote` acquires the token from the workstation auth-agent, see [Remote authentication](#remote-authentication).

- `YANDEX_MCP_TOKEN_TYPE` (optional, default: `iam`)
  * `iam` acquires an IAM token through `yc` and sends `Authorization: Bearer <token>`.
  * `oauth` reads a Yandex OAuth token from a file and sends `Authorization: OAuth <token>`, see [OAuth token authentication](#oauth-token-authentication).

| Source | Type | Token comes from | Required settings |
| --- | --- | --- | --- |
| `local` | `iam` | `yc` on the MCP machine | none, `YANDEX_CLI_PROFILE` optional |
| `remote` | `iam` | `yc` on the workstation | `YANDEX_CLI_PROFILE` |
| `local` | `oauth` | file on the MCP machine | `YANDEX_OAUTH_TOKEN_FILE` |
| `remote` | `oauth` | file on the workstation | `YANDEX_OAUTH_PROFILE` |

- `YANDEX_CLI_PROFILE` (optional for local IAM, required for remote IAM)
  * yc CLI profile used to retrieve IAM tokens.
  * In local mode, an empty value makes yc use the active profile.
  * When set, `yc iam create-token --profile <YANDEX_CLI_PROFILE>` runs on the MCP machine in local mode and on the workstation in remote mode.

- `YANDEX_OAUTH_TOKEN_FILE` (required for local OAuth)
  * Absolute path to a file on the MCP machine that contains one Yandex OAuth token.
  * The server reads the file again when `YANDEX_IAM_TOKEN_REFRESH_PERIOD` has passed since the last read or when the API answers `401`/`403`, so a replaced token is used without a restart.

- `YANDEX_OAUTH_PROFILE` (required for remote OAuth)
  * Name of the OAuth profile that auth-agent resolves to a token file on the workstation.
  * The token file stays in the auth-agent OAuth directory on the workstation; MCP sends only the profile name.

- `YANDEX_WIKI_BASE_URL` (optional, default: `https://api.wiki.yandex.net`)
  * Base URL for Yandex Wiki API.
  * Must be an `https://` URL.

- `YANDEX_TRACKER_BASE_URL` (optional, default: `https://api.tracker.yandex.net`)
  * Base URL for Yandex Tracker API.
  * Must be an `https://` URL.

- `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (optional, default: `10`)
  * Token refresh period in **hours**, for both token types.
  * The server caches the token and refreshes it when the cached token is older than this period.
  * IAM tokens are valid for **no more than 12 hours**; this refresh period should not exceed `12`.

- `YANDEX_MCP_TOOL_TIMEOUT`, optional, default `300`
  * Positive seconds for the complete tool call, including authentication and API retries.

- `YANDEX_MCP_ATTACH_EXT` (optional)
  * Comma-separated list of allowed attachment extensions **without dots**.
  * Fully replaces the default allowlist.
  * Default allowlist: txt, json, jsonc, yaml, yml, md, csv, tsv, rtf, html, htm, pdf, doc, docx, odt, xls, xlsx, ods, ppt, pptx, odp, jpg, jpeg, png, tiff, tif, gif, bmp, webp, zip, 7z, tar, tgz, tar.gz, gz, bz2, xz, rar.

- `YANDEX_MCP_ATTACH_VIEW_EXT` (optional)
  * Comma-separated list of allowed attachment extensions **without dots** for inline viewing.
  * Fully replaces the default text allowlist.
  * Default allowlist: txt, json, jsonc, yaml, yml, md, csv, tsv, rtf.

- `YANDEX_MCP_ATTACH_INLINE_MAX_BYTES` (optional, default: `10485760`)
  * Maximum size in bytes for attachment content returned inline.
  * Applies only to `get_content` inline responses; `save_path` uses streaming and is not limited by this setting.

- `YANDEX_MCP_ATTACH_DIR` (optional)
  * Comma-separated list of **absolute** directories allowed for saving attachments.
  * Fully replaces the default directory rules. When set, only the provided directories (and their subdirectories) are allowed.
  * Default rule: `save_path` must be inside the system temporary directory (`os.TempDir()`, for example `/tmp`), or inside the user home directory. In the home directory, `save_path` must not point to the home root and must not be within a hidden top-level home subdirectory (for example, `~/.ssh`).

## Authentication

The project acquires an IAM token through Yandex Cloud CLI `yc` or reads a Yandex OAuth token from a file, selected by `YANDEX_MCP_TOKEN_TYPE`. Either token type is acquired on the MCP machine or, in [remote mode](#remote-authentication), on the workstation, selected by `YANDEX_MCP_TOKEN_SOURCE`. The [OAuth token type](#oauth-token-authentication) needs no `yc` and is the only one that works with a Yandex 360 for Business organization.

**IAM token acquisition (`yc` prerequisites)**

Installation: https://yandex.cloud/en/docs/cli/operations/install-cli

In local mode, the server obtains IAM tokens by running:
- `yc iam create-token` when `YANDEX_CLI_PROFILE` is empty.
- `yc iam create-token --profile <YANDEX_CLI_PROFILE>` when `YANDEX_CLI_PROFILE` is set.

That means:

- You must have the **Yandex Cloud CLI** (`yc`) installed and available in `PATH`.
- You must have an initialized/authenticated `yc` profile (typically via `yc init`).
- Set `YANDEX_CLI_PROFILE` when the server must use a specific yc profile instead of the active profile.

Notes:

- Yandex IAM tokens are valid for **no more than 12 hours**, so long-running use requires periodic refresh.
- The server refreshes the token periodically based on `YANDEX_IAM_TOKEN_REFRESH_PERIOD` (by default every **10 hours**; you can set it to `12` to refresh roughly every 12 hours).
- When the refresh happens, the server calls `yc iam create-token` again. If your `yc` session/profile requires interactive authentication, `yc` may open your **default browser** and ask you to log in.

Official references:

- Tracker IAM token auth + lifetime: https://yandex.ru/support/tracker/en/concepts/access#iam-token
- Wiki IAM token auth + lifetime: https://yandex.ru/support/wiki/en/api-ref/access#iam-token

## OAuth token authentication

The `oauth` token type sends `Authorization: OAuth <token>` and works with both organization types. It combines with either acquisition location: the token file is read on the MCP machine (`local`) or on the workstation through auth-agent (`remote`).

Prerequisites:

- The account is not federated. Both official API guides direct federated accounts to IAM token authentication, so such accounts use the `iam` token type.
- API access is enabled in the organization. When a Wiki administrator has disabled API access, the Wiki integration cannot be configured.
- Requests run on behalf of the token owner. Tracker and Wiki return only what that user is allowed to read.

The user registers the OAuth application and issues the token while logged in to the account that works with the API, as the steps of the [Tracker](https://yandex.ru/support/tracker/en/concepts/access) and [Wiki](https://yandex.ru/support/wiki/en/api-ref/access) API access guides describe.

1. Create an application at https://oauth.yandex.ru: choose "For API access or debugging" and add the `tracker:read` and `wiki:read` permissions.
2. Open `https://oauth.yandex.ru/authorize?response_type=token&client_id=<ClientID>` under the account that works with Tracker and Wiki, and copy the token.
3. Save the token as a single line in a file outside any repository, for example `~/.yandex-oauth-token`.
4. Configure the server for local acquisition:

```json
"yandex": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_MCP_TOKEN_TYPE": "oauth",
    "YANDEX_OAUTH_TOKEN_FILE": "/absolute/path/to/.yandex-oauth-token",
    "YANDEX_ORG_ID": "yandex 360 organization id"
  }
}
```

For a Yandex Cloud organization, set `YANDEX_CLOUD_ORG_ID` instead of `YANDEX_ORG_ID`.

For remote acquisition, save the token on the workstation as a file named after the profile in the auth-agent OAuth directory, and name that profile on the server, see [Remote OAuth tokens](#remote-oauth-tokens).

Notes:

- The token never appears in the MCP client configuration; only the file path does.
- When the token expires, the API answers `401`. Issue a new token with the same link and replace the file contents.
- Official references: https://yandex.ru/support/tracker/en/concepts/access and https://yandex.ru/support/wiki/en/api-ref/access

## Remote authentication

MCP runs on the Linux server; token acquisition runs on your workstation: `yc` with browser login for IAM tokens, token files for OAuth tokens. Install `yandex-mcp` on both machines and prepare a workstation `yc` profile or an OAuth profile file. Server-side `yc` is not needed. Use only trusted machines and users.

### Server MCP client

Add this entry to the MCP client running on the server:

```json
{
  "mcpServers": {
    "yandex": {
      "command": "/absolute/server/path/to/yandex-mcp",
      "env": {
        "YANDEX_MCP_TOKEN_SOURCE": "remote",
        "YANDEX_MCP_AUTH_AGENT_PORT": "18765",
        "YANDEX_CLI_PROFILE": "work",
        "YANDEX_CLOUD_ORG_ID": "your-organization-id"
      }
    }
  }
}
```

Replace the executable path, `work` with your workstation yc profile, and the organization ID. This entry requests an IAM token; [Remote OAuth tokens](#remote-oauth-tokens) shows the entry for an OAuth token.

### Workstation setup

Install OpenSSH with `ssh` available through `PATH`. Install native `yc` with `yc` available through `PATH` only when you use IAM tokens. Verify the server's SSH host key and configure SSH-key access without password prompts. The server must permit loopback-only reverse forwarding.

From the extracted release archive or repository, run in your graphical desktop session:

```sh
export YANDEX_MCP_BINARY=/absolute/workstation/path/to/yandex-mcp
export YANDEX_MCP_SSH_TARGET=user@server.example
sh scripts/remote-auth/macos/install.sh
```

On a systemd-integrated Linux desktop, replace `macos` with `linux`. On Windows, use PowerShell:

```powershell
$env:YANDEX_MCP_BINARY = 'C:\Tools\yandex-mcp.exe'
$env:YANDEX_MCP_SSH_TARGET = 'user@server.example'
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\remote-auth\windows\install.ps1
```

Replace the workstation executable path and SSH destination. The installer configures one auth-agent startup entry. Auth-agent starts and reconnects SSH. `YANDEX_MCP_BINARY` is optional when `yandex-mcp` is in `PATH`.

When `yc` is not in `PATH`, the installer completes with a warning. Auth-agent then serves OAuth tokens, and IAM requests fail until `yc` is installed in a directory of the `PATH` that the installer saved. To use another `yc` executable, set `YANDEX_MCP_AUTH_AGENT_YC_PATH` before you run the installer; the installer stops when this value does not resolve to an executable.

Use the same `YANDEX_MCP_AUTH_AGENT_PORT` on both machines, default `18765`.

Restart the MCP connection and call a Wiki or Tracker tool. For IAM tokens, complete login in the workstation browser when `yc` prompts. Keep the workstation awake and connected when a new token is needed. The tool timeout defaults to 300 seconds; change `YANDEX_MCP_TOOL_TIMEOUT` in the server configuration if needed.

### Remote OAuth tokens

Auth-agent also serves OAuth tokens. It reads the token of a profile from the file with the profile's name in its OAuth directory.

| Workstation OS | Default OAuth directory |
| --- | --- |
| Linux | `$XDG_CONFIG_HOME/yandex-mcp/oauth`, usually `~/.config/yandex-mcp/oauth` |
| macOS | `~/Library/Application Support/yandex-mcp/oauth` |
| Windows | `%AppData%\yandex-mcp\oauth` |

Save each token as a single line in a file whose name is the profile name, for example `work`. A profile name starts with a letter or digit and contains only letters, digits, `.`, `_` and `-`. One auth-agent serves any number of profiles, so different MCP connections can use different accounts.

To use another directory, set `YANDEX_MCP_AUTH_AGENT_OAUTH_DIR` to an absolute path before you run the installer:

```sh
export YANDEX_MCP_AUTH_AGENT_OAUTH_DIR=/absolute/path
```

On Windows, use `$env:YANDEX_MCP_AUTH_AGENT_OAUTH_DIR = 'C:\absolute\path'`. The installer rejects a relative path and saves the value in the auth-agent startup configuration, so auth-agent uses this directory after each restart. When the variable is unset at installation, auth-agent uses the default directory from the table. On Linux, the default directory depends on `XDG_CONFIG_HOME` in the environment of the systemd user service; when you use a non-default `XDG_CONFIG_HOME`, set the directory explicitly.

The server entry names the token type and the profile:

```json
"yandex": {
  "command": "/absolute/server/path/to/yandex-mcp",
  "env": {
    "YANDEX_MCP_TOKEN_SOURCE": "remote",
    "YANDEX_MCP_TOKEN_TYPE": "oauth",
    "YANDEX_MCP_AUTH_AGENT_PORT": "18765",
    "YANDEX_OAUTH_PROFILE": "work",
    "YANDEX_ORG_ID": "yandex 360 organization id"
  }
}
```

MCP requests a token of the configured type for the profile name, and auth-agent returns a token of exactly that type or an error. An unknown OAuth profile and a failed acquisition are errors: auth-agent never substitutes another acquisition method. yc profile `work` and OAuth profile `work` are different tokens.

Remote OAuth tokens need auth-agent of a release after v1.0.4. Auth-agent v1.0.4 and earlier answers an OAuth request with the error `unknown object member name "token_type"`, and the tool returns this error. Remote IAM tokens work with any pair of MCP and auth-agent versions from v1.0.1 on.

## Client configuration examples

### Claude Code

```bash
claude mcp add -s user -e YANDEX_CLOUD_ORG_ID={yandex organization id} -e YANDEX_CLI_PROFILE={yc profile name} --transport stdio yandex /path/to/yandex-mcp
```

### VS Code, RooCode, etc.

```json
"yandex": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_CLOUD_ORG_ID": "yandex organization id",
    "YANDEX_CLI_PROFILE": "yc profile name"
  }
}
```

### Multiple organizations

When you need to work with two organizations, configure two independent MCP server entries. Each entry carries its own organization ID and its own authentication settings, matching the type of that organization:

- Yandex Cloud organization with IAM tokens: `YANDEX_CLOUD_ORG_ID` and `YANDEX_CLI_PROFILE`.
- Yandex 360 for Business organization: `YANDEX_ORG_ID`, `YANDEX_MCP_TOKEN_TYPE=oauth` and `YANDEX_OAUTH_TOKEN_FILE` (or `YANDEX_OAUTH_PROFILE` for remote acquisition). `YANDEX_CLI_PROFILE` is not used.

Set only one organization ID per entry, as [Environment variables](#environment-variables) describes.

Two Yandex Cloud organizations:

```json
"yandex_primary": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_CLOUD_ORG_ID": "primary organization id",
    "YANDEX_CLI_PROFILE": "primary yc profile"
  }
},
"yandex_partner": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_CLOUD_ORG_ID": "partner organization id",
    "YANDEX_CLI_PROFILE": "partner yc profile"
  }
}
```

A Yandex Cloud organization and a Yandex 360 organization:

```json
"yandex_cloud": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_CLOUD_ORG_ID": "cloud organization id",
    "YANDEX_CLI_PROFILE": "yc profile name"
  }
},
"yandex_360": {
  "command": "/path/to/yandex-mcp",
  "env": {
    "YANDEX_MCP_TOKEN_TYPE": "oauth",
    "YANDEX_OAUTH_TOKEN_FILE": "/absolute/path/to/.yandex-oauth-token",
    "YANDEX_ORG_ID": "yandex 360 organization id"
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
