// Package workflow 提供统一的配表检查工作流入口
// 封装 check_manager 的底层函数，为 CLI/Wails/MCP 三端提供一致的检查接口
package workflow

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/engine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// CheckMode 检查模式
type CheckMode string

const (
	// CheckModeFull 全量检查：检查本地磁盘所有 Excel 文件
	CheckModeFull CheckMode = "full"
	// CheckModeIncremental 增量检查：基于 git 变更或变更文件列表过滤检查
	CheckModeIncremental CheckMode = "incremental"
)

// CheckWorkflowConfig 检查工作流配置
type CheckWorkflowConfig struct {
	// ExcelPath Excel 文件目录路径
	ExcelPath string
	// CasePath 规则目录路径（Rules 为空时从此路径加载规则文件）
	CasePath string
	// Mode 检查模式：full 或 incremental
	Mode CheckMode
	// Rules 预加载的规则列表（非空时跳过磁盘加载，支持前端传入内存中的规则）
	Rules []*json_rule.SheetRule
	// ChangedFiles 增量模式：变更文件列表（用于 CheckWithFilter 过滤）
	ChangedFiles []string
	// Git 增量模式参数（CheckWithGitHistory 使用）
	RepoPath   string // git 仓库根目录
	CommitHash string // 要检查的 commit hash（空=HEAD）
	ParentHash string // 对比基准 commit hash
}

// WorkflowStats 工作流统计信息（独立类型，不引用 engine.CheckStats）
type WorkflowStats struct {
	TotalRules       int  // 加载的规则总数（JSON 配置 + 默认表级规则）
	FilteredRules    int  // 过滤后的规则数
	GeneralRuleCount int  // 通用规则数量
	IsIncremental    bool // 是否为增量检查
	ChangedFileCount int  // 变更文件数量
}

// SheetParseError Sheet 解析错误（统一类型，供 CLI/Wails/MCP 三端共用）
type SheetParseError struct {
	FileName  string `json:"fileName"`
	SheetName string `json:"sheetName"`
	Error     string `json:"error"`
}

// WorkflowResult 工作流执行结果
type WorkflowResult struct {
	ColResults   []*json_rule.ColCheckResult
	TableResults []*json_rule.TableCheckResult
	ParseErrors  []*SheetParseError
	Stats        *WorkflowStats
}

// RunCheckWorkflow 统一检查入口
// 内部根据 Mode 和参数自动选择执行路径：
//   - full + Rules非空: 使用传入规则执行全量检查
//   - full + Rules为空: 从 CasePath 加载规则后执行全量检查
//   - incremental + CommitHash非空: 通过 git 历史版本检查
//   - incremental + ChangedFiles非空: 基于变更文件过滤检查
func RunCheckWorkflow(cfg *CheckWorkflowConfig) (*WorkflowResult, error) {
	switch cfg.Mode {
	case CheckModeIncremental:
		return runIncrementalCheck(cfg)
	case CheckModeFull:
		fallthrough
	default:
		return runFullCheck(cfg)
	}
}

// runFullCheck 全量检查模式
// 调用 engine.CheckWithFilter 执行全量检查
// 如果 cfg.Rules 非空（来自 Wails 前端），额外收集 ParseErrors
func runFullCheck(cfg *CheckWorkflowConfig) (*WorkflowResult, error) {
	// 收集 ParseErrors（仅 Wails 端需要，即 Rules 非空时）
	var parseErrors []*SheetParseError
	if len(cfg.Rules) > 0 {
		excels, err := excelio.ReadFileOrDir(cfg.ExcelPath)
		if err != nil {
			return nil, fmt.Errorf("读取 Excel 目录失败: %w", err)
		}
		filter, err := excelio.ExcelFilter(excels)
		if err != nil {
			return nil, fmt.Errorf("过滤 Excel 文件失败: %w", err)
		}
		// 从过滤结果中收集解析错误
		for file, sheets := range filter {
			for _, sheet := range sheets {
				if sheet.Error != "" {
					parseErrors = append(parseErrors, &SheetParseError{
						FileName:  file.Path,
						SheetName: sheet.Name,
						Error:     sheet.Error,
					})
				}
			}
		}
		// 关闭文件（ParseErrors 已收集，不再需要）
		for _, excel := range excels {
			excel.Close()
		}
	}

	// 构建 check_manager 调用参数
	var opts []engine.CheckOption
	if len(cfg.Rules) > 0 {
		opts = append(opts, engine.WithPreloadedRules(cfg.Rules))
	}

	// 执行检查（changedFiles 传 nil 表示全量检查）
	colRes, tableRes, stats, err := engine.CheckWithFilter(
		cfg.ExcelPath, cfg.CasePath, nil, opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("全量检查失败: %w", err)
	}

	return &WorkflowResult{
		ColResults:   colRes,
		TableResults: tableRes,
		ParseErrors:  parseErrors,
		Stats:        mapStats(stats),
	}, nil
}

// runIncrementalCheck 增量检查模式
// 根据参数选择 CheckWithGitHistory 或 CheckWithFilter
func runIncrementalCheck(cfg *CheckWorkflowConfig) (*WorkflowResult, error) {
	var opts []engine.CheckOption
	if len(cfg.Rules) > 0 {
		opts = append(opts, engine.WithPreloadedRules(cfg.Rules))
	}

	// 优先使用 Git 历史检查（指定了 CommitHash 时）
	if cfg.CommitHash != "" {
		colRes, tableRes, stats, err := engine.CheckWithGitHistory(
			cfg.RepoPath, cfg.ExcelPath, cfg.CasePath,
			cfg.CommitHash, cfg.ParentHash, opts...,
		)
		if err != nil {
			return nil, fmt.Errorf("增量检查(git)失败: %w", err)
		}
		return &WorkflowResult{
			ColResults:   colRes,
			TableResults: tableRes,
			Stats:        mapStats(stats),
		}, nil
	}

	// 降级为基于变更文件列表的过滤检查
	colRes, tableRes, stats, err := engine.CheckWithFilter(
		cfg.ExcelPath, cfg.CasePath, cfg.ChangedFiles, opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("增量检查(filter)失败: %w", err)
	}

	return &WorkflowResult{
		ColResults:   colRes,
		TableResults: tableRes,
		Stats:        mapStats(stats),
	}, nil
}

// mapStats 将 engine.CheckStats 映射为 WorkflowStats
func mapStats(stats *engine.CheckStats) *WorkflowStats {
	if stats == nil {
		return nil
	}
	return &WorkflowStats{
		TotalRules:       stats.TotalRules,
		FilteredRules:    stats.FilteredRules,
		GeneralRuleCount: stats.GeneralRuleCount,
		IsIncremental:    stats.IsIncremental,
		ChangedFileCount: stats.ChangedFileCount,
	}
}
