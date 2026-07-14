// Package calculation 提供计算类列校验规则（权重和、日期一致性等）
// 本包中的规则用于检查涉及多列计算逻辑的数据一致性

package calculation

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// DateConsistencyCheckRule 活动时间一致性检查规则
type DateConsistencyCheckRule struct{}

func (c *DateConsistencyCheckRule) Check(sheetName string, cols [][]string, colIdx, startRowIdx int, params map[string]string, sheetMap map[string]*excelize.File) []*json_rule.CellError {
	breakLine := helpers.ParseBreakLine(params)

	endIdx := helpers.GetColEndIndex(cols, colIdx, startRowIdx, breakLine, params)
	myColData := cols[colIdx][startRowIdx:endIdx]
	res := make([]*json_rule.CellError, 0, len(myColData))

	var descSheetCols [][]string = nil
	if sheetName_, exist := params["descSheet"]; exist && sheetName_ != "" {
		if xlsx, exist := sheetMap[sheetName_]; exist {
			if sheetCols, err := xlsx.GetCols(sheetName_); err == nil {
				descSheetCols = sheetCols
			}
		}
	}

	// 获取活动描述列索引
	var descData = make([]string, 0, len(myColData))
	if descColAttrName, exist := params["descColAttrName"]; exist {
		if descSheetCols == nil {
			// 在本表查询
			for _, col := range cols {
				if len(col) > 2 && col[2] == descColAttrName {
					descData = col[startRowIdx:]
					break
				}
			}
		} else {
			// 在他表查询
			for _, col := range descSheetCols {
				if len(col) > 2 && col[2] == descColAttrName {
					descData = col[startRowIdx:]
					break
				}
			}
		}
	} else {
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("活动描述列索引名称为空: %s", params["descColAttrName"]),
			},
		}
	}

	if len(descData) == 0 {
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("未找到描述列: %s", params["descColAttrName"]),
			},
		}
	}

	// 获取活动配置的开始时间列索引
	startTimeData := myColData

	// 获取结束日期列的索引
	endColOffset, err := strconv.Atoi(params["endColOffset"])
	if err != nil || endColOffset == 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("结束日期列索引参数错误: %s", params["endColOffset"]),
			},
		}
	}
	if len(cols) < colIdx+endColOffset || colIdx+endColOffset < 0 {
		// 参数错误
		return []*json_rule.CellError{
			{
				Index:  0,
				Reason: fmt.Sprintf("结束日期列索引参数错误: 越界[%d], 实际:(%d)", colIdx+endColOffset, len(cols)),
			},
		}
	}
	endData := cols[colIdx+endColOffset][startRowIdx:]

	// 支持的日期格式
	dateFormats := []string{
		// 完整日期时间格式
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-1-2 15:4:5",
		"2006/1/2 15:4:5",
		"2006年01月02日 15时04分05秒",
		"2006年1月2日 15时4分5秒",
		"2006年01月02日 15:04:05",
		"2006年1月2日 15:4:5",

		// 日期+小时分钟
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006-1-2 15:4",
		"2006/1/2 15:4",
		"2006年01月02日 15时04分",
		"2006年1月2日 15时4分",
		"2006年01月02日 15:04",
		"2006年1月2日 15:4",

		// 日期+小时
		"2006-01-02 15点",
		"2006/01/02 15点",
		"2006-1-2 15点",
		"2006/1/2 15点",
		"2006年01月02日 15点",
		"2006年1月2日 15点",

		// 纯日期格式
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
		"20060102",
		"2006年01月02日",
		"2006年1月2日",

		// 月日格式（不带年份）
		"01月02日",
		"1月2日",
		"01-02",
		"1-2",
		"01/02",
		"1/2",
	}

	// 是否严格匹配（默认false，允许描述中的时间与配置时间不完全一致）
	strictMatch := false
	if strict, ok := params[string(json_rule.STRICT)]; ok {
		strictMatch = strings.ToLower(strict) == "true"
	}

	// 是否允许配置时间为空时跳过检查
	skipIfConfigEmpty := true
	if skip, ok := params[string(json_rule.ALLOW_EMPTY)]; ok {
		skipIfConfigEmpty = strings.ToLower(skip) == "true"
	}

	// 是否自动补充年份（当月日格式没有年份时）
	autoCompleteYear := true
	if autoYear, ok := params["autoCompleteYear"]; ok {
		autoCompleteYear = strings.ToLower(autoYear) == "true"
	}

	// 默认年份（当月日格式没有年份时使用）
	defaultYear := time.Now().Year()
	if yearStr, ok := params["defaultYear"]; ok {
		if year, err := strconv.Atoi(yearStr); err == nil {
			defaultYear = year
		}
	}

	// 是否允许空值
	allowEmpty := helpers.ParseAllowEmpty(params)

	// 是否允许注释
	allowCommit := helpers.ParseAllowCommit(params)

	for i := range myColData {
		configStart := startTimeData[i] // myColData
		configEnd := endData[i]
		desc := descData[i]

		// 处理空值和注释
		if helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx, i, allowEmpty, allowCommit) || helpers.SolveEmptyAndCommit(&res, cols, startRowIdx, colIdx+endColOffset, i, allowEmpty, allowCommit) {
			continue
		}

		// 如果配置时间为空且允许跳过
		if skipIfConfigEmpty && (strings.TrimSpace(configStart) == "" || strings.TrimSpace(configEnd) == "") {
			continue
		}

		// 解析配置时间
		configStartTime, err1 := helpers.ParseDateWithFormats(configStart, dateFormats)
		configEndTime, err2 := helpers.ParseDateWithFormats(configEnd, dateFormats)

		if err1 != nil || err2 != nil {
			if err1 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("配置开始时间格式错误: %s", configStart),
				})
			}
			if err2 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("配置结束时间格式错误: %s", configEnd),
				})
			}
			continue
		}

		// 从描述中提取所有时间字符串
		timeStrings := extractAllTimeStrings(desc)

		if len(timeStrings) < 2 {
			if reportNoTime, ok := params["reportNoTimeInDesc"]; ok && strings.ToLower(reportNoTime) == "true" {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("活动描述中没有找到足够的时间信息(需要2个, 找到%d个): %s", len(timeStrings), desc),
				})
			}
			continue
		}

		// 取前两个时间作为开始和结束时间
		descStartStr := timeStrings[0]
		descEndStr := timeStrings[1]

		// 尝试解析时间
		descStartTime, err1 := parseFlexibleTime(descStartStr, defaultYear, autoCompleteYear)
		descEndTime, err2 := parseFlexibleTime(descEndStr, defaultYear, autoCompleteYear)

		if err1 != nil || err2 != nil {
			if err1 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("描述开始时间解析失败: %s (原始: %s)", err1, descStartStr),
				})
			}
			if err2 != nil {
				res = append(res, &json_rule.CellError{
					Index:    i,
					ExcelRow: startRowIdx + i + 1,
					Reason:   fmt.Sprintf("描述结束时间解析失败: %s (原始: %s)", err2, descEndStr),
				})
			}
			continue
		}

		// 检查时间顺序（开始时间应该在结束时间之前）
		if descStartTime.After(descEndTime) {
			// 可能顺序反了，交换一下
			descStartTime, descEndTime = descEndTime, descStartTime
		}

		// 检查时间一致性
		var timeMismatch bool
		var reason string

		if strictMatch {
			// 严格匹配：时间必须完全一致
			if !descStartTime.Equal(configStartTime) || !descEndTime.Equal(configEndTime) {
				timeMismatch = true
				reason = fmt.Sprintf("描述中的时间(%s至%s)与配置时间(%s至%s)不一致",
					formatTimeForDisplay(descStartTime), formatTimeForDisplay(descEndTime),
					formatTimeForDisplay(configStartTime), formatTimeForDisplay(configEndTime))
			}
		} else {
			// 非严格匹配：只检查日期部分是否一致（忽略时间）
			descStartDate := time.Date(descStartTime.Year(), descStartTime.Month(), descStartTime.Day(), 0, 0, 0, 0, time.UTC)
			descEndDate := time.Date(descEndTime.Year(), descEndTime.Month(), descEndTime.Day(), 0, 0, 0, 0, time.UTC)
			configStartDate := time.Date(configStartTime.Year(), configStartTime.Month(), configStartTime.Day(), 0, 0, 0, 0, time.UTC)
			configEndDate := time.Date(configEndTime.Year(), configEndTime.Month(), configEndTime.Day(), 0, 0, 0, 0, time.UTC)

			if !descStartDate.Equal(configStartDate) || !descEndDate.Equal(configEndDate) {
				timeMismatch = true
				reason = fmt.Sprintf("描述中的日期(%s至%s)与配置日期(%s至%s)不一致",
					descStartDate.Format("2006-01-02"), descEndDate.Format("2006-01-02"),
					configStartDate.Format("2006-01-02"), configEndDate.Format("2006-01-02"))
			}
		}

		if timeMismatch {
			res = append(res, &json_rule.CellError{
				Index:    i,
				ExcelRow: startRowIdx + i + 1,
				Reason:   reason,
			})
		}
	}

	return res
}

// 改进的时间提取函数
func extractAllTimeStrings(desc string) []string {
	var timeStrings []string

	// 先尝试提取完整的中文日期时间格式
	chineseDateTimePattern := regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}日\s*\d{1,2}(?:点\d{1,2}(?:分\d{1,2}秒?)?)?`)
	chineseMatches := chineseDateTimePattern.FindAllString(desc, -1)
	timeStrings = append(timeStrings, chineseMatches...)

	// 提取带冒号的日期时间格式
	colonDateTimePattern := regexp.MustCompile(`\d{4}[-/]\d{1,2}[-/]\d{1,2}\s*\d{1,2}:\d{1,2}(?::\d{1,2})?`)
	colonMatches := colonDateTimePattern.FindAllString(desc, -1)
	timeStrings = append(timeStrings, colonMatches...)

	// 提取纯日期格式
	datePattern := regexp.MustCompile(`\d{4}[-/年]\d{1,2}[-/月]\d{1,2}[日]?`)
	dateMatches := datePattern.FindAllString(desc, -1)

	// 过滤掉已经包含在完整格式中的日期
	for _, date := range dateMatches {
		alreadyIncluded := false
		for _, ts := range timeStrings {
			if strings.Contains(ts, date) {
				alreadyIncluded = true
				break
			}
		}
		if !alreadyIncluded {
			timeStrings = append(timeStrings, date)
		}
	}

	// 提取单独的时间部分（当日期和时间分开写时）
	timePartPattern := regexp.MustCompile(`\d{1,2}(?:点\d{1,2}(?:分\d{1,2}秒?)?|:\d{1,2}(?::\d{1,2})?)`)
	timePartMatches := timePartPattern.FindAllString(desc, -1)

	// 尝试将单独的时间部分与前面的日期组合
	combinedTimes := combineDateTimeParts(desc, timeStrings, timePartMatches)
	if len(combinedTimes) > 0 {
		timeStrings = combinedTimes
	}

	// 去重并排序
	timeStrings = deduplicateAndSort(timeStrings, desc)

	return timeStrings
}

// 组合日期和时间部分
func combineDateTimeParts(desc string, dates, times []string) []string {
	if len(dates) == 0 || len(times) == 0 {
		return dates
	}

	var result []string

	// 查找日期和时间在文本中的位置
	positions := make(map[string]int)
	for _, d := range dates {
		positions[d] = strings.Index(desc, d)
	}
	for _, t := range times {
		positions[t] = strings.Index(desc, t)
	}

	// 尝试配对：时间紧跟在日期后面
	for i, date := range dates {
		datePos := positions[date]

		// 寻找紧接在日期后面的时间
		var closestTime string
		var closestDistance = 1000000 // 一个大数

		for _, timePart := range times {
			timePos := positions[timePart]
			if timePos > datePos {
				distance := timePos - (datePos + len(date))
				if distance < closestDistance && distance < 10 { // 距离小于10个字符
					closestDistance = distance
					closestTime = timePart
				}
			}
		}

		if closestTime != "" {
			// 组合日期和时间
			combined := strings.TrimSpace(date) + " " + strings.TrimSpace(closestTime)
			result = append(result, combined)

			// 从times中移除已使用的时间
			times = removeFromSlice(times, closestTime)
		} else {
			// 没有找到匹配的时间，只保留日期
			result = append(result, date)
		}

		// 如果已经找到了两个时间，就停止
		if len(result) >= 2 && i >= 1 {
			break
		}
	}

	// 添加剩余的时间（如果有的话且结果不足2个）
	if len(result) < 2 && len(times) > 0 {
		for _, t := range times {
			result = append(result, t)
			if len(result) >= 2 {
				break
			}
		}
	}

	return result
}

// 从切片中移除元素
func removeFromSlice(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// 去重并排序
func deduplicateAndSort(timeStrings []string, desc string) []string {
	// 去重
	unique := make(map[string]bool)
	var result []string
	for _, ts := range timeStrings {
		if !unique[ts] {
			unique[ts] = true
			result = append(result, ts)
		}
	}

	// 按在文本中的位置排序
	if len(result) > 1 {
		sort.Slice(result, func(i, j int) bool {
			posI := strings.Index(desc, result[i])
			posJ := strings.Index(desc, result[j])
			return posI < posJ
		})
	}

	return result
}

// 改进的灵活解析时间函数
func parseFlexibleTime(timeStr string, defaultYear int, autoCompleteYear bool) (time.Time, error) {
	// 标准化时间字符串
	timeStr = strings.TrimSpace(timeStr)

	// 定义解析函数
	parseFuncs := []func(string) (time.Time, error){
		parseChineseDateTime,
		parseColonDateTime,
		parseChineseDate,
		parseColonDate,
		parseMonthDay,
		parseTimeOnly,
	}

	for _, parseFunc := range parseFuncs {
		if t, err := parseFunc(timeStr); err == nil {
			// 如果时间没有年份且需要自动补充
			if autoCompleteYear && t.Year() == 0 {
				t = time.Date(defaultYear, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间: %s", timeStr)
}

// 解析中文日期时间：2025年1月1日15点30分55秒
func parseChineseDateTime(timeStr string) (time.Time, error) {
	// 完整格式：年-月-日 时:分:秒
	re := regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日\s*(\d{1,2})点(\d{1,2})分(\d{1,2})秒`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 7 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		minute, _ := strconv.Atoi(matches[5])
		second, _ := strconv.Atoi(matches[6])
		return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
	}

	// 缺少秒：年-月-日 时:分
	re = regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日\s*(\d{1,2})点(\d{1,2})分`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 6 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		minute, _ := strconv.Atoi(matches[5])
		return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC), nil
	}

	// 只有小时：年-月-日 时点
	re = regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日\s*(\d{1,2})点`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 5 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour, _ := strconv.Atoi(matches[4])
		return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("不是中文日期时间格式")
}

// 解析冒号日期时间：2025-01-01 15:30:55
func parseColonDateTime(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-1-2 15:4:5",
		"2006/1/2 15:4:5",
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006-1-2 15:4",
		"2006/1/2 15:4",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("不是冒号日期时间格式")
}

// 解析中文日期：2025年1月1日
func parseChineseDate(timeStr string) (time.Time, error) {
	re := regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 4 {
		year, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("不是中文日期格式")
}

// 解析冒号日期：2025-01-01
func parseColonDate(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-1-2",
		"2006/1/2",
		"20060102",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("不是冒号日期格式")
}

// 解析月日：1月1日
func parseMonthDay(timeStr string) (time.Time, error) {
	re := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 3 {
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])
		// 返回年份为0的时间，调用者会补充
		return time.Date(0, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("不是月日格式")
}

// 解析纯时间：15点30分55秒
func parseTimeOnly(timeStr string) (time.Time, error) {
	// 完整时间：时:分:秒
	re := regexp.MustCompile(`(\d{1,2})点(\d{1,2})分(\d{1,2})秒`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 4 {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		second, _ := strconv.Atoi(matches[3])
		// 使用今天日期
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.UTC), nil
	}

	// 时:分
	re = regexp.MustCompile(`(\d{1,2})点(\d{1,2})分`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC), nil
	}

	// 只有小时
	re = regexp.MustCompile(`(\d{1,2})点`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 2 {
		hour, _ := strconv.Atoi(matches[1])
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.UTC), nil
	}

	// 冒号格式：15:30:55
	re = regexp.MustCompile(`(\d{1,2}):(\d{1,2}):(\d{1,2})`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 4 {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		second, _ := strconv.Atoi(matches[3])
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, time.UTC), nil
	}

	// 冒号格式：15:30
	re = regexp.MustCompile(`(\d{1,2}):(\d{1,2})`)
	if matches := re.FindStringSubmatch(timeStr); len(matches) == 3 {
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC), nil
	}

	return time.Time{}, fmt.Errorf("不是时间格式")
}

// 按时间字符串在文本中的位置排序
func sortTimeStringsByPosition(desc string, timeStrings *[]string) {
	positionMap := make(map[string]int)
	for _, ts := range *timeStrings {
		positionMap[ts] = strings.Index(desc, ts)
	}

	// 按位置排序
	sorted := make([]string, len(*timeStrings))
	copy(sorted, *timeStrings)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if positionMap[sorted[i]] > positionMap[sorted[j]] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	*timeStrings = sorted
}

// 辅助函数：格式化时间用于显示
func formatTimeForDisplay(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("2006-01-02")
	}
	return t.Format("2006-01-02 15:04:05")
}
