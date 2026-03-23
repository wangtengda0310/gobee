package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
	"github.com/wangtengda0310/gobee/agent/pkg/memory"
	"github.com/wangtengda0310/gobee/agent/pkg/prompt"
	"github.com/wangtengda0310/gobee/agent/pkg/tool"
)

// Agent 智能体
// 核心执行单元，实现感知-思考-行动循环
type Agent struct {
	config   *Config
	registry *tool.Registry
	executor *tool.BatchExecutor
	history  *prompt.History
	memory   memory.Memory
	mu       sync.RWMutex
}

// New 创建新的 Agent
// opts: 配置选项
func New(opts ...Option) *Agent {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// 创建工具注册表
	registry := tool.NewRegistry()
	if len(config.Tools) > 0 {
		registry.MustRegister(config.Tools...)
	}

	// 创建批量执行器
	executor := tool.NewBatchExecutor(registry, 4)

	return &Agent{
		config:   config,
		registry: registry,
		executor: executor,
		history:  prompt.NewHistory(),
	}
}

// Run 执行任务
// ctx: 上下文
// input: 用户输入
// 返回: 执行结果
func (a *Agent) Run(ctx context.Context, input string) (*Result, error) {
	startTime := time.Now()

	// 设置超时
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	// 触发开始钩子
	if a.config.Hooks != nil && a.config.Hooks.OnStart != nil {
		a.config.Hooks.OnStart(input)
	}

	// 初始化状态
	state := &State{
		Input:     input,
		Messages:  make([]*llm.Message, 0),
		Done:      false,
		LoopCount: 0,
	}

	// 添加系统提示词
	if a.config.SystemPrompt != "" {
		state.Messages = append(state.Messages, &llm.Message{
			Role:    llm.RoleSystem,
			Content: llm.Text(a.config.SystemPrompt),
		})
	}

	// 添加用户输入
	state.Messages = append(state.Messages, &llm.Message{
		Role:    llm.RoleUser,
		Content: llm.Text(input),
	})

	// 执行循环
	result := &Result{
		History:    prompt.NewHistory(),
		ToolCalls:  make([]*llm.ToolCall, 0),
		LoopCount:  0,
		StopReason: llm.StopReasonEndTurn,
	}

	for !state.Done && state.LoopCount < a.config.MaxLoops {
		state.LoopCount++
		result.LoopCount = state.LoopCount

		// 触发循环钩子
		if a.config.Hooks != nil && a.config.Hooks.OnLoop != nil {
			a.config.Hooks.OnLoop(state.LoopCount, state)
		}

		// 调用 LLM
		resp, err := a.callLLM(ctx, state.Messages)
		if err != nil {
			state.Error = err
			// 触发错误钩子
			if a.config.Hooks != nil && a.config.Hooks.OnError != nil {
				a.config.Hooks.OnError(err, state)
			}
			return nil, fmt.Errorf("LLM 调用失败: %w", err)
		}

		state.Response = resp

		// 触发 LLM 响应钩子
		if a.config.Hooks != nil && a.config.Hooks.OnLLMResponse != nil {
			a.config.Hooks.OnLLMResponse(resp)
		}

		// 添加助手消息到历史
		assistantMsg := &llm.Message{
			Role:      llm.RoleAssistant,
			Content:   llm.Text(resp.Content),
			ToolCalls: resp.ToolCalls,
		}
		state.Messages = append(state.Messages, assistantMsg)

		// 检查是否有工具调用
		if len(resp.ToolCalls) > 0 {
			result.ToolCalls = append(result.ToolCalls, resp.ToolCalls...)

			// 执行工具
			toolResults := a.executeTools(ctx, resp.ToolCalls)
			state.ToolResults = toolResults

			// 添加工具结果消息
			for _, tr := range toolResults {
				state.Messages = append(state.Messages, &llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tr.ToolCallID,
					Name:       tr.Name,
					Content:    llm.Text(a.resultToString(tr)),
				})
			}

			// 继续循环，让 LLM 处理工具结果
			continue
		}

		// 没有工具调用，完成执行
		state.Done = true
		result.Content = resp.Content
		result.Usage = resp.Usage
		result.StopReason = resp.StopReason
	}

	// 构建历史
	for _, msg := range state.Messages {
		switch msg.Role {
		case llm.RoleUser:
			result.History.AddUser(llm.TextString(msg.Content))
		case llm.RoleAssistant:
			result.History.AddAssistantWithTools(llm.TextString(msg.Content), msg.ToolCalls)
		case llm.RoleTool:
			result.History.AddToolResult(msg.ToolCallID, msg.Name, llm.TextString(msg.Content))
		case llm.RoleSystem:
			result.History.AddSystem(llm.TextString(msg.Content))
		}
	}

	result.Duration = time.Since(startTime)

	// 触发完成钩子
	if a.config.Hooks != nil && a.config.Hooks.OnDone != nil {
		a.config.Hooks.OnDone(result)
	}

	return result, nil
}

// callLLM 调用 LLM
func (a *Agent) callLLM(ctx context.Context, messages []*llm.Message) (*llm.ChatResponse, error) {
	// 触发 LLM 调用钩子
	if a.config.Hooks != nil && a.config.Hooks.OnLLMCall != nil {
		a.config.Hooks.OnLLMCall(messages)
	}

	req := &llm.ChatRequest{
		Messages:    messages,
		MaxTokens:   a.config.MaxTokens,
		Temperature: a.config.Temperature,
	}

	// 添加工具定义
	if tools := a.registry.GetDefinitions(); len(tools) > 0 {
		req.Tools = tools
	}

	return a.config.LLM.Complete(ctx, req)
}

// executeTools 执行工具调用
func (a *Agent) executeTools(ctx context.Context, calls []*llm.ToolCall) []*tool.ToolResult {
	results := make([]*tool.ToolResult, len(calls))

	for i, call := range calls {
		// 触发工具调用钩子
		if a.config.Hooks != nil && a.config.Hooks.OnToolCall != nil {
			// 解析参数
			args := a.parseArguments(call.Function.Arguments)
			a.config.Hooks.OnToolCall(call.Function.Name, args)
		}

		// 执行工具
		args := a.parseArguments(call.Function.Arguments)
		result, err := a.registry.Execute(ctx, call.Function.Name, args)

		tr := &tool.ToolResult{
			ToolCallID: call.ID,
			Name:       call.Function.Name,
			Result:     result,
			Error:      err,
		}
		results[i] = tr

		// 触发工具结果钩子
		if a.config.Hooks != nil && a.config.Hooks.OnToolResult != nil {
			a.config.Hooks.OnToolResult(tr)
		}
	}

	return results
}

// parseArguments 解析参数
func (a *Agent) parseArguments(argsJSON string) map[string]any {
	args := make(map[string]any)
	if argsJSON == "" {
		return args
	}
	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		// 解析失败返回空 map
		return args
	}
	return args
}

// resultToString 将结果转换为字符串
func (a *Agent) resultToString(tr *tool.ToolResult) string {
	if tr.Error != nil {
		return fmt.Sprintf("错误: %s", tr.Error.Error())
	}
	return fmt.Sprintf("%v", tr.Result)
}

// AddTool 添加工具
func (a *Agent) AddTool(t tool.Tool) error {
	return a.registry.Register(t)
}

// SetMemory 设置记忆管理器
func (a *Agent) SetMemory(m memory.Memory) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.memory = m
}

// ClearHistory 清空历史
func (a *Agent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = prompt.NewHistory()
}
