package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected client to be created")
	}
	if len(client.Responses) != 0 {
		t.Error("expected empty responses queue")
	}
}

func TestClient_Complete_DefaultResponse(t *testing.T) {
	client := NewClient()

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("Hello")},
		},
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.Content != "default mock response" {
		t.Errorf("expected default response, got '%s'", resp.Content)
	}

	if client.CallCount != 1 {
		t.Errorf("expected CallCount 1, got %d", client.CallCount)
	}
}

func TestClient_Complete_PresetResponses(t *testing.T) {
	client := NewClient()

	// 添加预设响应
	client.AddTextResponse("first response")
	client.AddTextResponse("second response")

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("Hello")},
		},
	}

	// 第一次调用
	resp1, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp1.Content != "first response" {
		t.Errorf("expected 'first response', got '%s'", resp1.Content)
	}

	// 第二次调用
	resp2, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp2.Content != "second response" {
		t.Errorf("expected 'second response', got '%s'", resp2.Content)
	}

	// 第三次调用（超出预设，返回默认）
	resp3, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp3.Content != "default mock response" {
		t.Errorf("expected default response, got '%s'", resp3.Content)
	}
}

func TestClient_Complete_ToolCall(t *testing.T) {
	client := NewClient()

	client.AddToolCallResponse("get_time", `{"timezone": "UTC"}`)

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("What time is it?")},
		},
	}

	resp, err := client.Complete(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp.StopReason != llm.StopReasonToolUse {
		t.Errorf("expected StopReasonToolUse, got %v", resp.StopReason)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}

	if resp.ToolCalls[0].Function.Name != "get_time" {
		t.Errorf("expected tool name 'get_time', got '%s'", resp.ToolCalls[0].Function.Name)
	}
}

func TestClient_Complete_Error(t *testing.T) {
	client := NewClient()
	client.SetError(errors.New("API error"))

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("Hello")},
		},
	}

	_, err := client.Complete(context.Background(), req)
	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != "API error" {
		t.Errorf("expected 'API error', got '%s'", err.Error())
	}
}

func TestClient_Stream_Default(t *testing.T) {
	client := NewClient()

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("Hello")},
		},
	}

	ch, err := client.Stream(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var chunks []*llm.StreamChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}

	if len(chunks) == 0 {
		t.Error("expected at least one chunk")
	}
}

func TestClient_LastRequest(t *testing.T) {
	client := NewClient()

	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("test message")},
		},
		MaxTokens: 100,
	}

	_, _ = client.Complete(context.Background(), req)

	lastReq := client.GetLastRequest()
	if lastReq == nil {
		t.Fatal("expected last request to be recorded")
	}

	if lastReq.MaxTokens != 100 {
		t.Errorf("expected MaxTokens 100, got %d", lastReq.MaxTokens)
	}
}

func TestClient_Reset(t *testing.T) {
	client := NewClient()

	client.AddTextResponse("response")
	client.SetError(errors.New("error"))
	_, _ = client.Complete(context.Background(), &llm.ChatRequest{})

	client.Reset()

	if client.CallCount != 0 {
		t.Errorf("expected CallCount 0 after reset, got %d", client.CallCount)
	}
	if len(client.Responses) != 0 {
		t.Error("expected empty responses after reset")
	}
	if client.Error != nil {
		t.Error("expected nil error after reset")
	}
}

func TestClient_GetCallCount(t *testing.T) {
	client := NewClient()

	if client.GetCallCount() != 0 {
		t.Errorf("expected initial CallCount 0, got %d", client.GetCallCount())
	}

	_, _ = client.Complete(context.Background(), &llm.ChatRequest{})
	_, _ = client.Complete(context.Background(), &llm.ChatRequest{})

	if client.GetCallCount() != 2 {
		t.Errorf("expected CallCount 2, got %d", client.GetCallCount())
	}
}

func TestClient_AddResponse(t *testing.T) {
	client := NewClient()

	resp := &llm.ChatResponse{
		Content:    "custom response",
		StopReason: llm.StopReasonEndTurn,
		Usage:      &llm.Usage{InputTokens: 10, OutputTokens: 20},
	}

	client.AddResponse(resp)

	if len(client.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(client.Responses))
	}

	result, err := client.Complete(context.Background(), &llm.ChatRequest{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result.Content != "custom response" {
		t.Errorf("expected 'custom response', got '%s'", result.Content)
	}

	if result.Usage.InputTokens != 10 {
		t.Errorf("expected InputTokens 10, got %d", result.Usage.InputTokens)
	}
}
