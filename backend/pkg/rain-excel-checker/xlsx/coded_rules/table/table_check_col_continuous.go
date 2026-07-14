package coded_rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ColContinuousCheckRule 列连续性与唯一性检查规则
//
// ## 校验规则
// 检查指定列中数据的连续性或唯一性，支持六种模式：
// 1. INCREASE_STRICT: 严格递增（每个值 = 上一个值 + 1）
// 2. INCREASE_MONOTONE: 单调递增（允许跳跃）
// 3. DATE_CONTINUOUS: 日期间隔一致
// 4. ID_FORMAT_CONTINUOUS: {id;count} 格式ID递增
// 5. EXTRACT_NUMBER_STRICT: 从文本中提取数字后严格递增（如 "第2赛季" → 2）
// 6. SPLIT_UNIQUE: 按分隔符拆分后全局唯一性检查（如 byproduct 中的 ID 不能跨行重复）
//
// ## 相关表结构
// - 通用规则，适用于所有包含需要连续性/唯一性检查的列的表
//
// ## 检查流程
// 1. 解析参数（目标列、检查模式、范围等）
// 2. 查找目标列索引
// 3. 收集数据（跳过空值、注释行、排除行）
// 4. 按检查模式执行校验
// 5. 汇总错误并返回
type ColContinuousCheckRule struct{}

// 检查模式常量
const (
	CheckModeIncreaseStrict     = "INCREASE_STRICT"
	CheckModeIncreaseMonotone   = "INCREASE_MONOTONE"
	CheckModeDateContinuous     = "DATE_CONTINUOUS"
	CheckModeIdFormatContinuous = "ID_FORMAT_CONTINUOUS"
	CheckModeExtractNumber      = "EXTRACT_NUMBER_STRICT"
	CheckModeSplitUnique        = "SPLIT_UNIQUE"
	CheckModeDateMonthly        = "DATE_MONTHLY_PATTERN"
)

// extractNumberRe 从文本中提取第一个连续数字序列
var extractNumberRe = regexp.MustCompile(`\d+`)

// 参数 key 使用 ERuleParam 枚举（确保前端绑定能正确序列化）
const (
	paramTargetCol   = string(json_rule.TARGET_COL)
	paramCheckMode   = string(json_rule.CHECK_MODE)
	paramScope       = string(json_rule.SCOPE)
	paramStartValue  = string(json_rule.START_VALUE)
	paramTolerance   = string(json_rule.TOLERANCE)
	paramExcludeRows = string(json_rule.EXCLUDE_ROWS)
	paramSeparator   = string(json_rule.SEPARATOR)
)

// collectedRow 收集到的有效数据行
type collectedRow struct {
	rowIdx int    // 原始行索引（用于错误定位）
	value  string // 原始值字符串
}

// Meta 返回规则元数据
func (c *ColContinuousCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.COL_CONTINUOUS_CHECK,
		DisplayName:  "列连续性检查",
		Description:  "检查指定列的数据连续性或唯一性。支持7种模式：严格递增、单调递增、日期间隔一致、{id;count}格式ID递增、提取数字递增、拆分后全局唯一、日期月模式（月顺延、日/时一致）。支持排除行号、注释行跳过、空值跳过。",
		TargetSheets: []string{}, // 空 = 适用于所有表
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.TARGET_COL,
				Type:        "string",
				Label:       "检查列名",
				Description: "要检查连续性的列字段名（如 SeasonIndex、Id）",
				Required:    true,
			},
			{
				Key:         json_rule.CHECK_MODE,
				Type:        "select",
				Label:       "检查模式",
				Description: "选择连续性检查模式",
				Default:     CheckModeIncreaseStrict,
				Options: []json_rule.SelectOption{
					{Label: "严格递增（+1）", Value: CheckModeIncreaseStrict},
					{Label: "单调递增（允许跳跃）", Value: CheckModeIncreaseMonotone},
					{Label: "日期间隔一致", Value: CheckModeDateContinuous},
					{Label: "{id;count}格式ID递增", Value: CheckModeIdFormatContinuous},
					{Label: "提取数字严格递增（如\"第2赛季\"→2）", Value: CheckModeExtractNumber},
					{Label: "拆分后全局唯一（如\"1,2,3\"跨行不重复）", Value: CheckModeSplitUnique},
					{Label: "日期月模式（月顺延、日/时一致）", Value: CheckModeDateMonthly},
				},
			},
			{
				Key:         json_rule.SCOPE,
				Type:        "string",
				Label:       "检查范围",
				Description: "added(仅新增行) 或 all(全量检查)，默认 all",
				Default:     "all",
			},
			{
				Key:         json_rule.START_VALUE,
				Type:        "string",
				Label:       "起始值",
				Description: "期望的起始值（不填则从数据的第一个值开始检查）",
			},
			{
				Key:         json_rule.TOLERANCE,
				Type:        "number",
				Label:       "容差",
				Description: "日期模式的容差天数（默认0，仅 DATE_CONTINUOUS 有效）",
				Default:     "0",
			},
			{
				Key:         json_rule.ALLOW_EMPTY,
				Type:        "string",
				Label:       "允许空值",
				Description: "是否跳过空值单元格（默认 true）",
				Default:     "true",
			},
			{
				Key:         json_rule.ALLOW_COMMIT,
				Type:        "string",
				Label:       "允许注释行",
				Description: "是否跳过以 # 开头的注释行（默认 false）",
				Default:     "false",
			},
			{
				Key:         json_rule.EXCLUDE_ROWS,
				Type:        "string",
				Label:       "排除行号",
				Description: "排除的行号（数据行号，从1开始），多个用逗号分隔，支持范围（如 \"3,7\" 或 \"3-5\"）",
			},
			{
				Key:         json_rule.SEPARATOR,
				Type:        "string",
				Label:       "拆分分隔符",
				Description: "SPLIT_UNIQUE 模式的分隔符（默认逗号 \",\"）",
				Default:     ",",
			},
		},
	}
}

// Check 执行列连续性与唯一性检查
func (c *ColContinuousCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "列连续性检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 1. 解析参数
	targetCol := param.Params[paramTargetCol]
	if targetCol == "" {
		result.Ok = false
		result.Reason = "缺少必填参数: targetCol（检查列名）"
		return result
	}

	checkMode := param.Params[paramCheckMode]
	if checkMode == "" {
		checkMode = CheckModeIncreaseStrict
	}

	allowEmpty := true
	if v, ok := param.Params[string(json_rule.ALLOW_EMPTY)]; ok && strings.ToLower(v) == "false" {
		allowEmpty = false
	}

	allowCommit := false
	if v, ok := param.Params[string(json_rule.ALLOW_COMMIT)]; ok && strings.ToLower(v) == "true" {
		allowCommit = true
	}

	tolerance := 0
	if v, ok := param.Params[paramTolerance]; ok && v != "" {
		if t, err := strconv.Atoi(v); err == nil && t >= 0 {
			tolerance = t
		}
	}

	separator := ","
	if v, ok := param.Params[paramSeparator]; ok && v != "" {
		separator = v
	}

	excludeSet := parseExcludeRows(param.Params[paramExcludeRows])

	// 2. 查找目标列
	colIdx := helpers.GetColIndexByName(param.Cols, targetCol)
	if colIdx == -1 {
		result.Ok = false
		result.Reason = fmt.Sprintf("未找到列: %s", targetCol)
		return result
	}

	// 3. 收集有效数据行
	rows := collectRows(param.Cols, colIdx, param.StartRowIdx, param.EndIndex, allowEmpty, allowCommit, excludeSet)

	if len(rows) == 0 {
		return result // 无数据，直接通过
	}

	// 4. 解析 startValue
	var startValue *int
	if sv := param.Params[paramStartValue]; sv != "" {
		if v, err := strconv.Atoi(sv); err == nil {
			startValue = &v
		}
	}

	// 5. 按模式执行检查
	switch checkMode {
	case CheckModeIncreaseStrict:
		c.checkIncreaseStrict(result, rows, startValue, targetCol)
	case CheckModeIncreaseMonotone:
		c.checkIncreaseMonotone(result, rows, startValue, targetCol)
	case CheckModeDateContinuous:
		c.checkDateContinuous(result, rows, tolerance, targetCol)
	case CheckModeIdFormatContinuous:
		c.checkIdFormatContinuous(result, rows, startValue, targetCol)
	case CheckModeExtractNumber:
		c.checkExtractNumber(result, rows, startValue, targetCol)
	case CheckModeSplitUnique:
		c.checkSplitUnique(result, rows, separator, targetCol)
	case CheckModeDateMonthly:
		c.checkDateMonthly(result, rows, targetCol)
	default:
		result.Ok = false
		result.Reason = fmt.Sprintf("未知的检查模式: %s", checkMode)
	}

	return result
}

// parseExcludeRows 解析排除行号参数
// 格式: "3,7" 或 "3-5"，返回数据行号集合（从1开始的数据行号）
func parseExcludeRows(excludeStr string) map[int]bool {
	result := make(map[int]bool)
	if excludeStr == "" {
		return result
	}

	parts := strings.Split(excludeStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						result[i] = true
					}
				}
			}
		} else {
			if n, err := strconv.Atoi(part); err == nil {
				result[n] = true
			}
		}
	}
	return result
}

// collectRows 收集有效数据行
// excludeSet 中的 key 是数据行号（从1开始），需要转换为 rowIdx 来判断
func collectRows(cols [][]string, colIdx, startRowIdx, endIndex int, allowEmpty, allowCommit bool, excludeSet map[int]bool) []collectedRow {
	rows := make([]collectedRow, 0)

	for rowIdx := startRowIdx; rowIdx < endIndex; rowIdx++ {
		value := helpers.GetColValue(cols, colIdx, rowIdx)

		// 数据行号 = rowIdx - startRowIdx + 1（从1开始）
		dataRowNum := rowIdx - startRowIdx + 1

		// 排除指定行号
		if excludeSet[dataRowNum] {
			continue
		}

		// 跳过空值
		if value == "" && allowEmpty {
			continue
		}

		// 跳过注释行
		if allowCommit && strings.HasPrefix(value, "#") {
			continue
		}

		rows = append(rows, collectedRow{rowIdx: rowIdx, value: value})
	}

	return rows
}

// checkIncreaseStrict 检查严格递增（每个值 = 上一个值 + 1）
func (c *ColContinuousCheckRule) checkIncreaseStrict(result *json_rule.TableCheckResult, rows []collectedRow, startValue *int, colName string) {
	prevVal := 0
	if startValue != nil {
		prevVal = *startValue
	} else if len(rows) > 0 {
		if v, err := strconv.Atoi(rows[0].value); err == nil {
			prevVal = v
		} else {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  rows[0].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为整数: %s", rows[0].rowIdx, colName, rows[0].value),
			})
			result.Ok = false
			return
		}
	}

	startIdx := 0
	if startValue == nil {
		startIdx = 1 // 无 startValue 时跳过第一行
	}

	for i := startIdx; i < len(rows); i++ {
		curVal, err := strconv.Atoi(rows[i].value)
		if err != nil {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  rows[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为整数: %s", rows[i].rowIdx, colName, rows[i].value),
			})
			continue
		}

		expected := prevVal + 1
		if curVal != expected {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: rows[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值为 %d，期望为 %d（严格递增，上一个值为 %d）",
					rows[i].rowIdx, colName, curVal, expected, prevVal),
			})
		}
		prevVal = curVal
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个连续性问题", len(result.ErrCells))
	}
}

// checkIncreaseMonotone 检查单调递增（允许跳跃）
func (c *ColContinuousCheckRule) checkIncreaseMonotone(result *json_rule.TableCheckResult, rows []collectedRow, startValue *int, colName string) {
	prevVal := 0
	if startValue != nil {
		prevVal = *startValue
	} else if len(rows) > 0 {
		if v, err := strconv.Atoi(rows[0].value); err == nil {
			prevVal = v
		} else {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  rows[0].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为整数: %s", rows[0].rowIdx, colName, rows[0].value),
			})
			result.Ok = false
			return
		}
	}

	startIdx := 0
	if startValue == nil {
		startIdx = 1
	}

	for i := startIdx; i < len(rows); i++ {
		curVal, err := strconv.Atoi(rows[i].value)
		if err != nil {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  rows[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为整数: %s", rows[i].rowIdx, colName, rows[i].value),
			})
			continue
		}

		if curVal <= prevVal {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: rows[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值为 %d，必须大于上一个值 %d（单调递增）",
					rows[i].rowIdx, colName, curVal, prevVal),
			})
		}
		prevVal = curVal
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个连续性问题", len(result.ErrCells))
	}
}

// checkDateContinuous 检查日期间隔一致性
func (c *ColContinuousCheckRule) checkDateContinuous(result *json_rule.TableCheckResult, rows []collectedRow, tolerance int, colName string) {
	if len(rows) < 2 {
		return // 数据不足，跳过检查
	}

	// 解析所有日期
	type dateEntry struct {
		rowIdx int
		date   time.Time
	}
	dates := make([]dateEntry, 0, len(rows))

	for _, row := range rows {
		t := excelio.ParseDate(row.value)
		if t.IsZero() {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  row.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为日期: %s", row.rowIdx, colName, row.value),
			})
			continue
		}
		dates = append(dates, dateEntry{rowIdx: row.rowIdx, date: t})
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个日期解析错误", len(result.ErrCells))
		return
	}

	if len(dates) < 2 {
		return // 有效日期不足2个
	}

	// 计算标准间隔（前两个日期的差值）
	expectedDuration := dates[1].date.Sub(dates[0].date)
	if expectedDuration == 0 {
		return // 两个日期相同，无法确定间隔
	}

	toleranceDuration := time.Duration(tolerance) * 24 * time.Hour

	for i := 2; i < len(dates); i++ {
		actualDuration := dates[i].date.Sub(dates[i-1].date)
		diff := actualDuration - expectedDuration
		if diff < 0 {
			diff = -diff
		}

		if diff > toleranceDuration {
			expectedDays := int(expectedDuration.Hours() / 24)
			actualDays := int(actualDuration.Hours() / 24)
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: dates[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 与上一行日期间隔 %d 天，期望为 %d 天（日期连续，容差 %d 天）",
					dates[i].rowIdx, colName, actualDays, expectedDays, tolerance),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个日期间隔不一致问题", len(result.ErrCells))
	}
}

// checkDateMonthly 检查日期月模式（月顺延、日一致、时一致）
// 适用于 ArenaSeason 等周期性配置，如每月15号 00:00:00 更新
// 检查规则：
// 1. 月份逐行递增（允许跨年，如 12月 → 1月）
// 2. 所有行的"日"必须相同（如都是15号）
// 3. 所有行的"时:分:秒"必须相同（如都是 00:00:00）
func (c *ColContinuousCheckRule) checkDateMonthly(result *json_rule.TableCheckResult, rows []collectedRow, colName string) {
	if len(rows) < 2 {
		return // 数据不足，跳过检查
	}

	// 解析所有日期
	type dateEntry struct {
		rowIdx int
		date   time.Time
	}
	dates := make([]dateEntry, 0, len(rows))

	for _, row := range rows {
		t := excelio.ParseDate(row.value)
		if t.IsZero() {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  row.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析为日期: %s", row.rowIdx, colName, row.value),
			})
			continue
		}
		dates = append(dates, dateEntry{rowIdx: row.rowIdx, date: t})
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个日期解析错误", len(result.ErrCells))
		return
	}

	if len(dates) < 2 {
		return // 有效日期不足2个
	}

	// 以第一行为基准，提取日、时、分、秒
	baseDay := dates[0].date.Day()
	baseHour := dates[0].date.Hour()
	baseMin := dates[0].date.Minute()
	baseSec := dates[0].date.Second()

	// 检查所有行的日、时、分、秒是否与基准一致
	for i := 1; i < len(dates); i++ {
		d := dates[i].date
		if d.Day() != baseDay {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: dates[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 日期为 %d 号，期望为 %d 号（日期月模式，日必须一致）",
					dates[i].rowIdx, colName, d.Day(), baseDay),
			})
		}
		if d.Hour() != baseHour || d.Minute() != baseMin || d.Second() != baseSec {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: dates[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 时间为 %02d:%02d:%02d，期望为 %02d:%02d:%02d（日期月模式，时间必须一致）",
					dates[i].rowIdx, colName, d.Hour(), d.Minute(), d.Second(), baseHour, baseMin, baseSec),
			})
		}
	}

	// 检查月份是否逐行递增
	prevYear := dates[0].date.Year()
	prevMonth := int(dates[0].date.Month())
	for i := 1; i < len(dates); i++ {
		year := dates[i].date.Year()
		month := int(dates[i].date.Month())

		// 计算期望的月份（考虑跨年）
		expectedYear := prevYear
		expectedMonth := prevMonth + 1
		if expectedMonth > 12 {
			expectedYear++
			expectedMonth = 1
		}

		if year != expectedYear || month != expectedMonth {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: dates[i].rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 月份为 %d年%d月，期望为 %d年%d月（日期月模式，月份应逐月递增）",
					dates[i].rowIdx, colName, year, month, expectedYear, expectedMonth),
			})
		}
		prevYear = year
		prevMonth = month
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个日期月模式问题", len(result.ErrCells))
	}
}

// checkIdFormatContinuous 检查 {id;count} 格式ID递增
func (c *ColContinuousCheckRule) checkIdFormatContinuous(result *json_rule.TableCheckResult, rows []collectedRow, startValue *int, colName string) {
	type idRow struct {
		rowIdx int
		id     int
		raw    string
	}
	idRows := make([]idRow, 0, len(rows))

	for _, row := range rows {
		items := helpers.ParseItemCfg(row.value)
		if len(items) == 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  row.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值无法解析 {id;count} 格式: %s", row.rowIdx, colName, row.value),
			})
			continue
		}
		idRows = append(idRows, idRow{rowIdx: row.rowIdx, id: items[0].ItemId, raw: row.value})
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个格式解析错误", len(result.ErrCells))
		return
	}

	if len(idRows) == 0 {
		return
	}

	// 转换为 collectedRow 格式，复用严格递增检查
	convertedRows := make([]collectedRow, len(idRows))
	for i, ir := range idRows {
		convertedRows[i] = collectedRow{rowIdx: ir.rowIdx, value: strconv.Itoa(ir.id)}
	}

	c.checkIncreaseStrict(result, convertedRows, startValue, colName)

	// 补充格式信息到错误消息
	for _, errCell := range result.ErrCells {
		for _, ir := range idRows {
			if ir.rowIdx == errCell.Index {
				errCell.Reason = fmt.Sprintf("第 %d 行的 %s 格式 %s 中 ID 为 %d（ID格式连续递增，期望 %s）",
					ir.rowIdx, colName, ir.raw, ir.id, errCell.Reason)
				break
			}
		}
	}
}

// checkExtractNumber 从文本中提取数字后检查严格递增
// 适用于 "第2赛季"、"S3"、"版本10" 等混合文本格式
// 使用正则提取第一个连续数字序列，提取不到则报错
func (c *ColContinuousCheckRule) checkExtractNumber(result *json_rule.TableCheckResult, rows []collectedRow, startValue *int, colName string) {
	type extractedRow struct {
		rowIdx int
		num    int
		raw    string
	}
	extracted := make([]extractedRow, 0, len(rows))

	for _, row := range rows {
		match := extractNumberRe.FindString(row.value)
		if match == "" {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  row.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值 \"%s\" 中未找到数字", row.rowIdx, colName, row.value),
			})
			continue
		}
		num, err := strconv.Atoi(match)
		if err != nil {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:  row.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值 \"%s\" 中提取的数字 \"%s\" 解析失败", row.rowIdx, colName, row.value, match),
			})
			continue
		}
		extracted = append(extracted, extractedRow{rowIdx: row.rowIdx, num: num, raw: row.value})
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个无法提取数字的行", len(result.ErrCells))
		return
	}

	if len(extracted) == 0 {
		return
	}

	// 确定起始值
	prevVal := 0
	if startValue != nil {
		prevVal = *startValue
	} else {
		prevVal = extracted[0].num
	}

	startIdx := 0
	if startValue == nil {
		startIdx = 1
	}

	for i := startIdx; i < len(extracted); i++ {
		cur := extracted[i]
		expected := prevVal + 1
		if cur.num != expected {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index: cur.rowIdx,
				Reason: fmt.Sprintf("第 %d 行的 %s 值 \"%s\" 提取数字为 %d，期望为 %d（上一个为 %d）",
					cur.rowIdx, colName, cur.raw, cur.num, expected, prevVal),
			})
		}
		prevVal = cur.num
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个连续性问题", len(result.ErrCells))
	}
}

// checkSplitUnique 按分隔符拆分后检查全局唯一性
// 适用于 "1022201,1040010" 格式，拆分后的每个值在整列所有行中必须唯一
func (c *ColContinuousCheckRule) checkSplitUnique(result *json_rule.TableCheckResult, rows []collectedRow, separator string, colName string) {
	// seen 记录每个值第一次出现的行索引
	seen := make(map[string]int)

	for _, row := range rows {
		parts := strings.Split(row.value, separator)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if firstRowIdx, exists := seen[part]; exists {
				result.ErrCells = append(result.ErrCells, &json_rule.CellError{
					Index: row.rowIdx,
					Reason: fmt.Sprintf("第 %d 行的 %s 值 \"%s\" 重复，首次出现在第 %d 行",
						row.rowIdx, colName, part, firstRowIdx),
				})
			} else {
				seen[part] = row.rowIdx
			}
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个重复值", len(result.ErrCells))
	}
}
