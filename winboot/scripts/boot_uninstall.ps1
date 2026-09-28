#Requires -Version 5.1

<#
.SYNOPSIS
    boot uninstall: 卸载 Docker Desktop 并清理 WSL 残留数据盘.
.DESCRIPTION
    流程: 确认(默认需 -Yes 或交互确认) -> 停止 DD 进程 -> 静默卸载 -> wsl --unregister docker-desktop
          -> (可选 -RemoveWsl) wsl --uninstall 移除 WSL 本体.
    非管理员运行时会自动弹 UAC 以管理员重启自身.
.PARAMETER Yes
    跳过交互确认 (自动提权重启时使用).
.PARAMETER RemoveWsl
    额外卸载 WSL 本体 (默认保留; 若机器上还有其他发行版在用, 不要加).
.NOTES
    警告: docker-desktop 数据盘注销后, 本地所有镜像/容器/卷不可恢复地丢失.
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 兼容).
#>

param(
    [switch]$Yes,
    # 兼容别名: 等价于组件列表包含 wsl (新代码请用 -Components)
    [switch]$RemoveWsl,
    # 跳过"按任意键退出"暂停 (GUI/隐藏控制台调用必须加, 否则 ReadKey 永久阻塞)
    [switch]$NoPause,
    # 组件列表 (位置参数, GUI 点选/预设套装透传); 直调本脚本缺省仅 docker (保守).
    # 校验在运行时对 lib/components.ps1 注册表做 (ValidateSet 必须字面量, 无法引用注册表)
    [string[]]$Components = @('docker')
)

# 兼容: -RemoveWsl 等价于组件列表包含 wsl
if ($RemoveWsl -and ($Components -notcontains 'wsl')) { $Components += 'wsl' }

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$env:WSL_UTF8 = '1'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# --- 日志: 每次运行留档 (含提权实例), 位于 winboot/logs/uninstall_<时间戳>.log ---
$logDir = Join-Path (Split-Path $PSScriptRoot -Parent) 'logs'
New-Item -ItemType Directory -Force $logDir | Out-Null
Start-Transcript -Path (Join-Path $logDir ("uninstall_{0}.log" -f (Get-Date -Format 'yyyyMMdd_HHmmss'))) | Out-Null

# PS 5.1 坑: EAP=Stop 时原生 stderr 重定向 (2>$null) 会抛 NativeCommandError (与 boot_install 同源),
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

# 组件注册表 + 校验 (唯一事实源见 lib/components.ps1, 新增组件 checklist 也在那里)
. (Join-Path $PSScriptRoot 'lib\components.ps1')
$unknown = @($Components | Where-Object { $BootComponents -notcontains $_ })
if ($unknown.Count -gt 0) {
    Write-Host "未知组件: $($unknown -join ', ') (可用: $($BootComponents -join ', '))" -ForegroundColor Red
    Stop-AndExit 1
}

# --- 提权策略 (最小化): ---
# - 仅 docker 组件的卸载不提权: per-user 安装(默认)的卸载器免 UAC; all-users 安装仅在调用卸载器时经 RunAs 提权;
# - 含 wsl 组件 (移除 WSL 机器级组件)整体提权一次.
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (($Components -contains 'wsl') -and -not $isAdmin) {
    Write-Host "卸载 WSL 组件需要管理员 (机器级组件), 正在请求 UAC..." -ForegroundColor Yellow
    # 坑: powershell.exe -File 模式下 "-Components wsl,docker" 会作为单个字符串绑定, ValidateSet 直接拒绝
    # (2026-09-24 实测: 提权实例静默死亡、无日志无报错), 必须用 -Command 走真正的数组解析.
    $inner = "& '$PSCommandPath' -Yes -Components $($Components -join ',')"
    if ($NoPause) { $inner += ' -NoPause' }
    $argList = "-NoProfile -ExecutionPolicy Bypass -Command `"$inner`""
    # -Wait: 等提权实例真正干完 (否则 GUI 在 UAC 确认后立即报"完成", 实际卸载还要跑十几秒,
    # 用户此时点安装会撞上正在卸载的环境; 2026-09-24 日志实证: 父进程提前 19 秒退出)
    $proc = Start-Process powershell.exe -Verb RunAs -ArgumentList $argList -Wait -PassThru
    exit $proc.ExitCode
}

# --- 定位 Docker Desktop (per-user / all-users 双形态, 与 doctor 同逻辑) ---
function Find-DockerDesktop {
    foreach ($dir in @((Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop'), (Join-Path $env:ProgramFiles 'Docker\Docker'))) {
        $exe = Join-Path $dir 'Docker Desktop.exe'
        if (Test-Path $exe) { return $dir }
    }
    return $null
}
$ddDir = $null
if ($Components -contains 'docker') { $ddDir = Find-DockerDesktop }

if ($Components -notcontains 'docker') {
    Write-Host "跳过 Docker Desktop 组件 (组件列表: $($Components -join ', '))." -ForegroundColor Cyan
} elseif (-not $ddDir) {
    Write-Host "未找到 Docker Desktop 安装 (per-user / all-users 均未检出), 无需卸载." -ForegroundColor Green
} else {
    # --- 确认门 ---
    if (-not $Yes) {
        Write-Host "即将卸载组件: $($Components -join ', ')" -ForegroundColor Yellow
        if ($ddDir) { Write-Host "  Docker Desktop 位置: $ddDir (数据盘注销 => 本地镜像/容器/卷全部丢失!)" }
        if ($Components -contains 'wsl') { Write-Host "  卸载 WSL 本体 => 机器上所有发行版将被移除!" }
        $answer = Read-Host "确认卸载? 输入 YES 继续"
        if ($answer -ne 'YES') { Write-Host "已取消."; Stop-AndExit 1 }
    }

    # --- 停止 Docker Desktop 相关进程 (backend 随主进程退出) ---
    Write-Host "停止 Docker Desktop 进程..."
    foreach ($procName in 'Docker Desktop', 'com.docker.backend', 'com.docker.build', 'com.docker.dev-envs') {
        Stop-Process -Name $procName -Force -ErrorAction SilentlyContinue
    }
    Start-Sleep -Seconds 3

    # --- 静默卸载 ---
    $uninstaller = Join-Path $ddDir 'Docker Desktop Installer.exe'
    if (-not (Test-Path $uninstaller)) {
        Write-Host "卸载器不存在: $uninstaller, 尝试注册表卸载串..." -ForegroundColor Yellow
        $reg = Get-ItemProperty 'HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop' -ErrorAction SilentlyContinue
        if (-not $reg) { $reg = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop' -ErrorAction SilentlyContinue }
        if ($reg) { Write-Host "注册表卸载串: $($reg.UninstallString)" } else { Write-Host "注册表也无卸载项, 手动删除目录: $ddDir" }
        Stop-AndExit 1
    }
    # 卸载形态决定是否需要管理员: all-users(Program Files)的卸载器要提权, per-user 免 UAC
    $isAllUsers = $ddDir.StartsWith($env:ProgramFiles, [StringComparison]::OrdinalIgnoreCase)
    if ($isAllUsers) {
        Write-Host "执行静默卸载 (all-users 安装, 卸载器将请求 UAC)..."
        $proc = Start-Process -FilePath $uninstaller -ArgumentList 'uninstall', '--quiet' -Verb RunAs -Wait -PassThru
    } else {
        Write-Host "执行静默卸载 (per-user 安装, 免 UAC)..."
        $proc = Start-Process -FilePath $uninstaller -ArgumentList 'uninstall', '--quiet' -Wait -PassThru
    }
    Write-Host ("卸载器退出码: {0}" -f $proc.ExitCode)

    # --- 验证卸载结果 ---
    if (Test-Path (Join-Path $ddDir 'Docker Desktop.exe')) {
        Write-Host "警告: 主程序仍存在, 卸载可能未完全成功, 请用 boot doctor 复查." -ForegroundColor Red
    } else {
        Write-Host "Docker Desktop 主程序已移除." -ForegroundColor Green
    }
}

# --- 清理 docker-desktop 数据盘 (docker 组件, 仅当 DD 本体已缺席时才执行) ---
if ($Components -contains 'docker') {
    $ddAfter = Find-DockerDesktop
    if (-not $ddAfter) {
        # 经 cmd 包装捕获 stdout + 丢弃 stderr, 避免 PS 5.1 的 NativeCommandError 陷阱
        $distroLines = @(cmd /c "wsl --list --quiet 2>nul")
        if (($LASTEXITCODE -eq 0) -and ($distroLines -contains 'docker-desktop')) {
            Write-Host "注销 docker-desktop WSL 数据盘 (清除全部镜像/容器/卷)..."
            & wsl --unregister docker-desktop
            if ($LASTEXITCODE -eq 0) { Write-Host "docker-desktop 数据盘已注销." -ForegroundColor Green }
            else { Write-Host "注销失败, 请手动执行: wsl --unregister docker-desktop" -ForegroundColor Red }
        } else {
            Write-Host "无 docker-desktop 发行版残留." -ForegroundColor Green
        }
    }
}

# --- WSL 组件卸载 ---
if ($Components -contains 'wsl') {
    Write-Host "卸载 WSL 本体 (wsl --uninstall)..."
    $unExit = Invoke-Native 'wsl' @('--uninstall')
    if ($unExit -eq 0) {
        Write-Host "WSL 已卸载." -ForegroundColor Green
    } else {
        Write-Host "wsl --uninstall 失败, 尝试 Remove-AppxPackage..."
        Get-AppxPackage -Name 'MicrosoftCorporationII.WindowsSubsystemForLinux' -ErrorAction SilentlyContinue | Remove-AppxPackage
    }
} else {
    Write-Host "WSL 本体保留 (如需卸载: boot uninstall wsl)"
}

Write-Host "`n卸载流程结束. 建议执行 boot doctor 验收 (预期: DockerDesktop MISSING / distro 无残留)." -ForegroundColor Cyan
Stop-AndExit 0
