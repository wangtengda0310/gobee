package handlers

import (
	"strings"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

// TestFormatDMRecipients_全量检查跳过
func TestFormatDMRecipients_全量检查跳过(t *testing.T) {
	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{{Ok: false, Reason: "错误"}},
		CommitInfo: &notification.CommitInfo{
			Name:   "excel配置全量检查",
			SkipDM: true,
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	assert.Contains(t, got, "全量检查跳过")
}

// TestFormatDMRecipients_无错误未发送
func TestFormatDMRecipients_无错误未发送(t *testing.T) {
	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{{Ok: true}},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "zhangsan@ztgame.com",
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	assert.Contains(t, got, "无错误")
}

// TestFormatDMRecipients_普通commit单人
func TestFormatDMRecipients_普通commit单人(t *testing.T) {
	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{{Ok: false, Reason: "错误"}},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "zhangsan@ztgame.com",
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	assert.Equal(t, "zhangsan@ztgame.com", got)
}

// TestFormatDMRecipients_普通commit缺邮箱
func TestFormatDMRecipients_普通commit缺邮箱(t *testing.T) {
	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{{Ok: false, Reason: "错误"}},
		CommitInfo: &notification.CommitInfo{
			Name:  "张三",
			Email: "",
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	assert.Contains(t, got, "缺少邮箱")
}

// TestFormatDMRecipients_Merge场景多作者去重
func TestFormatDMRecipients_Merge场景多作者去重(t *testing.T) {
	event := &notification.CheckResultEvent{
		CommitInfo: &notification.CommitInfo{
			MergeInfo: &notification.MergeInfo{MergeAuthor: "merge作者"},
		},
		CommitSections: []notification.CommitSection{
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				ColResults:  []*json_rule.ColCheckResult{{Ok: false, Reason: "错误1"}},
			},
			{
				Author:      "李四",
				AuthorEmail: "lisi@ztgame.com",
				ColResults:  []*json_rule.ColCheckResult{{Ok: false, Reason: "错误2"}},
			},
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com", // 同作者重复，应去重
				ColResults:  []*json_rule.ColCheckResult{{Ok: false, Reason: "错误3"}},
			},
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	// 两个邮箱都在
	assert.Contains(t, got, "zhangsan@ztgame.com")
	assert.Contains(t, got, "lisi@ztgame.com")
	// 张三只出现一次（去重）
	assert.Equal(t, 1, strings.Count(got, "zhangsan@ztgame.com"))
}

// TestFormatDMRecipients_Merge场景部分作者无错误
func TestFormatDMRecipients_Merge场景部分作者无错误(t *testing.T) {
	event := &notification.CheckResultEvent{
		CommitInfo: &notification.CommitInfo{
			MergeInfo: &notification.MergeInfo{MergeAuthor: "merge作者"},
		},
		CommitSections: []notification.CommitSection{
			{
				Author:      "张三",
				AuthorEmail: "zhangsan@ztgame.com",
				ColResults:  []*json_rule.ColCheckResult{{Ok: false, Reason: "有错误"}},
			},
			{
				Author:      "李四",
				AuthorEmail: "lisi@ztgame.com",
				ColResults:  []*json_rule.ColCheckResult{{Ok: true}}, // 无错误
			},
		},
	}
	summary := notification.GetSummary(event)

	got := formatDMRecipients(event, summary)
	assert.Contains(t, got, "zhangsan@ztgame.com")
	assert.NotContains(t, got, "lisi@ztgame.com")
}

// TestBuildDebugMessage_包含私聊接收人字段
// 验证 debug 消息中确实出现了「私聊接收人」这一行
func TestBuildDebugMessage_包含私聊接收人字段(t *testing.T) {
	event := &notification.CheckResultEvent{
		ColResults: []*json_rule.ColCheckResult{{Ok: false, Reason: "错误"}},
		CommitInfo: &notification.CommitInfo{
			Name:    "张三",
			Email:   "zhangsan@ztgame.com",
			Branch:  "feature/test",
			Version: "abc1234567",
		},
	}
	summary := notification.GetSummary(event)

	msg := buildDebugMessage("FeishuDM", event, summary, nil)
	assert.Contains(t, msg, "私聊接收人: zhangsan@ztgame.com")
}
