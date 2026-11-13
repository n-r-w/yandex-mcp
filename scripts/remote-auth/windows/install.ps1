# Run as the desktop user. Administrator privileges are not required.
$ErrorActionPreference = 'Stop'
if (-not $env:YANDEX_MCP_SSH_TARGET) { throw 'Set YANDEX_MCP_SSH_TARGET before installation' }
$binaryCommand = 'yandex-mcp'
if ($env:YANDEX_MCP_BINARY) { $binaryCommand = $env:YANDEX_MCP_BINARY }
$binary = (Get-Command -Name $binaryCommand -CommandType Application -ErrorAction Stop).Source
# Auth-agent looks up the command name yc through the saved PATH at each IAM request.
$ycPath = 'yc'
if ($env:YANDEX_MCP_AUTH_AGENT_YC_PATH) {
    $ycPath = (Get-Command -Name $env:YANDEX_MCP_AUTH_AGENT_YC_PATH -CommandType Application -ErrorAction Stop).Source
} else {
    $foundYC = (Get-Command -Name 'yc' -CommandType Application -ErrorAction SilentlyContinue).Source
    if ($foundYC) { $ycPath = $foundYC } else { Write-Warning 'yc not found; auth-agent serves IAM tokens after yc is installed in PATH' }
}
Get-Command -Name ssh -CommandType Application -ErrorAction Stop | Out-Null
$port = 18765
if ($env:YANDEX_MCP_AUTH_AGENT_PORT) { $port = [int]$env:YANDEX_MCP_AUTH_AGENT_PORT }
if ($port -lt 1 -or $port -gt 65535) { throw 'Port must be from 1 to 65535' }
# Empty means the default OAuth directory that auth-agent computes at each start.
$oauthDir = [string]$env:YANDEX_MCP_AUTH_AGENT_OAUTH_DIR
# Absolute Windows paths start with a drive and a separator, or are UNC paths.
if ($oauthDir -and $oauthDir -notmatch '^([A-Za-z]:[\\/]|[\\/]{2})') {
    throw 'YANDEX_MCP_AUTH_AGENT_OAUTH_DIR must be an absolute path'
}
$directory = Join-Path $env:LOCALAPPDATA 'YandexMCP'
New-Item -ItemType Directory -Force -Path $directory | Out-Null
$values = @{
    '@PORT@' = [string]$port
    '@YC_PATH@' = $ycPath
    '@OAUTH_DIR@' = $oauthDir
    '@BINARY@' = $binary
    '@SSH_TARGET@' = $env:YANDEX_MCP_SSH_TARGET
    '@PATH@' = $env:PATH
}
$content = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'auth-agent.ps1')
foreach ($entry in $values.GetEnumerator()) {
    $content = $content.Replace($entry.Key, $entry.Value.Replace("'", "''"))
}
$script = Join-Path $directory 'auth-agent.ps1'
Set-Content -LiteralPath $script -Value $content -Encoding UTF8
$user = [Security.Principal.WindowsIdentity]::GetCurrent().Name
$principal = New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited
$logon = New-ScheduledTaskTrigger -AtLogOn -User $user
# Retry stopped auth-agent tasks; SSH reconnects inside auth-agent.
$retry = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes 1)
$settings = New-ScheduledTaskSettingsSet -MultipleInstances IgnoreNew -ExecutionTimeLimit ([TimeSpan]::Zero) `
    -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
$action = New-ScheduledTaskAction -Execute "$env:WINDIR\System32\WindowsPowerShell\v1.0\powershell.exe" `
    -Argument "-NoProfile -NonInteractive -ExecutionPolicy Bypass -File `"$script`""
$taskName = 'Yandex MCP auth-agent'
$taskPath = '\'
$existingTask = Get-ScheduledTask | Where-Object { $_.TaskName -eq $taskName -and $_.TaskPath -eq $taskPath }
if ($existingTask) { $existingTask | Stop-ScheduledTask }
Register-ScheduledTask -TaskName $taskName -TaskPath $taskPath -Action $action -Trigger @($logon, $retry) `
    -Settings $settings -Principal $principal -Force | Out-Null
Start-ScheduledTask -TaskName $taskName -TaskPath $taskPath
