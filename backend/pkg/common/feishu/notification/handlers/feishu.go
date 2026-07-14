// Package handlers 提供检查结果的输出处理器实现
package handlers

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
)

// FeishuCardHandler 飞书卡片消息处理器
// 将检查结果以卡片消息形式发送到飞书群
type FeishuCardHandler struct {
	robotGUID string   // 飞书机器人 GUID
	atUsers   []string // @ 用户 ID 列表
	formatter *notification.ErrorFormatter
	enabled   func() bool // 动态开关检查函数，nil 表示始终启用
}

// FeishuCardOption 飞书卡片配置选项
type FeishuCardOption func(*FeishuCardHandler)

// WithAtUser 设置 @ 用户
func WithAtUser(userID string) FeishuCardOption {
	return func(h *FeishuCardHandler) {
		h.atUsers = append(h.atUsers, userID)
	}
}

// WithAtUsers 设置多个 @ 用户
func WithAtUsers(userIDs []string) FeishuCardOption {
	return func(h *FeishuCardHandler) {
		h.atUsers = append(h.atUsers, userIDs...)
	}
}

// WithEnabledFunc 设置动态开关检查函数
// 每次执行 Handle 时调用此函数判断是否启用，nil 表示始终启用
func WithEnabledFunc(enabled func() bool) FeishuCardOption {
	return func(h *FeishuCardHandler) {
		h.enabled = enabled
	}
}

// NewFeishuCardHandler 创建飞书卡片处理器
// robotGUID: 飞书机器人的 GUID，为空或 "none" 时不发送消息
func NewFeishuCardHandler(robotGUID string, opts ...FeishuCardOption) *FeishuCardHandler {
	h := &FeishuCardHandler{
		robotGUID: robotGUID,
		formatter: notification.NewErrorFormatter(true), // 飞书卡片用颜色标签
		atUsers:   make([]string, 0),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Handle 处理检查结果事件
func (h *FeishuCardHandler) Handle(event *notification.CheckResultEvent) error {
	// 跳过无效的机器人配置
	if h.robotGUID == "" || h.robotGUID == "none" {
		return nil
	}

	// 动态开关检查（nil 表示始终启用）
	if h.enabled != nil && !h.enabled() {
		return nil
	}

	// Merge 场景：走独立格式化路径（汇聚为一条消息）
	if len(event.CommitSections) > 0 {
		return h.handleMergeEvent(event)
	}

	summary := notification.GetSummary(event)

	// 构建内容
	var contentLines []string

	// 添加 @ 用户
	for _, userID := range h.atUsers {
		contentLines = append(contentLines, fmt.Sprintf("<at id=\"%s\"></at>", userID))
	}

	// 格式化检查结果（上下文 + 解析错误 + 列级错误 + 表级错误 + 表级通知）
	contentLines = append(contentLines, h.formatter.FormatCheckResultLines(event)...)

	content := strings.Join(contentLines, "\n")

	// 根据检查结果选择卡片样式
	// 有错误或有通知都发送飞书消息
	if summary.HasErrors || summary.HasNotifications {
		title := "配表规则检查结果"
		var subTitle string
		if summary.HasErrors {
			subTitle = fmt.Sprintf("检查完成，发现 %d 个问题", summary.TotalErrors)
		} else {
			subTitle = fmt.Sprintf("检查完成，有 %d 条变更通知", summary.TableNotifications)
		}
		feishu.WarningRed(h.robotGUID, title, subTitle, content)
	} else {
		title := "配表规则检查通过"
		subTitle := "✅ 所有检查通过！"
		feishu.SuccessGreen(h.robotGUID, title, subTitle, content)
	}

	return nil
}

// Name 返回处理器名称
func (h *FeishuCardHandler) Name() string {
	return "FeishuCard"
}

// handleMergeEvent 处理 merge 场景的检查结果事件（汇聚为一条消息）
// 使用 FormatMergeContent 生成包含 merge 概览 + 各 commit 分段的内容
func (h *FeishuCardHandler) handleMergeEvent(event *notification.CheckResultEvent) error {
	summary := notification.GetSummary(event)
	content := h.formatter.FormatMergeContent(event)

	if summary.HasErrors || summary.HasNotifications {
		title := "配表规则检查结果"
		var subTitle string
		if summary.HasErrors {
			subTitle = fmt.Sprintf("检查完成，发现 %d 个问题", summary.TotalErrors)
		} else {
			subTitle = fmt.Sprintf("检查完成，有 %d 条变更通知", summary.TableNotifications)
		}
		feishu.WarningRed(h.robotGUID, title, subTitle, content)
	} else {
		feishu.SuccessGreen(h.robotGUID, "配表规则检查通过", "✅ 所有检查通过！", content)
	}

	return nil
}
