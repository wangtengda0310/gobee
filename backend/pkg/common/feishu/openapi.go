package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MessageSender 飞书消息发送接口（用于测试 mock）
type MessageSender interface {
	SendText(receiveID, idType, text string) error
}

// OpenAPIClient 飞书 OpenAPI 客户端（用于私聊消息发送）
// 使用 tenant_access_token 认证，支持按邮箱/open_id/chat_id 发送消息
type OpenAPIClient struct {
	appID      string
	appSecret  string
	baseURL    string // 可注入，默认 https://open.feishu.cn
	token      string
	tokenExpAt int64 // token 过期时间（Unix 时间戳）
}

// NewOpenAPIClient 创建 OpenAPI 客户端
func NewOpenAPIClient(appID, appSecret string) *OpenAPIClient {
	return &OpenAPIClient{
		appID:     appID,
		appSecret: appSecret,
		baseURL:   "https://open.feishu.cn",
	}
}

// SendText 向指定用户发送文本消息
// receiveID: 接收者 ID（邮箱 / open_id / chat_id 等）
// idType: ID 类型（"email" / "open_id" / "chat_id" 等）
// text: 消息内容
func (c *OpenAPIClient) SendText(receiveID, idType, text string) error {
	if err := c.ensureToken(); err != nil {
		return fmt.Errorf("获取 token 失败: %w", err)
	}

	content, _ := json.Marshal(map[string]string{"text": text})
	return c.send(receiveID, idType, "text", string(content))
}

// ensureToken 获取或刷新 tenant_access_token
// token 有效期约 2 小时，短任务无需刷新
func (c *OpenAPIClient) ensureToken() error {
	if c.token != "" && time.Now().Unix() < c.tokenExpAt {
		return nil
	}

	body, _ := json.Marshal(map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	})

	resp, err := http.Post(
		c.baseURL+"/open-apis/auth/v3/tenant_access_token/internal",
		"application/json; charset=utf-8",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("获取 token 失败: code=%d, msg=%s", result.Code, result.Msg)
	}

	c.token = result.TenantAccessToken
	// 提前 5 分钟过期，避免边界情况
	c.tokenExpAt = time.Now().Unix() + int64(result.Expire) - 300
	return nil
}

// send 发送消息到飞书
func (c *OpenAPIClient) send(receiveID, idType, msgType, content string) error {
	body, _ := json.Marshal(map[string]string{
		"receive_id": receiveID,
		"msg_type":   msgType,
		"content":    content,
	})

	req, err := http.NewRequest("POST",
		c.baseURL+"/open-apis/im/v1/messages?receive_id_type="+idType,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析响应失败: %s", string(respBody))
	}
	if result.Code != 0 {
		return fmt.Errorf("发送失败: code=%d, msg=%s", result.Code, result.Msg)
	}

	return nil
}
