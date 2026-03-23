// Package mock 提供 LLM 客户端的模拟实现
// 用于测试 Agent 框架而不依赖外部 API
package mock

import (
	"context"
	"sync"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// Client 模拟 LLM 客户端
// 支持预设响应队列、错误模拟和调用记录
type Client struct {
	mu sync.RWMutex

	// Responses 预设响应队列
	Responses []*llm.ChatResponse

	// StreamChunks 流式响应队列
	StreamChunks [][]*llm.StreamChunk

	// CallCount 调用计数
	CallCount int

	// LastRequest 最后一次请求
	LastRequest *llm.ChatRequest

	// Error 预设错误
	Error error

	// streamIndex 流式响应索引
	streamIndex int
}

// NewClient 创建新的 Mock 客户端
func NewClient() *Client {
	return &Client{
		Responses: make([]*llm.ChatResponse, 0),
	}
}

// Complete 实现 ChatCompleter 接口
func (m *Client) Complete(_ context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallCount++
	m.LastRequest = req

	// 如果预设了错误，返回错误
	if m.Error != nil {
		return nil, m.Error
	}

	// 如果有预设响应，返回队列中的响应
	if m.CallCount <= len(m.Responses) {
		return m.Responses[m.CallCount-1], nil
	}

	// 默认响应
	return &llm.ChatResponse{
		Content:    "default mock response",
		StopReason: llm.StopReasonEndTurn,
	}, nil
}

// Stream 实现 StreamCompleter 接口
func (m *Client) Stream(_ context.Context, req *llm.ChatRequest) (<-chan *llm.StreamChunk, error) {
	m.mu.Lock()
	m.CallCount++
	m.LastRequest = req
	m.mu.Unlock()

	// 如果预设了错误，返回错误
	if m.Error != nil {
		return nil, m.Error
	}

	ch := make(chan *llm.StreamChunk, 10)

	go func() {
		defer close(ch)

		m.mu.RLock()
		chunks := m.StreamChunks
		idx := m.streamIndex
		m.mu.RUnlock()

		if idx < len(chunks) {
			for _, chunk := range chunks[idx] {
				ch <- chunk
			}
			m.mu.Lock()
			m.streamIndex++
			m.mu.Unlock()
		} else {
			// 默认流式响应
			ch <- llm.NewContentChunk("default stream response")
			ch <- llm.NewDoneChunk(&llm.ChatResponse{
				Content:    "default stream response",
				StopReason: llm.StopReasonEndTurn,
			})
		}
	}()

	return ch, nil
}

// AddResponse 添加单个响应到队列
func (m *Client) AddResponse(resp *llm.ChatResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses = append(m.Responses, resp)
}

// AddToolCallResponse 添加工具调用响应
func (m *Client) AddToolCallResponse(name, argsJSON string) {
	m.AddResponse(&llm.ChatResponse{
		StopReason: llm.StopReasonToolUse,
		ToolCalls: []*llm.ToolCall{
			llm.NewToolCall("call_"+name, name, argsJSON),
		},
	})
}

// AddTextResponse 添加文本响应
func (m *Client) AddTextResponse(text string) {
	m.AddResponse(&llm.ChatResponse{
		Content:    text,
		StopReason: llm.StopReasonEndTurn,
	})
}

// SetError 设置错误响应
func (m *Client) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = err
}

// Reset 重置客户端状态
func (m *Client) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Responses = nil
	m.StreamChunks = nil
	m.CallCount = 0
	m.LastRequest = nil
	m.Error = nil
	m.streamIndex = 0
}

// GetCallCount 获取调用次数
func (m *Client) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.CallCount
}

// GetLastRequest 获取最后一次请求
func (m *Client) GetLastRequest() *llm.ChatRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.LastRequest
}
