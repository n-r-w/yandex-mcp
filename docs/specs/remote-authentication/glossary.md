# Remote Yandex MCP authentication glossary

- Workstation: the user's computer with an interactive desktop session and a browser.
- Agent: a program on the Linux server that performs user tasks and calls MCP tools.
- Yandex MCP server: the `yandex-mcp` program that provides Yandex Wiki and Yandex Tracker tools to the agent.
- MCP tool: an operation that the MCP server exposes for the agent to call.
- IAM token: a temporary access token for Yandex APIs.
- Federated session: an authenticated state that lets `yc` obtain IAM tokens for a federated user without another interactive login.
- yc profile: a named Yandex Cloud CLI configuration.
- Reauthentication: interactive user login after a federated session expires, including MFA when required.
- Auth-agent: the `yandex-mcp auth-agent` mode that obtains tokens on the workstation: IAM tokens through `yc`, OAuth tokens from OAuth profile files.
- Token source: a component that obtains a new token of one token type for the token provider or auth-agent: `yc` for IAM tokens, an OAuth token file or the OAuth directory for OAuth tokens, or auth-agent in remote mode.
- Token provider: the MCP component that caches a token of the configured token type and manages its refresh.
- Waiter: a call that waits for a shared token acquisition result and has its own cancellation context.
- Shared token acquisition: one acquisition of a token of one token type and profile whose result multiple calls can wait for.

Token type, OAuth profile, OAuth directory, and startup configuration are defined in the [Yandex 360 OAuth glossary](../yandex-360-oauth/glossary.md).
