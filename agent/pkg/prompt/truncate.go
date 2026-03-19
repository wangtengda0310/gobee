package prompt

import (
	"context"
	"fmt"
	"strings"
)

// TruncateStrategy 截断策略接口
// 不同实现提供不同的上下文窗口控制方式
type TruncateStrategy interface {
	// Truncate 执行截断操作
	// messages: 待截断的消息列表
	// 返回: 截断后的消息和可能的摘要
	Truncate(ctx context.Context, messages []HistoryItem) ([]HistoryItem, string, error)
}

// SummaryTruncateStrategy 使用 LLM 生成摘要的截断策略
// 这是最常用的策略，通过 LLM 将早期对话压缩为摘要
type SummaryTruncateStrategy struct {
	// MaxMessages 保留的最大消息数量（不包括摘要）
	MaxMessages int

	// SummaryPrompt 生成摘要时使用的提示词模板
	// 可使用 {{.Content}} 占位符
	SummaryPrompt string

	// generateSummary 生成摘要的函数
	// 由 History.Truncate 方法注入
	generateSummary func(ctx context.Context, content string) (string, error)
}

// NewSummaryTruncateStrategy 创建摘要截断策略
// maxMessages: 保留的最大消息数量，默认为 10
func NewSummaryTruncateStrategy(maxMessages int) *SummaryTruncateStrategy {
	if maxMessages <= 0 {
		maxMessages = 10
	}
	return &SummaryTruncateStrategy{
		MaxMessages: maxMessages,
		SummaryPrompt: `请将以下对话历史压缩为简洁的摘要，保留关键信息：

{{.Content}}

摘要：`,
	}
}

// Truncate 执行摘要截断
// 实现逻辑：
// 1. 消息数量 <= MaxMessages 时直接返回
// 2. 无摘要生成器时退化为滑动窗口
// 3. 有生成器时：前 N-MaxMessages 条消息生成摘要，保留最近 MaxMessages 条
func (s *SummaryTruncateStrategy) Truncate(ctx context.Context, messages []HistoryItem) ([]HistoryItem, string, error) {
	if len(messages) <= s.MaxMessages {
		return messages, "", nil
	}

	if s.generateSummary == nil {
		// 如果没有注入摘要生成函数，使用简单的滑动窗口
		// 这种退化为滑动窗口的行为确保即使没有配置 LLM 也能正常工作
		return messages[len(messages)-s.MaxMessages:], "", nil
	}

	// 提取需要摘要的内容
	// 格式: "role: content\n" 便于 LLM 理解对话结构
	var contentBuilder strings.Builder
	toSummarize := messages[:len(messages)-s.MaxMessages]
	for _, m := range toSummarize {
		contentBuilder.WriteString(m.Role)
		contentBuilder.WriteString(": ")
		contentBuilder.WriteString(m.Content)
		contentBuilder.WriteString("\n")
	}

	// 调用注入的摘要生成函数（通常由 History.SetSummaryGenerator 设置）
	summary, err := s.generateSummary(ctx, contentBuilder.String())
	if err != nil {
		return nil, "", fmt.Errorf("生成摘要失败: %w", err)
	}

	// 返回摘要 + 保留的消息
	result := []HistoryItem{
		{
			Role:    "system",
			Content: fmt.Sprintf("对话历史摘要:\n%s", summary),
		},
	}
	result = append(result, messages[len(messages)-s.MaxMessages:]...)

	return result, summary, nil
}

// SlidingWindowStrategy 滑动窗口截断策略
// 简单地保留最近的 N 条消息
type SlidingWindowStrategy struct {
	// MaxMessages 保留的最大消息数量
	MaxMessages int

	// PreserveSystem 是否保留系统消息
	PreserveSystem bool
}

// NewSlidingWindowStrategy 创建滑动窗口策略
func NewSlidingWindowStrategy(maxMessages int) *SlidingWindowStrategy {
	if maxMessages <= 0 {
		maxMessages = 10
	}
	return &SlidingWindowStrategy{
		MaxMessages:     maxMessages,
		PreserveSystem:  true,
	}
}

// Truncate 执行滑动窗口截断
func (s *SlidingWindowStrategy) Truncate(_ context.Context, messages []HistoryItem) ([]HistoryItem, string, error) {
	if len(messages) <= s.MaxMessages {
		return messages, "", nil
	}

	var systemMessages []HistoryItem
	var otherMessages []HistoryItem

	// 分离系统消息和其他消息
	for _, m := range messages {
		if m.Role == "system" && s.PreserveSystem {
			systemMessages = append(systemMessages, m)
		} else {
			otherMessages = append(otherMessages, m)
		}
	}

	// 截断非系统消息
	if len(otherMessages) > s.MaxMessages {
		otherMessages = otherMessages[len(otherMessages)-s.MaxMessages:]
	}

	// 合并结果
	result := append(systemMessages, otherMessages...)
	return result, "", nil
}

// FixedSystemStrategy 固定系统 + 滑动窗口策略
// 始终保留第一条系统消息，其余使用滑动窗口
type FixedSystemStrategy struct {
	// WindowSize 滑动窗口大小
	WindowSize int

	// delegate 内部使用的滑动窗口策略
	delegate *SlidingWindowStrategy
}

// NewFixedSystemStrategy 创建固定系统策略
func NewFixedSystemStrategy(windowSize int) *FixedSystemStrategy {
	if windowSize <= 0 {
		windowSize = 10
	}
	return &FixedSystemStrategy{
		WindowSize: windowSize,
		delegate:   NewSlidingWindowStrategy(windowSize),
	}
}

// Truncate 执行固定系统 + 滑动窗口截断
func (s *FixedSystemStrategy) Truncate(ctx context.Context, messages []HistoryItem) ([]HistoryItem, string, error) {
	if len(messages) <= s.WindowSize+1 {
		return messages, "", nil
	}

	// 确保第一条系统消息被保留
	s.delegate.PreserveSystem = true
	return s.delegate.Truncate(ctx, messages)
}
