package handlers

import (
	"fmt"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

// mockMessageSender 用于测试的 mock 消息发送器
type mockMessageSender struct {
	calls []struct {
		receiveID string
		idType    string
		text      string
	}
	shouldError bool
}

func (m *mockMessageSender) SendText(receiveID, idType, text string) error {
	m.calls = append(m.calls, struct {
		receiveID string
		idType    string
		text      string
	}{receiveID, idType, text})
	if m.shouldError {
		return fmt.Errorf("发送失败")
	}
	return nil
}

// TestFeishuDMHandler_普通commit有错误
func TestFeishuDMHandler_普通commit有错误(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "值超出范围"},
		},
		CommitInfo: &notification.CommitInfo{
			Name:    "张三",
			Email:   "zhangsan@ztgame.com",
			Branch:  "feature/test",
			Version: "abc1234567",
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 1)
	assert.Equal(t, "zhangsan@ztgame.com", mock.calls[0].receiveID)
	assert.Equal(t, "email", mock.calls[0].idType)
	assert.Contains(t, mock.calls[0].text, "发现问题")
}

// TestFeishuDMHandler_普通commit无错误
func TestFeishuDMHandler_普通commit无错误(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: true},
		},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "zhangsan@ztgame.com",
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 0)
}

// TestFeishuDMHandler_全量检查跳过
func TestFeishuDMHandler_全量检查跳过(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误"},
		},
		CommitInfo: &notification.CommitInfo{
			Name:   "excel配置全量检查",
			SkipDM: true,
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 0)
}

// TestFeishuDMHandler_邮箱为空跳过
func TestFeishuDMHandler_邮箱为空跳过(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误"},
		},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "",
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 0)
}

// TestFeishuDMHandler_Merge场景按作者分组
func TestFeishuDMHandler_Merge场景按作者分组(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误1"},
			{Ok: false, Reason: "错误2"},
		},
		CommitInfo: &notification.CommitInfo{
			Name: "merge作者",
			MergeInfo: &notification.MergeInfo{
				MergeAuthor: "merge作者",
			},
		},
		CommitSections: []notification.CommitSection{
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				CommitHash:  "aaa1111111",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: false, Reason: "张三的错误"},
				},
			},
			{
				Author:      "李四",
				AuthorEmail: "lisi@ztgame.com",
				CommitHash:  "bbb2222222",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: false, Reason: "李四的错误"},
				},
			},
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 2)

	// 验证两个邮箱都收到了消息
	emails := []string{mock.calls[0].receiveID, mock.calls[1].receiveID}
	assert.Contains(t, emails, "zhangsan@ztgame.com")
	assert.Contains(t, emails, "lisi@ztgame.com")
}

// TestFeishuDMHandler_Merge场景同作者聚合
func TestFeishuDMHandler_Merge场景同作者聚合(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误1"},
			{Ok: false, Reason: "错误2"},
		},
		CommitInfo: &notification.CommitInfo{
			Name: "merge作者",
			MergeInfo: &notification.MergeInfo{
				MergeAuthor: "merge作者",
			},
		},
		CommitSections: []notification.CommitSection{
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				CommitHash:  "aaa1111111",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: false, Reason: "错误1"},
				},
			},
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				CommitHash:  "ccc3333333",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: false, Reason: "错误2"},
				},
			},
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 1) // 同作者只发一条
	assert.Equal(t, "zhangsan@ztgame.com", mock.calls[0].receiveID)
	assert.Contains(t, mock.calls[0].text, "错误1")
	assert.Contains(t, mock.calls[0].text, "错误2")
}

// TestFeishuDMHandler_Merge场景部分作者无错误
func TestFeishuDMHandler_Merge场景部分作者无错误(t *testing.T) {
	mock := &mockMessageSender{}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误"},
		},
		CommitInfo: &notification.CommitInfo{
			Name: "merge作者",
			MergeInfo: &notification.MergeInfo{
				MergeAuthor: "merge作者",
			},
		},
		CommitSections: []notification.CommitSection{
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				CommitHash:  "aaa1111111",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: false, Reason: "有错误"},
				},
			},
			{
				Author:      "李四",
				AuthorEmail: "lisi@ztgame.com",
				CommitHash:  "bbb2222222",
				ColResults: []*json_rule.ColCheckResult{
					{Ok: true}, // 无错误
				},
			},
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)
	assert.Len(t, mock.calls, 1) // 只有张三收到
	assert.Equal(t, "zhangsan@ztgame.com", mock.calls[0].receiveID)
}

// TestFeishuDMHandler_发送失败不阻断
func TestFeishuDMHandler_发送失败不阻断(t *testing.T) {
	mock := &mockMessageSender{shouldError: true}
	handler := NewFeishuDMHandler(mock)

	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{
			{Ok: false, Reason: "错误"},
		},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "zhangsan@ztgame.com",
		},
	}

	err := handler.Handle(event)
	assert.Nil(t, err)           // 即使发送失败也返回 nil
	assert.Len(t, mock.calls, 1) // 确实尝试了发送
}
