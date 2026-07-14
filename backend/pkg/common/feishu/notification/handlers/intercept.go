// Package handlers 提供检查结果的输出处理器实现
package handlers

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
)

// InterceptHandler 劫持处理器
// 在测试阶段劫持飞书消息发送，在本地弹窗显示而非真正发送到飞书服务器
// 优先级最高，应该第一个注册，先于飞书卡片处理器执行
type InterceptHandler struct {
	enabled          func() bool // 劫持开关检查函数（延迟绑定，避免循环依赖）
	interceptService notification.InterceptService
	eventsEmit       notification.EventsEmit
}

// NewInterceptHandler 创建劫持处理器
// enabled: 劫持开关检查函数，返回 true 表示劫持开启
// interceptService: 劫持服务，用于添加劫持的消息
// eventsEmit: 事件发送函数，用于将劫持的消息推送到前端
func NewInterceptHandler(enabled func() bool, interceptService notification.InterceptService, eventsEmit notification.EventsEmit) *InterceptHandler {
	return &InterceptHandler{
		enabled:          enabled,
		interceptService: interceptService,
		eventsEmit:       eventsEmit,
	}
}

// Handle 处理检查结果事件
// 如果劫持开启，拦截消息并发送到前端，然后返回 ErrStopProcessing 阻止后续处理器
// 如果劫持未开启，返回 nil 让下一个处理器处理
func (h *InterceptHandler) Handle(event *notification.CheckResultEvent) error {
	// 检查劫持开关
	if !h.enabled() {
		// 劫持未开启，返回 nil 让后续处理器（飞书卡片处理器）继续处理
		return nil
	}

	// Merge 场景：走独立格式化路径
	if len(event.CommitSections) > 0 {
		return h.handleMergeEvent(event)
	}

	summary := notification.GetSummary(event)

	// 只有有错误或通知时才劫持
	if !summary.HasErrors && !summary.HasNotifications {
		return nil
	}

	// 格式化检查结果（上下文 + 解析错误 + 列级错误 + 表级错误 + 表级通知）
	formatter := notification.NewErrorFormatter(true)
	contentLines := formatter.FormatCheckResultLines(event)

	content := fmt.Sprintf("**%s**\n\n%s", event.CheckTime.Format("2006-01-02 15:04:05"), strings.Join(contentLines, "\n"))

	// 确定消息类型
	msgType := "text"
	if summary.TableNotifications > 0 {
		msgType = "interactive"
	}

	// 添加劫持的消息（内部调用会触发 Wails 事件推送到前端）
	h.interceptService.AddMessage("excel-check", msgType, content)

	// 返回 ErrStopProcessing 阻止后续处理器（如飞书卡片处理器）执行
	return notification.ErrStopProcessing
}

// Name 返回处理器名称
func (h *InterceptHandler) Name() string {
	return "Intercept"
}

// handleMergeEvent 处理 merge 场景的劫持消息（汇聚为一条）
func (h *InterceptHandler) handleMergeEvent(event *notification.CheckResultEvent) error {
	summary := notification.GetSummary(event)

	// 只有有错误或通知时才劫持
	if !summary.HasErrors && !summary.HasNotifications {
		return nil
	}

	formatter := notification.NewErrorFormatter(true)
	content := formatter.FormatMergeContent(event)
	content = fmt.Sprintf("**%s**\n\n%s", event.CheckTime.Format("2006-01-02 15:04:05"), content)

	// 确定消息类型
	msgType := "text"
	if summary.TableNotifications > 0 {
		msgType = "interactive"
	}

	h.interceptService.AddMessage("excel-check", msgType, content)
	return notification.ErrStopProcessing
}
