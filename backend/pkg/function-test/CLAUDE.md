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
