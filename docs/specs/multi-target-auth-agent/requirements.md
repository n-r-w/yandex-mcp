# Idea: One auth-agent for several servers

## Definitions

Terms are defined in the [problem statement](problem.md#key-definitions-and-abbreviations).

## Context and Problem

See the [problem statement](problem.md).

## Goal

MCP on each server of the user receives tokens from one auth-agent installation on the workstation, and installations of v1.0.1 to v1.1.0 keep working after an upgrade of auth-agent.

## Affected Scenarios

- An agent calls a Tracker or Wiki tool on any of several servers that use one workstation.
- The connection to one of the servers breaks and recovers.
- The user installs auth-agent for several servers and restarts the workstation.
- The user replaces the auth-agent executable of an earlier installation without running the installer again.

## Scope and Non-Scope

The scope includes auth-agent on the workstation, its installers for Linux, macOS, and Windows, compatibility with startup configurations of v1.0.1 to v1.1.0, and the user documentation in README and `.env.example`.

The scope excludes MCP on the server, different auth-agent ports for different servers, and different `yc` executables or OAuth directories for different servers.

## Requirements

### Functional Requirements

- [X] FRQ.01: One auth-agent serves token requests of MCP on each server that the user lists at installation, also after a restart of the workstation.
  - Origin: formulated. PRB.01.
  - Goal: Tools on each server receive tokens from one workstation.
  - Feasibility: Confirmed. Each server listens on its own loopback address, so reverse tunnels from one workstation use the same port on every server.
- [X] FRQ.02: When the connection to one server breaks, auth-agent keeps serving the other servers and restores the broken connection without user action.
  - Origin: formulated. PRB.01.
  - Goal: A network problem of one server does not stop tools on the other servers.
  - Feasibility: Confirmed.
- [X] FRQ.03: Auth-agent refuses to start when the list of servers holds an empty entry or the same server twice, with an error that names the setting.
  - Origin: formulated. Constraint for a mistyped list.
  - Goal: The user sees the typing error at startup.
  - Feasibility: Confirmed.
- [X] FRQ.04: README and `.env.example` describe how to list several servers and state that every server uses the port of the workstation.
  - Origin: formulated. PRB.01.
  - Goal: The user configures several servers from the documentation.
  - Feasibility: Confirmed.

### Non-Functional Requirements

- [X] NRQ.01: Auth-agent with this change starts with a startup configuration that an installer of v1.0.1 to v1.1.0 wrote and serves the server of that configuration. When a startup configuration holds both the list of servers and the single server of an earlier version, auth-agent uses the list.
  - Origin: formulated. Constraint for existing installations.
  - Goal: Replacing the auth-agent executable does not break remote authentication.
  - Feasibility: Confirmed. The earlier setting `YANDEX_MCP_SSH_TARGET` has a different name than the list.

## Justification of Requirements

| What does not work now | Requirements that fix it |
|---|---|
| PRB.01: auth-agent connects to one server, and a second installation replaces the first one | FRQ.01, FRQ.02, FRQ.04 |
| Constraint: a mistyped list of servers stops at startup | FRQ.03 |
| Constraint: installations of v1.0.1 to v1.1.0 keep working after the executable is replaced | NRQ.01 |

Sufficiency: each row has requirements.
Necessity: each requirement stands in a row.

## Critical Scenarios

- SCN.01: Two servers at the same time
  - Requirements: FRQ.01
  - Level: external
  - Area: auth-agent, SSH
  - Flow: Auth-agent started with two servers opens a tunnel to each of them, and both tunnels stay open at the same time.
- SCN.02: One connection breaks
  - Requirements: FRQ.02
  - Level: external
  - Area: auth-agent, SSH
  - Flow: The tunnel to one server ends. Auth-agent opens it again, and the tunnel to the other server stays open.
- SCN.03: Invalid list of servers
  - Requirements: FRQ.03
  - Level: external
  - Area: auth-agent
  - Flow: Startup with an empty entry or a repeated server in the list ends with an error that names `YANDEX_MCP_SSH_TARGETS`.
- SCN.04: Startup configuration of an earlier version
  - Requirements: NRQ.01
  - Level: external
  - Area: auth-agent
  - Flow: Auth-agent started with only `YANDEX_MCP_SSH_TARGET` connects to that server. Auth-agent started with both settings connects to the servers of `YANDEX_MCP_SSH_TARGETS`.
- SCN.05: Installation for several servers
  - Requirements: FRQ.01 (Linux, macOS)
  - Level: external
  - Area: installer, startup configuration
  - Flow: The installer writes the list of servers into the startup configuration, and auth-agent started with only this configuration connects to each server.
- FRQ.01, Windows: no scenario. No PowerShell is available for the check; the Windows installer is checked by reading its code.
- FRQ.04: no scenario. Documentation is checked by comparing its text with the code.

## Open Questions

There are no open questions.

## Technical Supplement

No technical supplement.

## References

- [Problem statement](problem.md)
- [Remote authentication requirements](../remote-authentication/requirements.md): requirements of remote authentication with one server
- [Yandex 360 OAuth requirements](../yandex-360-oauth/requirements.md): requirements of OAuth tokens through auth-agent
