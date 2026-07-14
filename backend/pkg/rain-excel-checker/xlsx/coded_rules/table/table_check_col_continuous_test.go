package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

// buildContinuousCheckParam 构建列连续性检查的参数
func buildContinuousCheckParam(cols [][]string, params map[string]string) json_rule.CheckParam {
	if params == nil {
		params = make(map[string]string)
	}
	if _, ok := params["targetCol"]; !ok {
		params["targetCol"] = "SeasonIndex"
	}
	if _, ok := params["checkMode"]; !ok {
		params["checkMode"] = "INCREASE_STRICT"
	}

	// 计算数据结束索引：找到最大列长度
	endIndex := 0
	for _, col := range cols {
		if len(col) > endIndex {
			endIndex = len(col)
		}
	}

	return json_rule.CheckParam{
		SheetName:   "测试表|TestTable",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    endIndex,
		Params:      params,
	}
}

// buildTestCols 构建测试用列数据
// header 行: 索引0(中文名), 1(类型), 2(字段名), 3(导出标记)
// 数据行: 索引4开始
func buildTestCols(colName string, values ...string) [][]string {
	col := make([]string, 4+len(values))
	col[0] = "测试列"
	col[1] = "int"
	col[2] = colName
	col[3] = "server"
	for i, v := range values {
		col[4+i] = v
	}
	return [][]string{col}
}

// ========== INCREASE_STRICT 测试 ==========

func TestColContinuousCheck_IncreaseStrict_AllPass(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "1", "2", "3", "4")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonIndex",
		"checkMode": "INCREASE_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "连续数据应通过检查")
	assert.Empty(t, result.ErrCells, "不应有错误")
}

func TestColContinuousCheck_IncreaseStrict_Gap(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "1", "2", "5", "6")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonIndex",
		"checkMode": "INCREASE_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "不连续数据应失败")
	assert.Len(t, result.ErrCells, 1, "应有1个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "期望为 3", "错误信息应指出期望值")
}

func TestColContinuousCheck_IncreaseStrict_WithStartValue(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "6", "7", "8")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":  "SeasonIndex",
		"checkMode":  "INCREASE_STRICT",
		"startValue": "5",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "从startValue开始连续应通过")
}

func TestColContinuousCheck_IncreaseStrict_StartValueMismatch(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "7", "8", "9")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":  "SeasonIndex",
		"checkMode":  "INCREASE_STRICT",
		"startValue": "5",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "第一个值不等于 startValue+1 应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "期望为 6")
}

func TestColContinuousCheck_MissingTargetCol(t *testing.T) {
	cols := buildTestCols("OtherCol", "1", "2", "3")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonIndex",
		"checkMode": "INCREASE_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "目标列不存在应失败")
	assert.Contains(t, result.Reason, "未找到列")
}

func TestColContinuousCheck_EmptyData(t *testing.T) {
	cols := buildTestCols("SeasonIndex")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonIndex",
		"checkMode": "INCREASE_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "空数据应通过")
}

func TestColContinuousCheck_MissingTargetColParam(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "1", "2")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "",
		"checkMode": "INCREASE_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "缺少 targetCol 参数应失败")
	assert.Contains(t, result.Reason, "缺少必填参数")
}

// ========== INCREASE_MONOTONE 测试 ==========

func TestColContinuousCheck_IncreaseMonotone_AllPass(t *testing.T) {
	cols := buildTestCols("SortOrder", "1", "3", "5", "10")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SortOrder",
		"checkMode": "INCREASE_MONOTONE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "单调递增数据应通过")
}

func TestColContinuousCheck_IncreaseMonotone_Decrease(t *testing.T) {
	cols := buildTestCols("SortOrder", "1", "5", "3")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SortOrder",
		"checkMode": "INCREASE_MONOTONE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "递减数据应失败")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "必须大于上一个值")
}

func TestColContinuousCheck_IncreaseMonotone_Duplicate(t *testing.T) {
	cols := buildTestCols("SortOrder", "1", "3", "3")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SortOrder",
		"checkMode": "INCREASE_MONOTONE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "重复值应失败（非严格递增）")
}

// ========== DATE_CONTINUOUS 测试 ==========

func TestColContinuousCheck_DateContinuous_AllPass(t *testing.T) {
	cols := buildTestCols("StartTime",
		"2026-01-01 00:00:00",
		"2026-01-08 00:00:00",
		"2026-01-15 00:00:00",
		"2026-01-22 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "间隔一致应通过")
}

func TestColContinuousCheck_DateContinuous_InconsistentGap(t *testing.T) {
	cols := buildTestCols("StartTime",
		"2026-01-01 00:00:00",
		"2026-01-08 00:00:00",
		"2026-01-18 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "间隔不一致应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "日期间隔")
}

func TestColContinuousCheck_DateContinuous_WithinTolerance(t *testing.T) {
	cols := buildTestCols("StartTime",
		"2026-01-01 00:00:00",
		"2026-01-08 00:00:00",
		"2026-01-16 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_CONTINUOUS",
		"tolerance": "2",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "容差内应通过")
}

func TestColContinuousCheck_DateContinuous_InsufficientData(t *testing.T) {
	cols := buildTestCols("StartTime", "2026-01-01 00:00:00")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "数据不足2行应跳过检查")
}

// ========== ID_FORMAT_CONTINUOUS 测试 ==========

func TestColContinuousCheck_IdFormatContinuous_AllPass(t *testing.T) {
	cols := buildTestCols("ItemCfg",
		"{1001;1}",
		"{1002;1}",
		"{1003;1}",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "ItemCfg",
		"checkMode": "ID_FORMAT_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "ID连续应通过")
}

func TestColContinuousCheck_IdFormatContinuous_Gap(t *testing.T) {
	cols := buildTestCols("ItemCfg",
		"{1001;1}",
		"{1005;1}",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "ItemCfg",
		"checkMode": "ID_FORMAT_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "ID跳跃应失败")
}

func TestColContinuousCheck_IdFormatContinuous_InvalidFormat(t *testing.T) {
	cols := buildTestCols("ItemCfg", "not_a_format")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "ItemCfg",
		"checkMode": "ID_FORMAT_CONTINUOUS",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "格式无效应失败")
}

// ========== 排除行和通用参数测试 ==========

func TestColContinuousCheck_ExcludeRows_Pass(t *testing.T) {
	// 数据: [1, 2, 99, 4] 排除第3行（数据行号3）→ 检查序列 [1, 2, 4]
	// INCREASE_STRICT: prevVal=1, 2(ok), 4(expect 3, fail)
	cols := buildTestCols("SeasonIndex", "1", "2", "99", "4")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":   "SeasonIndex",
		"checkMode":   "INCREASE_STRICT",
		"excludeRows": "3",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "排除后序列 [1,2,4] 不连续应失败")
}

func TestColContinuousCheck_ExcludeRows_Range(t *testing.T) {
	// 数据: [1, 2, 99, 99, 99, 6] 排除第3-5行 → 检查序列 [1, 2, 6]
	cols := buildTestCols("SeasonIndex", "1", "2", "99", "99", "99", "6")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":   "SeasonIndex",
		"checkMode":   "INCREASE_STRICT",
		"excludeRows": "3-5",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "排除后序列 [1,2,6] 不连续应失败")
}

func TestColContinuousCheck_AllowEmpty(t *testing.T) {
	// 数据: [1, 2, "", 4] allowEmpty=true → 检查序列 [1, 2, 4]
	cols := buildTestCols("SeasonIndex", "1", "2", "", "4")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":  "SeasonIndex",
		"checkMode":  "INCREASE_STRICT",
		"allowEmpty": "true",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "空值跳过后 [1,2,4] 不连续应失败")
}

func TestColContinuousCheck_AllowCommit(t *testing.T) {
	// 数据: [1, 2, "#注释", 3] allowCommit=true → 检查序列 [1, 2, 3]
	cols := buildTestCols("SeasonIndex", "1", "2", "#注释", "3")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":   "SeasonIndex",
		"checkMode":   "INCREASE_STRICT",
		"allowCommit": "true",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "注释行跳过后 [1,2,3] 连续应通过")
}

func TestColContinuousCheck_UnknownCheckMode(t *testing.T) {
	cols := buildTestCols("SeasonIndex", "1", "2", "3")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonIndex",
		"checkMode": "UNKNOWN_MODE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "未知模式应失败")
	assert.Contains(t, result.Reason, "未知的检查模式")
}

// ========== EXTRACT_NUMBER_STRICT 测试 ==========

func TestColContinuousCheck_ExtractNumber_AllPass(t *testing.T) {
	// 模拟 ArenaSeason 的 SeasonName 列: "汉武盛世", "第2赛季", "第3赛季", "第4赛季"
	// "汉武盛世" 没有数字会被跳过（需设置 startValue 或排除第一行）
	cols := buildTestCols("SeasonName", "1赛季", "第2赛季", "第3赛季", "第4赛季")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonName",
		"checkMode": "EXTRACT_NUMBER_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "提取数字后连续 [1,2,3,4] 应通过")
	assert.Empty(t, result.ErrCells)
}

func TestColContinuousCheck_ExtractNumber_WithGap(t *testing.T) {
	cols := buildTestCols("SeasonName", "第2赛季", "第3赛季", "第5赛季")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonName",
		"checkMode": "EXTRACT_NUMBER_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "提取数字后 [2,3,5] 不连续应失败")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "期望为 4")
}

func TestColContinuousCheck_ExtractNumber_WithStartValue(t *testing.T) {
	// 第一行没有数字的文本，配合 startValue 使用
	cols := buildTestCols("SeasonName", "汉武盛世", "第2赛季", "第3赛季")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":  "SeasonName",
		"checkMode":  "EXTRACT_NUMBER_STRICT",
		"startValue": "1",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "第一行无数字应报错")
	assert.Contains(t, result.ErrCells[0].Reason, "未找到数字")
}

func TestColContinuousCheck_ExtractNumber_ExcludeFirstRow(t *testing.T) {
	// 第一行没有数字，用排除行号跳过
	cols := buildTestCols("SeasonName", "汉武盛世", "第2赛季", "第3赛季")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol":   "SeasonName",
		"checkMode":   "EXTRACT_NUMBER_STRICT",
		"excludeRows": "1",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "排除第一行后 [2,3] 连续应通过")
}

func TestColContinuousCheck_ExtractNumber_NoNumber(t *testing.T) {
	cols := buildTestCols("SeasonName", "无数字文本", "也没有数字")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonName",
		"checkMode": "EXTRACT_NUMBER_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "所有行都无数字应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "未找到数字")
}

func TestColContinuousCheck_ExtractNumber_SFormat(t *testing.T) {
	// S3, S4, S5 格式
	cols := buildTestCols("SeasonLabel", "S3", "S4", "S5")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonLabel",
		"checkMode": "EXTRACT_NUMBER_STRICT",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "S3,S4,S5 提取数字 [3,4,5] 连续应通过")
}

// ========== SPLIT_UNIQUE 测试 ==========

func TestColContinuousCheck_SplitUnique_AllPass(t *testing.T) {
	// 每行的 byproduct 值不重复，跨行也不重复
	cols := buildTestCols("byproduct", "1001,1002", "1003,1004", "1005")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "byproduct",
		"checkMode": "SPLIT_UNIQUE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "所有值唯一应通过")
	assert.Empty(t, result.ErrCells)
}

func TestColContinuousCheck_SplitUnique_DuplicateAcrossRows(t *testing.T) {
	// 1003 出现在第1行和第2行
	cols := buildTestCols("byproduct", "1001,1003", "1003,1004")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "byproduct",
		"checkMode": "SPLIT_UNIQUE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "跨行重复应失败")
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "1003")
	assert.Contains(t, result.ErrCells[0].Reason, "重复")
}

func TestColContinuousCheck_SplitUnique_DuplicateInSameRow(t *testing.T) {
	// 同一行内重复: "1001,1001"
	cols := buildTestCols("byproduct", "1001,1001", "1002")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "byproduct",
		"checkMode": "SPLIT_UNIQUE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "同行内重复应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "1001")
}

func TestColContinuousCheck_SplitUnique_EmptyValues(t *testing.T) {
	// 空行跳过，只检查有值的行
	cols := buildTestCols("byproduct", "1001,1002", "", "1003,1004")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "byproduct",
		"checkMode": "SPLIT_UNIQUE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "空行跳过后所有值唯一应通过")
}

func TestColContinuousCheck_SplitUnique_CustomSeparator(t *testing.T) {
	// 使用自定义分隔符 "|"
	cols := buildTestCols("Items", "a|b", "c|d", "a|e")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "Items",
		"checkMode": "SPLIT_UNIQUE",
		"separator": "|",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "a 跨行重复应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "a")
	assert.Contains(t, result.ErrCells[0].Reason, "重复")
}

func TestColContinuousCheck_SplitUnique_SingleValuePerRow(t *testing.T) {
	// 每行只有一个值，相当于全局唯一检查
	cols := buildTestCols("Id", "1001", "1002", "1003")
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "Id",
		"checkMode": "SPLIT_UNIQUE",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "每行单值全部唯一应通过")
}

// ========== DATE_MONTHLY_PATTERN 测试 ==========

func TestColContinuousCheck_DateMonthly_AllPass(t *testing.T) {
	// 每月15号 00:00:00，连续3个月
	cols := buildTestCols("StartTime",
		"2026-01-15 00:00:00",
		"2026-02-15 00:00:00",
		"2026-03-15 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "月顺延、日/时一致应通过")
	assert.Empty(t, result.ErrCells)
}

func TestColContinuousCheck_DateMonthly_CrossYear(t *testing.T) {
	// 跨年月顺延：11月 → 12月 → 1月
	cols := buildTestCols("StartTime",
		"2025-11-15 00:00:00",
		"2025-12-15 00:00:00",
		"2026-01-15 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.True(t, result.Ok, "跨年月顺延应通过")
}

func TestColContinuousCheck_DateMonthly_DayMismatch(t *testing.T) {
	// 日不一致：15号 → 16号
	cols := buildTestCols("StartTime",
		"2026-01-15 00:00:00",
		"2026-02-16 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "日不一致应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "15 号")
}

func TestColContinuousCheck_DateMonthly_TimeMismatch(t *testing.T) {
	// 时间不一致
	cols := buildTestCols("StartTime",
		"2026-01-15 00:00:00",
		"2026-02-15 12:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "时间不一致应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "00:00:00")
}

func TestColContinuousCheck_DateMonthly_MonthGap(t *testing.T) {
	// 月份跳跃：1月 → 3月（跳过了2月）
	cols := buildTestCols("StartTime",
		"2026-01-15 00:00:00",
		"2026-03-15 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "StartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	assert.False(t, result.Ok, "月份跳跃应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "2月")
}

func TestColContinuousCheck_DateMonthly_RealArenaSeason(t *testing.T) {
	// 模拟 ArenaSeason 实际数据（从预览中获取的5行）
	cols := buildTestCols("SeasonStartTime",
		"2025-12-15 00:00:00",
		"2026-02-15 00:00:00",
		"2026-03-15 00:00:00",
		"2026-04-15 00:00:00",
		"2026-05-15 00:00:00",
	)
	param := buildContinuousCheckParam(cols, map[string]string{
		"targetCol": "SeasonStartTime",
		"checkMode": "DATE_MONTHLY_PATTERN",
	})

	rule := &ColContinuousCheckRule{}
	result := rule.Check(param)

	// 第1行到第2行跳过了1月（12月→2月），应该报错
	assert.False(t, result.Ok, "ArenaSeason 实际数据：12月→2月跳跃应失败")
	assert.Contains(t, result.ErrCells[0].Reason, "2026年1月")
}
