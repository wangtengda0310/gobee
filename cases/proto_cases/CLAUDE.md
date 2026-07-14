# cases/proto_cases — 协议测试用例

Proto 测试用例目录，存储通过"发包改包"页面多选保存的协议消息用例。

## 文件格式

每个用例是一个 JSON 文件，格式为 `Recording`（兼容 `LoadRecordFile` 和 `StartReplay`）：
- 版本号由 `cases.RecordingVersion` 常量定义（当前值为 1，位置：[record.go](../../backend/pkg/proto-test/cases/record.go:14)）
- 修改 `Recording` 结构体时需同步递增版本号并更新本说明

```json
{
  "version": 1,
  "recorded_at": "2026-06-03T08:00:00+08:00",
  "server_addr": "10.0.0.1:18000",
  "login_payload_b64": "...",
  "messages": [
    {
      "msg_id": 1001,
      "msg_name": "HelloReq",
      "seq_id": 0,
      "descript": "可选描述",
      "payload_json": {"name": "test1"}
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `version` | number | 文件版本号，由 `cases.RecordingVersion` 定义 |
| `recorded_at` | string | ISO8601 录制时间 |
| `server_addr` | string | 目标服务器地址 |
| `login_payload_b64` | string | (可选) LoginReq 解密后 payload |
| `messages` | array | 消息列表 |

### Message 字段（仅 Req）

用例文件**只保存客户端请求（Req）**，不保存 Ack/Ntf。加载时后端自动补全 `direction="→"` 供重放。

| 字段 | 类型 | 说明 |
|------|------|------|
| `msg_id` | number | 消息 ID |
| `msg_name` | string | 消息名称 |
| `seq_id` | number | 序列号 |
| `descript` | string | (可选) 用例描述，测试用例页签「描述」列展示 |
| `payload_json` | object | Proto 消息反序列化后的 JSON 对象 |
| `field_values` | object | (可选) 字段 4 态迭代配置，见下方说明 |

### field_values 4 态配置

`field_values` 为字段级配置，每个 key 对应消息 `payload_json` 中的字段名，value 结构如下：

```json
{
  "range_value": {"start": 0, "step": 1, "end": 10},
  "enum_value": [],
  "combo_value": [],
  "input_type": "variable",
  "variable_name": "cityId"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `input_type` | string | `"original" \| "range" \| "enum" \| "combo" \| "variable"` |
| `variable_name` | string | `input_type="variable"` 时必填，如 `cityId`、`openid` |

#### 内置变量

| 变量名 | 说明 | 示例场景 |
|--------|------|----------|
| `cityId` | 工会城战可选城池 ID | `TeamSelectGuildCityReq.cityId` |
| `openid` | 当前发送账号（如 `test1`） | 将玩家账号传入请求字段 |

变量实现位于 [`pkg/proto-test/variables/`](../../backend/pkg/proto-test/variables/CLAUDE.md)，注册表位于 `params` 包。

## 存储位置

路径：`cases/proto_cases/{name}.json`

## 相关代码

### 后端保存逻辑

核心方法：[`SaveTestCase()`](../backend/pkg/proto-test/service_test_case.go:54)

将前端 `RecordFileData` 转换为 `cases.Recording` 后，经 `SaveTestCaseToFile()` 过滤仅 Req 并省略 `direction`/`offset_ms` 写入文件。

### 后端加载逻辑

核心方法：[`LoadRecordFile()`](../backend/pkg/proto-test/service_record_file.go:20)

从文件读取后经 `LoadTestCaseFromFile()` 过滤 Ack/Ntf 并补全 direction，再通过 [`recordingToViews()`](../backend/pkg/proto-test/wails.go:48) 转换为前端视图格式。

### 4态值转换

核心方法：[`convertFieldValues()`](../backend/pkg/proto-test/wails.go:69)

将 `map[string]any` 转换为 `map[string]FieldValues`。

### 前端状态管理

[`useRecordData()`](../frontend/src/pages/stream-proxy/composables/use-record-data.ts:209) — 模块级单例状态。

### 数据流向

```
写入：前端 RecordFileData → SaveTestCase() → SaveTestCaseToFile() → JSON（仅 Req）
读取：JSON → LoadTestCaseFromFile() → NormalizeTestCaseRecording() → recordingToViews() → 前端 RecordFileData
```

## 关键文件索引

| 功能 | 文件路径 | 关键方法/类型 |
|------|----------|---------------|
| **写入用例** | `pkg/proto-test/service_test_case.go` | `SaveTestCase()` |
| **读取文件** | `pkg/proto-test/cases/testcase.go` | `LoadTestCaseFromFile()` |
| **写入用例精简格式** | `pkg/proto-test/cases/testcase.go` | `SaveTestCaseToFile()` |
| **转换视图** | `pkg/proto-test/wails.go` | `recordingToViews()` |
| **4态值转换** | `pkg/proto-test/wails.go` | `convertFieldValues()` |
| **前端状态** | `frontend/src/pages/stream-proxy/composables/use-record-data.ts` | `useRecordData()` |
| **数据结构** | `pkg/proto-test/wails.go` | `RecordFileData`, `RecordEntryView` |
