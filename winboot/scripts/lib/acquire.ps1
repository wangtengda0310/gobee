#Requires -Version 5.1

<#
.SYNOPSIS
    boot 工具的通用文件获取组件: 三级策略 本地缓存 -> 局域网源 -> 兜底.
.DESCRIPTION
    被 boot_install.ps1 点源引入; 新软件(WSL/Docker Desktop/未来的 git 等)复用 Get-BootFile,
    每个软件只需定义: 名字/缓存路径/体积阈值/局域网文件名/兜底动作.

    局域网源约定: 目录式服务, 文件按名存放 ($BootLanBase/<FileName>);
    未来 GUI 内置 HTTP 服务 + 局域网广播后, 由其解析实际地址并覆盖 $BootLanBase.

    下载安全: 先落 .part 再原子改名, 缓存文件永远完整; 缓存命中带体积校验(防半截/损坏文件).
    缓存语义: 命中即用 = 版本锁定 (团队部署已知版本时是特性; 想追新删缓存文件即可).
.NOTES
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 兼容).
#>

# 局域网源基地址 (当前为本机测试地址; GUI 广播接管后由调用方覆盖)
$script:BootLanBase = 'http://127.0.0.1:8765'

# 下载单文件到 $Dest (经 .part 原子改名), 返回 $true/$false.
# curl 主力 (PS5.1 的 IWR 进度条渲染会拖慢数倍); 不重定向 curl 的 stderr(进度条),
# EAP=Stop 下重定向原生 stderr 会抛 NativeCommandError.
function Save-BootFileFromUrl {
    param([string]$Name, [string]$Url, [string]$Dest, [int]$ConnectTimeoutSec = 0)
    Write-Host "  从 $Name 下载: $Url"
    $part = "$Dest.part"
    Remove-Item $part -Force -ErrorAction SilentlyContinue
    $dlStart = Get-Date
    $code = 1
    if (Get-Command curl.exe -ErrorAction SilentlyContinue) {
        $curlArgs = @('-L', '--fail', '--retry', '3', '--retry-delay', '2', '-C', '-', '-o', $part, $Url)
        if ($ConnectTimeoutSec -gt 0) { $curlArgs += @('--connect-timeout', "$ConnectTimeoutSec") }
        & curl.exe @curlArgs
        $code = $LASTEXITCODE
    } else {
        $ProgressPreference = 'SilentlyContinue'   # IWR 提速关键: 关进度条渲染
        try { Invoke-WebRequest -Uri $Url -OutFile $part -UseBasicParsing; $code = 0 } catch { }
    }
    if ($code -eq 0 -and (Test-Path $part)) {
        Move-Item $part $Dest -Force
        $sizeMB = [math]::Round((Get-Item $Dest).Length / 1MB, 0)
        $secs = [math]::Round(((Get-Date) - $dlStart).TotalSeconds, 0)
        Write-Host ("  下载完成 ($Name): {0}MB, 耗时 {1}s" -f $sizeMB, $secs) -ForegroundColor Green
        return $true
    }
    Write-Host "  $Name 下载失败 (exit $code)" -ForegroundColor Yellow
    return $false
}

# 三级获取: [1]本地缓存(体积校验) -> [2]局域网源($BootLanBase/$FileName, 3s 快速失败)
#           -> [3]兜底 (OfficialUrl 下载, 或调用方提供的 $Fallback 脚本块).
# 返回值: $CachePath(字符串) = 文件已就位; $true = 兜底脚本块自行完成了全部工作(无文件);
#         $null = 全部失败.
function Get-BootFile {
    param(
        [string]$Name,             # 显示名 (如 'DD 安装器')
        [string]$CachePath,        # 本地缓存路径
        [long]$MinBytes,           # 缓存有效性阈值 (如 200MB)
        [string]$FileName,         # 局域网源文件名 ($BootLanBase/$FileName)
        [string]$OfficialUrl = '', # 兜底: 官方下载地址
        [scriptblock]$Fallback = $null # 兜底: 自定义动作 (返回 $true = 已自行处理, 如 winget 直接安装)
    )
    # [1] 本地缓存
    if (Test-Path $CachePath) {
        $sizeMB = [math]::Round((Get-Item $CachePath).Length / 1MB, 0)
        if ((Get-Item $CachePath).Length -ge $MinBytes) {
            Write-Host "  [缓存] 命中 ${Name}: $CachePath ($sizeMB MB, 跳过下载)" -ForegroundColor Green
            return $CachePath
        }
        Write-Host "  [缓存] $Name 文件不完整 ($sizeMB MB 低于阈值), 删除后重新获取"
        Remove-Item $CachePath -Force -ErrorAction SilentlyContinue
    }
    # [2] 局域网源
    if (Save-BootFileFromUrl -Name "$Name(局域网源)" -Url "$script:BootLanBase/$FileName" -Dest $CachePath -ConnectTimeoutSec 3) {
        return $CachePath
    }
    # [3] 兜底
    if ($OfficialUrl -and (Save-BootFileFromUrl -Name "$Name(官方源)" -Url $OfficialUrl -Dest $CachePath)) {
        return $CachePath
    }
    if ($Fallback) {
        if (& $Fallback) { return $true }
    }
    Write-Host "  $Name 所有获取渠道均失败." -ForegroundColor Red
    return $null
}
