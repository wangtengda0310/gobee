package agent

import (
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// WithLLM 设置 LLM 客户端
func WithLLM(llm llm.ChatCompleter) Option {
	return func(c *Config) {
		c.LLM = llm
	}
}

// WithSystemPrompt 设置系统提示词
func WithSystemPrompt(prompt string) Option {
	return func(c *Config) {
		c.SystemPrompt = prompt
	}
}

// WithTools 设置工具列表
func WithTools(tools ...tool.Tool) Option {
	return func(c *Config) {
		c.Tools = append(c.Tools, tools...)
	}
}

// WithMaxLoops 设置最大循环次数
func WithMaxLoops(maxLoops int) Option {
	return func(c *Config) {
		if maxLoops > 0 {
			c.MaxLoops = maxLoops
		}
	}
}

// WithTimeout 设置执行超时
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

// WithHooks 设置钩子函数
func WithHooks(hooks *Hooks) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

// WithMaxTokens 设置最大生成 token 数
func WithMaxTokens(maxTokens int) Option {
	return func(c *Config) {
		if maxTokens > 0 {
			c.MaxTokens = maxTokens
		}
	}
}

// WithTemperature 设置采样温度
func WithTemperature(temp float64) Option {
	return func(c *Config) {
		if temp >= 0 && temp <= 2 {
			c.Temperature = temp
		}
	}
}

// WithMemory 设置记忆管理器
// 注意：这个选项会创建一个使用 memory 的 Agent
func WithMemory(m memory.Memory) Option {
	return func(c *Config) {
		// Memory 在 Agent 结构体中单独处理
		// 这里仅用于文档和未来扩展
		_ = m
	}
}
