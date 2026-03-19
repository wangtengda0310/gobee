package prompt

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// GenerateToolPrompt 将工具定义转换为人类可读的提示词格式
// 输出格式：
//
//	## 可用工具
//
//	### tool_name
//	工具描述
//
//	参数:
//	- param_name (必需): 参数描述 [类型]
func GenerateToolPrompt(tools ...*llm.Tool) string {
	if len(tools) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## 可用工具\n\n")

	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		fn := tool.Function
		fmt.Fprintf(&sb, "### %s\n", fn.Name)
		if fn.Description != "" {
			sb.WriteString(fn.Description)
			sb.WriteString("\n")
		}
		// 解析 JSON Schema 格式的参数定义
		// 提取 required 字段用于标记必需参数
		if len(fn.Parameters) > 0 {
			sb.WriteString("\n参数:\n")
			if props, ok := fn.Parameters["properties"].(map[string]interface{}); ok {
				required := make(map[string]bool)
				if req, ok := fn.Parameters["required"].([]interface{}); ok {
					for _, r := range req {
						if s, ok := r.(string); ok {
							required[s] = true
						}
					}
				}
				for name, prop := range props {
					if p, ok := prop.(map[string]interface{}); ok {
						reqMark := ""
						if required[name] {
							reqMark = " (必需)"
						}
						desc := ""
						if d, ok := p["description"].(string); ok {
							desc = d
						}
						typ := ""
						if t, ok := p["type"].(string); ok {
							typ = t
						}
						fmt.Fprintf(&sb, "- %s%s: %s [%s]\n", name, reqMark, desc, typ)
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// GenerateToolSchema 将工具定义转换为 JSON Schema 格式的提示词
func GenerateToolSchema(tools ...*llm.Tool) string {
	if len(tools) == 0 {
		return ""
	}

	schemas := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		schema := map[string]interface{}{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
			"parameters":  tool.Function.Parameters,
		}
		schemas = append(schemas, schema)
	}

	jsonBytes, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		return ""
	}

	return fmt.Sprintf("## 工具定义\n\n```json\n%s\n```", string(jsonBytes))
}

// GenerateToolCallPrompt 生成工具调用格式说明
func GenerateToolCallPrompt() string {
	return `## 工具调用格式

当需要调用工具时，请使用以下 JSON 格式:

` + "```json" + `
{
  "tool_calls": [
    {
      "name": "工具名称",
      "arguments": {
        "参数名": "参数值"
      }
    }
  ]
}
` + "```" + `

可以一次调用多个工具，它们会按顺序执行。`
}
