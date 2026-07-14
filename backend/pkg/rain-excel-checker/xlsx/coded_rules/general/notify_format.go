// Package coded_rules 提供通用表级别校验规则实现
// 本文件包含通知规则的共享格式化函数
// 从 table_check_new_row_notify.go 和 table_check_row_change_notify.go 提取
package coded_rules

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
)

// ==================== 通知格式化上下文 ====================

// formatNotifyContext 通知格式化上下文
// 包含通知消息中需要用到的元数据（Git 信息、表名、列名等）
type formatNotifyContext struct {
	sheetName   string
	idColName   string
	nameColName string
	commitTime  string
	committer   string
	baseCommit  string
	headCommit  string
}

// ==================== 行通知格式化 ====================

// formatAddedRowsReason 格式化新增行通知
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
func formatAddedRowsReason(rows []*diff.RowChange, ctx *formatNotifyContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 工作表变更通知 - %s\n\n", ctx.sheetName))
	sb.WriteString(fmt.Sprintf("<font color='green'>📁 %s | 🔄 新增行 | ✏️ +%d 行 | 🔑 %s %s</font>\n\n",
		ctx.sheetName, len(rows), ctx.idColName, formatRowIds(rows)))
	sb.WriteString(fmt.Sprintf("⏰ 变更时间: %s\n", ctx.commitTime))
	sb.WriteString(fmt.Sprintf("👤 提交人: %s", fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", ctx.committer)))
	return sb.String()
}

// formatRemovedRowsReason 格式化删除行通知
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
func formatRemovedRowsReason(rows []*diff.RowChange, ctx *formatNotifyContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 工作表变更通知 - %s\n\n", ctx.sheetName))
	sb.WriteString(fmt.Sprintf("~~📁 %s | 🔄 删除行 | ✏️ -%d 行 | 🔑 %s %s~~\n\n",
		ctx.sheetName, len(rows), ctx.idColName, formatRowIds(rows)))
	sb.WriteString(fmt.Sprintf("⏰ 变更时间: %s\n", ctx.commitTime))
	sb.WriteString(fmt.Sprintf("👤 提交人: %s", fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", ctx.committer)))
	return sb.String()
}

// formatRowIds 格式化行 ID 列表，最多显示 10 个，超过的用省略号
// 使用 RowChange.RowId（由 BuildSnapshot 根据主键列或第一列填充）
func formatRowIds(rows []*diff.RowChange) string {
	maxDisplay := 10
	if len(rows) <= maxDisplay {
		ids := make([]string, len(rows))
		for i, row := range rows {
			ids[i] = row.RowId
		}
		return strings.Join(ids, ",")
	}
	ids := make([]string, maxDisplay)
	for i := 0; i < maxDisplay; i++ {
		ids[i] = rows[i].RowId
	}
	return strings.Join(ids, ",") + fmt.Sprintf("...(共%d个)", len(rows))
}

// ==================== 列通知格式化 ====================

// formatAddedColsReason 格式化新增列通知
func formatAddedColsReason(cols []string, ctx *formatNotifyContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 工作表变更通知 - %s\n\n", ctx.sheetName))
	sb.WriteString(fmt.Sprintf("📁 %s | 🔄 新增列 | ✏️ +%d 列 | 📝 %s\n\n", ctx.sheetName, len(cols), formatColNames(cols)))
	sb.WriteString(fmt.Sprintf("⏰ 变更时间: %s\n", ctx.commitTime))
	sb.WriteString(fmt.Sprintf("👤 提交人: %s", fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", ctx.committer)))
	return sb.String()
}

// formatRemovedColsReason 格式化删除列通知
func formatRemovedColsReason(cols []string, ctx *formatNotifyContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 工作表变更通知 - %s\n\n", ctx.sheetName))
	sb.WriteString(fmt.Sprintf("📁 %s | 🔄 删除列 | ✏️ -%d 列 | 📝 %s\n\n", ctx.sheetName, len(cols), formatColNames(cols)))
	sb.WriteString(fmt.Sprintf("⏰ 变更时间: %s\n", ctx.commitTime))
	sb.WriteString(fmt.Sprintf("👤 提交人: %s", fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", ctx.committer)))
	return sb.String()
}

// formatColNames 格式化列名列表，最多显示 10 个，超过的用省略号
func formatColNames(cols []string) string {
	maxDisplay := 10
	if len(cols) <= maxDisplay {
		return strings.Join(cols, ", ")
	}
	return strings.Join(cols[:maxDisplay], ", ") + fmt.Sprintf("...(共%d个)", len(cols))
}

// ==================== 字段变更通知格式化 ====================

// fieldChangeRecord 字段变更记录
type fieldChangeRecord struct {
	rowId   string
	rowName string
	lineNo  int
	changes []struct {
		colName  string
		oldValue string
		newValue string
	}
}

// formatFieldChangeReason 格式化字段变更通知
// 消息格式变更时需同步更新飞书文档：https://ztgame.feishu.cn/wiki/UtAQwfxmmikkPRk26ivctT6lnMd（"当前实现的消息格式"章节）
func formatFieldChangeReason(records []*fieldChangeRecord, ctx *formatNotifyContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 表格名称: %s\n", ctx.sheetName))
	sb.WriteString("🔄 变更类型: 修改行\n")
	sb.WriteString(fmt.Sprintf("变更范围: 共 %d 行记录发生变更\n", len(records)))

	for i, record := range records {
		sb.WriteString(fmt.Sprintf("【变更记录 %d】\n", i+1))
		sb.WriteString(fmt.Sprintf("第 %d 行，%s %s\n", record.lineNo, ctx.idColName, record.rowId))
		for _, fc := range record.changes {
			// 先对原始值做 diff 高亮，再处理空值显示
			oldHighlighted, newHighlighted := diffTextHighlight(fc.oldValue, fc.newValue)
			// 空值替换为 (空)（diffTextHighlight 已处理空值情况，这里兜底）
			if fc.oldValue == "" {
				oldHighlighted = "(空)"
			}
			if fc.newValue == "" {
				newHighlighted = "(空)"
			}
			sb.WriteString(fmt.Sprintf("  ✏️ %s: \n [原值] %s \n [新值] %s\n", fc.colName, oldHighlighted, newHighlighted))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("⏰ 变更时间: %s\n", ctx.commitTime))
	sb.WriteString(fmt.Sprintf("👤 提交人: %s\n", fmt.Sprintf("<at email=\"%s@ztgame.com\"></at>", ctx.committer)))
	sb.WriteString(fmt.Sprintf("🔗 对比版本: %s → %s", ctx.baseCommit, ctx.headCommit))
	return sb.String()
}

// ==================== 文本差异高亮 ====================

// diffTextHighlight 对比两个字符串的差异，返回带飞书颜色标签的字符串
// 变化的部分用 <font color='red'>...</font> 标注
// 用于修改行通知中精确标注原值和新值的变化部分
func diffTextHighlight(oldValue, newValue string) (string, string) {
	// 边界：完全相同
	if oldValue == newValue {
		return oldValue, newValue
	}

	// 边界：oldValue 为空
	if oldValue == "" {
		return "(空)", fmt.Sprintf("<font color='red'>%s</font>", newValue)
	}

	// 边界：newValue 为空
	if newValue == "" {
		return fmt.Sprintf("<font color='red'>%s</font>", oldValue), "(空)"
	}

	// 按 rune（UTF-8 字符）比较，确保切片边界对齐到完整字符，避免切断多字节中文字符
	oldRunes := []rune(oldValue)
	newRunes := []rune(newValue)

	// 找公共前缀长度（rune 级别）
	prefixLen := 0
	minRuneLen := len(oldRunes)
	if len(newRunes) < minRuneLen {
		minRuneLen = len(newRunes)
	}
	for prefixLen < minRuneLen && oldRunes[prefixLen] == newRunes[prefixLen] {
		prefixLen++
	}

	// 找公共后缀长度（rune 级别，不能与前缀重叠）
	oldRemaining := len(oldRunes) - prefixLen
	newRemaining := len(newRunes) - prefixLen
	suffixLen := 0
	maxSuffix := oldRemaining
	if newRemaining < maxSuffix {
		maxSuffix = newRemaining
	}
	for suffixLen < maxSuffix &&
		oldRunes[len(oldRunes)-1-suffixLen] == newRunes[len(newRunes)-1-suffixLen] {
		suffixLen++
	}

	oldMiddleRunes := oldRunes[prefixLen : len(oldRunes)-suffixLen]
	newMiddleRunes := newRunes[prefixLen : len(newRunes)-suffixLen]

	// 如果中间部分为空（说明整个字符串都被前缀后缀覆盖，不应该发生因为已检查相等）
	if len(oldMiddleRunes) == 0 && len(newMiddleRunes) == 0 {
		return oldValue, newValue
	}

	prefixRunes := oldRunes[:prefixLen]
	oldSuffixRunes := oldRunes[len(oldRunes)-suffixLen:]
	newSuffixRunes := newRunes[len(newRunes)-suffixLen:]

	// 组装结果：前缀(公共) + 红色中间 + 后缀(公共)
	var oldResult, newResult string
	if len(oldMiddleRunes) > 0 {
		oldResult = string(prefixRunes) + fmt.Sprintf("<font color='red'>%s</font>", string(oldMiddleRunes)) + string(oldSuffixRunes)
	} else {
		oldResult = string(prefixRunes) + string(oldSuffixRunes)
	}
	if len(newMiddleRunes) > 0 {
		newResult = string(prefixRunes) + fmt.Sprintf("<font color='red'>%s</font>", string(newMiddleRunes)) + string(newSuffixRunes)
	} else {
		newResult = string(prefixRunes) + string(newSuffixRunes)
	}

	return oldResult, newResult
}

// ==================== 复合主键解析 ====================

// parseIdColNames 解析复合主键列名
// 如果 idColNamesStr 非空（逗号分隔），解析为多列数组
// 否则回退为单列 [idColName]
func parseIdColNames(idColNamesStr, idColName string) []string {
	if idColNamesStr != "" {
		parts := strings.Split(idColNamesStr, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return []string{idColName}
}
