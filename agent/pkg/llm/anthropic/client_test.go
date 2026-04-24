package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wangtengda0310/gobee/agent/pkg/llm"
)

func TestBuildEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "Anthropic 官方 API - BaseURL 包含 /v1",
			baseURL:  "https://api.anthropic.com/v1",
			path:     "/messages",
			expected: "https://api.anthropic.com/v1/messages",
		},
		{
			name:     "智谱 AI - BaseURL 不包含 /v1",
			baseURL:  "https://open.bigmodel.cn/api/anthropic",
			path:     "/messages",
			expected: "https://open.bigmodel.cn/api/anthropic/v1/messages",
		},
		{
			name:     "用户配置完整路径 - BaseURL 包含 /v1",
			baseURL:  "https://open.bigmodel.cn/api/anthropic/v1",
			path:     "/messages",
			expected: "https://open.bigmodel.cn/api/anthropic/v1/messages",
		},
		{
			name:     "BaseURL 末尾有斜杠",
			baseURL:  "https://api.anthropic.com/v1/",
			path:     "/messages",
			expected: "https://api.anthropic.com/v1/messages",
		},
		{
			name:     "BaseURL 末尾有斜杠且不含 /v1",
			baseURL:  "https://open.bigmodel.cn/api/anthropic/",
			path:     "/messages",
			expected: "https://open.bigmodel.cn/api/anthropic/v1/messages",
		},
		{
			name:     "自定义端点路径",
			baseURL:  "https://api.example.com/v1",
			path:     "/custom",
			expected: "https://api.example.com/v1/custom",
		},
		{
			name:     "不含 /v1 的自定义 BaseURL",
			baseURL:  "https://api.example.com/api",
			path:     "/chat",
			expected: "https://api.example.com/api/v1/chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildEndpoint(tt.baseURL, tt.path)
			if result != tt.expected {
				t.Errorf("buildEndpoint(%q, %q) = %q, want %q", tt.baseURL, tt.path, result, tt.expected)
			}
		})
	}
}

func TestBuildEndpoint_EdgeCases(t *testing.T) {
	// 测试边界情况
	if result := buildEndpoint("", "/messages"); result != "/v1/messages" {
		t.Errorf("empty baseURL should still work, got %q", result)
	}

	if result := buildEndpoint("https://api.example.com", ""); result != "https://api.example.com/v1" {
		t.Errorf("empty path should work, got %q", result)
	}

	if result := buildEndpoint("https://api.example.com/v1", ""); result != "https://api.example.com/v1" {
		t.Errorf("empty path with /v1 baseURL should work, got %q", result)
	}
}

// === 系统提示词丢失 Bug 回归测试 ===
//
// Bug 描述: agent.WithSystemPrompt() 配合 Anthropic provider 时，系统提示词被静默丢弃。
//
// 根因（三个环节配合断裂）:
//   1. Run()/RunStream() 将系统提示词以 RoleSystem 消息放入 Messages 数组
//   2. callLLM()/callLLMStream() 构建 ChatRequest 时未设置 System 字段
//   3. convertMessages() 跳过所有 RoleSystem 消息（见下方 TestConvertMessages_SkipsSystemRole）
//   → 结果: 系统提示词既不在顶层 system 字段，也不在 messages 数组中，被完全丢弃
//
// 修复: 在 callLLM()/callLLMStream() 中将 a.config.SystemPrompt 设置到 req.System
// 本文件验证 Anthropic converter 层的行为：System 字段被保留，RoleSystem 消息被过滤
//
// 关联测试:
//   - agent_test.go: TestAgent_SystemPromptPassedToChatRequest_* 验证 Agent 层修复
//   - 本文件: TestConvertMessages_SkipsSystemRole 验证 converter 层的过滤行为（bug 根因）

// TestConvertRequest_SystemFieldPassed 验证 convertRequest 正确传递 System 字段
func TestConvertRequest_SystemFieldPassed(t *testing.T) {
	req := &llm.ChatRequest{
		System: "你是一个专业的翻译助手",
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("你好")},
		},
	}

	anthReq := convertRequest(req)

	assert.Equal(t, "你是一个专业的翻译助手", anthReq.System,
		"Anthropic ChatRequest.System 应该等于原始请求的 System 字段")
}

// TestConvertRequest_SystemFieldEmptyWhenNotSet 验证未设置时 System 为空
func TestConvertRequest_SystemFieldEmptyWhenNotSet(t *testing.T) {
	req := &llm.ChatRequest{
		Messages: []*llm.Message{
			{Role: llm.RoleUser, Content: llm.Text("你好")},
		},
	}

	anthReq := convertRequest(req)

	assert.Empty(t, anthReq.System,
		"未设置 System 时应该为空")
}

// TestConvertMessages_SkipsSystemRole 验证 convertMessages 丢弃 RoleSystem 消息
// 这是 bug 的关键点：如果 system prompt 仅通过 Messages 传递，Anthropic converter 会完全丢弃它
func TestConvertMessages_SkipsSystemRole(t *testing.T) {
	messages := []*llm.Message{
		{Role: llm.RoleSystem, Content: llm.Text("你是助手")},
		{Role: llm.RoleUser, Content: llm.Text("你好")},
		{Role: llm.RoleAssistant, Content: llm.Text("你好！")},
		{Role: llm.RoleUser, Content: llm.Text("再见")},
	}

	result := convertMessages(messages)

	assert.Len(t, result, 3,
		"convertMessages 应该过滤掉 RoleSystem 消息，只保留 3 条")
	assert.Equal(t, "user", result[0].Role)
	assert.Equal(t, "assistant", result[1].Role)
	assert.Equal(t, "user", result[2].Role)
}

// TestConvertRequest_SystemPromptOnlyInTopLevelField 验证完整的端到端场景：
// system prompt 在顶层 System 字段中被保留，在 Messages 中被过滤
func TestConvertRequest_SystemPromptOnlyInTopLevelField(t *testing.T) {
	systemPrompt := "你是一个代码审查助手，请检查安全问题。"
	req := &llm.ChatRequest{
		System: systemPrompt,
		Messages: []*llm.Message{
			{Role: llm.RoleSystem, Content: llm.Text(systemPrompt)},
			{Role: llm.RoleUser, Content: llm.Text("请检查这段代码")},
		},
	}

	anthReq := convertRequest(req)

	// 顶层 System 字段保留
	assert.Equal(t, systemPrompt, anthReq.System,
		"顶层 System 字段应保留系统提示词")

	// Messages 中不含 system role
	for _, msg := range anthReq.Messages {
		assert.NotEqual(t, "system", msg.Role,
			"Messages 中不应包含 system role 消息")
	}

	// 只有 user 消息
	assert.Len(t, anthReq.Messages, 1,
		"过滤 system 后应只剩 1 条 user 消息")
}
