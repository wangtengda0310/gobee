# proto-test E2E 测试

协议重放模块的端到端测试，对应后端包 `pkg/proto-test/`、前端页面 `/ProtoTest`。

## 文件索引

### 拆分自主套件（proto-test.spec.ts）

| 文件 | 覆盖范围 |
|------|----------|
| `proto-test-base.spec.ts` | 页面结构、页签、目标服务输入、录制按钮 |
| `proto-test-multi-select.spec.ts` | 多选模式、测试用例管理 |
| `proto-test-components.spec.ts` | 重放控制、Range/Enum/Combo 输入组件、卡片编辑器、页面布局 |
| `proto-test-message-append.spec.ts` | 重放消息追加、测试用例执行重放 |
| `proto-test-payload.spec.ts` | 重发 Payload 修改同步、多次重发行为验证、自动路由迭代发送 |
| `proto-test-table-ux.spec.ts` | 表格列变体与 Req 过滤、测试用例描述编辑、步骤顺序与拖放 |

### 独立套件

| 文件 | 覆盖范围 |
|------|----------|
| `proto-test-replay.spec.ts` | 完整重放流程 |
| `proto-test-replay-simple.spec.ts` | 简化重放 |
| `proto-test-replay-correct.spec.ts` | 正确重放验证 |
| `proto-test-replay-result.spec.ts` | 重放结果页签（双事件通道） |
| `proto-test-replay-debug.spec.ts` | 重放调试 |
| `proto-test-retry-bug.spec.ts` | 重发次数 bug 回归 |
| `proto-test-save-case.spec.ts` | 配对消息索引错位回归 |
| `proto-test-account-range.spec.ts` | 账号序号范围迭代 |
| `proto-test-debug-tabs.spec.ts` | 页签切换调试 |
| `proto-test-variable.spec.ts` | 动态变量提取（变量选项始终可用） |
| `debug-replay.spec.ts` | 临时调试 |

## 运行方式

```bash
# 全部
npx playwright test proto-test/

# 按功能域（拆分后的主套件）
npx playwright test proto-test/proto-test-base.spec.ts
npx playwright test proto-test/proto-test-table-ux.spec.ts

# 指定用例
npx playwright test proto-test/ -g "描述编辑"
```

## 关键测试数据

- `cases/proto_cases/创号.json` — 描述编辑和排序测试的 fixture
- `cases/proto_cases/添加黄金.json` — 重放和重发测试的 fixture

## Page Object

`shared/pages/ProtoTestPage.ts` — 包含发包改包/测试用例/重放结果三个页签的全部操作方法。
