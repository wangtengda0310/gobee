# cdp_eval.ps1 - Connect to rain-qa-func's Android WebView via CDP remote debugging, execute JS, return result.
#
# Auto steps: find app pid -> adb forward -> pick page target -> connect WebSocket -> run JS.
# Capabilities: read DOM text, click nav/buttons to trigger backend calls, inspect globals, verify Service returns.
#
# Usage:
#   powershell -File build/android/scripts/cdp_eval.ps1 -Js "document.title"
#   powershell -File build/android/scripts/cdp_eval.ps1 -JsFile some-snippet.js
#
# Prerequisite: device/emulator connected via adb, rain-qa-func running in foreground.
# NOTE: Keep this script ASCII-only. Windows PowerShell 5.1 reads BOM-less UTF-8 .ps1 as GBK,
#       non-ASCII comments/strings get corrupted and cause parse errors.
param(
  [string]$Js,
  [string]$JsFile,
  [int]$CdpPort = 9223,
  [string]$Package = "com.wails.app",
  [string]$Activity = "com.wails.app/.MainActivity",
  [string]$Adb,
  [int]$TimeoutSec = 25
)

# Resolve adb path: -Adb param -> ANDROID_HOME -> common install locations
if (-not $Adb) {
  if ($env:ANDROID_HOME) { $Adb = Join-Path $env:ANDROID_HOME "platform-tools\adb.exe" }
  if (-not (Test-Path $Adb)) {
    $candidates = @("D:\Android\Sdk\platform-tools\adb.exe", "C:\Android\Sdk\platform-tools\adb.exe")
    foreach ($c in $candidates) { if (Test-Path $c) { $Adb = $c; break } }
  }
}
if (-not $Adb -or -not (Test-Path $Adb)) { throw "adb.exe not found; pass -Adb or set ANDROID_HOME" }

if ($JsFile) { $Js = Get-Content -Raw -Path $JsFile }
if (-not $Js) { throw "Provide JS via -Js or -JsFile" }

# 1. find app pid; start the app if it is not running
$pidval = (& $Adb shell pidof $Package).Trim()
if (-not $pidval) {
  Write-Host "App not running, starting $Activity ..."
  [void](& $Adb shell am start -n $Activity)
  Start-Sleep -Seconds 6
  $pidval = (& $Adb shell pidof $Package).Trim()
}
if (-not $pidval) { throw "Running $Package not found" }

# 2. adb forward to the WebView devtools socket
[void](& $Adb forward "tcp:$CdpPort" "localabstract:webview_devtools_remote_$pidval")

# 3. pick the page target WebSocket URL
try {
  $targets = (Invoke-WebRequest "http://127.0.0.1:$CdpPort/json" -UseBasicParsing -TimeoutSec 5).Content | ConvertFrom-Json
} catch { throw "CDP /json request failed (WebView debug not enabled?): $($_.Exception.Message)" }
$page = $targets | Where-Object { $_.type -eq "page" } | Select-Object -First 1
if (-not $page) { throw "No page target found; page may not be loaded" }
$wsUrl = $page.webSocketDebuggerUrl

# 4. connect WebSocket, send Runtime.evaluate, receive reply
$ws = New-Object System.Net.WebSockets.ClientWebSocket
$cts = New-Object System.Threading.CancellationTokenSource
$cts.CancelAfter([TimeSpan]::FromSeconds($TimeoutSec))
$ws.ConnectAsync($wsUrl, $cts.Token).Wait()

$payload = @{ id = 1; method = "Runtime.evaluate"; params = @{ expression = $Js; returnByValue = $true; awaitPromise = $true } } | ConvertTo-Json -Compress -Depth 10
$bytes = [Text.Encoding]::UTF8.GetBytes($payload)
$seg = [System.ArraySegment[byte]]::new($bytes, 0, $bytes.Length)
[void]$ws.SendAsync($seg, [System.Net.WebSockets.WebSocketMessageType]::Text, $true, $cts.Token).Wait()

$recvBuf = New-Object byte[] 1048576
$tot = ""
do {
  $seg2 = [System.ArraySegment[byte]]::new($recvBuf, 0, $recvBuf.Length)
  $r = $ws.ReceiveAsync($seg2, $cts.Token).Result
  $tot += [Text.Encoding]::UTF8.GetString($recvBuf, 0, $r.Count)
} until ($r.EndOfMessage)
$ws.Dispose()

# 5. parse and print result
try {
  $obj = $tot | ConvertFrom-Json
  if ($obj.result.exceptionDetails) { "[JS Error] " + $obj.result.exceptionDetails.text + " :: " + $obj.result.result.description }
  elseif ($obj.result.result.value) { $obj.result.result.value }
  elseif ($obj.result.result.description) { $obj.result.result.description }
  else { $obj.result.result | ConvertTo-Json -Depth 10 }
} catch { $tot }
