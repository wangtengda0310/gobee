---
name: write-e2e-test
description: 编写 E2E 测试用例。当用户要求编写、新增、补充 E2E 测试时触发。基于 Playwright + CDP 连接 WebView2 架构，使用 Page Object 模式组织代码。
---

# 编写 E2E 测试用例

本项目 E2E 测试基于 **Playwright 通过 CDP 连接运行中的 Wails WebView2 实例**，不是传统的浏览器测试。编写测试时必须遵循本规范。

## 目录结构

```
frontend/e2e/
├── fixtures/index.ts        # 测试 fixtures（CDP 连接 + Page Object 注入）
├── pages/                   # Page Object Model
│   ├── BasePage.ts          # 基类（Route 枚举、通用操作）
│   └── [PageName]Page.ts    # 各页面 Page Object
├── utils/helpers.ts         # Naive UI 组件操作工具函数
├── [page-name].spec.ts      # 测试用例文件
└── tsconfig.json
```

## 前置条件

测试依赖 `wails3 dev` 启动的运行中应用，通过 CDP 端口连接：

```
1. 启动应用: wails3 dev
2. 运行测试: cd frontend && npx playwright test <spec-file>
3. 单个文件: npx playwright test e2e/function-test/function-test.spec.ts
```

默认 CDP 端口 9223，可通过 `CDP_PORT` 环境变量修改。

## 文件模板

### 新页面 Page Object

```typescript
// e2e/shared/pages/NewPagePage.ts
import { Page, Locator, expect } from '@playwright/test';
import { BasePage, Route, resolveRoute } from './BasePage';
import { sleep, waitForVisible } from '../utils/helpers';

export class NewPagePage extends BasePage {
  // 定位器 —— 按页面区域组织
  readonly headerContainer: Locator;
  readonly mainContent: Locator;

  constructor(page: Page) {
    super(page);
    this.headerContainer = page.locator('.page-header');
    this.mainContent = page.locator('.main-content');
  }

  // 导航到页面（必须实现）
  async goto(): Promise<void> {
    await this.page.goto(resolveRoute(this.page, Route.NEW_PAGE));
    await sleep(500);
  }
}
```

### 新页面测试文件

```typescript
// e2e/new-page.spec.ts
import { test, expect, describe } from './fixtures';
import { sleep } from './utils/helpers';

describe('新页面名称', () => {
  // beforeEach: 确保在正确的页面上
  test.beforeEach(async ({ page }) => {
    // 关闭可能遮挡的模态框
    const modalMask = page.locator('.n-modal-mask');
    if (await modalMask.isVisible()) {
      await page.keyboard.press('Escape');
      await sleep(300);
    }

    // 导航到目标页面
    const currentUrl = page.url();
    if (!currentUrl.includes('/NewPage')) {
      // 使用菜单导航或直接 goto
      await page.goto(`http://wails.localhost:9245/NewPage`);
      await sleep(500);
    }
  });

  test('页面加载 - 主要元素可见', async ({ page }) => {
    await expect(page.locator('.main-content')).toBeVisible();
  });
});
```

### 注册 Page Object 到 Fixtures

在 `fixtures/index.ts` 中添加：

```typescript
// 1. 导入
import { NewPagePage } from '../pages/NewPagePage';

// 2. 添加到 AppFixtures 类型
type AppFixtures = {
  // ... 已有的
  newPagePage: NewPagePage;
};

// 3. 添加到 test.extend
export const test = base.extend<AppFixtures>({
  // ... 已有的
  newPagePage: async ({ page }, use) => {
    const newPagePage = new NewPagePage(page);
    await use(newPagePage);
  },
});
```

### 添加路由到 BasePage

在 `pages/BasePage.ts` 的 `Route` 枚举中添加：

```typescript
export enum Route {
  // ... 已有的
  NEW_PAGE = '/NewPage',
}
```

## 核心规则（必须遵守）

### 1. CDP 连接架构限制

本项目的 E2E 测试**不是传统浏览器测试**，是通过 CDP 连接到 WebView2：

- **禁止**在测试中启动新浏览器或新页面
- **禁止**使用 `page.evaluate()` 调用 `import()` 动态加载模块
- **禁止**在 `evaluate` 中访问 Wails runtime（`@wailsio/runtime`）
- **路由使用完整 URL**：`http://wails.localhost:9245/Test`，不能用相对路径 `/Test`
- 使用 `resolveRoute(page, Route.XXX)` 辅助函数处理路由

### 2. Naive UI 组件选择器

本项目使用 Naive UI 组件库，常见选择器：

| 组件 | 选择器 | 说明 |
|------|--------|------|
| 按钮 | `button:has-text("文本")` | **不要用** `.n-button--primary-type`，不稳定 |
| 确认对话框按钮 | `.n-button:has-text("确定")` | **不要用** `.n-dialog .n-button--primary-type` |
| 树节点 | `.n-tree-node-content:has-text("文本")` | |
| Tab 标签 | `.n-tabs-tab:has-text("文本")` | |
| 模态框遮罩 | `.n-modal-mask` | beforeEach 中需关闭 |
| 下拉选项 | `.n-base-select-option` | 展开 select 后出现 |
| 开关 | `.n-switch` | 用 `aria-checked` 判断状态 |
| 加载中 | `.n-spin` | |
| Toast 消息 | `.n-message:has-text("文本")` | |

**重要**：Naive UI 对话框的确认按钮用 `.n-button:has-text("确定")`，不要用 CSS class 匹配（如 `.n-button--primary-type`），因为 Naive UI 的 class 名会随版本和主题变化。

### 3. beforeEach 中的模态框清理

模态框遮罩 (`.n-modal-mask`) 会拦截所有点击操作，导致导航失败。**每个测试文件的 `beforeEach` 必须清理模态框**：

```typescript
test.beforeEach(async ({ page }) => {
  // 清理残留的模态框
  const modalMask = page.locator('.n-modal-mask');
  if (await modalMask.isVisible()) {
    await page.keyboard.press('Escape');
    await sleep(300);
  }

  // 导航到目标页面...
});
```

### 4. 数据验证方式

由于 CDP 环境限制，数据验证只能通过 DOM：

| 方法 | 可用性 | 说明 |
|------|--------|------|
| DOM 文本内容 | **推荐** | `locator.textContent()` / `expect(locator).toContainText()` |
| DOM 属性 | **推荐** | `locator.getAttribute()` / `locator.evaluate()` |
| 元素计数 | **推荐** | `locator.count()` |
| 可见性断言 | **推荐** | `expect(locator).toBeVisible()` |
| `page.evaluate()` | **受限** | 只能访问浏览器原生 API，不能 import |
| `page.on('console')` | **可用** | 收集前端 console.log 用于调试 |
| 截图 | **可用但谨慎** | Wails WebView2 截图可能与实际显示有差异 |

**调试技巧** — 收集前端日志：

```typescript
const consoleLogs: string[] = [];
page.on('console', msg => {
  if (msg.text().includes('[关键模块名]')) {
    consoleLogs.push(`[${msg.type()}] ${msg.text()}`);
  }
});
// ... 操作后检查 consoleLogs
```

### 5. 等待策略

Wails 应用中前端和 Go 后端通过 Bridge 通信，操作响应有延迟：

```typescript
// 点击按钮后等待后端处理
await button.click();
await sleep(500);  // 至少 500ms，复杂操作（如加载用例）需要更长

// 加载用例等耗时操作
await confirmButton.click();
await sleep(6000);  // Excel 初始化可能需要几秒

// 更好的方式：等待特定元素出现
await expect(page.locator('.result-element')).toBeVisible({ timeout: 10000 });
```

**不要过度使用 `sleep`**：优先使用 `waitFor` + `expect` 断言等待，`sleep` 作为兜底。

## 测试用例组织模式

### 按功能区域分组

```typescript
describe('战斗测试 - 页面名称', () => {
  describe('Header 区域', () => {
    test('加载用例按钮', async ({ page }) => { ... });
    test('保存用例按钮', async ({ page }) => { ... });
  });

  describe('Tree 区域', () => {
    test('搜索用例', async ({ page }) => { ... });
    test('选择用例节点', async ({ page }) => { ... });
  });

  describe('Config Tab', () => {
    test('下拉菜单数据加载', async ({ page }) => { ... });
  });
});
```

### 使用 Fixture 中的 Page Object

```typescript
// 通过 fixture 注入 Page Object（已配置好 CDP 连接）
test('使用 Page Object', async ({ functionTestPage }) => {
  await functionTestPage.goto();
  await functionTestPage.clickLoadCases();
});

// 也可以直接用 page（CDP 连接的 WebView2 页面）
test('直接操作 page', async ({ page }) => {
  await page.locator('button:has-text("加载用例")').click();
});
```

### 跳过测试

对于依赖后端服务的测试，使用 `test.skip()` 而非注释掉：

```typescript
test.skip('依赖后端 Excel 数据的测试', async ({ page }) => {
  // 需要 Excel 初始化成功的场景
});
```

## 常见场景代码片段

### 操作确认对话框

```typescript
// 点击触发对话框
await page.locator('button:has-text("操作")').click();
await sleep(300);

// 点击确定按钮（注意选择器）
const confirmBtn = page.locator('.n-button:has-text("确定")').first();
await confirmBtn.waitFor({ state: 'visible', timeout: 5000 });
await confirmBtn.click();
await sleep(500);
```

### 搜索并选择树节点

```typescript
// 搜索
await page.locator('input[placeholder="搜索"]').fill('关键词');
await sleep(500);

// 点击匹配的节点
const node = page.locator('.n-tree-node-content:has-text("关键词")');
await node.first().click();
await sleep(300);
```

### 检查下拉菜单选项

```typescript
// 展开下拉菜单
const select = page.locator('.n-select').first();
await select.click();
await sleep(1500);

// 检查选项数量
const optionCount = await page.locator('.n-base-select-option').count();
console.log(`选项数量: ${optionCount}`);

// 关闭下拉菜单
await page.keyboard.press('Escape');
```

### Tab 切换

```typescript
await page.locator('.n-tabs-tab:has-text("配置")').click();
await sleep(300);
```

## 完整流程：新增页面 E2E 测试

1. 创建 `e2e/shared/pages/NewPagePage.ts`（Page Object）
2. 在 `e2e/shared/pages/BasePage.ts` 的 `Route` 枚举中添加路由
3. 在 `e2e/fixtures/index.ts` 中注册 Page Object
4. 创建 `e2e/new-page.spec.ts`（测试用例）
5. **在前端页面 `pages/<page-name>/CLAUDE.md` 中添加 E2E 测试索引表格**
6. **在后端包 `backend/pkg/<package>/CLAUDE.md` 中添加 E2E 测试索引**
7. **更新 `frontend/src/CLAUDE.md` 全局 E2E 索引**
8. **更新根目录 `CLAUDE.md` 全局 E2E 索引（后端包→E2E 映射）**
9. 运行验证：`cd frontend && npx playwright test e2e/new-page.spec.ts`

### E2E 测试索引规范

**前端页面** — 每个 `pages/<page-name>/CLAUDE.md` 必须包含 `## E2E 测试` 章节：

```markdown
## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/new-page.spec.ts` | [`NewPagePage`](../../e2e/shared/pages/NewPagePage.ts) | 页面加载、导航、交互、数据展示、边界情况 |
```

**后端包** — 每个对应后端 `pkg/<package>/CLAUDE.md` 必须包含 `## E2E 测试` 章节：

```markdown
## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/new-page.spec.ts`](../../../frontend/e2e/new-page.spec.ts) | 页面加载、导航、交互、数据展示、边界情况 |
```

**要求**：
- 前端页面：测试文件列使用相对 `frontend/e2e/` 的路径；Page Object 列使用相对链接指向源码文件
- 后端包：测试文件列使用相对链接指向 `frontend/e2e/` 下的 spec 文件
- 覆盖范围列简明列出测试的主要功能区域
- 新增/修改测试时必须同步更新前后端索引
- 修改后端 API 或数据结构时，需检查并同步更新对应 E2E 测试

`frontend/src/CLAUDE.md` 和根目录 `CLAUDE.md` 中维护全局 E2E 测试索引，汇总所有页面的测试文件及后端包映射。

## 调试技巧

### 1. 运行单个测试并查看输出

```bash
cd frontend
npx playwright test e2e/debug.spec.ts --reporter=list
```

### 2. 收集前端 console.log

```typescript
const logs: string[] = [];
page.on('console', msg => {
  logs.push(`[${msg.type()}] ${msg.text()}`);
});
// 测试结束后输出
test.afterEach(() => {
  console.log('收集到的日志:', logs.join('\n'));
});
```

### 3. DOM 快照调试

```typescript
// 获取元素及其子元素的完整信息
const state = await page.locator('.target-element').evaluate((el: HTMLElement) => {
  return {
    text: el.textContent?.trim(),
    children: el.children.length,
    html: el.innerHTML.substring(0, 200),
  };
});
console.log('元素状态:', JSON.stringify(state, null, 2));
```

### 4. 临时调试脚本

调试用的临时 spec 文件放在 `e2e/` 下，文件名以 `debug-` 开头（如 `debug-cards-dropdown.spec.ts`），调试完成后删除。

## 禁止事项

- **禁止**在 `evaluate()` 中使用 `import()` 或 Wails runtime
- **禁止**用 Naive UI 内部 CSS class 作为选择器（如 `.n-button--primary-type`），用文本匹配代替
- **禁止**忽略模态框清理（`.n-modal-mask` 会拦截点击）
- **禁止**用相对路径导航（CDP 模式下必须用完整 URL）
- **禁止**注释掉测试而非使用 `test.skip()`
- **禁止**提交 `debug-*.spec.ts` 调试文件到版本控制
