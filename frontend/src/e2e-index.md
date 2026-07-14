# E2E 测试索引

各页面目录下的 `CLAUDE.md` 均包含 E2E 测试索引。全局测试文件：

| 测试文件                                  | 说明                                                                       |
| ----------------------------------------- | -------------------------------------------------------------------------- |
| `e2e/layout/layout.spec.ts`                    | 通用布局测试（导航栏渲染、页面路由切换、状态栏、布局结构、连续路由切换）   |
| `e2e/home/home.spec.ts`                      | AI 助手首页测试（见 `pages/llm/CLAUDE.md`）                              |
| `e2e/function-test/function-test.spec.ts`             | 战斗测试页测试（见 `pages/function-test/CLAUDE.md`）                     |
| `e2e/excel-test/excel-test.spec.ts`                | 配表测试页测试（见 `pages/excel-test/CLAUDE.md`）                        |
| `e2e/proto-test/proto-test-base.spec.ts` 等 6 个 | 协议重放页测试（拆分自主套件，见 `pages/proto-test/CLAUDE.md`） |
| `e2e/hero-wiki-check/hero-wiki-check.spec.ts`           | 武将 Wiki 检查页测试（见 `pages/hero-wiki-check/CLAUDE.md`）             |
| `e2e/activity-wiki/activity-wiki.spec.ts`             | 活动 Wiki 检查页测试（见 `pages/activity-wiki-check/CLAUDE.md`）         |
| `e2e/hero-voice-resource-check/hero-voice-resource-check.spec.ts` | 武将语音资源检查页测试（见 `pages/hero-voice-resource-check/CLAUDE.md`） |
| `e2e/settings/settings.spec.ts`                  | 设置页测试（见 `pages/settings/CLAUDE.md`）                              |
| `e2e/roadmap.spec.ts`                   | 路线图抽屉面板测试（见 `pages/settings/CLAUDE.md`）                      |

**规范要求**：

- 每个 `pages/<page-name>/CLAUDE.md` 必须包含 E2E 测试索引表格，列出对应测试文件、Page Object 和覆盖范围
- 新增页面时同步创建 E2E 测试并更新 `frontend/src/CLAUDE.md` 本索引
- 修改页面时按"前端变更强制同步规范"同步更新布局文档和 E2E 测试（见上方"前端变更强制同步规范"章节）



## 未完成的任务

- 将proto-test相关的e2e组织到一个目录下，并在CLAUDE.md中说明