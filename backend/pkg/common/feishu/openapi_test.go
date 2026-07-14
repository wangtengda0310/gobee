package feishu

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewOpenAPIClient_创建客户端 验证基本创建和默认值
func TestNewOpenAPIClient_创建客户端(t *testing.T) {
	client := NewOpenAPIClient("test_id", "test_secret")
	assert.Equal(t, "test_id", client.appID)
	assert.Equal(t, "test_secret", client.appSecret)
	assert.Equal(t, "https://open.feishu.cn", client.baseURL)
}

// TestOpenAPIClient_GetToken 验证 token 获取和缓存
func TestOpenAPIClient_GetToken(t *testing.T) {
	tokenCalled := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalled++
		// 验证请求方法和路径
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "tenant_access_token")

		// 返回成功 token
		resp := map[string]any{
			"code":                0,
			"msg":                 "ok",
			"tenant_access_token": "test_token_" + string(rune('0'+tokenCalled)),
			"expire":              7200,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAPIClient("test_id", "test_secret")
	client.baseURL = server.URL

	err := client.ensureToken()
	assert.NoError(t, err)
	assert.Equal(t, "test_token_1", client.token)
	assert.Equal(t, 1, tokenCalled)

	// 第二次调用应该使用缓存
	err = client.ensureToken()
	assert.NoError(t, err)
	assert.Equal(t, 1, tokenCalled) // 不应该再次调用
}

// TestOpenAPIClient_GetToken_失败响应 验证飞书返回错误时的处理
func TestOpenAPIClient_GetToken_失败响应(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"code": 99991401,
			"msg":  "permission denied",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAPIClient("bad_id", "bad_secret")
	client.baseURL = server.URL

	err := client.ensureToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "99991401")
}

// TestOpenAPIClient_SendText 验证消息发送请求格式
func TestOpenAPIClient_SendText(t *testing.T) {
	var receivedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal" {
			resp := map[string]any{
				"code":                0,
				"msg":                 "ok",
				"tenant_access_token": "test_token",
				"expire":              7200,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		// 验证发送消息请求
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
		assert.Contains(t, r.URL.RawQuery, "receive_id_type=email")

		json.NewDecoder(r.Body).Decode(&receivedBody)

		resp := map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"message_id": "om_test",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAPIClient("test_id", "test_secret")
	client.baseURL = server.URL

	err := client.SendText("user@ztgame.com", "email", "测试消息")
	assert.NoError(t, err)
	assert.Equal(t, "user@ztgame.com", receivedBody["receive_id"])
	assert.Equal(t, "text", receivedBody["msg_type"])
	assert.Contains(t, receivedBody["content"], "测试消息")
}

// TestOpenAPIClient_SendText_发送失败 验证发送失败时的错误返回
func TestOpenAPIClient_SendText_发送失败(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/open-apis/auth/v3/tenant_access_token/internal" {
			resp := map[string]any{
				"code":                0,
				"msg":                 "ok",
				"tenant_access_token": "test_token",
				"expire":              7200,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"code": 99991400,
			"msg":  "receive_id invalid",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAPIClient("test_id", "test_secret")
	client.baseURL = server.URL

	err := client.SendText("invalid", "email", "测试")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "99991400")
}
