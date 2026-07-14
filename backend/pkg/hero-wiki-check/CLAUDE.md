# hero-wiki-check

武将 Wiki 资源检查，对比新旧武将数据差异并输出检查报告。

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| HeroWikiResCheckService | wails.go:14 | Wiki 资源检查服务 |
| HeroWikiGameService | wails.go:58 | 前端 Game 服务包装器 |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| NewHeroWikiResCheckService | wails.go:17 | 创建检查服务实例 |
| Check | wails.go:25 | 执行 Wiki 资源检查，返回差异结果 |
| Save | wails.go:50 | 保存检查结果到指定路径 |
| RegisterWikiCheckTools | mcp.go:13 | 注册 Wiki 检查 MCP Tools |

## 开发注意事项
- 同时支持前端调用和 MCP 工具调用（@frontend 和 @mcp 标签）
- 检查结果通过 diff.DataContainer 结构返回
- 详见 [武将Wiki检查-实现详解](../../../docs/武将Wiki检查-实现详解.md)

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/hero-wiki-check/hero-wiki-check.spec.ts`](../../../frontend/e2e/hero-wiki-check/hero-wiki-check.spec.ts) | 页面加载、检查执行、差异展示、保存结果、筛选过滤 |

## 依赖关系
- 依赖：pkg/rain-resources-checker/diff、pkg/rain-resources-checker/herowiki、pkg/rain-resources-checker/mjs_excel、common/game
