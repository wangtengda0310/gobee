// Package notification 提供 Excel 检查结果的多通道输出能力
// 使用观察者模式，将检查结果分发到控制台、飞书、前端等多个通道
package notification

import (
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// CheckResultEvent 检查结果事件（统一的数据载体）
// 包含所有检查结果和上下文信息，用于分发到各个输出通道
type CheckResultEvent struct {
	// 检查结果
	ColResults   []*json_rule.ColCheckResult   // 列级检查结果
	TableResults []*json_rule.TableCheckResult // 表级检查结果
	ParseErrors  []*SheetParseError            // 文件解析错误

	// 上下文信息
	CheckTime     time.Time // 检查时间
	IsIncremental bool      // 是否增量检查
	ChangedFiles  []string  // 变更文件列表
	CommitInfo    *CommitInfo

	// 统计信息
	Stats *CheckStats

	// CommitSections merge 场景下按 commit 分段的检查结果
	// 非 merge 场景为 nil，handler 通过 len(CommitSections) > 0 判断是否走 merge 格式化路径
	// merge 模式下 ColResults/TableResults 保留聚合结果（GetSummary 不受影响）
	CommitSections []CommitSection
}

// CommitSection merge 场景下单个 commit 的检查结果分段
// 用于飞书/命令行/劫持三条通道统一格式化展示
type CommitSection struct {
	Author       string                        // 提交者
	AuthorEmail  string                        // 作者邮箱（用于飞书私聊）
	CommitHash   string                        // commit hash
	Branch       string                        // 提交分支
	DiffFiles    []string                      // 变更文件列表
	ColResults   []*json_rule.ColCheckResult   // 该 commit 的列级检查结果
	TableResults []*json_rule.TableCheckResult // 该 commit 的表级检查结果
	Stats        *CheckStats                   // 该 commit 的检查统计
}

// SheetParseError Sheet 解析错误
type SheetParseError struct {
	FileName  string `json:"fileName"`  // Excel文件名
	SheetName string `json:"sheetName"` // Sheet名称
	Error     string `json:"error"`     // 错误信息
}

// CommitInfo 提交信息
type CommitInfo struct {
	Name    string // 提交人
	Email   string // 提交人邮箱（用于飞书私聊）
	Branch  string // 提交分支
	Version string // 提交版本
	SkipDM  bool   // 是否跳过私聊通知（全量检查时设为 true）

	// Merge 场景下的分支 commit 摘要（可选）
	MergeInfo *MergeInfo
}

// MergeInfo merge commit 的分支信息摘要
type MergeInfo struct {
	// Merge 操作者信息
	MergeAuthor string // 执行 merge 操作的人
	FromBranch  string // 被合并的分支描述（如 commit message 中提取）
	ToBranch    string // 目标分支

	// 两个分支的 commit 摘要（各最多显示前 MaxMergeDisplayCommits 个）
	Parent1Commits []CommitSummary // 主分支 commit 摘要
	Parent1Total   int             // 主分支 commit 总数
	Parent2Commits []CommitSummary // 被合并分支 commit 摘要
	Parent2Total   int             // 被合并分支 commit 总数
}

// CommitSummary 单个 commit 的摘要信息（用于飞书消息展示）
type CommitSummary struct {
	Hash   string // commit hash（短格式，8位）
	Author string // 提交者
}

// MaxMergeDisplayCommits 飞书消息中每个分支最多显示的 commit 数量
const MaxMergeDisplayCommits = 5

// CheckStats 检查统计
type CheckStats struct {
	TotalRules       int // 总规则数（JSON 配置 + 默认表级规则）
	FilteredRules    int // 过滤后适用的规则数
	GeneralRuleCount int // 通用规则数量（NEW_ROW_NOTIFY + ROW_CHANGE_NOTIFY × 表数）
	ChangedFileCount int // 变更文件数
}

// CheckResultSummary 检查结果摘要
type CheckResultSummary struct {
	TotalErrors        int  // 总错误数
	ColErrors          int  // 列级错误数
	TableErrors        int  // 表级错误数（Ok=false 的数量）
	ParseErrors        int  // 解析错误数
	TableNotifications int  // 表级通知数（Ok=true 且有 ErrCells 的数量）
	HasErrors          bool // 是否有错误
	HasNotifications   bool // 是否有通知
}

// GetSummary 获取检查结果摘要
func GetSummary(event *CheckResultEvent) *CheckResultSummary {
	colErrors := countFailedColResults(event.ColResults)
	tableErrors := countFailedTableResults(event.TableResults)
	tableNotifications := countNotificationResults(event.TableResults)
	parseErrors := len(event.ParseErrors)

	return &CheckResultSummary{
		ColErrors:          colErrors,
		TableErrors:        tableErrors,
		ParseErrors:        parseErrors,
		TableNotifications: tableNotifications,
		TotalErrors:        colErrors + tableErrors + parseErrors,
		HasErrors:          colErrors+tableErrors+parseErrors > 0,
		HasNotifications:   tableNotifications > 0,
	}
}

// countFailedColResults 统计失败的列级检查数
func countFailedColResults(results []*json_rule.ColCheckResult) int {
	count := 0
	for _, r := range results {
		if r != nil && !r.Ok {
			count++
		}
	}
	return count
}

// countFailedTableResults 统计失败的表级检查数
func countFailedTableResults(results []*json_rule.TableCheckResult) int {
	count := 0
	for _, r := range results {
		if r != nil && !r.Ok {
			count++
		}
	}
	return count
}

// countNotificationResults 统计表级变更通知数
// 统计条件：Ok=true 且有 ErrCells（有实际变更内容）
func countNotificationResults(results []*json_rule.TableCheckResult) int {
	count := 0
	for _, r := range results {
		if r != nil && r.Ok && len(r.ErrCells) > 0 {
			count++
		}
	}
	return count
}
