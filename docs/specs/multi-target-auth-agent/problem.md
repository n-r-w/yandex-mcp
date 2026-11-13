# Problem Statement

## Key definitions and abbreviations

The terms auth-agent, workstation, MCP, remote authentication, and startup configuration are defined in the [Yandex 360 OAuth glossary](../yandex-360-oauth/glossary.md). A server is a machine that runs MCP with remote authentication.

## Context

With remote authentication, MCP on a server requests tokens from auth-agent on the workstation through a reverse SSH tunnel, as the [remote authentication problem](../remote-authentication/problem.md) describes. The user installs auth-agent with the installer for the operating system of the workstation.

## Observed Problem

- PRM.01: A user who runs MCP with remote authentication on several servers receives workstation tokens on only one of these servers through the installed auth-agent.
  - PRB.01: The user installs auth-agent for a second server. The agent calls a tool on the first server, and the tool returns an auth-agent connection error. Examples: `tracker_issue_get` on each of two servers that use one workstation.

## Affected Audience

Users of remote authentication whose agents run MCP on more than one server.

## Evidence

- `internal/server/authagent/config.go` in v1.1.0 reads one SSH destination from `YANDEX_MCP_SSH_TARGET`, and `internal/server/authagent/run.go` in v1.1.0 starts one SSH tunnel to it.
- The installers in v1.1.0 create one startup entry with a fixed name: `net.yandex-mcp.auth-agent` in `scripts/remote-auth/macos/install.sh`, `yandex-mcp-auth-agent.service` in `scripts/remote-auth/linux/install.sh`, and `Yandex MCP auth-agent` in `scripts/remote-auth/windows/install.ps1`. A second installation replaces the first entry.

## Impact

MEDIUM: on each server except one, the user either loses Tracker and Wiki tools with remote authentication or maintains a second auth-agent by hand.

## Current State

Auth-agent keeps one SSH tunnel to the server from `YANDEX_MCP_SSH_TARGET`. The installers keep one auth-agent startup entry per workstation user. A second auth-agent works only when the user starts it by hand with another `YANDEX_MCP_AUTH_AGENT_PORT` and sets the same port on the second server.

## Desired State

One auth-agent installation serves token requests of MCP on each server of the user.

## Problem Boundary

The problem covers token delivery from one workstation to MCP on several servers. Token acquisition on the workstation and MCP on the server are outside the problem: they work the same way for each server.

## Assumptions

No assumptions.

## Open Questions

No open questions.
