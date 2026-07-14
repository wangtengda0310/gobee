# Stream Proxy 页面 Req/Ack 配对展示设计

## 背景

当前协议录制只捕获客户端→服务端方向（Req），表格中每行展示一条消息。用户希望：
- **Req/Ack 配对展示**：同一业务请求和响应在同一行显示（左Req右Ack）
- **Ntf 单行展示**：通知类消息保持现有单行展示
- **动态更新**：录制过程中收到 Ack 时自动填充到对应行的 Ack 列

## 方案选择

**方案B：前端配对**（已确认）
- 后端扩展为双向录制（Req + Ack/Ntf），前端根据消息名动态配对展示
- 后端改动小，配对逻辑在前端完成

---

## 录制文件格式：JSON Lines

采用 **JSON Lines** 格式（`.jsonl`），每条消息独立一行，支持真正的流式追加写入。

### 文件格式 Schema

**第 1 行（Header）：**
```json
{"_type":"header","version":1,"recorded_at":"2026-06-02T10:00:00Z","server_addr":"10.0.0.1:18000","login_payload_b64":"..."}
```

**第 2~N 行（Message）：**
```json
{"_type":"msg","offset_ms":0,"msg_id":1001,"msg_name":"HelloReq","seq_id":0,"direction":"→","payload_json":"{}"}
{"_type":"msg","offset_ms":150,"msg_id":1002,"msg_name":"HelloAck","seq_id":1,"direction":"←","payload_json":"{}"}
```

### Header Line Schema

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `_type` | string | 是 | 固定值 `"header"` |
| `version` | number | 是 | 录制文件版本号 |
| `recorded_at` | string | 是 | ISO8601 格式录制时间 |
| `server_addr` | string | 是 | 目标服务器地址 |
| `login_payload_b64` | string | 否 | LoginReq 解密后 payload（base64） |

### Message Line Schema

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `_type` | string | 是 | 固定值 `"msg"` |
| `offset_ms` | number | 是 | 相对第一条消息的时间偏移（毫秒） |
| `msg_id` | number | 是 | 消息 ID |
| `msg_name` | string | 是 | 消息名称（如 `HelloReq`） |
| `seq_id` | number | 是 | 序列号 |
| `direction` | string | 是 | `"→"`（客户端→服务端）或 `"←"`（服务端→客户端） |
| `payload_json` | string | 是 | JSON 格式的 payload 字符串 |

### 为什么用 JSON Lines

- **流式追加**：每条消息到达立即写入文件一行，无需等待 Stop
- **崩溃安全**：程序崩溃时，已写入文件的消息不会丢失
- **逐行读取**：加载时逐行解析，内存占用低，支持大文件
- **向前兼容**：通过 `version` 字段支持未来格式升级

---

## 后端设计

### 1. Recorder 重构（流式写入）

```go
// Recorder 录制器（JSON Lines 流式写入）
type Recorder struct {
    mu          sync.Mutex
    filename    string
    serverAddr  string
    file        *os.File        // 文件句柄（追加写入）
    startTime   time.Time
    started     bool
    msgCount    int             // 消息计数（替代原 messages 数组）
    headerWritten bool          // 是否已写入 header
    loginPayloadB64 string
    onRecord    OnRecordCallback
}

// RecordFrame 记录一帧（流式追加到文件）
func (r *Recorder) RecordFrame(frame *DecodedFrame, dir string) error {
    // 解析 proto、转 JSON（原有逻辑）
    // ...

    r.mu.Lock()
    defer r.mu.Unlock()

    if !r.started {
        r.startTime = time.Now()
        r.started = true
    }

    // 首次写入：创建文件并写入 header
    if !r.headerWritten {
        if err := r.writeHeader(); err != nil {
            return err
        }
        r.headerWritten = true
    }

    // 构建消息行
    entry := map[string]any{
        "_type":       "msg",
        "offset_ms":   int(time.Since(r.startTime).Milliseconds()),
        "msg_id":      frame.MsgID,
        "msg_name":    GetMsgName(frame.MsgID),
        "seq_id":      frame.SeqID,
        "direction":   dir,
        "payload_json": string(jsonBytes),
    }

    // 追加写入文件
    line, _ := json.Marshal(entry)
    if _, err := r.file.Write(append(line, '\n')); err != nil {
        return fmt.Errorf("写入消息行失败: %v", err)
    }

    r.msgCount++

    // 触发回调
    if r.onRecord != nil {
        r.onRecord(r.msgCount, entry) // 附带最新消息数据
    }

    return nil
}

// writeHeader 写入文件头（JSON Lines 第一行）
func (r *Recorder) writeHeader() error {
    header := map[string]any{
        "_type":             "header",
        "version":           1,
        "recorded_at":       time.Now().Format(time.RFC3339),
        "server_addr":       r.serverAddr,
        "login_payload_b64": r.loginPayloadB64,
    }
    line, _ := json.Marshal(header)
    _, err := r.file.Write(append(line, '\n'))
    return err
}

// Save 关闭文件（JSON Lines 已实时写入，只需关闭）
func (r *Recorder) Save() error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if r.file != nil {
        return r.file.Close()
    }
    return nil
}
```

### 2. decoder.go 双向录制

```go
// 录制 Proto 消息（双向，两个方向都录）
if recorder != nil && frame.MsgID >= 1000 {
    if err := recorder.RecordFrame(frame, dir); err != nil {
        log.Printf("[录制] 错误: %v", err)
    }
}
```

### 3. 录制进度事件（附带最新消息）

```go
// OnRecordCallback 签名更新：附带最新消息数据
type OnRecordCallback func(msgCount int, latestMsg map[string]any)

// emitProgress 推送进度（包含最新消息）
func (w *RecordWorker) emitProgress(latestMsg map[string]any) {
    w.mu.Lock()
    p := w.progress
    w.mu.Unlock()

    data := map[string]any{
        "status":        p.Status,
        "message_count": p.MessageCount,
        "server_addr":   p.ServerAddr,
        "error_message": p.ErrorMessage,
    }
    if latestMsg != nil {
        data["latest_msg"] = latestMsg
    }
    w.app.Event.Emit("record:progress", data)
}
```

### 4. 加载录制文件（JSON Lines 解析）

```go
// LoadRecordFile 加载 JSON Lines 录制文件
func LoadRecordFile(path string) (*Recording, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var rec Recording
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        line := scanner.Text()
        if len(strings.TrimSpace(line)) == 0 {
            continue
        }

        var obj map[string]any
        if err := json.Unmarshal([]byte(line), &obj); err != nil {
            return nil, fmt.Errorf("第 %d 行解析失败: %v", lineNum+1, err)
        }

        lineType, _ := obj["_type"].(string)
        switch lineType {
        case "header":
            rec.Version = int(obj["version"].(float64))
            rec.RecordedAt, _ = obj["recorded_at"].(string)
            rec.ServerAddr, _ = obj["server_addr"].(string)
            rec.LoginPayloadB64, _ = obj["login_payload_b64"].(string)
        case "msg":
            payloadJSON, _ := obj["payload_json"].(string)
            rec.Messages = append(rec.Messages, &RecordEntry{
                OffsetMs:    int(obj["offset_ms"].(float64)),
                MsgID:       uint16(obj["msg_id"].(float64)),
                MsgName:     obj["msg_name"].(string),
                SeqID:       uint32(obj["seq_id"].(float64)),
                Direction:   obj["direction"].(string),
                PayloadJSON: json.RawMessage(payloadJSON),
            })
        }
        lineNum++
    }

    return &rec, scanner.Err()
}
```

---

## Wails 服务方法审查

### 方法清单与优化建议

| 方法 | 状态 | 说明 |
|------|------|------|
| `LoadRecordFile(path)` | **需修改** | 改为 JSON Lines 逐行解析 |
| `StartReplay(...)` | 无需修改 | 委托给 ReplayWorker |
| `StopReplay()` | 无需修改 | 委托给 ReplayWorker |
| `GetReplayStatus()` | 无需修改 | 委托给 ReplayWorker |
| `SaveRecordFile(path, data)` | **需修改** | 改为 JSON Lines 格式输出 |
| `UpdateMessagePayload(path, index, payloadJSON)` | **需适配** | 适配新的 SaveRecordFile |
| `StartRecord(...)` | 无需修改 | 委托给 RecordWorker |
| `StopRecord()` | 无需修改 | 委托给 RecordWorker |
| `GetRecordStatus()` | 无需修改 | 委托给 RecordWorker |
| ~~`SelectRecordFile()`~~ | **移除** | 前端改用 `Dialogs.OpenFile` 直接调用系统对话框 |
| ~~`SelectSavePath()`~~ | **移除** | 前端改用 `Dialogs.SaveFile` 直接调用系统对话框 |

### 问题解答

**1. LoadRecordFile 是否复用流式推送逻辑？**

不需要。"流式推送"是录制时的实时 Event.Emit，LoadRecordFile 是文件加载。改为 JSON Lines 后使用 `bufio.Scanner` 逐行读取即可，仍返回完整数据（前端表格需要全部消息进行配对计算）。

**2. SelectRecordFile / SelectSavePath 是否可以移除？**

可以移除。前端直接使用 Wails 运行时提供的 `Dialogs.OpenFile` / `Dialogs.SaveFile` 调用系统对话框，不需要后端提供对话框方法。

```typescript
// 前端直接调用（无需后端方法）
import { Dialogs } from '@wailsio/runtime'

const result = await Dialogs.OpenFile({
  Title: '选择录制文件',
  CanChooseFiles: true,
  Filters: [{ DisplayName: 'JSON', Pattern: '*.json' }]
})
const path = Array.isArray(result) ? result[0] : result
```

**3. SaveRecordFile 是否还有必要？**

有必要。用于保存 Payload 编辑后的修改。但需要改为输出 JSON Lines 格式（逐行写入 Header + Message）。

**4. replay.go 中的 Replay 函数**

也需要改为 JSON Lines 解析（当前是 `json.Unmarshal(data, &rec)` 解析 JSON 数组）。

---

## 时序图

### 录制时序（JSON Lines 流式写入）

```
用户点击"开始录制"
    │
    ▼
StartRecord(filePath, serverAddr, httpAddr)
    │
    ▼
RecordWorker.Start()
    │
    ├──► 创建/打开文件（os.Create）
    │
    ├──► 启动 TCP 代理监听
    │
    └──► 启动 HTTP 代理监听
    │
    ▼
客户端连接 → 发送 HelloReq
    │
    ▼
RelayAndParse(客户端→服务端)
    │
    ├──► 转发到远程服务器
    │
    └──► 解析帧 → RecordFrame(frame, "→")
                │
                ├──► 解析 Proto → JSON
                │
                ├──► writeHeader()（首次）
                │       {"_type":"header",...}
                │
                ├──► file.Write({"_type":"msg",...})
                │
                └──► onRecord(msgCount, latestMsg)
                            │
                            ▼
                        emitProgress(latestMsg)
                            │
                            ▼
                        Wails Event 'record:progress'
                            │
                            ▼
                        前端：追加消息到列表 → 重新计算配对 → 更新表格
    │
    ▼
服务端返回 HelloAck
    │
    ▼
RelayAndParse(服务端→客户端)
    │
    ├──► 转发到客户端
    │
    └──► 解析帧 → RecordFrame(frame, "←")
                │
                ├──► file.Write({"_type":"msg",...})
                │
                └──► onRecord → emitProgress → 前端动态填充 Ack 列
    │
    ▼
用户点击"停止录制"
    │
    ▼
StopRecord()
    │
    ▼
RecordWorker.Stop()
    │
    ├──► 关闭 TCP/HTTP 监听器
    │
    ├──► recorder.Save() → 关闭文件
    │
    └──► finish("stopped") → emitProgress
```

### 加载时序（JSON Lines 解析）

```
用户点击"加载"
    │
    ▼
LoadRecordFile(path)
    │
    ├──► os.Open(path)
    │
    ├──► bufio.NewScanner(file)
    │
    ├──► 第 1 行 → 解析 Header
    │       {"_type":"header",...}
    │
    ├──► 第 2~N 行 → 逐行解析 Message
    │       {"_type":"msg",...}
    │
    └──► 构建 RecordFileData → 返回前端
                │
                ▼
            前端：buildPairedEntries(messages)
                │
                ▼
            表格渲染（Req/Ack 配对行 + Ntf 单行）
```

### 编辑保存时序（JSON Lines 输出）

```
用户编辑 Payload → 点击"应用"
    │
    ▼
UpdateMessagePayload(path, index, payloadJSON)
    │
    ├──► 验证 JSON 合法性
    │
    ├──► LoadRecordFile(path) → 加载全部消息
    │
    ├──► 修改对应消息的 payload_json
    │
    ├──► SaveRecordFile(path, data)
    │       │
    │       └──► 逐行写入 JSON Lines 格式
    │               第 1 行：Header
    │               第 2~N 行：Message（修改后的 payload）
    │
    └──► LoadRecordFile(path) → 重新加载返回
```

---

## 前端设计

### 1. 配对数据结构

前端内部使用（不传给后端）：

```typescript
type PairedEntry = {
    id: number              // 行唯一ID
    baseName: string        // 消息基础名（如 "Hello"）
    type: 'pair' | 'single' // pair=Req/Ack配对, single=Ntf/其他

    // Req 部分（type='pair'）
    req: RecordEntryView | null

    // Ack 部分（type='pair'，动态到达）
    ack: RecordEntryView | null

    // Ntf 部分（type='single'）
    ntf: RecordEntryView | null

    // 显示用聚合字段
    offset_ms: number
    msg_id: number
    msg_name: string       // 如 "HelloReq | HelloAck" 或 "RoomGameActionNtf"
    direction: string      // "→←" / "→" / "←"
}
```

### 2. 配对算法

```typescript
function buildPairedEntries(messages: RecordEntryView[]): PairedEntry[] {
    const pairs: PairedEntry[] = []
    let nextId = 0

    for (const msg of messages) {
        const name = msg.msg_name

        if (name.endsWith('Ntf')) {
            // Ntf：单行展示
            pairs.push({
                id: nextId++, baseName: name, type: 'single',
                req: null, ack: null, ntf: msg,
                offset_ms: msg.offset_ms, msg_id: msg.msg_id,
                msg_name: name, direction: msg.direction,
            })
        } else if (name.endsWith('Req')) {
            // Req：创建配对行，Ack 留空等待
            const baseName = name.slice(0, -3)
            pairs.push({
                id: nextId++, baseName, type: 'pair',
                req: msg, ack: null, ntf: null,
                offset_ms: msg.offset_ms, msg_id: msg.msg_id,
                msg_name: name, direction: msg.direction,
            })
        } else if (name.endsWith('Ack')) {
            // Ack：向前查找最近一个未匹配的 Req
            const baseName = name.slice(0, -3)
            let matched = false
            for (let i = pairs.length - 1; i >= 0; i--) {
                const p = pairs[i]
                if (p.type === 'pair' && p.baseName === baseName && p.ack === null) {
                    p.ack = msg
                    p.msg_name = `${p.req?.msg_name} | ${msg.msg_name}`
                    p.direction = '→←'
                    matched = true
                    break
                }
            }
            // 找不到对应 Req，作为单行展示
            if (!matched) {
                pairs.push({
                    id: nextId++, baseName, type: 'single',
                    req: null, ack: msg, ntf: null,
                    offset_ms: msg.offset_ms, msg_id: msg.msg_id,
                    msg_name: name, direction: msg.direction,
                })
            }
        } else {
            // 其他消息，单行展示
            pairs.push({
                id: nextId++, baseName: name, type: 'single',
                req: null, ack: null, ntf: msg,
                offset_ms: msg.offset_ms, msg_id: msg.msg_id,
                msg_name: name, direction: msg.direction,
            })
        }
    }

    return pairs
}
```

### 3. 表格布局

```
┌────┬──────────┬──────┬─────────────┬─────────────┬──────┬──────────┬──────┐
│ #  │ 时间      │ MsgID │ 请求(Req)   │ 响应(Ack)   │ SeqID │ 方向     │ 结果 │
├────┼──────────┼──────┼─────────────┼─────────────┼──────┼──────────┼──────┤
│ 0  │ 0         │ 1001  │ HelloReq    │ HelloAck    │ 0     │ C->S,S->C│ 成功  │  ← 配对行
│ 1  │ 150       │ 1002  │ PlayCardReq │ 等待中...   │ 1     │ C->S     │ 超时  │  ← 配对行(Ack未到)
│ 2  │ 200       │ 1005  │ RoomGActNtf │ -           │ 0     │ S->C     │ 成功  │  ← Ntf行
└────┴──────────┴──────┴─────────────┴─────────────┴──────┴──────────┴──────┘
```

**列定义：**
| 列 | 配对行 | Ntf行 |
|---|---|---|
| # | 行号 | 行号 |
| 时间 | Req 的时间 | Ntf 的时间 |
| MsgID | Req 的 MsgID | Ntf 的 MsgID |
| 请求(Req) | Req 消息名 | Ntf 消息名 |
| 响应(Ack) | Ack 消息名 / "等待中..." | "-" |
| SeqID | Req 的 SeqID | Ntf 的 SeqID |
| 方向 | "C->S,S->C" / "C->S" / "S->C" | 实际方向 |
| 结果 | 成功 / 超时 | 成功 |

### 4. Payload 编辑器

**配对行点击** → 左右分栏：
```
┌──────────────────────────────┬──────────────────────────────┐
│ HelloReq (MsgID=1001)        │ HelloAck (MsgID=1002)        │
│ [格式化] [应用]               │ [格式化] [应用]               │
│ ┌──────────────────────────┐ │ ┌──────────────────────────┐ │
│ │ {                        │ │ │ {                        │ │
│ │   "name": "test1"        │ │ │   "result": 0            │ │
│ │ }                        │ │ │ }                        │ │
│ └──────────────────────────┘ │ └──────────────────────────┘ │
└──────────────────────────────┴──────────────────────────────┘
```

- 左侧：Req 的 JSON（**可编辑**）
- 右侧：Ack 的 JSON（**可编辑**）
- Ack 未到时右侧显示 "等待响应..."

**Ntf 行点击** → 保持现有单栏：
```
│ RoomGameActionNtf (MsgID=1005, SeqID=0)        [格式化] [应用] │
│ ┌────────────────────────────────────────────────────────────┐ │
│ │ {                                                          │ │
│ │   "action": "play_card"                                    │ │
│ │ }                                                          │ │
│ └────────────────────────────────────────────────────────────┘ │
```

### 5. 动态更新机制

录制过程中，后端通过 `record:progress` 事件推送最新消息：

```typescript
// use-record-data.ts 中
Events.On('record:progress', (data: any) => {
    recordProgress.value = data
    recordRunning.value = data.status === 'running'

    // 新增：如果有最新消息，添加到消息列表
    if (data.latest_msg) {
        const msg = data.latest_msg
        // 添加到 recordData.messages（如果已加载）
        if (recordData.value) {
            recordData.value.messages.push({
                index: recordData.value.messages.length,
                offset_ms: msg.offset_ms,
                msg_id: msg.msg_id,
                msg_name: msg.msg_name,
                seq_id: msg.seq_id,
                direction: msg.direction,
                payload_json: msg.payload_json,
            })
        }
    }
})
```

前端通过 `computed` 自动重新计算配对：
```typescript
const pairedMessages = computed(() => {
    if (!recordData.value) return []
    return buildPairedEntries(recordData.value.messages)
})
```

当 Ack 到达时：
1. 新消息追加到 `messages` 数组
2. `pairedMessages` computed 重新计算
3. 配对算法找到对应的 Req 行，填充 Ack 列
4. Vue 响应式更新表格显示

---

## 组件变更

### 新增文件

| 文件 | 说明 |
|------|------|
| `composables/use-paired-messages.ts` | 配对逻辑 composable |
| `components/paired-payload-editor.vue` | 左右分栏 Payload 编辑器 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `streamproto/record.go` | Recorder 重构为 JSON Lines 流式写入；RecordFrame 增加 dir 参数；OnRecordCallback 附带 latest_msg |
| `streamproto/decoder.go` | 双向录制（两个方向都调用 RecordFrame） |
| `record_worker.go` | 适配新 Recorder 接口；emitProgress 附带 latest_msg |
| `wails.go` | LoadRecordFile 改为 JSON Lines 解析；返回的 RecordEntryView 包含 direction |
| `composables/use-record-data.ts` | 处理 record:progress 中的 latest_msg |
| `components/message-table.vue` | 支持配对展示（新列定义、渲染逻辑） |
| `components/payload-editor.vue` | 或替换为 paired-payload-editor.vue |
| `index.vue` | 传递 pairedMessages 给子组件；移除 filePath 输入框的 clearable；增加浏览按钮调用 `Dialogs.OpenFile` / `Dialogs.SaveFile` |

---

## 边界情况处理

| 场景 | 处理方案 |
|------|----------|
| Ack 找不到对应 Req | 作为单行展示 |
| 多个相同类型 Req 连续发送 | 按顺序匹配最近的未匹配 Req |
| Req 发送后 Ack 迟迟未到 | 显示 "(等待中...)"，Ack 到后动态填充 |
| 消息名不以 Req/Ack/Ntf 结尾 | 作为单行展示 |
| 录制开始时 Ack 已到达（Req 未录到） | 作为单行展示 |
| 重放功能是否受影响 | 重放只发送 Req，不受 Ack 录制影响 |

---

## 验收标准

- [ ] 后端：Recorder 使用 JSON Lines 流式写入格式
- [ ] 后端：录制文件包含 Header 行（version/recorded_at/server_addr）
- [ ] 后端：双向录制（客户端→服务端和服务端→客户端的 Proto 消息都被录制）
- [ ] 后端：RecordEntry 包含 direction 字段
- [ ] 后端：LoadRecordFile 支持 JSON Lines 解析
- [ ] 后端：录制进度事件附带 latest_msg
- [ ] 后端：LoadRecordFile 改为 JSON Lines 逐行解析
- [ ] 后端：SaveRecordFile 改为 JSON Lines 格式输出
- [ ] 后端：replay.go 的 Replay 函数改为 JSON Lines 解析
- [ ] 前端：JSON Lines Schema 类型定义
- [ ] 前端：表格 Req/Ack 在同一行展示，Ntf 单行展示
- [ ] 前端：表格消息名列拆分为"请求(Req)"和"响应(Ack)"两列，新增"时间"和"结果"列
- [ ] 前端：配对行方向显示 "C->S,S->C"，单方向行显示 "C->S" 或 "S->C"
- [ ] 前端：结果列显示配对状态（成功/超时），成功为 success 类型 NTag，超时为 warning 类型 NTag
- [ ] 前端：Payload 编辑器配对行左右分栏，Ntf 行单栏
- [ ] 前端：动态更新——录制过程中 Ack 到达时自动填充对应行
- [ ] 前端：边界——无匹配 Req 的 Ack 作为单行展示
- [ ] 前端：移除 `SelectRecordFile` / `SelectSavePath` 调用，改用 `Dialogs.OpenFile` / `Dialogs.SaveFile`
- [ ] 前端：输入框增加浏览按钮，只更新输入框文本
- [ ] 重放功能不受影响
- [ ] Bindings 重新生成成功
