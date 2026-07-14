// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的右链反向步骤处理器（RightStepHandler）
// 右链反向执行：先调用 next（内层），然后根据内层返回的值做反向查找
// 这是洋葱模型的核心创新——反向执行避免全表扫描
package chain_reference

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
)

// RightStepHandler 右链反向步骤处理器
// 后向执行：先调用 next（内层），然后根据内层返回的值做反向查找
//
// 反向查找与正向查找的关键差异：
//
//	正向（LeftStep）：在 PreCol 列匹配输入值，提取 NextCol 列的值
//	反向（RightStep）：在 NextCol 列匹配输入值，提取 PreCol 列的值
//
// 当 NextCol 为空时，PreCol 既作搜索列也作提取列（验证存在性）
type RightStepHandler struct {
	Step     ChainStep   // 当前步骤配置
	StepIdx  int         // 步骤索引（在原右链中的位置，0-based）
	IsFirst  bool        // 是否右链第一步（反向执行的最后一步，产生 FirstStepInputValues）
	ChainCfg ChainConfig // 链配置（用于构建 RightResult）
}

// Handle 执行右链反向步骤
//
// 执行流程：
//  1. 调用 next(ctx) — 执行内层（更内层的右链步骤或 terminal）
//  2. 后置处理（反向查找）：
//     a. 获取 ctx.RightCurrentValues（来自更内层右链步骤或 Match 的输出）
//     b. 如果为空：设置 ctx.Err 并返回
//     c. 打开目标表 Step.Sheet
//     d. 反向查找：在 NextCol 列匹配输入值，提取 PreCol 列的值
//     e. 应用过滤条件
//     f. 写入 ctx.RightCurrentValues 和 ctx.RightStepValues
//     g. 如果 IsFirst：写入 ctx.FirstStepInputValues 并构建 ctx.RightResult
func (h *RightStepHandler) Handle(ctx *ChainContext, next NextFunc) error {
	// 前向：先执行内层
	if err := next(ctx); err != nil {
		return err
	}

	// 如果内层产生错误，直接返回
	if ctx.Err != nil {
		return ctx.Err
	}

	// 如果匹配阶段未通过（Match handler 短路），直接返回
	if !ctx.Matched {
		return nil
	}

	// 后置处理：反向查找
	h.handleReverseLookup(ctx)
	return nil
}

// handleReverseLookup 执行反向查找逻辑
//
// 执行流程：
//  1. 检查 ctx.RightCurrentValues 是否为空
//  2. 打开目标表
//  3. 获取列索引（反向：NextCol 为搜索列，PreCol 为提取列）
//  4. 遍历目标表，在搜索列中匹配输入值，从提取列取值
//  5. 将结果写入 ctx.RightCurrentValues
//  6. 预警时间提取（每一步都检查）
//  7. 如果 IsFirst，写入 ctx.FirstStepInputValues 并构建 ctx.RightResult
func (h *RightStepHandler) handleReverseLookup(ctx *ChainContext) {
	// 步骤 1: 检查输入值
	if len(ctx.RightCurrentValues) == 0 {
		ctx.Err = fmt.Errorf("右链步骤%d: 反向查找输入值为空", h.StepIdx+1)
		return
	}

	// 步骤 2: 打开目标表
	targetFile, targetSheetName, found := helpers.FindSheetBySuffix(ctx.SheetMap, h.Step.Sheet)
	if !found {
		ctx.Err = fmt.Errorf("右链步骤%d: 目标表不存在: %s", h.StepIdx+1, h.Step.Sheet)
		return
	}

	targetCols, err := ctx.GetCachedCols(targetSheetName, targetFile)
	if err != nil {
		ctx.Err = fmt.Errorf("右链步骤%d: 读取目标表失败: %w", h.StepIdx+1, err)
		return
	}

	// 步骤 3: 确定搜索列和提取列
	// 反向查找：NextCol 为搜索列，PreCol 为提取列
	// NextCol 为空时：PreCol 既搜索又提取（验证存在性）
	var searchColIdx, extractColIdx int

	if h.Step.NextCol != "" {
		searchColIdx = helpers.GetColIndexByName(targetCols, h.Step.NextCol)
		if searchColIdx < 0 {
			ctx.Err = fmt.Errorf("右链步骤%d: 目标表中未找到搜索列: %s", h.StepIdx+1, h.Step.NextCol)
			return
		}
	} else {
		// NextCol 为空：PreCol 既搜索又提取
		searchColIdx = helpers.GetColIndexByName(targetCols, h.Step.PreCol)
		if searchColIdx < 0 {
			ctx.Err = fmt.Errorf("右链步骤%d: 目标表中未找到列: %s", h.StepIdx+1, h.Step.PreCol)
			return
		}
	}

	extractColIdx = helpers.GetColIndexByName(targetCols, h.Step.PreCol)
	if extractColIdx < 0 {
		ctx.Err = fmt.Errorf("右链步骤%d: 目标表中未找到提取列: %s", h.StepIdx+1, h.Step.PreCol)
		return
	}

	// 步骤 4: 获取过滤条件
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
		ctx.Err = fmt.Errorf("右链步骤%d: 过滤条件错误: %w", h.StepIdx+1, ferr)
		return
	}
	hasFilter := h.Step.FilterCol != "" && (h.Step.FilterVal != "" || h.Step.FilterMode == "withinDays")

	// 将过滤行列表构建为 map（O(n) → O(1) 查找）
	filterRowSet := make(map[int]bool, len(filteredRows))
	for _, fr := range filteredRows {
		filterRowSet[fr] = true
	}

	targetEndIdx := helpers.GetDataEndIndex(targetCols, targetStartRow)

	// 步骤 5: 构建输入值的 set
	inputSet := make(map[string]bool)
	for _, v := range ctx.RightCurrentValues {
		inputSet[strings.TrimSpace(v)] = true
	}

	// 步骤 6: 遍历目标表执行反向查找
	var nextValues []string
	var nextMatchRows []int

	for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
		if hasFilter && !filterRowSet[rowI] {
			continue
		}

		// 在搜索列中匹配输入值
		searchVal := helpers.GetColValue(targetCols, searchColIdx, rowI)
		if !inputSet[searchVal] {
			continue
		}

		// 从提取列取值
		extractedVal := helpers.GetColValue(targetCols, extractColIdx, rowI)
		if extractedVal == "" {
			continue
		}

		nextValues = append(nextValues, extractedVal)
		nextMatchRows = append(nextMatchRows, rowI)
	}

	// 步骤 7: 写入中间结果
	ctx.RightStepValues = append(ctx.RightStepValues, append([]string{}, nextValues...))
	ctx.RightCurrentValues = nextValues

	// 步骤 7.5: 预警时间提取（每一步都检查，不仅限于 IsFirst）
	h.extractWarnValues(ctx, targetCols, nextMatchRows)

	// 步骤 8: 如果是第一步（反向执行的最后一步），收集 FirstStepInputValues 并构建 RightResult
	if h.IsFirst {
		ctx.FirstStepInputValues = nextValues
		h.buildRightResult(ctx, targetCols, nextMatchRows, nextValues)
	}
}

// extractWarnValues 预警时间提取（每一步都检查）
// 如果当前步骤的 sheet 匹配 chainWarnSheet 且有 chainWarnCol，从匹配行提取预警时间列的值
func (h *RightStepHandler) extractWarnValues(ctx *ChainContext, targetCols [][]string, matchRows []int) {
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

// buildRightResult 构建右链最终执行结果（RightResult）
// 当 IsFirst=true 时调用，将当前步骤提取的值封装为 ChainResult
func (h *RightStepHandler) buildRightResult(ctx *ChainContext, targetCols [][]string, matchRows []int, nextValues []string) {
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

	ctx.RightResult = &ChainResult{
		Values:               resultValues,
		StepValues:           ctx.RightStepValues,
		FirstStepInputValues: ctx.FirstStepInputValues,
	}
}
