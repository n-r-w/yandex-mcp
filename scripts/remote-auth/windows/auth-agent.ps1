# Native stderr contains normal slog output, not PowerShell failures.
$ErrorActionPreference = 'Continue'
$env:YANDEX_MCP_AUTH_AGENT_PORT = '@PORT@'
$env:YANDEX_MCP_AUTH_AGENT_YC_PATH = '@YC_PATH@'
$env:YANDEX_MCP_SSH_TARGET = '@SSH_TARGET@'
$env:PATH = '@PATH@'
& '@BINARY@' auth-agent >> "$env:LOCALAPPDATA\YandexMCP\auth-agent.log" 2>&1
exit $LASTEXITCODE
