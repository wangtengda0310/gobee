#Requires -Version 5.1
<#
.SYNOPSIS
    MCP 游戏数据 by_name 工具测试脚本（hero/card/skill）
.DESCRIPTION
    使用 PowerShell 调用 curl.exe 发送 UTF-8 JSON 请求，避免 bat 中 curl -d 中文 argv 编码损坏。
    直接解析 StreamableHTTP 返回的 SSE 响应。
    验证 get_hero_cfg_by_name / get_card_cfg_by_name / get_skill_cfg_by_name 的精准查询。
    注意：同名多条（如"杀"）按数组返回，断言用 .Count。
#>
[CmdletBinding()]
param(
    [string]$McpUrl = "http://127.0.0.1:8765"
)

$ErrorActionPreference = "Stop"

# 强制使用 UTF-8 输出
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'

$script:testId = 0
$script:passCount = 0
$script:failCount = 0

function Invoke-McpTool {
    param(
        [string]$Name,
        [object]$Arguments,
        [string]$TestDescription,
        [scriptblock]$Assert = $null
    )

    $script:testId++
    $id = $script:testId
    Write-Host "[$id] $Name ..."

    $body = @{
        jsonrpc = "2.0"
        method  = "tools/call"
        params  = @{
            name      = $Name
            arguments = $Arguments
        }
        id      = $id
    } | ConvertTo-Json -Compress -Depth 10

    $tempFile = [System.IO.Path]::GetTempFileName()
    try {
        $noBomUtf8 = New-Object System.Text.UTF8Encoding($false)
        [System.IO.File]::WriteAllText($tempFile, $body, $noBomUtf8)

        $output = & curl.exe -s -X POST $McpUrl `
            -H "Accept: application/json, text/event-stream" `
            -H "Content-Type: application/json; charset=utf-8" `
            --data-binary "@$tempFile"

        if ($LASTEXITCODE -ne 0) {
            throw "curl 请求失败，退出码: $LASTEXITCODE"
        }

        # 解析 SSE 响应，提取 data: 行
        $dataLines = $output -split "`r?`n" | Where-Object { $_ -match "^data:\s*(.+)$" } | ForEach-Object { $matches[1] }
        if ($dataLines.Count -eq 0) {
            throw "响应中未找到 SSE data 行，原始响应: $output"
        }

        $resp = $dataLines | Select-Object -Last 1 | ConvertFrom-Json
        if ($resp.error) {
            throw "MCP 返回错误: $($resp.error | ConvertTo-Json -Compress)"
        }
        $text = $resp.result.content | Where-Object { $_.type -eq 'text' } | Select-Object -ExpandProperty text
        if (-not $text) {
            throw "响应中未找到文本内容，原始响应: $output"
        }

        if ($Assert) {
            & $Assert $text
        }

        Write-Host "[PASS] $Name - $TestDescription"
        return $true
    }
    catch {
        Write-Host "[FAIL] $Name - $_"
        return $false
    }
    finally {
        if (Test-Path $tempFile) {
            Remove-Item $tempFile -Force
        }
    }
}

Write-Host "========================================"
Write-Host " MCP 游戏数据 by_name 工具测试"
Write-Host " URL: $McpUrl"
Write-Host "========================================"

# ---- get_hero_cfg_by_name ----
$result = Invoke-McpTool `
    -Name "get_hero_cfg_by_name" `
    -Arguments @{ hero_name = "赵云" } `
    -TestDescription "按单个英雄名查询（赵云）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.heroes."赵云".Count -lt 1) { throw "heroes 中未找到赵云" }
        if ($data.heroes."赵云"[0].Id -ne 10105) { throw "赵云 Id 应为 10105，实际: $($data.heroes.'赵云'[0].Id)" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

$result = Invoke-McpTool `
    -Name "get_hero_cfg_by_name" `
    -Arguments @{ hero_name = @("赵云", "孙尚香") } `
    -TestDescription "批量英雄名查询（赵云+孙尚香）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.heroes."赵云".Count -lt 1 -or $data.heroes."孙尚香".Count -lt 1) { throw "未同时找到赵云和孙尚香" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

$result = Invoke-McpTool `
    -Name "get_hero_cfg_by_name" `
    -Arguments @{ hero_name = "查无此武将" } `
    -TestDescription "查询不存在的英雄名" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.heroes."查无此武将".Count -gt 0) { throw "heroes 不应包含该项" }
        if ($data.notFound -notcontains "查无此武将") { throw "notFound 应包含该名称" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

# ---- get_card_cfg_by_name ----
$result = Invoke-McpTool `
    -Name "get_card_cfg_by_name" `
    -Arguments @{ card_name = "杀" } `
    -TestDescription "查卡牌'杀'（应返回多个 id：普杀/火杀/雷杀）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.cards."杀".Count -lt 2) { throw "'杀' 应返回多个卡牌（普杀/火杀/雷杀），实际数量: $($data.cards.'杀'.Count)" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

$result = Invoke-McpTool `
    -Name "get_card_cfg_by_name" `
    -Arguments @{ card_name = @("闪", "桃") } `
    -TestDescription "批量卡牌查询（闪+桃）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.cards."闪".Count -lt 1 -or $data.cards."桃".Count -lt 1) { throw "未同时找到闪和桃" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

# ---- get_skill_cfg_by_name ----
$result = Invoke-McpTool `
    -Name "get_skill_cfg_by_name" `
    -Arguments @{ skill_name = "七进七出" } `
    -TestDescription "按技能名查询（七进七出）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.skills."七进七出".Count -lt 1) { throw "skills 中未找到七进七出" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

$result = Invoke-McpTool `
    -Name "get_skill_cfg_by_name" `
    -Arguments @{ skill_name = @("七进七出", "枪出如龙") } `
    -TestDescription "批量技能查询（七进七出+枪出如龙）" `
    -Assert {
        param($text)
        $data = $text | ConvertFrom-Json
        if ($data.skills."七进七出".Count -lt 1 -or $data.skills."枪出如龙".Count -lt 1) { throw "未同时找到七进七出和枪出如龙" }
        if ($data.notFound.Count -gt 0) { throw "notFound 应为空" }
    }
if ($result) { $script:passCount++ } else { $script:failCount++ }

Write-Host ""
Write-Host "========================================"
Write-Host " Test Summary"
Write-Host "========================================"
Write-Host " Total Tests: $($script:testId)"
Write-Host " Passed:      $($script:passCount)"
Write-Host " Failed:      $($script:failCount)"
Write-Host "========================================"

if ($script:failCount -gt 0) {
    Write-Host "[RESULT] TEST FAILED - $($script:failCount) test(s) failed"
    exit 1
}
else {
    Write-Host "[RESULT] ALL TESTS PASSED"
    exit 0
}
