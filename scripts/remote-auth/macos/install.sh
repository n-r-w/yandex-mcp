#!/bin/sh
# Run in a logged-in macOS graphical session. No root privileges are needed.
set -eu
: "${YANDEX_MCP_BINARY:?Set the absolute yandex-mcp executable path}"
: "${YANDEX_MCP_AUTH_AGENT_YC_PATH:?Set the absolute native yc executable path}"
: "${YANDEX_MCP_AUTH_AGENT_PROFILES:?Set comma-separated allowed profiles}"
: "${YANDEX_MCP_SSH_TARGET:?Set the SSH host or user@host}"
listen=${YANDEX_MCP_AUTH_AGENT_LISTEN:-127.0.0.1:18765}
remote_port=${YANDEX_MCP_REMOTE_PORT:-18765}
case "$listen" in 127.0.0.1:*) local_port=${listen#127.0.0.1:} ;; *) echo 'Listen address must be 127.0.0.1:<port>' >&2; exit 1 ;; esac
for port in "$local_port" "$remote_port"; do
  case "$port" in ''|*[!0-9]*) echo 'Ports must be integers' >&2; exit 1 ;; esac
  [ "$port" -ge 1 ] && [ "$port" -le 65535 ] || exit 1
done
for executable in "$YANDEX_MCP_BINARY" "$YANDEX_MCP_AUTH_AGENT_YC_PATH"; do
  case "$executable" in /*) ;; *) echo 'Executable paths must be absolute' >&2; exit 1 ;; esac
  [ -x "$executable" ] || { echo "Not executable: $executable" >&2; exit 1; }
done
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
startup_dir="$HOME/Library/LaunchAgents"
log_dir="$HOME/Library/Logs/yandex-mcp"
mkdir -p "$startup_dir" "$log_dir"
agent="$startup_dir/net.yandex-mcp.auth-agent.plist"
tunnel="$startup_dir/net.yandex-mcp.ssh-tunnel.plist"
cp "$source_dir/auth-agent.plist" "$agent"
cp "$source_dir/ssh-tunnel.plist" "$tunnel"
plutil -replace ProgramArguments.0 -string "$YANDEX_MCP_BINARY" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_AUTH_AGENT_LISTEN -string "$listen" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_AUTH_AGENT_YC_PATH -string "$YANDEX_MCP_AUTH_AGENT_YC_PATH" "$agent"
plutil -replace EnvironmentVariables.YANDEX_MCP_AUTH_AGENT_PROFILES -string "$YANDEX_MCP_AUTH_AGENT_PROFILES" "$agent"
plutil -replace StandardErrorPath -string "$log_dir/auth-agent.log" "$agent"
plutil -replace ProgramArguments.16 -string "127.0.0.1:$remote_port:127.0.0.1:$local_port" "$tunnel"
plutil -replace ProgramArguments.18 -string "$YANDEX_MCP_SSH_TARGET" "$tunnel"
plutil -replace StandardErrorPath -string "$log_dir/ssh-tunnel.log" "$tunnel"
plutil -lint "$agent" "$tunnel"
for name in auth-agent ssh-tunnel; do
  label="net.yandex-mcp.$name"
  launchctl bootout "gui/$(id -u)/$label" 2>/dev/null || true
  launchctl bootstrap "gui/$(id -u)" "$startup_dir/$label.plist"
done
