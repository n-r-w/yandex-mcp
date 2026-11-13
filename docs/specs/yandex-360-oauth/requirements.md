# Idea: Yandex 360 organizations with OAuth tokens

## Definitions

Terms are defined in the [glossary](glossary.md).

## Context and Problem

The [problem statement](problem.md) describes the problem: members of Yandex 360 organizations cannot use `yandex-mcp`, because MCP and auth-agent work only with IAM tokens and Yandex Cloud organizations.

## Goal

A member of a Yandex 360 organization uses Tracker and Wiki tools with MCP on the user's machine or on a server, and members of Yandex Cloud organizations keep working with their current settings.

## Affected Scenarios

- An agent calls a Tracker or Wiki tool with MCP on the user's machine.
- An agent calls a Tracker or Wiki tool with MCP on a server and auth-agent on the workstation.
- The user installs auth-agent on the workstation and restarts the workstation.
- The user upgrades `yandex-mcp` on the server and on the workstation at different times.

## Scope and Non-Scope

The scope includes the choice of the token type, local and remote OAuth tokens, Yandex 360 organizations, auth-agent installation for OAuth, compatibility with released versions v1.0.1 to v1.0.4, and the user documentation in README and `.env.example`.

The scope excludes issuing and renewing OAuth tokens, federated accounts of Yandex 360 organizations, and write access to Tracker and Wiki.

## Requirements

### Functional Requirements

- [X] FRQ.01: The user selects the token type, IAM or OAuth, independently of the acquisition location, the machine that runs MCP or auth-agent, and each of the four combinations delivers a token of the selected type.
  - Origin: source. PRB.01, PRB.02. Review of pull request 15: `YANDEX_MCP_TOKEN_SOURCE=local|remote` and `YANDEX_MCP_TOKEN_TYPE=iam|oauth`.
  - Goal: An OAuth token is available in both scenarios.
  - Feasibility: Confirmed. Token sources for both types exist on both machines.
- [X] FRQ.02: With an OAuth token, MCP sends Tracker and Wiki requests with the `OAuth` scheme; with an IAM token, with the `Bearer` scheme. For a Yandex 360 organization MCP sends the organization ID in `X-Org-Id`, for a Yandex Cloud organization in `X-Cloud-Org-Id`. MCP refuses to start when a Yandex 360 organization is configured with the IAM token type, when no organization is configured, or when both are configured.
  - Origin: source. PRB.01, PRB.02. Review of pull request 15: the scheme and the organization compatibility follow the token type.
  - Goal: The API of a Yandex 360 organization accepts MCP requests.
  - Feasibility: Confirmed.
- [X] FRQ.03: When MCP acquires an OAuth token on its own machine, it reads the token from the file that the user configured.
  - Origin: formulated. PRB.01.
  - Goal: An OAuth token for MCP on the user's machine.
  - Feasibility: Confirmed.
- [X] FRQ.04: Auth-agent serves OAuth tokens of several OAuth profiles whose files stay on the workstation. MCP sends the profile name, so different MCP connections can use different accounts.
  - Origin: source. PRB.02. Review of pull request 15: one auth-agent supports multiple named OAuth files.
  - Goal: An OAuth token for MCP on a server.
  - Feasibility: Confirmed.
- [X] FRQ.05: When the token of the requested type cannot be acquired, for example because of a missing file, an unknown OAuth profile, a `yc` failure, or an unsupported token type, the tool returns an error with the cause. Neither MCP nor auth-agent tries another acquisition method.
  - Origin: source. PRB.01, PRB.02. Review of pull request 15: auth-agent must not substitute another acquisition method.
  - Goal: The user sees the actual cause and never receives a token of another type.
  - Feasibility: Confirmed.
- [X] FRQ.06: Concurrent requests for a token of one type and one profile wait for one acquisition. Requests for IAM profile `work` and OAuth profile `work` acquire their tokens independently.
  - Origin: source. PRB.02. Review of pull request 15: shared acquisitions distinguish token types.
  - Goal: Auth-agent never answers an OAuth request with an IAM token of the same profile name.
  - Feasibility: Confirmed.
- [X] FRQ.07: On a workstation without `yc`, auth-agent installs on Linux, macOS, and Windows and serves OAuth tokens. Such auth-agent answers an IAM request with an error that states that `yc` is not found.
  - Origin: source. PRB.02. Review of pull request 15, item "OAuth-only installation still requires `yc`".
  - Goal: OAuth on the workstation needs no `yc`.
  - Feasibility: Confirmed. The OAuth profile source does not run `yc`.
- [X] FRQ.08: Auth-agent uses the OAuth directory that the user set at installation after each restart.
  - Origin: source. PRB.02. Review of pull request 15, item "The selected OAuth directory must survive subsequent launches", for an explicitly set directory.
  - Goal: OAuth tokens remain available after the workstation restarts.
  - Feasibility: Confirmed. Each supported operating system has a startup configuration through which the installers already pass environment variables.
- [X] FRQ.09: README and `.env.example` describe the token type choice, the four combinations, the OAuth setup on the workstation, and the OAuth prerequisites: a non-federated account, API access enabled in the organization, and the permissions of the user.
  - Origin: formulated. PRB.01, PRB.02.
  - Goal: The user configures OAuth from the documentation.
  - Feasibility: Confirmed.

### Non-Functional Requirements

- [X] NRQ.01: Members of Yandex Cloud organizations keep working with their current local and remote IAM settings. Remote IAM acquisition works when MCP on the server and auth-agent on the workstation have different versions: one of them v1.0.1 to v1.0.4, the other with this change.
  - Origin: formulated. Constraint for existing users of released versions.
  - Goal: An upgrade does not break existing users.
  - Feasibility: Confirmed. Auth-agent v1.0.1 to v1.0.4 accepts a request that holds only the profile.

## Justification of Requirements

| What does not work now | Requirements that fix it |
|---|---|
| PRB.01: MCP acquires only an IAM token and sends the Yandex Cloud headers | FRQ.01 (local OAuth), FRQ.02, FRQ.03, FRQ.05, FRQ.09 |
| PRB.02: auth-agent acquires only an IAM token through `yc` | FRQ.01 (remote OAuth), FRQ.02, FRQ.04, FRQ.05, FRQ.06, FRQ.07, FRQ.08, FRQ.09 |
| Constraint: existing IAM users keep working | NRQ.01, FRQ.01 (local IAM, remote IAM) |

Sufficiency: each row has requirements.
Necessity: each requirement stands in a row.

## Critical Scenarios

- SCN.01: Local OAuth for a Yandex 360 organization
  - Requirements: FRQ.01 (local OAuth), FRQ.02, FRQ.03
  - Level: external
  - Area: MCP, Tracker and Wiki API
  - Flow: The agent calls a Tracker or Wiki tool. MCP reads the token from the file and sends the request with the `OAuth` scheme and `X-Org-Id`. The tool returns data.
- SCN.02: Remote OAuth with two profiles
  - Requirements: FRQ.01 (remote OAuth), FRQ.04
  - Level: external
  - Area: MCP, auth-agent, OAuth profile files
  - Flow: Two MCP connections with profiles `work` and `personal` acquire, through one auth-agent, the tokens from the matching files.
- SCN.03: IAM for a Yandex Cloud organization, local and remote
  - Requirements: FRQ.01 (local IAM, remote IAM), FRQ.02, NRQ.01
  - Level: external
  - Area: MCP, auth-agent, `yc`
  - Flow: MCP with the current settings acquires an IAM token through `yc` and sends the request with the `Bearer` scheme and `X-Cloud-Org-Id`.
- SCN.04: Invalid organization settings
  - Requirements: FRQ.02
  - Level: external
  - Area: MCP
  - Flow: Startup with a Yandex 360 organization and the IAM token type, without an organization, or with both organizations ends with a configuration error.
- SCN.05: Failure without another acquisition method
  - Requirements: FRQ.05
  - Level: external
  - Area: MCP, auth-agent
  - Flow: With a missing file, an unknown profile, a `yc` failure, or an unsupported token type, the tool returns an error with the cause, and no other token source runs.
- SCN.06: Equal profile names of different token types
  - Requirements: FRQ.06
  - Level: external
  - Area: auth-agent
  - Flow: Concurrent requests for IAM profile `work` and OAuth profile `work` each receive a token of their own type.
- SCN.07: Installation without `yc`
  - Requirements: FRQ.07 (Linux, macOS)
  - Level: external
  - Area: installer, startup configuration, auth-agent
  - Flow: The installer on a machine without `yc` ends successfully. Auth-agent serves an OAuth token and answers an IAM request with an error that `yc` is not found.
- SCN.08: OAuth directory after a restart
  - Requirements: FRQ.08
  - Level: external
  - Area: installer, startup configuration, auth-agent
  - Flow: Auth-agent is installed with the OAuth directory `/absolute/custom-oauth`. After a restart with only the environment of the startup configuration, profile `work` is read from `/absolute/custom-oauth/work`.
- SCN.09: Different versions of MCP and auth-agent
  - Requirements: NRQ.01
  - Level: external
  - Area: MCP, auth-agent
  - Flow: MCP v1.0.4 with the new auth-agent and the new MCP with auth-agent v1.0.4 acquire an IAM token. The new MCP requesting an OAuth token from auth-agent v1.0.4 returns an error with the auth-agent response.
- FRQ.07, Windows: no scenario. No PowerShell is available for the check; the Windows installer is checked by reading its code.
- FRQ.09: no scenario. Documentation is checked by comparing its text with the code.

## Open Questions

There are no open questions.

## Technical Supplement

No technical supplement.

## References

- [Problem statement](problem.md)
- [Technical solution](solution.md)
- [Remote authentication requirements](../remote-authentication/requirements.md)
- [Pull request 15](https://github.com/n-r-w/yandex-mcp/pull/15)
