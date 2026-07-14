# Android 运行时调试与后端可用性

> 将 rain-qa-func(Wails v3) 跑在 Android 模拟器/真机上后，如何调试、后端哪些功能可用。本文沉淀 2026-07-09 在 `Small_Phone`(x86_64) 模拟器上的实证方法与结论，供后续复现与排查。构建流程见 [Android APK 构建](Android-APK构建.md)。

## 一、后端可用性结论（实证）

**核心结论：后端 Go 代码在 Android 上基本全面可用，没有崩溃级或功能阻断问题。** 静态分析时担心的"平台不兼容"大多不成立，真正的降级集中在少数桌面专属/外部二进制依赖，且都容错、不崩溃。

### 实证确认可用（✅）

| 能力 | 证据 | 关联代码 |
|---|---|---|
| 应用启动、无崩溃 | 重启后 pid 稳定，logcat 无 `panic`/`AndroidRuntime`/`SIGSEGV` | [runWails](../cmd/rain-qa-func/wails.go:57) |
| 前后端运行时桥 | logcat `wails:runtime:ready` + 持续 `Runtime call`（422 已修） | 见 [Android APK 构建-已知问题](Android-APK构建.md) |
| 前端完整渲染 | CDP 读到全部 8 个功能模块入口 + 用例面板 | — |
| Service 调用链 + 文件 IO | 首页"1 条用例"→ `TestCaseService` 加载 + cases 文件读取成功 | [wails.go 服务注册](../cmd/rain-qa-func/wails.go:91) |
| MCP HTTP server | `0.0.0.0:8765` 在监听 → `mcpServer.Start()` 成功（失败会 `return`） | [runWails Start](../cmd/rain-qa-func/wails.go:224) |
| proto-test 代理监听 | TCP `18000` + HTTP `20144` 都在监听 → `StartListen` 成功 | [runWails StartListen](../cmd/rain-qa-func/wails.go:173) |
| Go 原生网络栈 | `net.Listen`/HTTP server 在 Android 完全工作（pprof `6061` 也在） | — |
| CGO / 平台特定代码 | 零 CGO；`attachconsole`、`x/sys/windows` 均 `//go:build` 隔离 | [main.go attachParentConsole](../main.go:21) |

### 降级项（⚠️ 容错、不崩溃，返回 error）

| 功能 | Android 表现 | 关联代码 | 处理建议 |
|---|---|---|---|
| 版本信息 `VersionService` | git 不存在→返回 `N/A` 占位（首页"最近更新:..."） | [getGitCommits](../backend/pkg/settings/wails.go:549) | 可接受；或内置版本号替代 |
| 打开 Excel `OpenExcel` | 走 default→`xdg-open` 不存在→失败 | [OpenExcel](../backend/pkg/common/fileutils/open_excel.go:34) | 仅"查看原表"按钮失效，不影响检查 |
| 客户端导出 `ExportClientConfig` | `cmd /c export_client.bat`，Windows 专属→失败 | [ExportClientConfig](../backend/pkg/proto-test/server-config/server_config.go:234) | Android 上隐藏入口 |

### 澄清的"非问题"（实证排除，排查时勿误判）

- **`Cleartext HTTP traffic ... not permitted` 警告**：仅约束 **WebView 内**的资源加载（如 Google 字体），**不影响后端 Go 原生网络栈**（飞书 webhook、游戏服务器、MCP 监听都不受此限制，Go 用自己的 `net` 栈）。
- **`/wails/custom.js` asset not found**：Wails 框架自身请求，前端无引用，无害。
- **Go 业务日志（`[main]`/`InitExcel`/`MCP`）不在 logcat**：不是 bug——`ServerLogService` 用 `os.Pipe` 拦截 stdout/stderr + `log.SetOutput`，经 `Event.Emit("serverLog")` 推到**前端日志面板**（[InitWithApp](../backend/pkg/settings/serverlog/service.go:37)）。`os.Pipe` 跨平台，Android 可用。

### 待真数据才能验证

`CheckAllExcelRules`、`InitExcel` 这类**需要输入数据**的功能逻辑跨平台可用，但模拟器上缺策划配表/资源目录——要端到端验证需先把测试 Excel 推进设备并在设置页配置路径。proto-test 的"代理到真实服务器"同理，需可达的游戏服地址（默认 `10.254.114.204:18000` 是内网，模拟器连不上）。

---

## 二、调试工具链

### 1. adb logcat —— 看启动与崩溃

```powershell
$adb = "D:\Android\Sdk\platform-tools\adb.exe"
# 干净复现启动日志
& $adb logcat -c
& $adb shell am force-stop com.wails.app
& $adb shell am start -n com.wails.app/.MainActivity
Start-Sleep -Seconds 8
& $adb logcat -d | Select-String "Wails|AndroidRuntime|panic|fatal"
```

要点：
- **Wails 自身日志**（`D Wails:`/`WailsBridge`）会进 logcat（经 JNI 桥）。
- **Go 业务代码的 `log.Printf` 不进 logcat**（见上文"非问题"），要看业务日志去前端 serverLog 面板。
- 关注 `E AndroidRuntime`（Java 崩溃）、`fatal error`/`SIGSEGV`（native 崩溃）。

### 2. CDP 远程调试 —— 最强渠道（执行 JS、调后端、读 DOM）

Wails Android 默认开启 WebView 调试，可用 Chrome DevTools Protocol 直接驱动前端。封装好的脚本：[`build/android/scripts/cdp_eval.ps1`](../build/android/scripts/cdp_eval.ps1)，自动完成「查 pid → adb forward → 取 page → 连 WebSocket → 执行 JS」。

```powershell
# 读首页可见文本，判断应用状态
& build/android/scripts/cdp_eval.ps1 -Js "String(document.body.innerText).slice(0,1000)"

# 读 JS 文件（复杂脚本用文件传入，避免命令行转义）
& build/android/scripts/cdp_eval.ps1 -JsFile snippet.js

# 点击导航按钮(示例:配表测试 = menuItems[3])并等待读取（Promise + awaitPromise 自动启用）
# ⚠️ -Js/-JsFile 含中文经 PowerShell 5.1 GBK 读取会乱码 → JS 不匹配。导航用 index、断言用英文,见下方"CDP JS ASCII 规范"
& build/android/scripts/cdp_eval.ps1 -Js 'new Promise(function(r){var b=document.querySelectorAll(".idea-icon-button")[3];if(b)b.click();setTimeout(function(){r(String(document.body.innerText).slice(0,1500))},2500)})'
```

#### CDP JS ASCII 规范（避坑）

`cdp_eval.ps1` 的 `-Js`/`-JsFile` 含中文时,Windows PowerShell 5.1 按 GBK 读取 UTF-8 文件 → 中文乱码 → JS 不匹配(返回 not-found)。脚本头注释明确 "Keep this script ASCII-only"。JS 须全 ASCII:

| 场景 | ❌ 中文写法 | ✅ ASCII 写法 |
|---|---|---|
| 点导航按钮 | `textContent.indexOf("配表测试")` | `document.querySelectorAll('.idea-icon-button')[3].click()`(menuItems 顺序:0设置/1AI助手/2战斗测试/3配表测试/4武将资源检查/5武将Wiki检查/6活动Wiki/7Proto测试) |
| DOM 文本断言 | `indexOf("武将")` | 英文片段 / class / 属性选择器 |
| 顶层变量 | `var a = ...`(cdp_eval 包装报 ReferenceError) | IIFE `(function(){...})()` 包裹 |
| 换行字面量 | `"\n"`(触发 PowerShell RemoveItem hook) | `String.fromCharCode(10)` 或避免 |

能力：读 DOM、点击导航/按钮触发后端调用、检查 `window.wails`、验证 Service 返回。手动等价操作：

```powershell
$adb forward tcp:9223 localabstract:webview_devtools_remote_<pid>
# 浏览器访问 http://127.0.0.1:9223/json 取 page webSocketDebuggerUrl，或直接 chrome://inspect
```

### 3. 截图 —— 看用户视角 UI

```powershell
$adb = "D:\Android\Sdk\platform-tools\adb.exe"
# 设备端写文件再 pull（避免 adb shell pty 对 PNG 的 LF→CRLF 损坏）
& $adb shell screencap -p /sdcard/s.png
& $adb pull /sdcard/s.png "$env:TEMP\shot.png"
& $adb shell rm /sdcard/s.png
```

注意：Read 工具对大尺寸 PNG 偶报 `Unsupported Image`，可用 `System.Drawing` 转 JPG 后再读；或用像素采样判断是否黑屏（见本仓库历史操作）。

### 4. 端口/网络验证 —— 确认 server 是否真监听

```powershell
$adb = "D:\Android\Sdk\platform-tools\adb.exe"
& $adb shell "cat /proc/net/tcp"   # state 0A = LISTEN，第2列本地地址端口为十六进制
```

实测监听端口对照：MCP `8765`、pprof `6061`、proto-test TCP `18000`/HTTP `20144`。看到这些端口即对应 server 启动成功。

### 5. Go 后端业务日志渠道

业务 `log.Printf` 走前端 serverLog 面板（事件名 `serverLog`，前端 `use-server-logs.ts` 监听）。要在 logcat 看后端日志，需额外桥接（待办，见下）。

---

## 三、待办/可改进

- **Go 业务日志桥到 logcat**：当前只前端可见，手机调试不便。可在 Android 入口把 `log`/`os.Stdout` 重定向到 Android log（需 cgo + `android/log`，或写文件后 `adb pull`）。
- **真机 arm64 端到端**：本文实证基于 x86_64 模拟器；arm64 真机行为应一致（同份 Go 代码 + libwails.so），但建议真机回归一次。

## 四、相关文档

- [Android APK 构建](Android-APK构建.md) — 构建/安装/启动流程、踩坑、版本对齐
- [Wails 开发注意事项](Wails开发注意事项.md)
