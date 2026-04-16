package chat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSettingsFile(t *testing.T, content map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(content)
	require.NoError(t, err)
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func TestLoadConfigFromFile(t *testing.T) {
	path := createTestSettingsFile(t, map[string]any{
		"env": map[string]any{
			"ANTHROPIC_AUTH_TOKEN":         "test-key-123",
			"ANTHROPIC_BASE_URL":           "https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL": "glm-4.5-air",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-4.7",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":  "glm-5.1",
		},
		"model": "opus[1m]",
	})

	cfg, err := loadConfigFromPath(path)
	require.NoError(t, err)

	assert.Equal(t, "test-key-123", cfg.APIKey)
	assert.Equal(t, "https://open.bigmodel.cn/api/anthropic", cfg.BaseURL)
	assert.Equal(t, "glm-4.5-air", cfg.Models["haiku"])
	assert.Equal(t, "glm-4.7", cfg.Models["sonnet"])
	assert.Equal(t, "glm-5.1", cfg.Models["opus"])
	assert.Equal(t, "glm-5.1", cfg.DefaultModel)
}

func TestParseModelTier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"opus[1m]", "opus"},
		{"sonnet", "sonnet"},
		{"haiku[2m]", "haiku"},
		{"opus", "opus"},
		{"unknown_model", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseModelTier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfigFallback(t *testing.T) {
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "env-key")
	os.Setenv("ANTHROPIC_BASE_URL", "https://env-url.com")
	defer os.Unsetenv("ANTHROPIC_AUTH_TOKEN")
	defer os.Unsetenv("ANTHROPIC_BASE_URL")

	cfg, err := LoadConfig("/nonexistent/path/settings.json")
	require.NoError(t, err)

	assert.Equal(t, "env-key", cfg.APIKey)
	assert.Equal(t, "https://env-url.com", cfg.BaseURL)
}

func TestLoadConfigDefaults(t *testing.T) {
	path := createTestSettingsFile(t, map[string]any{
		"env": map[string]any{
			"ANTHROPIC_AUTH_TOKEN": "test-key",
			"ANTHROPIC_BASE_URL":  "https://open.bigmodel.cn/api/anthropic",
		},
	})

	cfg, err := loadConfigFromPath(path)
	require.NoError(t, err)
	assert.Equal(t, "test-key", cfg.APIKey)
	assert.NotEmpty(t, cfg.DefaultModel)
}