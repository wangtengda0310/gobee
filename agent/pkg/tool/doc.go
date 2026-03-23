// Package tool 提供 Agent 工具集成。
//
// 核心概念:
//   - Tool: 工具接口，定义可被 Agent 调用的能力
//   - FunctionTool: 函数工具，包装 Go 函数为工具
//   - Registry: 工具注册表，管理工具的注册和查找
//   - BatchExecutor: 批量执行器，并行执行多个工具调用
//
// 使用示例:
//
//	// 创建函数工具
//	searchTool := tool.NewFunction("search", "搜索网络信息",
//	    func(ctx context.Context, args map[string]any) (any, error) {
//	        query := args["query"].(string)
//	        return map[string]any{"results": []string{query}}, nil
//	    },
//	    tool.WithStringParam("query", "搜索查询", true),
//	)
//
//	// 注册到注册表
//	registry := tool.NewRegistry()
//	registry.MustRegister(searchTool)
//
//	// 执行工具
//	result, err := registry.Execute(ctx, "search", map[string]any{"query": "hello"})
//
// # 工具定义格式
//
// 工具参数使用 JSON Schema 格式定义:
//
//	tool.WithParameters(map[string]any{
//	    "type": "object",
//	    "properties": map[string]any{
//	        "query": map[string]any{
//	            "type":        "string",
//	            "description": "搜索查询",
//	        },
//	    },
//	    "required": []string{"query"},
//	})
//
// # 便捷方法
//
// Package 提供了便捷方法快速定义常见参数类型:
//
//	tool.WithStringParam("name", "名称", true)    // 必需的字符串参数
//	tool.WithNumberParam("count", "数量", false) // 可选的数字参数
//	tool.WithBooleanParam("verbose", "详细输出", false) // 可选的布尔参数
package tool
