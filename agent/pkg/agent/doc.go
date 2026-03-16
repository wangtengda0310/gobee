// Package agent 提供 AI Agent 开发框架。
//
// 核心概念:
//   - Agent: 具有角色和能力的智能体
//   - Tool: Agent 可调用的工具
//   - Loop: Agent 执行循环 (感知-思考-行动)
//
// 使用示例:
//
//	agent := agent.New(
//	    agent.WithLLM(llm),
//	    agent.WithTools(tools...),
//	)
//	response, err := agent.Run(ctx, "帮我分析这个文件")
package agent
