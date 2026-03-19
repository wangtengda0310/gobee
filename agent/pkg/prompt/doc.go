// Package prompt 提供 AI Agent 提示词构建和管理工具。
//
// 本包提供以下核心功能：
//
//   - System Prompt 构建: 使用 SystemBuilder 链式构建系统提示词
//   - 对话历史管理: 使用 History 管理多轮对话消息
//   - 上下文截断: 支持多种策略（LLM摘要、滑动窗口等）
//   - 工具描述生成: 将 llm.Tool 转换为提示词格式
//   - 结构化输出: 生成 JSON/YAML 输出格式约束
//
// # 基本使用
//
// 构建系统提示词:
//
//	system := prompt.NewSystem("代码助手").
//	    WithDescription("你是一个专业的编程助手").
//	    WithConstraint("回答必须简洁准确").
//	    Build()
//
// 管理对话历史:
//
//	history := prompt.NewHistory()
//	history.AddUser("你好").
//	    AddAssistant("你好！有什么可以帮助你的？")
//	messages := history.ToMessages()
//
// 结构化输出:
//
//	format := prompt.JSONOutput(map[string]any{
//	    "type": "object",
//	    "properties": map[string]any{
//	        "result": map[string]any{"type": "string"},
//	    },
//	})
//	fmt.Println(format.ToPrompt())
package prompt
