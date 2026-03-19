# prompt 包功能设计计划

## 背景

`agent/pkg/prompt` 包用于组织 AI Agent 提示词相关逻辑，提供工具方法帮助使用者快速构建 Agent 并便捷地组织提示词。

---

## 核心功能（第一期）

### 1. System Prompt 构建器 (`system.go`)

构建 Agent 的系统提示词，包括角色定义、能力描述、约束条件、工具描述、输出格式等。

```go
type SystemBuilder struct { ... }

func NewSystem(role string) *SystemBuilder
func (b *SystemBuilder) WithDescription(desc string) *SystemBuilder
func (b *SystemBuilder) WithConstraint(c string) *SystemBuilder
func (b *SystemBuilder) WithTools(tools ...*llm.Tool) *SystemBuilder
func (b *SystemBuilder) WithOutputFormat(format *OutputFormat) *SystemBuilder
func (b *SystemBuilder) Build() string
```

### 2. 对话历史管理 (`history.go`)

管理多轮对话的消息历史，支持上下文窗口控制。

```go
type History struct { ... }

func NewHistory() *History
func (h *History) AddUser(content string) *History
func (h *History) AddAssistant(content string) *History
func (h *History) AddToolResult(name, result string) *History
func (h *History) Truncate(ctx context.Context, llm llm.ChatCompleter, strategy TruncateStrategy) (*History, error)
func (h *History) ToMessages() []*llm.Message
```

**截断需要注入 `llm.ChatCompleter` 用于生成摘要。**

**截断策略（策略模式）：**
- ✅ 摘要压缩（默认，使用 LLM 生成摘要）
- 🔜 滑动窗口（TODO）
- 🔜 固定系统 + 滑动窗口（TODO）
- 🔜 Token 计数动态调整（TODO）

### 3. 工具描述生成器 (`tool.go`)

将 `llm.Tool` 定义转换为 LLM 可理解的提示词格式。

```go
func GenerateToolPrompt(tools ...*llm.Tool) string
func GenerateToolSchema(tools ...*llm.Tool) string  // JSON Schema 格式
```

**直接复用 `pkg/llm.Tool` 结构体，无需额外适配器。**

### 4. 结构化输出 (`output.go`)

生成 JSON/YAML 输出格式约束，确保 LLM 输出符合预期格式。

```go
type OutputFormat struct {
    Type   string         // "json", "yaml", "markdown"
    Schema map[string]any // JSON Schema
}

func JSONOutput(schema map[string]any) *OutputFormat
func YAMLOutput(schema map[string]any) *OutputFormat
func (f *OutputFormat) ToPrompt() string
```

---

## 目录结构

```
pkg/prompt/
├── doc.go              # 包文档（已存在，需更新）
├── system.go           # System Prompt 构建器
├── history.go          # 对话历史管理
├── truncate.go         # 截断策略接口和实现
├── tool.go             # 工具描述生成
├── output.go           # 结构化输出格式
└── example_test.go     # 使用示例
```

---

## 关键设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 截断策略 | 策略模式，默认 LLM 摘要 | 可扩展，后续支持多种策略对照测试 |
| 工具集成 | 复用 llm.Tool | 减少重复定义，保持一致性 |
| 模板语法 | 稍后讨论 | 可先实现基础功能 |

---

## 后期扩展（TODO）

- Token 计数估算
- Few-shot 示例管理
- 模板引擎
- 多语言/多风格支持
- 滑动窗口截断策略
- Token 计数动态截断

---

## 验证方式

1. 编写单元测试覆盖核心功能
2. 使用 `go test ./pkg/prompt/...` 运行测试
3. 创建示例代码验证与 `pkg/llm` 的集成

---

## 实现状态

| 功能 | 状态 | 说明 |
|------|------|------|
| SystemBuilder | ✅ 已完成 | 链式构建，支持 Clone/Reset |
| History | ✅ 已完成 | 支持工具调用，依赖注入摘要生成 |
| TruncateStrategy | ✅ 已完成 | 摘要、滑动窗口、固定系统三种策略 |
| Tool 描述生成 | ✅ 已完成 | 复用 llm.Tool |
| OutputFormat | ✅ 已完成 | JSON/YAML/Markdown |
| 单元测试 | ✅ 已完成 | 14 个测试用例 |
