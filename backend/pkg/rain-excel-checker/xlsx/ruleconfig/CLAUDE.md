# ruleconfig 包 — 校验规则配置的 JSON 序列化与反序列化工具

校验规则配置的持久化存储和加载功能。支持将 SheetRule 结构体序列化为 JSON 文件存储，并从 JSON 文件反序列化重建规则对象。采用并发处理优化批量 IO 性能。

## 文件结构

```
ruleconfig/
└── excel_check_io.go    # 规则配置的保存与加载核心实现
```

## 核心组件

### excel_check_io.go — 规则配置 IO 操作

| 导出函数 | 说明 |
|----------|------|
| `SaveCheck()` | 并发保存校验规则到 JSON 文件，自动创建目录，处理 Sheet 名称中的特殊字符 |
| `LoadJsonRules()` | 并发加载多个规则 JSON 文件，保持原始顺序，支持数字类型的精确解析 |

## 包依赖

### 依赖
- `json_rule` — SheetRule 类型定义和规则数据结构

### 被依赖
- `engine/executor.go` — 检查执行器，根据模式调用 `LoadJsonRules` 加载规则
- `main.go` — 程序入口，使用 `LoadJsonRules` 加载校验规则配置
- `xlsx/tests/` — 测试用例，同时使用 `SaveCheck` 和 `LoadJsonRules`

## 关键行为

- **并发处理**：`LoadJsonRules` 使用 goroutine 并发读取多个规则文件，通过通道收集结果
- **数字精度保留**：使用 `json.Decoder` 的 `UseNumber()` 模式，确保大数字（如 int64 ID）不丢失精度
- **Sheet 名称特殊字符处理**：文件名中的 `/` 等特殊字符会被替换为安全字符
- **目录自动创建**：`SaveCheck` 会在目标目录不存在时自动创建
