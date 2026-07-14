---
name: golangci-lint
description: 使用项目内置的 golangci-lint v2 对 Go 代码执行静态分析。当用户提到 lint、golangci-lint、静态分析、代码检查、代码质量、代码规范、跑一下lint、检查代码质量时必须使用此技能。在以下场景也应主动触发：Go 代码修改完成后的质量自检、提交前的代码检查、code review 前的预检、CLAUDE.md 中"任务完成自我检查"要求的 lint 步骤。即使用户只是说"检查一下"而没有明确说 lint，只要当前工作是 Go 代码，也应使用此技能。注意：仅在处理 Go (.go) 代码时触发，Python/TypeScript/前端代码不适用。
---

# golangci-lint 代码静态分析

## 二进制文件

```
.claude/skills/golangci-lint/bin/golangci-lint.exe
```

Windows x86_64，v2.12.2，go1.26.2 构建。项目无需额外安装 golangci-lint。

## 路径定位

工作可能在主仓库或 worktree 中，CWD 不一定是项目根目录。必须动态定位：

```bash
ROOT=$(git rev-parse --show-toplevel)
LINT="$ROOT/.claude/skills/golangci-lint/bin/golangci-lint.exe"
```

后续所有命令中的 `$LINT` 即为此变量。

## 快速使用

```bash
# 定位二进制
ROOT=$(git rev-parse --show-toplevel)
LINT="$ROOT/.claude/skills/golangci-lint/bin/golangci-lint.exe"

# 检查整个项目（最常用，在 Go module 根目录下执行）
cd "$ROOT/rain-qa-func" && "$LINT" run ./...

# 检查指定包
cd "$ROOT/rain-qa-func" && "$LINT" run ./pkg/proto-test/...

# 自动修复可修复的问题
cd "$ROOT/rain-qa-func" && "$LINT" run --fix ./...

# 查看启用了哪些 linter
"$LINT" linters
```

## 工作流程

收到 lint 检查请求后，按以下步骤执行：

### 1. 确定范围

根据上下文自动判断检查范围，不要盲目跑整个项目：

| 场景 | 范围 | 命令 |
|------|------|------|
| "检查一下" / 任务完成自检 | 最近修改的包 | 基于本次修改的文件确定 |
| "检查整个项目" | 全量 | `run ./...` |
| "检查 proto-test" | 指定包 | `run ./pkg/proto-test/...` |
| "检查这个文件" | 单文件 | `run ./path/to/file.go` |

### 2. 执行并分析

运行 lint 后，对输出中的每个问题：

- **error 级别** — 必须修复，除非明确是误报
- **warning 级别** — 建议修复，报告给用户决定
- **识别误报** — 以下目录的输出通常是误报，应排除或忽略：
  - `xcard_pb/` — protobuf 生成代码
  - `vendor/` — 第三方依赖
  - 以 `_test.go` 结尾的测试文件中的某些模式

排除 protobuf 目录：
```bash
cd "$ROOT/rain-qa-func" && "$LINT" run --exclude-dirs "xcard_pb" ./...
```

### 3. 修复并验证

- 能用 `--fix` 自动修的问题优先自动修
- 手动修复时逐个处理，修完一个跑一次验证
- 所有修复完成后最终跑一次全量确认零问题

### 4. 汇报结果

向用户报告：
- 发现 X 个问题（Y error, Z warning）
- 已修复 A 个，B 个是误报已跳过
- 如有未修复的 warning，列出供用户决定

## 默认配置

项目无 `.golangci.yml`，使用 5 个默认 linter：`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`。

如需更严格检查，可临时启用额外 linter：

```bash
# 推荐：格式 + 拼写 + 风格
cd "$ROOT/rain-qa-func" && "$LINT" run --enable gofmt,misspell,gocritic ./...
```

| 推荐额外 linter | 检查什么 |
|----------------|---------|
| `gofmt` | 代码格式是否规范 |
| `misspell` | 注释中的英文拼写 |
| `gocritic` | 代码风格和常见坑 |
| `unconvert` | 多余的类型转换 |
| `prealloc` | slice 缺少预分配 |
