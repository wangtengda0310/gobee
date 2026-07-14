# Proto Test 页面

协议录制/重放页面，含"发包改包"、"测试用例"和"重放结果"三个页签。

## 1. 前端布局文档

- [布局总览](frontend/docs/layout/pages/proto-test/index.md) — ASCII 布局图、组件树、时序图
- [数据流](frontend/docs/layout/pages/proto-test/data-flow.md) — 事件路由、录制/拦截流程

## 2. requirement.ts 规范

### 组件列表

| 组件 | requirement.ts | 职责 | Wails Service |
|------|---------------|------|---------------|
| **stream-proxy/packet-tab.vue** | shared/protocol-content.requirement.ts | **发包改包页签**（拦截模式切换、拦截队列管理） | RecordControlService, ReplayControlService, TestCaseService, RecordFileService |
| **cases/testcase-tab.vue** | - | **测试用例页签**（独立状态管理） | TestCaseService, ReplayControlService, RecordFileService |
| **replay-result/replay-result-tab.vue** | - | **重放结果页签**（独立状态管理） | - |
| shared/case-selector.vue | shared/case-selector.requirement.ts | 用例选择/新建 | TestCaseService |
| shared/message-table.vue | shared/message-table.requirement.ts | 消息列表展示，拦截消息橙色标记 | RecordFileService |
| shared/paired-payload-editor.vue | shared/paired-payload-editor.requirement.ts | Payload 编辑器容器 | RecordFileService |
| shared/replay-control.vue | shared/replay-control.requirement.ts | 重发控制 + Ntf 显示 | ReplayControlService |
| **shared/target-service-config.vue** | **shared/target-service-config.requirement.ts** | 目标服务配置 + 监听端口配置 + 注入 unity 服务器列表 | ProtoTestConfigService, RecordControlService, **ServerConfigService** |
| replay-result/replay-result-selector.vue | replay-result/replay-result-service.requirement.ts | 重放结果选择器 | ReplayResultService |

### requirement.ts 模板

```typescript
// 详见 frontend/src/pages/CLAUDE.md 的"组件开发规范"章节
```

## 3. 设计决策

### 注入 unity 服务器列表（2026-06-17）

**决策**：在 `target-service-config.vue` 的设置抽屉「监听端口配置」下方新增「注入 unity 服务器列表」按钮，一键将本机作为 Unity 服务器写入策划配表并触发客户端导出。

**实现**：
1. 后端新增 `backend/pkg/proto-test/server-config` 包，定义 `ServerXlsxConfig` 结构体，实现 `InjectUnityServer(ServerXlsxConfig)` 和 `ExportClientConfig(string)`。
2. `InjectUnityServer` 打开统一策划配表目录下 `excel/服务器配置表.xlsx`，按字段名找列，写入/更新 `Id=999` 的行；`IpPort` 为空时自动使用本机 IP 和 HTTP 监听端口构造 `http://{ip}:{port}/authlogin`。
3. `ExportClientConfig` 在策划配表目录下执行 `export_client.bat`。
4. 新增 `settings.ExcelConfigService` 管理统一策划配表目录（`.rain-qa-func.json` 的 `excel_config` section），`ServerConfigService` 在 `ExcelDir` 为空时自动从该配置读取。
5. 前端 `target-service-config.requirement.ts` 封装 `ServerConfigService.InjectUnityServer` 和 `ServerConfigService.ExportClientConfig` 调用。

**相关代码**：
- 后端实现：`[InjectUnityServer](backend/pkg/proto-test/server-config/server_config.go:89)`、`[ExportClientConfig](backend/pkg/proto-test/server-config/server_config.go:234)`
- Wails 服务：`[ServerConfigService](backend/pkg/proto-test/server-config/wails.go:12)`
- 统一配置：`[ExcelConfigService](backend/pkg/settings/excel_config.go:14)`
- 前端调用：`[target-service-config.vue handleInjectUnityServer](frontend/src/pages/proto-test/shared/target-service-config.vue:236)`

### 测试用例自动加载

**决策（2026-06-04）**：测试用例页签组件挂载时自动加载用例列表。

**实现**：`testcase-tab.vue` 的 `onMounted` 中调用 `testCaseService.loadCaseList()`

### 两个独立的重放/重发操作

| 操作 | 触发位置 | 发送对象 | 次数 | 底部面板是否显示 |
|------|---------|---------|------|----------------|
| 开始重放 / 执行用例 | 顶部按钮行 | 表格中所有 Req | 1 次 | 否 |
| 重发 | 底部重放控制面板（选中行后显示） | 当前选中行的 Req | N 次 | 是，并显示 Ntf 消息 |

### 重发自动路由迭代发送（2026-06-11）

**背景 Bug**：卡片模式将字段配置为枚举（如 GmCommandReq 的 content 配置 3 个 `//AddItem` 命令）后点"重发"，原实现把逗号拼接的枚举串当作单条消息的字段值发送。

**决策**：
1. `field-item.vue` 的 `getActiveValue()` 一律返回字段原始值（`props.value`）——卡片模式不直接编辑字段值，range/enum/combo 是**迭代配置**而非字段值，仅通过 `getFourState()` 提交给后端展开
2. `packet-tab.vue` / `testcase-tab.vue` 的"重发"检测到迭代配置（任一字段 `input_type !== 'original'`）时自动路由到 `sendIterativeMessages`；收集顺序：卡片编辑器活跃状态优先，回退用例中持久化的 `field_values`
3. 拦截放行（已拦截消息的重发）必须精确发送当前一条，不参与迭代

**回归测试**：`proto-test-payload.spec.ts` 的"重发自动路由迭代发送"套件；后端展开逻辑见 `pkg/proto-test/params/iteration_test.go`（注意 `IterationConfig.Type` 与前端 `input_type` 一致，为 `combo` 而非 `compose`）。

### 消息表格按页签变体展示（2026-06-11）

`message-table.vue` 增加 `variant` prop 区分页签：

| variant | 使用页签 | 列差异 |
|---------|---------|--------|
| `packet` | 发包改包 | 完整列；"请求(Req)"标题旁有 [过滤] 按钮（只显示带 Req 数据的行） |
| `testcase` | 测试用例 | 隐藏 账号/时间/SeqID/结果 列；"响应(Ack)"列替换为"描述"列（消息 `descript` 字段） |
| 未传 | 重放结果 | 完整列，无过滤按钮 |

行点击高亮通过 `row-class-name`（`selected-row`）+ scoped CSS 覆盖 td 背景实现——tr 的 inline background 会被 naive-ui td 背景遮盖，不能用 rowProps style。

### 实时拦截改包模式

**决策（2026-06-09）**：为"发包改包"页签增加实时拦截并修改数据包的能力。

详见 [数据流文档](frontend/docs/layout/pages/proto-test/data-flow.md) 的"拦截模式流程"章节。

**2026-06-10 调查**：当前「重发」未在原连接放行，待改造为批量放行。完整结论见 [intercept-batch-release-investigation.md](../../../../backend/pkg/proto-test/docs/intercept-batch-release-investigation.md)；待办见 `docs/TODO.md`。

### 覆盖与追加语义分离（2026-06-17）

**背景 Bug**：`packet-tab.vue` 的"保存到用例"对话框按钮文案为"追加"，但实际调用 `TestCaseService.SaveTestCase`，后者会覆盖已存在用例，导致 `testcase-tab.vue` 中维护的用例数据被意外清空。

**决策**：
1. 后端 `TestCaseService` 拆分出 `AppendTestCase`（追加）与 `SaveTestCase`（覆盖）两个方法。
2. `SaveTestCase` 继续供 `testcase-tab.vue` 的"新增模块"/"删除消息"/"保存顺序"使用，保证覆盖语义正确。
3. `packet-tab.vue` 的"保存到用例"统一改调 `appendTestCase`，实现真正的追加。
4. `case-selector.requirement.ts` 同时暴露 `saveTestCase` 与 `appendTestCase`，并在注释中明确各自语义，防止误用。

**相关代码**：
- 后端方法：`[TestCaseService.SaveTestCase](backend/pkg/proto-test/wails_test_case.go:138)`、`[TestCaseService.AppendTestCase](backend/pkg/proto-test/wails_test_case.go:165)`
- 底层实现：`[cases.SaveTestCaseToFile](backend/pkg/proto-test/cases/testcase.go:58)`、`[cases.AppendTestCaseToFile](backend/pkg/proto-test/cases/testcase.go:66)`
- 前端封装：`[case-selector.requirement.ts](frontend/src/pages/proto-test/shared/case-selector.requirement.ts:21)`
- 前端调用：`[packet-tab.vue confirmSaveCase](frontend/src/pages/proto-test/stream-proxy/packet-tab.vue:483)`

**回归测试**：`proto-test-save-case.spec.ts` 的"追加到已存在用例不应覆盖原数据"用例。

### 启动自动监听与录制按钮职责拆分（2026-06-17）

**背景**：原先「开始录制」按钮同时负责"启动本地端口监听"和"开始协议录制"两步操作。用户需要应用启动后就自动监听端口，录制按钮只控制录制启停。

**决策**：
1. 后端 `RecordWorker` 拆分为 `StartListen`（启动 TCP/HTTP 监听）和 `StartRecord`（开始录制）两个阶段；`StopRecord` 只停止录制，保留监听。
2. 应用启动时（`cmd/rain-qa-func/wails.go`）自动调用 `StartListen`，读取 `ProtoTestConfigService` 中的监听端口和目标地址。
3. 前端「开始录制」按钮只调用 `StartRecord(filterMode)`；「停止录制」按钮只调用 `StopRecord()`。
4. 页面顶部「目标服务」输入框作为目标 TCP/HTTP 地址的唯一来源；修改后会自动保存到 `ProtoTestConfig` 并重启监听。
5. `target-service-config.vue` 设置抽屉仅保留 TCP/HTTP 监听端口配置。
6. 监听配置持久化到 `.rain-qa-func.json` 的 `proto_test` section，由 `ProtoTestConfigService` 管理。

**影响**：
- `protocol-content.requirement.ts` 新增 `ProtoTestConfigService` 接口和 `createWailsProtoTestConfigService()`。
- `packet-tab.vue` 开始录制前会先检查当前状态是否为 `listening`/`running`，未监听时给出提示。
- `RecordControlService` 暴露 `StartListen` / `StartRecord` / `StopListen` / `StopRecord` / `Stop` 方法。

### 目标服务地址保存的异步 props 陷阱（2026-06-17）

**背景 Bug**：在 `target-service-config.vue` 的 `handleServerAddrChange` / `handleHttpAddrChange` 中，先 `emit('update:serverAddr', value)` 再立即调用 `saveProtoTestSettings()`。`saveProtoTestSettings()` 内部从 `props.serverAddr` 读取值，但 Vue 子组件 props 在 emit 后不会同步更新，导致实际保存的是上一次（可能不完整的）旧值。表现为输入 `10.254.114.174:20144`，后端却按 `10.254.114.17:20144` 连接。

**决策**：
1. 将相关 change handler 改为 `async`，在 emit 后 `await nextTick()`，确保子组件 props 与父组件状态同步后再读取并保存。
2. 保持"输入即保存、保存后重启监听"的交互不变，仅修复时序问题。

**相关代码**：`[target-service-config.vue](frontend/src/pages/proto-test/shared/target-service-config.vue:177)`

### 用例保存仅保留 Req，多选模式禁止勾选非 Req 行（2026-06-17）

**背景 Bug**：用户多选保存到用例时勾选了 Ntf/Ack 单行，点击"保存到用例"后创建了一个空用例文件。由于早期没有提示，用户误以为保存失败。

**根因**：
- 用例文件格式只保留客户端请求（Req），`buildTestCaseFile` 会过滤掉 Ack/Ntf。
- `message-table.vue` 多选模式下所有行都可勾选，包括没有 `req` 的 Ntf/Ack 单行。
- `packet-tab.vue` 的 `handleSaveToCase` 虽然只取 `p.req`，但当选中行全部没有 Req 时，没有给用户任何提示，直接打开保存对话框并传空消息列表。

**决策**：
1. 多选模式下，非 Req 行的 checkbox 直接禁用，并 hover 提示"仅可选择 Req 消息"。
2. 即使绕过禁用（如旧状态残留），`handleSaveToCase` 在提取到 0 条 Req 消息时立即提示："选中的行不包含 Req 消息，无法保存到用例（用例仅保存客户端请求）"。
3. 右键"增加到用例"保持既有逻辑：无 Req 行显示"增加到用例（无 Req）"并禁用。

**相关代码**：
- 多选 checkbox 禁用：`[message-table.vue](frontend/src/pages/proto-test/shared/message-table.vue:145)`
- 保存前提示：`[packet-tab.vue handleSaveToCase](frontend/src/pages/proto-test/stream-proxy/packet-tab.vue:471)`

### 有字段配置的 Req 默认卡片模式（2026-07-01）

**背景**：选中 Req 后 payload 编辑器统一默认 JSON 模式。带 variable/range/enum/combo 字段配置的 Req 也在 JSON 模式，用户看不到已配置的迭代/变量信息；保存字段配置后 entry 重新加载，卡片模式也会跳回 JSON。

**决策**：`paired-payload-editor.vue` 在 entry 变化（含保存后重新加载）时，根据 `entry.req.field_values` 是否含**非 original** 类型字段决定默认编辑模式——有则卡片模式，无则 JSON 模式。

**相关代码**：
- 默认模式判定：`[paired-payload-editor.vue reqHasFieldConfig](frontend/src/pages/proto-test/shared/paired-payload-editor.vue:226)`、watch 回调中的 `reqEditMode.value` 赋值
- 回归测试：`stream-proxy-variable.spec.ts` 的「带字段配置的 Req 默认卡片模式 (G4)」套件

### 变量按 Req 过滤（AvailableReqs，2026-07-01）

**背景**：所有 Req 在卡片模式配置变量字段时，变量下拉都显示全部变量（cityId/roomCreator/roomID/openid...），用户要在无关变量里找需要的，且容易误选（如给 RoomLookOnReq 选了 cityId，运行时提取失败）。

**决策**：`VariableDef` 增加 `AvailableReqs []string`（存 proto 消息名 `msg_name`），限制变量只对指定 Req 可见；`nil`/空表示全可用（账号级变量 `openid`）。`VariableInfo` 同步携带 `available_reqs` 下发前端。`variable-select.vue` 接收当前 Req 的 `msg_name`（经 `req-card-editor → field-item → variable-select` 透传），按 `available_reqs` 过滤下拉选项。

注册映射：`cityId → TeamSelectGuildCityReq`、`roomCreator/roomID → RoomLookOnReq`、`openid → 全可用`。

**相关代码**：
- 后端字段：`[VariableDef.AvailableReqs](backend/pkg/proto-test/params/variable_defs.go:11)`
- 注册：`[registry.go](backend/pkg/proto-test/variables/registry.go)` 各变量的 `AvailableReqs` 值
- 前端过滤：`[variable-select.vue loadVariables](frontend/src/pages/proto-test/shared/variable-select.vue)`（按 `msgName` 过滤）
- 回归测试：`stream-proxy-variable.spec.ts` 的「变量按 Req 过滤 (G5)」套件

## 4. E2E 测试

- 测试文件：`frontend/e2e/proto-test/proto-test-base.spec.ts` 等拆分套件（见 `e2e/proto-test/CLAUDE.md`）
- Page Object：`frontend/e2e/shared/pages/ProtoTestPage.ts`
- 运行方式：`wails3 dev` 启动后 `npx playwright test proto-test/`
- 详细说明：`frontend/e2e/proto-test-replay-tests.md`

### 测试套件覆盖

| 测试套件 | 测试用例数 |
|---------|-----------|
| 页面结构与页签 | 4 |
| 目标服务输入 | 4 |
| 录制按钮 | 2 |
| 多选模式 | 3 |
| 测试用例管理 | 7 |
| 重放控制 | 2 |
| 卡片编辑器 | 4 |
| 页面布局 | 3 |
| 重放消息追加功能 | 7 |
| 测试用例执行重放 | 3 |
| 实时拦截改包 | 8 |
| 表格列变体与 Req 过滤 | 4 |

## 5. 相关文档

| 文档 | 路径 | 说明 |
|------|------|------|
| 布局文档 | `frontend/docs/layout/pages/proto-test/index.md` | ASCII 布局图、组件树、数据流、时序图 |
| 后端包文档 | `backend/pkg/proto-test/CLAUDE.md` | 后端 Service 索引、协议格式、设计决策 |
| 用例格式 | `cases/proto_cases/CLAUDE.md` | 用例文件格式 |
| 设计规格 | `docs/superpowers/specs/2026-06-02-stream-proxy-req-ack-pairing.md` | Req/Ack 配对展示设计 |
| 设计规格 | `docs/superpowers/specs/2026-06-09-stream-proxy-intercept-filter-mode.md` | 实时拦截改包功能设计 |

## 6. 路由

- 路径：`/ProtoTest`
- 菜单标签：`Proto测试`

## 7. 目录结构

```
pages/proto-test/
├── index.vue                                    # 页面入口（页签编排）
├── stream-proxy/
│   └── packet-tab.vue                           # 发包改包页签
├── cases/
│   └── testcase-tab.vue                         # 测试用例页签
├── replay-result/
│   ├── replay-result-tab.vue                    # 重放结果页签
│   └── replay-result-selector.vue              # 重放结果选择器
├── shared/
│   ├── protocol-content.requirement.ts          # RecordControlService/ProtoTestConfigService 接口封装
│   ├── case-selector.vue + .requirement.ts      # 用例选择下拉框
│   ├── message-table.vue + .requirement.ts      # 配对消息列表表格
│   ├── paired-payload-editor.vue + .requirement.ts  # 配对 JSON 编辑器容器
│   ├── replay-control.vue + .requirement.ts     # 重发控制 + Ntf 显示
│   ├── target-service-config.vue + .requirement.ts  # 目标服务配置
│   ├── req-card-editor.vue                      # Req 卡片式编辑器
│   ├── payload-editor.vue                       # 孤儿保留（疑似 paired-payload-editor 前身）
│   ├── field-item.vue                           # 单个字段编辑项
│   ├── combo-select.vue                         # 字段类型：下拉选择
│   ├── enum-select.vue                          # 字段类型：枚举选择
│   ├── range-input.vue                          # 字段类型：范围输入
│   ├── variable-select.vue + .requirement.ts    # 字段类型：变量选择
│   └── composables/
│       ├── use-paired-messages.ts               # Req/Ack 配对算法
│       └── use-selected-entry.ts                # 选中项管理
└── CLAUDE.md                                    # 本文件
```
