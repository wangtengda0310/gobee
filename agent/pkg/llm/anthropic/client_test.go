package anthropic

import (
	"testing"
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
