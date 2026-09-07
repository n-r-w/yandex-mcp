# Remote Yandex MCP authentication problem

## Context

The agent and `yandex-mcp` run on a Linux server through SSH or Orca. The user's browser runs on a separate workstation.

Terms are defined in the [glossary](glossary.md).

## Observed problem

Initial federated login and reauthentication through `yc` require a browser. When `yc` runs on the Linux server, the user must manually connect the server-side login process to the workstation browser. Installing `yc` on the server does not solve this problem.

## Affected audience

Users with a macOS, Linux, or Windows workstation who run the agent on a Linux server through SSH or Orca.

## Evidence

- The user reports that the local Mac scenario works. The server-side problem occurs during initial login, not only after the federated session expires.
- The user observes that reauthentication is required at least once a day. This frequency has not been independently reproduced.
- The [README authentication section](../../../README.md#authentication) describes token acquisition through `yc` and possible browser launch during refresh.

## Impact

Agent tasks that require Wiki or Tracker are interrupted. The user manually transfers the login URL and sets up the callback connection. The duration of these interruptions has not been measured.

## Current state

`yandex-mcp` obtains and caches an IAM token through `yc` on the machine that runs MCP. One-time profile setup does not remove the need for later interactive login.

## Desired state

The user continues to work with the agent on the Linux server. After initial setup, the only required authentication action is login in the workstation browser, without manual URL transfer or callback connection setup.

## Problem boundary

The problem covers the manual technical steps that connect the server-side process to the workstation browser for initial login and reauthentication. The need to enter a password or complete MFA is not the problem to remove.

## Open questions

There are no open questions about the problem statement.
