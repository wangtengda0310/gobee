// Package reference 提供引用关系相关的校验规则
// 本文件实现 CHAIN_REFERENCE（跨表关系链检查）规则
package reference

import (
	"fmt"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule/chain_reference"
	"github.com/xuri/excelize/v2"
)

// ChainReferenceCheckRule 跨表关系链检查规则
//
// 支持动态 N 步跳转的两条关系链汇合比较：
//   - 来源链（left）：从当前列出发，经多步跳转提取目标值集合
//   - 目标链（right）：从当前列出发，经多步跳转提取目标值集合
//   - 比较类型：verify_exists（验证存在）、time_overlap（时间点匹配）、date_*（日期比较）
//
// 典型用例：检查 DrawFix 中的武将是否违反战令保护期
// 来源链：SeasonPassReward → Item → Hero → StartTime
// 目标链：DrawFix.ItemIds → Item → Hero → EndTime
type ChainReferenceCheckRule struct{}

// Check 执行跨表关系链检查
//
// 执行流程：
//  1. 解析参数：chainSteps（JSON 链配置）、chainCompare（比较阶段类型）、chainMatchCompare（匹配阶段类型）
//  2. 解析 chainSteps JSON 获取两条链的配置
//  3. 获取当前列的数据范围（处理空值和注释）
//  4. 遍历当前列的每个单元格：
//     a. 跳过空值和注释行
//     b. 根据 chainMatchCompare 选择执行路径：
//     - 洋葱模型路径（chainMatchCompare 非空且右链有步骤）：BuildOnionChain 执行
//     - 旧路径（其他情况）：ExecuteChain 分别执行两条链
//     c. 比较结果
//     d. 不通过则记录错误
//  5. 返回错误列表
func (c *ChainReferenceCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)
	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]

	// 预警窗口启用检查：chainWarnBefore 为空时表示不启用此规则，直接返回空结果
	// chainWarnBefore="0h" 表示无条件启用（始终报错），chainWarnBefore="168h" 表示仅在7天窗口内报错
	warnBeforeVal := params["chainWarnBefore"]
	if warnBeforeVal == "" {
		return nil
	}

	res := make([]*json_rule.CellError, 0, len(myColData))

	// 步骤 1: 解析参数
	allowEmpty := helpers.ParseAllowEmpty(params)
	allowCommit := helpers.ParseAllowCommit(params)

	// 解析比较类型，默认为 verify_exists
	compareType := "verify_exists"
	if ct, ok := params["chainCompare"]; ok && ct != "" {
		compareType = ct
	}

	// 解析匹配阶段类型（两阶段门控模型的匹配规则）
	matchCompareType := ""
	if mct, ok := params["chainMatchCompare"]; ok && mct != "" {
		matchCompareType = mct
	}

	// 解析时间比较键
	leftKey := params["chainLeftKey"]
	rightKey := params["chainRightKey"]

	// 步骤 2: 解析链配置
	chainStepsJSON, ok := params["chainSteps"]
	if !ok || chainStepsJSON == "" {
		return []*json_rule.CellError{{
			Index:  0,
			Reason: "关系链参数 chainSteps 未配置",
		}}
	}

	pairConfig, err := chain_reference.ParseChainPairConfig(chainStepsJSON)
	if err != nil {
		return []*json_rule.CellError{{
			Index:  0,
			Reason: fmt.Sprintf("关系链配置解析失败: %v", err),
		}}
	}

	// 判断是否使用两阶段比较：time_overlap + 两链都有 compareCol（现有特化逻辑）
	hasLeftCompare := pairConfig.Left.CompareCol != ""
	hasRightCompare := pairConfig.Right.CompareCol != ""
	useTimeOverlapTwoPhase := compareType == "time_overlap" && hasLeftCompare && hasRightCompare

	// 判断是否使用洋葱模型路径：chainMatchCompare 非空且右链有步骤
	useOnionModel := matchCompareType != "" && len(pairConfig.Right.Steps) > 0

	// 如果使用洋葱模型，预构建洋葱链（所有行共用同一个执行函数）
	var onionFunc func(ctx *chain_reference.ChainContext) error
	if useOnionModel {
		onionFunc = chain_reference.BuildOnionChain(pairConfig, params)
	}

	// 步骤 3.5: 左链第一步本表行过滤
	// 当左链第一步 sheet=""（本表取值模式）且配置了 filterCol/filterVal 时，
	// 预计算符合条件的行集合，主循环中跳过不在集合中的行，避免对不相关行执行关系链查找
	//
	// 注意：leftFirstStepFilterSet 必须区分两种状态：
	//   - nil：过滤未启用，主循环对所有行执行检查
	//   - 非 nil（即使是空 map）：过滤已启用，主循环按集合过滤；空 map 表示无行匹配过滤条件
	// 此前 `if len(filteredRows) > 0` 把"过滤后无行匹配"误降级为"未启用过滤"，导致 filterDays
	// 窗口外时整表被检查（参见 commit 中的回归测试）
	var leftFirstStepFilterSet map[int]bool
	if len(pairConfig.Left.Steps) > 0 {
		firstStep := pairConfig.Left.Steps[0]
		filterEnabled := firstStep.Sheet == "" && firstStep.FilterCol != "" && (firstStep.FilterVal != "" || firstStep.FilterMode == "withinDays")
		if filterEnabled {
			filteredRows, ferr := chain_reference.FilterRowsByConditionEx(cols, chain_reference.FilterOptions{
				FilterColName: firstStep.FilterCol,
				FilterVal:     firstStep.FilterVal,
				StartRowIdx:   startRowIdx,
				FilterIsArray: firstStep.FilterIsArray,
				FilterMode:    firstStep.FilterMode,
				FilterDays:    firstStep.FilterDays,
			})
			if ferr != nil {
				return []*json_rule.CellError{{
					Index:  0,
					Reason: fmt.Sprintf("左链第一步过滤条件错误: %v", ferr),
				}}
			}
			// 即使 filteredRows 为空也要建立空 map，标记"过滤已启用但无行匹配"
			leftFirstStepFilterSet = make(map[int]bool, len(filteredRows))
			for _, rowIdx := range filteredRows {
				leftFirstStepFilterSet[rowIdx-startRowIdx] = true
			}
		}
	}

	// 步骤 3.8: 创建 GetCols 缓存，所有行共用同一个 onionFunc 避免重复调用 GetCols
	colsCache := make(map[string][][]string)

	// 步骤 4: 遍历每个单元格
	for i := range myColData {
		// 左链第一步本表行过滤：跳过不符合过滤条件的行（尽早裁剪，减少后续开销）
		if leftFirstStepFilterSet != nil && !leftFirstStepFilterSet[i] {
			continue
		}

		// 步骤 4a: 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		rowIdx := startRowIdx + i

		if useOnionModel {
			// === 洋葱模型路径 ===
			ctx := chain_reference.NewChainContext(cols, colsCache, colIdx, rowIdx, startRowIdx, sheetMap, pairConfig, params, myColData, i)
			if err := onionFunc(ctx); err != nil {
				// 执行错误（链路断开）也经过预警过滤
				if !chain_reference.ShouldSuppressByWarnBefore(ctx, time.Now()) {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   err.Error(),
					})
				}
				continue
			}
			if ctx.Violation {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   ctx.Reason,
				})
			}
		} else {
			// === 旧路径（保持不变） ===
			// 步骤 4b: 执行来源链（left）
			var leftResult *chain_reference.ChainResult
			if len(pairConfig.Left.Steps) > 0 {
				leftResult, err = chain_reference.ExecuteChain(cols, colIdx, rowIdx, startRowIdx, pairConfig.Left, sheetMap)
				if err != nil {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("来源链执行失败: %v", err),
					})
					continue
				}
			}

			// 步骤 4c: 执行目标链（right）
			var rightResult *chain_reference.ChainResult
			if len(pairConfig.Right.Steps) > 0 {
				rightResult, err = chain_reference.ExecuteChain(cols, colIdx, rowIdx, startRowIdx, pairConfig.Right, sheetMap)
				if err != nil {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("目标链执行失败: %v", err),
					})
					continue
				}
			}

			// 如果两个链都没提取到值，跳过
			if leftResult == nil || rightResult == nil {
				continue
			}
			leftVals := leftResult.MatchValues()
			rightVals := rightResult.MatchValues()

			if len(leftVals) == 0 && len(rightVals) == 0 {
				continue
			}

			// 步骤 4d: 根据比较类型进行比较
			var violation bool
			var reason string

			if matchCompareType != "" {
				// === 通用两阶段门控比较（旧路径中的两阶段）===
				leftLastVals := leftResult.LastStepValues()
				rightLastVals := rightResult.LastStepValues()
				matched, _ := chain_reference.MatchByType(matchCompareType, leftLastVals, rightLastVals)
				if !matched {
					continue
				}

				currentColVal := myColData[i]
				rightFirstInputVals := rightResult.GetFirstStepInputValues()
				// 左链第一步 isArray 控制当前列值的拆分
				var currentVals []string
				if len(pairConfig.Left.Steps) > 0 && strings.ToLower(pairConfig.Left.Steps[0].IsArray) == "true" {
					currentVals = chain_reference.SplitArrayElements(currentColVal, ",")
				} else {
					currentVals = []string{currentColVal}
				}
				violation, reason = chain_reference.CompareByType(compareType, currentVals, rightFirstInputVals)
			} else {
				// === 退化：单阶段比较（保持现有行为） ===
				if compareType == "time_overlap" {
					if useTimeOverlapTwoPhase {
						violation, reason = chain_reference.CompareTwoPhase(leftResult, rightResult, leftKey, rightKey)
					} else {
						violation, reason = chain_reference.CompareTimeMatch(leftVals, rightVals, leftKey, rightKey)
					}
				} else {
					violation, reason = chain_reference.CompareByType(compareType, leftVals, rightVals)
				}
			}

			// 旧路径预警窗口过滤：无 ChainContext，从 sheetMap 全表扫描判断
			if violation {
				warnBefore, _ := time.ParseDuration(params["chainWarnBefore"])
				if warnBefore > 0 && params["chainWarnSheet"] != "" && params["chainWarnCol"] != "" {
					if chain_reference.ShouldSuppressWarnBeforeLegacy(sheetMap, params["chainWarnSheet"], params["chainWarnCol"], warnBefore, time.Now()) {
						violation = false
						reason = ""
					}
				}
			}
			if violation {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   reason,
				})
			}
		}
	}

	return res
}
