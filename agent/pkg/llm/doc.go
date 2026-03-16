// Package llm 提供统一的大语言模型适配接口。
//
// 支持多种 LLM 提供商:
//   - OpenAI (GPT-4, GPT-3.5)
//   - Anthropic (Claude)
//   - 本地模型 (Ollama, LocalAI)
//
// 核心接口:
//
//	type ChatCompleter interface {
//	    Complete(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
//	    Stream(ctx context.Context, req *ChatRequest) (<-chan ChatChunk, error)
//	}
package llm
