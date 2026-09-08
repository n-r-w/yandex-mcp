#!/bin/sh
# Run in a logged-in macOS graphical session. No root privileges are needed.
set -eu
: "${YANDEX_MCP_SSH_TARGET:?Set the SSH host or user@host}"
port=${YANDEX_MCP_AUTH_AGENT_PORT:-18765}
case "$port" in ''|*[!0-9]*) echo 'Port must be an integer' >&2; exit 1 ;; esac
[ "$port" -ge 1 ] && [ "$port" -le 65535 ] || exit 1
binary=$(command -v "${YANDEX_MCP_BINARY:-yandex-mcp}") || { echo 'yandex-mcp not found; set YANDEX_MCP_BINARY' >&2; exit 1; }
yc_path=$(command -v "${YANDEX_MCP_AUTH_AGENT_YC_PATH:-yc}") || { echo 'yc not found; set YANDEX_MCP_AUTH_AGENT_YC_PATH' >&2; exit 1; }
command -v ssh >/dev/null || { echo 'Install OpenSSH and add ssh to PATH' >&2; exit 1; }
for executable in "$binary" "$yc_path"; do
  case "$executable" in /*) ;; *) echo 'Executable paths must resolve to absolute paths' >&2; exit 1 ;; esac
  [ -f "$executable" ] && [ -x "$executable" ] || { echo "Not executable: $executable" >&2; exit 1; }
done
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
startup_dir="$HOME/Library/LaunchAgents"
log_dir="$HOME/Library/Logs/yandex-mcp"
mkdir -p "$startup_dir" "$log_dir"
agent="$startup_dir/net.yandex-mcp.auth-agent.plist"
cp "$source_dir/auth-agent.plist" "$agent"
# plutil -replace inserts array elements on macOS; remove the placeholder first.
plutil -remove ProgramArguments.0 "$agent"
plutil -insert ProgramArguments.0 -string "$binary" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_AUTH_AGENT_PORT -string "$port" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_AUTH_AGENT_YC_PATH -string "$yc_path" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_SSH_TARGET -string "$YANDEX_MCP_SSH_TARGET" "$agent"
plutil -replace EnvironmentVariables.PATH -string "$PATH" "$agent"
plutil -replace StandardErrorPath -string "$log_dir/auth-agent.log" "$agent"
plutil -lint "$agent"
loaded_agents=$(launchctl list)
loaded_agent=$(printf '%s\n' "$loaded_agents" | awk '$3 == "net.yandex-mcp.auth-agent" { print $3 }')
if [ -n "$loaded_agent" ]; then
  launchctl bootout "gui/$(id -u)/net.yandex-mcp.auth-agent"
fi
launchctl bootstrap "gui/$(id -u)" "$agent"
