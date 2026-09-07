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
- `YANDEX_HTTP_TIMEOUT` and its associated `http.Client.Timeout` setting are removed without a compatibility alias. Lower layers use the call context.
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

The `singleflight` implementation in use passes the first call's context to the shared acquisition. Other calls wait without checking their own cancellation. It is therefore replaced with a small shared mechanism that tracks waiters.

Both the MCP provider and auth-agent use this mechanism:

- Each call waits for the result with its own context.
- The shared acquisition does not inherit the first call's deadline.
- Canceling one call removes only that waiter.
- The shared acquisition is canceled after the last waiter leaves.
- On the workstation, canceling the shared acquisition stops `yc`. The next `yc` process for that profile starts only after the previous process exits. The local source enforces this process lifecycle on each supported workstation OS.
- Shutdown of MCP or auth-agent cancels the acquisitions owned by that process.

Each tool call owns its deadline. A shared acquisition exists only while calls wait for it. No separate authentication timeout is introduced.

### D5. Token exchange and protection

Auth-agent accepts synchronous `POST /token` requests on `127.0.0.1`. A request contains the protocol version and an explicit profile. A response contains the protocol version and either an IAM token or an error code. The token-exchange protocol version is separate from the MCP version and the binary release version.

- The connection secret is sent in `Authorization: Bearer`.
- The secret, protocol version, and profile permission are checked before `yc` can start.
- The workstation runs the fixed command `yc iam create-token --profile <profile>` using the native executable. Shell execution and arbitrary arguments are not accepted.
- The request stays open until completion or cancellation. There are no separate jobs, status polling, or background authentication completion without waiters.
- Closing the HTTP request ends its wait on the workstation.
- The remote source contacts only the configured `http://127.0.0.1:<port>` address. Environment HTTP proxies and redirects are disabled.
- Secret-file validation uses mode `600` on macOS and Linux. On Windows, it checks file ACLs rather than Unix permission bits. Other nonprivileged users must not have read or write access to the secret files.
- Secret files remain outside Git. MCP does not persist IAM tokens to disk.
- Tokens, secrets, login URLs, and raw `yc` output do not appear in logs.

The secret grants access to specified profiles. It does not protect against compromise of the server account that can read that secret.

### D6. Errors

Auth-agent returns safe error codes without `yc` output:

- HTTP 400 indicates an invalid request or an incompatible protocol version. These causes have separate codes.
- HTTP 401 indicates an incorrect connection secret.
- HTTP 403 indicates a forbidden profile.
- HTTP 502 indicates that `yc` could not issue a token.
- HTTP 500 indicates an internal auth-agent error.
- A connection error indicates that the remote source is unavailable.
- An unexpected auth-agent response indicates a protocol error. Its body is not passed to the agent.

MCP converts these causes into tool errors. Token requests are not automatically retried after these errors. A subsequent tool call can retry token acquisition after the cause is resolved.

### D7. Configuration boundaries

Configuration loading is separate for the two invocation modes:

- MCP configuration selects `local` or `remote`, defaulting to `local`. Remote mode requires a loopback endpoint, a connection-secret file, and an explicit profile. Local mode can use the active `yc` profile.
- Auth-agent configuration contains the loopback listen address, the absolute native `yc` executable path, and client secrets with their profile allowlists. It does not load Wiki or Tracker settings and does not require `YANDEX_CLOUD_ORG_ID`.

The entry point loads only the configuration for the selected mode and constructs the required dependencies. This keeps CLI path handling and secret-file access separate from token caching and the HTTP protocol.

The [README remote authentication section](../../../README.md#planned-remote-authentication) contains the user-facing settings, installation, automatic startup, SSH configuration, and troubleshooting.

## Overengineering and overspecification considerations

- MCP remains a stdio server. The network interface is only for token acquisition.
- No custom SSO, database, job queue, status polling, or additional supervisor is introduced.
- Token-source and shared-waiting code is reused. A full project architecture rewrite is unnecessary.

## Solution verification

Implementation verification covers these mechanisms:

- The overall deadline across token acquisition and the subsequent API request, including the retry after HTTP 401 or 403.
- Cancellation of the first and last waiters at both acquisition-sharing levels, including `yc` termination before the next process for that profile starts.
- Rejection of an incorrect secret, forbidden profile, and incompatible version before `yc` starts.
- Local-mode operation without auth-agent.
- Native process cancellation and secret-file access checks on each supported workstation OS.

Browser login and connection recovery have not been verified end to end. The [README runtime checks](../../../README.md#remote-authentication-runtime-checks) describe the operational acceptance checks for each workstation platform.

## Open questions

There are no open design questions.

## References

- [MCP server wrapper](../../../internal/server/service.go) and MCP SDK `v1.6.1`, `mcp/server.go`, `Server.AddReceivingMiddleware`: tool registration and middleware support.
- [Token provider](../../../internal/adapters/ytoken/client.go) and `github.com/n-r-w/singleflight/v2 v2.0.0`, `singleflight.go`, `Group.Do`: caching and shared-acquisition cancellation constraints.
- [API client](../../../internal/adapters/apihelpers/client.go): token acquisition before HTTP requests and retry after authentication errors.
- [OpenSSH manual](https://man.openbsd.org/ssh.1): reverse port forwarding.
