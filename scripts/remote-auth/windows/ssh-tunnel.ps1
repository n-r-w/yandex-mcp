# Keep native stderr diagnostics in the log without terminating PowerShell early.
$ErrorActionPreference = 'Continue'
& "$env:WINDIR\System32\OpenSSH\ssh.exe" -N -T `
  -o BatchMode=yes -o ExitOnForwardFailure=yes -o StrictHostKeyChecking=yes `
  -o ServerAliveInterval=15 -o ServerAliveCountMax=3 -o ConnectTimeout=10 `
  -R '@FORWARD@' -- '@SSH_TARGET@' >> "$env:LOCALAPPDATA\YandexMCP\ssh-tunnel.log" 2>&1
exit $LASTEXITCODE
