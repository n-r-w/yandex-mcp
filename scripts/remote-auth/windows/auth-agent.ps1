# Native stderr contains normal slog output, not PowerShell failures.
$ErrorActionPreference = 'Continue'
$env:YANDEX_MCP_AUTH_AGENT_LISTEN = '@LISTEN@'
$env:YANDEX_MCP_AUTH_AGENT_YC_PATH = '@YC_PATH@'
$env:YANDEX_MCP_AUTH_AGENT_PROFILES = '@PROFILES@'
& '@BINARY@' auth-agent >> "$env:LOCALAPPDATA\YandexMCP\auth-agent.log" 2>&1
exit $LASTEXITCODE
