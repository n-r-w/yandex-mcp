# Technical Solution: One auth-agent for several servers

## Problem Statement

The [problem statement](problem.md) describes the problem, and the [requirements](requirements.md) define the expected behavior.

- Auth-agent installed by the installers serves MCP on only one server of the user.

## Proposed Solution

Auth-agent reads a list of servers and keeps one reverse SSH tunnel to each of them from one loopback listener. The installers save the list in the startup configuration. Auth-agent still reads the single server of startup configurations from earlier versions.

### S1. Tunnels to several servers (FRQ.01, FRQ.02)

- `YANDEX_MCP_SSH_TARGETS` holds a comma-separated list of SSH host aliases or `user@host` destinations. Auth-agent trims spaces around each destination.
- Auth-agent listens on `127.0.0.1` with `YANDEX_MCP_AUTH_AGENT_PORT` and starts one SSH process with the same reverse forwarding for each destination. Each server has its own loopback address, so every server uses the port of the workstation.
- Each tunnel restarts on its own 5 seconds after its SSH process ends. The end of one tunnel does not stop the other tunnels. Stopping auth-agent stops all SSH processes and waits for them.
- Log records and errors of a tunnel name its destination.
- Trade-off: all servers share one port, one `yc` executable, and one OAuth directory. Cost: a server whose port is taken by another program cannot use auth-agent without a change of the port on the workstation and on every other server.

### S2. Validation of the list (FRQ.03)

Auth-agent refuses to start when the list is empty, holds an empty entry, or holds a destination twice. The error names `YANDEX_MCP_SSH_TARGETS`. Two equal destinations would start two SSH processes that forward the same server port, and the second process would fail and restart every 5 seconds.

### S3. Startup configurations of earlier versions (NRQ.01)

- When `YANDEX_MCP_SSH_TARGETS` is unset or empty, auth-agent takes `YANDEX_MCP_SSH_TARGET` as the only destination. Example: a startup configuration of v1.1.0 with `YANDEX_MCP_SSH_TARGET=user@server` gives one tunnel to `user@server`.
- When both settings are set, auth-agent uses `YANDEX_MCP_SSH_TARGETS` and ignores `YANDEX_MCP_SSH_TARGET`. Example: `YANDEX_MCP_SSH_TARGETS=a,b` and `YANDEX_MCP_SSH_TARGET=c` give tunnels to `a` and `b`.
- When neither is set, the error names `YANDEX_MCP_SSH_TARGETS`.
- Auth-agent does not split the value of `YANDEX_MCP_SSH_TARGET` by commas, because earlier versions passed it to SSH as one destination.
- Trade-off: auth-agent keeps reading a setting that the installers no longer write. Cost: one more setting in the configuration code. A reinstall removes the old setting, because each installer writes the startup configuration anew.

### S4. Installation (FRQ.01)

The installers for Linux, macOS, and Windows require `YANDEX_MCP_SSH_TARGETS` and save its value in the startup configuration. They keep one startup entry per workstation user, so a reinstall with a longer list replaces the entry of the shorter one.

### S5. User documentation (FRQ.04)

README shows the installation with two servers, states that every server uses the port of the workstation, and states that auth-agent still reads `YANDEX_MCP_SSH_TARGET` of earlier installations. `.env.example` describes `YANDEX_MCP_SSH_TARGETS`.

### Verification

- SCN.01: `TestMultipleTunnels` in `internal/server/authagent/ssh_test.go` starts tunnels to two destinations with a test SSH executable and checks that both SSH processes run at the same time and stop after cancellation.
- SCN.02: `TestTunnelRestartIsIndependent` in `internal/server/authagent/ssh_test.go` lets the test SSH executable of one destination fail and checks that auth-agent starts it again while the SSH process of the other destination runs without a restart.
- SCN.03 and SCN.04: `TestLoadConfig` in `internal/server/authagent/config_test.go` covers an empty list, an empty entry, a repeated destination, the single destination of an earlier version, and both settings at once.
- SCN.05 on macOS: `test_first_install` in `scripts/remote-auth/macos/test_install.py` checks that the installer writes the list of servers to the plist. A real installation on a macOS workstation with two servers checks that Tracker and Wiki tools return data on both servers.
- SCN.05 on Linux and Windows: the installers are checked by reading their code.

### Changes outside the feature

- The macOS installer runs `launchctl bootstrap` up to three times with a pause of one second, because on real workstations this call sometimes failed.

## Overengineering and Overspecification Considerations

- S2 validation: the list comes from user input at installation, and a mistake would otherwise show only as repeated SSH errors in the log.
- S3 reading of the old setting: NRQ.01 requires that a replaced executable works with an existing startup configuration.
- Different ports, `yc` executables, or OAuth directories per server are outside the scope.
- Left to the implementer: error texts and the structure of tests.

## Open Questions

There are no open questions.

## References

- [Problem statement](problem.md) and [requirements](requirements.md)
- [Remote authentication solution](../remote-authentication/solution.md): tunnel, endpoint, and SSH options for one server
- [Yandex 360 OAuth solution](../yandex-360-oauth/solution.md): OAuth tokens through auth-agent and installation without `yc`
