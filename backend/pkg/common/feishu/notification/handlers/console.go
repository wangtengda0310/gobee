// Package handlers 提供检查结果的输出处理器实现
package handlers

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
)

// ConsoleHandler 控制台输出处理器
// 将检查结果以可读格式输出到控制台
type ConsoleHandler struct {
	formatter *notification.ErrorFormatter
}

// NewConsoleHandler 创建控制台处理器
func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{
		formatter: notification.NewErrorFormatter(false), // 控制台不用颜色标签
	}
}

// Handle 处理检查结果事件
func (h *ConsoleHandler) Handle(event *notification.CheckResultEvent) error {
	// Merge 场景：走独立格式化路径
	if len(event.CommitSections) > 0 {
		output := h.formatter.FormatMergeConsoleOutput(event)
		fmt.Print(output)
		return nil
	}

	// 使用格式化器输出（非 merge 场景，格式不变）
	output := h.formatter.FormatConsoleOutput(event)
	fmt.Print(output)
	return nil
}

// Name 返回处理器名称
func (h *ConsoleHandler) Name() string {
	return "Console"
}
