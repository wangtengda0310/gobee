#Requires -Version 5.1

<#
.SYNOPSIS
    boot 工具入口 (命令行形态; GUI (../main.py) 也经由此脚本调度, 用法一致).
.EXAMPLE
    .\boot.ps1 doctor      # 环境自检
    .\boot.ps1 up          # 启动开发环境 (原 win_boot.ps1)
    .\boot.ps1 stop        # 停止容器 (保留容器/网络, 下次 up 秒起)
    .\boot.ps1 restart     # 重启容器 (不动容器, 只重启内部进程)
    .\boot.ps1 logs        # 跟随查看容器日志
    .\boot.ps1 uninstall   # 卸载 (缺省全套 wsl+docker, 与 install 对称; 注意 wsl 会灭掉机器上所有发行版)
    .\boot.ps1 uninstall docker # 只卸指定组件 (wsl/docker)
    .\boot.ps1 uninstall-all # 兼容别名 (= uninstall wsl docker)
    .\boot.ps1 install     # 装 WSL + Docker Desktop
    .\boot.ps1 gui         # 引导获取并启动 GUI (第一份 exe 的入口; 见 boot_gui.ps1)
.NOTES
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 按 GBK 解析无 BOM 文件会语法错误).
#>

param(
    [Parameter(Position = 0)]
    [ValidateSet('doctor', 'up', 'stop', 'restart', 'logs', 'uninstall', 'uninstall-all', 'install', 'gui')]
    [string]$Command = 'doctor',
    # 收集子命令的透传参数 (如 uninstall -RemoveWsl / install -InstallerPath <path>).
    # 必须显式声明而非用 $args: StrictMode 下无额外参数时 $args 未定义, 引用即抛错 (Go 调用踩坑).
    [Parameter(ValueFromRemainingArguments = $true)]
    $Rest
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# 控制台按 UTF-8 解码原生程序输出 (wsl/docker/winget 均为 UTF-8), 避免 "鐗堟湰" 类乱码
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 组件注册表 (唯一事实源): 缺省全套 / uninstall-all 别名 / 组件名校验均取自此处
. (Join-Path $PSScriptRoot 'lib\components.ps1')

# 脚本位于 server\tools\winboot\scripts\, 项目 server 根在三级向上
# (compose 挂载 / .server_id / win_boot.ps1 都以 server\ 为根)
$serverRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
Set-Location -Path $serverRoot

# $Rest (字符串数组) 统一解析: 位置参数 = 组件名, '-Xxx' = flag (转 hashtable splat 才能绑定命名参数).
# 关键坑 (2026-09-28 跨机实测): 带 value 的 flag (-LanBase <url> / -ServerId <id> / -Mode 1) 的
# value 不以 '-' 开头, 按"前缀分类"会被误判成位置参数(组件名), ValidateSet 拒绝导致 install 失败.
# 必须: 已知带值 flag 走白名单顺序配对, 其余 flag 一律视为开关.
$ValuedFlags = @('LanBase', 'InstallerPath', 'InstallerUrl', 'UpdateBase', 'ServerId', 'Mode')

function Split-RestArgs {
    param($RestArgs)
    $comps = @(); $flags = @{}
    $items = @($RestArgs)
    for ($i = 0; $i -lt $items.Count; $i++) {
        $it = $items[$i]
        if ($it -is [string] -and $it -like '-*') {
            $name = $it.TrimStart('-')
            $next = $null
            if (($i + 1) -lt $items.Count) { $next = $items[$i + 1] }
            if (($ValuedFlags -contains $name) -and ($null -ne $next) -and ($next -is [string]) -and ($next -notlike '-*')) {
                $flags[$name] = $next; $i++   # 带值 flag: 吃掉下一个作为值
            } else {
                $flags[$name] = $true          # 开关
            }
        } elseif ($it -is [string]) {
            $comps += $it                      # 位置参数 = 组件名
        }
    }
    return @{ Comps = $comps; Flags = $flags }
}

$split = if ($PSBoundParameters.ContainsKey('Rest')) { Split-RestArgs $Rest } else { @{ Comps = @(); Flags = @{} } }
$restComps = @($split['Comps'])
# 缺省全套 (与一键安装对称); 前端点选/预设套装即透传位置参数
if ($restComps.Count -eq 0) { $restComps = @($BootComponents) }
$restFlags = $split['Flags']

switch ($Command) {
    'doctor' {
        & (Join-Path $PSScriptRoot 'boot_doctor.ps1')
    }
    'up' {
        # 透传 GUI 参数: -Detach(后台 up) / -Mode 1/2/3(免交互选模式) / -NoPause / -ServerId <id>
        & (Join-Path $serverRoot 'win_boot.ps1') @restFlags
    }
    'stop' {
        # PATH 未命中时从两种安装形态的已知路径补入本会话 PATH (与 win_boot 同逻辑)
        if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
            foreach ($bin in @((Join-Path $env:ProgramFiles 'Docker\Docker\resources\bin'), (Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop\resources\bin'))) {
                if (Test-Path (Join-Path $bin 'docker.exe')) { $env:PATH = "$bin;$env:PATH"; break }
            }
        }
        docker stop sw_server_dev
        if ($LASTEXITCODE -ne 0) { Write-Host "停止失败: 容器 sw_server_dev 不存在或未运行." -ForegroundColor Red }
    }
    'restart' {
        if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
            foreach ($bin in @((Join-Path $env:ProgramFiles 'Docker\Docker\resources\bin'), (Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop\resources\bin'))) {
                if (Test-Path (Join-Path $bin 'docker.exe')) { $env:PATH = "$bin;$env:PATH"; break }
            }
        }
        docker restart sw_server_dev
        if ($LASTEXITCODE -ne 0) { Write-Host "重启失败: 容器 sw_server_dev 不存在 (先 boot up)." -ForegroundColor Red }
    }
    'logs' {
        if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
            foreach ($bin in @((Join-Path $env:ProgramFiles 'Docker\Docker\resources\bin'), (Join-Path $env:LOCALAPPDATA 'Programs\DockerDesktop\resources\bin'))) {
                if (Test-Path (Join-Path $bin 'docker.exe')) { $env:PATH = "$bin;$env:PATH"; break }
            }
        }
        docker logs -f --tail 100 sw_server_dev
    }
    'uninstall' {
        # 位置参数 = 组件列表 (缺省全套, 与一键安装对称); flag 透传 (-Yes/-NoPause)
        & (Join-Path $PSScriptRoot 'boot_uninstall.ps1') @restFlags -Components $restComps
    }
    'uninstall-all' {
        # 兼容别名: 全组件卸载 (= uninstall <注册表全部组件>)
        & (Join-Path $PSScriptRoot 'boot_uninstall.ps1') @restFlags -Components $BootComponents
    }
    'install' {
        # 位置参数 = 组件列表 (缺省全套); flag 透传 (-Yes/-NoPause/-AllUsers/-InstallerPath/-InstallerUrl)
        & (Join-Path $PSScriptRoot 'boot_install.ps1') @restFlags -Components $restComps
    }
    'gui' {
        # 引导获取并启动 GUI: 已有 release\winboot.exe 直接启动, 缺失则三级获取
        # (缓存 -> 局域网同类(UDP 发现) -> 中心). flag 透传 (-UpdateBase/-NoLaunch/-Force)
        & (Join-Path $PSScriptRoot 'boot_gui.ps1') @restFlags
    }
}

# 透传子脚本/原生命令的退出码: & 调用的 exit 不会自动冒泡到本进程的退出码
# (实测 dispatcher 会把失败吞成 0, GUI 依赖进程退出码判定成败).
# 子脚本未显式 exit 时 $LASTEXITCODE 保持上一状态, null 则视为成功.
if ($null -ne $LASTEXITCODE) { exit $LASTEXITCODE }
