# E2E 测试同步规范

前端变更与 E2E 测试的同步规则。

## 何时更新 E2E 测试

| 场景 | 需要的操作 |
|------|-----------|
| 新增页面 | 创建 `e2e/shared/pages/[PageName]Page.ts` + `e2e/[page-name].spec.ts`，并在 `pages/<page-name>/CLAUDE.md` 和对应后端 `pkg/<package>/CLAUDE.md` 中添加 E2E 测试索引 |
| 新增组件 | 添加可见性/交互测试，并更新对应前后端 `CLAUDE.md` 的 E2E 索引 |
| 删除组件 | 删除或更新相关测试，并更新对应前后端 `CLAUDE.md` 的 E2E 索引 |
| UI 文本变更 | 更新测试断言 |
| 路由变更 | 更新 `e2e/shared/pages/BasePage.ts` 中的 Route 枚举 |
| 新增按钮/可交互元素 | 添加点击/交互测试，并更新对应前后端 `CLAUDE.md` 的 E2E 索引 |
| 新增表单字段 | 添加输入验证测试，并更新对应前后端 `CLAUDE.md` 的 E2E 索引 |
| 新增弹窗/对话框 | 添加打开/关闭/内容测试，并更新对应前后端 `CLAUDE.md` 的 E2E 索引 |

## E2E 测试文件结构

```
frontend/e2e/
├── fixtures/index.ts        # 共享测试配置（CDP 连接 + Page Object 注入）
├── pages/                   # Page Object Model
│   ├── BasePage.ts          # 基类 + Route 枚举
│   └── [PageName]Page.ts    # 页面定位器和操作
├── [page-name].spec.ts      # 测试用例
└── utils/
    └── helpers.ts           # 测试工具
```

## 测试覆盖要求

每个页面应覆盖：
1. **页面加载** — 主要元素可见
2. **导航** — 能从根路径到达
3. **交互** — 按钮、输入框、开关正常工作
4. **数据展示** — 表格、列表、卡片正确渲染
5. **边界情况** — 空状态、错误处理

## 新增页面完整流程

### 1. 创建 Page Object

参考现有 Page Object 模式（如 `e2e/shared/pages/StreamProxyPage.ts`）：

```typescript
export class NewPagePage extends BasePage {
  readonly pageContainer: Locator;
  constructor(page: Page) {
    super(page);
    this.pageContainer = page.locator('#new-page');
  }
  async goto(): Promise<void> {
    await this.page.locator('#layout-header button:has-text("NewPage")').click();
    await sleep(800);
  }
}
```

### 2. 在 BasePage.ts 添加路由枚举

### 3. 在 fixtures/index.ts 注册 Page Object

### 4. 创建测试文件

参考现有 spec 文件模式。导航使用 `goto()` 点击菜单按钮（Wails 内存路由不支持 URL 导航）。

### 5. 在页面 CLAUDE.md 中添加 E2E 测试索引

```markdown
## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/new-page.spec.ts` | [`NewPagePage`](../../e2e/shared/pages/NewPagePage.ts) | 页面加载、导航、交互、数据展示、边界情况 |
```

### 6. 在后端包 CLAUDE.md 中添加 E2E 测试索引

在对应后端包（如 `backend/pkg/<package>/CLAUDE.md`）中添加 E2E 索引：

```markdown
## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/new-page.spec.ts`](../../../frontend/e2e/new-page.spec.ts) | 页面加载、导航、交互、数据展示、边界情况 |
```

### 7. 更新 frontend/src/CLAUDE.md 全局 E2E 索引

在 `frontend/src/CLAUDE.md` 的 E2E 测试索引表格中添加新页面条目。

### 8. 更新根目录 CLAUDE.md 全局 E2E 索引

在根目录 `CLAUDE.md` 的 `## E2E 测试索引` 表格中添加后端包→E2E 测试映射条目。

## 页面 CLAUDE.md E2E 索引规范

每个 `pages/<page-name>/CLAUDE.md` 和对应后端 `pkg/<package>/CLAUDE.md` 必须包含 `## E2E 测试` 章节，格式如下：

```markdown
## E2E 测试

| 测试文件 | Page Object | 覆盖范围 |
|----------|-------------|----------|
| `e2e/<page-name>.spec.ts` | [`<PageName>Page`](../../e2e/shared/pages/<PageName>Page.ts) | 用逗号分隔的测试区域列表 |
```

**前端页面 CLAUDE.md 要求**：
- 测试文件列使用相对 `frontend/e2e/` 的路径
- Page Object 列使用相对链接指向源码文件
- 覆盖范围列简明列出测试的主要功能区域
- 新增/修改测试时必须同步更新索引

**后端包 CLAUDE.md 要求**：
- 测试文件列使用相对链接指向 `frontend/e2e/` 下的 spec 文件
- 覆盖范围列简明列出测试的主要功能区域
- 修改后端 API 或数据结构时，需检查并同步更新对应 E2E 测试

## 运行 E2E 测试

```bash
cd frontend
npx playwright test                      # 运行所有测试
npx playwright test --ui                 # UI 模式运行
npx playwright test --grep "page name"   # 运行指定测试
```
