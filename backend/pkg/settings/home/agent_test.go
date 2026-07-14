package home

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/agent/pkg/agent"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	mockllm "github.com/wangtengda0310/gobee/agent/pkg/llm/mock"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// TestProcessAgentEvents_ContentChunks content 事件按序发送，附带递增 seq
func TestProcessAgentEvents_ContentChunks(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 3)

	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "你好"}
	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "世界"}
	ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	chunks := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks, 2)

	// 验证 seq 递增
	data0 := chunks[0].Data[0].(map[string]any)
	data1 := chunks[1].Data[0].(map[string]any)
	assert.Equal(t, "你好", data0["content"])
	assert.Equal(t, 1, data0["seq"])
	assert.Equal(t, "世界", data1["content"])
	assert.Equal(t, 2, data1["seq"])
}

// TestProcessAgentEvents_DoneEventStops done 事件后停止处理
func TestProcessAgentEvents_DoneEventStops(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 3)

	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "hello"}
	ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
	// done 之后再发事件不会被处理
	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "after-done"}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	chunks := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks, 1)
}

// TestProcessAgentEvents_ErrorEventStops error 事件后停止处理
func TestProcessAgentEvents_ErrorEventStops(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 3)

	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "before"}
	ch <- &agent.StreamEvent{Type: agent.EventTypeError, Error: assert.AnError}
	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "after"}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	chunks := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks, 1)
}

// TestProcessAgentEvents_ContextCancelled context 取消时发送 chatStreamError
func TestProcessAgentEvents_ContextCancelled(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "test"}
	close(ch)

	svc.processAgentEvents(ctx, ch)

	errors := emitter.EventsByName("chatStreamError")
	assert.Len(t, errors, 1)
	assert.Equal(t, "请求已取消", errors[0].Data[0])
}

// TestProcessAgentEvents_EmptyContentSkipped 空 Content 不触发 emit
func TestProcessAgentEvents_EmptyContentSkipped(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 3)

	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: ""}
	ch <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "visible"}
	ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	chunks := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks, 1)
	data := chunks[0].Data[0].(map[string]any)
	assert.Equal(t, "visible", data["content"])
	assert.Equal(t, 1, data["seq"]) // seq 从 1 开始，空 content 不消耗 seq
}

// TestProcessAgentEvents_ToolCallOnlyLogs tool_call 事件只记录日志不 emit
func TestProcessAgentEvents_ToolCallOnlyLogs(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 2)

	ch <- &agent.StreamEvent{
		Type: agent.EventTypeToolCall,
		ToolCall: &llm.ToolCall{
			ID:   "call_1",
			Type: "function",
			Function: &llm.FunctionCall{
				Name:      "get_all_excels",
				Arguments: `{"dirPath":"/test"}`,
			},
		},
	}
	ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	// 不应产生任何 emit 事件（tool_call 只记日志）
	chunks := emitter.EventsByName("chatToolCall")
	assert.Len(t, chunks, 0)
}

// TestProcessAgentEvents_ToolResultOnlyLogs tool_result 事件只记录日志不 emit
func TestProcessAgentEvents_ToolResultOnlyLogs(t *testing.T) {
	svc, emitter := newTestChatService(t)
	ch := make(chan *agent.StreamEvent, 2)

	ch <- &agent.StreamEvent{
		Type: agent.EventTypeToolResult,
		ToolResult: &tool.ToolResult{
			ToolCallID: "call_1",
			Name:       "get_all_excels",
			Result:     "test-result",
		},
	}
	ch <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch)

	svc.processAgentEvents(context.Background(), ch)

	results := emitter.EventsByName("chatToolResult")
	assert.Len(t, results, 0)
}

// TestProcessAgentEvents_SeqResetsPerCall 每次 processAgentEvents 调用 seq 从 1 开始
func TestProcessAgentEvents_SeqResetsPerCall(t *testing.T) {
	svc, emitter := newTestChatService(t)

	// 第一次调用
	ch1 := make(chan *agent.StreamEvent, 2)
	ch1 <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "first"}
	ch1 <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch1)
	svc.processAgentEvents(context.Background(), ch1)

	chunks1 := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks1, 1)
	data1 := chunks1[0].Data[0].(map[string]any)
	assert.Equal(t, 1, data1["seq"])

	emitter.Reset()

	// 第二次调用 seq 重新从 1 开始
	ch2 := make(chan *agent.StreamEvent, 2)
	ch2 <- &agent.StreamEvent{Type: agent.EventTypeContent, Content: "second"}
	ch2 <- &agent.StreamEvent{Type: agent.EventTypeDone}
	close(ch2)
	svc.processAgentEvents(context.Background(), ch2)

	chunks2 := emitter.EventsByName("chatStreamChunk")
	assert.Len(t, chunks2, 1)
	data2 := chunks2[0].Data[0].(map[string]any)
	assert.Equal(t, 1, data2["seq"])
}

// TestToolCallEndToEnd 端到端工具调用测试
//
// 验证完整链路：mock LLM 返回 tool_use → Agent 执行注册的工具 → bool 变量被改变
//
// 流程：
//  1. 注册临时工具 set_flag，执行时将 *bool 设为 true
//  2. mock LLM 第一轮返回 tool_use（调用 set_flag）
//  3. mock LLM 第二轮返回纯文本（工具执行后 LLM 生成最终回复）
//  4. 断言 flag 被改为 true，emitter 收到 chatToolCall/chatToolResult 事件
func TestToolCallEndToEnd(t *testing.T) {
	svc, emitter := newTestChatService(t)

	// ---- 1. 注册临时工具 ----
	var flag bool
	svc.registry.MustRegister(
		tool.NewFunction("set_flag", "将 flag 设置为 true",
			func(ctx context.Context, args map[string]any) (any, error) {
				flag = true
				return "flag 已设置为 true", nil
			},
		),
	)

	// ---- 2. 配置 mock LLM ----
	// 第一轮：LLM 返回 tool_use（调用 set_flag）
	// 第二轮：LLM 返回纯文本（处理完工具结果后生成最终回复）
	mockClient := mockllm.NewClient()
	mockClient.StreamChunks = [][]*llm.StreamChunk{
		// 第一轮：LLM 决定调用工具
		{
			llm.NewToolUseChunk([]*llm.ToolCall{
				llm.NewToolCall("call_1", "set_flag", "{}"),
			}),
			llm.NewDoneChunk(&llm.ChatResponse{
				StopReason: llm.StopReasonToolUse,
				ToolCalls: []*llm.ToolCall{
					llm.NewToolCall("call_1", "set_flag", "{}"),
				},
			}),
		},
		// 第二轮：LLM 根据工具结果生成文本回复
		{
			llm.NewContentChunk("已将 flag 设置为 true"),
			llm.NewDoneChunk(&llm.ChatResponse{
				Content:    "已将 flag 设置为 true",
				StopReason: llm.StopReasonEndTurn,
			}),
		},
	}
	svc.currentClient = mockClient

	// 配置 ChatService
	svc.config = &ChatConfig{
		Provider:     "anthropic",
		SystemPrompt: "你是测试助手",
	}

	// ---- 3. 执行 ----
	assert.False(t, flag, "调用前 flag 应为 false")

	ctx := context.Background()
	ag, err := svc.getOrCreateAgent()
	assert.NoError(t, err)

	eventCh, err := ag.RunStream(ctx, "请设置 flag")
	assert.NoError(t, err)

	// 消费所有事件（阻塞直到完成）
	svc.processAgentEvents(ctx, eventCh)

	// ---- 4. 断言 ----
	assert.True(t, flag, "工具执行后 flag 应为 true")

	// 验证 emitter 收到工具调用事件
	toolCallEvents := emitter.EventsByName("chatToolCall")
	assert.Len(t, toolCallEvents, 1, "应收到一次 chatToolCall 事件")
	toolCallData := toolCallEvents[0].Data[0].(map[string]any)
	assert.Equal(t, "set_flag", toolCallData["name"])

	// 验证 emitter 收到工具结果事件
	toolResultEvents := emitter.EventsByName("chatToolResult")
	assert.Len(t, toolResultEvents, 1, "应收到一次 chatToolResult 事件")
	toolResultData := toolResultEvents[0].Data[0].(map[string]any)
	assert.Equal(t, "set_flag", toolResultData["name"])
	assert.Nil(t, toolResultData["error"])

	// 验证最终内容通过 chatStreamChunk 发送
	chunks := emitter.EventsByName("chatStreamChunk")
	assert.NotEmpty(t, chunks, "应收到内容 chunk")

	// 验证完成事件
	doneEvents := emitter.EventsByName("chatStreamDone")
	assert.Len(t, doneEvents, 1, "应收到一次 chatStreamDone 事件")
}
