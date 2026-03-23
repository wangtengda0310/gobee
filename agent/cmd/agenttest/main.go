// Package main 提供 Agent 集成测试
//
// 用法:
//
//	ANTHROPIC_API_KEY="xxx" go run ./cmd/agenttest \
//	  -model "glm-5" \
//	  -base-url "https://open.bigmodel.cn/api/anthropic/v1" \
//	  -prompt "现在几点了？"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/agent"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/llm/anthropic"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

func main() {
	// 命令行参数
	model := flag.String("model", "glm-5", "模型名称")
	baseURL := flag.String("base-url", "https://open.bigmodel.cn/api/anthropic/v1", "API 基础 URL")
	prompt := flag.String("prompt", "你好，请介绍一下你自己", "用户提示")
	maxLoops := flag.Int("max-loops", 5, "最大循环次数")
	systemPrompt := flag.String("system", "你是一个友好的AI助手", "系统提示词")
	flag.Parse()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置 ANTHROPIC_API_KEY 环境变量")
	}

	// 创建 LLM 客户端
	client, err := anthropic.NewClient(
		anthropic.WithAPIKey(apiKey),
		anthropic.WithModel(*model),
		anthropic.WithBaseURL(*baseURL),
	)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	// 创建工具
	timeTool := tool.NewFunction("get_current_time", "获取当前时间",
		func(ctx context.Context, args map[string]any) (any, error) {
			return map[string]any{
				"time":     time.Now().Format(time.RFC3339),
				"timezone": "UTC",
			}, nil
		},
		tool.WithDescription("返回当前的日期和时间"),
	)

	echoTool := tool.NewFunction("echo", "回显输入内容",
		func(ctx context.Context, args map[string]any) (any, error) {
			message, _ := args["message"].(string)
			return map[string]any{
				"echo":   message,
				"length": len(message),
			}, nil
		},
		tool.WithStringParam("message", "要回显的消息", true),
	)

	// 创建记忆
	mem := memory.NewSlidingWindow(50, memory.WithPreserveSystem(true))

	// 创建 Agent
	ag := agent.New(
		agent.WithLLM(client),
		agent.WithSystemPrompt(*systemPrompt),
		agent.WithTools(timeTool, echoTool),
		agent.WithMaxLoops(*maxLoops),
		agent.WithHooks(&agent.Hooks{
			OnStart: func(input string) {
				fmt.Printf("🚀 开始执行: %s\n\n", input)
			},
			OnLoop: func(loopCount int, state *agent.State) {
				fmt.Printf("📍 循环 %d\n", loopCount)
			},
			OnToolCall: func(name string, args map[string]any) {
				fmt.Printf("🔧 调用工具: %s, 参数: %v\n", name, args)
			},
			OnToolResult: func(result *tool.ToolResult) {
				if result.Error != nil {
					fmt.Printf("❌ 工具结果: %v\n", result.Error)
				} else {
					fmt.Printf("✅ 工具结果: %v\n", result.Result)
				}
			},
			OnLLMResponse: func(response *llm.ChatResponse) {
				fmt.Printf("🤖 LLM 响应: %s\n", response.Content)
				if len(response.ToolCalls) > 0 {
					fmt.Printf("   工具调用: %d 个\n", len(response.ToolCalls))
				}
			},
			OnDone: func(result *agent.Result) {
				fmt.Printf("\n✨ 执行完成\n")
				fmt.Printf("   循环次数: %d\n", result.LoopCount)
				fmt.Printf("   耗时: %v\n", result.Duration)
				if result.Usage != nil {
					fmt.Printf("   Token: 输入 %d, 输出 %d\n",
						result.Usage.InputTokens, result.Usage.OutputTokens)
				}
			},
		}),
	)

	// 注意：memory 暂未集成到 agent 中，这里仅展示创建
	_ = mem

	// 执行任务
	ctx := context.Background()
	result, err := ag.Run(ctx, *prompt)
	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📝 最终回复:")
	fmt.Println(result.Content)
}
