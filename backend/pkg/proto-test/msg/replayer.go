package protocol

import (
	"context"
	"fmt"
	"log"
)

// ReplayOptions 描述一次协议重放任务的所有配置。
// 推荐使用 NewReplayer(opts).SendMessages() 执行；旧的包级函数 SendMessages /
// SendMessagesWithRetry 仍保留为向后兼容的包装。
type ReplayOptions struct {
	// 连接配置
	ServerAddr string // TCP 游戏服务器地址
	HTTPAddr   string // HTTP 登录服务地址；空表示由调用方或 Authenticator 推导

	// 账号范围
	OpenID     string // 账号前缀
	RangeStart int    // 起始账号序号（含）
	RangeEnd   int    // 结束账号序号（含）

	// 用例配置
	Messages    []RecordMessage // 要重放的消息列表
	RepeatCount int             // 每账号重复轮数；<=0 时按 1 处理

	// 运行时回调与资源
	Context    context.Context        // 可取消的上下文；nil 时使用 context.Background()
	OnProgress ReplayProgressCallback // 进度回调（可选）
	OnMessage  ReplayMessageCallback  // 每收到一条服务端消息时的回调（可选）
	ConnPool   *AccountConnectionPool // 账号连接池（可选）

	// 认证器（可选）
	// nil 时使用基于 ServerAddr/HTTPAddr 的 HTTPAuthenticator。
	// 注入自定义实现可替换登录逻辑，也便于测试。
	Authenticator Authenticator

	// 并发与重试策略
	RetryConfig RetryConfig // 重试配置；零值时使用 GlobalRetryConfig
	Concurrency int         // 初始并发度；<=0 时使用 DefaultMaxConcurrency
}

// normalize 填充默认值并校验基本约束。
func (opts *ReplayOptions) normalize() error {
	if opts.RepeatCount <= 0 {
		opts.RepeatCount = 1
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = DefaultMaxConcurrency
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.RetryConfig.MaxRetries == 0 && opts.RetryConfig.RetryInterval == 0 && opts.RetryConfig.MinConcurrency == 0 {
		opts.RetryConfig = GlobalRetryConfig
	}
	if opts.RetryConfig.MinConcurrency <= 0 {
		opts.RetryConfig.MinConcurrency = MinConcurrency
	}
	if opts.Authenticator == nil {
		opts.Authenticator = &HTTPAuthenticator{ServerAddr: opts.ServerAddr, HTTPAddr: opts.HTTPAddr}
	}
	// 连接池纳入 Authenticator IoC：用 PooledAuthenticator 包装内部 authenticator
	if opts.ConnPool != nil {
		opts.Authenticator = NewPooledAuthenticator(opts.Authenticator, opts.ConnPool, opts.ServerAddr, opts.HTTPAddr)
	}
	if opts.RangeStart > opts.RangeEnd {
		return fmt.Errorf("账号范围无效: RangeStart=%d > RangeEnd=%d", opts.RangeStart, opts.RangeEnd)
	}
	return nil
}

// Replayer 负责执行协议重放任务。
// 它持有 ReplayOptions 的副本，执行过程中可安全修改内部状态而不影响调用方。
type Replayer struct {
	opts ReplayOptions
}

// NewReplayer 创建 Replayer。
func NewReplayer(opts ReplayOptions) *Replayer {
	return &Replayer{opts: opts}
}

// SendMessages 使用 opts 中的重试配置执行重放。
// 是 SendMessagesWithRetry 的别名，保留以更贴近旧 API 语义。
func (r *Replayer) SendMessages() error {
	return r.SendMessagesWithRetry()
}

// SendMessagesWithRetry 执行重放，支持动态并发度调整和限流失败重试。
func (r *Replayer) SendMessagesWithRetry() error {
	if err := r.opts.normalize(); err != nil {
		return err
	}

	opts := r.opts
	accountCount := opts.RangeEnd - opts.RangeStart + 1
	total := accountCount * len(opts.Messages) * opts.RepeatCount
	maxConc := opts.Concurrency
	if maxConc > accountCount {
		maxConc = accountCount
	}
	log.Printf("[重放] 账号数: %d (前缀=%s, 范围=%d~%d), 消息数: %d, 重复次数: %d, 总发送计划: %d, 初始并发: %d, 最大重试: %d",
		accountCount, opts.OpenID, opts.RangeStart, opts.RangeEnd, len(opts.Messages), opts.RepeatCount, total, maxConc, opts.RetryConfig.MaxRetries)

	idxs := make([]int, 0, accountCount)
	for i := opts.RangeStart; i <= opts.RangeEnd; i++ {
		idxs = append(idxs, i)
	}

	globalSent := 0
	executor := func(accountID string) (int, error) {
		runOpts := accountRunOptions{
			ReplayOptions: opts,
			accountID:     accountID,
			auth:          opts.Authenticator,
			grandTotal:    total,
			alreadySent:   globalSent,
		}
		sent, err := sendMessagesOnce(runOpts)
		if err == nil {
			globalSent += sent
		}
		return sent, err
	}

	results, err := ExecuteAccountsWithRetry(opts.Context, opts.OpenID, idxs, maxConc, opts.RetryConfig, executor, IsRateLimitError)
	if err != nil {
		return err
	}

	var permanentFailed []string
	for _, res := range results {
		if res.err != nil {
			permanentFailed = append(permanentFailed, res.accountID)
		}
	}

	if len(permanentFailed) > 0 {
		log.Printf("[重放] 完成: %d/%d 个账号成功, %d 个失败: %v, 共发送 %d 条",
			accountCount-len(permanentFailed), accountCount, len(permanentFailed), permanentFailed, globalSent)
		return fmt.Errorf("%d 个账号失败: %v (%d/%d 成功, 共发送 %d 条)",
			len(permanentFailed), permanentFailed, accountCount-len(permanentFailed), accountCount, globalSent)
	}

	log.Printf("[重放] 全部完成: %d 个账号, 共发送 %d 条", accountCount, globalSent)
	return nil
}
