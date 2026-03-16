package llm

// Tool 表示可用的工具定义
type Tool struct {
	// Type 工具类型，目前只有 "function"
	Type string `json:"type"`

	// Function 函数定义
	Function *FunctionDef `json:"function"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	// Name 函数名称
	Name string `json:"name"`

	// Description 函数描述
	Description string `json:"description"`

	// Parameters 函数参数 (JSON Schema)
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolCall 表示模型发起的工具调用
type ToolCall struct {
	// ID 工具调用唯一标识
	ID string `json:"id"`

	// Type 工具类型，目前只有 "function"
	Type string `json:"type"`

	// Function 函数调用
	Function *FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	// Name 函数名称
	Name string `json:"name"`

	// Arguments 函数参数 (JSON 字符串)
	Arguments string `json:"arguments"`
}

// NewTool 创建工具定义
func NewTool(name, description string, parameters map[string]interface{}) *Tool {
	return &Tool{
		Type: "function",
		Function: &FunctionDef{
			Name:        name,
			Description: description,
			Parameters:  parameters,
		},
	}
}

// NewToolCall 创建工具调用
func NewToolCall(id, name, arguments string) *ToolCall {
	return &ToolCall{
		ID:   id,
		Type: "function",
		Function: &FunctionCall{
			Name:      name,
			Arguments: arguments,
		},
	}
}
