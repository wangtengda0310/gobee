// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件提供公共提取和过滤辅助函数
package chain_reference

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
)

// ==================== 公共提取函数 ====================

// ExtractByRegexFromString 使用字符串格式的正则和捕获组提取子值
// 封装已有的 ExtractValuesByRegex + ParseCaptureGroups，提供字符串参数接口
//
// 参数：
//   - input: 待提取的原始字符串
//   - patternStr: 正则表达式字符串
//   - groupsStr: 捕获组配置（逗号分隔，如 "1" 或 "1,2"）
//
// 返回：
//   - 提取到的值列表
//   - 错误信息
func ExtractByRegexFromString(input string, patternStr string, groupsStr string) ([]string, error) {
	if input == "" || patternStr == "" {
		return nil, nil
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("正则表达式编译错误: %w", err)
	}

	groups := helpers.ParseCaptureGroups(groupsStr)
	result := helpers.ExtractValuesByRegex(input, re, groups)
	return result, nil
}

// FilterOptions 过滤条件选项
type FilterOptions struct {
	FilterColName string    // 过滤条件列名
	FilterVal     string    // 过滤条件值
	StartRowIdx   int       // 数据起始行索引
	FilterIsArray string    // "true" 时按逗号拆分多值匹配
	FilterMode    string    // 过滤模式: ""(单值) | "multi"(多值) | "withinDays"(距今<N天)
	FilterDays    string    // 距今天数（withinDays 模式）
	Now           time.Time // 注入当前时间，零值使用 time.Now().UTC()
}

// FilterRowsByConditionEx 扩展版过滤函数，支持时间比较模式
// 根据 FilterMode 选择不同的过滤策略：
//   - "withinDays": 保留单元格时间在 now ~ now+N天 的行
//   - "multi": 强制多值匹配（等同于 filterIsArray="true"）
//   - 默认: 单值/多值匹配（取决于 FilterIsArray）
//
// 返回值：
//   - []int: 匹配的行索引列表（即使过滤启用后无行匹配也返回空切片而非 nil）
//   - error: filterCol 在目标表中不存在等"配置错误"返回 error，调用方应当上报
//     filterColName 为空时返回 (nil, nil) 表示"未启用过滤"
//
// 注意：调用方必须区分 (nil, nil)="未启用过滤" 与 ([]int{}, nil)="过滤启用但 0 行匹配"
func FilterRowsByConditionEx(cols [][]string, opts FilterOptions) ([]int, error) {
	if opts.FilterColName == "" {
		return nil, nil
	}

	switch opts.FilterMode {
	case "withinDays":
		return filterRowsByWithinDays(cols, opts)
	case "multi":
		// 多值模式：强制 filterIsArray="true"
		return filterRowsByValue(cols, opts.FilterColName, opts.FilterVal, opts.StartRowIdx, "true")
	default:
		// 单值模式（含 filterMode=""）
		return filterRowsByValue(cols, opts.FilterColName, opts.FilterVal, opts.StartRowIdx, opts.FilterIsArray)
	}
}

// filterRowsByValue 值匹配过滤（原 FilterRowsByCondition 的核心逻辑）
// 根据 filterIsArray 选择单值精确匹配或多值集合匹配
// 返回 (matchedRows, error)；filterCol 不存在时返回 error
func filterRowsByValue(cols [][]string, filterColName, filterVal string, startRowIdx int, filterIsArray string) ([]int, error) {
	if filterColName == "" || filterVal == "" {
		return nil, nil
	}

	filterColIdx := helpers.GetColIndexByName(cols, filterColName)
	if filterColIdx < 0 {
		return nil, fmt.Errorf("过滤列 %q 在目标表中不存在", filterColName)
	}

	endIdx := helpers.GetDataEndIndex(cols, startRowIdx)

	matchedRows := []int{}
	if filterIsArray == "true" {
		filterSet := make(map[string]bool)
		for _, v := range SplitArrayElements(filterVal, ",") {
			filterSet[v] = true
		}
		for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
			val := helpers.GetColValue(cols, filterColIdx, rowIdx)
			if filterSet[val] {
				matchedRows = append(matchedRows, rowIdx)
			}
		}
		return matchedRows, nil
	}

	for rowIdx := startRowIdx; rowIdx < endIdx; rowIdx++ {
		val := helpers.GetColValue(cols, filterColIdx, rowIdx)
		if val == filterVal {
			matchedRows = append(matchedRows, rowIdx)
		}
	}
	return matchedRows, nil
}

// filterRowsByWithinDays 时间范围过滤：保留单元格时间在 now ~ now+N天 的行
// 用于筛选"未来N天内"到期的数据行
// 返回 (matchedRows, error)；filterDays 非法或 filterCol 不存在时返回 error
func filterRowsByWithinDays(cols [][]string, opts FilterOptions) ([]int, error) {
	days, err := strconv.Atoi(strings.TrimSpace(opts.FilterDays))
	if err != nil {
		return nil, fmt.Errorf("filterMode=withinDays 时 filterDays 必须是正整数: %q", opts.FilterDays)
	}
	if days <= 0 {
		return nil, fmt.Errorf("filterMode=withinDays 时 filterDays 必须>0，当前=%d", days)
	}

	filterColIdx := helpers.GetColIndexByName(cols, opts.FilterColName)
	if filterColIdx < 0 {
		return nil, fmt.Errorf("过滤列 %q 在目标表中不存在", opts.FilterColName)
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	deadline := now.AddDate(0, 0, days)

	endIdx := helpers.GetDataEndIndex(cols, opts.StartRowIdx)

	matchedRows := []int{}
	for rowIdx := opts.StartRowIdx; rowIdx < endIdx; rowIdx++ {
		val := helpers.GetColValue(cols, filterColIdx, rowIdx)
		t := helpers.ParseDate(val)
		if t.IsZero() {
			continue // 解析失败，跳过
		}
		// 保留 now <= t <= deadline 的行（未来N天内）
		if (t.Equal(now) || t.After(now)) && (t.Equal(deadline) || t.Before(deadline)) {
			matchedRows = append(matchedRows, rowIdx)
		}
	}
	return matchedRows, nil
}

// FilterRowsByCondition 根据过滤条件筛选目标表中满足条件的行索引
// 委托给 FilterRowsByConditionEx，保持向后兼容
// 注意：此函数忽略错误，仅用于不需要区分"配置错误 vs 过滤无匹配"的旧路径
// 新代码请使用 FilterRowsByConditionEx 并处理 error 返回值
func FilterRowsByCondition(cols [][]string, filterColName, filterVal string, startRowIdx int, filterIsArray string) []int {
	rows, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: filterColName,
		FilterVal:     filterVal,
		StartRowIdx:   startRowIdx,
		FilterIsArray: filterIsArray,
	})
	return rows
}

// SplitArrayElements 按分隔符拆分字符串，但保留花括号内的内容不被拆分
// 例如 "{1;2},{3;4}" 用 "," 拆分 → ["{1;2}", "{3;4}"]
//
// 参数：
//   - s: 待拆分的字符串
//   - separator: 分隔符（仅支持单字符，通常为 ","）
//
// 返回：拆分后的字符串切片
func SplitArrayElements(s string, separator string) []string {
	if separator == "" {
		return []string{s}
	}

	var result []string
	var current strings.Builder
	braceDepth := 0

	for _, ch := range s {
		switch ch {
		case '{':
			braceDepth++
			current.WriteRune(ch)
		case '}':
			braceDepth--
			current.WriteRune(ch)
		case rune(separator[0]):
			if braceDepth == 0 && len(separator) == 1 {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}
