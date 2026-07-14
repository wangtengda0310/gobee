# E2E 测试

rain-qa-func 前端的端到端测试，基于 Playwright 框架。
e2e测试中遇到的问题及解决方案及时记录在本文件中，供后续参考。

## 目录结构

```
e2e/
├── shared/                      # 共享：fixtures、pages、utils、docs
│   ├── fixtures/index.ts        # Playwright fixture + CDP 连接
│   ├── pages/*.ts               # 所有 Page Object
│   ├── utils/helpers.ts         # 通用辅助函数
│   └── docs/                    # 调试与分析文档
├── proto-test/                  # 协议重放（录制/重放/测试用例/拦截）
├── function-test/               # 战斗测试
├── excel-test/                  # 配表测试（含关系链/正则/过滤条件）
├── hero-wiki-check/             # 武将 Wiki 检查
├── hero-voice-resource-check/   # 武将语音资源检查
├── activity-wiki/               # 活动 Wiki 检查
├── settings/                    # 设置页（含路线图/P2P/布局调试）
├── home/                        # AI 助手首页
├── layout/                      # 通用布局与导航
└── CLAUDE.md                    # 本文件
```

### 模块与后端包对应关系

| E2E 目录 | 后端包 | 前端页面路由 |
|----------|--------|-------------|
| `proto-test/` | `pkg/proto-test/` | /ProtoTest |
| `function-test/` | `pkg/function-test/` | /Test |
| `excel-test/` | `pkg/excel-test/` | /Excel |
| `hero-wiki-check/` | `pkg/hero-wiki-check/` | /HeroWikiRes |
| `hero-voice-resource-check/` | `pkg/rain-resources-checker/` | /HeroRes |
| `activity-wiki/` | `pkg/activity-wiki-check/` | /ActivityWiki |
| `settings/` | `pkg/settings/` | /Settings |
| `home/` | `pkg/settings/home/` | /Home |
| `layout/` | 通用 | 全局 |

### 共享模块

| 路径 | 用途 |
|------|------|
| `shared/fixtures/index.ts` | CDP 连接 + Page Object fixture 注入 |
| `shared/pages/BasePage.ts` | Page Object 基类（路由、导航） |
| `shared/pages/*.ts` | 各页面的 Page Object |
| `shared/utils/helpers.ts` | sleep、截图、拖拽等通用函数 |

## 测试架构

E2E 测试**必须**通过 CDP（Chrome DevTools Protocol）连接运行中的 Wails WebView2 实例，不能直接通过 HTTP 访问 Vite dev server（Wails 使用内存路由，不支持 URL 导航）。

### 运行方式

```bash
wails3 dev                                    # 1. 启动应用（启用 CDP）
cd frontend && npx playwright test *.spec.ts  # 2. 运行测试

# 按模块运行
npx playwright test proto-test/             # 协议重放全部
npx playwright test excel-test/               # 配表测试全部
npx playwright test proto-test/ -g "描述编辑"  # 指定用例
```

CDP 端口默认 `9223`，可通过环境变量 `CDP_PORT` 自定义。

### 连接方式

```typescript
// shared/fixtures/index.ts — 正确方式
const browser = await chromium.connectOverCDP('http://127.0.0.1:9223');
const context = browser.contexts()[0];
const page = context.pages()[0];
```

❌ 错误方式：启动新 Chrome 实例（`npx playwright test` 独立模式）、codegen 连接。

### 调试方法

创建临时测试文件检查 DOM + 截图，在 `test-results/` 查看输出。

### Playwright 配置

详见 `playwright.config.ts`：无 `baseURL`、失败时截图和 trace、Chromium 单浏览器。

## 已知问题与解决方案

### 1. Selector 不匹配（Naive UI 深层 DOM）

**问题**：Naive UI 组件渲染为多层 DOM，Playwright 标准选择器有时不生效。

| 组件 | 错误选择器 | 正确选择器 |
|------|-----------|-----------|
| `n-input-number` | `.n-input-number-number-input` | `.n-input-number input` |
| `n-data-table` 行 | `tbody tr` | `[data-row-key]` |
| `n-button` 多选按钮 | `button:has-text("多选")` 匹配到 "退出多选" | `.filter({ hasText: '多选' }).first()` |
| 目标服务输入框 | `input[placeholder*="服务器地址"]` | `input[placeholder*="TCP"]` |

**解决方案**：使用 placeholder 属性选择器定位输入框，使用 `data-row-key` 属性定位表格行，使用 `.filter({ hasText }).first()` 定位按钮。

### 2. 多次录制文件混乱

**问题**：发包改包页签无路径输入框，开始录制自动生成 `record_{timestamp}.json`（相对路径，落在 **worktree 根目录**，非 `cases/proto_cases/`）。二次录制可能复用同一 `filePath` 覆盖文件。

**当前处理**：无自动清理机制，手动删除根目录下 `record_*.json`。

**改造计划**：见 `backend/pkg/proto-test/docs/intercept-batch-release-investigation.md` §5；待办 `docs/TODO.md`。

### 3. bindings 生成失败（Windows rename 锁）

**问题**：`wails3 dev` 内部执行 `generate bindings -clean=true` 时，或 **dev 仍在运行时手动** `wails3 generate bindings`，删除旧 `frontend/bindings/` 后把 `.bindings-tmp-*` rename 过去会失败（`Access is denied`），可能导致 bindings 目录被清空。

**解决方案**：先停 `wails3 dev` / `rain-qa-func.exe`，必要时 `git restore frontend/bindings/`，再单独生成后重启：

```bash
wails3 generate bindings -ts
wails3 dev
```

若需拆分前后端启动，可先 `wails3 task build` 再分别启动。完整说明与恢复步骤见 [`docs/Wails开发注意事项.md`](../../docs/Wails开发注意事项.md)「wails3 dev bindings rename 失败（Windows）」。

### 4. 测试依赖后端录制数据

**问题**：部分测试需要真实的录制文件（`cases/session.json`）才能执行，纯 UI 测试无法覆盖。

- ✅ `test.skip` 标记的测试需要后端录制数据支持
- 当前已在 spec 中用 `skip` 标注

### 5. Naive UI 表格行选择器问题

**问题**：使用 `[data-row-key]` 选择器匹配表格行失败，导致 E2E 测试中表格行数始终为 0。

**根本原因**：
- Naive UI 的 `n-data-table` 组件不会自动生成 `data-row-key` 属性
- 数据已正确加载到表格（`tbody tr` 存在），但选择器匹配失败

**解决方案**：
```typescript
// ❌ 错误的选择器
getTableRows(): Locator {
  return this.messageTable.locator('[data-row-key]');
}

// ✅ 正确的选择器
getTableRows(): Locator {
  return this.messageTable.locator('tbody tr');
}
```

**验证方法**：
```bash
# 调试测试检查表格 DOM 结构
npx playwright test debug-table-dom.spec.ts

# 输出示例：
# 表格元素数量: 1
# tbody 行数: 2
# [data-row-key] 元素数量: 0  ← 关键发现
```

**影响范围**：
- 所有使用 `getTableRowCount()` 的测试用例
- 表格数据验证测试

### 6. Vue watch 监听器状态更新陷阱

**问题**：`watch(selectedCase, ...)` 中调用 `loadTestCase()` 后表格无数据。

**根因**：Service 返回数据不会自动设置到响应式状态，必须显式调用 `setRecordData(data)`。

**要点**：从 `useRecordData()` 解构获取 `setRecordData`，Vue 3 响应式状态必须通过显式调用更新方法触发视图更新。

**相关代码**：`proto-test/index.vue:95`（解构）、`proto-test/index.vue:118-127`（watch 实现）、`use-record-data.ts:130-132`（`setRecordData` 定义）

### 7. 多 wails3 进程导致测试连接错误实例

**问题**：多次运行 `wails3 dev` 可能导致多个 wails3.exe 进程同时存在，E2E 测试可能连接到错误的应用实例，测试操作与观察到的界面不一致。

**检查方法**：
```bash
# Windows PowerShell
tasklist | findstr "wails3"

# 终止所有旧进程
Stop-Process -Name "wails3" -Force
```

**调试前强制检查清单**：
1. 终止所有 wails3 进程：`Stop-Process -Name "wails3" -Force`
2. 终止所有应用进程：`Stop-Process -Name "rain-qa-func" -Force`
3. 启动单一实例：`wails3 dev`
4. 验证单进程：`tasklist | grep wails3` 确认只有 1 个
5. 验证 CDP：`curl http://localhost:9223/json/list`

**进程残留原因**：
- `wails3 dev` 启动主进程 + 子进程（Vite、Go 编译守护）
- Claude Code "停止"按钮只终止主进程，子进程可能残留
- 守护进程会自动重启被终止的进程
- Windows 进程树管理不如 Unix 严格

**症状与解决方案**：
| 症状 | 根因 | 解决方案 |
|------|------|----------|
| 测试通过但界面无变化 | 连接到旧实例 | 终止所有 wails3 重新启动 |
| 点击操作无响应 | 连接到错误实例 | 检查进程数量 |
| 元素查找超时 | 应用未正确启动 | 确认单一进程运行 |

### 8. Naive UI 遮罩层拦截点击（串行测试）

**问题**：前一个测试触发了模态框或抽屉对话框（如保存用例），测试结束后未关闭。下一个测试的 `goto()` 中导航按钮被 `.n-modal-mask` 或 `.n-drawer-mask` 遮挡，导致 `click()` 超时。

**错误表现**：
```
TimeoutError: locator.click: Timeout 10000ms exceeded.
- <div aria-hidden="true" class="n-drawer-mask"></div> from ... intercepts pointer events
```

**解决方案**：在 `goto()` 中增加遮罩清理逻辑：
```typescript
const modalMask = this.page.locator('.n-modal-mask');
const drawerMask = this.page.locator('.n-drawer-mask');
if ((await modalMask.count()) > 0 || (await drawerMask.count()) > 0) {
  await this.page.keyboard.press('Escape');
  await sleep(300);
}
```

### 9. DevTools 打开后 CDP 连接到错误页面

**问题**：用户手动打开 F12 DevTools 后，CDP 页面列表中出现两个 `type=page`（应用页 + DevTools 页）。`context.pages()[0]` 可能取到 DevTools 页面而非 Wails 应用，导致所有元素选择器找不到（如 `#layout-header` 不存在）。

**错误表现**：
```
TimeoutError: locator.click: Timeout 10000ms exceeded.
- waiting for locator('#layout-header button:has-text("Proto测试")')
```

**诊断方法**：
```bash
curl -s http://localhost:9223/json/list
# 如果有多个 type=page 条目，说明 DevTools 被打开
```

**解决方案**：运行 E2E 测试前确保 **DevTools 窗口已关闭**。

### 10. v-show 导致元素存在但不可见

**问题**：Wails 应用使用 `v-show`（而非 `v-if`）切换页签，DOM 元素始终存在但 `display: none`。Playwright 的 `click()` 默认要求元素可见，会报 "Element is not visible"。

**典型场景**：
- 在发包改包页签下尝试点击重放结果页签的下拉选择器
- 三个页签各有 `.n-data-table`，选择器可能匹配到隐藏页签的表格

**解决方案**：
1. 操作前显式切换到目标页签
2. 必要时使用 `{ force: true }` 绕过可见性检查（如 Naive UI 下拉选项被遮挡时）
3. 多个同类元素时使用 `.last()` 或 `.nth()` 指定正确的实例

### 11. 串行测试的状态累积

**问题**：E2E 测试串行运行（`--workers=1`），前一个测试修改的状态（如重放结果、多选模式）会影响后续测试。例如测试 2 产生了"测试用例"来源的重放结果，测试 1 断言来源标签时匹配到旧结果。

**错误表现**：
```
Expected substring: "发包改包"
Received string:    "来源: 测试用例 - 执行用例"
```

**解决方案**：在 `goto()` 中重置所有可变状态：
1. 清理模态框/抽屉遮罩
2. 切到重放结果页签，点击"清空"按钮清除历史结果
3. 切回发包改包页签

### 12. Vue 3 defineExpose ref 自动 unwrap 陷阱

**问题**：子组件通过 `defineExpose({ replayResults })` 暴露 `ref`，父组件通过模板 ref 访问时，Vue 3 **自动解包**（unwrap）`ref`。因此 `childRef.value.replayResults` 直接就是数组，**不需要再 `.value`**。

**错误代码**：
```typescript
const resultsRef = replayResultTabRef.value?.replayResults  // 已经是数组
const resultsArray = (resultsRef as any).value               // undefined！
const currentResult = resultsArray?.find(...)                // 失败
```

**正确代码**：
```typescript
const resultsArray = replayResultTabRef.value?.replayResults as any[] | undefined
const currentResult = resultsArray?.find(...)  // 正常工作
```

**诊断方法**：在关键节点打 `console.log(typeof refValue)` 或 `console.log(Array.isArray(refValue))` 确认类型。

### 13. Escape 按键会清除表格行选中并关闭编辑器

**问题**：在 proto-test 页面用 `page.keyboard.press('Escape')` 关闭 NSelect 下拉菜单时，Escape 同时会清除消息表格的行选中状态，导致 paired-payload-editor（`v-if="entry"`）整个卸载，后续对 `.field-item` 的定位全部超时。

**错误表现**：编辑器相关 locator 解析为 0 个元素而 click/fill 超时；失败截图中表格无选中行、无编辑器。

**解决方案**：测试中不要按 Escape 关闭下拉菜单。单选 NSelect 点击选项后菜单自动关闭；多选（filterable+tag）输入后无需关闭菜单，对后续按钮使用 `click({ force: true })` 即可。

**诊断技巧**：用临时 CDP 调试脚本在每步后打印关键元素计数（`.field-item:visible` 等），可以快速二分定位是哪一步使编辑器消失。

### 14. 新增 Page Object 的 goto() 检查清单

编写新的 Page Object 的 `goto()` 方法时，**必须**逐一检查以下项目。详见 [goto-analysis.md](shared/docs/goto-analysis.md)。

**核心要点**：
- 遮罩清理用 `isVisible()` 检测（非 `count()>0`），Escape x 2 + click 遮罩兜底
- 导航确认：无 v-show 时用 `waitForSelector`；有 v-show 时用 `sleep` + 注释
- 状态重置用 `expect().toBeVisible()` 前置断言，提供 options 参数控制重置范围
- 导航前检查页面特有元素是否已存在，存在则跳过导航

**禁止**：`count()>0` 判断遮罩、`waitForSelector({state:'attached'})` 确认 v-show 页签、`if(await btn.isVisible())` 守卫关键操作、在 goto() 中做无关业务操作

### 15. 文本相同的按钮 + `.first()` 的 DOM 顺序陷阱

**问题**：`page.locator('button:visible').filter({ hasText: '设置' }).first()` 本意点 target-service-config 的重放设置抽屉触发按钮，但顶部 header 也有同名的全局「设置」导航按钮，且 DOM 顺序在前，`.first()` 永远命中 header 那个 → 点击后导航到全局设置页（飞书通知/MCP 配置），而非打开抽屉。抽屉里的元素（如 `input[placeholder*="不限"]`）永远找不到 → 确定性 TimeoutError。

**根因**：文本歧义（多处按钮同名「设置」）+ `.first()` 按 DOM 顺序（header 在前）= 每次都点错。**这是确定性失败，不是 flaky**——多次运行失败 test 完全一致。

**错误表现**：
```
TimeoutError: locator.inputValue: Timeout 10000ms exceeded.
- waiting for locator('input[placeholder*="不限"]').first()
```
查 error-context 的 Page snapshot：若显示 `button "设置" [active]` + 飞书通知配置/MCP 服务配置，即确认点了全局设置页。

**解决方案**：文本可能在多处出现的按钮/元素，**优先 `data-testid`**，不要依赖文本 + DOM 顺序。
```typescript
// ❌ 错误：文本 + .first() 命中 header 全局设置
this.settingsButton = page.locator('button:visible').filter({ hasText: '设置' }).first();

// ✅ 正确：data-testid 精确定位（target-service-config.vue 加 data-testid="target-service-settings-btn"）
this.settingsButton = page.locator('[data-testid="target-service-settings-btn"]').first();
```

**教训**：诊断「点错目标」类失败，必查 error-context 的 Page snapshot 确认点击后实际落到哪个页面，而非只看 TimeoutError 本身。

### 16. Naive UI 下拉关闭后 menu/option 仍留 DOM（strict mode 冲突）

**问题**：Naive UI 的 `n-select` 关闭后，`.n-base-select-menu` / `.n-base-select-option` 元素**仍保留在 DOM**（不可见，但被 Playwright locator 匹配）。当多个 select 共存（如「用例选择」+「字段类型选择」）时，定位下拉选项会解析到多个元素，触发 strict mode 失败；计数类断言也会偏大（如期望 5 个选项实测 14 个）。

**错误表现**：
```
Error: locator.waitFor: Error: strict mode violation: locator('.n-base-select-menu') resolved to 2 elements
  1) ...创号带工会 (19条)工会战...   ← 用例下拉（可见）
  2) ...原始值范围枚举组合变量...     ← 字段类型下拉（残留，不可见）

Error: expect(received).toBe(expected) ... Expected: 5, Received: 14   ← 残留选项被计入
```

残留还会跨测试累积（`goto` 只切内存路由，不重建 Vue 组件），首测尤其易踩。

**解决方案**：定位下拉选项一律加 `:visible`，只匹配当前打开菜单的可见选项。
```typescript
// ❌ 错误：匹配到关闭后残留的不可见选项
const options = page.locator('.n-base-select-option');

// ✅ 正确：:visible 过滤残留
const options = page.locator('.n-base-select-option:visible');
```

Page Object 的 `selectCaseFromDropdown` 已按此修复（见 `ProtoTestPage.ts`）：直接等「可见的目标选项」出现并点击，绕过残留 menu 容器。

**诊断技巧**：strict mode 报告多个元素时，看它们的 class——带 `fade-in-scale-up-transition-enter-active` 的是刚打开的可见 menu，没有的是残留。计数类断言偏大时，优先怀疑残留选项被计入。

### 17. 改 Go 后端后必须重启 wails3 dev（.exe 锁导致旧后端仍在跑）

**问题**：`wails3 dev` 运行中修改 Go 后端代码（新增字段、改 Service 返回值等），wails3 dev 的前端 Vite 会热重载，但**后端 `rain-qa-func.exe` 被运行进程锁定，无法被重新编译覆盖**。运行中的实例仍是旧后端——E2E 跑的是旧逻辑，新字段/新接口表现为"不存在"。

**典型踩坑**：给 `VariableInfo` 加了 `available_reqs` 字段，前端 bindings 已重新生成，vue-tsc 类型检查通过，但 E2E 里该字段始终 `undefined`——因为运行的后端还是旧版返回。

**解决方案**：改 Go 后端后，**停掉 wails3 dev 和 rain-qa-func 再重启**，确保新后端编译并加载：
```powershell
Stop-Process -Name "rain-qa-func" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "wails3" -Force -ErrorAction SilentlyContinue
# 然后 wails3 dev 重启，等 CDP（约 60s）
```

**如何判断是否需要重启**：改了 `backend/` 下任何 `.go` 文件后跑 E2E，若新行为未生效，先怀疑后端没重编译。bindings（`frontend/bindings/`）的更新不代表后端已重编译——bindings 由 `wails3 generate bindings` 独立生成，与运行中的 `.exe` 无关。

**与第 3 条的区别**：第 3 条是「dev 运行时手动 generate bindings 会 rename 失败」（bindings 生成侧）；本条是「dev 运行时改 Go 后端不会重编译进运行中的 .exe」（后端运行侧）。同源都是 Windows 文件锁，但表现和应对不同。

### 18. E2E 跑完后 webview 残留 select 菜单/overlay，手动操作被拦截

**问题**：E2E 跑完后，webview 上残留 select 菜单 / overlay（naive-ui 的 `.n-base-select-menu`）。用户手动操作（如步骤动作的 action select 选「出牌/弃牌」）时，**点选项无反应、点空白处也不关闭菜单**，必须重启 `wails3 dev` 才恢复。

**根因**：E2E 的 `afterEach` 只清了 modal/drawer 遮罩（`.n-modal-mask` / `.n-drawer-mask`，见第 8 条），**没清 select 菜单**。naive-ui select 菜单残留后其 overlay 拦截所有点击，表现为「选选项没反应、菜单关不掉」。与第 8 条同源（overlay 拦截），但作用对象是 select 菜单而非 modal。

**诊断特征**：跑完 E2E 后手动操作 GUI 出现「点选项无反应 + 点空白不关菜单」→ 基本是 select 菜单或 modal mask 残留；**重启 dev 干净后不复现 = 确认是残留**（非代码 bug）。

**解决方案**：fixture 的 `afterEach` 先关 select 菜单（Escape + 点 body 空白），再清 modal mask：
```typescript
// frontend/e2e/shared/fixtures/index.ts
test.afterEach(async ({page}) => {
    // Escape + 点 body 空白关闭残留的 select 菜单（naive-ui select 点外部/Escape 关闭）
    await page.keyboard.press('Escape');
    await page.locator('body').click({position: {x: 1, y: 1}}).catch(() => {});
    await page.waitForTimeout(150);
    // 再清 modal/drawer 遮罩（第 8 条）
    for (let i = 0; i < 3; i++) {
        const masks = page.locator('.n-modal-mask:visible, .n-drawer-mask:visible');
        if (await masks.count() === 0) break;
        await page.keyboard.press('Escape');
        await page.waitForTimeout(200);
    }
    // 孤儿 select 菜单：naive-ui unmounted 后 Follower 容器（.v-binder-follower-content）残留，
    // Escape/点 body 关不掉、display:none 会被 Follower 重新覆盖。只能 remove 含可见菜单的 Follower。
    // ⚠️ remove 偶发误伤还在用的 select（致后续 test flaky），需配合 playwright.config retries 兜底
    await page.evaluate(() => {
        document.querySelectorAll('.v-binder-follower-content').forEach((f) => {
            const menu = f.querySelector('.n-base-select-menu');
            if (menu && (menu as HTMLElement).offsetParent !== null) {
                f.remove();
            }
        });
    }).catch(() => {});
});
```

