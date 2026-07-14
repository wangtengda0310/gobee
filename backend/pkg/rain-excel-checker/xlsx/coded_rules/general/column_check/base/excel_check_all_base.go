// Package base 提供列级别的通用校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package base

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// AllBaseCheckRule 基础规则合并检查
// 支持的基础检查项：unique（唯一性）、chsOnly（仅中文）、increase（自增）
// 新增支持正则提取参数：pattern（正则表达式）、groups（捕获组）、fullMatch（是否完全匹配）
// 当配置了 pattern 时，会先用正则从单元格提取数据，然后对提取的值进行基础规则校验
type AllBaseCheckRule struct{}

// Check 执行通用基础检查
//
// 执行流程：
//  1. 解析参数：
//     a. BREAK_LINE: 空行连续阈值（默认3行）
//     b. ALLOW_EMPTY: 是否允许空值
//     c. ALLOW_COMMIT: 是否允许注释
//     d. unique: 是否检查唯一性
//     e. chsOnly: 是否检查仅中文
//     f. increase: 是否检查自增
//     g. pattern: 正则表达式（可选，用于提取数据）
//     h. groups: 捕获组索引（可选，配合pattern使用）
//     i. fullMatch: 是否要求完全匹配（可选，默认true）
//  2. 自动检测数据结束位置并提取列数据
//  3. 初始化检查结果切片和辅助数据结构
//  4. 遍历列数据，对每个单元格：
//     a. 处理空值和注释（根据配置决定是否跳过）
//     b. 如果配置了pattern，先用正则提取数据
//     c. 如果启用unique检查，验证值是否唯一
//     d. 如果启用chsOnly检查，验证是否仅含中文字符
//     e. 如果启用increase检查，验证是否自增
//  5. 返回所有错误信息
func (c *AllBaseCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	// 步骤1: 解析参数
	// 步骤1a: BREAK_LINE - 空行连续阈值
	breakLine := helpers.ParseBreakLine(params)

	// 步骤2: 自动检测数据结束位置并提取列数据
	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]

	// 步骤3: 初始化检查结果切片和辅助数据结构
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 步骤1b-1f: 解析检查参数
	allowEmpty := helpers.ParseAllowEmpty(params)

	allowCommit := helpers.ParseAllowCommit(params)

	unique := false
	if allow, ok := params["unique"]; ok {
		unique = strings.ToLower(allow) == "true"
	}

	chsOnly := false
	if allow, ok := params["chsOnly"]; ok {
		chsOnly = strings.ToLower(allow) == "true"
	}

	increase := false
	if allow, ok := params["increase"]; ok {
		increase = strings.ToLower(allow) == "true"
	}

	// 步骤1g-1i: 解析正则提取参数
	var extractPattern *regexp.Regexp
	var captureGroups []int
	hasPattern := false
	fullMatch := true

	if patternStr, ok := params[string(json_rule.PATTERN)]; ok && patternStr != "" {
		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("正则表达式编译错误: %v", err),
			}}
		}
		extractPattern = pattern
		hasPattern = true
		captureGroups = helpers.ParseCaptureGroups(params[string(json_rule.GROUPS)])

		if full, ok := params["fullMatch"]; ok {
			fullMatch = strings.ToLower(full) == "true"
		}
	}

	// 步骤3: 初始化辅助数据结构
	valueMap := make(map[string]int)    // unique 检查所需
	chsPattern := `^[\p{Han}\p{P}\s]+$` // chsOnly 检查所需
	idRef := int64(1)                   // increase 检查所需

	// 步骤4: 遍历列数据
	for i, s := range myColData {
		// 步骤4a: 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 步骤4b: 如果配置了pattern，先用正则提取数据
		valuesToCheck := []string{s}
		if hasPattern {
			extracted := helpers.ExtractValuesByRegex(s, extractPattern, captureGroups)
			if len(extracted) == 0 {
				// 正则未匹配到任何值
				if fullMatch {
					// 完全匹配模式下，未匹配到值视为错误
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("正则未匹配到任何可提取的值: %s", s),
					})
					continue
				}
				// 非完全匹配模式下，使用原值继续检查
			} else {
				valuesToCheck = extracted
			}
		}

		// 对每个提取的值进行基础规则校验
		for _, val := range valuesToCheck {
			val = strings.TrimSpace(val)
			if val == "" {
				continue
			}

			// 步骤4c: 检查唯一性
			if unique && val != "" {
				if prevIdx, exists := valueMap[val]; exists {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("值重复: %s, 第一次出现在第%d行", val, prevIdx+1),
					})
				} else {
					valueMap[val] = i
				}
			}

			// 步骤4d: 检查仅中文
			if chsOnly {
				if matched, _ := regexp.MatchString(chsPattern, val); !matched {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("包含非中文字符: %s", val),
					})
				}
			}

			// 步骤4e: 检查自增
			if increase {
				if id, err := strconv.ParseInt(val, 10, 64); err == nil {
					if id != idRef {
						res = append(res, &json_rule.CellError{
							Index:    i,
							ExcelRow: startRowIdx + i + 1,
							Reason:   fmt.Sprintf("ID自增有问题, 当前ID：%d, 应为ID: %d", id, i+1),
						})
					}
				} else {
					res = append(res, &json_rule.CellError{
						Index:    i,
						ExcelRow: startRowIdx + i + 1,
						Reason:   fmt.Sprintf("ID类型有问题, 当前id: %s", val),
					})
				}
			}
		}

		if increase {
			idRef++
		}
	}

	// 步骤5: 返回所有错误信息
	return res
}
