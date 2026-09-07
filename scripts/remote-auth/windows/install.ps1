# Run as the desktop user. Administrator privileges are not required.
$ErrorActionPreference = 'Stop'
foreach ($name in @('YANDEX_MCP_BINARY', 'YANDEX_MCP_AUTH_AGENT_YC_PATH', 'YANDEX_MCP_AUTH_AGENT_PROFILES', 'YANDEX_MCP_SSH_TARGET')) {
    if (-not [Environment]::GetEnvironmentVariable($name)) { throw "Set $name before installation" }
}
foreach ($path in @($env:YANDEX_MCP_BINARY, $env:YANDEX_MCP_AUTH_AGENT_YC_PATH)) {
    if (-not [IO.Path]::IsPathRooted($path) -or -not (Test-Path -LiteralPath $path -PathType Leaf)) {
        throw "Executable must have an absolute file path: $path"
    }
}
$listen = $env:YANDEX_MCP_AUTH_AGENT_LISTEN
if (-not $listen) { $listen = '127.0.0.1:18765' }
if ($listen -notmatch '^127\.0\.0\.1:([0-9]+)$') { throw 'Listen address must be 127.0.0.1:<port>' }
$localPort = [int]$Matches[1]
$remotePort = 18765
if ($env:YANDEX_MCP_REMOTE_PORT) { $remotePort = [int]$env:YANDEX_MCP_REMOTE_PORT }
foreach ($port in @($localPort, $remotePort)) {
    if ($port -lt 1 -or $port -gt 65535) { throw 'Ports must be from 1 to 65535' }
}
$directory = Join-Path $env:LOCALAPPDATA 'YandexMCP'
New-Item -ItemType Directory -Force -Path $directory | Out-Null
$values = @{
    '@LISTEN@' = $listen
    '@YC_PATH@' = $env:YANDEX_MCP_AUTH_AGENT_YC_PATH
    '@PROFILES@' = $env:YANDEX_MCP_AUTH_AGENT_PROFILES
    '@BINARY@' = $env:YANDEX_MCP_BINARY
    '@FORWARD@' = "127.0.0.1:${remotePort}:127.0.0.1:${localPort}"
    '@SSH_TARGET@' = $env:YANDEX_MCP_SSH_TARGET
}
$user = [Security.Principal.WindowsIdentity]::GetCurrent().Name
$principal = New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited
$logon = New-ScheduledTaskTrigger -AtLogOn -User $user
# The repeating trigger retries indefinitely after sleep or a long network outage.
# IgnoreNew prevents concurrent instances when a task is already running.
$retry = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 1)
$settings = New-ScheduledTaskSettingsSet -MultipleInstances IgnoreNew -ExecutionTimeLimit ([TimeSpan]::Zero) `
    -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
foreach ($name in @('auth-agent', 'ssh-tunnel')) {
    $content = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot "$name.ps1")
    foreach ($entry in $values.GetEnumerator()) {
        $content = $content.Replace($entry.Key, $entry.Value.Replace("'", "''"))
    }
    $script = Join-Path $directory "$name.ps1"
    Set-Content -LiteralPath $script -Value $content -Encoding UTF8
    $action = New-ScheduledTaskAction -Execute "$env:WINDIR\System32\WindowsPowerShell\v1.0\powershell.exe" `
        -Argument "-NoProfile -NonInteractive -ExecutionPolicy Bypass -File `"$script`""
    $taskName = "Yandex MCP $name"
    if (Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue) {
        Stop-ScheduledTask -TaskName $taskName
    }
    Register-ScheduledTask -TaskName $taskName -Action $action -Trigger @($logon, $retry) `
        -Settings $settings -Principal $principal -Force | Out-Null
    Start-ScheduledTask -TaskName $taskName
}
