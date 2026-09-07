# Remote Yandex MCP authentication technical solution

## Problem statement

The [problem statement](problem.md) gives the basis for this solution. The [requirements](requirements.md) define expected behavior and constraints. Terms are defined in the [glossary](glossary.md).

## Proposed solution

### D1. Two modes in one binary

- Normal `yandex-mcp` invocation exposes tools over stdio.
- `yandex-mcp auth-agent` starts only an HTTP interface for IAM token acquisition on the workstation.
- The token source is selected explicitly. The local source calls `yc` on the MCP machine. The remote source contacts the workstation.
- A remote-source error does not switch MCP to local `yc`.

Remote request path:

```text
Agent → MCP on Linux → reverse SSH tunnel
      → auth-agent on workstation → local yc → workstation browser
```

Auth-agent does not proxy the Wiki or Tracker APIs. Agent files and results remain on the Linux server. The browser login callback stays on the workstation and does not need a separate tunnel. The token protocol and cache logic do not depend on the workstation OS.

### D2. Overall tool timeout

Middleware is added in `internal/server` through `mcp.Server.AddReceivingMiddleware`. For `tools/call`, it creates a context with an overall deadline before passing control to the tool handler.

- `YANDEX_MCP_TOOL_TIMEOUT` specifies a positive number of seconds, defaulting to `300`.
- The duration includes tool argument validation, token acquisition, login, API requests, and result processing.
- Sequential requests and the retry after an authentication error use the remaining time, not another 300 seconds.
- API clients have no independent `http.Client.Timeout`. Lower layers use the call context.
- When its own timeout expires, MCP returns a tool error. Earlier MCP-client cancellation stops execution, but response delivery to the client that canceled the call is not guaranteed.

The middleware covers all `tools/call` requests, including tools without HTTP requests. Per-tool timeout configuration is therefore unnecessary.

### D3. Token sources and cache

`internal/adapters/ytoken.Provider` retains responsibility for caching and refreshing tokens. It obtains a new token from an injected source through an interface defined in the provider package.

- The `yc` invocation and output parsing move into the local source. Both local MCP and auth-agent use this source.
- The remote source sends a request to auth-agent.
- Dependencies are assembled in `cmd/yandex-mcp`. The provider does not select a concrete source implementation.
- Each MCP process retains its own cache. Auth-agent shares only active acquisitions and does not add a second token cache.
- `YANDEX_IAM_TOKEN_REFRESH_PERIOD` retains its role as the refresh period. Wiki and Tracker HTTP 401 or 403 responses trigger a refresh and one API request retry within the overall deadline.
- A refresh error is returned to the caller. The old token is not returned in place of a failed refresh result.

### D4. Shared waiting and cancellation

`internal/adapters/authwait.Group` tracks active acquisitions and their waiters. It owns one cancelable context per active profile, separate from each caller's context.

Both the MCP provider and auth-agent use this mechanism:

- Each call waits for the result with its own context.
- The shared acquisition does not inherit the first call's deadline.
- Canceling one call removes only that waiter.
- The shared acquisition is canceled after the last waiter leaves.
- On the workstation, canceling the shared acquisition stops `yc`. The next `yc` process for that profile starts only after the previous process exits. The local source enforces this process lifecycle on each supported workstation OS.
- Shutdown of MCP or auth-agent cancels the acquisitions owned by that process.

Each tool call owns its deadline. A shared acquisition exists only while calls wait for it. No separate authentication timeout is introduced.

### D5. Token exchange and protection

Auth-agent accepts synchronous `POST /token` requests on `127.0.0.1`. A request contains the protocol version and an explicit profile. A response contains the protocol version and either an IAM token or an error code with the original error message. The token-exchange protocol version is separate from the MCP version and the binary release version.

- Both machines and their local processes are trusted. SSH encrypts traffic between them. There is no HTTP authorization or connection secret.
- The protocol version and profile permission are checked before `yc` can start.
- The workstation runs the fixed command `yc iam create-token --profile <profile>` using the native executable. Shell execution and arbitrary arguments are not accepted.
- The request stays open until completion or cancellation. There are no separate jobs, status polling, or background authentication completion without waiters.
- Closing the HTTP request ends its wait on the workstation.
- The remote source contacts only the configured `http://127.0.0.1:<port>` address. Environment HTTP proxies and redirects are disabled.
- MCP does not persist IAM tokens to disk. Successful `yc` output is not logged.
- A failed `yc` process supplies its original execution error and complete `stderr` to auth-agent. These diagnostics are returned to MCP and written to the MCP process's structured `stderr` log without redaction.

Any process with access to the workstation listener or the server's forwarded loopback port can request an allowed profile's token. This design is not intended for machines with untrusted users or processes.

### D6. Errors

Auth-agent returns error codes and original messages:

- HTTP 400 indicates an invalid request or an incompatible protocol version. These causes have separate codes.
- HTTP 403 indicates a forbidden profile.
- HTTP 502 indicates that `yc` could not issue a token.
- HTTP 500 indicates an internal auth-agent error.
- A connection error indicates that the remote source is unavailable.
- An unexpected auth-agent response indicates a protocol error. A response outside the token protocol is not treated as a token-source diagnostic.

MCP preserves the original `message` in both its tool error and its structured `stderr` log. Authentication diagnostics retain their type through API wrappers so tool error handling does not replace them with `internal error`. The complete `stderr` of a failed `yc` process is included; successful token output is excluded. Token requests are not automatically retried after these errors. A subsequent tool call can retry token acquisition after the cause is resolved.

Protocol version 1 uses `POST /token` with `Content-Type: application/json`:

```json
{"version":1,"profile":"work"}
```

Success returns HTTP 200:

```json
{"version":1,"token":"<IAM-token>"}
```

A failed `yc` invocation returns HTTP 502:

```json
{"version":1,"error":"authentication_failed","message":"exit status 1\n<original stderr>"}
```

The other error codes are `invalid_request`, `incompatible_version`, `forbidden_profile`, and `internal_error`. Requests are limited to 4096 bytes. Error responses are not truncated. Unknown JSON fields and conflicting success/error fields are protocol errors.

### D7. Configuration boundaries

Configuration loading is separate for the two invocation modes:

- MCP configuration selects `local` or `remote`, defaulting to `local`. Remote mode requires a loopback endpoint and an explicit profile. Local mode can use the active `yc` profile.
- Auth-agent reads only environment variables: `YANDEX_MCP_AUTH_AGENT_LISTEN`, `YANDEX_MCP_AUTH_AGENT_YC_PATH`, and `YANDEX_MCP_AUTH_AGENT_PROFILES`. These specify the loopback listen address, absolute native `yc` executable path, and comma-separated allowed profiles. It does not load Wiki or Tracker settings and does not require `YANDEX_CLOUD_ORG_ID`.
- There is no auth-agent configuration-file parser. OS startup entries persist the environment settings for automatic startup.

The entry point loads only the configuration for the selected mode and constructs the required dependencies. This keeps CLI path handling separate from token caching and the HTTP protocol.

The [README remote authentication section](../../../README.md#remote-authentication) contains the user-facing settings, installation, automatic startup, SSH configuration, and troubleshooting.

## Overengineering and overspecification considerations

- MCP remains a stdio server. The network interface is only for token acquisition.
- No custom SSO, database, job queue, status polling, or additional supervisor is introduced.
- Token-source and shared-waiting code is reused. A full project architecture rewrite is unnecessary.

## Solution verification

Implementation verification covers these mechanisms:

- The overall deadline across token acquisition and the subsequent API request, including the retry after HTTP 401 or 403.
- Cancellation of the first and last waiters at both acquisition-sharing levels, including `yc` termination before the next process for that profile starts.
- Rejection of a forbidden profile and incompatible version before `yc` starts.
- Local-mode operation without auth-agent.
- Native process cancellation on macOS, Linux, and Windows.
- Original error delivery through a separate MCP process with no `yc` in `PATH`, including its tool result and structured log.
- Inclusion of workstation startup assets in release archives.

Local macOS and Linux tests include `-race`, native subprocess cancellation, and MCP error delivery. An SSH integration check between a macOS workstation and a test Linux server exercised a reverse tunnel with a test native `yc` executable. It checked token exchange, two separate MCP processes sharing one acquisition, independent cancellation, final-waiter process termination, and original diagnostic delivery to the tool result and MCP log.

A second SSH check used a Linux ARM64 workstation container and the Linux AMD64 test server. It repeated token exchange, shared waiting, cancellation, native process termination, and diagnostic delivery. Docker's restart policy restored the SSH tunnel after its process was terminated. A separate check disconnected the workstation container from its network, waited for OpenSSH to detect the loss and exit, then reconnected the network. Token exchange resumed without another tunnel command. These checks did not exercise the systemd user services.

Browser login, graphical-session startup, recovery after workstation sleep, and native Windows execution have not been verified end to end. Windows native tests are configured in `.github/workflows/ci.yml`; that workflow was not run during local development. The [README runtime checks](../../../README.md#remote-authentication-runtime-checks) describe the remaining operational acceptance checks.

## Open questions

There are no open design questions.

## References

- [MCP server wrapper](../../../internal/server/service.go) and MCP SDK `v1.6.1`, `mcp/server.go`, `Server.AddReceivingMiddleware`: tool registration and middleware support.
- [Token provider](../../../internal/adapters/ytoken/client.go) and [shared acquisition group](../../../internal/adapters/authwait/service.go): caching and independent waiter cancellation.
- [API client](../../../internal/adapters/apihelpers/client.go): token acquisition before HTTP requests and retry after authentication errors.
- [OpenSSH manual](https://man.openbsd.org/ssh.1): reverse port forwarding.
