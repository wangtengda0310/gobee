# variables — 协议测试内置变量

本包实现 proto-test 模块的内置变量，并在 `init()` 中注册到 `params` 包。

## 目录文件

| 文件 | 职责 |
|------|------|
| [registry.go](registry.go) | `init()` 注册 `cityId` / `roomCreator` / `roomID` / `openid` 到 `params.SetBuiltinVariables` |
| [cityid.go](cityid.go) | `ExtractGuildCityID` 从 `PveGuildCityDataNtf` / `TransportRawNtf` 提取城池 ID |
| [roomlist.go](roomlist.go) | `ExtractRoomCreator` / `ExtractRoomID` 从 `NewGetRoomListAck` 提取首个房间信息 |

## 核心类型与函数

| 符号 | 位置 | 说明 |
|------|------|------|
| `ExtractGuildCityID` | [cityid.go:14](cityid.go:14) | 从解码帧中提取可选中的城池 ID |
| `ExtractRoomCreator` | [roomlist.go:14](roomlist.go:14) | 从 `NewGetRoomListAck` 提取首个房间的 `Creator` |
| `ExtractRoomID` | [roomlist.go:25](roomlist.go:25) | 从 `NewGetRoomListAck` 提取首个房间的 `RoomID` |
| `parseNewGetRoomListAck` | [roomlist.go:36](roomlist.go:36) | 从解码帧中解析 `NewGetRoomListAck` |

## 房间列表变量

`roomCreator` 与 `roomID` 监听 `EGameMsgID_NewGetRoomListAck_id`（2255），从 `Infos[0]` 中分别提取 `Creator` 与 `RoomID`。
当消息 ID 不匹配时返回 `nil`；当 `Infos` 为空时返回错误，与 `cityId` 的行为保持一致。

## 账号级变量 openid

`openid` 不需要 Ntf 提取，由 [msg/replay.go](../msg/replay.go:290) 在 `prepareVariableContext` 中预置为当前 `accountID`。

## 变量按 Req 过滤（AvailableReqs）

每个变量可声明 `AvailableReqs []string`（proto 消息名 `msg_name`），限制它在前端卡片模式的变量选择器中对哪些 Req 可见；`nil`/空表示对所有 Req 可用（如账号级变量 `openid`）。前端 `variable-select.vue` 拿到全量 `VariableInfo`（含 `available_reqs`）后按当前 Req 的 `msg_name` 过滤，避免给无关 Req 暴露不相关的变量选项。

| 变量 | AvailableReqs |
|------|---------------|
| `cityId` | `["TeamSelectGuildCityReq"]` |
| `roomCreator` | `["RoomLookOnReq"]` |
| `roomID` | `["RoomLookOnReq"]` |
| `openid` | `nil`（全可用） |

## 依赖关系

- 依赖 `params` 包：注册表、类型别名、`DecodedFrame`、`UnmarshalProtoPayload`
- 不依赖 `msg` 包，避免循环导入
- `msg` 包在 `variable.go` 的 `init()` 中向 `params.MessageFactory` 注入 `NewMessage`，供本包反序列化 proto 消息

## 测试

| 文件 | 覆盖 |
|------|------|
| [cityid_test.go](cityid_test.go) | `ExtractGuildCityID` 各场景 |
| [roomlist_test.go](roomlist_test.go) | `ExtractRoomCreator` / `ExtractRoomID` 提取、非目标消息、空列表、错误帧类型、注册表查找 |
| [registry_test.go](registry_test.go) | 变量注册表存在性、`AvailableReqs` 注册值与下发 |
