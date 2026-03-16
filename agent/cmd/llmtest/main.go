package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"agent/pkg/llm"
	"agent/pkg/llm/anthropic"
	"agent/pkg/llm/openai"
)

func main() {
	// 命令行参数
	provider := flag.String("provider", "anthropic", "LLM 提供商: anthropic, openai")
	prompt := flag.String("prompt", "Say hello in one word", "提示词")
	stream := flag.Bool("stream", false, "使用流式响应")
	model := flag.String("model", "", "模型名称 (留空使用默认)")
	baseURL := flag.String("base-url", "", "API Base URL (留空使用默认)")
	maxTokens := flag.Int("max-tokens", 1024, "最大 token 数")
	timeout := flag.Duration("timeout", 60*time.Second, "请求超时")
	image := flag.String("image", "", "图像文件路径 (base64)")
	imageType := flag.String("image-type", "image/png", "图像 MIME 类型")
	flag.Parse()

	// 创建客户端
	var client llm.ChatCompleter
	var err error

	switch strings.ToLower(*provider) {
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "错误: 请设置 ANTHROPIC_API_KEY 环境变量")
			os.Exit(1)
		}
		opts := []anthropic.Option{
			anthropic.WithAPIKey(apiKey),
			anthropic.WithTimeout(*timeout),
		}
		if *model != "" {
			opts = append(opts, anthropic.WithModel(*model))
		}
		if *baseURL != "" {
			opts = append(opts, anthropic.WithBaseURL(*baseURL))
		}
		if *maxTokens > 0 {
			opts = append(opts, anthropic.WithMaxTokens(*maxTokens))
		}
		client, err = anthropic.NewClient(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建 Anthropic 客户端失败: %v\n", err)
			os.Exit(1)
		}

	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "错误: 请设置 OPENAI_API_KEY 环境变量")
			os.Exit(1)
		}
		opts := []openai.Option{
			openai.WithAPIKey(apiKey),
			openai.WithTimeout(*timeout),
		}
		if *model != "" {
			opts = append(opts, openai.WithModel(*model))
		}
		if *baseURL != "" {
			opts = append(opts, openai.WithBaseURL(*baseURL))
		}
		if *maxTokens > 0 {
			opts = append(opts, openai.WithMaxTokens(*maxTokens))
		}
		client, err = openai.NewClient(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建 OpenAI 客户端失败: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "不支持的提供商: %s\n", *provider)
		os.Exit(1)
	}

	// 构建请求
	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{
				Role:    llm.RoleUser,
				Content: llm.Text(*prompt),
			},
		},
	}

	// 添加图像 (如果有)
	if *image != "" {
		imageData, err := os.ReadFile(*image)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取图像文件失败: %v\n", err)
			os.Exit(1)
		}
		content := llm.NewContentList(
			llm.Text(*prompt),
			llm.ImageFromBase64Bytes(imageData, *imageType),
		)
		req.Messages[0].Content = content
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// 发送请求
	if *stream {
		err = streamRequest(ctx, client, req, *provider)
	} else {
		err = completeRequest(ctx, client, req, *provider)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		os.Exit(1)
	}
}

func completeRequest(ctx context.Context, client llm.ChatCompleter, req *llm.ChatRequest, provider string) error {
	fmt.Printf("[%s] 发送: %q\n", provider, llm.TextString(req.Messages[0].Content))

	start := time.Now()
	resp, err := client.Complete(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] 响应: %q\n", provider, resp.Content)
	fmt.Printf("[%s] 耗时: %v\n", provider, time.Since(start))

	if resp.Usage != nil {
		fmt.Printf("[%s] Token 使用: 输入=%d, 输出=%d, 总计=%d\n",
			provider, resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.TotalTokens)
	}

	if len(resp.ToolCalls) > 0 {
		fmt.Printf("[%s] 工具调用:\n", provider)
		for _, tc := range resp.ToolCalls {
			fmt.Printf("  - %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
		}
	}

	return nil
}

func streamRequest(ctx context.Context, client llm.ChatCompleter, req *llm.ChatRequest, provider string) error {
	fmt.Printf("[%s] 发送 (流式): %q\n", provider, llm.TextString(req.Messages[0].Content))
	fmt.Printf("[%s] 响应: ", provider)

	start := time.Now()
	stream, err := client.Stream(ctx, req)
	if err != nil {
		return err
	}

	var fullContent string
	var finalResp *llm.ChatResponse

	for chunk := range stream {
		if chunk.IsError() {
			fmt.Printf("\n[%s] 错误: %v\n", provider, chunk.Error)
			return chunk.Error
		}

		if chunk.Content != "" {
			fmt.Print(chunk.Content)
			fullContent += chunk.Content
		}

		if chunk.IsDone() {
			finalResp = chunk.Response
			break
		}
	}

	fmt.Println()
	fmt.Printf("[%s] 耗时: %v\n", provider, time.Since(start))

	if finalResp != nil && finalResp.Usage != nil {
		fmt.Printf("[%s] Token 使用: 输入=%d, 输出=%d, 总计=%d\n",
			provider, finalResp.Usage.InputTokens, finalResp.Usage.OutputTokens, finalResp.Usage.TotalTokens)
	}

	return nil
}
