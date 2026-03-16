package llm

import (
	"encoding/base64"
	"fmt"
)

// Role 消息角色
type Role string

const (
	RoleSystem    Role = "system"    // 系统消息
	RoleUser      Role = "user"      // 用户消息
	RoleAssistant Role = "assistant" // 助手消息
	RoleTool      Role = "tool"      // 工具响应消息
)

// Message 表示一条对话消息
type Message struct {
	// Role 消息角色
	Role Role `json:"role"`

	// Content 消息内容 (支持多模态)
	Content Content `json:"content"`

	// ToolCallID 工具调用 ID (用于工具响应消息)
	ToolCallID string `json:"tool_call_id,omitempty"`

	// Name 工具名称 (用于工具响应消息)
	Name string `json:"name,omitempty"`

	// ToolCalls 助手发起的工具调用
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}

// Content 接口表示消息内容 (支持文本和多模态)
type Content interface {
	// ContentType 返回内容类型
	ContentType() string

	// ToOpenAI 转换为 OpenAI 格式
	ToOpenAI() interface{}

	// ToAnthropic 转换为 Anthropic 格式
	ToAnthropic() interface{}
}

// TextContent 纯文本内容
type TextContent struct {
	Text string `json:"text"`
}

func (t *TextContent) ContentType() string {
	return "text"
}

func (t *TextContent) ToOpenAI() interface{} {
	return t.Text
}

func (t *TextContent) ToAnthropic() interface{} {
	return map[string]interface{}{
		"type": "text",
		"text": t.Text,
	}
}

// ImageContent 图像内容
type ImageContent struct {
	// Source 图像源
	Source ImageSource `json:"source"`
}

// ImageSource 图像源
type ImageSource struct {
	// Type 源类型: "base64" 或 "url"
	Type string `json:"type"`

	// MediaType 媒体类型 (如 "image/png", "image/jpeg")
	MediaType string `json:"media_type,omitempty"`

	// Data base64 编码数据
	Data string `json:"data,omitempty"`

	// URL 图像 URL
	URL string `json:"url,omitempty"`
}

func (i *ImageContent) ContentType() string {
	return "image"
}

func (i *ImageContent) ToOpenAI() interface{} {
	if i.Source.Type == "url" {
		return map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": i.Source.URL,
			},
		}
	}
	// OpenAI 使用带 data URL 的格式
	dataURL := fmt.Sprintf("data:%s;base64,%s", i.Source.MediaType, i.Source.Data)
	return map[string]interface{}{
		"type": "image_url",
		"image_url": map[string]string{
			"url": dataURL,
		},
	}
}

func (i *ImageContent) ToAnthropic() interface{} {
	return map[string]interface{}{
		"type": "image",
		"source": map[string]interface{}{
			"type":       "base64",
			"media_type": i.Source.MediaType,
			"data":       i.Source.Data,
		},
	}
}

// ContentList 多内容块列表
type ContentList struct {
	Items []Content `json:"-"`
}

func (c *ContentList) ContentType() string {
	return "list"
}

func (c *ContentList) ToOpenAI() interface{} {
	// 单个纯文本直接返回字符串
	if len(c.Items) == 1 {
		if text, ok := c.Items[0].(*TextContent); ok {
			return text.Text
		}
	}

	// 多个内容返回数组
	parts := make([]interface{}, len(c.Items))
	for i, item := range c.Items {
		parts[i] = item.ToOpenAI()
	}
	return parts
}

func (c *ContentList) ToAnthropic() interface{} {
	parts := make([]interface{}, len(c.Items))
	for i, item := range c.Items {
		parts[i] = item.ToAnthropic()
	}
	return parts
}

// 辅助函数

// Text 快速创建文本内容
func Text(text string) Content {
	return &TextContent{Text: text}
}

// TextString 从内容中提取纯文本
func TextString(c Content) string {
	switch v := c.(type) {
	case *TextContent:
		return v.Text
	case *ContentList:
		var result string
		for _, item := range v.Items {
			result += TextString(item)
		}
		return result
	}
	return ""
}

// ImageFromBase64 从 base64 数据创建图像内容
func ImageFromBase64(data, mediaType string) Content {
	return &ImageContent{
		Source: ImageSource{
			Type:      "base64",
			MediaType: mediaType,
			Data:      data,
		},
	}
}

// ImageFromURL 从 URL 创建图像内容
func ImageFromURL(url string) Content {
	return &ImageContent{
		Source: ImageSource{
			Type: "url",
			URL:  url,
		},
	}
}

// ImageFromBase64Bytes 从 base64 字节创建图像内容
func ImageFromBase64Bytes(data []byte, mediaType string) Content {
	return ImageFromBase64(base64.StdEncoding.EncodeToString(data), mediaType)
}

// NewContentList 创建内容列表
func NewContentList(items ...Content) *ContentList {
	return &ContentList{Items: items}
}

// Add 添加内容项
func (c *ContentList) Add(item Content) {
	c.Items = append(c.Items, item)
}

// AddText 添加文本内容
func (c *ContentList) AddText(text string) {
	c.Items = append(c.Items, Text(text))
}

// AddImage 添加图像内容
func (c *ContentList) AddImage(content *ImageContent) {
	c.Items = append(c.Items, content)
}
