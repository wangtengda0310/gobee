# streamproxy 设计决策

## 实时拦截改包模式

**拦截条件**：
- 仅拦截 **MsgID >= 1000** 的 Proto 消息（C→S 方向）
- 框架消息（MsgID < 1000）不拦截，原样转发
- Ack/Ntf 消息（S→C 方向）不拦截，原样转发

**拦截与录制的共存关系**：
- 拦截模式开启时，录制功能仍然正常工作
- 被拦截的 Req 消息会先录制到文件，再通过 `record:intercepted` 事件推送到前端
- 用户通过"重发"按钮放行时，复用 `ReplayControlService.SendMessages` 发送修改后的消息

## 数据契约原则（强制）

**一份数据结构，零冗余转换。**

| 类型 | 包 | 职责 | payload 字段 |
|------|-----|------|-------------|
| `RecordEntry` | streamproto | 存储/序列化（JSON 文件、TCP 帧） | `PayloadJSON json.RawMessage` |
| `RecordEntryView` | streamproxy | 前后端合约（Wails 返回值 + Event.Emit） | `Payload map[string]any` |

**唯一的转换点**：`singleEntryToView(entry, index) RecordEntryView`（定义在 `wails.go`）

**强制规则**：
1. **Event.Emit 禁止手动构建 `map[string]any`**：必须传递 `RecordEntryView` 结构体
2. **前端事件接收必须使用 `RecordEntryView.createFrom()`**
3. **新增字段只需改两处**：`RecordEntry` + `RecordEntryView`，`singleEntryToView` 中补充映射

**拦截事件的例外**：
`record:intercepted` 事件允许使用 `map[string]any` 包装 `RecordEntryView` + `conn_id`：
```go
data := map[string]any{
    "latest_msg": entryView,  // RecordEntryView 结构体
    "conn_id":    connID,     // 额外字段：连接标识
}
```

## Recording 版本号管理

录制文件版本号定义为 `cases.RecordingVersion` 常量（`cases/record.go:14`，`streamproto/record.go` 通过类型别名导出）。

**修改 Recording 结构体时的必检清单**：
1. 递增 `streamproto.RecordingVersion` 常量
2. 检查所有使用 `RecordingVersion` 的地方是否已同步
3. 同步更新 `cases/proto_cases/CLAUDE.md` 中的版本号和字段说明
4. 检查 `proto_cases/` 下已有 JSON 文件的兼容性

## 重放架构

**两个独立的重放操作，共享底层 `SendMessages`**：

```
顶部"开始重放"/"执行用例"  ──→ SendMessages(全部Req, repeatCount=1, rangeStart, rangeEnd)
底部"重发"面板（选中行）   ──→ SendMessages([选中Req], repeatCount=N, 1, 1)
```

`SendMessages` 流程（外层账号范围迭代 + 内层单账号重放）：
```
for i := rangeStart; i <= rangeEnd; i++:
    accountID := fmt.Sprintf("%s%d", openID, i)
    HTTP /authlogin 登录 → TCP 连接 → 发送 LoginReq
    → 等待推送完成 → 逐条发送 Req → readDrainer 接收 Ack/Ntf
    → 等待最后一批 Ack → 关闭连接
```

## 协议格式

### TCP 帧结构（Little-Endian）

```
帧头（4字节）:
  Byte 0-2: 消息长度（3B LE，值 = 总帧长 - 4）
  Byte 3:   标志位（Bit0=CompressFlag, Bit1=EncryptFlag）

消息体:
  Byte 0-1: MsgID（2B LE）
  Byte 2-5: SeqID（4B LE）
  Byte 6+:  Payload
```

### 消息分类

- MsgID < 1000：框架内部消息（LoginReq=1, LoginResp=2, Ping=3, Pong=4 等）
- MsgID >= 1000：Proto 消息（由 .pb.go 中的 EGameMsgID 定义，800+ 条）

### ByteStream 序列化格式（框架消息，MsgID < 1000）

服务端使用反射按 struct 字段顺序序列化（`go-service/base/serializer/serializerNew.go`）。

#### LoginReq 字段顺序

| 字段 | 类型 | 说明 |
|------|------|------|
| Account | string `[2B len][data]` | 账号 |
| Token | string `[2B len][data]` | HTTP 登录获取的 token |
| UID | uint64 `8B LE` | 玩家 ID（首次登录为 0） |
| Version | string `[2B len][data]` | 客户端版本号 |
| Metadata | map `[2B count][[key str][val uint64]]...` | 扩展元数据 |
| ExtData | []byte `[2B len][data]` | 设备信息 JSON 等 |
| ReqType | uint16 `2B LE` | 0=正常登录, 1=断线重连 |
| SeqID | uint32 `4B LE` | 序列号 |
| Entity | string `[2B len][data]` | SDK 验证数据 |
| Sign | string `[2B len][data]` | 签名 |

#### LoginResp 字段顺序

| 字段 | 类型 | 说明 |
|------|------|------|
| UID | uint64 `8B LE` | 玩家 ID |
| Result | uint32 `4B LE` | 0=成功 |
| ErrStr | string `[2B len][data]` | 错误信息 |
| Metadata | map `[2B count]...` | 含 Version 等 |
| ExtData | []byte `[2B len][data]` | 扩展数据 |
| ZoneID | int32 `4B LE` | 区服编号 |
| HttpToken | string `[2B len][data]` | HTTP token |

**注意**：LoginResp 没有 `[2B len]` 前缀，直接是 ByteStream 字段。Result 在 offset 8（UID 8 字节之后）。
