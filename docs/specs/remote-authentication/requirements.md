# Remote Yandex MCP authentication requirements

## Definitions

Terms are defined in the [glossary](glossary.md).

## Context and problem

The [problem statement](problem.md) gives the basis for these requirements.

## Goal

The user completes login on the workstation without repeated manual setup for token acquisition or connection recovery.

## Key scenarios

- Token acquisition with an active federated session.
- Initial user login and reauthentication on the workstation.
- Concurrent tool calls from multiple MCP processes.
- Recovery after workstation sleep or connection loss.

## Scope and non-scope

The scope includes remote authentication, its initial setup, automatic startup, connection recovery, and one overall tool timeout. Workstations running macOS, Linux with a graphical desktop, or Windows are supported. The agent and remote-mode MCP run on a Linux server. Local MCP execution also remains available.

The scope excludes password and MFA automation, running ordinary server-side `yc` commands through the workstation, and indefinite operation without the workstation.

## Requirements

### Functional requirements

- **FRQ-01:** In remote mode, IAM token acquisition, initial login, and reauthentication take place on the workstation. The Linux server does not require `yc` installation.
  - Goal: Use the browser on the user's workstation.
  - Goal achievement: Full. The server does not participate in interactive login.
- **FRQ-02:** A tool call waits for authentication within its overall timeout. After successful login, execution continues without another tool call.
  - Goal: Continue the task after login.
  - Goal achievement: Full when the call completes before its deadline.
- **FRQ-03:** Concurrent requests for one profile share one token acquisition, including requests from different MCP processes.
  - Goal: Prevent multiple concurrent login operations.
  - Goal achievement: Full. Concurrent requests do not open extra browser windows.
- **FRQ-04:** Cancellation or timeout ends that call's wait. Token acquisition stops only after the last waiter leaves.
  - Goal: Stop work that has no waiter without disrupting other calls.
  - Goal achievement: Full. Cancellation accounts for all waiters.
- **FRQ-05:** After initial setup, remote authentication programs start when the user logs in to the workstation's graphical desktop session. After sleep or network loss, the connection recovers without manual tunnel creation.
  - Goal: Remove repeated technical actions by the user.
  - Goal achievement: Full when the workstation, network, and server SSH are available.
- **FRQ-06:** Local mode obtains tokens through local `yc` without the remote authentication programs.
  - Goal: Preserve the local usage scenario.
  - Goal achievement: Full. Moving the agent to a server is not mandatory.
- **FRQ-07:** Errors distinguish workstation or connection unavailability, authentication failure, access denial, protocol incompatibility, and timeout. A remote-mode error does not start `yc` on the server.
  - Goal: Explain why execution stopped without silently changing the authentication method.
  - Goal achievement: Full. The user receives the failure reason.

### Non-functional requirements

- **NRQ-01:** Every tool call has one configurable overall timeout, defaulting to 300 seconds. Token acquisition, login, and API requests are included. Lower layers do not restart the timeout.
  - Goal: Limit the full duration of a call.
  - Goal achievement: Full. Waiting for authentication does not bypass the overall deadline.
- **NRQ-02:** Remote token acquisition is permitted only for a client with a valid connection secret and only for an explicitly specified allowed profile.
  - Goal: Restrict access to user tokens.
  - Goal achievement: Full. Access to the server's `localhost` alone does not grant permission.
- **NRQ-03:** Federated credentials remain on the workstation. IAM tokens, connection secrets, login URLs, and raw `yc` output do not appear in logs.
  - Goal: Prevent disclosure of authentication data.
  - Goal achievement: Partial. Transfer and storage protection are also necessary.

## Open questions

There are no blocking questions about the requirements.

## Related documents

The [technical solution](solution.md) defines how these requirements are met.
