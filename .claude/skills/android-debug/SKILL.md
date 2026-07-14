---
name: android-debug
description: 在 Android 模拟器/真机上调试 rain-qa-func(Wails v3) 前后端运行时问题。当需要查看应用是否正常启动、读前端 DOM、点击导航触发后端调用、抓启动崩溃日志、验证 MCP/proto-test server 是否监听、排查"某功能在手机上能否用"、Android 适配验证时使用。提供 CDP 远程调试(cdp_eval.ps1)、logcat、截图、端口扫描四条验证链。后端已实证基本可用，重心通常在前端适配或输入数据可达性，而非代码兼容性。
---

# Android 运行时调试（rain-qa-func）

调试 rain-qa-func 的 Android 构建（Wails v3 + WebView）。**后端 Go 代码在 Android 已实证基本全面可用**（零 CGO、Service 调用链/文件 IO/server 监听均工作），所以排查重心通常是「前端适配」「输入数据可达性（配表/资源目录）」或「真实外部依赖（游戏服地址）」，而非代码兼容性。完整结论见 [docs/Android-运行时调试.md](../../../docs/Android-运行时调试.md)。

## 前置

- 设备/模拟器已 adb 连接（`adb devices`）
- APK 已安装并启动：
  ```powershell
  $adb = "D:\Android\Sdk\platform-tools\adb.exe"
  & $adb shell am start -n com.wails.app/.MainActivity
  ```
- 构建/安装流程见 [docs/Android-APK构建.md](../../../docs/Android-APK构建.md)

## 四条调试链（按强度排序）

### 1. CDP 远程调试（首选 · 最强）

Wails Android **默认开启 WebView 调试**。封装脚本 [build/android/scripts/cdp_eval.ps1](../../../build/android/scripts/cdp_eval.ps1) 自动完成「查 pid → adb forward 9223 → 取 page → 连 WebSocket → 执行 JS」：

```powershell
# 读首页可见文本，判断应用状态/当前页面
& build/android/scripts/cdp_eval.ps1 -Js "String(document.body.innerText).slice(0,1500)"

# 复杂脚本用文件传入（避免命令行转义）
& build/android/scripts/cdp_eval.ps1 -JsFile snippet.js
```

能力：读 DOM、点击导航/按钮**触发后端 Service 调用**（功能可用性的直接证据）、检查 `window.wails`、Promise 自动 await。

点击导航触发后端调用并读取结果（每个页面 mount 时会调后端加载数据，加载成功 = 该后端功能可用）：
```javascript
// 导航用 menuItems index(全 ASCII,避免中文 GBK 坑,见下方"CDP JS ASCII 规范"):
// 0设置/1AI助手/2战斗测试/3配表测试/4武将资源检查/5武将Wiki检查/6活动Wiki/7Proto测试
new Promise(function(r){
  var b = document.querySelectorAll('.idea-icon-button')[3]; // 配表测试
  if (b) b.click();
  setTimeout(function(){ r(String(document.body.innerText).slice(0,1500)) }, 2500);
})
```

> 手动等价：`adb forward tcp:9223 localabstract:webview_devtools_remote_<pid>`，然后浏览器 `chrome://inspect` 或 `GET http://127.0.0.1:9223/json` 取 page webSocketDebuggerUrl。

### 2. adb logcat（看启动与崩溃）

```powershell
$adb = "D:\Android\Sdk\platform-tools\adb.exe"
& $adb logcat -c
& $adb shell am force-stop com.wails.app
& $adb shell am start -n com.wails.app/.MainActivity
Start-Sleep -Seconds 8
& $adb logcat -d | Select-String "Wails|AndroidRuntime|panic|fatal"
```

⚠️ **关键**：Go 业务代码的 `log.Printf`（`[main]`/`InitExcel`/`MCP` 等）**不进 logcat** —— 它们走前端 serverLog 面板（见下）。logcat 主要看：Wails 桥日志（`D Wails`/`WailsBridge`）、Java 崩溃（`E AndroidRuntime`）、native 崩溃（`fatal error`/`SIGSEGV`）。

### 3. 截图（用户视角 UI）

```powershell
& $adb shell screencap -p /sdcard/s.png
& $adb pull /sdcard/s.png "$env:TEMP\s.png"
& $adb shell rm /sdcard/s.png
```

设备端写文件再 pull，**避免 `adb shell` pty 对 PNG 做 LF→CRLF 损坏**。Read 工具对大 PNG 偶报 `Unsupported Image`，可用 `[System.Drawing]` 转 JPG 再读，或像素采样判断是否黑屏。

### 4. 端口/网络（验证 server 真监听）

```powershell
& $adb shell "cat /proc/net/tcp"   # state 0A = LISTEN，第2列本地地址端口为十六进制
```

预期监听端口：MCP `8765`、pprof `6061`、proto-test TCP `18000` / HTTP `20144`。看到这些端口即对应 server 启动成功；应用能 run 本身也证明 `mcpServer.Start()` 成功（失败会 `return` 不 run，见 [cmd/rain-qa-func/wails.go](../../../cmd/rain-qa-func/wails.go)）。

## CDP JS ASCII 规范（避 GBK 坑）⭐

`cdp_eval.ps1` 的 `-Js`/`-JsFile` 含中文时,Windows PowerShell 5.1 按 GBK 读 BOM-less UTF-8 → 中文乱码 → JS 不匹配(返回 not-found)。脚本头注释明确 "Keep this script ASCII-only"。JS **全 ASCII**:

| 场景 | ❌ 中文 | ✅ ASCII |
|---|---|---|
| 点导航按钮 | `textContent.indexOf("配表测试")` | `document.querySelectorAll('.idea-icon-button')[3].click()` |
| DOM 文本断言 | `indexOf("武将")` | 英文片段 / class / 属性选择器 |
| 顶层变量 | `var a=...`(cdp_eval 包装报 ReferenceError) | IIFE `(function(){...})()` 包裹 |
| 换行字面量 | `"\n"`(触发 PowerShell RemoveItem hook) | `String.fromCharCode(10)` |

menuItems index 顺序见上文导航示例。

## 横向白边/溢出排查（逐层 scrollWidth）

"向左拖出白边"、"内容屏外"、"半个字" = 某 `overflow-x:auto` 元素有子元素溢出。**CDP 逐层测 scrollWidth/clientWidth** 定位溢出源(`documentElement.scrollWidth` 可能被子 overflow 裁剪掩盖):

```javascript
(function(){
  function info(s){var e=document.querySelector(s);return e?{sw:e.scrollWidth,cw:e.clientWidth,canScroll:e.scrollWidth>e.clientWidth+1,ofx:getComputedStyle(e).overflowX}:null;}
  return JSON.stringify({de:info('html'),body:info('body'),layout:info('#layout'),header:info('#layout-header'),footer:info('#layout-footer')});
})()
```

`canScroll:true` 的层 = 溢出源。典型根因:
- `status-bar` `width:100%`+`padding` content-box 溢出 → 加 `box-sizing:border-box`
- `PathConfigInput` 固定 `input-width`×2 > 视口 → `:wrap` + `flex:1`
- 多栏 sider+content 和 > 视口 → sider 折叠/overlay

## Go 业务日志读取（serverLog 面板）

Go `log.Printf` 走前端 serverLog 面板(非 logcat)。CDP 排查 Go 后端(CheckUpdate 等)时,打开面板读:
```javascript
// 打开 serverLog 面板(服务端日志 card 第1个按钮)+ 读 [tag] 行
(function(){
  var logCard = document.querySelectorAll('.setting-card')[3]; // 服务端日志
  var btn = logCard ? logCard.querySelector('.n-button') : null;
  if (btn) btn.click(); // 打开抽屉
  // 读抽屉末尾(Go log 在此,非 logcat)
  setTimeout(function(){}, 1500);
})()
// 再读:var d=document.querySelector('.n-drawer-content'); d.innerText 末尾含 Go [tag] 日志
```

## 关键知识点（排查勿误判）

| 现象 | 真相 |
|---|---|
| logcat 找不到 Go 业务日志 | 不是 bug。`ServerLogService` 用 `os.Pipe` 拦截 stdout/stderr → `Event.Emit("serverLog")` 推**前端日志面板**（[serverlog/service.go](../../../backend/pkg/settings/serverlog/service.go)）。要看业务日志：前端打开日志面板，或 CDP 读 serverLog 事件。 |
| `Cleartext HTTP traffic ... not permitted` | 仅约束 **WebView 内**资源加载，**不影响后端 Go 原生网络栈**（飞书/游戏服/MCP 监听都不受限）。 |
| `/wails/custom.js` asset not found | Wails 框架自身请求，前端无引用，无害。 |
| 版本信息显示 N/A | `VersionService` 调 `git log`，Android 无 git→容错返回 N/A。 |
| 打开 Excel / 客户端导出失败 | `OpenExcel`(xdg-open)/`ExportClientConfig`(cmd bat) 是桌面专属，Android 失败属预期（容错不崩）。 |

## 已知降级（容错、不崩溃，无需"修兼容性"）

`VersionService`(git) → N/A 占位；`OpenExcel`(xdg-open) → 失败；`ExportClientConfig`(cmd bat) → Windows 专属。其余功能（配表检查、proto-test、Wiki 检查、飞书、MCP、AI Agent 等）逻辑跨平台可用，但**需要输入数据**（配表/资源目录/游戏服地址）才能端到端验证。

## 排查"某功能在手机能否用"的决策树

1. 应用能否启动？→ logcat 看崩溃 → 能启动则核心链路通。
2. 该功能页面能否加载/数据能否显示？→ CDP 点击导航读 body 文本 → 能显示则后端 Service 可用。
3. 功能依赖外部输入？→ 确认输入可达（配表目录是否存在、游戏服地址是否连得上）—— 这通常是"不可用"的真因，而非代码。

## 自迭代

每次使用后自问：是否暴露了新的 Android 调试陷阱/端口/命令？是否发现新的"非问题"误判模式？如是，更新本 skill 与 `docs/Android-运行时调试.md`，并按 `.learnings/LEARNINGS.md` 格式记录。
