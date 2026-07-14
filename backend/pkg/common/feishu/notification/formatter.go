package notification

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ErrorFormatter 统一的错误格式化器
// 支持带颜色和不带颜色两种输出模式
type ErrorFormatter struct {
	colorEnabled bool // 是否启用颜色（飞书卡片用 HTML 颜色标签）
}

// NewErrorFormatter 创建错误格式化器
// colorEnabled: true 时使用 <font color='red'> 标签，false 时输出纯文本
func NewErrorFormatter(colorEnabled bool) *ErrorFormatter {
	return &ErrorFormatter{colorEnabled: colorEnabled}
}

// errorLabel 返回错误标签（带颜色或不带颜色）
func (f *ErrorFormatter) errorLabel() string {
	if f.colorEnabled {
		return "**<font color='red'>错误</font>**"
	}
	return "错误"
}

// formatErrorLine 格式化整行错误描述（飞书卡片模式下整行加粗红色）
func (f *ErrorFormatter) formatErrorLine(format string, args ...interface{}) string {
	text := fmt.Sprintf(format, args...)
	if f.colorEnabled {
		return fmt.Sprintf("**<font color='red'>错误: %s</font>**", text)
	}
	return fmt.Sprintf("错误: %s", text)
}

// FormatColErrors 格式化列级检查错误
// 返回格式化的错误信息切片，每条错误一行
func (f *ErrorFormatter) FormatColErrors(results []*json_rule.ColCheckResult) []string {
	var lines []string
	for _, r := range results {
		if r == nil || r.Ok {
			continue
		}

		colName := "未知列"
		if r.ColName != nil {
			colName = *r.ColName
		}

		// 遍历错误单元格，显示每个错误的详细信息
		// 格式与前端面板一致：第N行, 错误原因:xxx
		for _, errCell := range r.ErrCells {
			// errCell.Index 是数组索引（从0开始），Excel 行号 = Index + 1
			excelRow := errCell.Index + 1
			if errCell.ExcelRow > 0 {
				excelRow = errCell.ExcelRow
			}
			lines = append(lines, f.formatErrorLine("%s [%s] 第%d行, %s",
				safeString(r.SheetName), colName, excelRow, errCell.Reason))
		}

		// 如果没有错误单元格但有整体原因，显示整体错误
		if len(r.ErrCells) == 0 && r.Reason != "" {
			lines = append(lines, f.formatErrorLine("%s [%s], %s",
				safeString(r.SheetName), colName, r.Reason))
		}
	}
	return lines
}

// FormatNotifications 格式化表级变更通知（用于飞书卡片和命令行日志）
//
// 只输出 Ok=true 且有 ErrCells 的通知规则结果
//
// Ok 字段语义：
//   - Ok=false: 错误检测规则（如赛季检查、武将开放时间检查）
//   - Ok=true:  通知规则（如 NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）
//
// 详细说明见：../rain-qa-func/docs/表级检查结果Ok字段语义.md
func (f *ErrorFormatter) FormatNotifications(results []*json_rule.TableCheckResult) []string {
	var lines []string
	for _, r := range results {
		if r == nil || !r.Ok {
			continue
		}
		// 只处理有 ErrCells 的通知规则（过滤"首次运行"和"无变更"）
		if len(r.ErrCells) == 0 {
			continue
		}

		// 新格式 Reason 已包含完整的结构化消息（📋 工作表变更通知），直接输出
		lines = append(lines, r.Reason)
	}
	return lines
}

// FormatTableErrors 格式化表级检查错误
//
// 只输出 Ok=false 的错误检测规则结果，跳过 Ok=true 的通知规则
//
// Ok 字段语义：
//   - Ok=false: 错误检测规则（如赛季检查、武将开放时间检查）
//   - Ok=true:  通知规则（如 NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）
//
// 详细说明见：../rain-qa-func/docs/表级检查结果Ok字段语义.md
func (f *ErrorFormatter) FormatTableErrors(results []*json_rule.TableCheckResult) []string {
	var lines []string
	for _, r := range results {
		if r == nil || r.Ok {
			continue
		}

		// 遍历错误单元格
		for _, errCell := range r.ErrCells {
			excelRow := errCell.Index + 1
			if errCell.ExcelRow > 0 {
				excelRow = errCell.ExcelRow
			}
			lines = append(lines, f.formatErrorLine("%s - %s 第%d行, %s",
				safeString(r.SheetName), r.DisplayName, excelRow, errCell.Reason))
		}

		// 整体错误
		if len(r.ErrCells) == 0 && r.Reason != "" {
			lines = append(lines, f.formatErrorLine("%s - %s, %s",
				safeString(r.SheetName), r.DisplayName, r.Reason))
		}
	}
	return lines
}

// FormatParseErrors 格式化解析错误
func (f *ErrorFormatter) FormatParseErrors(errors []*SheetParseError) []string {
	var lines []string
	for _, e := range errors {
		lines = append(lines, f.formatErrorLine("[%s] %s: %s",
			e.FileName, e.SheetName, e.Error))
	}
	return lines
}

// FormatContextInfo 格式化上下文信息（用于飞书卡片）
// 返回 Markdown 格式的上下文信息
func (f *ErrorFormatter) FormatContextInfo(event *CheckResultEvent) []string {
	var lines []string

	// 提交信息
	if event.CommitInfo != nil {
		if event.CommitInfo.Name != "" {
			lines = append(lines, fmt.Sprintf("**%s**提交了新配表", event.CommitInfo.Name))
		}
		if event.CommitInfo.Branch != "" {
			lines = append(lines, fmt.Sprintf("**提交分支**: %s", event.CommitInfo.Branch))
		}
		if event.CommitInfo.Version != "" {
			lines = append(lines, fmt.Sprintf("**版本**: %s", event.CommitInfo.Version))
		}

		// Merge 分支信息摘要
		if event.CommitInfo.MergeInfo != nil {
			mi := event.CommitInfo.MergeInfo
			if mi.MergeAuthor != "" {
				lines = append(lines, fmt.Sprintf("**Merge 操作**: %s 合并了被合并分支 → %s", mi.MergeAuthor, mi.ToBranch))
			}
			lines = append(lines, fmt.Sprintf("**Merge 提交**: 主分支 %d 个, 被合并分支 %d 个", mi.Parent1Total, mi.Parent2Total))
			if mi.Parent1Total > 0 {
				lines = append(lines, formatCommitList("主分支", mi.Parent1Commits, mi.Parent1Total))
			}
			if mi.Parent2Total > 0 {
				lines = append(lines, formatCommitList("被合并分支", mi.Parent2Commits, mi.Parent2Total))
			}
		}
	}

	// 变更文件
	if len(event.ChangedFiles) > 0 {
		lines = append(lines, fmt.Sprintf("**变更文件**: %s", strings.Join(event.ChangedFiles, ",")))
	}

	// 检查统计
	if event.Stats != nil {
		totalWithGeneral := event.Stats.TotalRules + event.Stats.GeneralRuleCount
		applicableWithGeneral := event.Stats.FilteredRules + event.Stats.GeneralRuleCount
		if event.IsIncremental {
			lines = append(lines, fmt.Sprintf("**增量检查**: 配置规则 %d 条 + 通用规则 %d 条 = %d 条，适用 %d 条 | 变更文件 %d 个",
				event.Stats.TotalRules, event.Stats.GeneralRuleCount, totalWithGeneral, applicableWithGeneral, event.Stats.ChangedFileCount))
		} else {
			lines = append(lines, fmt.Sprintf("**全量检查**: 配置规则 %d 条 + 通用规则 %d 条 = %d 条",
				event.Stats.TotalRules, event.Stats.GeneralRuleCount, totalWithGeneral))
		}
	}

	// 检查时间
	lines = append(lines, fmt.Sprintf("**检查时间**: %s", event.CheckTime.Format("2006-01-02 15:04:05")))

	return lines
}

// FormatCheckResultLines 格式化非 merge 场景的检查结果内容行
// 返回格式化的内容行切片，调用方可在此基础上添加前缀行（如 @ 用户）或包装格式（如检查时间标题）
// 适用于飞书卡片处理器和劫持处理器等需要灵活组合内容行的场景
func (f *ErrorFormatter) FormatCheckResultLines(event *CheckResultEvent) []string {
	summary := GetSummary(event)
	var lines []string

	// 上下文信息
	lines = append(lines, f.FormatContextInfo(event)...)

	// 解析错误
	if len(event.ParseErrors) > 0 {
		lines = append(lines, f.FormatParseErrors(event.ParseErrors)...)
	}

	// 列级检查错误
	if summary.ColErrors > 0 {
		lines = append(lines, f.FormatColErrors(event.ColResults)...)
	}

	// 表级检查错误
	if summary.TableErrors > 0 {
		lines = append(lines, f.FormatTableErrors(event.TableResults)...)
	}

	// 表级变更通知
	if summary.HasNotifications {
		lines = append(lines, f.FormatNotifications(event.TableResults)...)
	}

	return lines
}

// FormatConsoleOutput 格式化控制台输出
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
// 返回完整的控制台输出字符串
func (f *ErrorFormatter) FormatConsoleOutput(event *CheckResultEvent) string {
	var sb strings.Builder

	summary := GetSummary(event)

	sb.WriteString("\n========================================\n")
	sb.WriteString("📋 Excel 配表检查结果\n")
	sb.WriteString(fmt.Sprintf("检查时间: %s\n", event.CheckTime.Format("2006-01-02 15:04:05")))
	sb.WriteString("========================================\n")

	// 统计信息
	sb.WriteString("\n📊 检查统计:\n")
	sb.WriteString(fmt.Sprintf("  • 列级检查: %d 项 (失败: %d)\n", len(event.ColResults), summary.ColErrors))
	sb.WriteString(fmt.Sprintf("  • 表级检查: %d 项 (失败: %d)\n", len(event.TableResults), summary.TableErrors))

	// 列出执行的表级规则
	if len(event.TableResults) > 0 {
		sb.WriteString("  • 执行的表级规则:\n")
		// 提取规则名称（去重并排序）
		ruleNameSet := make(map[string]bool)
		var ruleNames []string
		for _, r := range event.TableResults {
			if r != nil && r.RuleType != "" {
				name := string(r.RuleType)
				if !ruleNameSet[name] {
					ruleNameSet[name] = true
					ruleNames = append(ruleNames, name)
				}
			}
		}
		sort.Strings(ruleNames)
		for _, ruleName := range ruleNames {
			sb.WriteString(fmt.Sprintf("    - %s\n", ruleName))
		}
	}
	sb.WriteString(fmt.Sprintf("  • 解析错误: %d 项\n", summary.ParseErrors))

	// 解析错误
	if len(event.ParseErrors) > 0 {
		sb.WriteString("\n❌ 解析错误:\n")
		for _, e := range event.ParseErrors {
			sb.WriteString(fmt.Sprintf("  • [%s] %s: %s\n", e.FileName, e.SheetName, e.Error))
		}
	}

	// 列级检查失败
	if summary.ColErrors > 0 {
		sb.WriteString("\n❌ 列级检查失败:\n")
		for _, line := range f.FormatColErrors(event.ColResults) {
			sb.WriteString(fmt.Sprintf("  • %s\n", line))
		}
	}

	// 表级检查失败
	if summary.TableErrors > 0 {
		sb.WriteString("\n❌ 表级检查失败:\n")
		for _, line := range f.FormatTableErrors(event.TableResults) {
			sb.WriteString(fmt.Sprintf("  • %s\n", line))
		}
	}

	// 表级变更通知（Ok=true 但有 ErrCells 的通知规则）
	// 通知规则不是错误，不需要发送到飞书，但需要在命令行日志中显示
	var notificationResults []*json_rule.TableCheckResult
	for _, r := range event.TableResults {
		if r != nil && r.Ok && len(r.ErrCells) > 0 {
			notificationResults = append(notificationResults, r)
		}
	}
	if len(notificationResults) > 0 {
		sb.WriteString("\n📝 表级变更通知:\n")
		for _, r := range notificationResults {
			sb.WriteString(fmt.Sprintf("  • %s\n", r.Reason))
		}
	}

	// 最终结果
	sb.WriteString("\n----------------------------------------\n")
	if summary.HasErrors {
		sb.WriteString(fmt.Sprintf("⚠️ 检查完成，发现 %d 个问题\n", summary.TotalErrors))
	} else {
		sb.WriteString("✅ 所有检查通过！\n")
	}
	sb.WriteString("========================================\n")

	return sb.String()
}

// FormatMergeContent 格式化 merge 场景的完整消息内容
// 用于飞书卡片和劫持处理器，生成包含 @ 标签的 Markdown 内容
// 结构：merge 概览 → 各 commit 分段检查结果 → 检查时间
//
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
func (f *ErrorFormatter) FormatMergeContent(event *CheckResultEvent) string {
	var parts []string

	// 1. Merge 概览区域（@ merge 作者 + 分支 commit 摘要 + 统计）
	if event.CommitInfo != nil && event.CommitInfo.MergeInfo != nil {
		mi := event.CommitInfo.MergeInfo
		// @ merge 作者
		if mi.MergeAuthor != "" {
			parts = append(parts, fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", mi.MergeAuthor))
		}
		// Merge 操作描述
		if mi.MergeAuthor != "" {
			parts = append(parts, fmt.Sprintf("**🔀 Merge 操作**: %s 合并了被合并分支 → %s", mi.MergeAuthor, mi.ToBranch))
		}
		parts = append(parts, fmt.Sprintf("**Merge 提交**: 主分支 %d 个, 被合并分支 %d 个", mi.Parent1Total, mi.Parent2Total))
		if mi.Parent1Total > 0 {
			parts = append(parts, formatCommitList("主分支", mi.Parent1Commits, mi.Parent1Total))
		}
		if mi.Parent2Total > 0 {
			parts = append(parts, formatCommitList("被合并分支", mi.Parent2Commits, mi.Parent2Total))
		}

		// 整体增量检查统计
		if event.Stats != nil {
			totalWithGeneral := event.Stats.TotalRules + event.Stats.GeneralRuleCount
			applicableWithGeneral := event.Stats.FilteredRules + event.Stats.GeneralRuleCount
			if event.IsIncremental {
				parts = append(parts, fmt.Sprintf("**增量检查**: 配置规则 %d 条 + 通用规则 %d 条 = %d 条，适用 %d 条 | 变更文件 %d 个",
					event.Stats.TotalRules, event.Stats.GeneralRuleCount, totalWithGeneral, applicableWithGeneral, event.Stats.ChangedFileCount))
			} else {
				parts = append(parts, fmt.Sprintf("**全量检查**: 配置规则 %d 条 + 通用规则 %d 条 = %d 条",
					event.Stats.TotalRules, event.Stats.GeneralRuleCount, totalWithGeneral))
			}
		}
	}

	// 2. 各 commit 分段检查结果
	for _, section := range event.CommitSections {
		parts = append(parts, "", "---")

		// @ 该 commit 的作者
		if section.Author != "" {
			parts = append(parts, fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", section.Author))
		}

		// 提交信息
		hash := section.CommitHash
		if len(hash) >= 8 {
			hash = hash[:8]
		}
		parts = append(parts, fmt.Sprintf("**提交**: %s %s提交了新配表", hash, section.Author))
		if section.Branch != "" {
			parts = append(parts, fmt.Sprintf("**提交分支**: %s", section.Branch))
		}
		parts = append(parts, fmt.Sprintf("**版本**: %s", hash))
		if len(section.DiffFiles) > 0 {
			parts = append(parts, fmt.Sprintf("**变更文件**: %s", strings.Join(section.DiffFiles, ",")))
		}

		// 该 commit 的统计
		if section.Stats != nil {
			totalWithGeneral := section.Stats.TotalRules + section.Stats.GeneralRuleCount
			applicableWithGeneral := section.Stats.FilteredRules + section.Stats.GeneralRuleCount
			parts = append(parts, fmt.Sprintf("**增量检查**: 配置规则 %d 条 + 通用规则 %d 条 = %d 条，适用 %d 条 | 变更文件 %d 个",
				section.Stats.TotalRules, section.Stats.GeneralRuleCount, totalWithGeneral, applicableWithGeneral, section.Stats.ChangedFileCount))
		}

		// 该 commit 执行的表级规则（去重并排序）
		if len(section.TableResults) > 0 {
			ruleNameSet := make(map[string]bool)
			var ruleNames []string
			for _, r := range section.TableResults {
				if r != nil && r.RuleType != "" {
					name := string(r.RuleType)
					if !ruleNameSet[name] {
						ruleNameSet[name] = true
						ruleNames = append(ruleNames, name)
					}
				}
			}
			sort.Strings(ruleNames)
			if len(ruleNames) > 0 {
				parts = append(parts, fmt.Sprintf("**执行的表级规则**: %s", strings.Join(ruleNames, ", ")))
			}
		}

		// 列级错误
		colLines := f.FormatColErrors(section.ColResults)
		if len(colLines) > 0 {
			parts = append(parts, "")
			parts = append(parts, colLines...)
		}

		// 表级错误
		tableLines := f.FormatTableErrors(section.TableResults)
		if len(tableLines) > 0 {
			parts = append(parts, "")
			parts = append(parts, tableLines...)
		}

		// 表级通知
		notifyLines := f.FormatNotifications(section.TableResults)
		if len(notifyLines) > 0 {
			parts = append(parts, "")
			parts = append(parts, notifyLines...)
		}
	}

	// 3. 检查时间
	parts = append(parts, "", "---")
	parts = append(parts, fmt.Sprintf("**检查时间**: %s", event.CheckTime.Format("2006-01-02 15:04:05")))

	return strings.Join(parts, "\n")
}

// FormatMergeConsoleOutput 格式化 merge 场景的控制台输出
// 结构与 FormatMergeContent 一致，但不含 <at> 和 <font> 标签
//
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
func (f *ErrorFormatter) FormatMergeConsoleOutput(event *CheckResultEvent) string {
	var sb strings.Builder

	summary := GetSummary(event)

	sb.WriteString("\n========================================\n")
	sb.WriteString("📋 Excel 配表检查结果（Merge）\n")
	sb.WriteString(fmt.Sprintf("检查时间: %s\n", event.CheckTime.Format("2006-01-02 15:04:05")))
	sb.WriteString("========================================\n")

	// Merge 概览
	if event.CommitInfo != nil && event.CommitInfo.MergeInfo != nil {
		mi := event.CommitInfo.MergeInfo
		sb.WriteString("\n🔀 Merge 概览:\n")
		if mi.MergeAuthor != "" {
			sb.WriteString(fmt.Sprintf("  • Merge 操作: %s 合并了被合并分支 → %s\n", mi.MergeAuthor, mi.ToBranch))
		}
		sb.WriteString(fmt.Sprintf("  • 主分支 %d 个提交, 被合并分支 %d 个提交\n", mi.Parent1Total, mi.Parent2Total))
		if mi.Parent1Total > 0 {
			sb.WriteString(fmt.Sprintf("    主分支: %s\n", formatCommitListPlain(mi.Parent1Commits, mi.Parent1Total)))
		}
		if mi.Parent2Total > 0 {
			sb.WriteString(fmt.Sprintf("    被合并分支: %s\n", formatCommitListPlain(mi.Parent2Commits, mi.Parent2Total)))
		}
	}

	// 整体统计
	sb.WriteString("\n📊 检查统计:\n")
	sb.WriteString(fmt.Sprintf("  • 列级检查: %d 项 (失败: %d)\n", len(event.ColResults), summary.ColErrors))
	sb.WriteString(fmt.Sprintf("  • 表级检查: %d 项 (失败: %d)\n", len(event.TableResults), summary.TableErrors))
	sb.WriteString(fmt.Sprintf("  • 解析错误: %d 项\n", summary.ParseErrors))

	// 各 commit 分段
	for i, section := range event.CommitSections {
		hash := section.CommitHash
		if len(hash) >= 8 {
			hash = hash[:8]
		}

		sb.WriteString(fmt.Sprintf("\n--- Commit %d: %s (%s) ---\n", i+1, hash, section.Author))
		if len(section.DiffFiles) > 0 {
			sb.WriteString(fmt.Sprintf("变更文件: %s\n", strings.Join(section.DiffFiles, ", ")))
		}

		// 该 commit 执行的表级规则（去重并排序）
		if len(section.TableResults) > 0 {
			ruleNameSet := make(map[string]bool)
			var ruleNames []string
			for _, r := range section.TableResults {
				if r != nil && r.RuleType != "" {
					name := string(r.RuleType)
					if !ruleNameSet[name] {
						ruleNameSet[name] = true
						ruleNames = append(ruleNames, name)
					}
				}
			}
			sort.Strings(ruleNames)
			if len(ruleNames) > 0 {
				sb.WriteString("  • 执行的表级规则:\n")
				for _, ruleName := range ruleNames {
					sb.WriteString(fmt.Sprintf("    - %s\n", ruleName))
				}
			}
		}

		// 列级错误
		for _, line := range f.FormatColErrors(section.ColResults) {
			sb.WriteString(fmt.Sprintf("  %s\n", line))
		}

		// 表级错误
		for _, line := range f.FormatTableErrors(section.TableResults) {
			sb.WriteString(fmt.Sprintf("  %s\n", line))
		}

		// 表级通知
		notifyLines := f.FormatNotifications(section.TableResults)
		for _, line := range notifyLines {
			sb.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	// 最终结果
	sb.WriteString("\n----------------------------------------\n")
	if summary.HasErrors {
		sb.WriteString(fmt.Sprintf("⚠️ 检查完成，发现 %d 个问题\n", summary.TotalErrors))
	} else {
		sb.WriteString("✅ 所有检查通过！\n")
	}
	sb.WriteString("========================================\n")

	return sb.String()
}

// formatCommitListPlain 格式化分支 commit 列表摘要（纯文本，不含 Markdown）
func formatCommitListPlain(commits []CommitSummary, total int) string {
	displayCount := len(commits)
	items := make([]string, displayCount)
	for i, c := range commits {
		if c.Author != "" {
			items[i] = fmt.Sprintf("%s(%s)", c.Hash, c.Author)
		} else {
			items[i] = c.Hash
		}
	}
	commitStr := strings.Join(items, ", ")
	if total > displayCount {
		return fmt.Sprintf("%s 等 %d 个提交（%d 个未显示）", commitStr, displayCount, total-displayCount)
	}
	return fmt.Sprintf("%s（共 %d 个）", commitStr, total)
}

// safeString 安全获取字符串指针的值
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// FormatCheckTime 格式化检查时间
func FormatCheckTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// formatCommitList 格式化分支的 commit 列表摘要
// 每个条目显示 hash(作者)，最多显示 MaxMergeDisplayCommits 个，超出的用文本提示
func formatCommitList(branchName string, commits []CommitSummary, total int) string {
	displayCount := len(commits)
	items := make([]string, displayCount)
	for i, c := range commits {
		if c.Author != "" {
			items[i] = fmt.Sprintf("%s(%s)", c.Hash, c.Author)
		} else {
			items[i] = c.Hash
		}
	}
	commitStr := strings.Join(items, ", ")
	if total > displayCount {
		return fmt.Sprintf("  %s: %s 等 %d 个提交（%d 个未显示）", branchName, commitStr, displayCount, total-displayCount)
	}
	return fmt.Sprintf("  %s: %s（共 %d 个）", branchName, commitStr, total)
}
