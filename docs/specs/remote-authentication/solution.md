# Remote Yandex MCP authentication technical solution

## Problem statement

The [problem statement](problem.md) gives the basis for this solution. The [requirements](requirements.md) define expected behavior and constraints. Terms are defined in the [glossary](glossary.md). The [Yandex 360 OAuth solution](../yandex-360-oauth/solution.md) describes token types, organization headers, and OAuth profiles.

## Proposed solution

### D1. Two modes in one binary

- Normal `yandex-mcp` invocation exposes tools over stdio. Local mode obtains tokens on the MCP machine; remote mode requests them from the workstation. On each machine the token type selects the token source, as [S1](../yandex-360-oauth/solution.md#s1-token-type-and-acquisition-location-frq01) describes.
- `yandex-mcp auth-agent` runs the workstation HTTP listener and manages a native OpenSSH reverse tunnel to the Linux server.
- The OS starts one auth-agent entry in the user's graphical session. There is no separate tunnel service. macOS uses a LaunchAgent, Linux uses a systemd user service, and Windows uses an interactive Task Scheduler task.
- Auth-agent restarts SSH five seconds after a connection attempt ends. OpenSSH uses `ServerAliveInterval=15`, `ServerAliveCountMax=3`, `ConnectTimeout=10`, `BatchMode=yes`, and `ExitOnForwardFailure=yes`. SSH reads the user's normal keys and host configuration and requires an already trusted server host key.
- Stopping auth-agent cancels and joins SSH and active acquisitions, including `yc` processes. A connection loss interrupts affected token requests; reconnection permits subsequent requests. Interrupted tool calls are not resumed automatically.

Request path:

```text
Agent -> MCP on Linux -> reverse SSH tunnel
      -> auth-agent on workstation -> IAM:   native yc -> workstation browser
                                   -> OAuth: OAuth profile file
```

Auth-agent does not proxy Wiki or Tracker API calls. The browser callback remains on the workstation and needs no additional tunnel.

### D2. Overall tool timeout

`internal/server` installs middleware through `mcp.Server.AddReceivingMiddleware`. Every `tools/call` gets one deadline before its handler runs.

- `YANDEX_MCP_TOOL_TIMEOUT` defaults to 300 positive seconds.
- Argument validation, token acquisition, login, API requests, and result processing share this deadline.
- Sequential API requests and the retry after HTTP 401 or 403 use the remaining time. API clients have no independent `http.Client.Timeout`.
- The deadline produces a tool error. Earlier client cancellation ends that client's wait.

### D3. Token sources and cache

`internal/adapters/ytoken.Provider` caches a token of the configured token type and uses an injected token source. `cmd/yandex-mcp` selects the source by the acquisition location and the token type: `authremote` in remote mode; `yc` for IAM tokens and `oauthfile` for OAuth tokens in local mode.

- Each MCP process caches its own token. Auth-agent shares active acquisitions but has no token cache.
- `YANDEX_IAM_TOKEN_REFRESH_PERIOD` controls cache refresh. HTTP 401 or 403 triggers a forced refresh and one API retry within the tool deadline.
- A failed refresh returns its error, not an old token. Remote errors never switch the server to a local token source.

### D4. Shared waiting and cancellation

`internal/adapters/authwait.Group` owns active acquisitions and their waiter counts. The shared context does not inherit the first caller's deadline.

- The MCP group combines calls within one process. The workstation group combines requests from different MCP processes.
- Canceling one caller removes only that waiter. The last waiter leaving cancels acquisition.
- The workstation group keys acquisitions by token type and profile, so IAM profile `work` and OAuth profile `work` never share one acquisition.
- The group keeps the profile occupied until acquisition returns. Native `yc` acquisition returns only after `exec.Cmd.Run` finishes, so the next process for that profile cannot overlap the previous process.
- Shutdown cancels the acquisitions owned by the stopping process and waits for cleanup.

### D5. Token exchange and protection

Auth-agent accepts `POST /token` on IPv4 loopback. The same configurable port, default 18765, is used on the workstation and the server. The bind address is fixed to `127.0.0.1`; SSH server policy must preserve loopback-only forwarding.

- Both machines and their local processes are trusted. There is no HTTP authorization or profile allowlist. Any process with access to a listener can request a workstation profile.
- The request names the profile and the token type in the format of D6. For `iam`, auth-agent executes only `yc iam create-token --profile <profile>`, without a shell or caller-supplied commands. For `oauth`, it reads the file of the OAuth profile in its OAuth directory; [S4](../yandex-360-oauth/solution.md#s4-oauth-profiles-on-the-workstation-frq04) defines profile names and the directory.
- Auth-agent returns a token of the requested type or an error. It never substitutes another acquisition method.
- The HTTP request waits for acquisition. Closing the request cancels its wait. There are no authentication jobs or polling endpoints.
- Remote requests use the fixed loopback address without HTTP proxies or redirects.
- Credential files stay on the workstation. MCP caches tokens in memory. Successful `yc` output is not logged.

### D6. Errors

Requests and responses use JSON without a separate protocol version or string error codes.

IAM token request:

```json
{"profile":"work"}
```

OAuth token request:

```json
{"token_type":"oauth","profile":"work"}
```

Auth-agent treats a missing or empty `token_type` as `iam`. MCP sends `token_type` only for OAuth tokens, because auth-agent v1.0.1 to v1.0.4 rejects unknown request members and accepts a request that holds only the profile. [S7](../yandex-360-oauth/solution.md#s7-request-compatibility-with-released-versions-nrq01) lists the behavior of each version combination.

HTTP 200 response:

```json
{"token":"<token>"}
```

The response does not repeat the token type: MCP builds `Bearer` or `OAuth` from the type it requested.

A failed acquisition returns HTTP 502 with the original message of the token source. Examples are a failed `yc` invocation, an unknown OAuth profile, and an unreadable OAuth token file. For a failed `yc` invocation the response is:

```json
{"message":"exit status 1\n<original stderr>"}
```

HTTP 400 reports invalid requests, including a missing profile and an unsupported token type. HTTP 500 reports an empty token returned by the source. Requests are limited to 4096 bytes; error responses are not truncated. Invalid JSON, unknown fields, and conflicting token/message fields are rejected.

MCP preserves the response's HTTP status and original message. `domain.AuthenticationError` prevents tool error handling from replacing authentication diagnostics with `internal error`. The complete failed-command stderr reaches the tool result and the structured MCP log without redaction. Connection errors retain their original transport cause. SSH process errors are logged on the workstation.

Token requests are not automatically retried. A subsequent tool call can request a token after the cause is resolved.

### D7. Configuration boundaries

- Remote MCP requires `YANDEX_MCP_TOKEN_SOURCE=remote`. [S1](../yandex-360-oauth/solution.md#s1-token-type-and-acquisition-location-frq01) and [S2](../yandex-360-oauth/solution.md#s2-authorization-scheme-and-organization-frq02) describe the token type, the organization, and the profile setting that each combination requires.
- Workstation auth-agent requires `YANDEX_MCP_SSH_TARGET`, an SSH alias or `user@host`. It does not load organization or Wiki/Tracker settings.
- `YANDEX_MCP_AUTH_AGENT_PORT` defaults to 18765 on both machines. Different port numbers at opposite tunnel ends are not supported.
- Auth-agent needs native `yc` only for IAM requests. At each IAM request it runs the command that `YANDEX_MCP_AUTH_AGENT_YC_PATH` names, default `yc` from `PATH`. OpenSSH is found through `PATH` and is always required.
- Auth-agent reads OAuth profiles from the OAuth directory that `YANDEX_MCP_AUTH_AGENT_OAUTH_DIR` sets or that [S4](../yandex-360-oauth/solution.md#s4-oauth-profiles-on-the-workstation-frq04) defines as the default. IAM profiles are not duplicated in auth-agent configuration.
- Installers locate `yandex-mcp` through `PATH` unless `YANDEX_MCP_BINARY` overrides it. They require OpenSSH and treat `yc` as optional. They persist the executable paths, `YANDEX_MCP_AUTH_AGENT_OAUTH_DIR`, and the installation session's `PATH` in the startup configuration of the single user-session startup entry. [S8](../yandex-360-oauth/solution.md#s8-installation-for-oauth-on-the-workstation-frq07-frq08) describes installation without `yc` and the saved OAuth directory. Auth-agent itself has no configuration-file parser.

The [README](../../../README.md#remote-authentication) contains the connection example and installation commands. GitHub release archives contain binaries, README, LICENSE, `.env.example`, and startup scripts. Internal specifications remain in the repository.

## Overengineering and overspecification considerations

- MCP stays on stdio. HTTP is used only for token acquisition.
- Native `yc` handles IAM login, native OpenSSH handles SSH, and the OS handles auth-agent startup. No custom SSO, Go SSH implementation, general process supervisor, database, or job queue is introduced.
- One shared-waiting implementation serves both acquisition levels. There is no additional per-profile lock in the CLI adapter.

## Solution verification

Verification covers token caching, the overall deadline across refresh and API retry, independent waiter cancellation, native process cleanup, arbitrary requested profiles, all four combinations of acquisition location and token type, OAuth profile selection, independent IAM and OAuth acquisitions for equal profile names, IAM acquisition for a request without a token type, rejection of unsupported token types, missing profiles, and unknown OAuth profiles, original diagnostics in MCP results and logs, and native SSH restart and shutdown.

A Linux ARM64 workstation container and Linux AMD64 server exercised auth-agent with its own native SSH child and no container restart policy. Token exchange recovered after killing SSH and after disconnecting and reconnecting the container network. Auth-agent stayed in the same process with zero container restarts. Two MCP processes shared one acquisition; cancellation preserved the other waiter, and final-waiter cancellation stopped native `yc`. Full diagnostics reached the tool result and MCP log. Auth-agent shutdown exited normally and released the server listener. The check used a test native `yc` executable, not browser login.

Browser login, graphical-session startup, workstation sleep, and native Windows execution require platform acceptance checks. Cross-compilation does not establish those behaviors.

## Open questions

There are no open design questions.

## References

- [MCP server wrapper](../../../internal/server/service.go): tool registration and overall timeout middleware.
- [Token provider](../../../internal/adapters/ytoken/client.go) and [shared acquisition group](../../../internal/adapters/authwait/service.go): caching and independent cancellation.
- [Auth-agent token handler](../../../internal/server/authagent/service.go): request validation, the default token type, and shared acquisitions by token type and profile.
- [Remote token source](../../../internal/adapters/authremote/client.go): the MCP side of the token request.
- [Auth-agent SSH management](../../../internal/server/authagent/ssh.go): native tunnel execution and restart.
- [Yandex 360 OAuth solution](../yandex-360-oauth/solution.md): token types, organization headers, OAuth profiles, and installation without `yc`.
- [API client](../../../internal/adapters/apihelpers/client.go): authenticated requests and forced-refresh retry.
- [OpenSSH manual](https://man.openbsd.org/ssh.1): reverse port forwarding and SSH options.
