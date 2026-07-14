// Package engine 提供检查功能的入口函数
// 本包提供执行所有检查的顶层函数
package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/gitutil"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/ruleconfig"
	"github.com/xuri/excelize/v2"
)

// CheckOption CheckWithGitHistory 的可选参数配置
type CheckOption func(*checkOptions)

type checkOptions struct {
	fallbackSheetMap  map[string]*excelize.File
	preloadedRules    []*json_rule.SheetRule     // 预加载的规则列表（跳过内部 LoadJsonRules）
	sheetMapOutputPtr *map[string]*excelize.File // 导出 sheetMap 的目标指针（供缓存传递）
}

// WithFallbackSheetMap 设置预加载的 fallback sheetMap（用于merge场景缓存）
func WithFallbackSheetMap(sheetMap map[string]*excelize.File) CheckOption {
	return func(opt *checkOptions) {
		opt.fallbackSheetMap = sheetMap
	}
}

// WithPreloadedRules 使用预加载的规则列表，跳过 CheckWithGitHistory 内部的 LoadJsonRules 调用
// 用于 merge 场景：在外层加载一次规则，所有 commit 共享同一份规则列表
func WithPreloadedRules(rules []*json_rule.SheetRule) CheckOption {
	return func(opt *checkOptions) {
		opt.preloadedRules = rules
	}
}

// WithSheetMapOutput 设置 sheetMap 导出目标，CheckWithGitHistory 完成后将构建的 sheetMap 写入此处
// 用于 merge 场景：将当前 commit 的 sheetMap 传递给下一个 commit 作为缓存
func WithSheetMapOutput(ptr *map[string]*excelize.File) CheckOption {
	return func(opt *checkOptions) {
		opt.sheetMapOutputPtr = ptr
	}
}

// CheckStats 检查统计信息
type CheckStats struct {
	TotalRules       int  // 加载的规则总数（JSON 配置 + 默认表级规则）
	FilteredRules    int  // 过滤后的规则数
	GeneralRuleCount int  // 通用规则数量（5 个拆分规则 × 表数）
	IsIncremental    bool // 是否为增量检查
	ChangedFileCount int  // 变更文件数量
}

// CheckAll 检查所有配置的表
// 这是执行 Excel 检查的入口函数，会加载规则并检查所有配置的表
// 参数:
//   - excelPath: Excel 文件目录路径
//   - casePath: 检查规则配置文件路径
//
// 返回:
//   - 不合格的检查结果列表
//   - 表级检查结果列表（包括通用规则）
//   - 检查统计信息
//   - 错误信息
func CheckAll(excelPath, casePath string) (okRes []*json_rule.ColCheckResult, tableRes []*json_rule.TableCheckResult, stats *CheckStats, err error) {
	return CheckWithFilter(excelPath, casePath, nil)
}

// shouldRunGeneralRules 判断是否应执行通用规则（NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY）
// 通用规则需要 git diff 基准，全量检查（changedFiles 为空/nil）时跳过
func shouldRunGeneralRules(changedFiles []string) bool {
	return len(changedFiles) > 0
}

// CheckWithFilter 检查所有配置的表，支持变更文件过滤
// 这是执行 Excel 检查的入口函数，支持增量检查模式
//
// 参数:
//   - excelPath: Excel 文件目录路径
//   - casePath: 检查规则配置文件路径
//   - changedFiles: 变更文件列表（空列表表示全量检查）
//   - opts: 可选参数配置（如 WithPreloadedRules 设置预加载规则）
//
// 返回:
//   - 不合格的检查结果列表
//   - 表级检查结果列表（包括通用规则如 NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）
//   - 检查统计信息
//   - 错误信息
//
// 分层设计:
//  1. 数据层：全量加载所有 Excel 文件到内存（GetSheetMap）
//  2. 规则层：加载列级规则（JSON）+ 表级规则（代码 init 注册）
//     2.1 补充默认表级规则（supplementDefaultTableRules）
//     2.2 补充 ParamDefs 默认参数（supplementDefaultParams）
//  3. 过滤层：根据 git diff 文件过滤规则（FilterRulesByChangedFiles）
//  4. 校验层：被选中的规则在全量数据上执行（CheckSheetCols + CheckTableRules）
//  5. 通用规则层：执行通用表级规则（NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY，全量模式跳过）
func CheckWithFilter(excelPath, casePath string, changedFiles []string, opts ...CheckOption) (okRes []*json_rule.ColCheckResult, tableRes []*json_rule.TableCheckResult, stats *CheckStats, err error) {
	// 初始化返回值
	okRes = make([]*json_rule.ColCheckResult, 0)
	tableRes = make([]*json_rule.TableCheckResult, 0)
	stats = &CheckStats{
		IsIncremental:    len(changedFiles) > 0,
		ChangedFileCount: len(changedFiles),
	}

	// 解析可选参数
	options := &checkOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 1. 数据层：全量加载所有 Excel 文件
	sheetMap, err := excelio.GetSheetMap(excelPath)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. 规则层：加载列级规则（JSON）
	// 支持外部预加载规则列表（WithPreloadedRules），跳过磁盘加载
	var list []*json_rule.SheetRule
	if options.preloadedRules != nil {
		list = deepCopyRules(options.preloadedRules)
	} else {
		list, err = ruleconfig.LoadJsonRules(casePath)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// ⚠️ 顺序关键：先预过滤 sheetMap，再补充默认规则，最后基于完整规则二次过滤
	//
	// 历史问题：
	// - 80ab4f1 将 filterSheetMapByRules 移到 supplementDefaultTableRules 之后，
	//   导致 supplementDefaultTableRules 遍历全量 sheetMap 为所有表补充默认规则，
	//   右键分类执行时实际运行了全量检查。
	// - 8974092 修复了 SEASON_PASS_HERO_OPEN_CHECK 等规则的 RequiredSheets 声明，
	//   但 filterSheetMapByRules 在 supplementDefaultTableRules 之前执行，此时 list
	//   中尚未包含默认规则，导致默认规则的 RequiredSheets 对应的关联表被过滤掉。
	//
	// 正确顺序：
	// 1. 预过滤：基于 list 中的 Sheet 字段，只保留规则列表中提到的表
	// 2. 补充默认规则：为预过滤后的 sheetMap 中的表补充默认规则
	// 3. 二次过滤：基于完整的规则列表（包含默认规则）从原始 sheetMap 中过滤
	//
	// 注意：步骤3必须从原始 sheetMap 过滤，因为步骤1过滤后的 sheetMap 可能缺少
	// 默认规则的 RequiredSheets 对应的关联表。
	var originalSheetMap map[string]*excelize.File
	if options.preloadedRules != nil {
		// 保存原始 sheetMap 引用（用于步骤3）
		originalSheetMap = sheetMap

		// 步骤1：预过滤 — 只保留 list 中提到的表
		preFiltered := make(map[string]*excelize.File)
		for _, rule := range list {
			for sheetName, file := range sheetMap {
				if isRelevantSheet(sheetName, map[string]bool{rule.Sheet: true}) {
					preFiltered[sheetName] = file
				}
			}
		}
		sheetMap = preFiltered
	}

	// 步骤2：补充有默认表级规则但没有 JSON 文件的表
	// 此时 sheetMap 已预过滤，只补充相关表的默认规则
	list = supplementDefaultTableRules(list, sheetMap)

	// 步骤3：基于完整的规则列表（包含默认规则）从原始 sheetMap 中过滤
	// 此时 list 中已包含默认规则，filterSheetMapByRules 会正确加载 RequiredSheets
	if options.preloadedRules != nil {
		sheetMap = filterSheetMapByRules(originalSheetMap, list)
	}

	// 2.3 为已有 JSON 配置的 TableRule 补充 ParamDefs 中定义的默认参数
	SupplementDefaultParams(list)

	// 3. 过滤层：根据变更文件过滤规则
	originalCount := len(list)
	list = FilterRulesByChangedFiles(list, changedFiles)

	// 统计通用规则数量
	// 通用规则（NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）需要 git diff 基准
	// 全量检查时无 diff 基准，跳过通用规则
	var generalRuleCount int
	if shouldRunGeneralRules(changedFiles) {
		// 增量检查：只对变更文件对应的表执行通用规则
		// 拆分后为 5 个独立规则（ADDED_ROW/REMOVED_ROW/ADDED_COL/REMOVED_COL/MODIFIED_ROW）
		generalRuleCount = 5 * len(list)
	}
	// else: 全量检查跳过通用规则，generalRuleCount = 0

	// 计算总规则数 = 配置规则 + 通用规则
	totalRulesWithGeneral := originalCount + generalRuleCount
	// 适用规则数 = 过滤后的配置规则 + 通用规则
	applicableRules := len(list) + generalRuleCount

	fmt.Printf("配置规则: %d 条 + 通用规则 %d 条 = %d 条, 适用规则: %d 条\n",
		originalCount, generalRuleCount, totalRulesWithGeneral, applicableRules)

	// 更新统计信息
	stats.TotalRules = originalCount          // 配置规则数（不含通用规则）
	stats.FilteredRules = len(list)           // 过滤后适用的配置规则
	stats.GeneralRuleCount = generalRuleCount // 通用规则数

	// 4. 校验层：执行检查
	for _, rule := range list {
		if xlsx, exist := sheetMap[rule.Sheet]; !exist {
			fmt.Printf("表不存在: %s\n", rule.Sheet)
		} else {
			fmt.Printf("检查表: %s\n", rule.Sheet)

			// 4.1 列级检查
			colResults, _ := CheckSheetCols(xlsx, rule, sheetMap)
			// 累积不合格的结果到返回值
			for _, re := range colResults {
				if !re.Ok {
					okRes = append(okRes, re)
				}
			}

			// 4.2 表级检查（从 JSON 配置中的 TableRules）
			tableResults, _ := CheckTableRules(xlsx, rule, sheetMap)
			// 收集所有表级结果（包括通知类型 Ok=true 的结果）
			tableRes = append(tableRes, tableResults...)

			// 打印当前表的不合格结果
			errRes := make([]*json_rule.ColCheckResult, 0)
			for _, re := range colResults {
				if !re.Ok {
					errRes = append(errRes, re)
				}
			}
			jsonData, _ := json.MarshalIndent(errRes, "", " ")
			fmt.Println(string(jsonData))
		}
	}

	// 5. 通用规则层：仅增量模式执行（全量模式无 diff 基准，跳过）
	if shouldRunGeneralRules(changedFiles) {
		generalTableResults := checkGeneralTableRules(sheetMap)
		tableRes = append(tableRes, generalTableResults...)

		fmt.Printf("通用表级规则检查完成，返回 %d 条结果\n", len(generalTableResults))
	}

	return okRes, tableRes, stats, nil
}

// checkGeneralTableRules 执行通用表级规则检查
// 通用规则是指 TargetSheets 为空的规则，它们自动应用于所有表
// 包括：NEW_ROW_NOTIFY（新增行/列通知）、ROW_CHANGE_NOTIFY（行变更字段通知）
func checkGeneralTableRules(sheetMap map[string]*excelize.File) []*json_rule.TableCheckResult {
	// 清理上一次检查的缓存
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	results := make([]*json_rule.TableCheckResult, 0)

	// 获取所有表级规则元数据
	allMetas := json_rule.GetAllTableRuleMetas()

	// 筛选通用规则（TargetSheets 为空）
	var generalMetas []*json_rule.TableRuleMeta
	for _, meta := range allMetas {
		if len(meta.TargetSheets) == 0 {
			generalMetas = append(generalMetas, meta)
		}
	}

	// 如果没有通用规则，直接返回
	if len(generalMetas) == 0 {
		return results
	}

	// 对每个 sheet 执行通用规则
	for sheetName, xlsxFile := range sheetMap {
		// 获取表的列数据
		cols, err := xlsxFile.GetCols(sheetName)
		if err != nil {
			continue
		}

		// 对每个通用规则执行检查
		for _, meta := range generalMetas {
			// 跳过已废弃的旧枚举（由拆分后的 5 个新规则替代）
			if meta.Type == json_rule.NEW_ROW_NOTIFY || meta.Type == json_rule.ROW_CHANGE_NOTIFY {
				continue
			}

			// 获取检查器
			checker := TableManager.GetChecker(meta.Type)
			if checker == nil {
				continue
			}

			// 构建默认参数，并合并 per-sheet 覆盖
			params := meta.ResolveParams(nil)
			if sheetOverrides := json_rule.GetGeneralRuleOverrides(sheetName); len(sheetOverrides) > 0 {
				if ruleOverrides, ok := sheetOverrides[meta.Type]; ok {
					for k, v := range ruleOverrides {
						// 保护不应被 per-sheet 覆盖的关键参数
						if k != string(json_rule.BASE_COMMIT) && k != string(json_rule.GIT_REPO_PATH) {
							params[k] = v
						}
					}
				}
			}

			// 执行检查
			result := checker.Check(json_rule.CheckParam{
				SheetName:   sheetName,
				Cols:        cols,
				StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
				Params:      params,
				SheetMap:    sheetMap,
			})
			if result == nil {
				continue
			}

			// 只收集有内容的结果（有 ErrCells 或有 Reason）
			hasContent := result.Ok || len(result.ErrCells) > 0
			if hasContent {
				result.SheetName = &sheetName
				result.RuleType = meta.Type
				result.DisplayName = meta.DisplayName
				result.TableName = &xlsxFile.Path
				results = append(results, result)
			}
		}
	}

	return results
}

// CheckWithGitHistory 基于 git 历史版本执行增量检查
// 与 CheckWithFilter 不同，本函数通过 git show 获取 commit 历史版本的文件，
// 不读取本地磁盘文件。用于 merge 场景遍历子 commit。
//
// 参数：
//   - repoPath: git 仓库根目录的绝对路径
//   - excelPath: Excel 文件在仓库中的相对目录（如 "" 或 "excel"）
//   - casePath: 检查规则目录路径
//   - commitHash: 要检查的 commit hash
//   - parentHash: 该 commit 的父 commit hash（作为通用规则的对比基准）
//   - opts: 可选参数配置（如 WithFallbackSheetMap 设置预加载缓存）
func CheckWithGitHistory(repoPath, excelPath, casePath, commitHash, parentHash string, opts ...CheckOption) (
	okRes []*json_rule.ColCheckResult,
	tableRes []*json_rule.TableCheckResult,
	stats *CheckStats,
	err error,
) {
	startTime := time.Now()

	// 解析可选参数
	options := &checkOptions{}
	for _, opt := range opts {
		opt(options)
	}

	okRes = make([]*json_rule.ColCheckResult, 0)
	tableRes = make([]*json_rule.TableCheckResult, 0)
	stats = &CheckStats{IsIncremental: true}

	// 1. 获取该 commit 的变更文件列表
	diffFiles, err := gitutil.GetCommitDiffFiles(repoPath, commitHash)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取 commit %s 变更文件失败: %w", commitHash[:8], err)
	}

	// 2. 过滤出 xlsx 文件
	var xlsxFiles []string
	for _, f := range diffFiles {
		if strings.HasSuffix(strings.ToLower(f), ".xlsx") {
			xlsxFiles = append(xlsxFiles, f)
		}
	}
	stats.ChangedFileCount = len(xlsxFiles)

	if len(xlsxFiles) == 0 {
		fmt.Printf("commit %s 没有变更的 xlsx 文件，跳过\n", commitHash[:8])
		return okRes, tableRes, stats, nil
	}

	// 3. 通过 git show 获取每个变更文件在该 commit 时的内容
	filesData := make(map[string][]byte)
	for _, relPath := range xlsxFiles {
		data, err := gitutil.GetFileAtCommit(repoPath, commitHash, filepath.Join(repoPath, relPath))
		if err != nil {
			fmt.Printf("[警告] 获取文件 %s@%s 失败: %v\n", relPath, commitHash[:8], err)
			continue
		}
		filesData[relPath] = data
	}

	if len(filesData) == 0 {
		fmt.Printf("commit %s 没有成功获取的 xlsx 文件，跳过\n", commitHash[:8])
		return okRes, tableRes, stats, nil
	}

	// 4. 从字节数据构建 sheetMap
	sheetMap, err := excelio.GetSheetMapFromBytes(filesData, repoPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("构建 sheetMap 失败: %w", err)
	}

	// 4.1 【性能优化】在补充 sheetMap 之前执行通用规则检查
	// 此时 sheetMap 只包含变更的文件，通用规则只需检查这些 sheet
	// 避免 supplement 后检查全量 sheet，大幅提升性能
	generalTableResults := checkGeneralTableRulesWithBaseCommit(sheetMap, parentHash, commitHash)

	// 4.1.1 记录通用规则计数（supplement 前的 sheetMap 大小 × 5 个拆分规则）
	generalRuleCount := 5 * len(sheetMap)

	// 5. 加载规则（在 supplement 之前），以便收集 RequiredSheets
	// 支持外部预加载规则列表（WithPreloadedRules），避免每个 commit 重复加载
	var list []*json_rule.SheetRule
	if options.preloadedRules != nil {
		list = deepCopyRules(options.preloadedRules) // 深拷贝，避免跨 commit 修改共享数据
	} else {
		list, err = ruleconfig.LoadJsonRules(casePath)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// 5.1 补充有默认表级规则但没有 JSON 文件的表
	// 此时 sheetMap 只包含变更文件，supplementDefaultTableRules 会为这些表补充默认规则
	// 其中包含 RequiredSheets 声明（如 HERO_DROP_CHECK 需要 DropItem、SeasonPassReward 等）
	list = supplementDefaultTableRules(list, sheetMap)

	// 5.2 为已有 JSON 配置的 TableRule 补充 ParamDefs 中定义的默认参数
	SupplementDefaultParams(list)

	// 4.2 【性能优化】按需补充关联表
	// 收集规则声明的所有 RequiredSheets，只加载这些关联表而非全量 261 个 sheet
	requiredSheets := collectRequiredSheets(list, sheetMap)
	// [DEBUG] 补充前状态
	supplementSheetMap(sheetMap, repoPath, parentHash, options.fallbackSheetMap, requiredSheets)
	// [DEBUG] 补充后状态

	// 6. 过滤规则（使用变更文件列表）
	// 注意：在 git 历史场景下，sheetMap 已经只包含变更的文件，
	// supplementDefaultTableRules 也只为 sheetMap 中的表补充规则，
	// 所以不需要再过滤，否则会导致规则被错误过滤掉。
	originalCount := len(list)
	// list = FilterRulesByChangedFiles(list, xlsxFiles) // 注释掉，不过滤

	stats.TotalRules = originalCount
	stats.FilteredRules = len(list)
	stats.GeneralRuleCount = generalRuleCount

	fmt.Printf("commit %s: 变更 %d 个 xlsx, 配置规则 %d, 通用规则 %d\n",
		commitHash[:8], len(xlsxFiles), originalCount, generalRuleCount)

	// 7. 执行列级和表级检查
	for _, rule := range list {
		fmt.Printf("- rule.Sheet %s \n", rule.Sheet)

		if xlsx, exist := sheetMap[rule.Sheet]; !exist {
			fmt.Printf("表不存在: %s\n", rule.Sheet)
		} else {
			colResults, _ := CheckSheetCols(xlsx, rule, sheetMap)
			for _, re := range colResults {
				if !re.Ok {
					okRes = append(okRes, re)
				}
			}

			tableResults, _ := CheckTableRules(xlsx, rule, sheetMap)
			tableRes = append(tableRes, tableResults...)
		}
	}

	// 8. 将前面检查的通用规则结果追加到结果中
	tableRes = append(tableRes, generalTableResults...)

	fmt.Printf("[耗时] commit %s 总检查耗时: %v\n", commitHash[:8], time.Since(startTime))

	// 导出 sheetMap 供调用方缓存复用（用于 merge 场景的跨 commit 缓存传递）
	if options.sheetMapOutputPtr != nil {
		*options.sheetMapOutputPtr = sheetMap
	}

	return okRes, tableRes, stats, nil
}

// collectRequiredSheets 从规则列表中收集所有需要的关联表名
// 包括：表级规则的 RequiredSheets + 列级规则的 REFERENCE_SHEET 参数
// 已存在于 sheetMap 中的表不需要再补充
func collectRequiredSheets(rules []*json_rule.SheetRule, sheetMap map[string]*excelize.File) map[string]bool {
	required := make(map[string]bool)

	for _, rule := range rules {
		// 1. 收集表级规则的 RequiredSheets
		for _, tableRule := range rule.TableRules {
			if !tableRule.Enabled {
				continue
			}
			meta := json_rule.TableRuleMetas[tableRule.Type]
			if meta == nil {
				continue
			}
			for _, sheet := range meta.RequiredSheets {
				// 跳过已存在于 sheetMap 的表
				if _, exists := sheetMap[sheet]; !exists {
					required[sheet] = true
				}
			}
		}

		// 2. 收集列级规则中的跨表引用参数
		for _, colRule := range rule.Rules {
			for _, cr := range colRule.PropRules {
				for _, paramKey := range json_rule.ReferenceSheetParamKeys {
					if refSheet, ok := cr.Params[paramKey]; ok && refSheet != "" {
						if _, exists := sheetMap[refSheet]; !exists {
							required[refSheet] = true
						}
					}
				}
				// 3. CHAIN_REFERENCE: chainSteps JSON 中提取引用的表名
				if cr.Params["chainSteps"] != "" {
					for _, sheet := range extractChainStepSheets(cr.Params) {
						if _, exists := sheetMap[sheet]; !exists {
							required[sheet] = true
						}
					}
				}
			}
		}
	}

	return required
}

// supplementSheetMap 补充 sheetMap 中缺失的表
// 优先使用预加载缓存，缓存不可用时降级到从 git commit 加载
// 支持按需补充：只加载 requiredSheets 中声明的关联表，而非全量加载
func supplementSheetMap(sheetMap map[string]*excelize.File, repoPath, commitHash string, cache map[string]*excelize.File, requiredSheets map[string]bool) {
	startTime := time.Now()
	beforeCount := len(sheetMap)

	// requiredSheets 为 nil 表示按需补充但未指定具体表，此时跳过
	// requiredSheets 非空但过滤后 needed 为空，也跳过
	if requiredSheets != nil {
		needed := make(map[string]bool)
		for sheet := range requiredSheets {
			if _, exists := sheetMap[sheet]; !exists {
				needed[sheet] = true
			}
		}

		if len(needed) == 0 {
			fmt.Printf("[缓存] 无需补充关联表，跳过\n")
			return
		}

		if cache != nil {
			count := supplementFromCache(sheetMap, cache, needed)
			fmt.Printf("[缓存] 从缓存按需补充了 %d 个关联表（共需 %d 个），耗时 %v\n", count, len(needed), time.Since(startTime))
			// 缓存补充不足时，降级到从 git commit 加载剩余缺失的表
			if count < len(needed) {
				remainingNeeded := make(map[string]bool)
				for sheet := range needed {
					if _, exists := sheetMap[sheet]; !exists {
						remainingNeeded[sheet] = true
					}
				}
				if len(remainingNeeded) > 0 {
					supplementFromCommit(sheetMap, repoPath, commitHash, remainingNeeded)
					fmt.Printf("[补充] 缓存不足，从 commit 降级加载剩余关联表，sheetMap %d→%d\n",
						beforeCount, len(sheetMap))
				}
			}
		} else {
			supplementFromCommit(sheetMap, repoPath, commitHash, needed)
			fmt.Printf("[补充] 从 commit 按需加载，sheetMap %d→%d，耗时 %v\n",
				beforeCount, len(sheetMap), time.Since(startTime))
		}
		return
	}

	// requiredSheets 为 nil：全量补充（兼容旧逻辑）
	if cache != nil {
		count := supplementFromCache(sheetMap, cache, nil)
		fmt.Printf("[缓存] 使用预加载缓存，补充 %d 个 sheet，耗时 %v\n", count, time.Since(startTime))
	} else {
		supplementFromCommit(sheetMap, repoPath, commitHash, nil)
		fmt.Printf("[补充] 从 commit 加载，sheetMap %d→%d，耗时 %v\n",
			beforeCount, len(sheetMap), time.Since(startTime))
	}
}

// supplementFromCache 从预加载的缓存补充 sheetMap 中缺失的表
// 如果指定了 requiredSheets，只补充声明需要的关联表（支持后缀匹配）；否则全量补充
// 后缀匹配规则：RequiredSheets 中的 "Item" 可匹配缓存中的 "道具|Item"
// 返回实际补充的数量
func supplementFromCache(sheetMap, cache map[string]*excelize.File, requiredSheets map[string]bool) int {
	supplementCount := 0
	for sheetName, f := range cache {
		if _, exists := sheetMap[sheetName]; !exists {
			// 如果指定了 requiredSheets，只补充需要的表（支持后缀匹配）
			if len(requiredSheets) > 0 && !helpers.MatchSheetBySuffix(sheetName, requiredSheets) {
				continue
			}
			sheetMap[sheetName] = f
			supplementCount++
		}
	}
	return supplementCount
}

// mayFileContainRequiredSheet 判断文件名是否可能包含需要的 sheet
// 用于在 git show 之前过滤掉明显不需要的文件，避免 I/O 开销
// 匹配策略：
//   - 精确匹配：fileName == req（如 "Item" == "Item"）
//   - 后缀匹配：req 的最后一段 == fileName（如 "SeasonPassReward" 的文件可能是 "SeasonPassReward_xxx.xlsx"）
//   - 前缀匹配：fileName 以 req 开头（如 "ArenaScore" 匹配 "ArenaScoreReward"）
//   - 包含匹配：req 包含 fileName 或 fileName 包含 req
func mayFileContainRequiredSheet(fileName string, requiredSheets map[string]bool) bool {
	for req := range requiredSheets {
		// 精确匹配
		if fileName == req {
			return true
		}
		// 文件名是 req 的前缀（如 "Arena" 匹配 "ArenaSeason"）
		if strings.HasPrefix(req, fileName) {
			return true
		}
		// req 是文件名的前缀（如 "SeasonPass" 匹配 "SeasonPassReward_xxx"）
		if strings.HasPrefix(fileName, req) {
			return true
		}
		// 文件名包含 req（如 "Drop" in "DropItem" 或 "Item" in "GiftItem"）
		if strings.Contains(fileName, req) || strings.Contains(req, fileName) {
			return true
		}
	}
	return false
}

// supplementFromCommit 从指定 commit 加载 xlsx 文件补充 sheetMap 中缺失的表
// 如果指定了 requiredSheets，只加载包含所需 sheet 的 xlsx 文件（按需加载）
// 已存在于 sheetMap 中的表不会被覆盖（commit 变更的文件优先级更高）
func supplementFromCommit(sheetMap map[string]*excelize.File, repoPath, commitHash string, requiredSheets map[string]bool) {
	// 获取该 commit 中所有 xlsx 文件列表
	xlsxFiles, err := gitutil.ListXlsxAtCommit(repoPath, commitHash)
	if err != nil {
		fmt.Printf("[警告] 获取 commit %s 文件列表失败，跳过补充: %v\n", commitHash[:8], err)
		return
	}

	// 加载 sheetMap 中缺失的文件
	supplementCount := 0
	loadedFiles := 0
	for _, relPath := range xlsxFiles {
		// 按需优化：如果指定了 requiredSheets，先判断文件名是否可能包含需要的 sheet
		// 匹配策略：requiredSheets 中的名称是文件名的子串（如 "Drop" in "Drop.xlsx"）
		// 或文件名是 requiredSheets 名称的子串（如 "Item.xlsx" contains "Item"）
		if len(requiredSheets) > 0 {
			fileName := strings.TrimSuffix(filepath.Base(relPath), ".xlsx")
			if !mayFileContainRequiredSheet(fileName, requiredSheets) {
				continue
			}
		}

		data, err := gitutil.GetFileAtCommit(repoPath, commitHash, filepath.Join(repoPath, relPath))
		if err != nil {
			continue
		}

		f, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			continue
		}
		f.Path = filepath.Join(repoPath, relPath)
		loadedFiles++

		// 检查该文件的 sheet 是否有 sheetMap 缺失的
		fileNeeded := false
		for _, sheetName := range f.GetSheetList() {
			if _, exists := sheetMap[sheetName]; !exists {
				// 如果指定了 requiredSheets，只补充需要的表（支持后缀匹配）
				if len(requiredSheets) > 0 && !helpers.MatchSheetBySuffix(sheetName, requiredSheets) {
					continue
				}
				sheetMap[sheetName] = f
				supplementCount++
				fileNeeded = true
			}
		}

		// 如果该文件的 sheet 都不需要，关闭文件释放资源
		if !fileNeeded {
			f.Close()
		}
	}

	if supplementCount > 0 {
		fmt.Printf("[补充] 从 commit %s 按需加载了 %d 个文件中的 %d 个缺失 sheet\n",
			commitHash[:8], loadedFiles, supplementCount)
	}
}

// checkGeneralTableRulesWithBaseCommit 执行通用表级规则检查，指定对比基准 commit 和当前 commit
// 与 checkGeneralTableRules 逻辑相同，但将 BASE_COMMIT 和 HEAD_COMMIT 设为指定值
func checkGeneralTableRulesWithBaseCommit(sheetMap map[string]*excelize.File, baseCommit, headCommit string) []*json_rule.TableCheckResult {
	// 清理上一次检查的缓存
	diff.ClearDiffCache()
	defer diff.ClearDiffCache()

	results := make([]*json_rule.TableCheckResult, 0)

	allMetas := json_rule.GetAllTableRuleMetas()

	// 筛选通用规则（TargetSheets 为空）
	var generalMetas []*json_rule.TableRuleMeta
	for _, meta := range allMetas {
		if len(meta.TargetSheets) == 0 {
			generalMetas = append(generalMetas, meta)
		}
	}

	if len(generalMetas) == 0 {
		return results
	}

	checkedCount := 0
	for sheetName, xlsxFile := range sheetMap {
		checkedCount++
		cols, err := xlsxFile.GetCols(sheetName)
		if err != nil {
			continue
		}

		for _, meta := range generalMetas {
			// 跳过已废弃的旧枚举（由拆分后的 5 个新规则替代）
			if meta.Type == json_rule.NEW_ROW_NOTIFY || meta.Type == json_rule.ROW_CHANGE_NOTIFY {
				continue
			}

			checker := TableManager.GetChecker(meta.Type)
			if checker == nil {
				continue
			}

			// 构建参数，显式设置 BASE_COMMIT 和 HEAD_COMMIT，并合并 per-sheet 覆盖
			params := meta.ResolveParams(nil)
			params[string(json_rule.BASE_COMMIT)] = baseCommit
			if headCommit != "" {
				params[string(json_rule.HEAD_COMMIT)] = headCommit
			}
			if sheetOverrides := json_rule.GetGeneralRuleOverrides(sheetName); len(sheetOverrides) > 0 {
				if ruleOverrides, ok := sheetOverrides[meta.Type]; ok {
					for k, v := range ruleOverrides {
						if k != string(json_rule.BASE_COMMIT) && k != string(json_rule.GIT_REPO_PATH) {
							params[k] = v
						}
					}
				}
			}

			result := checker.Check(json_rule.CheckParam{
				SheetName:   sheetName,
				Cols:        cols,
				StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
				Params:      params,
				SheetMap:    sheetMap,
			})
			if result == nil {
				continue
			}

			hasContent := result.Ok || len(result.ErrCells) > 0
			if hasContent {
				result.SheetName = &sheetName
				result.RuleType = meta.Type
				result.DisplayName = meta.DisplayName
				result.TableName = &xlsxFile.Path
				results = append(results, result)
			}
		}
	}

	fmt.Printf("[通用规则] 检查了 %d 个 sheet（只包含变更的 sheet）\n", checkedCount)

	return results
}

// 确保 excelize.File 在本包中被引用（checkGeneralTableRules 使用）
var _ *excelize.File
