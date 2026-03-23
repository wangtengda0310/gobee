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

	// 从 memory 获取历史消息
	if a.memory != nil {
		historyMsgs, err := a.memory.GetContext(ctx)
		if err == nil && len(historyMsgs) > 0 {
			state.Messages = append(state.Messages, historyMsgs...)
		}
	}

	// 添加系统提示词（如果没有历史或历史中没有系统消息）
	if a.config.SystemPrompt != "" {
		hasSystemMsg := false
		for _, msg := range state.Messages {
			if msg.Role == llm.RoleSystem {
				hasSystemMsg = true
				break
			}
		}
		if !hasSystemMsg {
			state.Messages = append([]*llm.Message{{
				Role:    llm.RoleSystem,
				Content: llm.Text(a.config.SystemPrompt),
			}}, state.Messages...)
		}
	}

	// 添加用户输入
	userMsg := &llm.Message{
		Role:    llm.RoleUser,
		Content: llm.Text(input),
	}
	state.Messages = append(state.Messages, userMsg)

	// 将用户消息添加到 memory
	if a.memory != nil {
		a.memory.Add(ctx, userMsg)
	}

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

		// 将助手消息添加到 memory
		if a.memory != nil {
			a.memory.Add(ctx, assistantMsg)
		}

		// 检查是否有工具调用
		if len(resp.ToolCalls) > 0 {
			result.ToolCalls = append(result.ToolCalls, resp.ToolCalls...)

			// 执行工具
			toolResults := a.executeTools(ctx, resp.ToolCalls)
			state.ToolResults = toolResults

			// 添加工具结果消息
			for _, tr := range toolResults {
				toolMsg := &llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tr.ToolCallID,
					Name:       tr.Name,
					Content:    llm.Text(a.resultToString(tr)),
				}
				state.Messages = append(state.Messages, toolMsg)

				// 将工具结果消息添加到 memory
				if a.memory != nil {
					a.memory.Add(ctx, toolMsg)
				}
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

// RunStream 流式执行任务
// 返回事件通道，调用者可以从通道中读取流式事件
func (a *Agent) RunStream(ctx context.Context, input string) (<-chan *StreamEvent, error) {
	// 创建事件通道
	eventCh := make(chan *StreamEvent, 100)

	go func() {
		defer close(eventCh)

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

		// 从 memory 获取历史消息
		if a.memory != nil {
			historyMsgs, err := a.memory.GetContext(ctx)
			if err == nil && len(historyMsgs) > 0 {
				state.Messages = append(state.Messages, historyMsgs...)
			}
		}

		// 添加系统提示词
		if a.config.SystemPrompt != "" {
			hasSystemMsg := false
			for _, msg := range state.Messages {
				if msg.Role == llm.RoleSystem {
					hasSystemMsg = true
					break
				}
			}
			if !hasSystemMsg {
				state.Messages = append([]*llm.Message{{
					Role:    llm.RoleSystem,
					Content: llm.Text(a.config.SystemPrompt),
				}}, state.Messages...)
			}
		}

		// 添加用户输入
		userMsg := &llm.Message{
			Role:    llm.RoleUser,
			Content: llm.Text(input),
		}
		state.Messages = append(state.Messages, userMsg)

		// 将用户消息添加到 memory
		if a.memory != nil {
			a.memory.Add(ctx, userMsg)
		}

		// 初始化结果
		result := &Result{
			History:    prompt.NewHistory(),
			ToolCalls:  make([]*llm.ToolCall, 0),
			LoopCount:  0,
			StopReason: llm.StopReasonEndTurn,
		}

		// 执行循环
		for !state.Done && state.LoopCount < a.config.MaxLoops {
			state.LoopCount++
			result.LoopCount = state.LoopCount

			// 触发循环钩子
			if a.config.Hooks != nil && a.config.Hooks.OnLoop != nil {
				a.config.Hooks.OnLoop(state.LoopCount, state)
			}

			// 调用 LLM 流式 API
			streamCh, err := a.callLLMStream(ctx, state.Messages)
			if err != nil {
				state.Error = err
				eventCh <- &StreamEvent{Type: EventTypeError, Error: err, LoopCount: state.LoopCount}
				if a.config.Hooks != nil && a.config.Hooks.OnError != nil {
					a.config.Hooks.OnError(err, state)
				}
				return
			}

			// 处理流式响应
			var contentBuffer string
			var toolCalls []*llm.ToolCall
			var usage *llm.Usage
			var stopReason llm.StopReason

			for chunk := range streamCh {
				if chunk.IsError() {
					eventCh <- &StreamEvent{Type: EventTypeError, Error: chunk.Error}
					return
				}

				// 发送内容增量
				if chunk.Content != "" {
					contentBuffer += chunk.Content
					eventCh <- &StreamEvent{
						Type:      EventTypeContent,
						Content:   chunk.Content,
						LoopCount: state.LoopCount,
					}
				}

				// 收集工具调用
				if len(chunk.ToolCalls) > 0 {
					toolCalls = append(toolCalls, chunk.ToolCalls...)
					for _, tc := range chunk.ToolCalls {
						eventCh <- &StreamEvent{
							Type:      EventTypeToolCall,
							ToolCall:  tc,
							LoopCount: state.LoopCount,
						}
					}
				}

				// 获取最终响应
				if chunk.IsDone() && chunk.Response != nil {
					usage = chunk.Response.Usage
					stopReason = chunk.Response.StopReason
				}
			}

			// 创建响应
			resp := &llm.ChatResponse{
				Content:    contentBuffer,
				ToolCalls:  toolCalls,
				Usage:      usage,
				StopReason: stopReason,
			}
			state.Response = resp

			// 触发 LLM 响应钩子
			if a.config.Hooks != nil && a.config.Hooks.OnLLMResponse != nil {
				a.config.Hooks.OnLLMResponse(resp)
			}

			// 添加助手消息到历史
			assistantMsg := &llm.Message{
				Role:      llm.RoleAssistant,
				Content:   llm.Text(contentBuffer),
				ToolCalls: toolCalls,
			}
			state.Messages = append(state.Messages, assistantMsg)

			// 将助手消息添加到 memory
			if a.memory != nil {
				a.memory.Add(ctx, assistantMsg)
			}

			// 检查是否有工具调用
			if len(toolCalls) > 0 {
				result.ToolCalls = append(result.ToolCalls, toolCalls...)

				// 执行工具
				toolResults := a.executeTools(ctx, toolCalls)
				state.ToolResults = toolResults

				// 添加工具结果消息并发送事件
				for _, tr := range toolResults {
					toolMsg := &llm.Message{
						Role:       llm.RoleTool,
						ToolCallID: tr.ToolCallID,
						Name:       tr.Name,
						Content:    llm.Text(a.resultToString(tr)),
					}
					state.Messages = append(state.Messages, toolMsg)

					// 将工具结果消息添加到 memory
					if a.memory != nil {
						a.memory.Add(ctx, toolMsg)
					}

					// 发送工具结果事件
					eventCh <- &StreamEvent{
						Type:       EventTypeToolResult,
						ToolResult: tr,
						LoopCount:  state.LoopCount,
					}
				}

				// 继续循环，让 LLM 处理工具结果
				continue
			}

			// 没有工具调用，完成执行
			state.Done = true
			result.Content = contentBuffer
			result.Usage = usage
			result.StopReason = stopReason
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

		// 发送完成事件
		eventCh <- &StreamEvent{
			Type:      EventTypeDone,
			Result:    result,
			LoopCount: result.LoopCount,
		}
	}()

	return eventCh, nil
}

// callLLMStream 调用 LLM 流式 API
func (a *Agent) callLLMStream(ctx context.Context, messages []*llm.Message) (<-chan *llm.StreamChunk, error) {
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

	return a.config.LLM.Stream(ctx, req)
}
