# Wails v3 开发注意事项

rain-qa-func 基于 Wails v3 框架开发过程中的关键陷阱和解决方案。

## 禁止使用 Type Alias

**问题**：Wails v3 的 bindings 生成器不支持 type alias，使用 type alias 会导致前端无法正确调用后端服务。

**错误示例**：
```go
// ❌ 错误：使用 type alias
type FeishuNotifyConfig = feishuconfig.FeishuNotifyConfig

type MyService struct {}
func (s *MyService) GetConfig() *FeishuNotifyConfig { ... }
```

**正确示例**：
```go
// ✅ 正确：直接定义完整的结构体
type FeishuNotifyConfig struct {
    Enabled         bool   `json:"enabled"`
    RobotGUID       string `json:"robotGuid"`
    MessageTemplate string `json:"messageTemplate"`
}

type MyService struct {}
func (s *MyService) GetConfig() *FeishuNotifyConfig { ... }
```

**原因**：
- Wails v3 的 bindings 生成器通过解析 Go 类型定义来生成 TypeScript 接口
- Type alias 会被解析为别名指向的类型，导致前端导入路径错误
- 错误信息类似：`Cannot find module "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/wails/xxx'`

**解决方案**：
1. 如需跨包共享类型，将类型定义放在 `internal/common/` 目录下
2. 各包直接引用该类型，而不是使用 type alias
3. 如确实需要类型兼容，使用接口（interface）而非 type alias

## wails3 dev bindings rename 失败（Windows）

**现象**：`wails3 dev` 启动时会内部执行 `generate bindings -clean=true`。生成器先把 TypeScript 写到 `frontend/.bindings-tmp-*`，删除旧的 `frontend/bindings/`，再把临时目录 rename 为 `bindings`。Windows 上若目标目录或其中文件仍被占用，rename 会失败（`Access is denied` / `rename ... bindings: Access is denied`）。此时旧 bindings 可能已被删掉，只剩 `.bindings-tmp-*`，前端 `@bindings` 导入大面积报错。

**常见占用来源**：

| 来源 | 说明 |
|------|------|
| `wails3 dev` / `rain-qa-func.exe` | Vite、WebView2、TS 语言服务持有 `frontend/bindings/` 句柄 |
| dev 运行时手动 `generate bindings` | 与 dev 内部生成争抢同一目录，极易删光 bindings 后 rename 失败 |
| IDE 索引 | Cursor/VS Code 正在索引 bindings 目录 |
| 多终端并行 generate | 多个进程同时 `-clean=true` |

**恢复与正确流程**：

```bash
# 1. 先停占用进程（PowerShell）
powershell.exe -Command "Stop-Process -Name 'wails3','rain-qa-func' -Force -ErrorAction SilentlyContinue"

# 2. 若 bindings 已损坏或大量缺失，从 git 恢复
git restore frontend/bindings/

# 3. 单独生成（dev 未运行时）
wails3 generate bindings -ts

# 4. 再启动开发
wails3 dev
```

**generate 仍失败时**：

- 用 `-clean=false` 写到备用目录，确认内容后再替换：`wails3 generate bindings -ts -d frontend/bindings-gen -clean=false`
- 检查是否仍有进程占用：`tasklist | findstr /i "wails3 rain-qa-func"`
- 清理残留的 `frontend/.bindings-tmp-*`（仅在已 `git restore` 且确认无有用内容后）

**预防**：

- 修改 Go 暴露给前端的 API 后：**先停 dev，再 `generate bindings`，再启 dev**；不要在 `wails3 dev` 运行时执行 generate（含 AI 子任务）。
- `frontend/bindings/` 已纳入版本控制；日常拉取/合并后若与 Go 代码一致，不一定每次 dev 都需要成功执行 `-clean`。
- merge/rebase 时 bindings 冲突的处理见根 `CLAUDE.md`「Bindings 冲突处理」。

**相关文档**：`frontend/e2e/CLAUDE.md` §3（E2E 场景下的同一问题与变通启动方式）。

## Event.Emit 序列化限制

`app.Event.Emit(name, data)` 中 `data` **不能使用自定义 Go 结构体**，必须用 `map[string]any`。

**现象**：静默失败，无编译时或运行时错误。只有通过 CDP 拦截才能发现。

| 数据类型 | 是否可用 | 示例 |
|----------|----------|------|
| `string` | 可用 | `Emit("time", now)` |
| `map[string]any` | 可用 | `Emit("chatToolCall", map[string]any{...})` |
| `nil` | 可用 | `Emit("chatStreamDone", nil)` |
| 多参数 | 可用 | `Emit("robotLog", log, seq)` |
| **自定义 struct** | **不可用** | `Emit("serverLog", LogEntry{...})` → 静默失败 |

**修复模式**：为结构体添加 `ToMap() map[string]any` 方法。

## 前端 Events.On 回调接收 WailsEvent 包装

`Events.On(eventName, callback)` 的回调接收的是 `WailsEvent` 对象，**不是**后端 `Event.Emit(name, data)` 中直接传入的 `data`。

### 数据结构

```typescript
// callback 实际接收的参数
{
  name: "record:progress",   // 事件名
  data: { ... },             // 后端传入的实际数据（在这里）
  sender: "..."              // 发送方窗口名（可选）
}
```

### 正确的访问方式

```typescript
// ✅ 正确：解包 WailsEvent
Events.On('record:progress', (raw: any) => {
  const data = raw.data ?? raw   // raw.data 是后端传入的 map[string]any
  console.log(data.status)       // 正确访问业务字段
})

// ❌ 错误：直接当原始数据用
Events.On('record:progress', (data: any) => {
  console.log(data.status)       // undefined！data 是 {name, data} 对象
})
```

### 版本信息

此行为在 `@wailsio/runtime@3.0.0-alpha.78` 及 Go module `wails/v3@v3.0.0-alpha.96` 中确认。`dispatchWailsEvent` 函数会创建 `new WailsEvent(event.name, event.data)` 包装，Listener.dispatch 将整个 WailsEvent 传给回调。

## GUI / CLI 构建入口

项目提供两个独立的 Go 入口：

```text
main.go
└── cmd/rain-qa-func/
    ├── root.go   # cobra RootCmd，无子命令时启动 Wails GUI
    └── wails.go  # Wails GUI 启动逻辑 + Service 注册

cmd/rain-qa-func-cli/
└── main.go       # 纯 CLI 入口，不依赖 Wails
```

- `main.go` 构建出的 `rain-qa-func.exe` **永远是 GUI 二进制**（Windows 平台使用）。
- `cmd/rain-qa-func-cli/main.go` 构建出的是 **纯 CLI 二进制**，用于 macOS 以及任何不需要 GUI 的场景。

> **相关文档**：proto-test CLI 的具体用法及重新编译命令见 [`.claude/skills/proto-test-cli/SKILL.md`](../../.claude/skills/proto-test-cli/SKILL.md)；其预编译二进制正是基于 `cmd/rain-qa-func-cli/` 构建的纯 CLI 版本。

### 平台差异

Wails v3 alpha.96 的 `pkg/application` 中：

- **Windows**：没有 `import "C"`，只有 `//go:build windows` 约束，**不依赖 cgo**。
- **Linux / Darwin / Android**：大量使用 `import "C"`，**必须开启 cgo**。

因此：

| 平台 | 入口 | 构建要求 |
|------|------|---------|
| Windows | `main.go` | `CGO_ENABLED` 可以是 0 |
| Linux / Darwin | `main.go` | 必须 `CGO_ENABLED=1` |
| 任何平台 | `cmd/rain-qa-func-cli/main.go` | 不需要 cgo |

### 故障恢复

若 `wails3 dev` 出现 `Cannot find module '@bindings/...'` 类的 TypeScript 错误，通常是因为 `frontend/bindings/` 被意外清空或损坏。按以下步骤恢复：

```bash
# 1. 停止占用进程
powershell.exe -Command "Stop-Process -Name 'wails3','rain-qa-func' -Force -ErrorAction SilentlyContinue"

# 2. 从 git 恢复 bindings
git restore frontend/bindings/

# 3. 重新生成
wails3 generate bindings -ts

# 4. 再启动开发
wails3 dev
```

**预防**：修改 Go 暴露给前端的 API 后，**先停 dev，再 `generate bindings`，再启 dev**；不要在 `wails3 dev` 运行时执行 generate。

## 服务端日志架构

自动捕获 Go 后端 stdout/stderr 输出，通过 Wails 事件实时推送到前端。

```
数据流: os.Stdout/os.Stderr → pipe 写端 → pipe 读端 → goroutine → app.Event.Emit("serverLog", entry.ToMap()) → 前端
```

| 文件 | 职责 |
|------|------|
| `internal/serverlog/types.go` | LogEntry 定义 + ToMap() |
| `internal/serverlog/writer.go` | PipeWriter: os.Pipe 拦截 stdout |
| `internal/serverlog/service.go` | Wails 服务 + 手动日志方法 (Debug/Info/Warn/Error) |

**初始化顺序** (main.go):
1. `NewServerLogService()` — 在 app 创建前
2. `InitWithApp(app)` — app 创建后：替换 stdout/stderr + `log.SetOutput(pipeFile)`
3. `RegisterService` — 注册 Wails 服务

**关键知识点**：Go `log` 包在 init 时捕获 `os.Stderr` 引用，必须显式调用 `log.SetOutput()` 重定向。

## 双击 .exe 启动 GUI（cobra Mousetrap + PE subsystem）

**问题**：双击 `rain-qa-func.exe` 提示 "This is a command line tool. You need to open cmd.exe and run it from there."，GUI 不启动（但 `cmd /c rain-qa-func.exe` 正常）。

**根因**：两个独立机制共同导致，必须分别处理：

| 机制 | 控制什么 | 设置 |
|------|---------|------|
| **PE subsystem**（build 时）| 双击是否开 cmd 窗口 | `-H windowsgui`（GUI=2 无 cmd）/ 默认 console(3 开 cmd) |
| **cobra Mousetrap**（运行时）| 是否阻塞 runWails（GUI 启动）| `MousetrapHelpText`（非空=检测 explorer.exe 双击并阻塞 RootCmd.Run）|

**两者独立**：
- 单改 subsystem 不能让双击启动 GUI（Mousetrap 仍阻塞 runWails，只是无 cmd 窗口看不到提示）
- 单禁用 Mousetrap 双击会开 cmd 显示日志（console subsystem）

**修复**（commit 1d2b898，两者都做）：
1. `cmd/rain-qa-func/root.go` init 设 `cobra.MousetrapHelpText = ""`（禁用 Mousetrap，双击 → runWails 执行）
2. `build/windows/Taskfile.yml` PRODUCTION=false 分支加 `-ldflags="-H windowsgui"`（GUI subsystem，双击无 cmd 窗口）

**原理细节**：
- cobra `command_win.go`：`if MousetrapHelpText != "" && mousetrap.StartedByExplorer() { c.Print(MousetrapHelpText) }`（阻塞 RootCmd.Run = runWails）
- `StartedByExplorer()` 检测父进程 explorer.exe（双击），**与 subsystem 无关**（进程树检测）
- wails3 task build 默认 PRODUCTION=false 无 `-H windowsgui` → console subsystem

**残留**：GUI subsystem .exe 双击时 WebView2 启动有 ~2 个 console 快速闪动（wails3/WebView2 启动副作用，非项目代码，不影响 GUI 使用）。

## GUI 子系统下 CLI 输出不可见（AttachConsole 修复）

**问题**：`rain-qa-func.exe` 命令行带参数运行（如 `--help`、`proto-test` 子命令）时，在 cmd/PowerShell 里「没反应」——不输出、秒退、无报错。但程序逻辑正常（用管道捕获可拿到完整 help 文本）。

**根因**：`rain-qa-func.exe` 用 `-H windowsgui` 构建（PE Subsystem=GUI，见上节）。Windows 加载器启动 GUI 子系统程序时**不附加父进程的控制台**，进程的 stdout/stderr 句柄无效（`GetStdHandle` 返回 NULL），cobra 写到 `os.Stdout` 的内容全部丢失。Linux/macOS 的 ELF/Mach-O 没有 PE 子系统概念，进程默认继承父进程 fd，不存在此问题。

**修复**（[attachconsole_windows.go](../attachconsole_windows.go)，build tag `windows`；非 Windows 空实现 [attachconsole_other.go](../attachconsole_other.go)；由 [main.go](../main.go) 在 `cmd.Execute()` 前调用）：三件套针对 GUI 子系统 + cmd/PowerShell 的不同副作用分层处理。

| 函数 | 解决的问题 | 机制 |
|------|----------|------|
| `attachParentConsole` | GUI 子系统 stdout 丢失，CLI 输出不可见 | `AttachConsole(ATTACH_PARENT_PROCESS)` + `CreateFile("CONOUT$")` + `SetStdHandle` + 重建 `os.Stdout/os.Stderr` |
| （有参数才 attach）| AttachConsole 干扰 Wails/WebView2，GUI 闪退 | 函数开头 `if len(os.Args) <= 1 { return }`，GUI 模式（无参数）不 attach |
| `releaseParentConsole` | cmd.exe 因控制台被子进程持有而等待，提示符不刷新 | 退出前 `FreeConsole()` 释放 attachment（`main.go` defer） |
| `nudgeConsoleRefresh` | PowerShell 的 PSReadLine 提示符不重绘（屏幕状态模型与实际不同步） | `FreeConsole` 前用 `WriteConsoleInputW` 写一个假回车键事件，PSReadLine 收到后重绘提示符 |

**关键 Windows 控制台陷阱**：

- **AttachConsole 成功后必须重建 `os.Stdout`**：attach 不会自动更新进程的标准句柄，`os.Stdout` 仍指向启动时的无效 fd。必须 `CreateFile("CONOUT$")` + `SetStdHandle(STD_OUTPUT_HANDLE, h)` + `os.Stdout = os.NewFile(...)`，否则 cobra 仍写到无效 fd。
- **PowerShell 与 cmd 的差异**：cmd.exe 用简单提示符刷新，CONOUT$ 直写不影响它。PowerShell 的 PSReadLine 靠**跟踪 PS 自己经 stdout 管道写的输出**维护屏幕状态；GUI 子系统 exe 写 CONOUT$ 绕过管道，PSReadLine 看不到，exe 退出后不重绘提示符——用户看到光标闪烁误以为程序卡住，实际 PS 已在等下一条命令（此时输入会被 PS 当命令执行）。这是 AttachConsole 方案在 PowerShell 下的固有限制，`nudgeConsoleRefresh` 的假回车是已验证的 workaround。
- **`golang.org/x/sys/windows` 未封装** `AttachConsole`/`WriteConsoleInput`，需通过 `kernel32.dll` 的 `LazyProc` 调用；`ATTACH_PARENT_PROCESS = ^uint32(0)` 需手动定义。

**验证方法**：GUI 子系统 exe 的 stdout 行为只能用「强制传管道句柄」的方式编程式验证——.NET `Process.Start(RedirectStandardOutput=true)` 或 Go `exec.Command` 的 StdoutPipe。修复前管道能捕获到 help（说明程序正常、仅控制台看不到），修复后管道捕获变 0 字节（说明 attach 把 stdout 改向到 CONOUT$、接管了输出）。最终裁判必须在真实交互式 cmd/PowerShell 窗口里做。

### 纯 CLI 入口（cmd/rain-qa-func-cli）

Windows PowerShell 下的提示符问题虽由 `nudgeConsoleRefresh` 缓解，但 AttachConsole 方案的本质决定了它依赖 workaround。项目另提供一个**纯 CLI 入口** [cmd/rain-qa-func-cli/main.go](../cmd/rain-qa-func-cli/main.go)（console 子系统，`CGO_ENABLED=0`，无 Wails 依赖），子命令集与主 exe CLI 模式一致（proto-test、mcp）。它的用途：

- **macOS/Linux 交叉编译**：提供与主 exe 一致的 cobra 子命令（`main.go` 的 GUI 入口在非 Windows 平台需要 cgo）。
- **proto-test-cli skill 预编译二进制的来源**：见 [.claude/skills/proto-test-cli/SKILL.md](../../.claude/skills/proto-test-cli/SKILL.md)。

console 子系统 exe 的 stdout 走管道，PowerShell 的 PSReadLine 能正确跟踪、提示符渲染正常——这是 AttachConsole 方案无法企及的干净路径。
