// Package memory 提供对话记忆管理。
//
// 核心概念:
//   - Memory: 记忆接口，管理对话历史
//   - SlidingWindow: 滑动窗口实现，保留最近 N 条消息
//   - Session: 会话，包含完整的对话上下文
//   - Compressor: 压缩器，用于压缩长对话
//
// 使用示例:
//
//	// 创建滑动窗口记忆
//	mem := memory.NewSlidingWindow(50,
//	    memory.WithPreserveSystem(true),
//	)
//
//	// 添加消息
//	mem.Add(ctx, &llm.Message{
//	    Role:    llm.RoleUser,
//	    Content: llm.Text("Hello"),
//	})
//
//	// 获取上下文
//	msgs, _ := mem.GetContext(ctx)
//
// # 持久化
//
// 实现 PersistentMemory 接口可以支持会话持久化:
//
//	type FileMemory struct {
//	    memory.Memory
//	    filePath string
//	}
//
//	func (m *FileMemory) Save(ctx context.Context) error {
//	    // 序列化并保存到文件
//	}
//
//	func (m *FileMemory) Load(ctx context.Context) error {
//	    // 从文件加载并反序列化
//	}
package memory
