// Package agent 提供 AI Agent 开发框架。
//
// 核心概念:
//   - Agent: 具有角色和能力的智能体
//   - Tool: Agent 可调用的工具
//   - Loop: Agent 执行循环 (感知-思考-行动)
//
// 执行流程:
//
//	用户输入 → 构建请求 → LLM 调用 → 解析响应
//	    ↑                              ↓
//	    ←←←←← 有工具调用 ←←← 执行工具 ←←←
//
// 使用示例:
//
//	// 创建 Agent
//	agent := agent.New(
//	    agent.WithLLM(llmClient),
//	    agent.WithSystemPrompt("你是一个游戏QA助手"),
//	    agent.WithTools(
//	        tool.NewFunction("get_time", "获取当前时间",
//	            func(ctx context.Context, args map[string]any) (any, error) {
//	                return time.Now().Format(time.RFC3339), nil
//	            }),
//	    ),
//	    agent.WithHooks(&agent.Hooks{
//	        OnToolCall: func(name string, args map[string]any) {
//	            log.Printf("[调用] %s: %v", name, args)
//	        },
//	    }),
//	)
//
//	// 执行任务
//	result, err := agent.Run(ctx, "现在几点了？")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Content)
package agent
