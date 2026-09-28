#Requires -Version 5.1

<#
.SYNOPSIS
    组件注册表 (唯一事实源): install/uninstall 的缺省全套、uninstall-all 别名、
    组件名校验全部从这里取值.
.DESCRIPTION
    ── 新增组件 checklist (以 git 为例) ─────────────────────────────
    1. 此处登记:           $BootComponents 数组加 'git'
    2. boot_install.ps1:   新增 [x/n] 检测+获取(acquire.ps1 复用)+安装段
    3. boot_uninstall.ps1: 新增对应卸载清理段
    4. (可选) lanshare.py 的 SHAREABLE: 安装包想进局域网共享/预取才登记
    5. (可选) GUI: 组件点选列表 (透传位置参数即生效)
    ────────────────────────────────────────────────────────────────
.NOTES
    注意: 本文件必须保存为 UTF-8 with BOM (PS 5.1 兼容).
#>

$BootComponents = @('wsl', 'docker')
