// Package handlers 提供检查结果的输出处理器实现
package handlers

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// 最大消息长度（超出截断，飞书文本上限 150KB）
const maxDMContentLength = 4000

// FeishuDMHandler 飞书私聊消息处理器
// 向校验失败的 git 提交者发送私聊通知
type FeishuDMHandler struct {
	client    feishu.MessageSender
	formatter *notification.ErrorFormatter
	dryRun    bool // -noDM 时为 true：Handle 正常执行（触发装饰器 debug），但 sendDM 跳过实际发送
}

// NewFeishuDMHandler 创建私聊消息处理器
// client: 飞书 OpenAPI 客户端（实现 MessageSender 接口）
func NewFeishuDMHandler(client feishu.MessageSender) *FeishuDMHandler {
	return &FeishuDMHandler{
		client:    client,
		formatter: notification.NewErrorFormatter(false), // 私聊用纯文本模式
	}
}

// SetDryRun 设置 dryRun 模式（-noDM 时调用）
// Handle 正常执行（允许装饰器触发 debug 监控），但 sendDM 跳过实际发送
func (h *FeishuDMHandler) SetDryRun(dryRun bool) {
	h.dryRun = dryRun
}

// Handle 处理检查结果事件
// 核心逻辑：按作者聚合失败结果，向每个有失败的作者发送一条汇总消息
// 始终返回 nil，发送失败不影响其他 Handler
func (h *FeishuDMHandler) Handle(event *notification.CheckResultEvent) error {
	// 全量检查跳过（由 main.go 设置 CommitInfo.SkipDM=true）
	if event.CommitInfo != nil && event.CommitInfo.SkipDM {
		return nil
	}

	summary := notification.GetSummary(event)
	// 无错误不发送
	if !summary.HasErrors {
		return nil
	}

	if len(event.CommitSections) > 0 {
		h.handleMerge(event)
	} else {
		h.handleNormal(event)
	}

	return nil
}

// Name 返回处理器名称
func (h *FeishuDMHandler) Name() string {
	return "FeishuDM"
}

// handleNormal 处理普通 commit 的私聊通知
func (h *FeishuDMHandler) handleNormal(event *notification.CheckResultEvent) {
	if event.CommitInfo == nil || event.CommitInfo.Email == "" {
		log.Printf("[FeishuDM] 跳过：无邮箱信息")
		return
	}

	content := h.buildNormalContent(event)
	h.sendDM(event.CommitInfo.Email, content)
}

// handleMerge 处理 merge 场景的私聊通知
// 按 AuthorEmail 分组，每组聚合该作者的失败结果
func (h *FeishuDMHandler) handleMerge(event *notification.CheckResultEvent) {
	// 按作者分组，只包含有错误的 section
	type authorGroup struct {
		email    string
		sections []notification.CommitSection
	}
	groups := make(map[string]*authorGroup)

	for _, section := range event.CommitSections {
		// 检查该 section 是否有错误
		if !h.hasErrors(section.ColResults, section.TableResults) {
			continue
		}
		email := section.AuthorEmail
		if email == "" {
			continue
		}
		if _, exists := groups[email]; !exists {
			groups[email] = &authorGroup{email: email}
		}
		groups[email].sections = append(groups[email].sections, section)
	}

	// 向每个作者发送汇总消息
	for _, group := range groups {
		content := h.buildMergeContent(event, group.sections)
		h.sendDM(group.email, content)
	}
}

// hasErrors 检查是否有错误结果
func (h *FeishuDMHandler) hasErrors(colResults []*json_rule.ColCheckResult, tableResults []*json_rule.TableCheckResult) bool {
	for _, r := range colResults {
		if r != nil && !r.Ok {
			return true
		}
	}
	for _, r := range tableResults {
		if r != nil && !r.Ok {
			return true
		}
	}
	return false
}

// buildNormalContent 构建普通 commit 的私聊内容
func (h *FeishuDMHandler) buildNormalContent(event *notification.CheckResultEvent) string {
	var sb strings.Builder

	sb.WriteString("配表规则检查 - 发现问题\n\n")
	if event.CommitInfo != nil {
		shortHash := ""
		if len(event.CommitInfo.Version) >= 8 {
			shortHash = event.CommitInfo.Version[:8]
		}
		fmt.Fprintf(&sb, "提交: %s (%s)\n", shortHash, event.CommitInfo.Name)
		fmt.Fprintf(&sb, "分支: %s\n", event.CommitInfo.Branch)
	}
	sb.WriteString("\n")

	// 错误详情
	lines := h.formatter.FormatCheckResultLines(event)
	content := sb.String() + strings.Join(lines, "\n")

	return truncateContent(content)
}

// buildMergeContent 构建 merge 场景某作者的私聊内容
func (h *FeishuDMHandler) buildMergeContent(event *notification.CheckResultEvent, sections []notification.CommitSection) string {
	var sb strings.Builder

	sb.WriteString("配表规则检查 - Merge 发现问题\n\n")
	if event.CommitInfo != nil && event.CommitInfo.MergeInfo != nil {
		fmt.Fprintf(&sb, "Merge 操作: %s\n", event.CommitInfo.MergeInfo.MergeAuthor)
		fmt.Fprintf(&sb, "目标分支: %s\n\n", event.CommitInfo.Branch)
	}

	for _, section := range sections {
		shortHash := ""
		if len(section.CommitHash) >= 8 {
			shortHash = section.CommitHash[:8]
		}
		fmt.Fprintf(&sb, "---\n提交: %s (%s)\n", shortHash, section.Author)

		// 构建该 section 的错误内容
		sectionEvent := &notification.CheckResultEvent{
			ColResults:   section.ColResults,
			TableResults: section.TableResults,
		}
		lines := h.formatter.FormatCheckResultLines(sectionEvent)
		sb.WriteString(strings.Join(lines, "\n"))
		sb.WriteString("\n")
	}

	totalErrors := 0
	for _, s := range sections {
		for _, r := range s.ColResults {
			if r != nil && !r.Ok {
				totalErrors++
			}
		}
		for _, r := range s.TableResults {
			if r != nil && !r.Ok {
				totalErrors++
			}
		}
	}
	fmt.Fprintf(&sb, "\n共 %d 个问题，详情请查看群通知。", totalErrors)

	return truncateContent(sb.String())
}

// sendDM 发送私聊消息（失败只 log）
// dryRun 模式下跳过实际发送，仅记录日志
func (h *FeishuDMHandler) sendDM(email, content string) {
	// 简单邮箱格式校验
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		log.Printf("[FeishuDM] 跳过：邮箱格式无效 (%s)", email)
		return
	}
	if h.dryRun {
		log.Printf("[FeishuDM] dryRun 模式，跳过发送 (%s, 内容 %d 字符)", email, len(content))
		return
	}
	if err := h.client.SendText(email, "email", content); err != nil {
		log.Printf("[FeishuDM] 发送失败 (%s): %v", email, err)
	}
}

// truncateContent 截断过长消息，在完整 UTF-8 字符边界处截断
func truncateContent(content string) string {
	if len(content) <= maxDMContentLength {
		return content
	}
	// 安全截断到最近的完整 UTF-8 字符边界
	truncated := content[:maxDMContentLength]
	for !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "\n\n[消息过长已截断，详情请查看群通知]"
}
