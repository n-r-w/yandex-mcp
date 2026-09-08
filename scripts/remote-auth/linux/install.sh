#!/bin/sh
# Run in a systemd-integrated graphical desktop session, without sudo.
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
  write_setting YANDEX_MCP_BINARY "$binary"
  write_setting YANDEX_MCP_AUTH_AGENT_PORT "$port"
  write_setting YANDEX_MCP_AUTH_AGENT_YC_PATH "$yc_path"
  write_setting YANDEX_MCP_SSH_TARGET "$YANDEX_MCP_SSH_TARGET"
  write_setting PATH "$PATH"
} > "$config_dir/auth-agent.env"
cp "$source_dir/auth-agent.service" "$startup_dir/yandex-mcp-auth-agent.service"
# The desktop must also import its environment into the user manager at login.
for name in DISPLAY WAYLAND_DISPLAY XAUTHORITY DBUS_SESSION_BUS_ADDRESS SSH_AUTH_SOCK; do
  if printenv "$name" >/dev/null; then systemctl --user import-environment "$name"; fi
done
systemctl --user daemon-reload
systemctl --user enable yandex-mcp-auth-agent.service
systemctl --user restart yandex-mcp-auth-agent.service
