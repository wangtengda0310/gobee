// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现洋葱模型的匹配处理器（MatchHandler）
// Match 是两阶段门控的 Phase 1：在左链执行完毕后，用左链最后一步的值与右链最终表的值做交汇判断
// 匹配成功时将键值传递给右链反向步骤；匹配失败时短路返回，不执行右链
package chain_reference

import (
	"fmt"
	"regexp"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
)

// MatchHandler 匹配处理器
// 在左链执行完毕后，用左链最后一步的值与右链最终表的值做门控判断
//
// 前置处理（在调用 next 之前）：
//   - 获取左链最终值（ctx.LeftCurrentValues）
//   - 打开右链最后一步的目标表，提取值作为右链最终值集合
//   - 用 MatchByType 判断是否匹配
//   - 匹配成功：计算传递给右链的键值，写入 ctx.RightCurrentValues，调用 next
//   - 匹配失败：设置 ctx.Matched=false，不调用 next（短路返回）
type MatchHandler struct {
	MatchType     string      // 匹配类型（verify_exists/date_* 等）
	RightLastStep ChainStep   // 右链最后一步配置（用于获取右链最终表和列信息）
	RightConfig   ChainConfig // 右链完整配置（用于 CompareCol）
}

// Handle 执行匹配判断
//
// 执行流程：
//  1. 前置处理：获取左链最终值和右链最终值集合
//  2. 用 MatchByType 判断是否匹配
//  3. 匹配成功：计算传递键值，写入 ctx，调用 next（传递给右链反向步骤）
//  4. 匹配失败：
//     - 门控类型（verify_exists/date_*）：设置 ctx.Matched=false，短路返回（静默）
//     - 强制类型（verify_must_exist）：设置 ctx.Matched=false + ctx.Violation=true + Reason，短路返回（报错）
func (h *MatchHandler) Handle(ctx *ChainContext, next NextFunc) error {
	// 前置处理：执行匹配判断
	matched, passToRight, reason := h.performMatch(ctx)

	if !matched {
		ctx.Matched = false
		// 强制类型：将 Match 失败升级为 Violation，由 Compare 层放行 ctx.Violation
		if IsMatchTypeStrict(h.MatchType) && reason != "" {
			ctx.Violation = true
			ctx.Reason = reason
		}
		return nil
	}

	// 匹配成功：将键值传递给右链反向步骤
	ctx.Matched = true
	ctx.MatchedKeys = passToRight
	ctx.RightCurrentValues = passToRight

	// 调用 next，将控制权传递给右链反向步骤
	return next(ctx)
}

// performMatch 执行匹配判断
//
// 执行流程：
//  1. 从 ctx.LeftCurrentValues 获取左链最终值
//  2. 打开右链最后一步的目标表
//  3. 提取 NextCol 列所有值（或全表扫描时的 PreCol 值）作为右链最终值集合
//  4. 用 MatchByType 判断是否匹配
//  5. 返回匹配结果、传递给右链反向步骤的键值、失败时的详细 reason
func (h *MatchHandler) performMatch(ctx *ChainContext) (bool, []string, string) {
	// 步骤 1: 获取左链最终值
	leftVals := ctx.LeftCurrentValues
	if len(leftVals) == 0 {
		// 左链无值，匹配失败（无具体缺失值描述）
		return false, nil, ""
	}

	// 步骤 2: 打开右链最后一步的目标表
	targetFile, targetSheetName, found := helpers.FindSheetBySuffix(ctx.SheetMap, h.RightLastStep.Sheet)
	if !found {
		ctx.Err = fmt.Errorf("匹配阶段: 右链最后一步目标表不存在: %s", h.RightLastStep.Sheet)
		return false, nil, ""
	}

	targetCols, err := ctx.GetCachedCols(targetSheetName, targetFile)
	if err != nil {
		ctx.Err = fmt.Errorf("匹配阶段: 读取右链最终表失败: %w", err)
		return false, nil, ""
	}

	// 步骤 3: 提取右链最终值
	rightVals := h.getRightFinalValues(ctx, targetCols)
	if ctx.Err != nil {
		return false, nil, ""
	}

	// 步骤 4: 用 MatchByType 判断是否匹配
	matched, reason := MatchByType(h.MatchType, leftVals, rightVals)
	if !matched {
		return false, nil, reason
	}

	// 步骤 5: 传递右链全部值给反向查找步骤
	return true, rightVals, ""
}

// getRightFinalValues 从右链最终表中提取值集合
//
// 执行流程：
//  1. 确定提取列（NextCol 优先，为空时用 PreCol）
//  2. 遍历数据行，提取非空值
//  3. 返回去重后的值列表
func (h *MatchHandler) getRightFinalValues(ctx *ChainContext, targetCols [][]string) []string {
	// 确定提取列
	var colName string
	if h.RightLastStep.NextCol != "" {
		colName = h.RightLastStep.NextCol
	} else {
		colName = h.RightLastStep.PreCol
	}

	colIdx := helpers.GetColIndexByName(targetCols, colName)
	if colIdx < 0 {
		return nil
	}

	// 应用过滤条件
	targetStartRow := excelio.MJS_FIXED_ROWS_NUM
	filteredRows, ferr := FilterRowsByConditionEx(targetCols, FilterOptions{
		FilterColName: h.RightLastStep.FilterCol,
		FilterVal:     h.RightLastStep.FilterVal,
		StartRowIdx:   targetStartRow,
		FilterIsArray: h.RightLastStep.FilterIsArray,
		FilterMode:    h.RightLastStep.FilterMode,
		FilterDays:    h.RightLastStep.FilterDays,
	})
	if ferr != nil {
		ctx.Err = fmt.Errorf("匹配阶段: 过滤条件错误: %w", ferr)
		return nil
	}
	hasFilter := h.RightLastStep.FilterCol != "" && (h.RightLastStep.FilterVal != "" || h.RightLastStep.FilterMode == "withinDays")

	filterRowSet := make(map[int]bool, len(filteredRows))
	for _, fr := range filteredRows {
		filterRowSet[fr] = true
	}

	targetEndIdx := helpers.GetDataEndIndex(targetCols, targetStartRow)

	// 预编译右链最后一步的正则（如有）
	// 修复：之前直接用 GetColValue 取原始字符串作为右链最终值集合，未应用 Pattern；
	// 导致 ShopGoods.Item="{1039037;1}" 与左链提取出的纯数字 "1039037" 不匹配 → 全部误报缺失
	var stepPattern *regexp.Regexp
	var stepGroups []int
	if h.RightLastStep.Pattern != "" {
		re, err := regexp.Compile(h.RightLastStep.Pattern)
		if err != nil {
			ctx.Err = fmt.Errorf("匹配阶段: 右链最后一步正则编译失败: %w", err)
			return nil
		}
		stepPattern = re
		stepGroups = helpers.ParseCaptureGroups(h.RightLastStep.Groups)
	}

	// 遍历提取值
	seen := make(map[string]bool)
	var result []string
	for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
		if hasFilter && !filterRowSet[rowI] {
			continue
		}
		val := helpers.GetColValue(targetCols, colIdx, rowI)
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		// 应用 pattern 正则提取子值（如有）
		var extracted []string
		if stepPattern != nil {
			extracted = helpers.ExtractValuesByRegex(val, stepPattern, stepGroups)
		} else {
			extracted = []string{val}
		}
		for _, sub := range extracted {
			sub = strings.TrimSpace(sub)
			if sub != "" && !seen[sub] {
				seen[sub] = true
				result = append(result, sub)
			}
		}
	}

	return result
}
