# Problem Statement

## Key definitions and abbreviations

Terms are defined in the [glossary](glossary.md).

## Context

`yandex-mcp` gives an agent read access to Yandex Tracker and Yandex Wiki. MCP acquires tokens either on the machine that runs MCP or, when MCP runs on a Linux server, through auth-agent on the workstation, as the [remote authentication problem](../remote-authentication/problem.md) describes.

## Observed Problem

- PRM.01: A member of a Yandex 360 organization cannot read Tracker issues and Wiki pages of the organization through `yandex-mcp`. Section Current State describes the cause.
  - PRB.01: MCP acquires the token on the machine that runs MCP. The agent calls a tool, and the API of the Yandex 360 organization rejects the request of MCP. Examples: `tracker_issue_get` and `wiki_page_get` for a Yandex 360 organization with the token acquired on the machine that runs MCP.
  - PRB.02: MCP runs on a server, and the OAuth token of the user is stored on the workstation. MCP requests a token from auth-agent, and the OAuth token does not reach the server. Examples: `tracker_issue_get` and `wiki_page_get` for a Yandex 360 organization with MCP on a server and the OAuth token on the workstation.

## Affected Audience

Members of Yandex 360 organizations with a non-federated account who use `yandex-mcp` locally or through remote authentication.

## Evidence

- `internal/adapters/apihelpers/client.go` in v1.0.4 sets `Authorization: Bearer <token>` and `X-Cloud-Org-Id` on every request, and `internal/config/config.go` in v1.0.4 requires `YANDEX_CLOUD_ORG_ID`.
- The [Tracker API access guide](https://yandex.ru/support/tracker/en/concepts/access) and the [Wiki API access guide](https://yandex.ru/support/wiki/en/api-ref/access) state that OAuth tokens work in Yandex 360 and Yandex Cloud organizations, and IAM tokens only in Yandex Cloud organizations. Both guides direct federated accounts to IAM tokens. The Wiki guide states that `X-Org-Id` identifies a Yandex 360 organization.

## Impact

HIGH: members of Yandex 360 organizations cannot use any Tracker or Wiki tool of `yandex-mcp`.

## Current State

MCP and auth-agent acquire only IAM tokens through `yc`. MCP always sends the `Bearer` scheme and the `X-Cloud-Org-Id` header and refuses to start without `YANDEX_CLOUD_ORG_ID`.

## Desired State

A member of a Yandex 360 organization calls Tracker and Wiki tools in the same way as a member of a Yandex Cloud organization, with the token acquired on the machine that runs MCP or through auth-agent.

## Problem Boundary

The problem covers Tracker and Wiki access for members of Yandex 360 organizations through local and remote token acquisition. Federated accounts are outside the problem: the API guides direct them to IAM tokens, which only Yandex Cloud organizations accept. Issuing and renewing OAuth tokens is outside the problem: the user issues a token in Yandex OAuth.

## Assumptions

No assumptions.

## Open Questions

No open questions.
