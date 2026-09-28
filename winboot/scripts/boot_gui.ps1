#Requires -Version 5.1

<#
.SYNOPSIS
    boot gui: 引导获取并启动 GUI (winboot.exe) —— 解决"第一份 exe"的鸡生蛋.
.DESCRIPTION
    新同事 clone 仓库后运行本命令即可, 无需预装任何东西:
      [1] release\winboot.exe 已存在 -> 直接启动 (升级交给其自更新机制)
      [2] 缺失 -> 三级获取 (lib/acquire.ps1): %TEMP% 缓存 -> 局域网同类 -> 中心部署地址
          局域网同类经 UDP 广播发现 (与 lanshare.py 同协议, 无中心也能人传人)
    下载到的可能是任意版本, 首次运行后其自更新机制会拉齐到最新.
.PARAMETER UpdateBase
    中心部署地址 (兜底下载源).
.PARAMETER NoLaunch
    只获取不启动.
.PARAMETER Force
    已有 release\winboot.exe 也重新获取 (默认直接用现有版).
.NOTES
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 兼容).
#>

param(
    [string]$UpdateBase = 'http://10.68.115.208:8760',
    [switch]$NoLaunch,
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

. (Join-Path $PSScriptRoot 'lib\acquire.ps1')

$winbootRoot = Split-Path $PSScriptRoot -Parent
$dest = Join-Path $winbootRoot 'release\winboot.exe'

# --- 版本比较 ('1.10.0' > '1.9.1'; 非数字段按 0 兜底, 引导场景够用) ---
function Test-VersionGt {
    param([string]$a, [string]$b)
    $pa = @($a -split '\.'); $pb = @($b -split '\.')
    for ($i = 0; $i -lt [Math]::Max($pa.Count, $pb.Count); $i++) {
        $x = 0; $y = 0
        if ($i -lt $pa.Count -and $pa[$i] -match '^\d+$') { $x = [int]$pa[$i] }
        if ($i -lt $pb.Count -and $pb[$i] -match '^\d+$') { $y = [int]$pb[$i] }
        if ($x -gt $y) { return $true }
        if ($x -lt $y) { return $false }
    }
    return $false
}

# --- 局域网同类发现 (与 lanshare.py 同协议: 广播 WINBOOT_DISCOVER, 收单播 JSON 清单) ---
function Find-LanPeer {
    param([int]$TimeoutMs = 3000)
    $sock = New-Object System.Net.Sockets.Socket ([System.Net.Sockets.AddressFamily]::InterNetwork,
        [System.Net.Sockets.SocketType]::Dgram, [System.Net.Sockets.ProtocolType]::Udp)
    try {
        $sock.SetSocketOption([System.Net.Sockets.SocketOptionLevel]::Socket,
            [System.Net.Sockets.SocketOptionName]::Broadcast, $true)
        $msg = [System.Text.Encoding]::UTF8.GetBytes('WINBOOT_DISCOVER')
        # 受限广播 + 各网卡 /24 定向广播 (与 Python 侧 _broadcast_targets 一致)
        $targets = @()
        $targets += [System.Net.IPEndPoint]::new([System.Net.IPAddress]::Broadcast, 58765)
        foreach ($ip in (Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
                Where-Object { $_.IPAddress -notlike '127.*' -and $_.IPAddress -notlike '169.254.*' })) {
            $p = $ip.IPAddress.Split('.')
            $targets += [System.Net.IPEndPoint]::new([System.Net.IPAddress]::Parse("$($p[0]).$($p[1]).$($p[2]).255"), 58765)
        }
        foreach ($t in ($targets | Select-Object -Unique)) {
            try { $sock.SendTo($msg, $t) | Out-Null } catch { }
        }
        # 500ms 接收窗口轮询到总超时; 多个应答取版本最高者.
        # PS 坑: 无连接 socket 收包必须 ReceiveFrom(buffer, [ref]EndPoint) ——
        # Receive 没有 ref EndPoint 重载, 且 [ref] 必须套在显式类型的变量上
        # (无类型变量的 [ref] 是 [ref]Object, 重载解析失败, 实测 MethodException).
        $sock.ReceiveTimeout = 500
        [byte[]]$buf = New-Object byte[] 65536
        $best = $null
        $sw = [Diagnostics.Stopwatch]::StartNew()
        while ($sw.ElapsedMilliseconds -lt $TimeoutMs) {
            [System.Net.EndPoint]$remote = New-Object System.Net.IPEndPoint ([System.Net.IPAddress]::Any, 0)
            try { $n = $sock.ReceiveFrom($buf, [ref]$remote) } catch { continue }   # 本窗口无应答
            try { $m = [System.Text.Encoding]::UTF8.GetString($buf, 0, $n) | ConvertFrom-Json } catch { continue }
            if ($m.app -ne 'winboot') { continue }
            if (-not $m.PSObject.Properties['tool'] -or -not $m.tool) { continue }  # 对方不服务工具本体 (源码模式)
            $peer = @{ Url = "http://$($remote.Address):$($m.port)"; Version = [string]$m.tool.version }
            if (-not $best -or (Test-VersionGt $peer.Version $best.Version)) { $best = $peer }
        }
        return $best
    } finally { $sock.Close() }
}

# --- 主流程 ---
if ((Test-Path $dest) -and -not $Force) {
    Write-Host "GUI 已就位: $dest" -ForegroundColor Green
    Write-Host "  (升级由其自更新机制负责; 强制重新获取: boot.ps1 gui -Force)"
} else {
    Write-Host "[gui] 引导获取 winboot.exe (缓存 -> 局域网同类 -> 中心)..."
    $peer = Find-LanPeer
    if ($peer) {
        $script:BootLanBase = $peer.Url
        Write-Host "  发现局域网同类: $($peer.Url) (v$($peer.Version)), 优先从其下载" -ForegroundColor Cyan
    } else {
        Write-Host "  局域网无同类在线, 走缓存/中心源"
    }
    $cache = Join-Path $env:TEMP 'winboot.exe'
    $got = Get-BootFile -Name 'winboot GUI' -CachePath $cache -MinBytes 10MB -FileName 'winboot.exe' `
        -OfficialUrl "$($UpdateBase.TrimEnd('/'))/winboot.exe"
    if (-not $got -or $got -eq $true) {
        Write-Host "获取失败 (可 -UpdateBase 指定中心地址)" -ForegroundColor Red
        exit 1
    }
    New-Item -ItemType Directory -Force (Split-Path $dest -Parent) | Out-Null
    Copy-Item $cache $dest -Force
    Write-Host "GUI 已就位: $dest" -ForegroundColor Green
}

if (-not $NoLaunch) {
    Start-Process -FilePath $dest -WorkingDirectory $winbootRoot
    Write-Host "已启动. 后续版本由其自更新机制接管, 本命令只在 exe 缺失时才需要再跑."
}
