#Requires -Version 5.1

<#
.SYNOPSIS
    boot install: 安装 WSL + Docker Desktop 并等待 daemon 就绪.
.DESCRIPTION
    提权范围最小化: 主流程始终以普通用户运行 ——
    - DD 默认"当前用户安装"(%LOCALAPPDATA%, 免 UAC, 日志全程可被调用方捕获);
    - 仅当 WSL/VMP 特性缺失时, 派一个提权 helper 只做该阶段(弹一次 UAC), 完成后回到本进程继续;
    - -AllUsers 显式选择所有用户安装(Program Files, 安装器需要 UAC).
    流程: [1] WSL(必要时提权 helper) [2] DD 三级下载+安装+daemon.json 预置 [3] 启动并等 daemon 就绪.
.PARAMETER InstallerPath
    直接使用本地安装器文件 (优先级最高).
.PARAMETER InstallerUrl
    显式指定下载源 (跳过本地缓存与局域网源, 只从这里下);
    缺省三级策略: %TEMP% 缓存 -> 局域网源(-LanBase 覆盖, 缺省 acquire.ps1 内置地址) -> 官方 evergreen.
.PARAMETER AllUsers
    DD 安装到 Program Files 对所有用户可见 (安装器需要管理员, 会弹 UAC);
    缺省当前用户安装 (免 UAC).
.PARAMETER Yes
    跳过确认.
.NOTES
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 兼容).
#>

param(
    [string]$InstallerPath = '',
    [string]$InstallerUrl = '',
    # 局域网源覆盖 (GUI 发现广播源后传入, 如 http://192.168.1.23:8765);
    # 缺省用 lib/acquire.ps1 内置地址. 会同时透传给提权 helper (WSL 获取在其阶段).
    [string]$LanBase = '',
    # 私仓 insecure-registries 地址 (HTTP 明文 Harbor 才需要; 因环境而异, 不内置)
    # 例: boot.ps1 install -InsecureRegistries harbor.local:5000
    [string[]]$InsecureRegistries = @(),
    [switch]$AllUsers,
    [switch]$Yes,
    # 跳过"按任意键退出"暂停 (GUI/隐藏控制台调用必须加, 否则 ReadKey 永久阻塞)
    [switch]$NoPause,
    # 组件列表 (位置参数, GUI 点选/预设套装透传); 缺省全套.
    # 校验在运行时对 lib/components.ps1 注册表做 (ValidateSet 必须字面量, 无法引用注册表)
    [string[]]$Components = @('wsl', 'docker'),
    # 内部参数: GUI exe 路径 (经 WINBOOT_EXE 环境变量来, 穿透提权 helper 后
    # 在已有管理员上下文里顺带放行防火墙, 零新增 UAC)
    [string]$WinbootExe = '',
    # 内部参数: 提权 helper 阶段标记 (只做 WSL+VMP, 不对外)
    [switch]$ElevatedWslPhase
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$env:WSL_UTF8 = '1'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# --- 日志: 每次运行留档, 位于 winboot/logs/<阶段>_<时间戳>.log ---
$logDir = Join-Path (Split-Path $PSScriptRoot -Parent) 'logs'
New-Item -ItemType Directory -Force $logDir | Out-Null
$phaseTag = if ($ElevatedWslPhase) { 'install_wsladmin' } else { 'install' }
Start-Transcript -Path (Join-Path $logDir ("{0}_{1}.log" -f $phaseTag, (Get-Date -Format 'yyyyMMdd_HHmmss'))) | Out-Null

# PS 5.1 坑: EAP=Stop 时对原生程序 stderr 重定向 (2>$null) 会把 stderr 每行包成 ErrorRecord 直接抛异常;
# wsl/docker 在未安装/未就绪时向 stderr 写信息 => 脚本早夭 => 提权窗口闪退 (2026-09 踩坑根因).
# 统一经此 helper 静默执行原生命令, 只回传退出码.
function Invoke-Native {
    param([string]$FilePath, [string[]]$NativeArgs)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        if ($NativeArgs) { $null = & $FilePath @NativeArgs 2>$null }
        else { $null = & $FilePath 2>$null }
        return $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $prev
    }
}

function Stop-AndExit {
    param([int]$Code)
    if (-not $NoPause -and $Host.Name -eq 'ConsoleHost') {
        Write-Host "`n按任意键退出..." -ForegroundColor Cyan
        $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown') | Out-Null
    }
    try { Stop-Transcript | Out-Null } catch { }
    exit $Code
}

# Invoke-Native 的带输出版本: 返回 @{ Exit = 退出码; Out = 首行 stdout }.
# 用途: docker info 探活 —— DD 后端管道在而引擎未起时, "Error response from daemon: ..."
# 打到 stderr 且退出码为 0, 只看退出码会被骗 (2026-09-24 实测: 打出空的"daemon 已就绪").
function Invoke-NativeCapture {
    param([string]$FilePath, [string[]]$NativeArgs)
    $prev = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $out = & $FilePath @NativeArgs 2>$null
        return @{ Exit = $LASTEXITCODE; Out = [string]($out | Select-Object -First 1) }
    } finally {
        $ErrorActionPreference = $prev
    }
}

function Test-Admin {
    return ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# WSL 注册稳定等待: MSIX 刚装完的窗口期, DD 安装器/后端可能探测不到 WSL (实测瞬态报
# "WSL not installed"后自愈). --version 经 System32 桩转发立即通过, 用 --status 轮询 + grace 确保落定.
function Wait-WslSettled {
    $total = 0
    while (((Invoke-Native 'wsl' @('--status')) -ne 0) -and ($total -lt 30)) {
        Start-Sleep -Seconds 3
        $total += 3
        Write-Host "  ... 等待 WSL 注册稳定 (${total}s)"
    }
    Start-Sleep -Seconds 3   # 注册落定后的 grace: 覆盖 DD 探测的窗口期
}

# VMP 特性状态: CBS 注册表直读, 免提权 (CurrentState 112=Enabled)
function Test-VmpEnabled {
    try {
        $pkgs = Get-ChildItem 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\Packages' -ErrorAction Stop |
            Where-Object { $_.PSChildName -like 'HyperV-Feature-VirtualMachinePlatform-Client-Package~*amd64~~*' }
        return @($pkgs | Where-Object { (Get-ItemProperty $_.PSPath -ErrorAction SilentlyContinue).CurrentState -eq 112 }).Count -gt 0
    } catch { return $false }
}

function Find-DockerDesktop {
    foreach ($dir in @((Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop'), (Join-Path $env:ProgramFiles 'Docker\Docker'))) {
        if (Test-Path (Join-Path $dir 'Docker Desktop.exe')) { return $dir }
    }
    return $null
}

# 组件注册表 + 校验 (唯一事实源见 lib/components.ps1, 新增组件 checklist 也在那里)
. (Join-Path $PSScriptRoot 'lib\components.ps1')
$unknown = @($Components | Where-Object { $BootComponents -notcontains $_ })
if ($unknown.Count -gt 0) {
    Write-Host "未知组件: $($unknown -join ', ') (可用: $($BootComponents -join ', '))" -ForegroundColor Red
    Stop-AndExit 1
}

# 防火墙放行 (只在已有管理员上下文里调用 —— 提权 helper 末尾/管理员直装路径;
# GUI 侧 lanshare.ensure_firewall_rule 用同名规则, 两处显示名必须一致).
# 程序级规则 (-Program) + Any profile: 公司网常被 Windows 归类为公用,
# 限定专用/域的规则在实测环境里不生效.
function Enable-WinbootFirewall {
    if (-not $WinbootExe -or -not (Test-Path $WinbootExe)) { return }
    # 注意: -Enabled True 参数在 PS5.1 会静默过滤掉所有规则 (实测), 用属性过滤
    if (Get-NetFirewallRule -DisplayName 'winboot 局域网共享' -ErrorAction SilentlyContinue | Where-Object { $_.Enabled -eq 'True' }) { return }
    try {
        New-NetFirewallRule -DisplayName 'winboot 局域网共享' -Direction Inbound -Action Allow `
            -Program $WinbootExe -Profile Any | Out-Null
        Write-Host '  已顺带放行防火墙 (winboot 入站, 全部 profile, 零新增 UAC)' -ForegroundColor Green
    } catch {
        Write-Host "  防火墙放行失败 (可能组策略/EDR 拦截; GUI 启动后还会再试并给手动指引): $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# 通用文件获取组件 (三级: 缓存 -> 局域网 -> 兜底); 新软件在此之上各定义一个获取条目
. (Join-Path $PSScriptRoot 'lib\acquire.ps1')
# 局域网源覆盖 (GUI 广播发现后传入; 必须在点源之后, 否则被 acquire.ps1 内置值盖回)
if ($LanBase) { $script:BootLanBase = $LanBase }

# --- WSL 安装(含 VMP 特性): 可在提权 helper 或管理员上下文中调用 ---
function Install-WslIfNeeded {
    if (-not (Test-VmpEnabled)) {
        Write-Host "  启用 VirtualMachinePlatform 特性 (dism)..."
        cmd /c "dism /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart" | ForEach-Object { Write-Host "    $_" }
        if ($LASTEXITCODE -eq 3010) {
            Write-Host "`n需要重启 Windows 使特性生效. 重启后请重新执行 boot install 继续." -ForegroundColor Yellow
            return $false
        }
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  启用 VirtualMachinePlatform 失败 (exit $LASTEXITCODE)." -ForegroundColor Red
            return $false
        }
    }
    # WSL 获取四级策略: 缓存(.msix) -> 局域网 -> winget download(产物入缓存) -> winget install 直装.
    # 缓存红利关键: winget download 只下载不安装, 版本化产物重命名为稳定名 wsl-x64.msix 落缓存,
    # 下次循环直接命中缓存层 (2026-09-24 之前兜底直装导致缓存层形同虚设).
    # (MSIX 版 WSL 移除后只剩 System32 内置 wsl.exe 桩, 不支持 --no-distribution; 裸 wsl --install 会强装默认发行版)
    $wslPkg = Join-Path $env:TEMP 'wsl-x64.msix'
    $wslGot = Get-BootFile -Name 'WSL 安装包' -CachePath $wslPkg -MinBytes 100MB -FileName 'wsl-x64.msix' -Fallback {
        # 三级兜底: winget download 下载入缓存 (产物为版本化命名 .msix, 重命名入缓存)
        $dlDir = Join-Path $env:TEMP 'wsl_winget_dl'
        Remove-Item $dlDir -Recurse -Force -ErrorAction SilentlyContinue
        New-Item -ItemType Directory -Force $dlDir | Out-Null
        Write-Host "  [兜底] winget download Microsoft.WSL (下载入缓存)..."
        cmd /c "winget download --id Microsoft.WSL -e -d `"$dlDir`" --accept-source-agreements --accept-package-agreements 2>&1" | ForEach-Object { Write-Host "    $_" }
        if ($LASTEXITCODE -eq 0) {
            $msix = Get-ChildItem $dlDir -Filter '*.msix' -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($msix) {
                Move-Item $msix.FullName $wslPkg -Force
                Remove-Item $dlDir -Recurse -Force -ErrorAction SilentlyContinue
                Write-Host "  [兜底] 已入缓存: $wslPkg"
                return $true
            }
        }
        # 末级兜底: winget install 直装 (不产生缓存, 下次仍需下载)
        Write-Host "  [末级] winget install Microsoft.WSL (直装)..."
        cmd /c "winget install --id Microsoft.WSL -e --silent --accept-source-agreements --accept-package-agreements 2>&1" | ForEach-Object { Write-Host "    $_" }
        return ($LASTEXITCODE -eq 0)
    }
    if (-not $wslGot) {
        Write-Host "  WSL 安装失败: 缓存/局域网/winget 全部失败." -ForegroundColor Red
        return $false
    }
    if (Test-Path $wslPkg) {
        # 文件就位 (缓存/局域网/winget download): 静默安装, 扩展名分派
        if ($wslPkg -like '*.msi') {
            Write-Host "  msiexec 静默安装: $wslPkg"
            $proc = Start-Process msiexec.exe -ArgumentList '/i', "`"$wslPkg`"", '/qn', '/norestart' -Wait -PassThru
            if ($proc.ExitCode -eq 3010) {
                Write-Host "`n需要重启 Windows 使 WSL 生效. 重启后请重新执行 boot install 继续." -ForegroundColor Yellow
                return $false
            }
            if ($proc.ExitCode -ne 0) {
                Write-Host "  WSL MSI 安装失败 (exit $($proc.ExitCode))." -ForegroundColor Red
                return $false
            }
        } else {
            # .msix / .msixbundle (需管理员; 本函数仅在提权上下文被调用)
            Write-Host "  Add-AppxPackage 静默安装: $wslPkg"
            try { Add-AppxPackage -Path $wslPkg -ErrorAction Stop }
            catch {
                Write-Host "  WSL msix 安装失败: $_" -ForegroundColor Red
                return $false
            }
        }
    } else {
        Write-Host "  WSL 已由 winget 直装完成 (未产生缓存)." -ForegroundColor Green
    }
    $null = Invoke-Native 'wsl' @('--set-default-version', '2')
    return $true
}

# ======================================================================
# 提权 helper 阶段: 只做 WSL + VMP, 完成即退 (由主流程经 UAC 派生, -Wait 等待)
# ======================================================================
if ($ElevatedWslPhase) {
    if (-not (Test-Admin)) { Write-Host "helper 未获得管理员权限, 异常退出." -ForegroundColor Red; Stop-AndExit 1 }
    Write-Host "[提权 helper] 安装 WSL / 启用特性..."
    if (Install-WslIfNeeded) {
        Enable-WinbootFirewall   # 顺带放行 (同一次 UAC 内, 零新增弹窗)
        Write-Host "[提权 helper] 完成." -ForegroundColor Green
        Stop-AndExit 0
    } else {
        Stop-AndExit 1
    }
}

# ======================================================================
# 主流程: 普通用户上下文
# ======================================================================

# --- [1] WSL 检查 (免提权) ---
Write-Host "`n[1/3] WSL 检查..." -ForegroundColor Cyan
$isAdmin = Test-Admin

if ($Components -notcontains 'wsl') {
    Write-Host "  跳过 (组件列表: $($Components -join ', '))"
} else {
    $wslOk = ((Invoke-Native 'wsl' @('--version')) -eq 0)
    $vmpOk = Test-VmpEnabled

    if ($wslOk -and $vmpOk) {
        $wslVer = (& wsl --version | Select-Object -First 1)
        Write-Host "  WSL 已就绪: $wslVer (VMP 特性已启用)" -ForegroundColor Green
    } elseif ($isAdmin) {
        # 用户从管理员终端手动运行: 直接在本上下文完成 (DD 安装形态见 [2/3] 提示)
        Write-Host "  管理员上下文检测到 WSL/VMP 缺失, 直接安装..."
        if (-not (Install-WslIfNeeded)) { Stop-AndExit 1 }
        if ((Invoke-Native 'wsl' @('--version')) -ne 0) { Write-Host "  WSL 安装后校验失败." -ForegroundColor Red; Stop-AndExit 1 }
        Wait-WslSettled
        Enable-WinbootFirewall   # 已是管理员上下文, 顺带放行
        Write-Host "  WSL 安装完成." -ForegroundColor Green
    } else {
        # 提权范围最小化: 只为 WSL/VMP 阶段弹一次 UAC, 主流程保持普通权限
        # helper 恒带 -NoPause: 否则其独立控制台窗口干完活停在"按任意键", 卡住整个安装 (2026-09-24 实测);
        # helper 输出在其 transcript (logs\install_wsladmin_*.log) 里, 不依赖窗口驻留.
        Write-Host "  WSL/VMP 缺失, 派提权 helper 安装 (仅此阶段需要 UAC, 完成后自动继续)..." -ForegroundColor Yellow
        try {
            # -LanBase / -WinbootExe 穿透给 helper: WSL 获取在其阶段; 防火墙放行要 exe 路径
            $helperArgs = "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`" -ElevatedWslPhase -NoPause"
            if ($LanBase) { $helperArgs += " -LanBase `"$LanBase`"" }
            if ($env:WINBOOT_EXE) { $helperArgs += " -WinbootExe `"$($env:WINBOOT_EXE)`"" }
            Start-Process powershell.exe -Verb RunAs -Wait -ArgumentList $helperArgs
        } catch {
            Write-Host "  UAC 被取消或 helper 启动失败: $_" -ForegroundColor Red
            Stop-AndExit 1
        }
        # helper 完成后回本进程校验
        if (((Invoke-Native 'wsl' @('--version')) -ne 0) -or -not (Test-VmpEnabled)) {
            Write-Host "  helper 执行后 WSL/VMP 仍未就绪, 详见 logs\install_wsladmin_*.log" -ForegroundColor Red
            Stop-AndExit 1
        }
        Wait-WslSettled
        Write-Host "  WSL 安装完成 (helper 阶段日志: logs\install_wsladmin_*.log)" -ForegroundColor Green
    }
}

# --- [2] Docker Desktop ---
Write-Host "`n[2/3] Docker Desktop 检查..." -ForegroundColor Cyan
$ddDir = $null
if ($Components -contains 'docker') { $ddDir = Find-DockerDesktop }
if ($Components -notcontains 'docker') {
    Write-Host "  跳过 (组件列表: $($Components -join ', '))"
} elseif ($ddDir) {
    Write-Host "  Docker Desktop 已安装: $ddDir" -ForegroundColor Green
} else {
    if (-not $Yes) {
        $answer = Read-Host "  将安装 Docker Desktop (默认当前用户, 免 UAC), 继续? (回车确认)"
        if ($answer -ne '') { Write-Host "已取消."; Stop-AndExit 1 }
    }

    if ($InstallerPath -and (Test-Path $InstallerPath)) {
        $installer = (Resolve-Path $InstallerPath).Path
        Write-Host "  使用本地安装器(显式指定): $installer"
    } else {
        # 三级获取走通用组件 (lib/acquire.ps1): 缓存 -> 局域网 -> 官方
        $installer = Join-Path $env:TEMP 'DockerDesktopInstaller.exe'
        if ($InstallerUrl) {
            # 显式指定源 = 明确意图: 跳过缓存与局域网, 只从这里下
            if (-not (Save-BootFileFromUrl -Name 'DD 安装器(指定源)' -Url $InstallerUrl -Dest $installer)) {
                Write-Host "  指定源下载失败. 可 -InstallerPath 指定本地安装器." -ForegroundColor Red
                Stop-AndExit 1
            }
        } else {
            $got = Get-BootFile -Name 'DD 安装器' -CachePath $installer -MinBytes 200MB -FileName 'DockerDesktopInstaller.exe' `
                -OfficialUrl 'https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe'
            if (-not $got) {
                Write-Host "  可 -InstallerPath 指定本地安装器, 或 -InstallerUrl 指定其他源." -ForegroundColor Red
                Stop-AndExit 1
            }
        }
    }

    # --- 安装: --user = 当前用户安装 (%LOCALAPPDATA%, 免 UAC, 官方文档 flag); 缺省 all-users 需管理员 ---
    $installArgs = @('install', '--accept-license', '--quiet', '--backend=wsl-2')
    if ($AllUsers) {
        if (-not $isAdmin) {
            Write-Host "  静默安装 (所有用户模式, 安装器将请求 UAC)..."
            $proc = Start-Process -FilePath $installer -ArgumentList $installArgs -Verb RunAs -Wait -PassThru
        } else {
            Write-Host "  静默安装 (所有用户模式, 管理员上下文)..."
            $proc = Start-Process -FilePath $installer -ArgumentList $installArgs -Wait -PassThru
        }
    } else {
        # 当前用户安装: 注意 per-user 模式不支持 Windows 容器 (本工具用 WSL2 后端跑 Linux 容器, 无影响)
        Write-Host "  静默安装 (当前用户, --user, 免 UAC)..."
        $proc = Start-Process -FilePath $installer -ArgumentList ($installArgs + '--user') -Wait -PassThru
    }
    Write-Host ("  安装器退出码: {0}" -f $proc.ExitCode)

    $ddDir = Find-DockerDesktop
    if (-not $ddDir) {
        Write-Host "  安装后未检出 Docker Desktop, 请用 boot doctor 复查." -ForegroundColor Red
        Stop-AndExit 1
    }
    Write-Host "  Docker Desktop 安装完成: $ddDir" -ForegroundColor Green
}

if ($Components -contains 'docker') {
    # --- 预置 daemon.json (必须在 DD 首启前; 私仓是 HTTP 明文 Harbor, 不配 insecure-registries 则登录/拉取全失败) ---
    $daemonJson = Join-Path $env:USERPROFILE '.docker\daemon.json'
    if (-not (Test-Path $daemonJson)) {
        New-Item -ItemType Directory -Force (Split-Path $daemonJson) | Out-Null
        # daemon.json 预置: live-restore + 100GB 构建缓存 + 调用方传入的 insecure 私仓
        # (私仓地址因环境而异, 经 -InsecureRegistries 传入, 不内置任何特定地址)
        $regLines = @($InsecureRegistries | ForEach-Object { '    "{0}"' -f $_ })
        $regJson = if ($regLines.Count) { $regLines -join ",`n" } else { '' }
        $daemonConfig = @"
{
  "builder": {
    "gc": {
      "defaultKeepStorage": "100GB",
      "enabled": true
    }
  },
  "experimental": false,
  "insecure-registries": [
$regJson
  ],
  "live-restore": true
}
"@
        [System.IO.File]::WriteAllText($daemonJson, $daemonConfig)
        Write-Host "  已预置 daemon.json (insecure-registries / live-restore / 100GB 构建缓存)" -ForegroundColor Green
    } else {
        Write-Host "  daemon.json 已存在, 保留现有配置: $daemonJson"
    }
}

# --- [3] 启动并等待 daemon 就绪 (始终用户上下文; 提权上下文启动 DD 曾出现 daemon 不起的问题) ---
if ($Components -notcontains 'docker') {
    Write-Host "`n[3/3] 跳过 (未安装 docker 组件)." -ForegroundColor Cyan
} else {
    Write-Host "`n[3/3] 启动 Docker Desktop 并等待 daemon 就绪 (首次启动含 WSL 数据盘初始化, 最长 5 分钟)..." -ForegroundColor Cyan
    $ddExe = Join-Path $ddDir 'Docker Desktop.exe'
    $dockerExe = Join-Path $ddDir 'resources\bin\docker.exe'
    if (-not (Test-Path $dockerExe)) { $dockerExe = 'docker' }

    # 就绪判定 = 退出码 0 且输出了非空版本号 (防 "Error response from daemon" 退出码 0 的假就绪)
    $probe = Invoke-NativeCapture $dockerExe @('info', '--format', '{{.ServerVersion}}')
    if (($probe.Exit -ne 0) -or -not $probe.Out) {
        Start-Process -FilePath $ddExe
    }

    $deadline = (Get-Date).AddMinutes(5)
    $ready = $false
    $ver = ''
    while ((Get-Date) -lt $deadline) {
        $probe = Invoke-NativeCapture $dockerExe @('info', '--format', '{{.ServerVersion}}')
        if (($probe.Exit -eq 0) -and $probe.Out) { $ready = $true; $ver = $probe.Out; break }
        Start-Sleep -Seconds 5
        Write-Host '  ... 等待 daemon'
    }
    if ($ready) {
        Write-Host "  daemon 已就绪, server $ver" -ForegroundColor Green
    } else {
        Write-Host "  等待超时. 请确认 Docker Desktop 窗口无报错后用 boot doctor 复查." -ForegroundColor Red
        Stop-AndExit 1
    }
}

Write-Host "`n安装流程结束. 建议: boot doctor 验收 -> boot up 启动开发环境." -ForegroundColor Cyan
Stop-AndExit 0
