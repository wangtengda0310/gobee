// Package tool 提供 Agent 工具集成。
//
// 内置工具:
//   - 代码执行 (Code Execution)
//   - 文件操作 (File Operations)
//   - 搜索集成 (Search)
//   - Web 访问 (HTTP Client)
//
// 自定义工具:
//
//	tool := tool.NewFunction(
//	    "search",
//	    "搜索网络信息",
//	    func(ctx context.Context, args map[string]any) (any, error) {
//	        // 实现逻辑
//	    },
//	)
package tool
