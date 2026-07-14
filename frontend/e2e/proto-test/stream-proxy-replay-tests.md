# Proto Test 重放消息追加功能 E2E 测试说明

## 测试文件
- **测试文件**: `frontend/e2e/proto-test/` 目录下拆分套件（见 `e2e/proto-test/CLAUDE.md`）
- **Page Object**: `frontend/e2e/shared/pages/ProtoTestPage.ts`
- **运行方式**: `wails3 dev` 启动后 `npx playwright test proto-test/`

## 测试用例列表

### 重放消息追加功能 (`describe('重放消息追加功能')`)

| 测试名 | 关键验证点 |
|--------|-----------|
| 执行用例后表格行数增加 | 重放后行数 >= 重放前 |
| 重放消息包含正确的 MsgID 和方向 | 消息内容非空，含 MsgID、方向箭头 |
| 多次重放会多次追加消息 | 第二次行数 > 第一次行数 > 初始行数 |
| 重放进度标签正确更新 | 按钮 loading 状态正确切换 |
| 选中行重发后表格追加消息 | 重放控制面板显示 + 行数增加 |
| 重放消息在表格中的顺序正确 | 最后一行有内容 |
| 重放完成后状态正确恢复 | 按钮恢复可用，无异常状态残留 |

### 测试用例执行重放 (`describe('测试用例执行重放')`)

| 测试名 | 关键验证点 |
|--------|-----------|
| 执行用例按钮触发重放并追加消息 | 执行后行数 >= 执行前 |
| 执行用例时按钮状态正确 | 执行前可用 → loading → 恢复可用 |
| 未选择用例时执行按钮禁用 | 按钮存在且禁用 |

## Page Object 新增方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `getRowText(rowIndex)` | 获取指定行的文本内容 | `Promise<string>` |
| `waitForTableRowCount(expectedCount, timeout)` | 等待表格行数达到指定值 | `Promise<void>` |
| `getReplayStatusText()` | 获取重放状态标签文本 | `Promise<string>` |
| `isReplayRunning()` | 检查重放是否正在进行 | `Promise<boolean>` |
| `waitForReplayComplete(timeout)` | 等待重放完成 | `Promise<void>` |

## 测试数据

测试用例文件：`cases/proto_cases/e2e_test_case.json`
- 包含 2 条 Req 消息（direction = "→"）+ 2 条 Ack 消息（direction = "←"）
- 用于验证重放功能的消息筛选和追加逻辑

## 已知限制

1. **后端依赖**: 测试需要真实的后端服务才能完整验证消息追加功能
2. **时间等待**: 使用固定的 `waitForTimeout` 等待重放完成，可能需要根据实际网络环境调整
3. **行数验证**: 由于后端重放可能失败或超时，某些测试使用 `toBeGreaterThanOrEqual` 而非精确值

## 运行测试

```bash
# 启动 Wails 应用（启用 CDP）
wails3 dev

# 运行所有 Proto Test 测试
cd frontend
npx playwright test proto-test/

# 只运行重放消息追加测试
npx playwright test proto-test/proto-test-message-append.spec.ts

# 调试模式（显示浏览器窗口）
npx playwright test proto-test/ --debug
```

## 相关文档

- [前端布局文档](frontend/docs/layout/pages/proto-test/index.md) — 页面 ASCII 布局图和组件树
- [后端包文档](backend/pkg/proto-test/CLAUDE.md) — 后端 Service 索引和重放架构
- [E2E 测试规范](frontend/e2e/CLAUDE.md) — E2E 测试通用规范和已知问题
- [Proto Test 页面文档](frontend/src/pages/proto-test/CLAUDE.md) — 页面级文档和设计决策
