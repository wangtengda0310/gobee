package tool

// ToolResult 工具执行结果
// 封装了工具调用的结果信息
type ToolResult struct {
	// ToolCallID 工具调用唯一标识
	// 用于关联 LLM 返回的 tool_call
	ToolCallID string

	// Name 工具名称
	Name string

	// Result 执行结果
	// 可以是任意 JSON 可序列化的值
	Result any

	// Error 执行错误
	// 如果执行成功则为 nil
	Error error
}

// IsError 检查是否有错误
func (r *ToolResult) IsError() bool {
	return r.Error != nil
}

// ToMap 将结果转换为 map 格式
// 便于序列化为 JSON
func (r *ToolResult) ToMap() map[string]any {
	m := map[string]any{
		"tool_call_id": r.ToolCallID,
		"name":         r.Name,
	}
	if r.Error != nil {
		m["error"] = r.Error.Error()
		m["success"] = false
	} else {
		m["result"] = r.Result
		m["success"] = true
	}
	return m
}

// BatchResult 批量执行结果
type BatchResult struct {
	// Results 所有工具执行结果
	Results []*ToolResult

	// SuccessCount 成功执行的数量
	SuccessCount int

	// ErrorCount 执行失败的数量
	ErrorCount int
}

// HasErrors 检查是否有错误
func (r *BatchResult) HasErrors() bool {
	return r.ErrorCount > 0
}

// GetByToolCallID 根据 ToolCallID 获取结果
func (r *BatchResult) GetByToolCallID(id string) *ToolResult {
	for _, res := range r.Results {
		if res.ToolCallID == id {
			return res
		}
	}
	return nil
}

// GetByName 根据工具名称获取结果
func (r *BatchResult) GetByName(name string) []*ToolResult {
	var results []*ToolResult
	for _, res := range r.Results {
		if res.Name == name {
			results = append(results, res)
		}
	}
	return results
}
