# Problem Statement

## Key definitions and abbreviations

Terms are defined in the [glossary](glossary.md).

## Context

`yandex-mcp` gives an agent read access to Yandex Tracker and Yandex Wiki. MCP runs either on the user's machine or on a Linux server; in the second case auth-agent acquires tokens on the user's workstation, as the [remote authentication problem](../remote-authentication/problem.md) describes.

## Observed Problem

- PRM.01: A member of a Yandex 360 organization cannot read Tracker issues and Wiki pages of the organization through `yandex-mcp`. MCP adds an IAM token as `Authorization: Bearer` and the organization ID as `X-Cloud-Org-Id` to every API request. Tracker and Wiki of a Yandex 360 organization accept an OAuth token as `Authorization: OAuth` with the organization ID in `X-Org-Id`.
  - PRB.01: MCP runs on the user's machine. The agent calls a tool. MCP acquires the token only through `yc` and sends it with the Yandex Cloud headers, so the API of the Yandex 360 organization rejects the request.
  - PRB.02: MCP runs on a server, and the user's OAuth token is stored on the workstation. MCP requests the token from auth-agent, but auth-agent acquires only IAM tokens through `yc`, so the OAuth token never reaches the server.

## Affected Audience

Members of Yandex 360 organizations with a non-federated account who use `yandex-mcp` locally or through remote authentication.

## Evidence

- `internal/adapters/apihelpers/client.go` on `main` sets `Authorization: Bearer <token>` and `X-Cloud-Org-Id` on every request, and `internal/config/config.go` on `main` requires `YANDEX_CLOUD_ORG_ID`.
- Per the description and the review of pull request 15, the [Tracker API access guide](https://yandex.ru/support/tracker/en/concepts/access) and the [Wiki API access guide](https://yandex.ru/support/wiki/en/api-ref/access) state that a Yandex 360 organization requires an OAuth token and the `X-Org-Id` header, and direct federated accounts to IAM tokens; the guides were not reread for this document.
- Per the description of pull request 15, `tracker_user_current`, `tracker_issue_get`, and `wiki_page_get` returned data for a real Yandex 360 organization with a local OAuth token; not checked independently.

## Impact

HIGH: members of Yandex 360 organizations cannot use any Tracker or Wiki tool of `yandex-mcp`.

## Current State

MCP and auth-agent acquire only IAM tokens through `yc`. MCP always sends the `Bearer` scheme and the `X-Cloud-Org-Id` header and refuses to start without `YANDEX_CLOUD_ORG_ID`.

## Desired State

A member of a Yandex 360 organization calls Tracker and Wiki tools in the same way as a member of a Yandex Cloud organization, with MCP on the user's machine or on a server.

## Problem Boundary

The problem covers Tracker and Wiki access for members of Yandex 360 organizations through local and remote token acquisition. Federated accounts of Yandex 360 organizations are outside the problem: the API guides direct them to IAM tokens. Issuing and renewing OAuth tokens is outside the problem: the user issues a token in Yandex OAuth.

## Assumptions

No assumptions.

## Open Questions

There are no open questions about the problem statement.
