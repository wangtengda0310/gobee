# 待完成任务

## streamproxy / stream-proxy 命名混淆梳理（2026-06-17 发现，待单独会话）

- [x] **命名重构已完成（2026-06-18，squash 进单个 commit）** — 三个易混名字梳理落地：`pkg/streamproxy` → `backend/pkg/proto-test/server-config`（Unity 服务器注入 + 客户端配置导出）；前端 `pages/stream-proxy` → `pages/proto-test`；保留 `cmd/tests/streamproxy`（CLI 流量代理，功能独立）。配套完成前端页签重组（stream-proxy/cases/replay-result/shared）+ 内层 `rain-qa-func/` → `backend/`（见下两条 [x]）。详见 [specs/2026-06-18-streamproxy-rename-design.md](superpowers/specs/2026-06-18-streamproxy-rename-design.md)。

- [x] **前端页签目录重组（命名重构的后续独立工作，2026-06-18 完成）** — 已完成：`pages/proto-test/` 扁平 `components/`（25 文件）重组为 `stream-proxy/`/`cases/`/`replay-result/`/`shared/`(+`shared/composables/`) 4 目录（严格按依赖归类，shared 扁平+composables，UI 不变）。详见 [specs/2026-06-18-proto-test-tab-reorganize-design.md](superpowers/specs/2026-06-18-proto-test-tab-reorganize-design.md)。

- [x] **内层目录 `rain-qa-func/` → `backend/`（独立重构，2026-06-18 完成 cbf33b2..0ff1a12）** — 已完成：内层改名 `backend/`（与 `frontend/` 对称，消除内外层 rain-qa-func 同名混淆），module 名 `rain-qa-func` 保留。A1 Go import + B1 bindings/前端 + C1 文档 + D1 build/验证，final review clean（0 findings，全局源码零残留）。详见 [specs/2026-06-18-backend-rename-design.md](superpowers/specs/2026-06-18-backend-rename-design.md)。

## streamproxy 拦截改包（2026-06-10 调查+产品确认，待单独会话实现）

- [x] **streamproxy 拦截改包批量放行改造（已完成 38656a6 + 97f604a + 63baa42）** — 后端 `ReleasePendingMessages`/`ReleaseAllPending`（按 connID、seq_id 排序）+ `onLoginReqParsed`（LoginReq 解析）+ `SetFilterMode`（关闭后透传）；前端 `message-table.vue` 拦截标记（interceptedSeqIds）+ `protocol-content.requirement` 放行接口。详见 [intercept-batch-release-investigation.md](../backend/pkg/proto-test/docs/intercept-batch-release-investigation.md) §4.5、§6
  - 后端：`pending` + `ReleasePendingMessages`（原连接、seq_id 排序、fail-fast）；LoginReq/Ping 透传；修复 `RecordLoginPayload`；停录/关改包时发送已编辑内容后透传
  - 前端：**新增**改包专用放行按钮；`filterMode` 下隐藏重发/N 次/迭代/应用，禁用「开始重放」；待放行 UI + 合并 toast
  - `filePath`：默认不改；仅当影响改包时用时间戳新文件名做最小修复
  - 已定：Ping P0、批次快照、空 pending 放行=空遍历、录制以实际发送为准

- [x] **streamproxy 多连接 conn_id（已完成）** — `ConnIDCounter` + 连接池（`pooled_authenticator`）+ `ReleasePendingMessages` 按 connID 放行 + 前端多连接展示。

## Bug

- [ ] **战斗测试 reverse_translate.go color 翻译语义错误（2026-06-23 发现）** — 座位 [`CustomHero.Color`](../backend/pkg/function-test/services.go:59) 真实语义是"身份阵营颜色编号"，权威定义为游戏配表 `IdentityEncodeRule_身份编码规则表`（`config/excel/`，字段 IdentityId→Color→Encode；主公 1~8、反贼 65、内奸 97、黄巾 105、刺客 113、伪帝 121，前端 `Identity.ts` 的 `excelIdentityColorMap` 是其复刻）。但 [`reverse_translate.go`](../backend/pkg/function-test/reverse_translate.go:716) 误用 `countryMap`（ECountry 国家表：1=Qin 秦 … 16=Yan 燕）翻译 `hero.Color`，导致 color∈1~16 被错译为"秦/西楚/孙吴…"国家名、color>16 仅显示纯数字。修复方向：复刻 IdentityEncodeRule 的 color→阵营名映射替代 countryMap（参考 `IdentityCampsTemplate` 阵营大类）。当前状态：4 处代码注释已标注 bug（services.go / reverse_translate.go / 前端 Identity.ts），翻译逻辑未改。

- [ ] 协议录制根目录 `record_*.json` — 低优先级；仅在与改包冲突时按上待办 `filePath` 最小策略处理

- [ ] wails3 dev被ctrl+c停止后，对应的node和goalng程序子进程没有一同被停止

## proto-test 惰性变量提取后续（2026-06-15 审核遗留）

基础修复已落地（提交 `02cb126` + `e7e7b74` + `8e24e99`），以下为四份实施后审核（正确性/并发/测试/架构）标注但本次范围外的后续改进：

- [ ] **5s 超时硬编码** — `ExtractVariablesForMessage` 的 `WaitMsg(watchID, 5*time.Second)` 超时硬编码且阻塞发送循环，导致重放取消响应延迟 bounded 5s/条变量消息。应给 `WaitMsg` 增加 `context.Context` 参数，发送循环每条消息前探测 `ctx.Done()`（设计计划 §7 已记录）
- [ ] **`ScanFieldValuesForVariables` 重复扫描** — `sendMessagesOnce` 阶段0 与 `prepareVariableContext` 各扫一次。应让 `prepareVariableContext` 接收扫描结果而非重新扫描，使"阶段0"与"阶段2"数据流显式化
- [ ] **测试 Medium：`FullGuildWarSequence` 魔数索引** — `varMsgIdx := 8` 硬编码 0-based 索引，序列增删一条即错位。改为按 `MsgName` 查找 payload；同时该测试与 `GuildWarTimeline` 功能重叠，可考虑合并
- [ ] **测试 Medium：`ConcurrentExtractSafety` 断言过弱** — 仅断言 `count >= 1`，无法证明并发无 data race。应断言所有 goroutine 都在超时窗口内返回（无挂起）+ `require.NotPanics`；建议 CI 跑 `go test -race`（需本机装 gcc/cgo）
- [ ] **正确性 Low：repeatCount>1 跨轮变量不刷新** — `variableStore` exists 跳过 + cache 覆盖式永不清除 = 第二轮起永不重新提取。对当前 cityId（请求触发型、单场稳定）无影响，但属潜在陷阱。应在 `VariableDef` 注释固化"请求触发型"契约；后续若需轮次敏感变量，每轮开始清空 store + cache
- [ ] **正确性 Low：`conn_pool.IsConnAlive` 可能吞字节** — idle 连接上 `Read(1 byte)` 探活可能读走服务端推送帧的 1 字节致 readLoop 解析错位。连接池既存逻辑，skipDrain=true 路径下 `GetOrCreate` 内部仍调用。当前工会战场景 Ntf 在 Req 后推送（连接 Borrowed 非 idle）不触发，可后续改用零字节 Read 探活
- [ ] **架构 Critical（合并时自动消除）：`02cb126` 中间提交不可编译** — 删除 `ExtractVariableValues` 但未改 `replay.go`，bisect/revert 会失败。squash 合并到 dev 时自动消除；流程改进：后续禁止删除被引用函数时不改调用方

## E2E 测试改进

详见 `frontend/e2e/CLAUDE.md`

- [x] **proto-test base spec「flaky」实际是确定性 bug（2026-06-18 修复 d0eb203）** — 原「约 6/12 flaky 每次变化」为误判：连续跑 4 次（`--retries=0 --workers=1`）失败 test 完全一致，实为 2/12 确定性失败。真实根因：`settingsButton` locator（`button:visible 设置 .first()`）误命中 header 全局「设置」导航按钮（DOM 顺序在前），抽屉从没打开 → 抽屉内 `input[placeholder*="不限"]` 找不到。修复：`target-service-config.vue` 加 `data-testid="target-service-settings-btn"` + `ProtoTestPage.ts` locator 改用 testid + `stream-proxy-base.spec.ts` 误改的 `192.168.1.1` 恢复 `10.254.114.204`（c048ce8 引入）。验证 12/12 passed，教训记 `frontend/e2e/CLAUDE.md` §15。**原假设的「test 间状态隔离/受控组件」根因不成立**（5b08f36 的 native setter 仍保留，属合理改进；但 memory `stream-proxy-e2e-test-quality` 记录的 P0-P2 质量问题——31 个 if 守卫假通过/断言弱——仍待独立处理）。
