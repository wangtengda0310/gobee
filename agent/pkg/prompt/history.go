package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// HistoryItem 表示历史消息项
type HistoryItem struct {
	// Role 消息角色: "user", "assistant", "system", "tool"
	Role string

	// Content 消息内容
	Content string

	// ToolCallID 工具调用 ID（用于工具响应消息）
	ToolCallID string

	// ToolName 工具名称（用于工具响应消息）
	ToolName string

	// ToolCalls 助手发起的工具调用
	ToolCalls []*llm.ToolCall
}

// History 管理多轮对话的消息历史
type History struct {
	items []HistoryItem

	// summaryGenerator 用于生成摘要的函数
	// 由 Truncate 方法在需要时注入到策略中
	summaryGenerator func(ctx context.Context, content string) (string, error)
}

// NewHistory 创建新的对话历史管理器
func NewHistory() *History {
	return &History{
		items: make([]HistoryItem, 0),
	}
}

// AddUser 添加用户消息
func (h *History) AddUser(content string) *History {
	h.items = append(h.items, HistoryItem{
		Role:    "user",
		Content: content,
	})
	return h
}

// AddAssistant 添加助手消息
func (h *History) AddAssistant(content string) *History {
	h.items = append(h.items, HistoryItem{
		Role:    "assistant",
		Content: content,
	})
	return h
}

// AddAssistantWithTools 添加包含工具调用的助手消息
func (h *History) AddAssistantWithTools(content string, toolCalls []*llm.ToolCall) *History {
	h.items = append(h.items, HistoryItem{
		Role:      "assistant",
		Content:   content,
		ToolCalls: toolCalls,
	})
	return h
}

// AddSystem 添加系统消息
func (h *History) AddSystem(content string) *History {
	h.items = append(h.items, HistoryItem{
		Role:    "system",
		Content: content,
	})
	return h
}

// AddToolResult 添加工具响应消息
func (h *History) AddToolResult(toolCallID, toolName, result string) *History {
	h.items = append(h.items, HistoryItem{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
		ToolName:   toolName,
	})
	return h
}

// Add 添加自定义历史项
func (h *History) Add(item HistoryItem) *History {
	h.items = append(h.items, item)
	return h
}

// Len 返回消息数量
func (h *History) Len() int {
	return len(h.items)
}

// Clear 清空历史
func (h *History) Clear() *History {
	h.items = make([]HistoryItem, 0)
	return h
}

// Items 返回所有历史项
func (h *History) Items() []HistoryItem {
	return h.items
}

// Last 返回最后一条消息
func (h *History) Last() *HistoryItem {
	if len(h.items) == 0 {
		return nil
	}
	return &h.items[len(h.items)-1]
}

// SetSummaryGenerator 设置摘要生成函数
// 该函数用于在 Truncate 时生成对话摘要
func (h *History) SetSummaryGenerator(fn func(ctx context.Context, content string) (string, error)) {
	h.summaryGenerator = fn
}

// Truncate 使用指定策略截断历史
// 需要注入 llm.ChatCompleter 用于生成摘要（当策略需要时）
//
// 设计说明：
// - 摘要生成器通过依赖注入方式设置，避免 History 直接依赖 llm.ChatCompleter
// - 这样可以在测试中使用 mock 函数，也支持不使用 LLM 的简单截断策略
//
// 使用示例:
//
//	history.SetSummaryGenerator(func(ctx context.Context, content string) (string, error) {
//	    resp, err := client.Complete(ctx, &llm.ChatRequest{
//	        Messages: []*llm.Message{
//	            {Role: llm.RoleUser, Content: llm.Text("请总结以下对话:\n" + content)},
//	        },
//	    })
//	    if err != nil {
//	        return "", err
//	    }
//	    return resp.Content, nil
//	})
//	truncated, summary, err := history.Truncate(ctx, prompt.NewSummaryTruncateStrategy(10))
func (h *History) Truncate(ctx context.Context, strategy TruncateStrategy) (*History, string, error) {
	// 如果是摘要策略，注入摘要生成函数
	// 这种类型断言+注入的方式允许策略按需使用 LLM 能力
	if s, ok := strategy.(*SummaryTruncateStrategy); ok && h.summaryGenerator != nil {
		s.generateSummary = h.summaryGenerator
	}

	items, summary, err := strategy.Truncate(ctx, h.items)
	if err != nil {
		return nil, "", err
	}

	return &History{
		items:            items,
		summaryGenerator: h.summaryGenerator,
	}, summary, nil
}

// ToMessages 将历史转换为 llm.Message 列表
func (h *History) ToMessages() []*llm.Message {
	messages := make([]*llm.Message, 0, len(h.items))

	for _, item := range h.items {
		msg := &llm.Message{}

		switch item.Role {
		case "user":
			msg.Role = llm.RoleUser
		case "assistant":
			msg.Role = llm.RoleAssistant
		case "system":
			msg.Role = llm.RoleSystem
		case "tool":
			msg.Role = llm.RoleTool
		default:
			msg.Role = llm.Role(item.Role)
		}

		msg.Content = llm.Text(item.Content)
		msg.ToolCallID = item.ToolCallID
		msg.Name = item.ToolName
		msg.ToolCalls = item.ToolCalls

		messages = append(messages, msg)
	}

	return messages
}

// Clone 克隆历史
func (h *History) Clone() *History {
	newHistory := NewHistory()
	newHistory.items = append([]HistoryItem{}, h.items...)
	newHistory.summaryGenerator = h.summaryGenerator
	return newHistory
}

// String 返回历史的字符串表示（用于调试）
func (h *History) String() string {
	var sb strings.Builder
	for i, item := range h.items {
		fmt.Fprintf(&sb, "[%d] %s: %s\n", i, item.Role, item.Content)
		if len(item.ToolCalls) > 0 {
			fmt.Fprintf(&sb, "  ToolCalls: %d\n", len(item.ToolCalls))
		}
	}
	return sb.String()
}

// GetLastUserMessage 获取最后一条用户消息
func (h *History) GetLastUserMessage() *HistoryItem {
	for i := len(h.items) - 1; i >= 0; i-- {
		if h.items[i].Role == "user" {
			return &h.items[i]
		}
	}
	return nil
}

// GetLastAssistantMessage 获取最后一条助手消息
func (h *History) GetLastAssistantMessage() *HistoryItem {
	for i := len(h.items) - 1; i >= 0; i-- {
		if h.items[i].Role == "assistant" {
			return &h.items[i]
		}
	}
	return nil
}
