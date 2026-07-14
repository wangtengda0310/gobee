# params — proto-test 纯数据类型与算法

本包提供协议测试所需的纯数据类型和算法，不依赖 `msg` 包或 `variables` 包，避免循环导入。

## 核心类型

| 类型 | 位置 | 说明 |
|------|------|------|
| `RangeValue` | [types.go:13](types.go:13) | 范围输入值（start/step/end） |
| `FieldValues` | [types.go:21](types.go:21) | 字段 4 态值（original/range/enum/combo/variable） |
| `IterationConfig` | [iteration.go:11](iteration.go:11) | 内部迭代配置 |
| `VariableDef` | [variable_defs.go:11](variable_defs.go:11) | 变量定义（含提取函数、`AvailableReqs` 可见 Req 列表） |
| `VariableInfo` | [variable_defs.go:19](variable_defs.go:19) | 前端用变量信息（含 `available_reqs`） |
| `DecodedFrame` | [frame.go:7](frame.go:7) | 解码后的协议帧（供 variables 包使用） |

## 核心函数

| 函数 | 位置 | 说明 |
|------|------|------|
| `GenerateIterativeMessages` | [iteration.go:32](iteration.go:32) | 字段迭代展开算法 |
| `FieldValuesToIterationConfig` | [iteration.go:188](iteration.go:188) | 前端 FieldValues 转内部迭代配置 |
| `SetBuiltinVariables` | [variable_defs.go:29](variable_defs.go:29) | 设置内置变量列表 |
| `AppendBuiltinVariables` | [variable_defs.go:34](variable_defs.go:34) | 追加内置变量（测试用） |
| `FindVariableByShortName` | [variable_defs.go:56](variable_defs.go:56) | 按短名查找变量定义 |
| `StripByteStreamPrefix` | [payload.go:19](payload.go:19) | 剥离 ByteStream 2字节LE长度前缀 |
| `UnmarshalProtoPayload` | [payload.go:30](payload.go:30) | 反序列化 proto 消息（依赖注入的 `MessageFactory`） |

## MessageFactory 注入

`params.UnmarshalProtoPayload` 需要 `msg` 包注入 `MessageFactory`（即 `NewMessage`）。
注入点：[msg/variable.go:12](../msg/variable.go:12)

## 测试

| 文件 | 覆盖 |
|------|------|
| [iteration_test.go](iteration_test.go) | 迭代展开算法 |
