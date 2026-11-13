# Yandex 360 OAuth support glossary

- Yandex Cloud organization: an organization of Yandex Cloud Organization; Tracker and Wiki identify it by the `X-Cloud-Org-Id` header.
- Yandex 360 organization: an organization of Yandex 360 for Business; Tracker and Wiki identify it by the `X-Org-Id` header.
- Token type: the kind of token that MCP sends to Tracker and Wiki: an IAM token or a Yandex OAuth token.
- Acquisition location: the machine where the token is acquired: the machine that runs MCP, or the workstation through auth-agent.
- OAuth profile: a named file that holds one Yandex OAuth token; auth-agent reads it by the profile name that MCP sends.
- OAuth directory: the workstation directory that holds the OAuth profile files of auth-agent.
- Startup configuration: the file from which the operating system passes the environment to auth-agent at startup: `auth-agent.env` on Linux, the LaunchAgent plist on macOS, the generated `auth-agent.ps1` on Windows.

Workstation, auth-agent, token source, and token provider are defined in the [remote authentication glossary](../remote-authentication/glossary.md).
