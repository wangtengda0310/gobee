package prompt

import (
	"strings"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// SystemBuilder 用于构建 Agent 的系统提示词
// 支持链式调用，方便组合各个部分
type SystemBuilder struct {
	role        string
	description string
	capabilities []string
	constraints  []string
	tools       []*llm.Tool
	outputFormat *OutputFormat
	examples    []string
	context     string
}

// NewSystem 创建一个新的系统提示词构建器
// role: Agent 的角色名称（如"代码助手"、"数据分析师"）
func NewSystem(role string) *SystemBuilder {
	return &SystemBuilder{
		role:         role,
		capabilities: make([]string, 0),
		constraints:  make([]string, 0),
		tools:        make([]*llm.Tool, 0),
		examples:     make([]string, 0),
	}
}

// WithDescription 设置 Agent 的详细描述
func (b *SystemBuilder) WithDescription(desc string) *SystemBuilder {
	b.description = desc
	return b
}

// WithCapabilities 添加 Agent 的能力描述
func (b *SystemBuilder) WithCapabilities(capabilities ...string) *SystemBuilder {
	b.capabilities = append(b.capabilities, capabilities...)
	return b
}

// WithConstraint 添加约束条件
func (b *SystemBuilder) WithConstraint(c string) *SystemBuilder {
	b.constraints = append(b.constraints, c)
	return b
}

// WithConstraints 批量添加约束条件
func (b *SystemBuilder) WithConstraints(constraints ...string) *SystemBuilder {
	b.constraints = append(b.constraints, constraints...)
	return b
}

// WithTools 添加可用工具
func (b *SystemBuilder) WithTools(tools ...*llm.Tool) *SystemBuilder {
	b.tools = append(b.tools, tools...)
	return b
}

// WithOutputFormat 设置输出格式约束
func (b *SystemBuilder) WithOutputFormat(format *OutputFormat) *SystemBuilder {
	b.outputFormat = format
	return b
}

// WithExample 添加示例（Few-shot）
func (b *SystemBuilder) WithExample(example string) *SystemBuilder {
	b.examples = append(b.examples, example)
	return b
}

// WithContext 设置额外上下文信息
func (b *SystemBuilder) WithContext(ctx string) *SystemBuilder {
	b.context = ctx
	return b
}

// Build 构建最终的系统提示词
// 输出格式为 Markdown，包含以下可选部分：
// - 角色定义（必需）
// - 描述
// - 能力列表
// - 约束条件
// - 工具描述（调用 GenerateToolPrompt）
// - 输出格式
// - 示例
// - 上下文
func (b *SystemBuilder) Build() string {
	var sb strings.Builder

	// 角色定义 - 所有系统提示词的核心部分
	sb.WriteString("# 角色定义\n\n")
	sb.WriteString("你是一个")
	sb.WriteString(b.role)
	sb.WriteString("。\n")

	// 详细描述
	if b.description != "" {
		sb.WriteString("\n## 描述\n\n")
		sb.WriteString(b.description)
		sb.WriteString("\n")
	}

	// 能力列表
	if len(b.capabilities) > 0 {
		sb.WriteString("\n## 能力\n\n")
		for _, cap := range b.capabilities {
			sb.WriteString("- ")
			sb.WriteString(cap)
			sb.WriteString("\n")
		}
	}

	// 约束条件
	if len(b.constraints) > 0 {
		sb.WriteString("\n## 约束条件\n\n")
		for _, c := range b.constraints {
			sb.WriteString("- ")
			sb.WriteString(c)
			sb.WriteString("\n")
		}
	}

	// 工具描述
	if len(b.tools) > 0 {
		sb.WriteString("\n")
		sb.WriteString(GenerateToolPrompt(b.tools...))
	}

	// 输出格式
	if b.outputFormat != nil {
		sb.WriteString("\n## 输出格式\n\n")
		sb.WriteString(b.outputFormat.ToPrompt())
		sb.WriteString("\n")
	}

	// 示例
	if len(b.examples) > 0 {
		sb.WriteString("\n## 示例\n\n")
		for _, ex := range b.examples {
			sb.WriteString(ex)
			sb.WriteString("\n\n")
		}
	}

	// 额外上下文
	if b.context != "" {
		sb.WriteString("\n## 上下文\n\n")
		sb.WriteString(b.context)
		sb.WriteString("\n")
	}

	return sb.String()
}

// String 实现 Stringer 接口，等同于 Build()
func (b *SystemBuilder) String() string {
	return b.Build()
}

// Reset 重置构建器，保留角色定义
func (b *SystemBuilder) Reset() *SystemBuilder {
	b.description = ""
	b.capabilities = make([]string, 0)
	b.constraints = make([]string, 0)
	b.tools = make([]*llm.Tool, 0)
	b.outputFormat = nil
	b.examples = make([]string, 0)
	b.context = ""
	return b
}

// Clone 克隆当前构建器
func (b *SystemBuilder) Clone() *SystemBuilder {
	newBuilder := NewSystem(b.role)
	newBuilder.description = b.description
	newBuilder.capabilities = append([]string{}, b.capabilities...)
	newBuilder.constraints = append([]string{}, b.constraints...)
	newBuilder.tools = append([]*llm.Tool{}, b.tools...)
	newBuilder.outputFormat = b.outputFormat
	newBuilder.examples = append([]string{}, b.examples...)
	newBuilder.context = b.context
	return newBuilder
}
