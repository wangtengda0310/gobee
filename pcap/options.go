package pcap

// Option 是 functional option，用于在 NewCapturer 时修改默认配置。
// 沿用本仓库 agent 模块的配置风格（Option func(*Config) + WithXxx）。
type Option func(*Config)

// WithBufferSize 设置每个处理函数的有界队列容量。
// 默认 1024。设置过小会导致 OverflowDrop 策略下大量丢包；
// 设置过大会增加内存占用并放大背压延迟（处理函数看到的包越来越「旧」）。
// 传入非正值会被忽略，保留默认值。
func WithBufferSize(n int) Option {
	return func(c *Config) {
		if n > 0 {
			c.BufferSize = n
		}
	}
}

// WithOverflowStrategy 设置背压（队列满）时的应对策略。默认 OverflowDrop。
// 实时抓包建议保持 OverflowDrop；离线分析（读 pcap 文件）可改用 OverflowBlock。
func WithOverflowStrategy(s OverflowStrategy) Option {
	return func(c *Config) {
		c.Overflow = s
	}
}

// WithBPFFilter 设置 BPF 过滤表达式（如 "tcp port 80 and host itsnot.fun"）。
// 在数据源层过滤，是降低投递压力最有效的手段。
// 注意：并非所有 Source 都支持（见 Target.BPF 说明）。
func WithBPFFilter(expr string) Option {
	return func(c *Config) {
		c.BPFFilter = expr
	}
}

// WithHooks 设置生命周期回调。所有回调字段可选（nil 即不触发）。
func WithHooks(h Hooks) Option {
	return func(c *Config) {
		c.Hooks = h
	}
}
