# fight-test — 战斗测试用例工具

## 简介

fight-test 子命令提供战斗测试用例（`cases/fight_cases/`）的 CLI 能力：用例列表浏览与战斗测试执行。适用于 AI Agent 自动化测试和批量回归场景。

战斗测试执行复用 GUI / MCP 的同一套机器人跑测逻辑（`RunRobotTest`），配置默认值读取 `.rain-qa-func.json` 中的 `function_test` 配置段。

## 用法

```
rain-qa-func fight-test [子命令] [选项]
```

## 子命令

```
fight-test
├── list          # 列出战斗测试用例
└── run           # 运行战斗测试（核心命令）
```

## fight-test list — 列出用例

```bash
rain-qa-func fight-test list [--dir <path>] [--hero <id>] [--format table|json]
```

| Flag | 默认 | 说明 |
|------|------|------|
| `--dir` | 取配置 `jsons_dir`（默认 `cases/fight_cases`） | 用例目录 |
| `--hero` | `0`（不限） | 按英雄 ID 过滤（匹配 `{id}_*.json` 和 `{id}-*.json`） |
| `--format` | `table` | 输出格式：`table`（默认）或 `json` |

table 格式输出示例：

```
CASE                           STEPS  HEROS  FILE
司马懿_狼顾鹰视_1                   1      3  10006_司马懿.json
...
共 88 条用例（目录: cases/fight_cases）
```

## fight-test run — 运行战斗测试（核心命令）

```bash
rain-qa-func fight-test run [flags]
```

| Flag | 默认 | 说明 |
|------|------|------|
| `--server` | 取配置 `server_addr` | 目标服务器 IP |
| `--port` | 取配置 `server_port`（20144） | 目标服务器端口 |
| `--dir` | 取配置 `jsons_dir` | 用例目录 |
| `--hero` | `0`（不限） | 英雄 ID，>0 时匹配 `{id}_*.json` / `{id}-*.json` |
| `--case` | `""` | 仅运行指定用例名（`Case` 字段） |
| `--file` | `[]`（可多次指定） | 直接指定用例文件名，优先级高于 `--hero` |
| `--prefix` | 取配置 `robot_prefix` | 机器人账号前缀 |
| `--op-time` | `30000` | 操作超时（毫秒） |
| `--concurrency` | 取配置 `concurrency` | 并发数 |
| `--timeout` | `10m` | 整体超时 |

### 过滤优先级

1. `--file`（可多次指定）：直接指定文件名列表，忽略 `--hero`。
2. `--hero`：>0 时按英雄 ID 匹配文件。
3. 均未指定：运行 `--dir` 下全部用例。

`--case` 可与上述任意方式组合，用于在选中的文件内进一步按用例名过滤。

### 执行日志输出

执行过程中，每条 step 日志实时输出到终端（由 `cli_emitter.go` 驱动），格式：

```
[2026-07-07T12:00:00.000000], 动作[1], Step[赵云_七进七出], name[pf_qax0x1], [INFO], UseHeroSkill Start
```

字段：`Time`（微秒时间戳）/ `动作[ID]`（step 动作编号，-1 为非动作日志）/ `Step[Case 名]`（CLI 无前端用例 steps 上下文，用 Case 名兜底）/ `name[机器人名]` / `[Level]`（INFO/WARN/ERROR）/ 消息正文。逐条日志会与 robot 库自带的粗粒度日志（如「进度: 1/1」「成功加载文件」）交错输出，属预期。

## 示例

```bash
# 列出所有战斗测试用例
rain-qa-func fight-test list

# 列出英雄 10006 的用例
rain-qa-func fight-test list --hero 10006

# 运行英雄 10006 的全部用例
rain-qa-func fight-test run --hero 10006

# 指定服务器运行单个文件
rain-qa-func fight-test run --server 10.254.114.241 --file 10006_司马懿.json

# 运行指定用例名
rain-qa-func fight-test run --hero 10006 --case 司马懿_狼顾鹰视_1

# 查看帮助
rain-qa-func fight-test --help
rain-qa-func fight-test run --help
```

## 更多信息

详见 `backend/pkg/function-test/` 目录。战斗用例数据结构定义见 `cases/fight_cases_schema.json` 与 `services.go` 中的 `QAFuncCase` 结构体。
