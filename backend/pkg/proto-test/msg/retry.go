package protocol

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// DefaultMaxConcurrency 默认最大并发账号数。
// MaxConcurrency=0 时表示使用此默认值，避免大量账号同时 HTTP 登录触发服务端限流。
const DefaultMaxConcurrency = 10

// MinConcurrency 动态调整并发度时的下限。
const MinConcurrency = 1

// MaxConcurrency 最大并发账号数（0=使用 DefaultMaxConcurrency）。
var MaxConcurrency int = 0

// RetryConfig 重试与动态并发配置。
type RetryConfig struct {
	MaxRetries     int           // 最大重试次数
	RetryInterval  time.Duration // 重试间隔
	MinConcurrency int           // 最小并发度
}

// DefaultRetryConfig 返回默认重试配置。
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		RetryInterval:  500 * time.Millisecond,
		MinConcurrency: MinConcurrency,
	}
}

// GlobalRetryConfig 全局默认重试配置，可被 CLI/Wails 覆盖。
var GlobalRetryConfig = DefaultRetryConfig()

// accountIndex 从 accountID 中提取原始序号。
func accountIndex(openID, accountID string) int {
	if !strings.HasPrefix(accountID, openID) {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimPrefix(accountID, openID))
	if err != nil {
		return 0
	}
	return n
}

// AccountExecutor 执行单个账号任务的函数签名。
// 返回发送成功的消息数和错误（如有）。
type AccountExecutor func(accountID string) (int, error)

// ExecuteAccountsWithRetry 对一批账号执行任务，支持动态并发度调整和可重试错误重试。
// 返回每个账号的最终结果（成功或最终失败）。
func ExecuteAccountsWithRetry(
	ctx context.Context,
	openID string,
	idxs []int,
	initialConcurrency int,
	cfg RetryConfig,
	execute AccountExecutor,
	isRetryable func(error) bool,
) ([]accountResult, error) {
	if len(idxs) == 0 {
		return nil, nil
	}

	maxConc := initialConcurrency
	if maxConc <= 0 {
		maxConc = DefaultMaxConcurrency
	}
	if maxConc > len(idxs) {
		maxConc = len(idxs)
	}

	// 取消检查：启动前先判断
	if ctx != nil {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("重放已取消")
		default:
		}
	}

	pending := make([]int, len(idxs))
	copy(pending, idxs)
	currentConc := maxConc

	var allResults []accountResult

	for attempt := 0; attempt <= cfg.MaxRetries && len(pending) > 0; attempt++ {
		// 每轮开始前检查取消
		if ctx != nil {
			select {
			case <-ctx.Done():
				return allResults, fmt.Errorf("重放已取消")
			default:
			}
		}

		if attempt > 0 {
			log.Printf("[重放] 第 %d 次重试，剩余 %d 个账号，并发度=%d", attempt, len(pending), currentConc)
			if cfg.RetryInterval > 0 {
				if ctx != nil {
					select {
					case <-ctx.Done():
						return allResults, fmt.Errorf("重放已取消")
					case <-time.After(cfg.RetryInterval):
					}
				} else {
					time.Sleep(cfg.RetryInterval)
				}
			}
		}

		results := executeAccountBatch(openID, pending, currentConc, execute, ctx)

		var nextPending []int
		rateLimited := 0
		for _, r := range results {
			if r.err == nil {
				allResults = append(allResults, r)
				continue
			}
			if isRetryable(r.err) && attempt < cfg.MaxRetries {
				idx := accountIndex(openID, r.accountID)
				if idx > 0 {
					nextPending = append(nextPending, idx)
					rateLimited++
					continue
				}
			}
			allResults = append(allResults, r)
		}

		// 动态调整并发度：限流失败率高时降低并发
		if len(nextPending) > 0 && attempt < cfg.MaxRetries {
			failureRate := float64(rateLimited) / float64(len(pending))
			if failureRate > 0.5 {
				currentConc = max(currentConc/2, cfg.MinConcurrency)
			} else if failureRate > 0.2 {
				currentConc = max(currentConc-1, cfg.MinConcurrency)
			}
			log.Printf("[重放] 限流失败率 %.0f%%，下次重试并发度调整为 %d", failureRate*100, currentConc)
		}

		pending = nextPending
	}

	// 剩余 pending 为最终失败，补充结果
	for _, idx := range pending {
		allResults = append(allResults, accountResult{
			accountID: fmt.Sprintf("%s%d", openID, idx),
			sent:      0,
			err:       fmt.Errorf("重试 %d 次后仍失败", cfg.MaxRetries),
		})
	}

	return allResults, nil
}

// SendMessagesWithRetry 并发发送消息，支持动态并发度调整和限流失败重试（向后兼容包装）。
// 实际实现已迁移到 Replayer.SendMessagesWithRetry。
func SendMessagesWithRetry(
	serverAddr, httpAddr, openID string,
	messages []RecordMessage,
	repeatCount int,
	rangeStart, rangeEnd int,
	ctx context.Context,
	onProgress ReplayProgressCallback,
	onMessage ReplayMessageCallback,
	connPool *AccountConnectionPool,
	cfg RetryConfig,
	initialConcurrency int,
) error {
	return NewReplayer(ReplayOptions{
		ServerAddr:  serverAddr,
		HTTPAddr:    httpAddr,
		OpenID:      openID,
		Messages:    messages,
		RepeatCount: repeatCount,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
		Context:     ctx,
		OnProgress:  onProgress,
		OnMessage:   onMessage,
		ConnPool:    connPool,
		RetryConfig: cfg,
		Concurrency: initialConcurrency,
	}).SendMessagesWithRetry()
}

// executeAccountBatch 对指定账号索引列表执行单轮任务。
// 返回每个账号的执行结果，已启动的 goroutine 在 ctx 取消时仍会完成并收集结果。
func executeAccountBatch(
	openID string,
	accountIdxs []int,
	concurrency int,
	execute AccountExecutor,
	ctx context.Context,
) []accountResult {
	accountCount := len(accountIdxs)
	sem := make(chan struct{}, concurrency)
	resultCh := make(chan accountResult, accountCount)

	started := 0
	for _, idx := range accountIdxs {
		// 检查取消：停止启动新的 goroutine，但完成已启动的并收集结果
		if ctx != nil {
			select {
			case <-ctx.Done():
				results := make([]accountResult, 0, started)
				for range started {
					results = append(results, <-resultCh)
				}
				return results
			default:
			}
		}

		accountID := fmt.Sprintf("%s%d", openID, idx)
		sem <- struct{}{}
		started++
		go func(aid string) {
			defer func() { <-sem }()
			log.Printf("[重放] [account=%s] 开始发送", aid)
			sent, err := execute(aid)
			if err != nil {
				log.Printf("[重放] [account=%s] 发送失败: %v", aid, err)
			} else {
				log.Printf("[重放] [account=%s] 发送完成: %d 条", aid, sent)
			}
			resultCh <- accountResult{accountID: aid, sent: sent, err: err}
		}(accountID)
	}

	return collectResults(resultCh, accountCount)
}

// collectResults 从 resultCh 中收集指定数量的结果。
func collectResults(resultCh <-chan accountResult, count int) []accountResult {
	results := make([]accountResult, 0, count)
	for range count {
		results = append(results, <-resultCh)
	}
	return results
}
