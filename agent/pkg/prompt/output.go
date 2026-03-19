package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat 定义结构化输出格式
type OutputFormat struct {
	// Type 输出类型: "json", "yaml", "markdown"
	Type string

	// Schema JSON Schema 定义
	Schema map[string]any
}

// JSONOutput 创建 JSON 输出格式约束
func JSONOutput(schema map[string]any) *OutputFormat {
	return &OutputFormat{
		Type:   "json",
		Schema: schema,
	}
}

// YAMLOutput 创建 YAML 输出格式约束
func YAMLOutput(schema map[string]any) *OutputFormat {
	return &OutputFormat{
		Type:   "yaml",
		Schema: schema,
	}
}

// MarkdownOutput 创建 Markdown 输出格式约束
// schema 可用于定义章节结构
func MarkdownOutput(schema map[string]any) *OutputFormat {
	return &OutputFormat{
		Type:   "markdown",
		Schema: schema,
	}
}

// ToPrompt 将输出格式转换为提示词
// 根据类型生成不同的格式说明：
// - json: 包含 JSON Schema 和解析要求
// - yaml: 包含数据结构定义和解析要求
// - markdown: 包含描述和章节结构建议
func (f *OutputFormat) ToPrompt() string {
	var sb strings.Builder

	switch f.Type {
	case "json":
		sb.WriteString("你的输出必须是有效的 JSON 格式。\n")
		if f.Schema != nil {
			sb.WriteString("JSON Schema:\n")
			schemaJSON, _ := json.MarshalIndent(f.Schema, "", "  ")
			sb.WriteString("```json\n")
			sb.WriteString(string(schemaJSON))
			sb.WriteString("\n```\n")
		}
		sb.WriteString("确保输出可以被 JSON 解析器正确解析。")

	case "yaml":
		sb.WriteString("你的输出必须是有效的 YAML 格式。\n")
		if f.Schema != nil {
			sb.WriteString("数据结构定义:\n")
			schemaJSON, _ := json.MarshalIndent(f.Schema, "", "  ")
			sb.WriteString("```json\n")
			sb.WriteString(string(schemaJSON))
			sb.WriteString("\n```\n")
		}
		sb.WriteString("确保输出可以被 YAML 解析器正确解析。")

	case "markdown":
		sb.WriteString("你的输出必须是 Markdown 格式。\n")
		if f.Schema != nil {
			if desc, ok := f.Schema["description"].(string); ok {
				sb.WriteString(desc)
				sb.WriteString("\n")
			}
			if sections, ok := f.Schema["sections"].([]string); ok {
				sb.WriteString("建议的章节结构:\n")
				for _, s := range sections {
					fmt.Fprintf(&sb, "- %s\n", s)
				}
			}
		}
	}

	return sb.String()
}

// String 实现 Stringer 接口
func (f *OutputFormat) String() string {
	return f.ToPrompt()
}
