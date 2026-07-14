// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件提供解析函数和单链执行引擎
package chain_reference

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"github.com/xuri/excelize/v2"
)

// ==================== 解析函数 ====================

// ParseChainPairConfig 从 JSON 字符串解析两条关系链配置
//
// 参数：
//   - jsonStr: JSON 格式的关系链配置字符串
//
// 返回：
//   - 解析后的 ChainPairConfig
//   - 错误信息（JSON 格式错误时返回）
func ParseChainPairConfig(jsonStr string) (*ChainPairConfig, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("关系链配置为空")
	}

	var config ChainPairConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("关系链配置解析失败: %w", err)
	}

	if len(config.Left.Steps) == 0 && len(config.Right.Steps) == 0 {
		return nil, fmt.Errorf("关系链配置中两条链都没有步骤")
	}

	return &config, nil
}

// ==================== 单链执行引擎 ====================

// ExecuteChain 执行单条关系链，返回最终提取的值对集合
//
// 执行流程：
//  1. 遍历链中的每个步骤（ChainStep）
//  2. 第一步有两种模式：
//     a. Sheet 为空（左链仅取值模式）：从当前表 nextCol 列取值作为初始输入（nextCol 为空时回退到当前列）
//     b. Sheet 非空（右链全表扫描模式）：无需输入值，直接遍历目标表
//  3. 后续步骤：使用上一步提取的值集合作为输入
//  4. 对每个输入值（或全表扫描模式）：
//     a. 如果有 Pattern，使用正则提取子值列表
//     b. 打开目标表（step.Sheet），获取列数据
//     c. 在目标表中查找 PreCol 列值匹配的行（支持过滤条件）
//     d. 从匹配行提取 NextCol 列的值
//  5. 收集所有提取的值作为下一步的输入
//  6. 最后一步：
//     a. 如果 cfg.CompareCol 非空，查找比较列索引
//     b. 对每个匹配行，提取 NextCol 值作为 ChainValue.Match
//     c. 如果 CompareCol 存在，提取对应值作为 ChainValue.Compare
//     d. 返回 ChainResult
//
// 参数：
//   - cols: 当前表的列数据
//   - colIdx: 当前检查列的索引（左链第一步 nextCol 为空时作为回退取值来源）
//   - rowIdx: 当前行索引（绝对行号）
//   - startRowIdx: 数据起始行索引
//   - cfg: 关系链配置
//   - sheetMap: 所有表的数据映射
//
// 返回：
//   - 关系链执行结果（包含所有提取的值对）
//   - 错误信息
//
// applyIsArray 按 isArray 参数拆分值列表中的每个元素
// 当 isArray == "true" 时，对每个值调用 SplitArrayElements 按逗号拆分（保留花括号内容）
// 否则原样返回
func applyIsArray(values []string, isArray string) []string {
	if strings.ToLower(isArray) != "true" {
		return values
	}
	var result []string
	for _, v := range values {
		result = append(result, SplitArrayElements(v, ",")...)
	}
	return result
}

func ExecuteChain(cols [][]string, colIdx int, rowIdx int, startRowIdx int, cfg ChainConfig, sheetMap map[string]*excelize.File) (*ChainResult, error) {
	var currentValues []string
	var stepValues [][]string
	var firstStepSearchVals []string // 第一步的所有查找值（正则提取后、去重），用于两阶段比较 Phase 2

	for stepIdx, step := range cfg.Steps {
		// 步骤 2: 获取当前输入值
		var inputValues []string

		if stepIdx == 0 {
			if step.Sheet == "" {
				// 左链第一步（仅取值模式）：从 nextCol 指定的列取值
				// nextCol 为空时回退到当前列（colIdx）
				var val string
				if step.NextCol != "" {
					nextColIdx := helpers.GetColIndexByName(cols, step.NextCol)
					if nextColIdx < 0 {
						return nil, fmt.Errorf("步骤%d: 未找到列: %s", stepIdx+1, step.NextCol)
					}
					val = helpers.GetColValue(cols, nextColIdx, rowIdx)
				} else {
					if colIdx < 0 || colIdx >= len(cols) {
						return nil, fmt.Errorf("步骤%d: 当前列索引无效: %d", stepIdx+1, colIdx)
					}
					val = helpers.GetColValue(cols, colIdx, rowIdx)
				}
				if val == "" {
					return nil, nil
				}
				inputValues = []string{val}
			}
			// else: 有 sheet（右链全表扫描模式），inputValues 保持为空
		} else {
			// 后续步骤：使用上一步的结果
			inputValues = currentValues
		}

		// inputValues 为空时：只有右链全表扫描模式（第一步 + 有 sheet）允许继续
		if len(inputValues) == 0 && !(stepIdx == 0 && step.Sheet != "") {
			return nil, nil
		}

		// 左链第一步"仅取值"模式：sheet 为空时不做跨表查找，直接将取到的值传给下一步
		if stepIdx == 0 && step.Sheet == "" {
			var nextValues []string
			if step.Pattern != "" {
				re, err := regexp.Compile(step.Pattern)
				if err != nil {
					return nil, fmt.Errorf("步骤%d: 正则编译失败: %w", stepIdx+1, err)
				}
				groups := helpers.ParseCaptureGroups(step.Groups)
				extracted := helpers.ExtractValuesByRegex(inputValues[0], re, groups)
				nextValues = applyIsArray(extracted, step.IsArray)
			} else {
				nextValues = inputValues
			}
			stepValues = append(stepValues, append([]string{}, nextValues...))

			// 保存查找值（仅取值模式下 nextValues 即为查找值）
			firstStepSearchVals = nextValues

			// 如果是唯一步骤（没有后续跳转），构建 ChainResult 返回
			if stepIdx == len(cfg.Steps)-1 {
				var resultValues []ChainValue
				for _, nv := range nextValues {
					resultValues = append(resultValues, ChainValue{Match: nv})
				}
				return &ChainResult{Values: resultValues, StepValues: stepValues, FirstStepInputValues: firstStepSearchVals}, nil
			}
			currentValues = nextValues
			if len(currentValues) == 0 {
				return nil, nil
			}
			continue
		}

		// 步骤 4a: 打开目标表
		targetFile, targetSheetName, found := helpers.FindSheetBySuffix(sheetMap, step.Sheet)
		if !found {
			return nil, fmt.Errorf("步骤%d: 目标表不存在: %s", stepIdx+1, step.Sheet)
		}

		targetCols, err := targetFile.GetCols(targetSheetName)
		if err != nil {
			return nil, fmt.Errorf("步骤%d: 读取目标表失败: %w", stepIdx+1, err)
		}

		// 查找目标列索引
		preColIdx := helpers.GetColIndexByName(targetCols, step.PreCol)
		if preColIdx < 0 {
			return nil, fmt.Errorf("步骤%d: 目标表中未找到查找列: %s", stepIdx+1, step.PreCol)
		}

		nextColIdx := -1
		if step.NextCol != "" {
			nextColIdx = helpers.GetColIndexByName(targetCols, step.NextCol)
			if nextColIdx < 0 {
				return nil, fmt.Errorf("步骤%d: 目标表中未找到提取列: %s", stepIdx+1, step.NextCol)
			}
		}

		// 预编译正则（如果有）
		var stepPattern *regexp.Regexp
		var stepGroups []int
		if step.Pattern != "" {
			re, err := regexp.Compile(step.Pattern)
			if err != nil {
				return nil, fmt.Errorf("步骤%d: 正则编译失败: %w", stepIdx+1, err)
			}
			stepPattern = re
			stepGroups = helpers.ParseCaptureGroups(step.Groups)
		}

		// 获取过滤后的行列表（如果有过滤条件）
		targetStartRow := excelio.MJS_FIXED_ROWS_NUM
		filteredRows, ferr := FilterRowsByConditionEx(targetCols, FilterOptions{
			FilterColName: step.FilterCol,
			FilterVal:     step.FilterVal,
			StartRowIdx:   targetStartRow,
			FilterIsArray: step.FilterIsArray,
			FilterMode:    step.FilterMode,
			FilterDays:    step.FilterDays,
		})
		if ferr != nil {
			return nil, fmt.Errorf("步骤%d: 过滤条件错误: %w", stepIdx+1, ferr)
		}
		hasFilter := step.FilterCol != "" && (step.FilterVal != "" || step.FilterMode == "withinDays")

		// 将过滤行列表构建为 map，避免内层循环线性搜索（O(n) → O(1)）
		filterRowSet := make(map[int]bool, len(filteredRows))
		for _, fr := range filteredRows {
			filterRowSet[fr] = true
		}

		targetEndIdx := helpers.GetDataEndIndex(targetCols, targetStartRow)

		// 步骤 4b-4d: 对每个输入值执行查找和提取
		var nextValues []string
		var nextMatchRows []int // 记录匹配的行索引，用于最后一步提取 Compare 值

		if stepIdx == 0 && len(inputValues) == 0 {
			// 右链全表扫描模式：无输入值，直接遍历目标表
			// 不需要从当前表取值，直接扫描目标表所有行（或过滤后的行）
			for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
				if hasFilter && !filterRowSet[rowI] {
					continue
				}
				// 收集 preCol 值作为 FirstStepInputValues（用于 Phase 2）
				findVal := helpers.GetColValue(targetCols, preColIdx, rowI)
				if findVal == "" {
					continue
				}
				firstStepSearchVals = append(firstStepSearchVals, findVal)

				// 提取 nextCol 值传给下一步
				if nextColIdx >= 0 {
					extractedVal := helpers.GetColValue(targetCols, nextColIdx, rowI)
					if extractedVal != "" {
						nextValues = append(nextValues, extractedVal)
						if stepIdx == len(cfg.Steps)-1 {
							nextMatchRows = append(nextMatchRows, rowI)
						}
					}
				} else {
					// 没有 nextCol，preCol 值既作为 FirstStepInputValues 也传给下一步
					nextValues = append(nextValues, findVal)
					if stepIdx == len(cfg.Steps)-1 {
						nextMatchRows = append(nextMatchRows, rowI)
					}
				}
			}
		} else {
			// 原有逻辑：对每个输入值执行查找和提取
			for _, inputVal := range inputValues {
				// 如果有 Pattern，先提取子值
				var searchValues []string
				if stepPattern != nil {
					extracted := helpers.ExtractValuesByRegex(inputVal, stepPattern, stepGroups)
					if len(extracted) == 0 {
						continue
					}
					searchValues = applyIsArray(extracted, step.IsArray)
				} else {
					searchValues = applyIsArray([]string{inputVal}, step.IsArray)
				}

				// 构建查找值的 set
				searchSet := make(map[string]bool)
				for _, sv := range searchValues {
					searchSet[strings.TrimSpace(sv)] = true
				}

				// 遍历目标表查找匹配行
				for rowI := targetStartRow; rowI < targetEndIdx; rowI++ {
					// 如果有过滤条件，O(1) 检查是否在过滤行集合中
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
							// 如果是最后一步，记录匹配行索引
							if stepIdx == len(cfg.Steps)-1 {
								nextMatchRows = append(nextMatchRows, rowI)
							}
						}
					} else {
						// 没有提取列，使用匹配到的值
						nextValues = append(nextValues, cellVal)
						if stepIdx == len(cfg.Steps)-1 {
							nextMatchRows = append(nextMatchRows, rowI)
						}
					}
				}
			}

			// 第一步时收集所有查找值（去重）— 仅在有 inputValues 的路径执行
			if stepIdx == 0 {
				seen := make(map[string]bool)
				for _, inputVal := range inputValues {
					var sv []string
					if stepPattern != nil {
						sv = applyIsArray(helpers.ExtractValuesByRegex(inputVal, stepPattern, stepGroups), step.IsArray)
					} else {
						sv = applyIsArray([]string{inputVal}, step.IsArray)
					}
					for _, v := range sv {
						v = strings.TrimSpace(v)
						if !seen[v] {
							seen[v] = true
							firstStepSearchVals = append(firstStepSearchVals, v)
						}
					}
				}
			}
		}

		// 对提取的 nextCol 值应用 isArray 拆分
		// 例如 DropGroup="91001,91003,91004" → ["91001","91003","91004"]
		nextValues = applyIsArray(nextValues, step.IsArray)

		// 保存当前步的 nextValues（在最后一步判断之前，确保最后一步也被记录）
		stepValues = append(stepValues, append([]string{}, nextValues...))

		// 如果是最后一步，构建 ChainResult
		if stepIdx == len(cfg.Steps)-1 {
			var resultValues []ChainValue

			// 查找比较列索引
			compareColIdx := -1
			if cfg.CompareCol != "" {
				compareColIdx = helpers.GetColIndexByName(targetCols, cfg.CompareCol)
			}

			// 对每个匹配的行，提取 Match 和 Compare 值
			for i, rowIdx := range nextMatchRows {
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

			return &ChainResult{Values: resultValues, StepValues: stepValues, FirstStepInputValues: firstStepSearchVals}, nil
		}

		currentValues = nextValues
		if len(currentValues) == 0 {
			return nil, nil
		}
	}

	// 没有步骤，返回空结果
	return &ChainResult{Values: nil, StepValues: stepValues, FirstStepInputValues: firstStepSearchVals}, nil
}
