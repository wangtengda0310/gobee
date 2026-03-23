package agent

import (
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/prompt"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// State 执行状态
// 记录 Agent 执行过程中的状态信息
type State struct {
	// Input 用户输入
	Input string

	// Messages 当前对话消息列表
	Messages []*llm.Message

	// ToolCalls 当前轮次的工具调用
	ToolCalls []*llm.ToolCall

	// ToolResults 工具执行结果
	ToolResults []*tool.ToolResult

	// Response LLM 响应
	Response *llm.ChatResponse

	// LoopCount 循环计数
	LoopCount int

	// Done 是否完成
	Done bool

	// Error 执行错误
	Error error
}

// Result 执行结果
// Agent 执行完成后的最终结果
type Result struct {
	// Content 最终回复内容
	Content string

	// ToolCalls 所有工具调用记录
	ToolCalls []*llm.ToolCall

	// Usage token 使用统计
	Usage *llm.Usage

	// StopReason 停止原因
	StopReason llm.StopReason

	// History 对话历史
	History *prompt.History

	// Duration 执行耗时
	Duration time.Duration

	// LoopCount 总循环次数
	LoopCount int
}

// Hooks 钩子函数
// 用于在执行过程中注入自定义逻辑
type Hooks struct {
	// OnStart 执行开始时调用
	OnStart func(input string)

	// OnLoop 每次循环开始时调用
	OnLoop func(loopCount int, state *State)

	// OnLLMCall LLM 调用前调用
	OnLLMCall func(messages []*llm.Message)

	// OnLLMResponse LLM 响应后调用
	OnLLMResponse func(response *llm.ChatResponse)

	// OnToolCall 工具调用时调用
	OnToolCall func(name string, args map[string]any)

	// OnToolResult 工具执行结果
	OnToolResult func(result *tool.ToolResult)

	// OnError 发生错误时调用
	OnError func(err error, state *State)

	// OnDone 执行完成时调用
	OnDone func(result *Result)
}

// Config Agent 配置
type Config struct {
	// LLM LLM 客户端
	LLM llm.ChatCompleter

	// SystemPrompt 系统提示词
	SystemPrompt string

	// Tools 工具列表
	Tools []tool.Tool

	// MaxLoops 最大循环次数
	MaxLoops int

	// Timeout 执行超时
	Timeout time.Duration

	// Hooks 钩子函数
	Hooks *Hooks

	// MaxTokens 最大生成 token 数
	MaxTokens int

	// Temperature 采样温度
	Temperature float64
}

// Option Agent 配置选项
type Option func(*Config)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxLoops:    10,
		Timeout:     5 * time.Minute,
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}
