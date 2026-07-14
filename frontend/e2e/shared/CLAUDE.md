# shared 共享模块

所有 E2E 测试共享的基础设施。

## 目录结构

| 路径 | 用途 |
|------|------|
| `fixtures/index.ts` | Playwright fixture 定义（CDP 连接、Page Object 注入） |
| `pages/BasePage.ts` | Page Object 基类（路由表、导航方法） |
| `pages/ProtoTestPage.ts` | 协议重放页 Page Object（最复杂，700+ 行） |
| `pages/FunctionTestPage.ts` | 战斗测试页 |
| `pages/ExcelTestPage.ts` | 配表测试页 |
| `pages/HeroWikiCheckPage.ts` | 武将 Wiki 检查页 |
| `pages/HeroVoiceResourceCheckPage.ts` | 武将语音资源检查页 |
| `pages/ActivityWikiPage.ts` | 活动 Wiki 检查页 |
| `pages/SettingsPage.ts` | 设置页 |
| `pages/RoadmapPage.ts` | 路线图页 |
| `pages/HomePage.ts` | AI 助手首页 |
| `utils/helpers.ts` | sleep、截图比较、拖拽、滚动等通用函数 |
| `docs/goto-analysis.md` | `goto()` 方法设计分析与改进建议 |

## 新增 Page Object 指南

1. 继承 `BasePage`，在构造函数中初始化关键 Locator
2. 实现 `goto()` 方法（参考 `ProtoTestPage.goto()` 和 `docs/goto-analysis.md`）
3. 在 `fixtures/index.ts` 中注册 fixture
4. 使用 `data-testid` 属性定位元素（而非依赖 Naive UI 内部 DOM）
