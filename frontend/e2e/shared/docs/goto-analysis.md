---
name: streamproxy-e2e-goto-method-analysis
description: ProtoTestPage.goto() 方法的完整分析——设计原理、演进历史、已知问题、设计缺陷和改进建议
metadata:
  type: reference
  related: [[proto-test-e2e-test-quality]] [[Wails createMemoryHistory 内存路由导致 E2E goto() 无法通过 URL 导航]]
---

# ProtoTestPage.goto() 方法完整分析

## 概述

`ProtoTestPage.goto()` 是 proto-test E2E 测试套件中最关键的初始化方法。每次测试 `beforeEach` 都会调用它，负责**导航到目标页面 + 重置所有可变状态**。由于 Wails 使用 `createMemoryHistory`（内存路由），无法通过 URL 导航，goto() 必须模拟用户操作（点击菜单按钮）来触发路由切换。

**当前实现**：`frontend/e2e/shared/pages/ProtoTestPage.ts:143-183`

## 设计原理

### 为什么不用 URL 导航？

Wails 应用使用 `createMemoryHistory()` 而非 `createWebHistory()`，路由状态存储在内存对象中，**浏览器地址栏 URL 不随页面切换而变化**。因此 `page.goto('/StreamProxy')` 无效，必须通过点击菜单按钮触发内存路由。

### goto() 必须完成的职责

| 步骤 | 操作 | 目的 |
|------|------|------|
| 1 | 清理 `.n-modal-mask` 和 `.n-drawer-mask` 遮罩 | 防止遗留遮罩拦截后续点击 |
| 2 | 点击 `#layout-header button:has-text("Proto测试")` | 导航到 Proto 测试页面 |
| 3 | 清理"取消多选"按钮（如果可见） | 防止遗留的多选模式干扰 |
| 4 | 切换到"发包改包"页签 | 确保起始页签一致 |
| 5 | 切换到"重放结果"页签 → 点击"清空" | 清除上一次测试遗留的重放结果 |
| 6 | 切换回"发包改包"页签 | 最终回到默认页签 |

## 演进历史

| 阶段 | 改动 |
|------|------|
| 初始 | 仅点击菜单按钮 |
| 遮罩问题 | + modal/drawer mask 清理 + 二次 Escape 确认 |
| 状态累积 | + 多选模式清理 + 重放结果清空 |
| v-show 不可见 | `click()` → `click({ force: true })` |
| count() 误判 | `count()>0` → `isVisible()` 检查 |

## 已知问题

详见 `frontend/e2e/CLAUDE.md` 问题 6/8/10/11/12。核心问题：

- **遮罩拦截**（问题 8）：前一个测试的模态框未关闭，遮罩拦截 goto() 中的 click
- **v-show 不可见**（问题 10）：三个页签使用 v-show，DOM 元素始终存在但 `display: none`
- **状态累积**（问题 11）：串行测试，前一个测试的重放结果影响后续断言
- **Vue watch 陷阱**（问题 6）：只调用 `loadTestCase()` 未将返回数据设置到响应式状态

## 当前实现的设计缺陷

| 缺陷 | 问题 | 影响 |
|------|------|------|
| 过度重置 | goto() 清空重放结果同时丢失用例页签数据 | 每个测试需重新加载用例（~4-5秒） |
| 遮罩不可靠 | Escape 不保证关闭所有遮罩，静默继续 | 后续操作可能因遮罩拦截超时 |
| clearBtn 脆弱 | 依赖"清空"文本匹配，v-show 下可能不可见 | 渲染未完成时跳过清空 |
| 多选清理 | 依赖"取消多选"按钮文本定位 | 文本变化导致选择器失效 |

## 本会话中 goto() 导致的实际问题

| 问题 | 表现 | 根因 | 修复 |
|------|------|------|------|
| A | goto() 后表格行数为 0 | 切换页签清空数据，Vue 状态未保留 | 每个测试手动重新加载用例 |
| B | 选择器匹配到隐藏页签元素 | v-show 下所有 DOM 同时存在，`.first()` 取到隐藏实例 | `dispatchEvent('click')` + `isVisible()` 过滤 |
| C | 清空按钮 click 超时 | `count()>0` 但元素 hidden，`click()` 要求 visible | `isVisible()` + `click({ force: true })` |

## 改进建议

- ✅ **已实现**：`GotoOptions.skipClearResults` 控制重置深度；逐 mask 检查 + 三级兜底
- **中期**：将"加载测试用例"提取为 Page Object 方法统一处理
- **长期**：考虑 `v-if` 替代 `v-show`；通过 `page.evaluate` 直接操作 Vue 状态

## 反模式总结

通过 2 个 agent 的代码审核和本会话的实践经验，总结以下 goto() 编写中的反模式：

| 反模式 | 为什么错误 | 正确方式 |
|--------|-----------|---------|
| `count() > 0` 判断元素存在 | 对 hidden 元素也返回 true | `isVisible()` |
| `if (await btn.isVisible())` 守卫关键操作 | 不可见时静默跳过，测试"假通过" | `expect(btn).toBeVisible()` 前置断言 |
| `waitForSelector({ state: 'attached' })` 在 v-show 环境 | 元素始终 attached，瞬间返回，无法确认导航 | sleep + 注释说明 |
| 遮罩清理只 Escape 一次 | 嵌套 modal/drawer 需要多次 Escape | Escape x 2 + click 遮罩兜底 |
| goto() 中硬编码所有重置逻辑 | 每个测试都承受全部开销 | options 参数控制重置范围 |
| 遮罩清理失败静默继续 | 后续操作可能因遮罩拦截而失败，难以定位 | console.warn 或 throw
8. **为每个测试文件创建独立的页面实例**：隔离状态，但可能增加 CDP 连接开销

## 相关文档

- `frontend/e2e/CLAUDE.md` → 问题 8/10/11/12
- `frontend/e2e/shared/pages/ProtoTestPage.ts:143` → 当前实现
- `frontend/e2e/proto-test/proto-test-base.spec.ts` → 使用方式
- [[proto-test-e2e-test-quality]] — E2E 测试质量问题
- [[Wails createMemoryHistory 内存路由导致 E2E goto() 无法通过 URL 导航]] — forgetful 记忆 #81
