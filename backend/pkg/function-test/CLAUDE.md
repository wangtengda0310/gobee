# function-test

功能测试核心包，提供测试用例管理、机器人测试执行、配置管理和 MCP 工具集成。

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| JsonCaseService | mcp.go:23 | JSON 用例服务（用例管理、执行） |
| QAFuncCase | mcp.go:36 | 测试用例 JSON 数据结构 |
| CaseInfo | mcp.go:89 | 用例基本信息（列表展示用） |
| CategoryInfo | mcp.go:101 | 分类信息 |
| LogEntry | mcp.go:235 | 日志条目 |
| FuncCaseConfigService | mcp.go:856 | 配置管理服务 |
| FightCaseService | mcp.go:476 | 战斗用例服务接口 |
| FightCaseTools | mcp.go:482 | 战斗测试工具集合 |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| NewJsonCaseService | mcp.go:30 | 创建 JSON 用例服务实例 |
| RunRobotTest | mcp.go:110 | 执行机器人测试 |
| GetCaseList | mcp.go:685 | 获取用例列表（支持文件/目录） |
| GetCategories | mcp.go:733 | 获取分类信息 |
| SearchCases | mcp.go:836 | 搜索用例（关键词匹配） |
| GetTestCase | mcp.go:562 | 获取特定测试用例详情 |
| SaveCase | mcp.go:448 | 保存测试用例到 JSON 文件 |
| RegisterJsonCaseTools | mcp.go:25 | 注册功能测试 MCP Tools |
| RegisterFightCaseTools | mcp.go:499 | 注册战斗测试 MCP Tools |

## L6 结构化报告落盘（飞书用例翻译 skill 依赖）

`RunRobotTest`（GUI「执行」/ MCP `run_fight_test`）结束后，将内存中的 ERROR/FAIL 日志聚合并写入文件，供 skill 读 `latest.md` 做 P-* 归因，**无需从 GUI 手抄日志**。

| 文件 | 职责 |
|------|------|
| [`fight_test_report.go`](fight_test_report.go) | 报告结构体、`failureKind` 分类、`writeFightTestReport` 写 `latest.md` / `latest.json` 与时间戳归档 |
| [`fight_test_report_test.go`](fight_test_report_test.go) | 单元测试：`classifyFailureKind`、报告目录、`writeFightTestReport` |
| [`services.go`](services.go) `RunRobotTest` 末尾 | 调用 `persistFightTestReportAfterRun`（**hook 点**；缺此调用则不会落盘） |

**输出路径**（相对 GUI 配置的用例目录 `casesDir`）：

```
{casesDir}/../fight_test_reports/latest.md      ← skill 优先读
{casesDir}/../fight_test_reports/latest.json
{casesDir}/../fight_test_reports/{时间戳}_{任务}.md/.json
```

控制台会打印：`战斗测试报告: ... latest.md: ...`

**Git 提交**：以上三个 Go 文件为**同一功能**，须**一起**提交；只改 skill 文档而不提交后端时，旧 `rain-qa-func.exe` 不会产生报告。技能侧说明见 `.claude/skills/feishu-fight-case-translator/reference/9.l6-failure-patterns.md`。

## CLI fight-test run 实时日志

`RunRobotTest` 三入口的逐条 step 日志去向（robot 库 → `wails3LogChan` → `emitRobotLog` → `emitter.Emit`）：

| 入口 | 构造点 | emitter | 逐条日志去向 |
|------|--------|---------|------------|
| GUI | `cmd/rain-qa-func/wails.go:123` | `app.Event` | 实时推前端 `robotLog` 事件 |
| MCP | `backend/pkg/settings/mcp/services.go:37` | `nil` | 丢弃；`buildFightTestResult`（mcp.go）把全量 `Logs` 塞进返回值给 AI |
| **CLI** | `cobra.go` `newFightTestCaseServiceForCLI` | `newCLILogEmitter()` | 实时格式化打印到 stdout |

| 文件 | 职责 |
|------|------|
| [`cli_emitter.go`](cli_emitter.go) | `cliLogEmitter` 实现 `RobotLogEmitter`（emitter.go），把 `*log_service.Wails3Log` 格式化为 `[Time], 动作[ID], Step[Case], name[RobotName], [Level], Msg` 写入 stdout |
| [`cli_emitter_test.go`](cli_emitter_test.go) | 单元测试：格式化、事件名过滤、nil/类型断言/空 data 守卫、Level 经 String() |

输出格式对齐前端 `robot-test-log.vue`，但 `Step` 字段前端用「用例 steps 的 desc」，CLI 无此上下文用 `Case` 名兜底；终端会与 robot 库自带的粗粒度 stdout 日志（如「进度: 1/1」「成功加载文件」）交错，属预期。

## 开发注意事项
- 测试运行状态使用 atomic.Bool 确保线程安全
- 支持飞书消息劫持功能，通过 InterceptService 控制
- 配置文件默认 `.rain-qa-func.json` (section: `function_test`)
- MCP 工具支持异步测试（run_fight_test_async）和进度查询（get_test_progress）
- 日志使用 vars.SaveQAFuncLogCache 缓存

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/function-test/function-test.spec.ts`](../../../frontend/e2e/function-test/function-test.spec.ts) | 页面加载、用例树、配置面板、步骤编辑、运行测试、保存用例 |

## 依赖关系
- 依赖：rain-robot（日志/客户端）、common（配置/游戏服务/飞书通知）
