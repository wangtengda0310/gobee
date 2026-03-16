// Package memory 提供对话记忆管理。
//
// 功能:
//   - 会话持久化 (Session Persistence)
//   - 上下文压缩 (Context Compression)
//   - 向量存储集成 (Vector Store)
//   - 滑动窗口 (Sliding Window)
//
// 使用示例:
//
//	mem := memory.NewSlidingWindow(100)
//	mem.Add(ctx, message)
//	context := mem.GetContext(ctx)
package memory
