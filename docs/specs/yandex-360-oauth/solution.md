# Technical Solution: Yandex 360 organizations with OAuth tokens

## Problem Statement

The [problem statement](problem.md) describes the problem, and the [requirements](requirements.md) define the expected behavior. Terms are defined in the [glossary](glossary.md).

- Members of Yandex 360 organizations cannot use Tracker and Wiki tools of `yandex-mcp`.

## Proposed Solution

MCP gets a token type setting that is independent of the acquisition location. The token type selects the token source on each machine, the `Authorization` scheme, and the allowed organization type. Auth-agent serves OAuth tokens from named files on the workstation and keeps the request format of released versions for IAM tokens.

### S1. Token type and acquisition location (FRQ.01)

- `YANDEX_MCP_TOKEN_SOURCE=local|remote` selects the acquisition location; `YANDEX_MCP_TOKEN_TYPE=iam|oauth`, default `iam`, selects the token type. The two settings are independent.
- MCP builds its token source from both settings:

| Location | Type | Token source | Setting that names the token |
|---|---|---|---|
| `local` | `iam` | `yc` on the MCP machine | `YANDEX_CLI_PROFILE`, optional |
| `local` | `oauth` | file on the MCP machine | `YANDEX_OAUTH_TOKEN_FILE`, absolute path |
| `remote` | `iam` | auth-agent, `yc` on the workstation | `YANDEX_CLI_PROFILE`, required |
| `remote` | `oauth` | auth-agent, OAuth profile on the workstation | `YANDEX_OAUTH_PROFILE`, required |

- The token provider caches a token of either type in the same way; `YANDEX_IAM_TOKEN_REFRESH_PERIOD` controls the refresh of both, and HTTP 401 or 403 forces a refresh.

### S2. Authorization scheme and organization (FRQ.02)

- MCP sends an IAM token as `Authorization: Bearer <token>` and an OAuth token as `Authorization: OAuth <token>`, for local and remote acquisition alike.
- `YANDEX_CLOUD_ORG_ID` selects a Yandex Cloud organization and the `X-Cloud-Org-Id` header; `YANDEX_ORG_ID` selects a Yandex 360 organization and the `X-Org-Id` header.
- Configuration loading rejects three cases at startup: no organization ID, both organization IDs, and `YANDEX_ORG_ID` with the IAM token type, because a Yandex 360 organization accepts only OAuth tokens.

### S3. Local OAuth token (FRQ.03)

The local OAuth source reads `YANDEX_OAUTH_TOKEN_FILE` at each acquisition, so a replaced token takes effect at the next refresh without a restart. The source strips a byte order mark and surrounding whitespace. A missing file is an authentication error that names the file. An empty file, or a file whose content has spaces or non-ASCII characters, is an authentication error that omits the content.

### S4. OAuth profiles on the workstation (FRQ.04)

- An OAuth profile is a file in the OAuth directory whose name is the profile name. The default OAuth directory is `<user config dir>/yandex-mcp/oauth`, where Go `os.UserConfigDir` gives the user config dir; `YANDEX_MCP_AUTH_AGENT_OAUTH_DIR` replaces it with an absolute path.
- A profile name starts with a letter or digit and contains only letters, digits, `.`, `_`, and `-`, so a request cannot read a file outside the OAuth directory. Auth-agent reports a profile without a file as an unknown OAuth profile.
- One auth-agent serves any number of OAuth profiles; yc profiles stay in `yc` and are not duplicated in auth-agent configuration.

### S5. Strict source selection (FRQ.05)

Auth-agent selects the token source only by the token type of the request, and MCP selects its source only by its configuration. Each failure returns the original message of the selected source; no component retries the acquisition with another source. Auth-agent rejects an unsupported token type with HTTP 400.

### S6. Shared acquisitions by token type (FRQ.06)

Auth-agent keys shared acquisitions by the token type and the profile name. yc profile `work` and OAuth profile `work` therefore have separate acquisitions, and a waiter never receives a token of another type.

### S7. Request compatibility with released versions (NRQ.01)

- The [remote authentication solution](../remote-authentication/solution.md) defines the auth-agent endpoint `POST /token`, the tunnel, HTTP statuses, and shared waiting; this change keeps them and adds the token type to the request.
- A request without `token_type`, or with an empty `token_type`, asks for an IAM token. MCP sends `token_type` only for OAuth tokens, so an IAM request keeps the released format `{"profile":"work"}`; an OAuth request is `{"token_type":"oauth","profile":"work"}`.
- The released auth-agent rejects unknown request members, and this change keeps the IAM request free of new members. Version combinations then behave as follows:

| MCP | Auth-agent | IAM request | OAuth request |
|---|---|---|---|
| v1.0.1 to v1.0.4 | this change | token | not sent: these MCP versions have no OAuth |
| this change | v1.0.1 to v1.0.4 | token | HTTP 400 `unknown object member name "token_type"`, returned in the tool error |
| this change | this change | token | token |

- Trade-off: the contract has a default token type. An explicit error for a missing token type would break every mixed-version pair for IAM. Cost: a client that omits the token type receives an IAM token without an error.

### S8. Installation for OAuth on the workstation (FRQ.07, FRQ.08)

- When `YANDEX_MCP_AUTH_AGENT_YC_PATH` is unset and `yc` is not found, the installer completes and writes the command name `yc` to the startup configuration. Auth-agent then looks up `yc` through the saved `PATH` at each IAM request, and the request fails with the lookup error while `yc` is absent. An installation of `yc` into a directory of the saved `PATH` works without reinstalling auth-agent. When `YANDEX_MCP_AUTH_AGENT_YC_PATH` is set and does not resolve to an executable, the installation fails, because the user named an executable that does not exist.
- The installer rejects a relative `YANDEX_MCP_AUTH_AGENT_OAUTH_DIR` and writes its value to the startup configuration. An unset variable is written as an empty value, and auth-agent then computes the default OAuth directory at each startup.
- Trade-off: on Linux the default OAuth directory depends on `XDG_CONFIG_HOME` in the environment of the systemd user service. When the service sees a different value than the installation shell, auth-agent reads profiles from a default directory other than the one that the README table shows. The installers do not compute the default directory, because a copy of the Go rule in three scripts would drift from the code; a user with a non-default `XDG_CONFIG_HOME` sets the directory explicitly.

### S9. User documentation (FRQ.09)

README describes the token type choice, the four combinations, local OAuth, remote OAuth with the OAuth directory per operating system, the OAuth prerequisites, and the installation without `yc`. `.env.example` lists each variable of S1, S2, S4, and S8 with the combinations that require it.

### Verification

- SCN.01 and SCN.03: tests of the Tracker and Wiki clients send requests to a local HTTP server and check the scheme and the organization header for both token types.
- SCN.02 and SCN.03: `TestTokenProviderCombinations` in `cmd/yandex-mcp/main_test.go` builds the token provider for each of the four combinations against a test `yc` executable and a real auth-agent HTTP handler. The test builds auth-agent with the same function that the `auth-agent` command uses, so a wrong source in the production composition fails the test.
- SCN.05: `TestTokenProviderCombinations` covers a missing token file and an unknown OAuth profile, and checks that `yc` stays unused. `TestTokenProtocol` in `internal/server/authagent/service_test.go` covers an unsupported token type, which auth-agent rejects without running a source. No test runs a failing `yc` through auth-agent. `TestRemoteMCPPreservesErrorAndLogs` in `cmd/yandex-mcp/main_test.go` covers the path above `yc`: a real auth-agent handler with a mocked IAM source that fails, and the MCP process that returns the error in the tool result. `TestAcquire` in `internal/adapters/yc/client_test.go` checks that a failing `yc` gives an error with its exit status and output.
- SCN.04: configuration tests cover each rejected organization case.
- SCN.06: an auth-agent test runs concurrent IAM and OAuth requests with profile `work`.
- SCN.07 and SCN.08 on Linux: a manual run checked the installer; it is not a test in the repository. The run used a copy of the installer with redirected output directories and stub commands, then started auth-agent with only the environment of the generated `auth-agent.env` and another `XDG_CONFIG_HOME`.
- SCN.07 and SCN.08 on macOS: `scripts/remote-auth/macos/test_install.py` checks that the macOS installer completes without `yc`, writes the command name `yc`, rejects a missing explicit `yc` path and a relative OAuth directory, and writes the OAuth directory to the plist. The macOS installer and auth-agent also ran once on macOS arm64 with the real `plutil`: auth-agent, started with only the plist environment, served the OAuth token from the configured directory and answered the IAM request with the `yc` lookup error.
- SCN.07 and SCN.08 on Windows: the installer is checked by reading its code.
- SCN.09: builds of v1.0.4 and of this change run as auth-agent, and requests in both formats go to each of them.

### Changes outside the feature

- `go.mod`, CI and release workflows raise Go to 1.27.4, because govulncheck reports standard library advisories for 1.27.2.
- The CI and release workflows drop their separate installation of golangci-lint and ifaceguard. `task lint` runs both through `go tool`, so the versions in `go.mod` are the only linter versions.
- `AGENTS.md` requires a specification in `docs/specs/{feature or bug name}/` for each change: `problem.md`, `requirements.md`, `solution.md`, and `glossary.md` when the specification introduces terms. `AGENTS.md` no longer forbids backward compatibility.

## Overengineering and Overspecification Considerations

- S7 default token type: NRQ.01 requires mixed-version IAM acquisition, and the released auth-agent rejects unknown request members.
- S4 profile name restriction: profile names come from MCP requests over the tunnel, and the restriction keeps a request inside the OAuth directory.
- S8 validation of an explicit `yc` path and a relative OAuth directory: both come from user input at installation, and auth-agent would otherwise fail at each startup.
- The installers compute no default OAuth directory; S8 names the cost.
- Left to the implementer: error texts, the structure of tests, and the way each installer writes an empty value to its startup configuration.

## Open Questions

There are no open questions.

## References

- [Problem statement](problem.md) and [requirements](requirements.md)
- [Remote authentication solution](../remote-authentication/solution.md): design of remote IAM acquisition before this change: tunnel, endpoint, HTTP statuses, shared waiting
- [Tracker API access](https://yandex.ru/support/tracker/en/concepts/access) and [Wiki API access](https://yandex.ru/support/wiki/en/api-ref/access): token types and organization headers
- [Pull request 15](https://github.com/n-r-w/yandex-mcp/pull/15)
