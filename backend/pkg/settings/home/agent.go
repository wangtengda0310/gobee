package home

import (
	"context"
	"fmt"
	"log"

	"github.com/wangtengda0310/gobee/agent/pkg/agent"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// EventEmitter 事件发射器接口
// *application.EventManager 隐式实现此接口（签名完全一致）
// 用于解耦 ChatService 对 Wails application.App 的直接依赖
type EventEmitter interface {
	Emit(name string, data ...any) bool
}

// noopEventEmitter 空实现，app 为 nil 时降级使用
type noopEventEmitter struct{}

func (e *noopEventEmitter) Emit(name string, data ...any) bool { return false }

// getOrCreateAgent 获取或创建 Agent 实例
func (s *ChatService) getOrCreateAgent() (*agent.Agent, error) {
	// 如果 Agent 已存在且配置未变化，直接返回
	if s.agent != nil {
		return s.agent, nil
	}

	// 确保 LLM 客户端存在
	if s.currentClient == nil {
		client, err := s.createClient()
		if err != nil {
			return nil, err
		}
		s.currentClient = client
	}

	// 获取工具列表
	tools := s.registry.ListTools()
	log.Printf("[ChatService] 注册到 Agent 的工具数量: %d", len(tools))

	// 创建 Agent
	systemPrompt := s.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是一个QA测试助手。你可以使用提供的工具来帮助用户检查Excel配表、查询数据、执行测试等。当用户提出与配表相关的问题时，请主动使用工具获取数据后再回答，不要凭猜测回答。"
	}

	// 调试：输出工具名称
	for _, t := range tools {
		log.Printf("[ChatService] 工具: %s", t.Name())
	}
	// 发射调试事件
	s.emitter.Emit("chatDebug", map[string]any{
		"toolCount":    len(tools),
		"systemPrompt": systemPrompt,
		"model":        s.config.AnthropicConfig.Model,
	})

	s.agent = agent.New(
		agent.WithLLM(s.currentClient),
		agent.WithMemory(s.memory.FileMemory),
		agent.WithTools(tools...),
		agent.WithSystemPrompt(systemPrompt),
		agent.WithHooks(&agent.Hooks{
			OnStart: func(input string) {
				log.Printf("[Agent] 开始处理输入: %s", truncateString(input, 50))
				s.emitter.Emit("chatStreamStart", input)
			},
			OnLLMCall: func(messages []*llm.Message) {
				log.Printf("[Agent] LLM 调用, 消息数: %d, 工具数: %d", len(messages), len(tools))
				s.emitter.Emit("chatDebug", map[string]any{
					"event":     "llmCall",
					"msgCount":  len(messages),
					"toolCount": len(tools),
				})
			},
			OnToolCall: func(name string, args map[string]any) {
				log.Printf("[Agent] 工具调用: %s", name)
				s.emitter.Emit("chatToolCall", map[string]any{
					"name": name,
					"args": args,
				})
			},
			OnToolResult: func(result *tool.ToolResult) {
				if result.Error != nil {
					log.Printf("[Agent] 工具执行错误: %s, 错误: %v", result.Name, result.Error)
				} else {
					log.Printf("[Agent] 工具执行结果: %s", result.Name)
				}
				s.emitter.Emit("chatToolResult", map[string]any{
					"name":   result.Name,
					"result": result.Result,
					"error":  result.Error,
				})
			},
			OnError: func(err error, state *agent.State) {
				log.Printf("[Agent] 错误: %v, 状态: %+v", err, state)
				s.emitter.Emit("chatStreamError", err.Error())
			},
			OnDone: func(result *agent.Result) {
				log.Printf("[Agent] 完成, 内容长度: %d", len(result.Content))
				s.emitter.Emit("chatStreamDone", nil)
			},
		}),
	)

	return s.agent, nil
}

// processAgentEvents 处理 Agent 的流式事件，转发给前端渲染
//
// 为什么需要序列号：Wails Event.Emit 经 WebView2 bridge 传递到 JS 时，
// 短间隔内连续 emit 的事件可能乱序到达前端（观察日志中事件 #1-#7 到达顺序为 1,2,6,5,3,4,7）。
// 中文字符被拆成多个 chunk 时乱序会导致文本错乱（如"才能全量"变成"执行才能Excel量全"）。
// 通过为每个 content chunk 附带递增序列号 seq，前端按 seq 排序后拼接，保证显示正确。
func (s *ChatService) processAgentEvents(ctx context.Context, eventCh <-chan *agent.StreamEvent) {
	eventCount := 0
	seq := 0

	for event := range eventCh {
		eventCount++
		log.Printf("[ChatService] 事件 #%d, type: %s", eventCount, event.Type)

		select {
		case <-ctx.Done():
			s.emitter.Emit("chatStreamError", "请求已取消")
			return
		default:
		}

		switch event.Type {
		case agent.EventTypeContent:
			if event.Content != "" {
				seq++
				s.emitter.Emit("chatStreamChunk", map[string]any{
					"content": event.Content,
					"seq":     seq,
				})
			}

		case agent.EventTypeToolCall:
			log.Printf("[ChatService] 工具调用: %s, 参数: %v", event.ToolCall.Function.Name, event.ToolCall.Function.Arguments)

		case agent.EventTypeToolResult:
			if event.ToolResult.Error != nil {
				log.Printf("[ChatService] 工具执行错误: %v", event.ToolResult.Error)
			} else {
				log.Printf("[ChatService] 工具结果: %s, 结果长度: %d", event.ToolResult.Name, len(fmt.Sprintf("%v", event.ToolResult.Result)))
			}

		case agent.EventTypeDone:
			log.Printf("[ChatService] Agent 完成, 共 %d 个事件, seq: %d", eventCount, seq)
			return

		case agent.EventTypeError:
			log.Printf("[ChatService] Agent 错误: %v", event.Error)
			return

		default:
			log.Printf("[ChatService] 未知事件类型: %s", event.Type)
		}
	}
}

// truncateString 截断字符串用于日志
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
