# sheetcmp Git/UGit 接入脚本: 配置当前用户的差异与合并工具。
# 用法: powershell -NoProfile -ExecutionPolicy Bypass -File .\configure-git.ps1 [-ExePath <绝对路径>]
# 默认使用本仓库构建产物 ..\..\bin\sheetcmp.exe; 目录移动后需重跑。

param(
    [string]$ExePath = ""
)

$ErrorActionPreference = "Stop"

# 默认定位: 脚本在 sheetcmp/scripts/ 下, 产物在 sheetcmp/bin/
if (-not $ExePath) {
    $ExePath = Join-Path (Split-Path $PSScriptRoot -Parent) "bin\sheetcmp.exe"
}
if (-not (Test-Path $ExePath)) {
    Write-Host "未找到 sheetcmp.exe: $ExePath" -ForegroundColor Red
    Write-Host "先构建: cd frontend; npm install; npm run build; cd ..; go build -o bin/sheetcmp.exe ./cmd/sheetcmp"
    exit 1
}
$ExePath = (Resolve-Path $ExePath).Path

function Set-GitConfig {
    param([string]$Key, [string]$Value)
    git config --global $Key $Value
    Write-Host ("  {0} = {1}" -f $Key, $Value) -ForegroundColor Gray
}

Write-Host "配置 sheetcmp 为全局差异/合并工具 (exe: $ExePath)"
# 差异工具: 左右两文件
Set-GitConfig "difftool.sheetcmp.cmd" "`"$ExePath`" `"`$LOCAL`" `"`$REMOTE`""
# 合并工具: -MergeTool REMOTE MERGED, 退出码可信 (0=已解决, 1=放弃)
Set-GitConfig "mergetool.sheetcmp.cmd" "`"$ExePath`" -MergeTool `"`$REMOTE`" `"`$MERGED`""
Set-GitConfig "mergetool.sheetcmp.trustExitCode" "true"
# 设为默认 GUI 工具 (UGit/GitGUI 读取)
Set-GitConfig "diff.guitool" "sheetcmp"
Set-GitConfig "merge.guitool" "sheetcmp"

Write-Host ""
Write-Host "完成。验证: git config --global --get-regexp 'difftool|mergetool'" -ForegroundColor Green
Write-Host "UGit 已打开时需重启后生效。"
