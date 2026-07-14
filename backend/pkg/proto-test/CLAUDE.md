# proto-test (prototest) -- 协议录制/重放/拦截包

本包提供协议测试的核心能力：TCP+HTTP 代理录制、实时拦截改包、录制文件/用例管理、协议重放与迭代发送。

## 目录结构

```
backend/pkg/proto-test/
├── wails.go                   # 共享类型（RecordFileData、RecordEntryView）与转换函数
├── wails_record_control.go    # RecordControlService：监听/录制启停、状态、penging 放行
├── wails_record_file.go       # RecordFileService：录制文件加载、保存、Payload 编辑
├── wails_replay_control.go    # ReplayControlService：重放、迭代发送、连接池管理
├── wails_test_case.go         # TestCaseService：用例目录、保存、加载、删除
├── config.go                  # ProtoTestConfigService：监听端口与目标地址配置
├── record_worker.go           # RecordWorker：TCP+HTTP 代理录制 + 拦截模式
├── replay_worker.go           # ReplayWorker：异步重放任务与进度推送
├── llm.go                     # AI Agent 工具注册
├── testcase.go                # 用例类型别名 → cases 包
├── cobra.go                   # proto-test CLI 子命令（当前仅 --help）
├── cobra-help.md              # CLI help 文本
├── params/                    # 参数迭代 + 变量定义
│   ├── types.go               # RangeValue、FieldValues（前端 4 态输入）
│   ├── iteration.go           # IterationConfig、GenerateIterativeMessages
│   ├── variable_defs.go       # 变量注册表与查找
│   └── *_test.go
├── cases/                     # 录制/用例纯数据类型 + 文件 I/O
│   ├── record.go              # Recording、RecordEntry 定义与序列化
│   ├── testcase.go            # 用例文件读写与归一化
│   └── *_test.go
├── server-config/             # Unity 服务器注入 + 客户端配置导出
│   ├── wails.go               # ServerConfigService：Wails 服务接口
│   ├── server_config.go       # InjectUnityServer、ExportClientConfig
│   ├── server_config_test.go  # 单元测试
│   └── CLAUDE.md              # 子包文档
├── msg/                       # 协议解析与运行时
│   ├── frame.go               # TCP 帧编解码
│   ├── crypto.go              # 加密/解密
│   ├── compress.go            # 压缩/解压
│   ├── decoder.go             # Payload 解码
│   ├── record.go              # Recorder 运行时（内存录制，不自动落盘）
│   ├── replay.go              # 重放编排（账号循环、会话、变量上下文）
│   ├── auth.go                # 登录/鉴权（AuthLogin、LoginReq/Resp、payload 提取）
│   ├── transport.go           # 传输底层（EncodeClientMessage、payloadToJSON、readDrainer）
│   ├── frame_mux.go           # FrameMux 帧多路复用
│   ├── conn_pool.go           # 账号连接池
│   ├── msg_registry.go        # 消息注册表
│   ├── variable.go            # 变量协议解码（PveGuildCityData/TransportRawNtf 提取）
│   ├── variable_runtime.go    # 变量编排（注册、扫描、提取、注入应用层流程）
│   └── *_test.go
├── docs/
│   ├── design-decisions.md    # 设计决策（拦截模式、数据契约、Recording 版本）
│   ├── known-issues.md        # 已知问题与限制
│   └── intercept-batch-release-investigation.md # 批量放行改造调查
├── DEVELOPMENT_ISSUES.md      # 开发经验沉淀
└── ISSUES-RESOLVED.md         # 历史问题修复记录
```

## 核心 Service

| Service | 文件 | 职责 | 对应前端组件 |
|---------|------|------|------------|
| `RecordControlService` | [wails_record_control.go](wails_record_control.go) | 启动/停止监听、启动/停止录制、获取状态、放行 pending 消息、切换拦截模式 | `packet-tab.vue` 录制按钮、`target-service-config.vue` 设置抽屉 |
| `RecordFileService` | [wails_record_file.go](wails_record_file.go) | 加载/保存录制文件、更新消息 Payload/字段值/描述 | `packet-tab.vue`、`paired-payload-editor.vue` |
| `ReplayControlService` | [wails_replay_control.go](wails_replay_control.go) | 重放、迭代发送、连接池管理、变量展开 | `packet-tab.vue`、`testcase-tab.vue` 重放按钮 |
| `TestCaseService` | [wails_test_case.go](wails_test_case.go) | 用例列表、保存（覆盖）、追加、加载、删除 | `testcase-tab.vue`、`packet-tab.vue` 用例保存对话框 |
| `ProtoTestConfigService` | [config.go](config.go) | 监听端口与目标地址配置的持久化 | `target-service-config.vue` 设置抽屉 |

**覆盖与追加语义分离**：`SaveTestCase` 为覆盖写入，供 `testcase-tab.vue` 的"新增模块"/"删除消息"/"保存顺序"使用；`AppendTestCase` 为追加写入，供 `packet-tab.vue` 的"保存到用例"对话框使用。两者不可互换，否则会导致用例数据被意外覆盖或顺序/删除操作失效。

底层 `msg/replay.go` 中的 `SendMessages` 负责外层账号范围循环与内层单账号发送，[wails_replay_control.go](wails_replay_control.go) 的方法最终都会委托到 `ReplayWorker`。

## 核心类型

| 类型 | 文件 | 说明 |
|------|------|------|
| `RecordFileData` | [wails.go:19](wails.go:19) | 返回给前端的录制文件视图 |
| `RecordEntryView` | [wails.go:28](wails.go:28) | 单条消息视图，Payload 为 `map[string]any` |
| `RecordProgress` | [record_worker.go:64](record_worker.go:64) | 录制/监听进度状态 |
| `ReplayProgress` | [replay_worker.go:14](replay_worker.go:14) | 重放进度状态 |
| `RecordWorker` | [record_worker.go:72](record_worker.go:72) | TCP+HTTP 代理录制与拦截工作器 |
| `ReplayWorker` | [replay_worker.go:25](replay_worker.go:25) | 异步重放工作器 |
| `ProtoTestConfig` | [config.go:13](config.go:13) | 监听端口与目标地址配置 |

## 变量系统

字段 `input_type: "variable"` 支持两类变量：

| 变量 | 短名 | 来源 | 说明 |
|------|------|------|------|
| 城池 ID | `cityId` | 服务端 `PveGuildCityDataNtf` / `TransportRawNtf` | 工会城战场景，由 `FrameMux` 缓存并惰性提取 |
| 当前账号 | `openid` | 发送循环预置 `accountID` | 账号级变量，无需 Ntf 提取，适合将当前账号传入 payload |

变量注册表位于 `params` 包，由 [variables 包](variables/CLAUDE.md) 在 `init()` 中注册；`msg` 包负责运行时扫描、惰性提取和 payload 注入。`DecodedFrame` 实际定义在 [params/frame.go](params/frame.go:7)。

## 数据契约

- 存储层使用 `msg.RecordEntry`，Payload 为 `json.RawMessage`。
- 前后端交互使用 `RecordEntryView`，Payload 为 `map[string]any`。
- 唯一转换点：`singleEntryToView` 与 `viewsToEntries`，分别定义在 [wails.go:43](wails.go:43) 和 [wails_record_file.go:59](wails_record_file.go:59)。
- 修改 `Recording` 结构体时，必须同步递增 `msg.RecordingVersion` 并更新 `cases/proto_cases/CLAUDE.md`，详见 [docs/design-decisions.md](docs/design-decisions.md)。

## LLM 工具

[llm.go](llm.go) 注册的 AI Agent 工具：

| 工具名 | 说明 |
|--------|------|
| `start_record` | 启动协议录制 |
| `stop_record` | 停止录制 |
| `load_case_list` | 获取用例列表 |
| `load_case` | 加载指定用例 |
| `delete_case` | 删除用例 |
| `save_case` | 保存用例 |

## CLI 子命令 <!-- layered -->

[cobra.go](cobra.go) + [cobra-help.md](cobra-help.md) 定义 proto-test 的 CLI 子命令，供 AI Agent skill 调用。完整参数说明见 cobra-help.md。

### 命令树

```
proto-test
├── case                    # 用例管理
│   ├── list                # 列出所有用例
│   ├── show <name>         # 查看用例详情（--format summary 仅输出序号+名称+描述）
│   ├── edit <name>         # 修改已有用例（--merge/--payload/--desc/--append/--remove）
│   ├── create <name>       # 从文件/标准输入创建用例
│   └── delete <name>       # 删除用例（--force 跳过确认）
├── replay <case-name>      # 重放用例（--select/--select-name 选择消息子集）
└── send-msg                # 发送单条自定义消息（--msg-id/--msg-name/--payload）
```

### 设计原则

- **CLI 范围**：用例管理、重放、单条发送。**录制和实时拦截改包依赖 GUI 交互，不在 CLI 范围内**。
- **索引一致性**：`case show` 输出的 1-based 消息序号与 `replay --select`、`case edit --msg` 的索引完全一致，AI 可先 show 再操作。
- **JSON merge patch**：`case edit --merge` 采用 RFC 7396 语义，AI 只需构造要改的字段，不用关心 payload 完整结构。
- **选择重放**：`replay --select 3-7` / `--select-name GmCommandReq` 覆盖前端"多选 Req 重发"场景，后端复用 `SendMessages` 的消息子集能力。

### 实现状态

| 子命令 | 状态 | 说明 |
|--------|------|------|
| `case list` | ✅ 已实现 | 复用 `TestCaseService.LoadTestCaseList`，支持 `--format table|json` |
| `case show` | ✅ 已实现 | 复用 `TestCaseService.LoadTestCase` + `RecordFileService`，支持 `--format summary|json`，序号为 1-based `seq` |
| `replay` | ✅ 已实现 | 复用 `msg.SendMessages`，支持 `--range`/`--repeat`/`--select`/`--select-name`/`--concurrency`/`--print-ack` 等 |
| `send-msg` | ✅ 已实现 | 构造单条 `RecordMessage` 调用 `msg.SendMessages`，`--server` 必填 |
| `case edit` | 🔲 占位 | 目标设计见 `cobra-help.md`，含 `--merge`(RFC7396)/`--append`/`--remove` |
| `case create` | 🔲 占位 | 目标设计见 `cobra-help.md` |
| `case delete` | 🔲 占位 | 目标设计见 `cobra-help.md` |

所有子命令在 `NewProtoTestCmd()` 内通过 `cmd.AddCommand()` 注册。

### 与现有工具的关系

| CLI 命令 | 复用的后端能力 | 参考原型 |
|----------|--------------|----------|
| `case *` | `TestCaseService`（wails_test_case.go） | LLM 工具 `load_case_list`/`load_case`/`delete_case`/`save_case` |
| `replay` | `msg.SendMessages`（replay.go） | `cmd/tests/streamproxy -replay`、`cmd/tests/replay_guild_war` |
| `send-msg` | `msg.SendMessages`（单条） | LLM 工具未覆盖，新增 |

## 设计决策与已知问题

- 设计决策详见 [docs/design-decisions.md](docs/design-decisions.md)。
- 已知问题与限制详见 [docs/known-issues.md](docs/known-issues.md)。
- 2026-06-10 拦截批量放行改造调查：[docs/intercept-batch-release-investigation.md](docs/intercept-batch-release-investigation.md)。
- 开发经验沉淀见 [DEVELOPMENT_ISSUES.md](DEVELOPMENT_ISSUES.md)。
- 历史问题修复记录见 [ISSUES-RESOLVED.md](ISSUES-RESOLVED.md)。

### 录制数据驻留内存（2026-06-15）

**决策**：录制过程中产生的协议数据驻留内存（`Recorder.messages`），**不自动落盘**。持久化由前端"保存为用例"按钮手动触发，通过 `SaveRecordFile` 写入 `cases/proto_cases/`。

**背景**：此前实现曾在 `RecordWorker.Stop()` 和 TCP 连接关闭时自动调用 `Recorder.Save()` 落盘到 `record_YYYYMMDD_HHMMSS.json`，这是偏离原始设计的自动发挥。它导致仓库根目录残留临时录制文件，且与"保存为用例"职责重叠。

**影响**：
- `StartRecord` / `RecordWorker.Start` 移除了 `filePath` 参数
- `Recorder` 移除了 `filename` 字段、`Save()` 方法、`GetFilename()` 方法
- `packet-tab.vue` 的 payload 编辑改为纯内存操作（直接改 `recordData.messages`），不再依赖落盘文件
- `.gitignore` 增加 `record_*.json` 防回归

## 测试编写规范 <!-- layered -->

本包曾因**虚假/过拟合测试**导致严重的时序竞态 bug（工会战 cityId 变量提取失效）未被捕获——测试全绿但真实重放仍用写死值。以下规范是血泪教训，**编写或修改本包测试前必须遵守**。

### 禁止虚假/过拟合测试（强约束）

测试必须验证**可观测的真实行为**，而非内部实现细节。以下模式属于禁止项：

| 禁止模式 | 为什么是虚假测试 | 正确做法 |
|----------|----------------|----------|
| 断言内部私有函数的返回值，而非端到端结果 | 实现重构后测试即失效，但真实行为未验证 | 断言最终发送的 payload / cache 状态 / 提取后的 store |
| 构造输入后直接 `assert.Equal(预期)`，不调用被测主流程 | 等于把"预期"写进测试，永远为真 | 必须走完 `ExtractVariablesForMessage` → `resolveVariablePayload` → 校验 payload 全链路 |
| 用 `if 判断条件 { 记为跳过 }` 模拟失败分支，不触发真实失败 | 没有调用真实代码，测的是测试自己的逻辑 | 让真实路径产生失败（如不推送 Ntf 触发超时），再断言失败后的 continue 语义 |
| 测试函数名声称验证某场景，但函数体只 mock 了该场景的"结果" | 名实不符，无法捕获回归 | 场景的前置条件必须真实构造（如延迟推送 Ntf），让被测代码自己走到该分支 |
| 永真断言：`assert.NotNil(永远非nil的对象)` / `assert.True(true)` | 不提供任何信息 | 断言必须能区分"修复生效"与"修复失效" |

**自检清单**（提交测试前自问）：
1. 如果把被测函数的实现整体替换成错误的，这个测试会失败吗？若不会，就是虚假测试。
2. 断言的值，是真实经过被测代码计算出来的，还是我在测试里直接写死的？
3. 测试覆盖的场景，其触发条件是真实构造的，还是我用 `if` 模拟的？

### 必须用 FakeConn 验证时序（强约束）

涉及协议帧收发、变量提取、readLoop/WaitMsg 等**时序相关**的逻辑，**禁止**用纯函数单测或 mock 接口测试，**必须**用 `msg/fakconn_impl_test.go` 的 `FakeConn` 真实模拟网络时序。

原因：真实游戏服务器的 Ntf 推送时机不确定，无法稳定复现时序 bug；FakeConn 让测试代码精确控制"服务器在第几步推送什么帧"，从而确定性复现时序窗口。

### FakeConn 使用指南

`FakeConn`（[msg/fakconn_impl_test.go](msg/fakconn_impl_test.go)）是实现 `net.Conn` 的内存连接，设计参考 Netty EmbeddedChannel（单对象脚本式编排），核心价值是**让测试精确控制服务端推送时机**。文件用 `_test.go` 后缀确保只在测试构建编译，不进入生产二进制。

#### 数据流向（与真实 TCP 一致）

```
客户端写(conn.Write)       → outboundBuf → 测试读取(WaitClientWrite / ClientWrites)
测试推入(PushServerFrame)  → inboundCh   → 客户端读(conn.Read，readLoop 消费)
```

#### 核心 API

| 方法 | 方向 | 用途 |
|------|------|------|
| `NewFakeConn()` | — | 创建内存连接 |
| `PushServerFrame(msgID, payload)` | 服务端→客户端 | 编码并推入一帧（`FlagEncrypt`，readLoop 可解码）。channel 满会 panic（防止静默丢帧） |
| `PushRawFrame(rawFrame)` | 服务端→客户端 | 推入预编码原始字节（精确控制 flags/seqID 时用） |
| `WaitClientWrite(timeout)` | 客户端→服务端 | 阻塞等待客户端写出下一帧，返回解码后的 `DecodedFrame`。**断言客户端发出了预期消息用此** |
| `ClientWrites()` | 客户端→服务端 | 非阻塞取出当前所有客户端写出帧（无数据返回 nil） |
| `MakeServerPayload(protoData)` | helper | 构造带 2 字节 LE 长度前缀的 payload（`TransportRawNtf` 等内层 ByteStream 格式需要） |

#### 标准测试骨架

```go
func TestXxx(t *testing.T) {
    fc := NewFakeConn()
    defer fc.Close()

    // watchedIDs: 本测试关心的服务端消息 ID
    mux := NewFrameMux(fc, []uint16{uint16(pb.EGameMsgID_TransportRawNtf_id)})
    mux.wg.Add(1)
    go mux.readLoop(nil, "test")  // 启动真实读循环
    defer mux.Stop()

    time.Sleep(50 * time.Millisecond) // 确保 readLoop 就绪

    // 步骤1: 客户端发前置消息（模拟 sendMessagesOnce 发送循环）
    fc.Write(EncodeFrame(reqMsgID, seqID, FlagEncrypt, []byte(payload), true))
    got, err := fc.WaitClientWrite(2 * time.Second) // 断言确实发出
    require.Equal(t, reqMsgID, got.MsgID)

    // 步骤2: 在"前置消息发出后、变量消息发出前"推送 Ntf（关键时序窗口）
    fc.PushServerFrame(ntfMsgID, ntfPayload)

    // 步骤3: 等待 readLoop 缓存（轮询 cache）
    require.Eventually(t, func() bool {
        _, err := mux.WaitMsg(ntfMsgID, 2*time.Second)
        return err == nil
    }, 3*time.Second, 50*time.Millisecond)

    // 步骤4: 惰性提取（此时 cache 已命中，模拟发送循环中的提取点）
    store := map[string]any{}
    require.NoError(t, ExtractVariablesForMessage(varMsg, mux, store))

    // 步骤5: 校验最终 payload 被替换（端到端断言，而非只查 store）
    resolved := resolveVariablePayload(varMsg, store)
    var result map[string]any
    require.NoError(t, json.Unmarshal([]byte(resolved), &result))
    assert.Equal(t, float64(expectedCityID), result["cityId"])
}
```

#### 验证时序窗口的关键技巧

惰性提取修复的核心是"Ntf 必须在 Req 之前的消息发出后、Req 发出前到达"。要验证这个窗口，**不要预推 Ntf**，而是在另一个 goroutine 里阻塞调用提取，再延迟推送：

```go
// 验证"阻塞→被唤醒"路径（见 TestLazyExtract_BlockingThenWakeOnDelayedNtf）
done := make(chan error, 1)
go func() { done <- ExtractVariablesForMessage(msg, mux, store) }()

// 先确认它真的在阻塞（短超时探测）
select {
case <-done:
    t.Fatal("不应在 Ntf 推送前返回")
case <-time.After(200 * time.Millisecond):
    // 预期：仍在阻塞
}

// 推送 Ntf，应唤醒
fc.PushServerFrame(ntfMsgID, payload)
select {
case err := <-done:
    require.NoError(t, err)
case <-time.After(6 * time.Second):
    t.Fatal("notifyCh 未唤醒 WaitMsg")
}
```

#### 常见错误

| 错误 | 后果 | 正确做法 |
|------|------|----------|
| 预推 Ntf 后立刻提取，不验证阻塞路径 | 无法发现 WaitMsg 唤醒机制的回归 | 关键场景用"延迟推送 + goroutine 阻塞探测" |
| 只断言 `store["cityId"]`，不断言最终 payload | 实现可能 store 填对了但 payload 替换坏了 | 必须走到 `resolveVariablePayload` 并解析 payload 断言 |
| 占位值用真实可能值（如用 4293 既当占位又当预期） | 断言永远为真，无法区分替换是否生效 | 占位值故意与 Ntf 值不同（如占位 9999、预期 5566） |
| 不 `defer fc.Close()` / 不 `defer mux.Stop()` | goroutine 泄漏 | 必须 defer 清理 |
| 测试用例里造了永不触发的失败分支 | 虚假测试（见上文禁止模式） | 失败场景要由真实代码路径产生 |

### 测试文件索引

| 文件 | 覆盖范围 |
|------|---------|
| `msg/lazy_extract_test.go` | 惰性变量提取：命中/超时/未注册/阻塞唤醒/DrainAndStart 池路径/失败跳过/最新 Ntf 胜出 |
| `msg/fakconn_test.go` | FakeConn 本身的 net.Conn 语义 |
| `msg/frame_mux_test.go` | FrameMux：DrainAndStart 排空丢弃、cache 覆盖、WaitMsg |
| `msg/replay_test.go` | 重放编排 |

## 相关文档

| 文档 | 路径 | 说明 |
|------|------|------|
| 前端页面文档 | `frontend/src/pages/stream-proxy/CLAUDE.md` | requirement.ts 索引、E2E 测试 |
| 布局文档 | `frontend/docs/layout/pages/stream-proxy/index.md` | ASCII 布局图、时序图 |
| 数据流 | `frontend/docs/layout/pages/stream-proxy/data-flow.md` | 事件路由、拦截流程 |
| 用例格式 | `cases/proto_cases/CLAUDE.md` | 用例文件格式 |
| variables 包 | `pkg/proto-test/variables/CLAUDE.md` | 内置变量实现（cityId/openid） |
| params 包 | `pkg/proto-test/params/CLAUDE.md` | 纯数据类型、注册表、DecodedFrame |
| 设计规格 | `docs/superpowers/specs/2026-06-02-stream-proxy-req-ack-pairing.md` | Req/Ack 配对展示设计 |
| 设计规格 | `docs/superpowers/specs/2026-06-09-stream-proxy-intercept-filter-mode.md` | 实时拦截改包功能设计（初版，部分与现状不符） |
