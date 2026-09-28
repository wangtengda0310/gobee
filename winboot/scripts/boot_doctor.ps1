#Requires -Version 5.1

<#
.SYNOPSIS
    boot doctor 原型: 开发环境自检脚本 (未来 onboarding 工具的检测子命令).
.DESCRIPTION
    检测 WSL / Docker Desktop 的安装状态、残留与可用性, 输出结构化结论:
    [OK] 正常 / [MISSING] 缺失待安装 / [RESIDUE] 半卸载残留待清理 / [WARN] 异常待关注.
    纯只读, 不做任何变更; 不要求管理员权限 (特性检测失败时降级为推断).
.NOTES
    v2 修正:
    - Docker Desktop 支持 per-user 安装 (%LOCALAPPDATA%\Programs\DockerDesktop, 注册表在 HKCU)
      与 all-users 安装 (Program Files, 注册表在 HKLM), 两处都要查
    - docker-desktop 发行版只有在 Docker Desktop 本体缺席时才算残留, 否则是运行中的后端数据盘
    - Get-WindowsOptionalFeature 需要提权, 非提权时降级为按 WSL 功能可用性推断
    - wsl -l -v 行首的 '*' 运行标记会干扰名字解析
    注意: 本文件必须保存为 UTF-8 with BOM, 否则 PS 5.1 按 GBK 解析中文导致语法错误.
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Continue'

# wsl.exe 默认输出 UTF-16, 在 PS 控制台里是乱码; WSL_UTF8=1 强制 UTF-8 (WSL 0.64+)
$env:WSL_UTF8 = '1'
# 控制台按 UTF-8 解码原生程序输出, 与 WSL_UTF8 配套
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

$findings = New-Object System.Collections.Generic.List[object]
function Add-Finding {
    param([string]$Area, [string]$Status, [string]$Detail)
    $script:findings.Add([pscustomobject]@{ Area = $Area; Status = $Status; Detail = $Detail })
    $color = switch ($Status) {
        'OK'      { 'Green' }
        'MISSING' { 'Yellow' }
        'RESIDUE' { 'Magenta' }
        default   { 'Red' }
    }
    Write-Host ("  [{0}] {1}: {2}" -f $Status, $Area, $Detail) -ForegroundColor $color
}

Write-Host "`n=== boot doctor: 开发环境自检 ===" -ForegroundColor Cyan

# --- 1. Windows 版本 (决定安装路径: Win10 2004+ 可用 wsl --install) ---
Write-Host "`n[1/7] Windows 版本" -ForegroundColor Cyan
$os = Get-CimInstance Win32_OperatingSystem
$build = [int]$os.BuildNumber
Add-Finding 'OS' 'OK' ("{0} (build {1})" -f $os.Caption, $build)
if ($build -lt 19041) {
    Add-Finding 'OS' 'WARN' '低于 Win10 2004 (19041), 需走 DISM 传统安装路径'
}

# --- 2. Docker Desktop 安装状态 (先查清, 供后面残留判定使用) ---
Write-Host "`n[2/7] Docker Desktop 安装状态" -ForegroundColor Cyan
# 两种安装形态: per-user (HKCU + %LOCALAPPDATA%) 与 all-users (HKLM + Program Files)
$ddRegKeys = @(
    @{ Hive = 'HKCU'; Path = 'HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop' },
    @{ Hive = 'HKLM'; Path = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop' }
)
$ddReg = $null; $ddHive = $null
foreach ($k in $ddRegKeys) {
    $reg = Get-ItemProperty $k.Path -ErrorAction SilentlyContinue
    if ($reg) { $ddReg = $reg; $ddHive = $k.Hive; break }
}
$ddInstallDirs = @(
    @{ Mode = 'per-user';  Dir = Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop' },
    @{ Mode = 'all-users'; Dir = Join-Path $env:ProgramFiles 'Docker\Docker' }
)
$ddDir = $null; $ddMode = $null
foreach ($d in $ddInstallDirs) {
    if (Test-Path (Join-Path $d.Dir 'Docker Desktop.exe')) { $ddDir = $d.Dir; $ddMode = $d.Mode; break }
}

if ($ddDir) {
    if ($ddReg) {
        Add-Finding 'DockerDesktop' 'OK' ("{0} ({1} 安装, 注册表 {2}, v{3})" -f $ddDir, $ddMode, $ddHive, $ddReg.DisplayVersion)
    } else {
        Add-Finding 'DockerDesktop' 'WARN' ("{0} ({1} 安装), 但注册表卸载项缺失, 卸载/升级可能异常" -f $ddDir, $ddMode)
    }
    $uninstaller = Join-Path $ddDir 'Docker Desktop Installer.exe'
    if (Test-Path $uninstaller) {
        Add-Finding 'DD-Uninstaller' 'OK' "静默卸载器存在 (uninstall --quiet 可用)"
    } else {
        Add-Finding 'DD-Uninstaller' 'WARN' "安装目录中无卸载器: $uninstaller"
    }
} else {
    Add-Finding 'DockerDesktop' 'MISSING' '未找到 Docker Desktop 主程序 (已查 per-user 与 all-users 两个安装位置)'
}

# --- 3. WSL 发行版 + docker-desktop 残留判定 ---
Write-Host "`n[3/7] WSL 发行版" -ForegroundColor Cyan
$distroNames = @()
$distroLines = & wsl --list --verbose 2>$null
if ($LASTEXITCODE -eq 0 -and $distroLines) {
    # 行格式: "* docker-desktop    Running    2" / "  Ubuntu    Stopped    2", 先剥掉行首 '*' 运行标记
    $distroNames = @($distroLines | Select-Object -Skip 1 |
        Where-Object { $_ -match '\S' } |
        ForEach-Object { (($_ -replace '^\s*\*\s*', '') -split '\s+')[0] } |
        Where-Object { $_ })
    if ($distroNames.Count -gt 0) {
        Add-Finding 'WSL-Distros' 'OK' ("共 {0} 个: {1}" -f $distroNames.Count, ($distroNames -join ', '))
    } else {
        Add-Finding 'WSL-Distros' 'MISSING' '无发行版 (对 Docker Desktop 场景非必需, 它自建 docker-desktop 发行版)'
    }
} else {
    Add-Finding 'WSL-Distros' 'MISSING' 'wsl --list 失败或为空'
}
# 残留判定: 只有 Docker Desktop 本体缺席时, docker-desktop 发行版才是残留
if ($distroNames -contains 'docker-desktop') {
    if ($ddDir) {
        Add-Finding 'docker-desktop-distro' 'OK' 'Docker Desktop 的 WSL2 后端数据盘, 正常'
    } else {
        Add-Finding 'docker-desktop-distro' 'RESIDUE' "Docker Desktop 已卸载但数据盘残留, 需 'wsl --unregister docker-desktop' 清理"
    }
}

# --- 4. WSL 本体 (MSIX 包 + CLI 版本) ---
Write-Host "`n[4/7] WSL 本体" -ForegroundColor Cyan
$wslAppx = Get-AppxPackage -Name 'MicrosoftCorporationII.WindowsSubsystemForLinux' -ErrorAction SilentlyContinue
if ($wslAppx) {
    Add-Finding 'WSL' 'OK' ("MSIX 包已安装: {0}" -f $wslAppx.Version)
} else {
    Add-Finding 'WSL' 'MISSING' 'MSIX 包未安装 (可能仅剩系统内置 WSL1)'
}
$null = & wsl --version 2>$null
if ($LASTEXITCODE -eq 0) {
    $wslVer = (& wsl --version 2>$null | Select-Object -First 1)
    Add-Finding 'WSL-CLI' 'OK' "wsl.exe 可用: $wslVer"
} else {
    Add-Finding 'WSL-CLI' 'MISSING' 'wsl.exe 不可用'
}

# --- 5. WSL 系统特性 (CBS 注册表非提权直读; CurrentState: 112=Enabled, 64=Staged, 80=Superseded) ---
# 注: MSIX 版 WSL (Store/独立安装) 只依赖 VirtualMachinePlatform;
#     老特性 Microsoft-Windows-Subsystem-Linux 仅 WSL1/内置版 WSL 需要, 未启用不算问题.
Write-Host "`n[5/7] WSL 系统特性" -ForegroundColor Cyan
$featurePackages = [ordered]@{
    'VirtualMachinePlatform'            = 'HyperV-Feature-VirtualMachinePlatform-Client-Package~*amd64~~*'
    'Microsoft-Windows-Subsystem-Linux' = 'Microsoft-Windows-Lxss-merged-Package~*amd64~~*'
}
foreach ($feat in $featurePackages.Keys) {
    $enabled = $false; $cbsRead = $false
    try {
        $pkgs = Get-ChildItem 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\Packages' -ErrorAction Stop |
            Where-Object { $_.PSChildName -like $featurePackages[$feat] }
        $enabled = @($pkgs | Where-Object { (Get-ItemProperty $_.PSPath -ErrorAction SilentlyContinue).CurrentState -eq 112 }).Count -gt 0
        $cbsRead = $true
    } catch { }
    if ($cbsRead) {
        if ($enabled) {
            Add-Finding $feat 'OK' '已启用 (CBS 注册表确认, 无需提权)'
        } elseif ($feat -eq 'Microsoft-Windows-Subsystem-Linux' -and $wslAppx) {
            Add-Finding $feat 'OK' '未启用, 但 MSIX 版 WSL 不依赖此旧特性 (仅 WSL1 需要), 无影响'
        } else {
            Add-Finding $feat 'MISSING' '未启用 (需要 dism 启用 + 可能重启)'
        }
    } else {
        # CBS 不可读时降级为推断
        if ($wslAppx -or ($distroNames.Count -gt 0)) {
            Add-Finding $feat 'OK' '已启用 (推断; CBS 注册表不可读)'
        } else {
            Add-Finding $feat 'WARN' '无法确认 (CBS 不可读且无法推断)'
        }
    }
}

# --- 6. Docker CLI 与 daemon ---
Write-Host "`n[6/7] Docker CLI / daemon" -ForegroundColor Cyan
$dockerCli = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerCli) {
    # PATH 未刷新 (安装后未重开终端): 从两种安装形态的已知路径探测
    foreach ($bin in @((Join-Path $env:ProgramFiles 'Docker\Docker\resources\bin'), (Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop\resources\bin'))) {
        $probe = Join-Path $bin 'docker.exe'
        if (Test-Path $probe) { $dockerCli = @{ Source = $probe }; break }
    }
}
if ($dockerCli) {
    Add-Finding 'Docker-CLI' 'OK' ("docker.exe: {0}{1}" -f $dockerCli.Source, $(if ($env:PATH -notlike "*$($dockerCli.Source | Split-Path)*") { ' (注意: 当前会话 PATH 未包含, 新终端自动恢复)' } else { '' }))
    $null = & $dockerCli.Source info --format '{{.ServerVersion}}' 2>$null
    if ($LASTEXITCODE -eq 0) {
        $ver = & $dockerCli.Source info --format '{{.ServerVersion}}' 2>$null
        Add-Finding 'Docker-daemon' 'OK' "daemon 运行中, server $ver"
    } else {
        Add-Finding 'Docker-daemon' 'MISSING' 'daemon 未运行 (Docker Desktop 未启动或已损坏)'
    }
} else {
    Add-Finding 'Docker-CLI' 'MISSING' 'docker 不在 PATH 中'
}

# --- 7. 环境配置 (settings-store.json / winget / 私仓登录态) ---
Write-Host "`n[7/7] 环境配置" -ForegroundColor Cyan
$settingsFile = Join-Path $env:APPDATA 'Docker\settings-store.json'
if (Test-Path $settingsFile) {
    Add-Finding 'DD-Settings' 'OK' $settingsFile
} else {
    Add-Finding 'DD-Settings' 'MISSING' 'settings-store.json 不存在 (未首启过, 可预置代理/后端配置)'
}
$winget = Get-Command winget -ErrorAction SilentlyContinue
if ($winget) {
    Add-Finding 'winget' 'OK' ("{0}" -f (& winget --version))
} else {
    Add-Finding 'winget' 'MISSING' 'winget 不可用, 安装需走直连下载路径'
}
# --- 私仓登录检查: 地址留空则整项跳过 (私仓地址因环境而异, 不内置) ---
$PrivateRegistry = ''
if ($PrivateRegistry) {
    $configFile = Join-Path $env:USERPROFILE '.docker\config.json'
    if (Test-Path $configFile) {
        $escaped = $PrivateRegistry -replace '\.', '\.'
        $logged = Select-String -Path $configFile -Pattern $escaped -Quiet
        if ($logged) { Add-Finding 'Registry-Login' 'OK' ("已登录私仓 {0}" -f $PrivateRegistry) }
        else         { Add-Finding 'Registry-Login' 'MISSING' ("未登录私仓 {0}" -f $PrivateRegistry) }
    } else {
        Add-Finding 'Registry-Login' 'MISSING' '~/.docker/config.json 不存在 (未登录过私仓)'
    }
}

# --- 汇总 (单次判定, 不随 finding 数量重复) ---
Write-Host "`n=== 汇总 ===" -ForegroundColor Cyan
$findings | Group-Object Status | ForEach-Object {
    Write-Host ("  {0}: {1} 项" -f $_.Name, $_.Count)
}
$issues = @($findings | Where-Object { $_.Status -in 'MISSING', 'RESIDUE', 'WARN' })
if ($issues.Count -eq 0) {
    Write-Host "`n结论: 环境健康, 可直接 boot up." -ForegroundColor Green
} else {
    Write-Host "`n结论: 环境不完整, 建议动作:" -ForegroundColor Yellow
    foreach ($f in ($issues | Where-Object Status -eq 'RESIDUE')) {
        Write-Host ("  - [清理] {0}: {1}" -f $f.Area, $f.Detail) -ForegroundColor Magenta
    }
    foreach ($f in ($issues | Where-Object Status -in 'MISSING', 'WARN')) {
        Write-Host ("  - [安装/修复] {0}: {1}" -f $f.Area, $f.Detail) -ForegroundColor Yellow
    }
}
