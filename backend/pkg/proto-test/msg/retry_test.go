package protocol

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRateLimitError(t *testing.T) {
	assert.True(t, IsRateLimitError(&AuthError{Code: -600, Message: "登录失败"}))
	assert.False(t, IsRateLimitError(&AuthError{Code: -1, Message: "登录失败"}))
	assert.False(t, IsRateLimitError(errors.New("其他错误")))
	assert.False(t, IsRateLimitError(nil))
}

func TestExecuteAccountsWithRetry_AllSuccess(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:     2,
		RetryInterval:  10 * time.Millisecond,
		MinConcurrency: 1,
	}

	var callCount atomic.Int32
	executor := func(accountID string) (int, error) {
		callCount.Add(1)
		return 1, nil
	}

	results, err := ExecuteAccountsWithRetry(context.Background(), "test", []int{1, 2, 3}, 2, cfg, executor, IsRateLimitError)
	require.NoError(t, err)
	require.Len(t, results, 3)
	for _, r := range results {
		assert.NoError(t, r.err)
		assert.Equal(t, 1, r.sent)
	}
	assert.Equal(t, int32(3), callCount.Load())
}

func TestExecuteAccountsWithRetry_RateLimitRetry(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:     2,
		RetryInterval:  10 * time.Millisecond,
		MinConcurrency: 1,
	}

	// 第一次调用 test1/test2 限流失败，第二次成功
	failOnce := map[string]bool{"test1": true, "test2": true}
	var mu sync.Mutex
	executor := func(accountID string) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		if failOnce[accountID] {
			failOnce[accountID] = false
			return 0, &AuthError{Code: -600, Message: "登录失败"}
		}
		return 1, nil
	}

	results, err := ExecuteAccountsWithRetry(context.Background(), "test", []int{1, 2, 3}, 3, cfg, executor, IsRateLimitError)
	require.NoError(t, err)
	require.Len(t, results, 3)
	for _, r := range results {
		assert.NoError(t, r.err, "账号 %s 应该最终成功", r.accountID)
		assert.Equal(t, 1, r.sent)
	}
}

func TestExecuteAccountsWithRetry_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:     2,
		RetryInterval:  10 * time.Millisecond,
		MinConcurrency: 1,
	}

	executor := func(accountID string) (int, error) {
		if accountID == "test2" {
			return 0, errors.New("不可重试错误")
		}
		return 1, nil
	}

	results, err := ExecuteAccountsWithRetry(context.Background(), "test", []int{1, 2, 3}, 3, cfg, executor, IsRateLimitError)
	require.NoError(t, err)
	require.Len(t, results, 3)

	failures := 0
	for _, r := range results {
		if r.accountID == "test2" {
			assert.Error(t, r.err)
			failures++
		} else {
			assert.NoError(t, r.err)
		}
	}
	assert.Equal(t, 1, failures)
}

func TestExecuteAccountsWithRetry_DynamicConcurrencyReduction(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:     3,
		RetryInterval:  10 * time.Millisecond,
		MinConcurrency: 1,
	}

	// 第一轮：全部 4 个都限流失败
	// 第二轮：并发度应降到 2，仍全部失败
	// 第三轮：并发度应降到 1，全部成功
	attemptsForAccount := map[string]int{}
	var mu sync.Mutex
	executor := func(accountID string) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		attemptsForAccount[accountID]++
		if attemptsForAccount[accountID] < 3 {
			return 0, &AuthError{Code: -600, Message: "登录失败"}
		}
		return 1, nil
	}

	results, err := ExecuteAccountsWithRetry(context.Background(), "test", []int{1, 2, 3, 4}, 4, cfg, executor, IsRateLimitError)
	require.NoError(t, err)
	require.Len(t, results, 4)
	for _, r := range results {
		assert.NoError(t, r.err)
	}
}

func TestExecuteAccountsWithRetry_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxRetries:     3,
		RetryInterval:  1 * time.Second,
		MinConcurrency: 1,
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	executor := func(accountID string) (int, error) {
		// 阻塞直到 ctx 取消
		<-ctx.Done()
		return 0, ctx.Err()
	}

	// 立刻取消
	cancel()

	results, err := ExecuteAccountsWithRetry(ctx, "test", []int{1, 2}, 2, cfg, executor, IsRateLimitError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "重放已取消")
	assert.Empty(t, results)
}

func TestExecuteAccountsWithRetry_EmptyTasks(t *testing.T) {
	results, err := ExecuteAccountsWithRetry(context.Background(), "test", nil, 10, DefaultRetryConfig(), nil, IsRateLimitError)
	require.NoError(t, err)
	assert.Empty(t, results)
}
