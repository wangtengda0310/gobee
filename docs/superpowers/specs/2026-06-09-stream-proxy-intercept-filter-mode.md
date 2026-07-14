# Stream Proxy 实时拦截改包功能设计

## 1. 需求概述

为 proto 测试页面的"发包改包"功能增加实时拦截并修改数据包的能力。

### 1.1 功能描述

1. 在"发包改包"页签增加一个"开启/关闭实时修改"切换按钮
2. 开启实时修改模式后，客户端发往服务端的 **Req 消息**（MsgID ≥ 1000，方向 C→S）会被拦截
3. 被拦截的 Req 消息推送到前端表格，自动选中并滚动到可视区域
4. 用户在前端编辑 Payload 后，点击"重发"按钮将修改后的消息放行到服务端
5. **Ack 和 Ntf 消息（方向 S→C）不拦截**，原样转发

### 1.2 不拦截的消息类型

| 消息类型 | MsgID 范围 | 方向 | 是否拦截 | 原因 |
|---------|-----------|------|---------|------|
| 框架消息 | < 1000 | C→S | ❌ | LoginReq 等框架消息，不拦截 |
| Proto Req | ≥ 1000 | C→S | ✅ | 业务请求，需要拦截修改 |
| Ack / Ntf | ≥ 1000 | S→C | ❌ | 服务端响应，原样转发 |

## 2. 架构设计

### 2.1 整体数据流

```
┌─────────────────────────────────────────────────────────────┐
│                      拦截模式数据流                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   游戏客户端        录制代理(RecordWorker)        游戏服务端  │
│       │                    │                        │       │
│       │───TCP Req────────>│                        │       │
│       │                   │───拦截，不转发──────────│       │
│       │                   │                        │       │
│       │                   │───Event.Emit───────────┐       │
│       │                   │   'record:intercepted' │       │
│       │                   │                        ▼       │
│       │              ┌─────────────────────────────────┐   │
│       │              │      packet-tab.vue (前端)       │   │
│       │              │  1. 追加消息到表格               │   │
│       │              │  2. 自动选中该行                 │   │
│       │              │  3. 滚动到可视区域               │   │
│       │              │  4. 用户编辑 Payload             │   │
│       │              │  5. 点击"重发"按钮               │   │
│       │              └─────────────────────────────────┘   │
│       │                   │                        ▲       │
│       │                   │───SendMessages()───────┘       │
│       │                   │   (修改后的 Req)               │
│       │                   │                                │
│       │                   │───转发到服务端────────────────>│ │
│       │                   │                                │
│                                                             │
│   ───────────────────────────────────────────────────────   │
│   Ack/Ntf 方向（不拦截）：                                   │
│   服务端───TCP Ack/Ntf────>录制代理────原样转发────>客户端   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 组件职责

| 组件 | 职责 |
|------|------|
| `RecordWorker` (后端) | TCP 代理，拦截 Req 消息，推送事件 |
| `RecordControlService` (后端) | 接收 `filterMode` 参数，启动/停止录制 |
| `packet-tab.vue` (前端) | 切换拦截模式，监听拦截事件，管理拦截队列 |
| `message-table.vue` (前端) | 显示被拦截消息的特殊标记，提供滚动方法 |
| `replay-control.vue` (前端) | 复用"重发"按钮放行被拦截消息 |
| `paired-payload-editor.vue` (前端) | 编辑被拦截消息的 Payload |

## 3. 后端设计

### 3.1 RecordWorker 修改

**新增字段**：
```go
type RecordWorker struct {
    // ... 现有字段 ...
    filterMode bool  // 是否启用拦截模式
}
```

**Start 方法**：接收 filterMode 参数并保存
```go
func (w *RecordWorker) Start(filePath string, serverAddr string, httpAddr string, filterMode bool) error {
    w.filterMode = filterMode
    // ... 现有逻辑 ...
}
```

**handleTCPConn 修改**：

客户端→服务端方向不再直接转发，改为：

```go
// 客户端 -> 服务端（录制 + 可能的拦截方向）
go func() {
    defer wg.Done()
    if w.filterMode {
        // 拦截模式：解析但不转发，通过事件推送到前端
        w.interceptAndParse(clientConn, remoteConn, id, streamproto.DirClientToServer, true, w.recorder)
    } else {
        // 普通模式：原样转发
        n := streamproto.RelayAndParse(clientConn, remoteConn, id, streamproto.DirClientToServer, true, w.recorder)
        log.Printf("[RecordWorker] TCP #%d 客户端->服务端 完成, %d 字节", id, n)
    }
    remoteConn.Close()
}()

// 服务端 -> 客户端（始终原样转发，不拦截）
go func() {
    defer wg.Done()
    n := streamproto.RelayAndParse(remoteConn, clientConn, id, streamproto.DirServerToClient, false, w.recorder)
    log.Printf("[RecordWorker] TCP #%d 服务端->客户端 完成, %d 字节", id, n)
    clientConn.Close()
}()
```

### 3.2 新增 interceptAndParse 方法

```go
// interceptAndParse 拦截并解析 TCP 数据
// 从 src 读取数据，解析协议帧，如果是 Req 则拦截（不写入 dst），通过事件推送到前端
// 如果是框架消息（MsgID < 1000）则原样转发到 dst
func (w *RecordWorker) interceptAndParse(
    src net.Conn, dst net.Conn, connID uint64,
    dir string, isClientData bool, recorder *streamproto.Recorder,
) int64 {
    var total int64

    for {
        // 读取帧头
        header := make([]byte, streamproto.FrameHeaderSize)
        n, err := io.ReadFull(src, header)
        if err != nil {
            if err != io.EOF && total > 0 {
                log.Printf("[TCP] #%d %s 读取帧头结束: %v", connID, dir, err)
            }
            return total
        }
        total += int64(n)

        msgLen, _, herr := streamproto.ParseFrameHeader(header)
        if herr != nil {
            log.Printf("[TCP] #%d %s 帧头解析失败: %v", connID, dir, herr)
            dst.Write(header)
            return total
        }

        // 读取消息体
        body := make([]byte, msgLen)
        n, err = io.ReadFull(src, body)
        if err != nil {
            log.Printf("[TCP] #%d %s 读取消息体失败: %v", connID, dir, err)
            dst.Write(header)
            return total
        }
        total += int64(n)

        // 解析帧以判断是否拦截
        raw := make([]byte, len(header)+len(body))
        copy(raw, header)
        copy(raw[len(header):], body)

        frame, derr := streamproto.DecodeFrame(raw, isClientData)
        if derr != nil {
            log.Printf("[TCP] #%d %s 解码失败: %v", connID, dir, derr)
            // 解码失败时原样转发，不拦截
            dst.Write(header)
            dst.Write(body)
            continue
        }

        // 录制（所有消息都录）
        if recorder != nil && frame.MsgID >= 1000 {
            if isClientData && frame.MsgID == 1 {
                recorder.RecordLoginPayload(frame.Payload)
            }
            recorder.RecordFrame(frame, dir)
        }

        // 判断是否需要拦截：Proto Req（MsgID >= 1000，C→S 方向）
        shouldIntercept := frame.MsgID >= 1000 && dir == streamproto.DirClientToServer

        if shouldIntercept {
            // 拦截：不写入 dst，通过事件推送到前端
            log.Printf("[RecordWorker] 拦截消息: conn=%d msg=%s msg_id=%d", connID, frame.MsgName, frame.MsgID)
            w.emitInterceptedMessage(frame, connID)
            // 更新录制计数
            w.updateMessageCount()
        } else {
            // 不拦截：原样转发
            dst.Write(header)
            dst.Write(body)
        }
    }
}
```

### 3.3 新增 emitInterceptedMessage 方法

```go
// emitInterceptedMessage 推送被拦截的消息到前端
func (w *RecordWorker) emitInterceptedMessage(frame *streamproto.DecodedFrame, connID uint64) {
    entryView := singleEntryToView(&streamproto.RecordEntry{
        OffsetMs:    0,
        MsgID:       frame.MsgID,
        MsgName:     frame.MsgName,
        SeqID:       frame.SeqID,
        PayloadJSON: frame.Payload,
        Direction:   streamproto.DirClientToServer,
    }, 0)

    data := map[string]any{
        "latest_msg": entryView,
        "conn_id":    connID,
    }
    w.app.Event.Emit("record:intercepted", data)
}
```

### 3.4 RecordControlService 接口

`StartRecord` 方法签名已有 `filterMode bool` 参数（当前代码第 46 行），无需修改：

```go
func (s *RecordControlService) StartRecord(filePath string, serverAddr string, httpAddr string, filterMode bool) error
```

## 4. 前端设计

### 4.1 packet-tab.vue 修改

**新增状态**：
```typescript
// 拦截队列：记录被拦截但未放行的消息 SeqID
const interceptedSeqIDs = ref<Set<number>>(new Set())
```

**监听拦截事件**：
```typescript
Events.On('record:intercepted', (raw: any) => {
    const data = raw.data ?? raw
    if (!data.latest_msg) return

    const newEntry = RecordEntryView.createFrom({
        ...data.latest_msg,
        index: recordData.value?.messages.length ?? 0,
    })

    // 追加到消息列表
    if (recordData.value) {
        recordData.value.messages = [...recordData.value.messages, newEntry]
        recordData.value.message_count = recordData.value.messages.length
    }

    // 记录到拦截队列
    interceptedSeqIDs.value.add(newEntry.seq_id)

    // 自动选中并滚动
    const newIndex = (recordData.value?.messages.length ?? 1) - 1
    selectedIndex.value = newIndex

    // 滚动到可视区域（通过 message-table 暴露的方法）
    nextTick(() => {
        messageTableRef.value?.scrollToRow(newIndex)
    })
})
```

**修改 handleRetryMessage**：
```typescript
async function handleRetryMessage(count: number) {
    if (!selectedPairedEntry.value || !replayServerAddr.value) return
    const entry = selectedPairedEntry.value
    let targetMsg = entry.req
    if (!targetMsg) {
        message.warning('未找到 Req 消息')
        return
    }

    // 从编辑器获取当前 payload
    const editorPayload = pairedPayloadEditorRef.value?.getCurrentReqPayload()
    if (editorPayload) {
        targetMsg = { ...targetMsg, payload: editorPayload }
    }

    // 检查是否是拦截消息的放行
    const isIntercepted = interceptedSeqIDs.value.has(targetMsg.seq_id)

    try {
        emit('replay-start', 'retry')
        await replayControlService.sendMessages(
            replayServerAddr.value, replayHttpAddr.value, replayOpenID.value,
            [targetMsg], count
        )

        if (isIntercepted) {
            message.success(`已放行 ${targetMsg.msg_name}`)
            interceptedSeqIDs.value.delete(targetMsg.seq_id)
        } else {
            message.info(`正在重发 ${targetMsg.msg_name} (${count} 次)`)
        }
    } catch (e: any) {
        message.error('发送失败: ' + (e.message || e))
    }
}
```

**修改 toggleFilterMode**：
```typescript
function toggleFilterMode() {
    filterMode.value = !filterMode.value
    if (filterMode.value) {
        message.info('实时修改模式已开启：客户端 Req 消息将被拦截，编辑后点击"重发"放行')
    } else {
        message.info('实时修改模式已关闭')
        // 清空拦截队列
        interceptedSeqIDs.value.clear()
    }
}
```

**修改 startRecord 调用**：
```typescript
async function handleStartRecord() {
    // ... 初始化 recordData ...
    await recordControlService.startRecord(
        filePath.value, replayServerAddr.value, replayHttpAddr.value,
        filterMode.value  // 传递拦截模式状态
    )
    recordRunning.value = true
}
```

### 4.2 protocol-content.requirement.ts 修改

`startRecord` 方法签名增加 `filterMode` 参数：

```typescript
export interface RecordControlService {
    startRecord(filePath: string, serverAddr: string, httpAddr: string, filterMode: boolean): Promise<void>
    // ... 其他方法不变 ...
}

export function createWailsRecordControlService(): RecordControlService {
    return {
        async startRecord(filePath: string, serverAddr: string, httpAddr: string, filterMode: boolean): Promise<void> {
            await StartRecord(filePath, serverAddr, httpAddr, filterMode)
        },
        // ... 其他方法不变 ...
    }
}
```

### 4.3 message-table.vue 修改

**新增方法**：
```typescript
// 暴露给父组件，用于自动滚动到指定行
function scrollToRow(index: number) {
    // 获取表格行的 DOM 元素并滚动到可视区域
    const rowEl = tableRef.value?.$el.querySelector(`[data-row-index="${index}"]`)
    if (rowEl) {
        rowEl.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
}

defineExpose({ scrollToRow })
```

**被拦截消息的 UI 标记**：
```vue
<!-- 在 rowProps 中增加拦截状态样式 -->
<tr :style="{
    // ... 现有样式 ...
    borderLeft: isIntercepted(row) ? '3px solid #f0a020' : undefined,
    backgroundColor: isSelected ? 'var(--n-primary-color-hover)' : isIntercepted(row) ? 'rgba(240, 160, 32, 0.08)' : undefined,
}">
```

**Props 新增**：
```typescript
const props = defineProps<{
    // ... 现有 props ...
    interceptedSeqIDs?: Set<number>  // 被拦截的消息 SeqID 集合
}>()

function isIntercepted(row: any): boolean {
    return props.interceptedSeqIDs?.has(row.seq_id) ?? false
}
```

### 4.4 UI 状态说明

| 状态 | 视觉表现 | 含义 |
|------|---------|------|
| 普通行 | 默认背景 | 正常录制/重放的消息 |
| 被拦截行 | 左侧橙色边框 + 微黄背景 | 被拦截等待编辑的 Req |
| 选中行 | 主题色高亮背景 | 当前正在查看/编辑的行 |
| 已放行 | 边框消失，恢复普通行样式 | 拦截消息已发送给服务端 |

## 5. Wails Bindings 调整

### 5.1 需要重新生成的 Bindings

| 文件 | 变更内容 |
|------|---------|
| `recordcontrolservice.ts` | `StartRecord` 增加第 4 个参数 `filterMode: boolean` |

### 5.2 重新生成流程

```bash
# 1. 确认后端代码修改完成后，重新生成 TypeScript bindings
wails3 generate bindings -ts

# 2. 验证生成的 recordcontrolservice.ts 中 StartRecord 签名：
#    StartRecord(filePath: string, serverAddr: string, httpAddr: string, filterMode: boolean)

# 3. 同步更新 protocol-content.requirement.ts 的 startRecord 方法签名
```

### 5.3 无需 Bindings 改动的部分

- `record:intercepted` 事件：后端通过 `app.Event.Emit()` 推送，前端通过 `Events.On()` 监听，**不需要 Service 方法绑定**
- `SendMessages`：复用现有 `ReplayControlService` 方法，签名不变

## 6. 关键时序图

### 6.1 拦截 → 编辑 → 放行 完整流程

```
客户端    RecordWorker    packet-tab.vue    replay-control.vue    服务端
  │            │               │                    │               │
  │──Req─────>│               │                    │               │
  │           │──拦截，不转发──│                    │               │
  │           │               │                    │               │
  │           │──Event.Emit──>│                    │               │
  │           │ 'record:intercepted'                │               │
  │           │               │                    │               │
  │           │               │──追加到表格────────│               │
  │           │               │──自动选中─────────│               │
  │           │               │──滚动到可视区域────│               │
  │           │               │                    │               │
  │           │               │<──用户编辑 Payload─│               │
  │           │               │                    │               │
  │           │               │<──点击"重发"按钮───│               │
  │           │               │                    │               │
  │           │<──SendMessages──│                  │               │
  │           │   (修改后的 Req) │                  │               │
  │           │               │                    │               │
  │           │───────────────────────────────────────────────────>│
  │           │               │                    │    转发到服务端 │
  │           │               │                    │               │
  │           │               │──从拦截队列移除────│               │
  │           │               │──UI 恢复普通样式───│               │
  │           │               │                    │               │
  │<────────────────────────────Ack/Ntf（不拦截，直接转发）────────│
```

## 7. 错误处理

| 场景 | 处理策略 |
|------|---------|
| 拦截模式下客户端断开连接 | 清理该连接相关的拦截状态，已拦截但未放行的消息丢弃 |
| 用户关闭实时修改模式 | 清空拦截队列，后续消息不再拦截（但已拦截的仍需放行） |
| 停止录制 | 所有拦截状态清空，未放行消息丢弃 |
| 解码失败 | 原样转发，不拦截，记录日志 |
| 前端未响应（长时间不放行） | 消息一直阻塞在拦截状态，直到用户放行或断开连接 |

## 8. 测试要点

1. **开启拦截模式后，Req 消息是否被拦截**：验证消息不直接到达服务端
2. **Ack/Ntf 是否正常转发**：验证服务端响应能到达客户端
3. **拦截消息是否正确显示在表格中**：验证消息内容、SeqID、MsgName 正确
4. **自动选中是否生效**：验证新拦截消息自动选中并滚动到可视区域
5. **编辑后放行是否正确**：验证修改后的 Payload 发送到服务端
6. **拦截队列是否正确清理**：验证放行后 UI 标记消失
7. **关闭拦截模式后行为恢复**：验证后续消息正常转发
8. **多消息连续拦截**：验证多条消息依次拦截、依次处理

## 9. 相关文件索引

### 后端

| 文件 | 修改内容 |
|------|---------|
| `backend/pkg/proto-test/record_worker.go` | 新增 `filterMode` 字段、`interceptAndParse`、`emitInterceptedMessage` 方法 |
| `backend/pkg/proto-test/wails_record_control.go` | 已有 `filterMode` 参数，无需修改 |
| `backend/pkg/proto-test/streamproto/decoder.go` | 无需修改（复用 `DecodeFrame`） |

### 前端

| 文件 | 修改内容 |
|------|---------|
| `frontend/src/pages/stream-proxy/components/packet-tab.vue` | 监听 `record:intercepted` 事件，管理拦截队列，修改 `handleRetryMessage` |
| `frontend/src/pages/stream-proxy/components/protocol-content.requirement.ts` | `startRecord` 增加 `filterMode` 参数 |
| `frontend/src/pages/stream-proxy/components/message-table.vue` | 新增 `scrollToRow` 方法，拦截消息 UI 标记 |
| `frontend/src/pages/stream-proxy/components/replay-control.vue` | 无需修改（复用"重发"按钮） |

### Bindings

| 文件 | 变更 |
|------|------|
| `frontend/bindings/.../proto-test/recordcontrolservice.ts` | `StartRecord` 增加 `filterMode: boolean` 参数 |
