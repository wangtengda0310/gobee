---
name: browser-env-debug
description: 浏览器环境复杂 Bug 排查技能 — 当遇到涉及浏览器 API、框架响应式系统干扰第三方库、Vite 缓存、shim 不完整等问题时触发。提供系统化排查流程、常见陷阱清单和自动化调试脚本模板。确保在处理任何涉及浏览器环境的第三方库集成问题、Vue/React 响应式系统与库冲突、Vite/Webpack 构建差异、WebRTC/WebSocket/IndexedDB 等 API 问题时使用此 skill，即使用户没有明确提到"浏览器环境调试"。每次使用后检查是否有新经验需要沉淀到 skill 中。
---

# 浏览器环境复杂 Bug 排查

**触发条件**（满足任一即触发）：
- Bug 涉及浏览器专有 API（WebRTC、Canvas、IndexedDB、Service Worker 等）
- 第三方浏览器库运行异常（WebTorrent、PDF.js、Monaco Editor 等）
- 框架响应式系统（Vue reactive、React state）与库内部状态冲突
- Vite/Webpack 打包缓存导致修改不生效
- 仅在浏览器中复现、Node 环境无法复现的 Bug
- 用户报告"浏览器控制台有错误但 AI 无法直接查看"

**不触发**：纯后端 Bug、纯逻辑 Bug、不涉及浏览器环境的常规开发

---

## 排查流程：隔离 → 对比 → 定向修复 → 验证

### 第一步：隔离复现

**目标**：排除框架干扰，确认是库本身的问题还是环境集成问题。

在浏览器上下文中直接测试库的核心 API，不经过任何框架包装。

```javascript
// 使用 Playwright page.evaluate 进行隔离测试
const results = await page.evaluate(async () => {
    const logs = []
    function log(level, ...args) {
        logs.push({ level, text: args.join(' ') })
    }
    try {
        // 直接导入并使用库，不经过 Vue/React
        const { default: Lib } = await import('/node_modules/.vite/deps/target-lib.js')
        const instance = new Lib()
        log('info', 'Instance created, constructor:', instance.constructor.name)
        // ... 测试核心操作
        log('success', 'Test passed')
    } catch (err) {
        log('error', err.message)
    }
    return logs
})
```

**关键点**：
- 使用 `page.evaluate` 在浏览器真实环境中执行
- 直接 `import` 库，绕过框架包装
- 收集所有日志返回给 AI 分析

**如果隔离测试通过** → 问题在框架集成，进入第二步
**如果隔离测试失败** → 问题在库本身或 shim，进入陷阱清单排查

### 第二步：对比分析

**目标**：找出框架哪个环节干扰了库的正常运行。

对比以下场景：
1. `new Lib()` — 原始实例
2. `reactive(new Lib())` — Vue 包装后
3. `ref(new Lib())` — ref 包装后

检查关键属性是否被改变：
```javascript
const raw = new Lib()
const wrapped = reactive(new Lib())

// 检查数组属性是否变成 Proxy
log('info', 'raw.array:', raw.someArray.constructor.name)  // 应为 "Array"
log('info', 'wrapped.array:', wrapped.someArray.constructor.name)  // 可能是 "Proxy"
```

### 第三步：定向修复

根据干扰源选择修复策略（见陷阱清单）。

### 第四步：回归验证

用 Playwright 自动验证修复是否生效：
```javascript
// 在应用页面上下文中执行修复后的代码
const results = await page.evaluate(async () => {
    // ... 使用修复后的代码进行完整测试
})
// 检查 results 中是否有 error 级别日志
```

---

## 常见陷阱清单

### 陷阱 1：Vue reactive Proxy 包装第三方库实例

**症状**：`client.torrents === Proxy(Array)`、库内部 `indexOf`/`push` 行为异常、`this !== this` 引用不一致

**原因**：Vue 的 `ref()` / `reactive()` 会递归地把对象所有属性包装成 Proxy，包括库内部的数组、Map 等数据结构

**修复**：
```typescript
import { markRaw, shallowRef } from 'vue'

// 方案 A：markRaw（推荐）
client.value = markRaw(new ThirdPartyLib())

// 方案 B：shallowRef（不递归包装）
const client = shallowRef<ThirdPartyLib | null>(null)
client.value = new ThirdPartyLib()
```

**识别方法**：在控制台检查 `instance.someArray.constructor.name`，如果为 `"Proxy"` 则命中此陷阱

### 陷阱 2：Vite 预构建缓存

**症状**：修改了 `node_modules` 中的文件，但浏览器中运行的代码没有变化。浏览器控制台错误指向原始行号而非修改后的代码。

**原因**：Vite dev server 使用 `node_modules/.vite/deps/` 中的预构建缓存，不会自动检测 `node_modules` 的变化

**修复**：
```bash
rm -rf node_modules/.vite
# 重启 vite dev server
```

**识别方法**：在浏览器控制台执行 `import('/node_modules/.vite/deps/target-lib.js')` 检查是否包含修改

### 陷阱 3：Browser shim 不完整

**症状**：`TypeError: this.dht.once is not a function`、`Cannot read properties of undefined` 等运行时错误

**原因**：Vite 配置中用 alias 替换 Node.js 专有模块的 shim 文件不完整，缺少库内部调用的方法

**修复**：对照原始库的 `package.json` 中的 `browser` 字段，逐个检查被替换模块的 API 使用情况：
```bash
# 查找库内部调用了哪些 shim 的方法
grep -r "dht\." node_modules/target-lib/ | grep -v node_modules/target-lib/node_modules
```

### 陷阱 4：事件循环时序

**症状**：`queueMicrotask`、`Promise.then`、`setTimeout` 之间的时序问题导致状态不一致

**原因**：库内部使用 `queueMicrotask` 调度回调，而 `await` 会暂停当前 async 函数，让 microtask 提前执行

**修复**：理解库的异步时序，避免在关键操作之间插入 `await`

### 陷阱 5：ESM/CJS 模块兼容

**症状**：`Class constructor X cannot be invoked without 'new'`、`default` 导出为空对象

**原因**：Vite 预构建时 CJS → ESM 转换可能引入兼容性问题

**修复**：在 `vite.config.ts` 中配置 `optimizeDeps.include` 或使用 `?url` 导入

---

## 调试脚本模板

> 详见 [templates/playwright-isolate.mjs](templates/playwright-isolate.mjs)

### 快速使用

```bash
# 1. 确保 dev server 正在运行
# 2. 复制模板并修改 target-lib 和测试逻辑
# 3. 运行
node debug-isolate.mjs
```

### page.evaluate 内联测试模式

当不想创建独立文件时，直接在 AI 会话中用 `page.evaluate` 执行：

```javascript
// 在 Playwright MCP 中使用
const results = await page.evaluate(async () => {
    const { default: Lib } = await import('/node_modules/.vite/deps/target-lib.js')
    // ... 测试逻辑
    return testResults
})
```

---

## 案例参考

### WebTorrent + Vue Proxy（2026-04）

**问题**：WebTorrent 做种时报 `Cannot add duplicate torrent` 和 `torrent is destroyed`

**排查过程**：
1. 用户报告浏览器控制台错误 → AI 修改 shim 修复首次错误
2. shim 修复后仍有错误 → AI 添加调试日志但 Vite 缓存导致不生效
3. 用 Playwright + `page.evaluate` 隔离测试 → 纯 webtorrent 环境测试通过
4. 对比发现 `client.torrents === Proxy(Array)` → 确认 Vue Proxy 是根因
5. 使用 `markRaw(new WebTorrent())` 修复

**关键教训**：
- Vue `ref()` 会递归包装对象，对第三方库实例必须用 `markRaw`
- Vite 预构建缓存需要手动清除
- `page.evaluate` 隔离测试是定位环境干扰的最高效方法

---

## 自迭代机制

本 skill 设计为可进化的知识库。每次使用后，AI 应检查是否有新的经验需要沉淀。

### 触发迭代的时机

1. **排查过程发现新的陷阱模式** — 不在现有 5 个陷阱清单中
2. **某个陷阱的修复方案有更优解** — 比如发现 `shallowRef` 在某些场景比 `markRaw` 更合适
3. **隔离测试方法有改进** — 比如发现比 `page.evaluate` 更好的隔离方式
4. **新的框架/工具链陷阱** — 比如 Next.js SSR、Nuxt SSR、Electron 等环境的新问题
5. **案例复盘有新洞察** — 真实 bug 排查完成后，总结出可复用的经验

### 迭代方式

#### 添加新陷阱

在"常见陷阱清单"中按此格式追加：

```markdown
### 陷阱 N：标题

**症状**：具体表现

**原因**：根因解释

**修复**：解决方案代码

**识别方法**：如何快速确认命中此陷阱
```

#### 补充案例

在"案例参考"中追加新的真实案例，包含：
- 问题描述（一句话）
- 排查过程（关键步骤）
- 关键教训（1-3 条可复用的经验）

#### 拓展参考资料

当陷阱涉及特定框架/工具的深度知识时，创建独立参考文件避免 SKILL.md 膨胀：

```
browser-env-debug/
├── SKILL.md                          # 核心流程和陷阱清单（<500行）
├── templates/playwright-isolate.mjs  # 隔离测试脚本模板
└── references/                       # 按需加载的参考资料
    ├── vue-traps.md                  # Vue 响应式系统深度陷阱
    ├── react-traps.md                # React 生命周期和闭包陷阱
    ├── vite-traps.md                 # Vite 构建/缓存/Worker 陷阱
    └── electron-traps.md             # Electron 环境特有陷阱
```

当某个陷阱的解释超过 30 行，应拆分到 `references/` 目录并在 SKILL.md 中保留简短摘要和链接。

### 迭代检查清单

每次使用本 skill 完成 bug 排查后，AI 应自问：

1. 这次排查是否暴露了新的陷阱模式？（不在现有清单中）
2. 是否使用了现有陷阱清单中未覆盖的修复方法？
3. 排查过程中是否有值得记录的"顿悟时刻"？
4. 是否涉及了新的浏览器 API 或框架特性？

如果任一问题答案为"是"，主动向用户提议更新 skill。

### 经验沉淀（使用 self-improvement）

每次使用后评估是否产生了值得记录的知识：
- 发现了新的浏览器 API 兼容性问题？
- Vue/React 响应式系统与第三方库的新冲突模式？
- Vite/Webpack 构建差异的新案例？
- WebRTC/WebSocket/IndexedDB 的新的 gotcha？

如果是，记录到 `.learnings/LEARNINGS.md`，category 为 `browser_env`。
