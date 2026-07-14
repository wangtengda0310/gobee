---
CLAUDE.md中会使用一些用户自定义的标签供你参考

<!-- pinned --> 意思是这些内容用户认为比较重要，你再进行文档重构时避免将他们提取到单独的文件导致ai agent无法第一时间直接加载到
<!-- layered --> 意思是这些内容推荐使用渐进式披露进行组织，当前可能内容不多直接内联到了主文档，后期内容膨胀后需要按照渐进式披露的组织方式进行优化
<!-- warnning --> 意思是这不分内容需要提醒用户，选择在[开启会话、完成一个工作]中的任意时机想用户发出警告
<!-- todo --> 意思是需要你重点关注的内容，一般为用户发现文档不符合预期，需要你进行迭代
<!-- task --> 用户列出的代办任务，需要今后实现。触发时机：每次完成一个工作单元时（**git 提交后必触发**，其他收尾场景也应触发）。扫描范围：从 worktree 根目录到当前工作文件夹的所有 CLAUDE.md，以及根目录 docs/TODO.md。找到后向用户列出待办供选择下一步任务

---
# rain-qa-func 项目说明

当前项目为名将杀的 QA 测试桌面应用，基于 Wails v3 框架开发，提供测试用例配置和执行功能。支持GUI（windows）、CLI（windows/macOs）、程序内嵌AI Agent、MCP接口。名将杀是类三国杀的游戏，在三国杀的基础上进行了扩展和改良。

## first of all <!-- pinned -->
- 抑制你使用Read/Grep/Search等工具的冲动。
  1. 你应总是优先使用 `claude-mem`、`forgetful`、`mempalace`、可用的技能、mcp等工具查询历史记录和相关信息，避免重复劳动和信息孤岛。
  2. 在你安排Search/Explorer等subagent工作时也应当这么要求它们。
  3. 在找到你所需要的信息后使用所有这些工具同步你所获得的信息。

## 系统干净约束（重要）<!-- pinned -->
应用长期使用**不留垃圾历史数据**。实现任何功能时必须考虑数据生命周期,避免累积:
- **临时/缓存有上限或定期清理**:下载文件安装后清理、示例数据可重置(先清旧再放新)、日志/历史缓存有 max 上限(如 serverLog maxHistory=500)、cache 目录不无限增长
- **配置/数据变更不残留**:路径切换、重置、删除操作要彻底清理旧数据(文件 + 配置 + 缓存)
- **Android 端尤其注意**:私有目录(files/cache)、自动更新下载的 APK(安装后删)、go:embed 释放的示例数据(可重置)、wails-picker 复制的临时文件 等,需有清理机制
- **禁止私有目录外文件操作**:所有文件读写必须在应用私有目录(files/cache,Android 即 `/data/data/<pkg>/`)内,禁止写 `/sdcard` 等外部存储(卸载残留,破坏卸载干净)。需用户数据导入时,导入到私有目录(C/A/B 方案),不用 D(MANAGE_EXTERNAL_STORAGE)。详见 [Android 系统干净](docs/Android-系统干净.md)
- **判定标准**:同一操作重复执行 N 次,存储占用不应单调增长;卸载/重置后无残留

## 各关键时间节点你需要注意的事情

### 调研的时候
1. 你需要随时将你获取到的信信息记录下来，方便后续调用subagent的时候共享这些信息，避免重复劳动和信息孤岛。
2. 某些技能或文档中可能要求你隔离subagent的上下文并禁止subagent继承你的记忆和历史记录，这时候你需要尊巡这些要求，但你仍需给出提示让agent知道如何在需要时候正确地找到这些信息尤其是你已经花费大量精力调研的结果。

todo 补充更详细的指导手册

### 开始新工作的时候
1. 想用户罗列不使用Read工具的情况下你所看到的skills和各CLAUDE.md文档中指明此时你需要注意的事项，询问用户是否需要你使用、尊巡这些技能和事项。

### 调整 golang 代码目录的时候
优先使用 `go-pkg-refactor` 技能。简要流程：

1. `git mv` 移动文件
2. `goimports` 修复 import（简单移动通常足够；若涉及 blank import 或同名符号冲突，改用 `gofiximports` 显式替换路径）
3. `gofmt -w` 格式化
4. `go build ./...` 验证编译
5. 同步受影响的 CLAUDE.md 文档
6. git 提交

### 调用工具的时候

如果你是kimi agent:

1. **调用 Bash/PowerShell 前必须先明确具体命令**
   - `command` 字段必须非空，禁止传空字符串
   - 调用前能明确说出该命令的预期输出
   - 不要把空调用当作"占位"或"稍后填充"

2. **禁止空命令调用**
   - 不允许 `{"command": ""}` 这种调用
   - 如果还没想好具体命令，用文字说明状态，不要调用 shell

3. **调用后验证输出**
   - 如果预期有输出但返回 `(Bash completed with no output)`，立即检查是否传了空命令
   - 如果命令失败，先分析错误原因再决定下一步，不要盲目重试

4. **优先使用非 shell 工具做文件操作**
   - 读文件用 `Read`，写文件用 `Write`，改文件用 `Edit`
   - 只在需要运行程序、git 操作、构建测试时使用 Bash/PowerShell

5. **跨目录操作时使用绝对路径**
   - 当前 worktree 路径可能不是目标路径，避免依赖相对路径解析
   - 文件操作优先使用完整绝对路径

### 你认为任务已经完成的时候
1. 尊巡各CLAUDE.md中标明的需要你此时完成的事项。
2. 使用`/渐进式披露`技能检查文档是否需要优化。
3. **触发时机**：git 提交后必触发；其他工作收尾场景也应触发。扫描范围：从 worktree 根目录到当前工作文件夹路径上的**所有 CLAUDE.md**，以及根目录 `docs/TODO.md`。找到 `<!-- task -->` 标记的待办或 TODO.md 列表时，必须向用户列出待办项供选择下一步。
4. 回顾任务过程中遇到的卡点，复盘有价值的经验教训，分析是否有可提取为技能的重复性工作或特殊工作。

### 安排subagent工作的时候
1. 你需要在安排subagent工作的时候提醒它们也要遵守以上的规范，优先使用claude-mem、forgetful、mempalace、可用的技能、mcp等工具查询历史记录和相关信息，避免重复劳动和信息孤岛。

## 技术栈

- 后端: Go 1.25 + Wails v3
- 前端: Vue 3 + Vite + TypeScript + Naive UI

开发过程中使用 `wails3 dev` 运行程序查看效果，如已运行则不需要重复执行。
git 提交前使用 `wails3 task build` 构建可执行程序，产物位于项目根目录一并提交。

## 分支策略

| 分支         | 用途   | 操作规则                               |
|------------|------|------------------------------------|
| **dev**    | 集成分支 | 汇总各 worktree 的开发成果，用户审核后手动推送到 main |
| **main**   | 稳定发布 | 仅由用户手动推送，禁止直接在上面开发和提交              |
| **master** | 废弃分支 | 禁止操作，任何情况下都不要使用                    |

**开发流程**（用户通过 `claude -w` 创建 worktree，不需要 AI 创建）：

1. **开发** → 在 worktree 中开发和提交（可能多次提交），注意不要修改到了主目录的文件文件。
2. 在任何验收通过的场景下提交保存工作进度。
2. **squash** → 将相关的零碎提交合并为一个大提交，在 worktree 上 squash 多个 commit 为一个 使用herodoc配合rebase -i，禁止使用 `git reset` (会被运行中的rain-qa-func.exe 锁定儿导致操作失败) <!-- todo: PowerShell 不支持 heredoc 语法-->
3. **rebase onto main** → `git fetch origin && git rebase origin/main`
4. **同步到 dev** → `git push . <worktree-branch>:dev`（不切换分支）
5. **用户审核** → 用户审核 dev 上的改动，手动推送到 main

git提交、推送、合并、rebase等操作需要遵守 [Git 操作安全规范](docs/git-safe.md)。

## 依赖项目
rain-robot（外部）
`../rain-robot/`（这个项目module不符合规范不能直接import，所以使用`rain-qa-func`分支单独为此项目维护了一个改过module名的版本）


## 目录结构

```
rain-qa-func/
├── main.go                  # 应用入口（调用 cmd.Execute）
├── assets.go                # 前端资源嵌入（go:embed 必须在根目录）
├── cmd/rain-qa-func/        # Cobra CLI 命令定义
│   ├── root.go              # 根命令 + 子命令注册
│   └── wails.go             # Wails GUI 启动逻辑
├── cmd/rain-qa-func/        # 纯CLI不依赖wails，用于proto-test技能，需要适配windows,macOS
├── frontend/                # 前端代码
├── cases/                   # 测试用例存储
├── docker/                   # 流水线任务用到的docker镜像级脚本
├── build/                   # 构建配置
├── backend/            # 内部 pkg 包（原 rain-qa-func/）
│   └── pkg/
│       ├── proto-test/           # 协议录制/重放
│       ├── function-test/         # 战斗测试
│       ├── excel-test/            # 配表测试
│       ├── hero-wiki-check/       # 武将 Wiki
│       ├── activity-wiki-check/   # 活动 Wiki
│       ├── proto-test/      # Proto 协议测试（cobra.go + cobra-help.md）
│       ├── settings/mcp/    # MCP 服务器（cobra.go + cobra-help.md）
│       ├── rain-resources-checker/# 资源检查，当前功能较单一，后续可能扩展为多个检查模块
│       ├── settings/              # 设置/首页
│       ├── common/                # 通用功能
│       └── table_relations/       # 表关系解析
└── rain-qa-func.exe         # 构建产物，需要提交git，只针对windows平台使用
```

## 关键入口

| 功能           | 文件位置                            |
|--------------|---------------------------------|
| Cobra 根命令    | cmd/rain-qa-func/root.go        |
| Wails GUI 启动 | cmd/rain-qa-func/wails.go       |
| 前端资源嵌入       | assets.go                       |
| 应用入口         | main.go:34                      |
| 服务注册         | main.go:48-56                   |
| 功能测试服务       | internal/functionCaseService.go |
| Excel检查服务    | internal/excelCheckService.go   |

## Cobra CLI 子命令规范 <!-- layered -->

应用通过 cobra 组织命令行接口。无子命令时启动 Wails GUI，指定子命令时执行对应的 CLI 功能。

### 目录组织规则

1. **一级子命令** — 在 `cmd/rain-qa-func/` 目录下定义注册逻辑（root.go 中 AddCommand）
2. **二级及以下子命令** — 在模块目录的 `cobra.go` 中定义，与 `wails.go`/`mcp.go` 同级
3. **帮助文档** — 在 `cobra.go` 同级创建 `cobra-help.md`，通过 `//go:embed cobra-help.md` 嵌入为 `--help` 的 Long 描述

### 文件职责

| 文件 | 职责 |
|------|------|
| `cmd/rain-qa-func/root.go` | RootCmd 定义 + 一级子命令注册 + Execute() 入口 |
| `cmd/rain-qa-func/wails.go` | Wails GUI 完整启动逻辑（RootCmd.Run 调用） |
| `backend/pkg/<module>/cobra.go` | 模块暴露 `New<Module>Cmd()` 返回 `*cobra.Command` |
| `backend/pkg/<module>/cobra-help.md` | 模块的 --help 文档内容 |

### 新增子命令流程

1. 在目标模块目录创建 `cobra.go`，实现 `NewXXXCmd() *cobra.Command`
2. 同目录创建 `cobra-help.md`，通过 `//go:embed cobra-help.md` 嵌入
3. 在 `cmd/rain-qa-func/root.go` 的 `init()` 中 `RootCmd.AddCommand(xxx.NewXXXCmd())`
4. 如果是已有模块的二级子命令，在模块 `cobra.go` 的 `NewXXXCmd()` 内部 `AddCommand`

### 前端资源嵌入

`//go:embed all:frontend/dist` 必须在根目录定义（go:embed 不支持 `../` 路径）。
`assets.go` 通过 `cmd.SetAssets()` 传递给 cmd 包，由 `wails.go` 使用。

### 现有子命令

| 子命令 | 模块目录 | 状态 |
|--------|----------|------|
| `proto-test` | `backend/pkg/proto-test/` | 仅 --help |
| `mcp` | `backend/pkg/settings/mcp/` | 仅 --help |

## 子模块文档

| 模块                 | 说明         | 文档                                                                      |
|--------------------|------------|-------------------------------------------------------------------------|
| proto-test          | 协议录制/重放/拦截 | [backend/pkg/proto-test/CLAUDE.md](backend/pkg/proto-test/CLAUDE.md)     |
| function-test      | 战斗测试       | [backend/pkg/function-test/CLAUDE.md](backend/pkg/function-test/CLAUDE.md) |
| excel-test         | 配表测试       | [backend/pkg/excel-test/CLAUDE.md](backend/pkg/excel-test/CLAUDE.md)       |
| proto-test/server-config | Unity 服务器注入 + 客户端配置导出（proto-test 子包） | [backend/pkg/proto-test/server-config/CLAUDE.md](backend/pkg/proto-test/server-config/CLAUDE.md) |
| frontend           | 前端代码组织     | [frontend/src/CLAUDE.md](frontend/src/CLAUDE.md)                        |
| rain-excel-checker | 策划配表检查     | [rain-excel-checker/CLAUDE.md](rain-excel-checker/CLAUDE.md)            |
| cmd/tests/streamproxy    | CLI 流量代理工具 | [cmd/tests/streamproxy/CLAUDE.md](cmd/tests/streamproxy/CLAUDE.md)                  |

## 前端代码组织

> 详见 [frontend/src/CLAUDE.md](frontend/src/CLAUDE.md)

代码按功能内聚组织，每个页面的组件、状态、配置自包含在 `pages/` 子目录下。

## 文档组织规范

本项目采用前后端双向索引的文档组织方式：

1. **后端代码目录中的 CLAUDE.md** — 必须包含：前端页面关键元素索引、时序图
2. **前端页面代码目录中的 CLAUDE.md** — 必须包含：后端关键代码索引、布局文档引用、时序图
3. **前端布局文档** — 必须包含：E2E 测试用例引用、时序图
4. **父文件夹中的 CLAUDE.md 规范子文件夹、模块中的 CLAUDE.md 组织结构**

**示例**（proto-test）：

| 文档位置                                               | 内容                           |
|----------------------------------------------------|------------------------------|
| `pkg/proto-test/CLAUDE.md`                        | 前端页面关键元素索引 + 时序图（录制/重放流程）    |
| `frontend/src/pages/proto-test/CLAUDE.md`        | 后端关键代码索引 + 布局文档引用 + 时序图      |
| `frontend/docs/layout/pages/proto-test/index.md` | ASCII 布局图 + E2E 测试用例索引 + 时序图 |

## 更多文档

### 开发规范
- [Git 操作安全规范](docs/git-safe.md) — 仓库上下文验证、worktree 防误操作
- [Wails 开发注意事项](docs/Wails开发注意事项.md) — Type Alias、Event.Emit、bindings rename 锁（Windows）、GUI/CLI 构建入口
- [Android APK 构建](docs/Android-APK构建.md) — 桌面应用构建 Android APK 的流程、必装补丁与踩坑（wails 版本对齐、linux_cgo 守卫、Windows NDK .cmd CC、go-task shell）
- [Android 运行时调试](docs/Android-运行时调试.md) — 模拟器/真机调试工具链（logcat/CDP/截图/端口扫描）与后端可用性实证结论（后端基本可用，含 cdp_eval.ps1）
- [Android 前端适配](docs/Android-前端适配.md) — 4 核心页 CDP 体检实测 + 适配全景（数据供给是真难点，非代码兼容）+ 候选方案
- [Android APK 自动更新](docs/Android-自动更新.md) — Wails updater 桌面专用(不适用 Android) + 非 Play 自升级标准方案 + 本项目实施方案(待定更新源)
- [Android 适配复盘](docs/Android-适配复盘.md) — 任务复盘(踩坑根因/通用教训/方法论/后续优化/技能提取建议)
- [Android 数据加载](docs/Android-数据加载.md) — 数据供给瓶颈方案(C 内置示例/A zip 导入/B 选目录/D 外部存储)+ OpenFileDialog 选文件工作(纠正旧误判)+ game resources 待决策
- 记忆系统使用规范 — claude-mem/mempalace/forgetful 使用方式
- [前端调试方法](docs/前端调试方法.md) — DevTools/远程调试/Playwright

### 后端文档
- [MCP 接口使用手册](docs/MCP-USAGE.md) — MCP 工具完整使用指南
- [战斗测试流程实现](docs/战斗测试流程实现.md) — 战斗测试架构、数据结构、执行流程
- [表级规则变更检测-实现详解](rain-excel-checker/docs/表级规则变更检测-实现详解.md)
- [飞书机器人消息发送机制](docs/发送飞书机器人消息.md)
- [ErrCells 与 Ok 字段关系分析](docs/ErrCells与Ok字段关系分析.md) — 表级检查结果数据流与显示逻辑
- [责任人归一逻辑](docs/责任人归一逻辑.md) — 树形结构右键菜单同步负责人功能

### 前端文档
- [前端布局文档](frontend/docs/layout/CLAUDE.md) — 各页面 ASCII 布局可视化
- [截图注释文档](frontend/docs/screenshots/CLAUDE.md) — 页面截图 UI 元素标注

### 功能实现详解
- [武将Wiki检查-实现详解](docs/武将Wiki检查-实现详解.md)
- [武将保护期检查框架设计](rain-excel-checker/docs/武将保护期检查框架设计.md)
- [WebTorrent P2P 文件传输文档](docs/WebTorrent-实现详解.md)

### 待完成任务 <!-- task -->
- [TODO.md](docs/TODO.md)
