// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的左链步骤处理器（LeftStepHandler）
// 左链前向执行：在调用 next 之前完成本步的查找/提取，结果写入 ChainContext 供内层使用
package chain_reference

import (
	"fmt"
	"regexp"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
)

// LeftStepHandler 左链步骤处理器
// 左链前向执行：在调用 next 之前完成本步的查找/提取
// 结果写入 ctx.LeftCurrentValues 和 ctx.LeftStepValues 供内层 handler 使用
//
// 执行模式：
//   - StepIdx == 0 && Sheet == ""（仅取值模式）：从当前表取值，不跨表查找
//   - StepIdx > 0 或 Sheet != ""（跨表查找模式）：在目标表中查找匹配行并提取值
type LeftStepHandler struct {
	Step     ChainStep   // 当前步骤配置
	StepIdx  int         // 步骤索引（0-based）
	IsLast   bool        // 是否左链最后一步
	ChainCfg ChainConfig // 链配置（用于最后一步的 CompareCol）
}

// Handle 执行左链步骤的前向查找逻辑
//
// 执行流程：
//  1. 根据步骤类型选择执行模式（仅取值 / 跨表查找）
//  2. 提取值后写入 ctx.LeftCurrentValues 和 ctx.LeftStepValues
//  3. 如果是最后一步，构建 ctx.LeftResult
//  4. 调用 next 传递给内层
func (h *LeftStepHandler) Handle(ctx *ChainContext, next NextFunc) error {
	// 根据步骤类型选择执行模式
	if h.StepIdx == 0 && h.Step.Sheet == "" {
		h.handleValueOnlyMode(ctx)
	} else {
		h.handleCrossTableLookup(ctx)
	}

	// 如果当前步骤产生错误，直接返回
	if ctx.Err != nil {
		return ctx.Err
	}

	// 前向执行完成，传递给内层
	return next(ctx)
}

// handleValueOnlyMode 左链第一步仅取值模式
// 从当前表的指定列取值，不跨表查找
// 如果有 Pattern，使用正则提取子值
//
// 执行流程：
//  1. 从 ctx.Cols 中找到 NextCol 列（NextCol 为空时用 ctx.ColIdx）
//  2. 取 ctx.RowIdx 行的值
//  3. 如果有 Pattern，正则提取子值
//  4. 结果写入 ctx.LeftCurrentValues 和 ctx.LeftStepValues
//  5. 如果 IsLast，构建 ctx.LeftResult
func (h *LeftStepHandler) handleValueOnlyMode(ctx *ChainContext) {
	// 步骤 1: 从指定列取值
	var val string
	if h.Step.NextCol != "" {
		nextColIdx := helpers.GetColIndexByName(ctx.Cols, h.Step.NextCol)
		if nextColIdx < 0 {
			ctx.Err = fmt.Errorf("步骤%d: 未找到列: %s", h.StepIdx+1, h.Step.NextCol)
			return
		}
		val = helpers.GetColValue(ctx.Cols, nextColIdx, ctx.RowIdx)
	} else {
		if ctx.ColIdx < 0 || ctx.ColIdx >= len(ctx.Cols) {
			ctx.Err = fmt.Errorf("步骤%d: 当前列索引无效: %d", h.StepIdx+1, ctx.ColIdx)
			return
		}
		val = helpers.GetColValue(ctx.Cols, ctx.ColIdx, ctx.RowIdx)
	}

	// 空值直接返回，LeftCurrentValues 保持为 nil
	if val == "" {
		return
	}

	inputValues := []string{val}

	// 步骤 2: 正则提取（如果有 Pattern）
	var nextValues []string
	if h.Step.Pattern != "" {
		re, err := regexp.Compile(h.Step.Pattern)
		if err != nil {
			ctx.Err = fmt.Errorf("步骤%d: 正则编译失败: %w", h.StepIdx+1, err)
			return
		}
		groups := helpers.ParseCaptureGroups(h.Step.Groups)
		nextValues = applyIsArray(helpers.ExtractValuesByRegex(inputValues[0], re, groups), h.Step.IsArray)
	} else {
		nextValues = applyIsArray(inputValues, h.Step.IsArray)
	}

	// 步骤 3: 写入中间结果
	ctx.LeftStepValues = append(ctx.LeftStepValues, append([]string{}, nextValues...))
	ctx.LeftCurrentValues = nextValues
	ctx.FirstStepInputValues = nextValues

	// 步骤 4: 如果是唯一步骤，构建 LeftResult
	if h.IsLast {
		h.buildLeftResult(ctx, nil, nil, nextValues)
	}
}

// handleCrossTableLookup 左链跨表查找模式
// 在目标表中匹配 ctx.LeftCurrentValues 的值并提取新值
//
// 执行流程：
//  1. 打开目标表
//  2. 获取列数据和列索引
//  3. 获取过滤条件行
//  4. 遍历目标表，在 PreCol 列匹配 ctx.LeftCurrentValues
//  5. 提取 NextCol 列值
//  6. 写入 ctx.LeftCurrentValues 和 ctx.LeftStepValues
//  7. 如果 IsLast，构建 ctx.LeftResult
func (h *LeftStepHandler) handleCrossTableLookup(ctx *ChainContext) {
	if len(ctx.LeftCurrentValues) == 0 && !(h.StepIdx == 0) {
		// 无输入值且不是全表扫描模式，直接返回
		return
	}

	// 步骤 1: 打开目标表
	targetFile, targetSheetName, found := helpers.FindSheetBySuffix(ctx.SheetMap, h.Step.Sheet)
	if !found {
		ctx.Err = fmt.Errorf("步骤%d: 目标表不存在: %s", h.StepIdx+1, h.Step.Sheet)
		return
	}

	targetCols, err := ctx.GetCachedCols(targetSheetName, targetFile)
	if err != nil {
		ctx.Err = fmt.Errorf("步骤%d: 读取目标表失败: %w", h.StepIdx+1, err)
		return
	}

	// 步骤 2: 查找列索引
	preColIdx := helpers.GetColIndexByName(targetCols, h.Step.PreCol)
	if preColIdx < 0 {
		ctx.Err = fmt.Errorf("步骤%d: 目标表中未找到查找列: %s", h.StepIdx+1, h.Step.PreCol)
		return
	}

	nextColIdx := -1
	if h.Step.NextCol != "" {
		nextColIdx = helpers.GetColIndexByName(targetCols, h.Step.NextCol)
		if nextColIdx < 0 {
			ctx.Err = fmt.Errorf("步骤%d: 目标表中未找到提取列: %s", h.StepIdx+1, h.Step.NextCol)
			return
		}
	}

	// 预编译正则（如果有）
	var stepPattern *regexp.Regexp
	var stepGroups []int
	if h.Step.Pattern != "" {
		re, err := regexp.Compile(h.Step.Pattern)
		if err != nil {
			ctx.Err = fmt.Errorf("步骤%d: 正则编译失败: %w", h.StepIdx+1, err)
			return
		}
		stepPattern = re
		stepGroups = helpers.ParseCaptureGroups(h.Step.Groups)
	}

	// 步骤 3: 获取过滤条件
	targetStartRow := excelio.MJS_FIXED_ROWS_NUM
	filteredRows, ferr := FilterRowsByConditionEx(targetCols, FilterOptions{
		FilterColName: h.Step.FilterCol,
		FilterVal:     h.Step.FilterVal,
		StartRowIdx:   targetStartRow,
		FilterIsArray: h.Step.FilterIsArray,
		FilterMode:    h.Step.FilterMode,
		FilterDays:    h.Step.FilterDays,
	})
	if ferr != nil {
		ctx.Err = fmt.Errorf("左链步骤%d: 过滤条件错误: %w", h.StepIdx+1, ferr)
		return
	}
	hasFilter := h.Step.FilterCol != "" && (h.Step.FilterVal != "" || h.Step.FilterMode == "withinDays")

	// 将过滤行列表构建为 map（O(n) → O(1) 查找）
	filterRowSet := make(map[int]bool, len(filteredRows))
	for _, fr := range filteredRows {
		filterRowSet[fr] = true
	}

	targetEndIdx := helpers.GetDataEndIndex(targetCols, targetStartRow)

	// 步骤 4: 遍历目标表查找匹配行
	var nextValues []string
	var nextMatchRows []int

	// 全表扫描模式：左链第一步 + 有 sheet + 无输入值
	// 不需要从当前表取值，直接遍历目标表所有行
	fullTableScan := h.StepIdx == 0 && len(ctx.LeftCurrentValues) == 0

	if fullTableScan {
		// 全表扫描：遍历目标表所有行（或过滤后的行）
		for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
			if hasFilter && !filterRowSet[rowI] {
				continue
			}
			// 收集 preCol 值作为 FirstStepInputValues
			findVal := helpers.GetColValue(targetCols, preColIdx, rowI)
			if findVal == "" {
				continue
			}
			ctx.FirstStepInputValues = append(ctx.FirstStepInputValues, findVal)

			// 提取 nextCol 值传给下一步
			if nextColIdx >= 0 {
				extractedVal := helpers.GetColValue(targetCols, nextColIdx, rowI)
				if extractedVal != "" {
					nextValues = append(nextValues, extractedVal)
					if h.IsLast {
						nextMatchRows = append(nextMatchRows, rowI)
					}
				}
			} else {
				nextValues = append(nextValues, findVal)
				if h.IsLast {
					nextMatchRows = append(nextMatchRows, rowI)
				}
			}
		}
	} else {
		// 原有逻辑：对每个输入值执行查找和提取
		for _, inputVal := range ctx.LeftCurrentValues {
			// 如果有 Pattern，先提取子值
			var searchValues []string
			if stepPattern != nil {
				extracted := helpers.ExtractValuesByRegex(inputVal, stepPattern, stepGroups)
				if len(extracted) == 0 {
					continue
				}
				searchValues = applyIsArray(extracted, h.Step.IsArray)
			} else {
				searchValues = applyIsArray([]string{inputVal}, h.Step.IsArray)
			}

			// 构建查找值的 set
			searchSet := make(map[string]bool)
			for _, sv := range searchValues {
				searchSet[strings.TrimSpace(sv)] = true
			}

			// 遍历目标表查找匹配行
			for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
				if hasFilter && !filterRowSet[rowI] {
					continue
				}

				cellVal := helpers.GetColValue(targetCols, preColIdx, rowI)
				if !searchSet[cellVal] {
					continue
				}

				// 匹配成功，提取值
				if nextColIdx >= 0 {
					extractedVal := helpers.GetColValue(targetCols, nextColIdx, rowI)
					if extractedVal != "" {
						nextValues = append(nextValues, extractedVal)
						if h.IsLast {
							nextMatchRows = append(nextMatchRows, rowI)
						}
					}
				} else {
					// 没有提取列，使用匹配到的值
					nextValues = append(nextValues, cellVal)
					if h.IsLast {
						nextMatchRows = append(nextMatchRows, rowI)
					}
				}
			}
		}

	} // end else

	// 步骤 4.5: 对提取的 nextCol 值应用 isArray 拆分
	// 例如 DropGroup="91001,91003,91004" → ["91001","91003","91004"]
	nextValues = applyIsArray(nextValues, h.Step.IsArray)
	// 引用完整性检查：无过滤条件时，有输入值但查找结果为空报错
	// 有过滤条件时，查找为空可能是正常的类型不匹配（如道具存在但 Type 不是 Upgrade），不报错
	// 无过滤条件时，查找为空意味着引用的 ID 不存在，应报错
	if !hasFilter && len(ctx.LeftCurrentValues) > 0 && len(nextValues) == 0 {
		colVal := ctx.CurrentCellValue()
		colName := ""
		if ctx.ColIdx >= 0 && ctx.ColIdx < len(ctx.Cols) && len(ctx.Cols[ctx.ColIdx]) > excelio.MJS_FIXED_ROWS_NAME {
			colName = ctx.Cols[ctx.ColIdx][excelio.MJS_FIXED_ROWS_NAME]
		}
		// 构建链路路径：左链当前步骤到末尾 → 右链反向（末尾到第一步）
		path := buildChainPath(h.StepIdx, ctx.Config)
		if len(ctx.Config.Right.Steps) > 0 {
			rs := ctx.Config.Right.Steps[0]
			ctx.Err = fmt.Errorf("%s=\"%s\" 经过 %s 在 %s 列中未找到",
				colName, colVal, path, rs.PreCol)
		} else {
			ctx.Err = fmt.Errorf("%s=\"%s\" 经过 %s 中未找到",
				colName, colVal, path)
		}
		return
	}
	// 步骤 5: 写入中间结果
	ctx.LeftStepValues = append(ctx.LeftStepValues, append([]string{}, nextValues...))
	ctx.LeftCurrentValues = nextValues

	// 步骤 5.5: 预警时间提取（每一步都检查，不仅限于最后一步）
	h.extractWarnValues(ctx, targetCols, nextMatchRows)

	// 步骤 6: 如果是最后一步，构建 LeftResult
	if h.IsLast {
		h.buildLeftResult(ctx, targetCols, nextMatchRows, nextValues)
	}
}

// buildChainPath 构建从断裂步骤到终点的链路路径描述
// 格式：左链当前步骤表 → ... → 左链末尾表 → 右链末尾表 → ... → 右链第一步表
// 例如：道具表|Item → 武将表|Hero → 赛季战令表|SeasonPass
func buildChainPath(stepIdx int, cfg *ChainPairConfig) string {
	var tables []string

	// 左链：从当前步骤到末尾
	for i := stepIdx; i < len(cfg.Left.Steps); i++ {
		if cfg.Left.Steps[i].Sheet != "" {
			tables = append(tables, cfg.Left.Steps[i].Sheet)
		}
	}

	// 右链：反向（末尾到第一步）
	for i := len(cfg.Right.Steps) - 1; i >= 0; i-- {
		if cfg.Right.Steps[i].Sheet != "" {
			tables = append(tables, cfg.Right.Steps[i].Sheet)
		}
	}

	// 去重相邻重复表名
	var deduped []string
	for _, t := range tables {
		if len(deduped) == 0 || deduped[len(deduped)-1] != t {
			deduped = append(deduped, t)
		}
	}

	return strings.Join(deduped, " → ")
}

// sheetNameMatch 判断步骤配置中的表名是否匹配预警表名
// 支持精确匹配和后缀匹配（"中文|英文" 格式）
func sheetNameMatch(stepSheet, warnSheet string) bool {
	if stepSheet == warnSheet {
		return true
	}
	// 后缀匹配：stepSheet 可能是 "赛季战令表|SeasonPass"，warnSheet 是 "SeasonPass"
	if strings.HasSuffix(stepSheet, "|"+warnSheet) {
		return true
	}
	// 反向后缀匹配
	if strings.HasSuffix(warnSheet, "|"+stepSheet) {
		return true
	}
	return false
}

// extractWarnValues 预警时间提取（每一步都检查）
// 如果当前步骤的 sheet 匹配 chainWarnSheet 且有 chainWarnCol，从匹配行提取预警时间列的值
func (h *LeftStepHandler) extractWarnValues(ctx *ChainContext, targetCols [][]string, matchRows []int) {
	if h.Step.Sheet == "" || targetCols == nil {
		return
	}
	warnSheet := ctx.WarnSheet()
	warnCol := ctx.WarnCol()
	if warnSheet == "" || warnCol == "" {
		return
	}
	if !sheetNameMatch(h.Step.Sheet, warnSheet) {
		return
	}
	warnColIdx := helpers.GetColIndexByName(targetCols, warnCol)
	if warnColIdx < 0 {
		return
	}
	for _, rowIdx := range matchRows {
		if v := helpers.GetColValue(targetCols, warnColIdx, rowIdx); v != "" {
			ctx.WarnValues = append(ctx.WarnValues, v)
		}
	}
}

// buildLeftResult 构建左链最终执行结果（LeftResult）
// 当 IsLast=true 时调用，将当前步骤提取的值封装为 ChainResult
//
// 执行流程：
//  1. 如果有 CompareCol，查找比较列索引
//  2. 对每个匹配的行/值，提取 Match 和 Compare 值
//  3. 如果匹配预警表，从匹配行提取预警时间列的值写入 ctx.WarnValues
//  4. 写入 ctx.LeftResult
func (h *LeftStepHandler) buildLeftResult(ctx *ChainContext, targetCols [][]string, matchRows []int, nextValues []string) {
	var resultValues []ChainValue

	if h.ChainCfg.CompareCol != "" && targetCols != nil {
		// 有比较列：从匹配行提取比较值
		compareColIdx := helpers.GetColIndexByName(targetCols, h.ChainCfg.CompareCol)

		for i, rowIdx := range matchRows {
			matchValue := nextValues[i]
			var compareValue string
			if compareColIdx >= 0 {
				compareValue = helpers.GetColValue(targetCols, compareColIdx, rowIdx)
			}
			resultValues = append(resultValues, ChainValue{
				Match:   matchValue,
				Compare: compareValue,
			})
		}
	} else {
		// 无比较列：仅构建 Match 值
		for _, nv := range nextValues {
			resultValues = append(resultValues, ChainValue{Match: nv})
		}
	}

	ctx.LeftResult = &ChainResult{
		Values:               resultValues,
		StepValues:           ctx.LeftStepValues,
		FirstStepInputValues: ctx.FirstStepInputValues,
	}
}
