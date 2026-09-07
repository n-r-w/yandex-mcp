#!/bin/sh
# Run in a systemd-integrated graphical desktop session, without sudo.
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
systemctl --user is-active --quiet graphical-session.target || {
  echo 'Run this installer from an active systemd graphical desktop session' >&2
  exit 1
}
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
startup_dir="$HOME/.config/systemd/user"
config_dir="$HOME/.config/yandex-mcp"
mkdir -p "$startup_dir" "$config_dir"
# EnvironmentFile uses double-quoted values, not shell evaluation.
write_setting() {
  case "$2" in *'
'*) echo 'Settings cannot contain newlines' >&2; exit 1 ;; esac
  printf '%s="' "$1"
  printf '%s' "$2" | sed 's/\\/\\\\/g; s/"/\\"/g'
  printf '"\n'
}
{
  write_setting YANDEX_MCP_BINARY "$YANDEX_MCP_BINARY"
  write_setting YANDEX_MCP_AUTH_AGENT_LISTEN "$listen"
  write_setting YANDEX_MCP_AUTH_AGENT_YC_PATH "$YANDEX_MCP_AUTH_AGENT_YC_PATH"
  write_setting YANDEX_MCP_AUTH_AGENT_PROFILES "$YANDEX_MCP_AUTH_AGENT_PROFILES"
  write_setting YANDEX_MCP_SSH_TARGET "$YANDEX_MCP_SSH_TARGET"
  write_setting YANDEX_MCP_FORWARD "127.0.0.1:$remote_port:127.0.0.1:$local_port"
} > "$config_dir/auth-agent.env"
cp "$source_dir/auth-agent.service" "$startup_dir/yandex-mcp-auth-agent.service"
cp "$source_dir/ssh-tunnel.service" "$startup_dir/yandex-mcp-ssh-tunnel.service"
# The desktop must also import its environment into the user manager at login.
for name in DISPLAY WAYLAND_DISPLAY XAUTHORITY DBUS_SESSION_BUS_ADDRESS; do
  if printenv "$name" >/dev/null; then systemctl --user import-environment "$name"; fi
done
systemctl --user daemon-reload
systemctl --user enable yandex-mcp-auth-agent.service yandex-mcp-ssh-tunnel.service
systemctl --user restart yandex-mcp-auth-agent.service yandex-mcp-ssh-tunnel.service
