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
  * Default rule: `save_path` must be inside the user home directory, must not point to the home root, and must not be within a hidden top-level home subdirectory (for example, `~/.ssh`).

## Authentication

The project obtains IAM tokens through Yandex Cloud CLI `yc`. Local mode runs `yc` on the MCP machine. [Remote mode](#remote-authentication) runs it on the workstation.

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

## Remote authentication

The agent and MCP run on a trusted Linux server. Token acquisition and browser login run on the user's macOS, Linux, or Windows workstation. The workstation must be available when MCP needs a new token. This does not automate passwords or MFA.

MCP uses stdio. Token exchange uses synchronous JSON over `POST /token` through a reverse SSH tunnel. SSH protects traffic between the machines. There is no HTTP authorization or connection secret. Every process that can access either loopback listener can request a token for an allowed profile. Do not use this mode on a server with untrusted users or processes.

### Prerequisites

- A workstation with an interactive desktop session, a browser, native Yandex Cloud CLI, and an OpenSSH client. Linux workstations need a graphical desktop. Windows does not require WSL.
- A Linux server running the agent and MCP, with SSH access that permits loopback-only reverse forwarding. Remote-mode MCP does not require server-side `yc` installation.
- Prepared federated `yc` profiles on the workstation and an organization ID for the server's MCP configuration.

Yandex provides [CLI installation instructions](https://yandex.cloud/en/docs/cli/operations/install-cli) for macOS, Linux, and Windows. Microsoft documents [OpenSSH Client availability on Windows](https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh-overview).

### Configuration

For the server's MCP process:

- `YANDEX_MCP_TOKEN_SOURCE`: `local` or `remote`, defaulting to `local`. Select `remote` for workstation token acquisition.
- `YANDEX_MCP_AUTH_AGENT_URL`: the source address through the server's loopback port, for example `http://127.0.0.1:18765`.
- `YANDEX_CLI_PROFILE`: the explicit workstation profile to request. Remote mode does not select the workstation's active profile implicitly.
- `YANDEX_CLOUD_ORG_ID`: the organization used for Wiki and Tracker API requests. Obtain this value on the workstation; no server-side CLI is needed.
- `YANDEX_MCP_TOOL_TIMEOUT`: positive seconds for the full tool call, defaulting to `300`. Authentication, API requests, and the retry after HTTP 401 or 403 share this deadline.

The refresh period remains controlled by `YANDEX_IAM_TOKEN_REFRESH_PERIOD`.

The workstation's `yandex-mcp auth-agent` uses only environment variables. It does not load organization or API settings:

- `YANDEX_MCP_AUTH_AGENT_LISTEN`: IPv4 loopback address and port, default `127.0.0.1:18765`.
- `YANDEX_MCP_AUTH_AGENT_YC_PATH`: required absolute path to the native `yc` executable.
- `YANDEX_MCP_AUTH_AGENT_PROFILES`: required comma-separated list of allowed profiles, for example `work,personal`.

Example for macOS or Linux:

```sh
YANDEX_MCP_AUTH_AGENT_YC_PATH=/absolute/path/to/yc \
YANDEX_MCP_AUTH_AGENT_PROFILES=work \
/absolute/path/to/yandex-mcp auth-agent
```

Each MCP process caches its token in memory. Auth-agent shares active acquisitions by profile but does not cache tokens.

### Initial setup

1. Install the matching MCP binaries on the Linux server and workstation. Release archives include `scripts/remote-auth`.
2. Prepare the required federated `yc` profiles on the workstation.
3. Verify and trust the SSH server's host key. Make SSH-key authentication available to background processes without password or passphrase prompts.
4. Set the server MCP environment to `YANDEX_MCP_TOKEN_SOURCE=remote`, its loopback endpoint, an explicit profile, and the organization ID. Set these in the MCP client's process configuration.
5. Configure workstation startup as described below. Use absolute executable paths.
6. Call a Wiki or Tracker tool. If `yc` needs authentication, complete login in the workstation browser. The original call continues within its remaining timeout.

### Reverse SSH tunnel

The workstation initiates the tunnel. Inbound SSH access to the workstation is not required.

For port `18765`, the OpenSSH command is:

```sh
ssh -NT -R 127.0.0.1:18765:127.0.0.1:18765 -o BatchMode=yes -o ExitOnForwardFailure=yes -o StrictHostKeyChecking=yes -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o ConnectTimeout=10 user@server.example
```

Replace the SSH destination and ports with the configured values. The first address and port identify the Linux server listener. The second address and port identify auth-agent on the workstation. These are not browser callback ports.

Verify that the effective server listener is loopback-only. The SSH server's forwarding policy must not expose it to the network. The [OpenSSH manual](https://man.openbsd.org/ssh.1) describes reverse forwarding and the effect of server policy on the bind address.

`BatchMode=yes` prevents interactive credential prompts. `ExitOnForwardFailure=yes` stops SSH when it cannot establish forwarding. The `ServerAlive` settings let SSH detect a lost connection so the workstation's startup mechanism can restart it.

### Workstation automatic startup

Run auth-agent and the SSH tunnel in the user's graphical session. The installation scripts configure the OS startup mechanism; they do not install binaries or change SSH server policy.

Set these installation-only environment variables alongside the auth-agent settings:

- `YANDEX_MCP_BINARY`: absolute workstation path to `yandex-mcp`.
- `YANDEX_MCP_SSH_TARGET`: SSH host alias or `user@host`.
- `YANDEX_MCP_REMOTE_PORT`: Linux server port, default `18765`.

On macOS or a systemd-integrated Linux desktop:

```sh
export YANDEX_MCP_BINARY=/absolute/path/to/yandex-mcp
export YANDEX_MCP_AUTH_AGENT_YC_PATH=/absolute/path/to/yc
export YANDEX_MCP_AUTH_AGENT_PROFILES=work
export YANDEX_MCP_SSH_TARGET=user@server.example
sh scripts/remote-auth/macos/install.sh
# On a Linux workstation, use scripts/remote-auth/linux/install.sh instead.
```

On Windows, run in PowerShell as the desktop user:

```powershell
$env:YANDEX_MCP_BINARY = 'C:\Tools\yandex-mcp.exe'
$env:YANDEX_MCP_AUTH_AGENT_YC_PATH = 'C:\Tools\yc.exe'
$env:YANDEX_MCP_AUTH_AGENT_PROFILES = 'work'
$env:YANDEX_MCP_SSH_TARGET = 'user@server.example'
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\scripts\remote-auth\windows\install.ps1
```

The installers store environment settings in LaunchAgent entries on macOS, an EnvironmentFile on Linux, and task launcher scripts on Windows. Auth-agent itself reads no configuration file. Rerun the installer after changing its environment settings.

- **macOS:** use two user LaunchAgents, one for auth-agent and one for OpenSSH. Auth-agent runs in the logged-in user's graphical session. See [Apple's launchd documentation](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html).
- **Linux:** on desktops integrated with systemd, use two user services associated with the graphical session. Auth-agent needs the desktop environment required to open the browser; an SSH-only session or system service is not equivalent. See [systemd's graphical session documentation](https://www.freedesktop.org/software/systemd/man/latest/systemd.special.html).
- **Windows:** use Task Scheduler tasks triggered at user logon, configured to run in that user's interactive session. Use native `yc.exe` and `ssh.exe`. A repeating one-minute trigger retries stopped tasks indefinitely. Running instances are not duplicated. See [Microsoft's task security contexts](https://learn.microsoft.com/en-us/windows/win32/taskschd/security-contexts-for-running-tasks). Do not run auth-agent as a Windows service, because [services cannot directly interact with the user](https://learn.microsoft.com/en-us/windows/win32/services/interactive-services).

Auth-agent is unavailable before the user logs in and while the workstation sleeps. After wake and network recovery, the startup mechanism and OpenSSH restore the tunnel without a new manual tunnel command.

### Use and troubleshooting

- With an active federated session, token refresh does not need user interaction. During reauthentication, complete login in the workstation browser. The original tool call continues when login finishes within its remaining timeout.
- A cached token can still be used while the workstation is unavailable. Workstation availability is needed when the next token acquisition occurs.
- An unavailable-source error means the workstation, auth-agent, or tunnel cannot be reached. Check those processes and the SSH connection.
- A forbidden-profile error requires adding the requested profile to `YANDEX_MCP_AUTH_AGENT_PROFILES` or correcting `YANDEX_CLI_PROFILE`.
- A protocol error requires compatible MCP and auth-agent versions and the correct endpoint. There is no automatic switch to server-side `yc`.
- A tool timeout ends that call's wait. The last waiter leaving also cancels token acquisition. A later tool call can try again.
- Closing a browser tab is not a reliable cancellation signal. Waiting ends when `yc` exits, the call is canceled, or the tool timeout expires.

Auth-agent returns original errors, including the complete `stderr` of a failed `yc` process. MCP includes these diagnostics in its tool error and structured `stderr` log. Successful token output is not logged. Failed-command diagnostics are not redacted and can contain login URLs or account details; inspect them before sharing logs.

Workstation startup logs are in `~/Library/Logs/yandex-mcp` on macOS, the user journal on Linux, and `%LOCALAPPDATA%\YandexMCP` on Windows. On Linux, use `journalctl --user -u yandex-mcp-auth-agent -u yandex-mcp-ssh-tunnel`.

### Remote authentication runtime checks

Run these operational checks on each workstation platform. Unit and subprocess tests do not establish browser behavior or recovery after a real workstation sleep.

- Start through the configured user-session mechanism and complete initial login and reauthentication. Confirm that the browser is usable and the original Linux tool call finishes within its deadline.
- Request one profile from multiple MCP processes. Confirm that only one `yc` acquisition is active. Cancel one call, then all remaining calls, and check process cleanup.
- Test workstation sleep and a separate SSH interruption. After restoring the session and network, confirm that a new tool call succeeds without manual tunnel creation.
- Inspect the server listener. Confirm loopback-only forwarding and rejection of forbidden profiles.
- Cause `yc` to fail. Confirm that the original diagnostic reaches both the tool result and MCP's `stderr` log.

Development checks use `task fmt`, `task test`, `task lint`, and `task build`.

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
