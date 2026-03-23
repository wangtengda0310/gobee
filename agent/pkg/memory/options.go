package memory

// Option Memory 配置选项
type Option func(*SlidingWindow)

// WithMaxSize 设置最大消息数量
func WithMaxSize(size int) Option {
	return func(s *SlidingWindow) {
		if size > 0 {
			s.maxSize = size
		}
	}
}

// WithPreserveSystem 设置是否保留系统消息
func WithPreserveSystem(preserve bool) Option {
	return func(s *SlidingWindow) {
		s.preserveSys = preserve
	}
}

// WithCompressor 设置压缩器
func WithCompressor(compressor Compressor) Option {
	return func(s *SlidingWindow) {
		s.compressor = compressor
	}
}

// SessionOption 会话配置选项
type SessionOption func(*Session)

// WithMetadata 设置会话元数据
func WithMetadata(key string, value any) SessionOption {
	return func(s *Session) {
		if s.Metadata == nil {
			s.Metadata = make(map[string]any)
		}
		s.Metadata[key] = value
	}
}
