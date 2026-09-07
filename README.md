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

- **Linux (AMD64)**: `yandex-mcp-v*-linux-amd64.tar.gz`
- **macOS (Intel)**: `yandex-mcp-v*-darwin-amd64.tar.gz`
- **macOS (Apple Silicon)**: `yandex-mcp-v*-darwin-arm64.tar.gz`
- **Windows (AMD64)**: `yandex-mcp-v*-windows-amd64.zip`

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

- `YANDEX_CLI_PROFILE` (optional)
  * yc CLI profile used to retrieve IAM tokens.
  * When empty, yc uses the active profile.
  * When set, the server runs `yc iam create-token --profile <YANDEX_CLI_PROFILE>`.

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

- `YANDEX_HTTP_TIMEOUT` (optional, default: `30`)
  * HTTP timeout for Yandex API requests in **seconds**.

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
  * Default rule: `save_path` must be inside the user home directory, must not point to the home root, and must not be within a hidden top-level home subdirectory (for example, `~/.ssh`).

## Authentication

The project supports IAM token authentication via the Yandex Cloud CLI (`yc`) only.

**IAM token acquisition (`yc` prerequisites)**

Installation: https://yandex.cloud/en/docs/cli/operations/install-cli

This server obtains IAM tokens by running:
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

## Planned remote authentication

Remote authentication is not implemented yet. The `auth-agent` subcommand, new settings, and installation assets described in this section are not available in the current implementation. The setup below describes intended usage after implementation; it does not replace the local authentication instructions above.

The agent and MCP will run on a Linux server. Token acquisition and browser login will run on the user's macOS, Linux, or Windows workstation. The workstation must be available when MCP needs a new token. This does not automate passwords or MFA.

The [remote authentication specification](docs/specs/remote-authentication/requirements.md) defines the behavior. The [technical solution](docs/specs/remote-authentication/solution.md) describes the program changes.

### Prerequisites

- A workstation with an interactive desktop session, a browser, native Yandex Cloud CLI, and an OpenSSH client. Linux workstations need a graphical desktop. Windows does not require WSL.
- A Linux server running the agent and MCP, with SSH access that permits loopback-only reverse forwarding. Remote-mode MCP does not require server-side `yc` installation.
- Prepared federated `yc` profiles on the workstation and an organization ID for the server's MCP configuration.

Yandex provides [CLI installation instructions](https://yandex.cloud/en/docs/cli/operations/install-cli) for macOS, Linux, and Windows. Microsoft documents [OpenSSH Client availability on Windows](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh-overview).

### Planned configuration

For the server's MCP process:

- `YANDEX_MCP_TOKEN_SOURCE`: `local` or `remote`, defaulting to `local`. Select `remote` for workstation token acquisition.
- `YANDEX_MCP_AUTH_AGENT_URL`: the source address through the server's loopback port, for example `http://127.0.0.1:18765`.
- `YANDEX_MCP_AUTH_SECRET_FILE`: the path to the server-side connection-secret file.
- `YANDEX_CLI_PROFILE`: the explicit workstation profile to request. Remote mode does not select the workstation's active profile implicitly.
- `YANDEX_CLOUD_ORG_ID`: the organization used for Wiki and Tracker API requests. Obtain this value on the workstation; no server-side CLI is needed.
- `YANDEX_MCP_TOOL_TIMEOUT`: positive seconds for the full tool call, defaulting to `300`. This planned setting replaces `YANDEX_HTTP_TIMEOUT` without an alias. The existing HTTP timeout setting above remains the one implemented today.

The refresh period remains controlled by `YANDEX_IAM_TOKEN_REFRESH_PERIOD`.

For the workstation's `yandex-mcp auth-agent` process, `YANDEX_MCP_AUTH_AGENT_CONFIG_FILE` identifies the protected configuration file. It contains the loopback listen address, absolute native `yc` executable path, and each client's connection secret and allowed profiles. Auth-agent does not need the server's organization or API settings.

### Initial setup after implementation

1. Install the MCP binary on the Linux server and the matching workstation binary on the workstation.
2. Prepare the required `yc` profiles on the workstation and complete initial browser login there.
3. Prepare a connection secret for each authorized client. Store it in the server-side secret file and the corresponding workstation configuration entry. Keep these files outside Git and do not put secrets in process arguments or logs.
4. On macOS and Linux, set secret-file permissions to `600`. On Windows, restrict the file ACL so other nonprivileged users cannot read or write the file.
5. Verify and trust the SSH server's host key. Make SSH-key authentication available to background processes without password or passphrase prompts. Do not disable host-key verification to bypass setup failures.
6. Configure the server's MCP entry and the workstation's auth-agent. Use absolute executable and configuration paths rather than relying on interactive shell startup files.
7. Configure the reverse tunnel and automatic startup as described below.

### Reverse SSH tunnel

The workstation initiates the tunnel. Inbound SSH access to the workstation is not required.

For port `18765`, the OpenSSH command is:

```sh
ssh -NT -R 127.0.0.1:18765:127.0.0.1:18765 -o BatchMode=yes -o ExitOnForwardFailure=yes -o StrictHostKeyChecking=yes -o ServerAliveInterval=30 -o ServerAliveCountMax=3 user@server.example
```

Replace the SSH destination and ports with the configured values. The first address and port identify the Linux server listener. The second address and port identify auth-agent on the workstation. These are not browser callback ports.

Verify that the effective server listener is loopback-only. The SSH server's forwarding policy must not expose it to the network. The [OpenSSH manual](https://man.openbsd.org/ssh.1) describes reverse forwarding and the effect of server policy on the bind address.

`BatchMode=yes` prevents interactive credential prompts. `ExitOnForwardFailure=yes` stops SSH when it cannot establish forwarding. The `ServerAlive` settings let SSH detect a lost connection so the workstation's startup mechanism can restart it.

### Workstation automatic startup

Run auth-agent and the SSH tunnel in the user's session. Configure automatic restart with a delay between attempts. The intended delivery includes platform-specific startup templates and setup scripts, not a new universal supervisor.

- **macOS:** use two user LaunchAgents, one for auth-agent and one for OpenSSH. Auth-agent runs in the logged-in user's graphical session. See [Apple's launchd documentation](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html).
- **Linux:** on desktops integrated with systemd, use two user services associated with the graphical session. Auth-agent needs the desktop environment required to open the browser; an SSH-only session or system service is not equivalent. See [systemd's graphical session documentation](https://www.freedesktop.org/software/systemd/man/latest/systemd.special.html).
- **Windows:** use Task Scheduler tasks triggered at user logon, configured to run in that user's interactive session. Use native `yc.exe` and `ssh.exe`. Configure task restart behavior for failed processes. See [Microsoft's task security contexts](https://learn.microsoft.com/en-us/windows/win32/taskschd/security-contexts-for-running-tasks). Do not run auth-agent as a Windows service, because [services cannot directly interact with the user](https://learn.microsoft.com/en-us/windows/win32/services/interactive-services).

Auth-agent is unavailable before the user logs in and while the workstation sleeps. After wake and network recovery, the startup mechanism and OpenSSH restore the tunnel without a new manual tunnel command.

### Use and troubleshooting

- With an active federated session, token refresh does not need user interaction. During reauthentication, complete login in the workstation browser. The original tool call continues when login finishes within its remaining timeout.
- A cached token can still be used while the workstation is unavailable. Workstation availability is needed when the next token acquisition occurs.
- An unavailable-source error means the workstation, auth-agent, or tunnel cannot be reached. Check those processes and the SSH connection.
- An incorrect-secret or forbidden-profile error requires correcting the client configuration or profile permission. Access to the server's `localhost` alone is not authorization.
- A protocol error requires compatible MCP and auth-agent versions and the correct endpoint. There is no automatic switch to server-side `yc`.
- A tool timeout ends that call's wait. The last waiter leaving also cancels token acquisition. A later tool call can try again.
- Closing a browser tab is not a reliable cancellation signal. Waiting ends when `yc` exits, the call is canceled, or the tool timeout expires.

Do not include tokens, connection secrets, login URLs, or raw `yc` output in diagnostic reports.

### Remote authentication runtime checks

These checks are required after implementation on each supported workstation platform. They have not been performed for the proposed feature.

- Start through the configured user-session mechanism and complete initial login and reauthentication. Confirm that the browser is usable and the original Linux tool call finishes within its deadline.
- Request one profile from multiple MCP processes. Confirm that only one `yc` acquisition is active. Cancel one call, then all remaining calls, and check process cleanup.
- Test workstation sleep and a separate SSH interruption. After restoring the session and network, confirm that a new tool call succeeds without manual tunnel creation.
- Inspect the server listener and secret-file access. Confirm loopback-only forwarding and rejection of incorrect secrets and forbidden profiles.

Development checks after implementation use `task test`, `task lint`, and `task build`.

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

When you need to work with two organizations, configure two independent MCP server entries. Each entry should use its own `YANDEX_CLOUD_ORG_ID` and `YANDEX_CLI_PROFILE`.

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
