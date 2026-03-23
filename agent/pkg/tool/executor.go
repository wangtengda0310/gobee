package tool

import (
	"context"
	"sync"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

// BatchExecutor 批量工具执行器
// 支持并行执行多个工具调用
type BatchExecutor struct {
	registry *Registry
	workers  int // 并行执行的最大 worker 数
}

// NewBatchExecutor 创建批量执行器
// registry: 工具注册表
// workers: 并行执行的最大数量，默认为 4
func NewBatchExecutor(registry *Registry, workers int) *BatchExecutor {
	if workers <= 0 {
		workers = 4
	}
	return &BatchExecutor{
		registry: registry,
		workers:  workers,
	}
}

// Execute 执行单个工具调用
func (e *BatchExecutor) Execute(ctx context.Context, name string, args map[string]any) (any, error) {
	return e.registry.Execute(ctx, name, args)
}

// ExecuteBatch 并行执行多个工具调用
// 返回所有执行结果
func (e *BatchExecutor) ExecuteBatch(ctx context.Context, calls []*llm.ToolCall) *BatchResult {
	if len(calls) == 0 {
		return &BatchResult{}
	}

	results := make([]*ToolResult, len(calls))
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 使用 semaphore 控制并发数
	sem := make(chan struct{}, e.workers)

	for i, call := range calls {
		wg.Add(1)
		go func(idx int, tc *llm.ToolCall) {
			defer wg.Done()

			// 获取信号量
			sem <- struct{}{}
			defer func() { <-sem }()

			result := &ToolResult{
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
			}

			// 执行工具
			res, err := e.registry.Execute(ctx, tc.Function.Name, nil)
			if err != nil {
				result.Error = err
			} else {
				result.Result = res
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, call)
	}

	wg.Wait()

	// 统计结果
	batch := &BatchResult{Results: results}
	for _, r := range results {
		if r.IsError() {
			batch.ErrorCount++
		} else {
			batch.SuccessCount++
		}
	}

	return batch
}

// ExecuteSequential 顺序执行多个工具调用
// 适用于工具之间有依赖关系的场景
func (e *BatchExecutor) ExecuteSequential(ctx context.Context, calls []*llm.ToolCall) *BatchResult {
	if len(calls) == 0 {
		return &BatchResult{}
	}

	results := make([]*ToolResult, len(calls))
	batch := &BatchResult{Results: results}

	for i, call := range calls {
		result := &ToolResult{
			ToolCallID: call.ID,
			Name:       call.Function.Name,
		}

		res, err := e.registry.Execute(ctx, call.Function.Name, nil)
		if err != nil {
			result.Error = err
			batch.ErrorCount++
		} else {
			result.Result = res
			batch.SuccessCount++
		}

		results[i] = result
	}

	return batch
}

// Registry 返回底层注册表
func (e *BatchExecutor) Registry() *Registry {
	return e.registry
}
