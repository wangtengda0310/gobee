// Package reference 提供引用关系相关的校验规则
// 本包中的规则用于检查单列数据的格式和有效性

package reference

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule/chain_reference"
	"github.com/xuri/excelize/v2"
)

// CrossReferenceCheckRule 跨表引用检查, 某个字段的值在另外一个表（当前sheet、当前表的另外一个sheet、或者另外一个表）的某个字段中
type CrossReferenceCheckRule struct{}

func (c *CrossReferenceCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否排除特定值
	var exceptVals []string
	if expt, ok := params[string(json_rule.EXCEPTS)]; ok {
		exceptVals = strings.Split(expt, ",")
	}

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	// 是否为数组模式（逗号分隔的多个值，每个值都需要单独检查）
	isArray := false
	if isArrayStr, ok := params["isArray"]; ok {
		isArray = strings.ToLower(isArrayStr) == "true"
	}

	// 行条件过滤：预先计算满足过滤条件的行集合，主循环中通过 map O(1) 查询
	var filteredRowSet map[int]bool
	if filterColName, ok := params[string(json_rule.FILTER_COL)]; ok && filterColName != "" {
		filterVal := strings.TrimSpace(params[string(json_rule.FILTER_VAL)])
		filterIsArray := params["filterIsArray"]
		filterMode := params["filterMode"]
		filterDays := params["filterDays"]

		// 验证过滤列是否存在
		filterColIdx := helpers.GetColIndexByName(cols, filterColName)
		if filterColIdx < 0 {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("条件过滤列不存在: %s", filterColName),
			}}
		}

		// 使用 chain_reference.FilterRowsByConditionEx 获取满足条件的行索引
		matchedRows, ferr := chain_reference.FilterRowsByConditionEx(cols, chain_reference.FilterOptions{
			FilterColName: filterColName,
			FilterVal:     filterVal,
			StartRowIdx:   startRowIdx,
			FilterIsArray: filterIsArray,
			FilterMode:    filterMode,
			FilterDays:    filterDays,
		})
		if ferr != nil {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("过滤条件错误: %v", ferr),
			}}
		}
		filteredRowSet = make(map[int]bool, len(matchedRows))
		for _, rowIdx := range matchedRows {
			filteredRowSet[rowIdx] = true
		}
	}

	// 获取参照表的整列数据
	refValues := make(map[string]bool) // 使用map提高查找效率
	refValueList := make([]string, 0)  // 参照值有序列表（用于遍历型比较操作）
	refColName := params["refCol"]     // 提前声明以便后续同行关联模式使用
	refSheetName := params["refSheet"]
	if refSheetName != "" {
		refFile, refFileExist := sheetMap[refSheetName]
		if !refFileExist {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("参照表不存在: %s", refSheetName),
				},
			}
		}

		// 校验参照列名称
		if refColName == "" {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: "必须指定参照列参数 'refCol'",
				},
			}
		}

		// 尝试获取参照表的所有列
		refCols, err := refFile.GetCols(refSheetName)
		if err != nil {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("无法读取参照表列数据: %v", err),
				},
			}
		}

		// 查找参照列
		var refColData []string
		var isEnumTable = strings.Contains(path.Base(refFile.Path), "_enum.xlsx")

		for _, col := range refCols {
			if len(col) == 0 {
				continue
			}

			// 检查是否是参照列
			var isTargetCol bool
			if isEnumTable {
				// 枚举表：第MJS_FIXED_ENUM_ROWS_CHS行是中文列名
				if len(col) > excelio.MJS_FIXED_ENUM_ROWS_CHS && col[excelio.MJS_FIXED_ENUM_ROWS_CHS] == refColName {
					isTargetCol = true
					refColData = col[excelio.MJS_FIXED_ENUM_ROWS_CHS:]
				}
			} else {
				// 普通表：第MJS_FIXED_ROWS_NAME行是列名
				if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == refColName {
					isTargetCol = true
					refColData = col[excelio.MJS_FIXED_ROWS_NUM:]
				}
			}

			if isTargetCol {
				// 将参照列的所有值添加到map中
				for _, value := range refColData {
					if value != "" {
						refValues[value] = true
						refValueList = append(refValueList, value)

						// 如果检查方式需要，也可以添加处理后的值
						if strings.Contains(params["matchType"], "ignorecase") {
							refValues[strings.ToLower(value)] = true
							refValues[strings.ToUpper(value)] = true
						}
					}
				}
				break
			}
		}

		if len(refValues) == 0 {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("在表%s中未找到列: %s", refSheetName, refColName),
				},
			}
		}
	} else {
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: "必须指定参照表参数 'refSheet'",
			},
		}
	}

	// 同行关联模式参数解析
	// matchCol: 当前表中用于关联参照表的列名，非空时启用同行关联模式
	// matchRefCol: 参照表中用于关联当前表的列名，默认等于 matchCol
	// 当 matchCol 非空时，只比较关联键对应行的参照值，而非全表多对多匹配
	matchCol := ""
	matchRefCol := ""
	matchColIdx := -1
	refValuesByRow := make(map[string][]string) // 关联键 → 参照值列表（同行关联模式用）
	if mc, ok := params["matchCol"]; ok && strings.TrimSpace(mc) != "" {
		matchCol = strings.TrimSpace(mc)
		matchRefCol = matchCol // 默认等于 matchCol
		if mrc, ok := params["matchRefCol"]; ok && strings.TrimSpace(mrc) != "" {
			matchRefCol = strings.TrimSpace(mrc)
		}

		// 查找当前表中 matchCol 列的索引
		for idx, col := range cols {
			if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == matchCol {
				matchColIdx = idx
				break
			}
		}
		if matchColIdx < 0 {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("同行关联列不存在: %s", matchCol),
			}}
		}

		// 构建 refValuesByRow：在参照表中找到 matchRefCol 列和 refCol 列，按行对齐
		refFile := sheetMap[refSheetName]
		refCols2, _ := refFile.GetCols(refSheetName)
		isEnum := strings.Contains(path.Base(refFile.Path), "_enum.xlsx")

		var matchRefColData []string  // matchRefCol 列数据
		var refColDataForRow []string // refCol 列数据
		nameRow := excelio.MJS_FIXED_ROWS_NAME
		dataStartRow := excelio.MJS_FIXED_ROWS_NUM
		if isEnum {
			nameRow = excelio.MJS_FIXED_ENUM_ROWS_CHS
			dataStartRow = excelio.MJS_FIXED_ENUM_ROWS_CHS
		}

		for _, col := range refCols2 {
			if len(col) == 0 {
				continue
			}
			if len(col) > nameRow {
				colName := col[nameRow]
				if colName == matchRefCol && matchRefColData == nil {
					matchRefColData = col[dataStartRow:]
				}
				if colName == refColName && refColDataForRow == nil {
					refColDataForRow = col[dataStartRow:]
				}
			}
			if matchRefColData != nil && refColDataForRow != nil {
				break
			}
		}

		if matchRefColData == nil {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("参照表中未找到关联列: %s", matchRefCol),
			}}
		}
		if refColDataForRow == nil {
			return []*json_rule.CellError{{
				Index:  0,
				Reason: fmt.Sprintf("参照表中未找到参照列: %s", refColName),
			}}
		}

		// 按行对齐构建映射（matchRefCol 不唯一时一个键对应多个参照值）
		maxLen := len(matchRefColData)
		if len(refColDataForRow) < maxLen {
			maxLen = len(refColDataForRow)
		}
		for rowIdx := 0; rowIdx < maxLen; rowIdx++ {
			key := strings.TrimSpace(matchRefColData[rowIdx])
			val := strings.TrimSpace(refColDataForRow[rowIdx])
			if key != "" && val != "" {
				refValuesByRow[key] = append(refValuesByRow[key], val)
			}
		}
	}

	// 检查方式：包含 不包含
	matchType := "exists" // 默认是存在性检查
	if mt, exist := params["matchType"]; exist && mt != "" {
		matchType = mt
	}

	// 是否不区分大小写
	ignoreCase := false
	if ignore, ok := params["ignoreCase"]; ok {
		ignoreCase = strings.ToLower(ignore) == "true"
	}

	// 是否要求完全匹配（针对多值情况）
	exactMatch := true
	if exact, ok := params["exactMatch"]; ok {
		exactMatch = strings.ToLower(exact) == "true"
	}

	// 比较操作（新增维度，为空时使用原有逻辑）
	compareOp := ""
	if op, ok := params["compareOp"]; ok && op != "" {
		compareOp = op
	}
	// 正则表达式提取模式（用于多值提取）
	var extractPattern *regexp.Regexp
	var captureGroups []int
	if patternStr, ok := params[string(json_rule.PATTERN)]; ok && patternStr != "" {
		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return []*json_rule.CellError{
				{
					Index:  0,
					Reason: fmt.Sprintf("提取正则表达式编译错误: %v", err),
				},
			}
		}
		extractPattern = pattern

		// 获取捕获组配置
		if groupsStr, ok := params[string(json_rule.GROUPS)]; ok && groupsStr != "" {
			groupStrs := strings.Split(groupsStr, ",")
			for _, groupStr := range groupStrs {
				groupStr = strings.TrimSpace(groupStr)
				if groupStr == "" {
					continue
				}
				if idx, err := strconv.Atoi(groupStr); err == nil && idx >= 0 {
					captureGroups = append(captureGroups, idx)
				}
			}
		} else {
			// 默认提取所有捕获组（不包括第0组）
			captureGroups = []int{1}
		}
	}

	// 数组模式：先按逗号拆分，再对每个元素应用正则提取
	var arraySeparator string
	if isArray {
		arraySeparator = ","
	}

	for i, s := range myColData {
		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) {
			continue
		}

		// 行条件过滤：行不在过滤结果集中则跳过
		if filteredRowSet != nil && !filteredRowSet[startRowIdx+i] {
			continue
		}

		if slices.Contains(exceptVals, s) {
			continue
		}

		// 提取要检查的值（单值或多值）
		var valuesToCheck []string

		// 数组模式：先按逗号拆分
		var elements []string
		if isArray && arraySeparator != "" {
			// 按逗号拆分，但保留花括号内的内容
			// 例如 "{1;2},{3;4}" -> ["{1;2}", "{3;4}"]
			elements = splitArrayElements(s, arraySeparator)
		} else {
			elements = []string{s}
		}

		// 对每个元素应用正则提取
		for _, element := range elements {
			element = strings.TrimSpace(element)
			if element == "" {
				continue
			}

			if extractPattern != nil {
				// 使用正则表达式提取多值
				matches := extractPattern.FindAllStringSubmatch(element, -1)
				if matches != nil {
					for _, match := range matches {
						for _, groupIdx := range captureGroups {
							if groupIdx < len(match) && match[groupIdx] != "" {
								valuesToCheck = append(valuesToCheck, match[groupIdx])
							}
						}
					}
				}

				if len(matches) == 0 {
					// 如果没有提取到值，尝试匹配整个字符串
					if extractPattern.MatchString(element) {
						// 如果有匹配但没有捕获组，使用整个匹配
						valuesToCheck = append(valuesToCheck, element)
					}
				}
			} else {
				// 单值检查
				valuesToCheck = append(valuesToCheck, element)
			}
		}

		// 根据检查方式验证
		isValid := false
		var foundValues []string
		var extractedValues []string
		extractedValues = valuesToCheck

		// 同行关联模式：通过关联键获取行级参照值
		if matchCol != "" {
			// 取当前行 matchCol 列的值作为关联键
			matchKey := strings.TrimSpace(helpers.GetColValue(cols, matchColIdx, startRowIdx+i))
			rowRefValues, keyExists := refValuesByRow[matchKey]
			if !keyExists || len(rowRefValues) == 0 {
				// 关联键在参照表中未找到
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason: fmt.Sprintf("值%s未通过跨表引用检查(关联键:%s=%s 未在参照表%s的%s列中找到)",
						s, matchCol, matchKey, refSheetName, matchRefCol),
				})
				continue
			}

			// 构建行级参照值的 map 和 list
			rowRefMap := make(map[string]bool, len(rowRefValues))
			for _, rv := range rowRefValues {
				rowRefMap[rv] = true
			}

			if compareOp != "" {
				// 同行关联 + 比较操作
				allValid := true
				for _, v := range valuesToCheck {
					v = strings.TrimSpace(v)
					if v == "" {
						continue
					}
					if ok, _ := compareWithOp(v, rowRefValues, rowRefMap, compareOp, ignoreCase); ok {
						foundValues = append(foundValues, v)
					} else {
						allValid = false
						break
					}
				}
				isValid = allValid
			} else {
				// 同行关联 + 精确匹配（exists 语义）
				allValid := true
				for _, v := range valuesToCheck {
					v = strings.TrimSpace(v)
					if v == "" {
						continue
					}
					if rowRefMap[v] {
						foundValues = append(foundValues, v)
					} else {
						allValid = false
						break
					}
				}
				isValid = allValid
			}
		} else if compareOp != "" {
			// 比较操作模式（全表多对多）：使用 compareWithOp 替代 map 精确查找
			// 语义与 matchType=exists 一致：匹配成功 → 通过，匹配失败 → 报错
			allValid := true
			for _, v := range valuesToCheck {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				if ok, _ := compareWithOp(v, refValueList, refValues, compareOp, ignoreCase); ok {
					foundValues = append(foundValues, v)
				} else {
					allValid = false
					break
				}
			}
			isValid = allValid
		} else {
			// 原有逻辑（全表多对多 + matchType 分支）
			switch matchType {
			case "exists":
				// 存在性检查：值是否在参照列中
				if exactMatch {
					// 精确匹配：所有值都必须在参照列中
					allValid := true
					for _, v := range valuesToCheck {
						v = strings.TrimSpace(v)
						if v == "" {
							continue
						}

						checkValue := v
						if ignoreCase {
							checkValue = strings.ToLower(v)
						}

						if refValues[checkValue] || (ignoreCase && (refValues[strings.ToUpper(v)] || refValues[v])) {
							foundValues = append(foundValues, v)
						} else {
							allValid = false
							break
						}
					}
					isValid = allValid
				} else {
					// 非精确匹配：至少有一个值在参照列中
					for _, v := range valuesToCheck {
						v = strings.TrimSpace(v)
						if v == "" {
							continue
						}

						checkValue := v
						if ignoreCase {
							checkValue = strings.ToLower(v)
						}

						if refValues[checkValue] || (ignoreCase && (refValues[strings.ToUpper(v)] || refValues[v])) {
							foundValues = append(foundValues, v)
							isValid = true
							break
						}
					}
				}

			case "not_exists":
				// 反向检查：所有值都不能在参照列中
				allNotExist := true
				for _, v := range valuesToCheck {
					v = strings.TrimSpace(v)
					if v == "" {
						continue
					}

					checkValue := v
					if ignoreCase {
						checkValue = strings.ToLower(v)
					}

					if refValues[checkValue] || (ignoreCase && (refValues[strings.ToUpper(v)] || refValues[v])) {
						allNotExist = false
						foundValues = append(foundValues, v)
						break
					}
				}
				isValid = allNotExist

			default:
				// 默认按exists处理

			}
		}

		if !isValid {
			var reason string
			if matchCol != "" {
				// 同行关联模式错误信息
				matchKey := strings.TrimSpace(helpers.GetColValue(cols, matchColIdx, startRowIdx+i))
				if compareOp != "" {
					reason = fmt.Sprintf("值%s未通过跨表引用检查(规则:%s, 比较:%s, 关联键:%s=%s)", s, matchType, compareOp, matchCol, matchKey)
				} else {
					reason = fmt.Sprintf("值%s未通过跨表引用检查(规则:%s, 关联键:%s=%s)", s, matchType, matchCol, matchKey)
				}
				if len(foundValues) > 0 {
					reason += fmt.Sprintf(", 已找到:%s", strings.Join(foundValues, ", "))
				}
			} else if compareOp != "" {
				reason = fmt.Sprintf("值%s未通过跨表引用检查(规则:%s, 比较:%s)", s, matchType, compareOp)
				if len(foundValues) > 0 {
					reason = fmt.Sprintf("值%s未通过跨表引用检查(规则:%s, 比较:%s, 已找到:%s)", s, matchType, compareOp, strings.Join(foundValues, ", "))
				}
			} else {
				reason = fmt.Sprintf("值%s未通过跨表引用检查(规则:%s)", s, matchType)

				if len(extractedValues) > 0 {
					reason = fmt.Sprintf("提取的值[%s]未通过跨表引用检查(规则:%s)",
						strings.Join(extractedValues, ", "), matchType)
				}

				if len(foundValues) > 0 {
					reason = fmt.Sprintf("值%s未找到匹配的参照值(规则:%s, 已找到:%s)",
						s, matchType, strings.Join(foundValues, ", "))
				}
			}

			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
		}
	}

	return res
}

// compareWithOp 使用指定比较操作检查值是否匹配参照值
// cellValue: 当前单元格值, refValueList: 参照值列表, refValues: 参照值map(O(1)查找用)
// compareOp: 比较操作, ignoreCase: 是否忽略大小写
// 返回: (是否匹配, 匹配到的参照值)
func compareWithOp(cellValue string, refValueList []string, refValues map[string]bool, compareOp string, ignoreCase bool) (bool, string) {
	cellValue = strings.TrimSpace(cellValue)
	if cellValue == "" {
		return false, ""
	}

	switch compareOp {
	case "exact_match":
		// 使用 map 查找（O(1)）
		checkValue := cellValue
		if ignoreCase {
			checkValue = strings.ToLower(cellValue)
		}
		if refValues[checkValue] || (ignoreCase && (refValues[strings.ToUpper(cellValue)] || refValues[cellValue])) {
			return true, checkValue
		}
		return false, ""

	case "contains":
		// 参照值包含在当前值中：strings.Contains(cellValue, refValue)
		cell := cellValue
		if ignoreCase {
			cell = strings.ToLower(cellValue)
		}
		for _, rv := range refValueList {
			ref := strings.TrimSpace(rv)
			if ref == "" {
				continue
			}
			if ignoreCase {
				ref = strings.ToLower(rv)
			}
			if strings.Contains(cell, ref) {
				return true, rv // 返回原始参照值
			}
		}
		return false, ""

	case "date_equals":
		// 秒级日期相同比较
		cellTime := helpers.ParseDate(cellValue)
		if cellTime.IsZero() {
			return false, ""
		}
		cellKey := cellTime.Format("2006-01-02 15:04:05")
		for _, rv := range refValueList {
			refTime := helpers.ParseDate(strings.TrimSpace(rv))
			if !refTime.IsZero() && refTime.Format("2006-01-02 15:04:05") == cellKey {
				return true, rv
			}
		}
		return false, ""

	case "date_before_or_equal":
		// 当前值日期早于或等于参照值日期
		cellTime := helpers.ParseDate(cellValue)
		if cellTime.IsZero() {
			return false, ""
		}
		for _, rv := range refValueList {
			refTime := helpers.ParseDate(strings.TrimSpace(rv))
			if !refTime.IsZero() && (cellTime.Before(refTime) || cellTime.Equal(refTime)) {
				return true, rv
			}
		}
		return false, ""

	case "date_after_or_equal":
		// 当前值日期晚于或等于参照值日期
		cellTime := helpers.ParseDate(cellValue)
		if cellTime.IsZero() {
			return false, ""
		}
		for _, rv := range refValueList {
			refTime := helpers.ParseDate(strings.TrimSpace(rv))
			if !refTime.IsZero() && (cellTime.After(refTime) || cellTime.Equal(refTime)) {
				return true, rv
			}
		}
		return false, ""

	default:
		return false, ""
	}
}

// splitArrayElements 按分隔符拆分字符串，但保留花括号内的内容
// 例如 "{1;2},{3;4}" 按 "," 拆分得到 ["{1;2}", "{3;4}"]
func splitArrayElements(s string, separator string) []string {
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
				// 只有在花括号外才拆分
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
