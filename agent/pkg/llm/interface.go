package llm

import "context"

// ChatCompleter 聊天补全接口
//
// 这是 LLM 适配器的核心接口，所有提供商都需要实现此接口。
// 支持流式和非流式两种响应模式。
type ChatCompleter interface {
	// Complete 发送非流式聊天补全请求
	//
	// ctx: 上下文，用于取消和超时控制
	// req: 聊天请求
	// 返回: 聊天响应或错误
	Complete(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// Stream 发送流式聊天补全请求
	//
	// ctx: 上下文，用于取消和超时控制
	// req: 聊天请求 (req.Stream 会被设为 true)
	// 返回: 流式数据块通道或错误
	//
	// 使用示例:
	//   stream, err := client.Stream(ctx, req)
	//   for chunk := range stream {
	//       if chunk.IsError() {
	//           return chunk.Error
	//       }
	//       fmt.Print(chunk.Content)
	//       if chunk.IsDone() {
	//           break
	//       }
	//   }
	Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)
}

// ModelInfo 模型信息接口
type ModelInfo interface {
	// ModelName 返回当前使用的模型名称
	ModelName() string

	// ProviderName 返回提供商名称
	ProviderName() string
}

// Client 完整的 LLM 客户端接口
type Client interface {
	ChatCompleter
	ModelInfo
}
