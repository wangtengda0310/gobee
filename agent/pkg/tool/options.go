package tool

// FunctionOption 函数工具配置选项
// 用于配置 FunctionTool 的行为
type FunctionOption func(*FunctionTool)

// WithDescription 设置工具描述
func WithDescription(desc string) FunctionOption {
	return func(f *FunctionTool) {
		f.description = desc
	}
}

// WithParameters 设置工具参数定义
// params 应该是符合 JSON Schema 格式的 map
// 示例:
//
//	WithParameters(map[string]any{
//	    "type": "object",
//	    "properties": map[string]any{
//	        "query": map[string]any{
//	            "type":        "string",
//	            "description": "搜索查询",
//	        },
//	    },
//	    "required": []string{"query"},
//	})
func WithParameters(params map[string]any) FunctionOption {
	return func(f *FunctionTool) {
		f.parameters = params
	}
}

// WithStringParam 添加字符串类型参数
// 这是一个便捷方法，用于快速定义简单参数
func WithStringParam(name, description string, required bool) FunctionOption {
	return func(f *FunctionTool) {
		if f.parameters == nil {
			f.parameters = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			}
		}

		props, _ := f.parameters["properties"].(map[string]any)
		if props == nil {
			props = make(map[string]any)
			f.parameters["properties"] = props
		}

		props[name] = map[string]any{
			"type":        "string",
			"description": description,
		}

		if required {
			req, _ := f.parameters["required"].([]string)
			f.parameters["required"] = append(req, name)
		}
	}
}

// WithNumberParam 添加数字类型参数
func WithNumberParam(name, description string, required bool) FunctionOption {
	return func(f *FunctionTool) {
		if f.parameters == nil {
			f.parameters = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			}
		}

		props, _ := f.parameters["properties"].(map[string]any)
		if props == nil {
			props = make(map[string]any)
			f.parameters["properties"] = props
		}

		props[name] = map[string]any{
			"type":        "number",
			"description": description,
		}

		if required {
			req, _ := f.parameters["required"].([]string)
			f.parameters["required"] = append(req, name)
		}
	}
}

// WithBooleanParam 添加布尔类型参数
func WithBooleanParam(name, description string, required bool) FunctionOption {
	return func(f *FunctionTool) {
		if f.parameters == nil {
			f.parameters = map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			}
		}

		props, _ := f.parameters["properties"].(map[string]any)
		if props == nil {
			props = make(map[string]any)
			f.parameters["properties"] = props
		}

		props[name] = map[string]any{
			"type":        "boolean",
			"description": description,
		}

		if required {
			req, _ := f.parameters["required"].([]string)
			f.parameters["required"] = append(req, name)
		}
	}
}
