// Package llm 提供统一的大语言模型适配接口。
//
// 支持多种 LLM 提供商:
//   - OpenAI (GPT-4, GPT-4o)
//   - Anthropic (Claude)
//
// 核心接口:
//
//	type ChatCompleter interface {
//	    Complete(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
//	    Stream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)
//	}
//
// 使用示例:
//
//	// 创建 Anthropic 客户端
//	client, err := anthropic.NewClient(
//	    anthropic.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
//	    anthropic.WithModel("claude-sonnet-4-20250514"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 发送请求
//	req := &llm.ChatRequest{
//	    Messages: []*llm.Message{
//	        {Role: llm.RoleUser, Content: llm.Text("Hello!")},
//	    },
//	}
//	resp, err := client.Complete(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Content)
//
// 多模态支持:
//
//	// 发送图像和文本
//	content := llm.NewContentList(
//	    llm.Text("What's in this image?"),
//	    llm.ImageFromBase64(imageData, "image/png"),
//	)
//	req.Messages[0].Content = content
//
// 流式响应:
//
//	stream, err := client.Stream(ctx, req)
//	for chunk := range stream {
//	    if chunk.IsError() {
//	        log.Fatal(chunk.Error)
//	    }
//	    fmt.Print(chunk.Content)
//	    if chunk.IsDone() {
//	        break
//	    }
//	}
package llm
