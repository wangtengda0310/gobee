// Package diff 提供 Excel 配表差异检测与上下文管理功能
// 本文件提供 DiffResult 缓存层，避免多个通知规则重复执行 git show + Excel 解析
package diff

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/gitutil"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"github.com/xuri/excelize/v2"
)

// ==================== 缓存数据结构 ====================

// DiffCacheKey 差异结果缓存键
// 包含所有影响 diff 计算结果的参数，确保不同参数不会错误地命中缓存
type DiffCacheKey struct {
	ExcelPath   string // 当前 Excel 文件的完整路径
	BaseCommit  string // 对比的基准 Git commit
	SheetName   string // 工作表名称
	IdColKey    string // 规范化的 ID 列键（排序后 "\x01" 分隔）
	NameColName string // 名称列名（影响 BuildSnapshot 结果）
}

// DiffCacheEntry 差异结果缓存条目
type DiffCacheEntry struct {
	DiffResult *ExcelDiffResult  // 差异检测结果
	GitCtx     *GitNotifyContext // Git 提交元数据（供通知格式化使用）
	IsNewFile  bool              // 是否为新文件（git show 失败，无历史版本）
	NewFileMsg string            // 新文件通知消息
	Err        error             // 计算过程中的错误
}

// GitNotifyContext Git 提交元数据
// 缓存后供所有通知规则共享，避免每个规则重复执行 git log
type GitNotifyContext struct {
	Committer  string // 提交者
	CommitTime string // 提交时间（格式化后的字符串）
	HeadCommit string // HEAD commit hash
	BaseCommit string // 基准 commit hash
	IdColName  string // 展示用 ID 列名（用于通知格式化）
}

// DiffComputeParam 差异计算参数
// 封装完整的 diff 计算所需参数，避免依赖 json_rule 包
type DiffComputeParam struct {
	GitRepoPath string     // Git 仓库路径
	BaseCommit  string     // 基准 commit
	HeadCommit  string     // 当前 commit
	SheetName   string     // 工作表名称
	Cols        [][]string // 当前版本的列数据
	StartRowIdx int        // 数据起始行索引
	IdColNames  []string   // ID 列名列表（可能为复合主键）
	NameColName string     // 名称列名
	ExcelPath   string     // 当前 Excel 文件路径（用于 git show）
}

// ==================== 缓存管理 ====================

// diffCache 包级缓存，生命周期由调用方通过 ClearDiffCache 控制
var diffCache sync.Map

// NormalizeIdColKey 规范化 ID 列缓存键
// 多列 ID 时排序后用 "\x01" 拼接，确保顺序不影响缓存命中
func NormalizeIdColKey(idColNames []string) string {
	if len(idColNames) == 0 {
		return "Id"
	}
	if len(idColNames) == 1 {
		return idColNames[0]
	}
	sorted := make([]string, len(idColNames))
	copy(sorted, idColNames)
	sort.Strings(sorted)
	return strings.Join(sorted, "\x01")
}

// GetOrComputeDiff 获取或计算差异结果（带缓存）
// 如果缓存中存在结果直接返回，否则执行完整的 git show + 解析 + diff 流程并缓存
func GetOrComputeDiff(key DiffCacheKey, param DiffComputeParam) *DiffCacheEntry {
	if cached, ok := diffCache.Load(key); ok {
		return cached.(*DiffCacheEntry)
	}

	entry := computeDiff(param)
	diffCache.Store(key, entry)
	return entry
}

// ClearDiffCache 清除所有缓存条目
// 在 checkGeneralTableRules 入口和出口时调用
func ClearDiffCache() {
	diffCache.Range(func(key, _ interface{}) bool {
		diffCache.Delete(key)
		return true
	})
}

// PrefillDiffCache 预填充缓存（供测试使用）
// 单元测试中可通过此函数注入预设的 DiffResult，避免依赖真实 git 仓库
func PrefillDiffCache(key DiffCacheKey, entry *DiffCacheEntry) {
	diffCache.Store(key, entry)
}

// ResolveExcelPath 从 SheetMap 中解析当前 Excel 文件路径
// 遍历 SheetMap 找到包含目标 sheet 的 Excel 文件
func ResolveExcelPath(sheetMap map[string]*excelize.File, sheetName string) (string, *excelize.File) {
	for _, f := range sheetMap {
		sheets := f.GetSheetList()
		for _, s := range sheets {
			if s == sheetName {
				return f.Path, f
			}
		}
	}
	return "", nil
}

// ==================== 差异计算核心逻辑 ====================

// computeDiff 执行完整的差异计算流程
// 整合了原 table_check_new_row_notify.go 和 table_check_row_change_notify.go 中
// 约 70 行重复的 git show + ParseExcelFromBytes + BuildSnapshot + DetectDiff 逻辑
//
// 执行流程：
//  1. 解析 gitRepoPath
//  2. git show 获取历史版本文件（失败则标记为新文件）
//  3. ParseExcelFromBytes 解析历史 Excel
//  4. 根据 idColNames 长度选择 BuildSnapshot / BuildSnapshotWithCompositeKey
//  5. DetectDiff / DetectDiffWithCompositeKey 计算差异
//  6. 获取 Git 提交元数据（committer、date 等）
func computeDiff(param DiffComputeParam) *DiffCacheEntry {
	entry := &DiffCacheEntry{}

	// 步骤 1: 解析 gitRepoPath
	gitRepoPath := param.GitRepoPath
	if gitRepoPath == "." {
		gitRepoPath = filepath.Dir(param.ExcelPath)
	}

	// 步骤 2: git show 获取历史版本文件
	oldCols, err := gitutil.GetFileAtCommit(gitRepoPath, param.BaseCommit, param.ExcelPath)
	if err != nil {
		// parent commit 中不存在该文件 → 新增文件
		entry.IsNewFile = true
		entry.NewFileMsg = fmt.Sprintf("新增文件: %s", filepath.Base(param.ExcelPath))
		return entry
	}

	// 步骤 3: 解析历史 Excel 文件
	oldFileCols, err := ParseExcelFromBytes(oldCols, param.SheetName)
	if err != nil {
		entry.Err = fmt.Errorf("解析历史版本文件失败: %v", err)
		return entry
	}

	// 步骤 4-5: 构建快照并检测差异
	var diffResult *ExcelDiffResult
	if len(param.IdColNames) > 1 {
		oldSnapshot := BuildSnapshotWithCompositeKey(param.SheetName, oldFileCols, param.StartRowIdx, param.IdColNames, param.NameColName)
		diffResult = DetectDiffWithCompositeKey(oldSnapshot, param.Cols, param.StartRowIdx, param.IdColNames, param.NameColName)
	} else {
		idColName := "Id"
		if len(param.IdColNames) == 1 {
			idColName = param.IdColNames[0]
		}
		oldSnapshot := BuildSnapshot(param.SheetName, oldFileCols, param.StartRowIdx, idColName, param.NameColName)
		diffResult = DetectDiff(oldSnapshot, param.Cols, param.StartRowIdx, idColName, param.NameColName)
	}

	// 步骤 6: 获取 Git 提交元数据
	gitCtx := computeGitNotifyContext(gitRepoPath, param.BaseCommit, param.HeadCommit, param.IdColNames)

	entry.DiffResult = diffResult
	entry.GitCtx = gitCtx
	return entry
}

// computeGitNotifyContext 获取 Git 提交元数据
// 包括提交者、提交时间、HEAD commit hash 等
func computeGitNotifyContext(gitRepoPath, baseCommit, headCommit string, idColNames []string) *GitNotifyContext {
	ctx := &GitNotifyContext{
		BaseCommit: baseCommit,
	}

	// 展示用 ID 列名
	if len(idColNames) > 0 {
		ctx.IdColName = idColNames[0]
	} else {
		ctx.IdColName = "Id"
	}

	repoRoot, err := gitutil.GetRepoRoot(gitRepoPath)
	if err != nil {
		ctx.CommitTime = time.Now().Format("2006-01-02 15:04:05")
		return ctx
	}

	// 使用 HEAD_COMMIT 参数或 fallback 到 HEAD
	lookupHash := headCommit
	if lookupHash == "" {
		lookupHash = "HEAD"
	}

	if _, _, author, err := gitutil.GetCommitInfo(repoRoot, lookupHash); err == nil {
		ctx.Committer = author
	}

	if t, err := gitutil.GetCommitDate(repoRoot, lookupHash); err == nil {
		ctx.CommitTime = t.Format("2006-01-02 15:04:05")
	}

	// 如果 headCommit 为空，回填实际 HEAD hash
	if headCommit == "" {
		if hash, _, _, err := gitutil.GetCommitInfo(repoRoot, ""); err == nil {
			ctx.HeadCommit = hash
		}
	} else {
		ctx.HeadCommit = headCommit
	}

	if ctx.CommitTime == "" {
		ctx.CommitTime = time.Now().Format("2006-01-02 15:04:05")
	}

	return ctx
}

// BuildDiffCacheKey 构建缓存键的便捷函数
// 封装了 NormalizeIdColKey 的调用，方便规则使用
func BuildDiffCacheKey(excelPath, baseCommit, sheetName, nameColName string, idColNames []string) DiffCacheKey {
	return DiffCacheKey{
		ExcelPath:   excelPath,
		BaseCommit:  baseCommit,
		SheetName:   sheetName,
		IdColKey:    NormalizeIdColKey(idColNames),
		NameColName: nameColName,
	}
}

// ParseDiffParams 从参数 map 中解析 diff 计算所需的参数
// 各通知规则可调用此函数统一解析参数，避免重复代码
func ParseDiffParams(params map[string]string) (idColNames []string, nameColName, gitRepoPath, baseCommit, headCommit string) {
	idColName := params["idColName"]
	idColNamesStr := params["idColNames"]
	nameColName = params["nameColName"]
	gitRepoPath = params["gitRepoPath"]
	baseCommit = params["baseCommit"]
	headCommit = params["headCommit"]

	// 解析复合主键：idColNames 优先
	idColNames = ParseIdColNamesForKey(idColNamesStr, idColName)
	return
}

// ParseIdColNamesForKey 解析复合主键列名（用于缓存键规范化）
// idColNamesStr 优先级高于 idColName
// 复制自 coded_rules/general/parse_id_col_names.go，供 check_internal 包使用
func ParseIdColNamesForKey(idColNamesStr, idColName string) []string {
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

	// 回退到单列 ID
	if idColName == "" {
		idColName = "Id"
	}
	return []string{idColName}
}

// BuildDiffComputeParam 从规则参数构建 DiffComputeParam
// 封装了参数解析和 Excel 路径解析，方便规则使用
//
// now 参数用于 Excel 内部处理（保留兼容，当前未使用）
func BuildDiffComputeParam(param map[string]string, sheetMap map[string]*excelize.File) (DiffComputeParam, bool) {
	idColNames, nameColName, gitRepoPath, baseCommit, headCommit := ParseDiffParams(param)

	excelPath, xlsxFile := ResolveExcelPath(sheetMap, param["sheetName"])
	if xlsxFile == nil {
		return DiffComputeParam{}, false
	}

	return DiffComputeParam{
		GitRepoPath: gitRepoPath,
		BaseCommit:  baseCommit,
		HeadCommit:  headCommit,
		SheetName:   param["sheetName"],
		Cols:        nil, // 由调用方填充（从 xlsxFile.GetCols 获取）
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		IdColNames:  idColNames,
		NameColName: nameColName,
		ExcelPath:   excelPath,
	}, true
}
