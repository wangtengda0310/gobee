package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	feishu "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu/notification/handlers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/dispatch"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/gitutil"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/engine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/ruleconfig"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/workflow"
	"github.com/xuri/excelize/v2"
)

// 默认配置值
const (
	defaultExcelPath      = "../../config/excel"
	defaultExcelCheckRule = "cases/excel_cases"
	//defaultFeishuRobot    = "36732a0b-9b65-4456-8294-17044223114f" // ljh
	// https://open.feishu.cn/open-apis/bot/v2/hook/db06f82a-4dad-43f6-bbef-97503e0b953a
	defaultFeishuRobot = "db06f82a-4dad-43f6-bbef-97503e0b953a" // ljh
)

// maxMergeDisplayCommits merge 场景下最多展示详细检查结果的 commit 数量
// 所有 commit 都会执行检查，但只展示前 N 个的详细分段，其余在概览区域简要列出
const maxMergeDisplayCommits = 5

// Config 定义命令行参数结构体
type Config struct {
	feishuRobot       string // 飞书机器人 webhook ID
	excelPath         string // Excel 配表目录
	casePath          string // 检查规则保存目录
	mode              string // 检查模式: incremental(增量) | full(全量-本地磁盘)
	feishuDMAppID     string // 飞书私聊 App ID
	feishuDMAppSecret string // 飞书私聊 App Secret
	feishuDMDisabled  bool   // 是否禁用私聊通知
}

// argParse 解析命令行参数
func argParse() *Config {
	var config Config

	flag.StringVar(&config.feishuRobot, "feishuRobot", defaultFeishuRobot, "飞书机器人")
	flag.StringVar(&config.excelPath, "excelPath", defaultExcelPath, "excel文件夹目录")
	flag.StringVar(&config.casePath, "casePath", defaultExcelCheckRule, "检查规则保存目录")
	flag.StringVar(&config.mode, "mode", "incremental", "检查模式: incremental(增量), full(全量-本地磁盘)")
	flag.BoolVar(&config.feishuDMDisabled, "noDM", false, "禁用飞书私聊通知")

	flag.Parse()

	// 私聊凭证仅通过环境变量读取（不使用命令行参数，避免密钥出现在 ps aux 中）
	config.feishuDMAppID = os.Getenv("FEISHU_DM_APP_ID")
	config.feishuDMAppSecret = os.Getenv("FEISHU_DM_APP_SECRET")

	return &config
}

// ==================== 检查结果数据结构 ====================

// CommitCheckResult 单个 commit 的检查结果
type CommitCheckResult struct {
	CommitHash   string
	Author       string
	Email        string // 作者邮箱（用于飞书私聊）
	DiffFiles    []string
	ColResults   []*json_rule.ColCheckResult
	TableResults []*json_rule.TableCheckResult
	Stats        *engine.CheckStats
	MergeInfo    *notification.MergeInfo // merge 场景下的分支摘要（仅第一个 result 携带）
	Error        error
}

// ==================== 通知策略 ====================

// NotifyStrategy 通知合并策略
type NotifyStrategy int

const (
	// NotifyStrategySingle 将多个 commit 结果聚合为一个事件发送（默认）
	NotifyStrategySingle NotifyStrategy = iota
	// NotifyStrategyEach 每个 commit 单独发送通知（预留）
	NotifyStrategyEach
)

// dispatchResults 根据策略和分支分发检查结果通知
func dispatchResults(results []*CommitCheckResult, config *Config, strategy NotifyStrategy) {
	branch, _ := gitutil.GetCurrentBranch(config.excelPath)
	mode := dispatch.ResolveNotifyMode(branch, dispatch.DefaultRules(), dispatch.DefaultMode())
	fmt.Printf("分支: %s, 通知模式: group=%v dm=%v\n", branch, mode.Group, mode.DM)

	dmHandler := newDMHandler(config)
	// 装饰器：用 debug 监控包装私聊 handler
	// 只要私聊通道启用（dmHandler 非 nil），每次 Handle 调用后自动发送 debug 摘要
	// v0.0.8-pre-release 分支 dmHandler 为 nil → 装饰器不生效 → 不发 debug
	monitorClient := feishu.NewOpenAPIClient(config.feishuDMAppID, config.feishuDMAppSecret)
	dmHandler = handlers.WrapDMMonitor(dmHandler, monitorClient)
	router := dispatch.NewNotifyRouter(mode, config.feishuRobot, dmHandler)

	switch strategy {
	case NotifyStrategySingle:
		dispatchAsSingle(results, config, router)
	case NotifyStrategyEach:
		for _, r := range results {
			dispatchOne(r, config, router)
		}
	}
}

// dispatchAsSingle 将所有 commit 的结果聚合为一个通知事件发送
// 非 merge 场景：直接聚合为一个事件
// merge 场景：先发送 merge 操作信息段落，再按 commit 分段发送检查结果（每段 @ 对应提交者）
func dispatchAsSingle(results []*CommitCheckResult, config *Config, router *dispatch.NotifyRouter) {
	// 检查是否有 merge 信息
	mergeInfo := getMergeInfo(results)

	if mergeInfo != nil {
		dispatchMergeResults(results, config, mergeInfo, router)
		return
	}

	// 非 merge 场景：聚合为一个事件
	dispatcher := router.BuildDispatcher([]string{getFirstAuthor(results)})

	var allColResults []*json_rule.ColCheckResult
	var allTableResults []*json_rule.TableCheckResult
	var allChangedFiles []string
	totalRules := 0
	filteredRules := 0
	generalRuleCount := 0

	for _, r := range results {
		if r.Error != nil {
			continue
		}
		allColResults = append(allColResults, r.ColResults...)
		allTableResults = append(allTableResults, r.TableResults...)
		allChangedFiles = mergeUnique(allChangedFiles, r.DiffFiles)
		if r.Stats != nil {
			totalRules += r.Stats.TotalRules
			filteredRules += r.Stats.FilteredRules
			generalRuleCount += r.Stats.GeneralRuleCount
		}
	}

	branch, _ := gitutil.GetCurrentBranch(config.excelPath)
	headHash, _ := gitutil.GetHeadHash(config.excelPath)
	author := getFirstAuthor(results)

	// 使用程序运行时间
	checkTime := time.Now()

	event := &notification.CheckResultEvent{
		ColResults:    allColResults,
		TableResults:  allTableResults,
		CheckTime:     checkTime,
		IsIncremental: len(allChangedFiles) > 0,
		ChangedFiles:  allChangedFiles,
		CommitInfo: &notification.CommitInfo{
			Name:    author,
			Email:   getFirstEmail(results),
			Branch:  branch,
			Version: headHash,
			SkipDM:  author == "excel配置全量检查",
		},
		Stats: &notification.CheckStats{
			TotalRules:       totalRules,
			FilteredRules:    filteredRules,
			ChangedFileCount: len(allChangedFiles),
			GeneralRuleCount: generalRuleCount,
		},
	}

	dispatcher.Dispatch(event)
}

// dispatchMergeResults merge 场景下汇聚为一条通知消息发送
// 包含 merge 概览 + 各 commit 分段检查结果，每个段落 @ 对应提交者
func dispatchMergeResults(results []*CommitCheckResult, config *Config, mergeInfo *notification.MergeInfo, router *dispatch.NotifyRouter) {
	branch, _ := gitutil.GetCurrentBranch(config.excelPath)

	// 聚合所有 commit 的结果（用于 GetSummary 统计）
	var allColResults []*json_rule.ColCheckResult
	var allTableResults []*json_rule.TableCheckResult
	var allChangedFiles []string
	totalRules := 0
	filteredRules := 0
	generalRuleCount := 0

	// 构建 commit 分段数据（只展示前 maxMergeDisplayCommits 个的详细结果）
	var commitSections []notification.CommitSection
	// 收集所有需要 @ 的用户（去重）
	authorSet := map[string]bool{mergeInfo.MergeAuthor: true}
	var allAuthors []string
	allAuthors = append(allAuthors, mergeInfo.MergeAuthor)

	for _, r := range results {
		if r.Error != nil {
			continue
		}

		// 聚合到顶层（用于统计）
		allColResults = append(allColResults, r.ColResults...)
		allTableResults = append(allTableResults, r.TableResults...)
		allChangedFiles = mergeUnique(allChangedFiles, r.DiffFiles)
		if r.Stats != nil {
			totalRules += r.Stats.TotalRules
			filteredRules += r.Stats.FilteredRules
			generalRuleCount += r.Stats.GeneralRuleCount
		}

		// 构建分段（只展示前 maxMergeDisplayCommits 个）
		if len(commitSections) < maxMergeDisplayCommits {
			section := notification.CommitSection{
				Author:       r.Author,
				AuthorEmail:  r.Email,
				CommitHash:   r.CommitHash,
				Branch:       branch,
				DiffFiles:    r.DiffFiles,
				ColResults:   r.ColResults,
				TableResults: r.TableResults,
			}
			if r.Stats != nil {
				section.Stats = &notification.CheckStats{
					TotalRules:       r.Stats.TotalRules,
					FilteredRules:    r.Stats.FilteredRules,
					ChangedFileCount: r.Stats.ChangedFileCount,
					GeneralRuleCount: r.Stats.GeneralRuleCount,
				}
			}
			commitSections = append(commitSections, section)
		}

		// 去重收集作者
		if r.Author != "" && !authorSet[r.Author] {
			authorSet[r.Author] = true
			allAuthors = append(allAuthors, r.Author)
		}
	}

	// 使用程序运行时间
	checkTime := time.Now()

	// 构建单个事件
	event := &notification.CheckResultEvent{
		ColResults:    allColResults,
		TableResults:  allTableResults,
		CheckTime:     checkTime,
		IsIncremental: len(allChangedFiles) > 0,
		ChangedFiles:  allChangedFiles,
		CommitInfo: &notification.CommitInfo{
			Name:      mergeInfo.MergeAuthor,
			Branch:    branch,
			MergeInfo: mergeInfo,
		},
		Stats: &notification.CheckStats{
			TotalRules:       totalRules,
			FilteredRules:    filteredRules,
			ChangedFileCount: len(allChangedFiles),
			GeneralRuleCount: generalRuleCount,
		},
		CommitSections: commitSections,
	}

	// 只发送一条消息
	dispatcher := router.BuildDispatcher(allAuthors)
	dispatcher.Dispatch(event)
}

// dispatchOne 发送单个 commit 的检查结果通知（格式与重构前一致）
func dispatchOne(r *CommitCheckResult, config *Config, router *dispatch.NotifyRouter) {
	if r.Error != nil {
		return
	}

	branch, _ := gitutil.GetCurrentBranch(config.excelPath)

	dispatcher := router.BuildDispatcher([]string{r.Author})

	// 使用程序运行时间
	checkTime := time.Now()

	event := &notification.CheckResultEvent{
		ColResults:    r.ColResults,
		TableResults:  r.TableResults,
		CheckTime:     checkTime,
		IsIncremental: r.Stats != nil && r.Stats.IsIncremental,
		ChangedFiles:  r.DiffFiles,
		CommitInfo: &notification.CommitInfo{
			Name:    r.Author,
			Email:   r.Email,
			Branch:  branch,
			Version: r.CommitHash,
		},
	}

	if r.Stats != nil {
		event.Stats = &notification.CheckStats{
			TotalRules:       r.Stats.TotalRules,
			FilteredRules:    r.Stats.FilteredRules,
			ChangedFileCount: r.Stats.ChangedFileCount,
			GeneralRuleCount: r.Stats.GeneralRuleCount,
		}
	}

	dispatcher.Dispatch(event)
}

// ==================== 工具函数 ====================

// getFirstAuthor 获取第一个有作者的 commit 的作者名
func getFirstAuthor(results []*CommitCheckResult) string {
	for _, r := range results {
		if r.Author != "" {
			return r.Author
		}
	}
	return ""
}

// getFirstEmail 获取第一个有邮箱的 commit 的邮箱
func getFirstEmail(results []*CommitCheckResult) string {
	for _, r := range results {
		if r.Email != "" {
			return r.Email
		}
	}
	return ""
}

// newDMHandler 创建飞书私聊处理器
// 有凭证时始终创建 handler：-noDM 时设置 dryRun 跳过实际发送，但 Handle 仍执行（允许装饰器触发 debug）
// 无凭证时返回 nil
func newDMHandler(config *Config) notification.CheckResultHandler {
	if config.feishuDMAppID == "" || config.feishuDMAppSecret == "" {
		return nil
	}
	h := handlers.NewFeishuDMHandler(
		feishu.NewOpenAPIClient(config.feishuDMAppID, config.feishuDMAppSecret),
	)
	h.SetDryRun(config.feishuDMDisabled)
	return h
}

// getMergeInfo 从结果列表中获取 MergeInfo（由第一个 result 携带）
func getMergeInfo(results []*CommitCheckResult) *notification.MergeInfo {
	for _, r := range results {
		if r.MergeInfo != nil {
			return r.MergeInfo
		}
	}
	return nil
}

// mergeUnique 合并两个字符串切片并去重
func mergeUnique(a, b []string) []string {
	seen := make(map[string]bool)
	for _, s := range a {
		seen[s] = true
	}
	result := make([]string, len(a))
	copy(result, a)
	for _, s := range b {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	return result
}

// ==================== 主入口 ====================

//go:generate go run main.go -feishuRobot=none
func main() {

	// 5步流程说明（三种检查模式对比）：
	//
	//   步骤1: 确定检查范围
	//     全量模式(-mode=full): 全部 Excel 文件，无需 git
	//     普通 commit: HEAD 单个 commit 的变更文件
	//     Merge commit: 遍历两个分支的所有子 commit
	//
	//   步骤2: 加载校验规则（LoadJsonRules + supplementDefaultTableRules + supplementDefaultParams）
	//     全量模式: 在 CheckWithFilter 内部完成
	//     普通 commit: 在 CheckWithGitHistory 内部完成
	//     Merge commit: 外部一次加载，deepCopyRules 后各 commit 共享
	//
	//   步骤3: 过滤适用规则（FilterRulesByChangedFiles）
	//     全量模式: changedFiles=nil → 返回全部规则
	//     普通/Merge commit: 按 git diff 文件过滤
	//
	//   步骤4: 加载数据到 sheetMap
	//     全量模式: GetSheetMap 从本地磁盘全量加载
	//     普通 commit: CheckWithGitHistory 通过 git show 按 commit 版本加载
	//     Merge commit: 同上 + 跨 commit 缓存传递（fallbackSheetMap）
	//
	//   步骤5: 执行校验（CheckSheetCols + CheckTableRules + 通用规则）
	//     全量模式: 执行全部规则，跳过通用规则（无 diff 基准）
	//     普通/Merge commit: 执行过滤后的规则 + 通用规则（NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY）

	config := argParse()

	fmt.Printf("ExcelPath: %s\n", config.excelPath)
	fmt.Printf("CasePath: %s\n", config.casePath)

	var results []*CommitCheckResult

	switch config.mode {
	case "full":
		// 全量检查模式：直接读取本地磁盘文件，不依赖 git 历史
		handleFullCheck(config, &results)
	default:
		// 增量检查模式：通过 git 历史获取变更文件
		headHash, err := gitutil.GetHeadHash(config.excelPath)
		if err != nil {
			log.Fatalf("获取 HEAD hash 失败: %v", err)
		}
		fmt.Printf("HEAD: %s\n", headHash)

		isMerge, err := gitutil.IsMergeCommit(config.excelPath, headHash)
		if err != nil {
			log.Fatalf("判断 merge commit 失败: %v", err)
		}

		if isMerge {
			fmt.Printf("检测到 merge commit: %s\n", headHash[:8])
			handleMergeCommit(config, headHash, &results)
		} else {
			handleNormalCommit(config, headHash, &results)
		}
	}

	// 统一分发通知
	if len(results) > 0 {
		dispatchResults(results, config, NotifyStrategySingle)
	}
}

// handleMergeCommit 处理 merge commit：遍历子 commit 执行配表检查
// 遍历顺序保持分支分组（先主分支 parent1，后被合并分支 parent2），各分支内部按时间正序
// 不读取本地磁盘文件，每个 commit 通过 git show 获取历史版本
//
// 5步流程分布：
//
//	步骤1（本函数）: 获取子 commit 列表 → 遍历
//	步骤2（本函数）: 外部一次 LoadJsonRules + deepCopyRules 共享
//	步骤3-5（CheckWithGitHistory 内部）: 过滤规则 → supplementSheetMap → 执行校验
func handleMergeCommit(config *Config, headHash string, results *[]*CommitCheckResult) {
	// ── 步骤1: 获取按时间排序的子 commit 列表（区分 merge/普通） ──
	mergeCommits, err := gitutil.GetMergeChildCommits(config.excelPath, headHash)
	if err != nil {
		log.Printf("获取 merge 子提交失败: %v，回退为普通 commit 处理", err)
		handleNormalCommit(config, headHash, results)
		return
	}

	fmt.Printf("主分支 commit 数: %d, 被合并分支 commit 数: %d\n",
		len(mergeCommits.Parent1Commits), len(mergeCommits.Parent2Commits))

	repoRoot, err := gitutil.GetRepoRoot(config.excelPath)
	if err != nil {
		log.Printf("获取仓库根目录失败: %v", err)
		return
	}

	// ── 步骤2: 加载所有校验规则 ──
	// 在遍历 commit 之前一次性加载规则列表，所有 commit 共享同一份规则
	allRules, err := ruleconfig.LoadJsonRules(config.casePath)
	if err != nil {
		log.Printf("加载校验规则失败: %v", err)
		return
	}
	fmt.Printf("加载规则数: %d\n", len(allRules))

	// 构建 author 缓存（从 mergeCommits.SortedCommits 提取，避免循环中逐个调用 git）
	authorCache := make(map[string]string, len(mergeCommits.SortedCommits))
	emailCache := make(map[string]string, len(mergeCommits.SortedCommits))
	for _, c := range mergeCommits.SortedCommits {
		authorCache[c.Hash] = c.Author
		emailCache[c.Hash] = c.Email
	}

	// 初始化跨 commit 的关联表缓存，前一个 commit 加载的关联表供后续 commit 复用
	fallbackSheetMap := make(map[string]*excelize.File)
	// 遍历结束后清理缓存中的 excelize.File，释放内存
	defer func() {
		for _, f := range fallbackSheetMap {
			f.Close()
		}
	}()

	// 遍历顺序：保持分支分组（先 parent1 后 parent2），各分支内部按时间正序
	allCommits := append(mergeCommits.Parent1Commits, mergeCommits.Parent2Commits...)
	// 记录 parent2 分支的起始索引，用于跨分支时清空关联表缓存
	// 缓存版本安全说明：
	//   - 同分支内缓存传递是安全的：版本沿提交链自然演进（commit1.parent=base，commit2.parent=commit1）
	//   - 跨分支时不安全：parent2 分支的 parentHash 指向 parent2 的历史，而非 parent1
	//     fallbackSheetMap 中的关联表来自 parent1 分支的 parentHash，版本不匹配
	//   - 解决方案：在跨分支边界清空缓存，parent2 第一个 commit 从 git 重新加载（约 +1.5 秒）
	parent2StartIdx := len(mergeCommits.Parent1Commits)

	mergeStart := time.Now()
	for i, commitHash := range allCommits {
		// 跨分支边界：清空关联表缓存，避免 parent1 分支版本的关联表污染 parent2 分支
		if i == parent2StartIdx && i > 0 {
			for _, f := range fallbackSheetMap {
				f.Close()
			}
			fallbackSheetMap = make(map[string]*excelize.File)
			fmt.Println("[缓存] 跨分支边界，清空关联表缓存，parent2 分支将从 git 重新加载")
		}

		author := authorCache[commitHash]

		// 获取该 commit 的父 commit 作为对比基准
		parentHash, err := gitutil.GetParentCommit(repoRoot, commitHash)
		if err != nil {
			log.Printf("获取 commit %s 父提交失败: %v，跳过\n", commitHash[:8], err)
			continue
		}

		fmt.Printf("\n=== [git] 检查 commit %s (作者: %s, 对比: %s) ===\n", commitHash[:8], author, parentHash[:8])

		// ── 步骤3-5 在 CheckWithGitHistory 内部执行 ──
		// 步骤3: 获取变更文件 → 过滤适用规则（git 历史场景不过滤）
		//   补充: supplementDefaultTableRules + supplementDefaultParams
		// 步骤4: collectRequiredSheets → supplementSheetMap（优先从 fallbackSheetMap 缓存取，跨分支时已清空）
		// 步骤5: CheckSheetCols + CheckTableRules + 通用规则 + CheckStats 统计
		var commitSheetMap map[string]*excelize.File
		colRes, tableRes, stats, err := engine.CheckWithGitHistory(
			repoRoot, "", config.casePath, commitHash, parentHash,
			engine.WithFallbackSheetMap(fallbackSheetMap),
			engine.WithPreloadedRules(allRules),
			engine.WithSheetMapOutput(&commitSheetMap),
		)
		if err != nil {
			log.Printf("检查 commit %s 失败: %v\n", commitHash[:8], err)
			continue
		}

		// 将本次 commit 的 sheetMap 合并到跨 commit 缓存（新版本覆盖旧版本，旧版本 Close 释放内存）
		// 注意：同一个 *excelize.File 可能被多个 sheet key 引用（同一个 xlsx 文件的多个 sheet）
		// 需要去重 Close：跳过 oldFile == f 的缓存传递场景，避免关闭仍在使用中的 File
		for name, f := range commitSheetMap {
			if oldFile, exists := fallbackSheetMap[name]; exists && oldFile != f {
				oldFile.Close()
			}
			fallbackSheetMap[name] = f
		}

		diffFiles, _ := gitutil.GetCommitDiffFiles(repoRoot, commitHash)

		result := &CommitCheckResult{
			CommitHash:   commitHash,
			Author:       author,
			Email:        emailCache[commitHash],
			DiffFiles:    diffFiles,
			ColResults:   colRes,
			TableResults: tableRes,
			Stats:        stats,
		}

		// 为第一个 commit 携带 merge 信息
		if len(*results) == 0 {
			mergeAuthor, _ := gitutil.GetCommitAuthor(config.excelPath, headHash)
			branch, _ := gitutil.GetCurrentBranch(config.excelPath)
			result.MergeInfo = &notification.MergeInfo{
				MergeAuthor:    mergeAuthor,
				ToBranch:       branch,
				Parent1Commits: buildCommitSummaries(config.excelPath, mergeCommits.Parent1Commits),
				Parent1Total:   len(mergeCommits.Parent1Commits),
				Parent2Commits: buildCommitSummaries(config.excelPath, mergeCommits.Parent2Commits),
				Parent2Total:   len(mergeCommits.Parent2Commits),
			}
		}

		*results = append(*results, result)
	}

	fmt.Printf("[耗时] merge 遍历 %d 个 commit 总耗时: %v\n", len(allCommits), time.Since(mergeStart))
}

// maxDisplayCommits 飞书消息中每个分支最多显示的 commit 数量（暂不支持命令行参数）
const maxDisplayCommits = 5

// buildCommitSummaries 构建分支的 commit 摘要列表，最多保留前 maxDisplayCommits 个
func buildCommitSummaries(repoPath string, commits []string) []notification.CommitSummary {
	max := maxDisplayCommits
	if len(commits) < max {
		max = len(commits)
	}
	result := make([]notification.CommitSummary, max)
	for i := 0; i < max; i++ {
		hash := commits[i]
		if len(hash) >= 8 {
			hash = hash[:8]
		}
		author, _ := gitutil.GetCommitAuthor(repoPath, commits[i])
		result[i] = notification.CommitSummary{
			Hash:   hash,
			Author: author,
		}
	}
	return result
}

// handleNormalCommit 处理普通 commit：通过 git show 获取 HEAD 历史版本检查
func handleNormalCommit(config *Config, headHash string, results *[]*CommitCheckResult) {
	fmt.Printf("普通 commit: %s\n", headHash[:8])
	repoRoot, err := gitutil.GetRepoRoot(config.excelPath)
	if err != nil {
		log.Printf("获取仓库根目录失败: %v", err)
		return
	}

	// 获取 commit 作者和父 commit
	author, err := gitutil.GetCommitAuthor(repoRoot, headHash)
	if err != nil {
		log.Printf("获取 commit %s 作者失败: %v", headHash[:8], err)
	}
	email, _ := gitutil.GetCommitAuthorEmail(repoRoot, headHash)
	parentHash, err := gitutil.GetParentCommit(repoRoot, headHash)
	if err != nil {
		log.Printf("获取 commit %s 父提交失败: %v", headHash[:8], err)
	}

	// 使用统一工作流执行增量检查（git 历史模式）
	wfResult, err := workflow.RunCheckWorkflow(&workflow.CheckWorkflowConfig{
		ExcelPath:  config.excelPath,
		CasePath:   config.casePath,
		Mode:       workflow.CheckModeIncremental,
		RepoPath:   repoRoot,
		CommitHash: headHash,
		ParentHash: parentHash,
	})
	if err != nil {
		log.Printf("检查 commit %s 失败: %v", headHash[:8], err)
		return
	}

	diffFiles, _ := gitutil.GetCommitDiffFiles(repoRoot, headHash)

	// 映射 stats
	var stats *engine.CheckStats
	if wfResult.Stats != nil {
		stats = &engine.CheckStats{
			TotalRules:       wfResult.Stats.TotalRules,
			FilteredRules:    wfResult.Stats.FilteredRules,
			GeneralRuleCount: wfResult.Stats.GeneralRuleCount,
			IsIncremental:    wfResult.Stats.IsIncremental,
			ChangedFileCount: wfResult.Stats.ChangedFileCount,
		}
	}

	result := &CommitCheckResult{
		CommitHash:   headHash,
		Author:       author,
		Email:        email,
		DiffFiles:    diffFiles,
		ColResults:   wfResult.ColResults,
		TableResults: wfResult.TableResults,
		Stats:        stats,
	}
	*results = append(*results, result)
}

// handleFullCheck 全量检查本地所有 Excel 文件
// 不依赖 git 历史，直接读取本地磁盘文件
// 注意：通用规则（NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY）在全量模式下跳过（无 diff 基准）
func handleFullCheck(config *Config, results *[]*CommitCheckResult) {
	start := time.Now()
	fmt.Println("=== 全量检查模式（本地磁盘）===")

	// 使用统一工作流执行全量检查
	wfResult, err := workflow.RunCheckWorkflow(&workflow.CheckWorkflowConfig{
		ExcelPath: config.excelPath,
		CasePath:  config.casePath,
		Mode:      workflow.CheckModeFull,
	})
	if err != nil {
		log.Fatalf("全量检查失败: %v", err)
	}

	// 映射 stats
	var stats *engine.CheckStats
	if wfResult.Stats != nil {
		stats = &engine.CheckStats{
			TotalRules:       wfResult.Stats.TotalRules,
			FilteredRules:    wfResult.Stats.FilteredRules,
			GeneralRuleCount: wfResult.Stats.GeneralRuleCount,
			IsIncremental:    wfResult.Stats.IsIncremental,
			ChangedFileCount: wfResult.Stats.ChangedFileCount,
		}
	}

	result := &CommitCheckResult{
		CommitHash:   "FULL_CHECK",
		Author:       "excel配置全量检查",
		ColResults:   wfResult.ColResults,
		TableResults: wfResult.TableResults,
		Stats:        stats,
	}
	*results = append(*results, result)

	fmt.Printf("[耗时] 全量检查: %v, 配置规则: %d, 通用规则: %d\n",
		time.Since(start), stats.TotalRules, stats.GeneralRuleCount)
}
