package handlers

import (
	"fmt"
	"log"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// debugMonitorEmail debug 监控消息接收邮箱
const debugMonitorEmail = "v-wangtengda@ztgame.com"

// maxDebugFailedTables debug 消息中列出的失败表级规则上限，避免消息过长
const maxDebugFailedTables = 8

// DMMonitorDecorator 装饰器：包装 CheckResultHandler
// 每次被装饰的 handler.Handle 执行后，自动发送 debug 摘要到监控邮箱
// 通知失败不阻塞流水线（仅 log）
type DMMonitorDecorator struct {
	inner  notification.CheckResultHandler
	sender feishu.MessageSender
}

// WrapDMMonitor 创建 debug 监控装饰器
// inner 或 sender 为 nil 时返回原始 handler（不装饰）
func WrapDMMonitor(inner notification.CheckResultHandler, sender feishu.MessageSender) notification.CheckResultHandler {
	if inner == nil || sender == nil {
		return inner
	}
	return &DMMonitorDecorator{inner: inner, sender: sender}
}

// Handle 执行被装饰的 handler，然后发送 debug 摘要
func (d *DMMonitorDecorator) Handle(event *notification.CheckResultEvent) error {
	err := d.inner.Handle(event)

	summary := notification.GetSummary(event)
	msg := buildDebugMessage(d.inner.Name(), event, summary, err)

	if sendErr := d.sender.SendText(debugMonitorEmail, "email", msg); sendErr != nil {
		log.Printf("[DMMonitorDecorator] debug 发送失败: %v", sendErr)
	}

	return err
}

// buildDebugMessage 构造 debug 监控消息
// 相比旧的「通知已分发 | handler | 错误 | 表级」单行摘要，这里输出多行上下文：
// 提交人/分支/commit、私聊接收人、增量/变更文件、列级/表级/解析错误/通知计数、失败表名、handler 报错、检查时间。
// 各字段为空时对应行省略，避免噪音。
func buildDebugMessage(innerName string, event *notification.CheckResultEvent, summary *notification.CheckResultSummary, handlerErr error) string {
	var b strings.Builder
	b.WriteString("[DEBUG-MONITOR] 通知已分发\n")
	b.WriteString("handler: " + innerName)

	// 提交信息
	if event.CommitInfo != nil {
		var ci = event.CommitInfo
		// Merge 场景优先显示 merge 摘要，普通场景显示单提交
		if ci.MergeInfo != nil && ci.MergeInfo.MergeAuthor != "" {
			fmt.Fprintf(&b, "\nMerge: %s → %s", ci.MergeInfo.MergeAuthor, ci.MergeInfo.ToBranch)
			if ci.MergeInfo.Parent1Total+ci.MergeInfo.Parent2Total > 0 {
				fmt.Fprintf(&b, " (主 %d + 被 %d)", ci.MergeInfo.Parent1Total, ci.MergeInfo.Parent2Total)
			}
		} else if ci.Name != "" {
			version := ci.Version
			if len(version) > 8 {
				version = version[:8]
			}
			fmt.Fprintf(&b, "\n提交: %s @ %s", ci.Name, version)
		}
		if ci.Branch != "" {
			b.WriteString("\n分支: " + ci.Branch)
		}
	}

	// 私聊接收人（与 FeishuDMHandler 的发送逻辑保持一致）
	b.WriteString("\n私聊接收人: " + formatDMRecipients(event, summary))

	// 增量/全量 + 变更文件
	mode := "全量"
	if event.IsIncremental {
		mode = "增量"
	}
	filePart := ""
	if len(event.ChangedFiles) > 0 {
		filePart = " | 变更文件: " + strings.Join(event.ChangedFiles, ", ")
	}
	fmt.Fprintf(&b, "\n%s检查%s", mode, filePart)

	// 统计：列级 / 表级 / 解析错误 / 总计
	fmt.Fprintf(&b, "\n统计: 列级 %d(失败%d) | 表级 %d(失败%d, 通知%d) | 解析错误 %d | 总计 %d",
		len(event.ColResults), summary.ColErrors,
		len(event.TableResults), summary.TableErrors, summary.TableNotifications,
		summary.ParseErrors, summary.TotalErrors)

	// 失败的表级规则（去重，最多 maxDebugFailedTables 个）
	if failed := formatFailedTableRules(event.TableResults); failed != "" {
		b.WriteString("\n失败表: " + failed)
	}

	// handler 自身报错
	errPart := "(无)"
	if handlerErr != nil {
		errPart = handlerErr.Error()
	}
	b.WriteString("\nhandler 报错: " + errPart)

	// 检查时间
	if !event.CheckTime.IsZero() {
		b.WriteString("\n检查时间: " + event.CheckTime.Format("2006-01-02 15:04:05"))
	}

	return b.String()
}

// formatDMRecipients 推断私聊消息的接收人
// 逻辑与 FeishuDMHandler.Handle 保持一致：
//   - 全量检查（SkipDM=true）或无错误 → 不发送
//   - Merge 场景 → 按有错误的 section 聚合作者邮箱（去重）
//   - 普通场景 → CommitInfo.Email
//
// 用于 debug 监控消息，让排障时能直观看到「通知发给了谁」。
func formatDMRecipients(event *notification.CheckResultEvent, summary *notification.CheckResultSummary) string {
	// 全量检查跳过私聊
	if event.CommitInfo != nil && event.CommitInfo.SkipDM {
		return "无（全量检查跳过私聊）"
	}

	// Merge 场景：按有错误的 section 聚合作者邮箱（去重）
	// 注意：merge 模式下顶层 ColResults/TableResults 是聚合结果，但这里以 section 为准，
	// 避免依赖调用方是否填充了聚合结果（与 FeishuDMHandler.handleMerge 一致）。
	if len(event.CommitSections) > 0 {
		seen := make(map[string]bool)
		var emails []string
		for _, section := range event.CommitSections {
			if section.AuthorEmail == "" {
				continue
			}
			// 只有该 section 存在失败结果才计入（与 handleMerge 一致）
			if !sectionHasErrors(section.ColResults, section.TableResults) {
				continue
			}
			if seen[section.AuthorEmail] {
				continue
			}
			seen[section.AuthorEmail] = true
			emails = append(emails, section.AuthorEmail)
		}
		if len(emails) == 0 {
			return "无（无有错误的提交者）"
		}
		return strings.Join(emails, ", ")
	}

	// 普通场景：无错误不发送
	if summary == nil || !summary.HasErrors {
		return "无（无错误）"
	}

	// 普通场景：CommitInfo.Email
	if event.CommitInfo != nil && event.CommitInfo.Email != "" {
		return event.CommitInfo.Email
	}
	return "无（缺少邮箱信息）"
}

// sectionHasErrors 检查某个 CommitSection 是否存在失败结果
// 与 FeishuDMHandler.hasErrors 逻辑一致，供 formatDMRecipients 复用
func sectionHasErrors(colResults []*json_rule.ColCheckResult, tableResults []*json_rule.TableCheckResult) bool {
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

// formatFailedTableRules 提取失败的表级规则摘要
// 格式: RuleType[SheetName] 去重后逗号分隔，超过 maxDebugFailedTables 个则附「等 N 个」
func formatFailedTableRules(results []*json_rule.TableCheckResult) string {
	seen := make(map[string]bool)
	var items []string
	for _, r := range results {
		if r == nil || r.Ok {
			continue
		}
		rule := string(r.RuleType)
		if rule == "" {
			rule = r.DisplayName
		}
		sheet := derefStr(r.SheetName)
		key := rule + "|" + sheet
		if seen[key] {
			continue
		}
		seen[key] = true
		if sheet != "" {
			items = append(items, rule+"["+sheet+"]")
		} else {
			items = append(items, rule)
		}
	}

	total := len(items)
	if total == 0 {
		return ""
	}
	if total > maxDebugFailedTables {
		items = items[:maxDebugFailedTables]
		return strings.Join(items, ", ") + fmt.Sprintf(" 等 %d 个", total)
	}
	return strings.Join(items, ", ")
}

// derefStr 安全解引用字符串指针
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Name 返回处理器名称
func (d *DMMonitorDecorator) Name() string {
	return fmt.Sprintf("DMMonitor(%s)", d.inner.Name())
}
