package notification

import (
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

func TestErrorFormatter_FormatColErrors(t *testing.T) {
	// 测试数据
	sheetName := "Hero"
	colName := "Id"
	tableName := "Hero.xlsx"

	// 创建一个失败的检查结果
	result := &json_rule.ColCheckResult{
		TableName: &tableName,
		SheetName: &sheetName,
		ColName:   &colName,
		Ok:        false,
		Reason:    "ID不能为空",
		ErrCells: []*json_rule.CellError{
			{Index: 4, Reason: "值为空"}, // 第5行
			{Index: 9, Reason: "值为空"}, // 第10行
		},
	}

	// 测试带颜色的格式化
	formatter := NewErrorFormatter(true)
	lines := formatter.FormatColErrors([]*json_rule.ColCheckResult{result})

	if len(lines) != 2 {
		t.Errorf("期望 2 行错误, 实际 %d 行", len(lines))
	}

	// 检查颜色标签
	if !strings.Contains(lines[0], "<font color='red'>") {
		t.Error("期望包含红色字体标签")
	}

	// 检查行号转换 (Index+1)
	if !strings.Contains(lines[0], "第5行") {
		t.Error("期望包含 '第5行'")
	}
	if !strings.Contains(lines[1], "第10行") {
		t.Error("期望包含 '第10行'")
	}

	// 测试不带颜色的格式化
	formatterNoColor := NewErrorFormatter(false)
	linesNoColor := formatterNoColor.FormatColErrors([]*json_rule.ColCheckResult{result})

	if strings.Contains(linesNoColor[0], "<font") {
		t.Error("不期望包含字体标签")
	}
}

func TestErrorFormatter_FormatTableErrors(t *testing.T) {
	sheetName := "SeasonPass"
	tableName := "SeasonPass.xlsx"

	result := &json_rule.TableCheckResult{
		TableName:   &tableName,
		SheetName:   &sheetName,
		DisplayName: "赛季检查",
		Ok:          false,
		Reason:      "赛季时间不足",
		ErrCells:    nil,
	}

	formatter := NewErrorFormatter(true)
	lines := formatter.FormatTableErrors([]*json_rule.TableCheckResult{result})

	if len(lines) != 1 {
		t.Errorf("期望 1 行错误, 实际 %d 行", len(lines))
	}

	if !strings.Contains(lines[0], "赛季检查") {
		t.Error("期望包含显示名称")
	}
}

func TestGetSummary(t *testing.T) {
	sheetName := "Test"
	colName := "Col"

	// 创建 2 个失败 + 1 个通过
	results := []*json_rule.ColCheckResult{
		{SheetName: &sheetName, ColName: &colName, Ok: false},
		{SheetName: &sheetName, ColName: &colName, Ok: false},
		{SheetName: &sheetName, ColName: &colName, Ok: true},
	}

	parseErrors := []*SheetParseError{
		{FileName: "test.xlsx", SheetName: "Sheet1", Error: "解析错误"},
	}

	event := &CheckResultEvent{
		ColResults:  results,
		ParseErrors: parseErrors,
		CheckTime:   time.Now(),
	}

	summary := GetSummary(event)

	if summary.ColErrors != 2 {
		t.Errorf("期望 ColErrors=2, 实际=%d", summary.ColErrors)
	}
	if summary.ParseErrors != 1 {
		t.Errorf("期望 ParseErrors=1, 实际=%d", summary.ParseErrors)
	}
	if summary.TotalErrors != 3 {
		t.Errorf("期望 TotalErrors=3, 实际=%d", summary.TotalErrors)
	}
	if !summary.HasErrors {
		t.Error("期望 HasErrors=true")
	}
}

func TestErrorFormatter_FormatNotifications(t *testing.T) {
	sheetName := "Hero"
	tableName := "Hero.xlsx"

	// 创建通知规则结果（Ok=true, 有 ErrCells）
	notifyResult := &json_rule.TableCheckResult{
		TableName:   &tableName,
		SheetName:   &sheetName,
		DisplayName: "新增行/列通知",
		Ok:          true,
		Reason:      "🆕 新增行 (2条)\n• ID=32, 名称=测试\n• ID=33, 名称=测试2",
		ErrCells: []*json_rule.CellError{
			{Index: 31, Reason: "新增行: ID=32, 名称=测试"},
			{Index: 32, Reason: "新增行: ID=33, 名称=测试2"},
		},
	}

	// 创建错误规则结果（Ok=false）- 不应被 FormatNotifications 处理
	errorResult := &json_rule.TableCheckResult{
		TableName:   &tableName,
		SheetName:   &sheetName,
		DisplayName: "赛季检查",
		Ok:          false,
		Reason:      "赛季已结束",
		ErrCells: []*json_rule.CellError{
			{Index: 7, Reason: "赛季结束"},
		},
	}

	// 创建无 ErrCells 的通知结果（如"首次运行"）- 不应被 FormatNotifications 处理
	emptyNotifyResult := &json_rule.TableCheckResult{
		TableName:   &tableName,
		SheetName:   &sheetName,
		DisplayName: "新增行/列通知",
		Ok:          true,
		Reason:      "首次运行，无历史数据",
		ErrCells:    []*json_rule.CellError{},
	}

	formatter := NewErrorFormatter(true)
	results := []*json_rule.TableCheckResult{notifyResult, errorResult, emptyNotifyResult}
	lines := formatter.FormatNotifications(results)

	// 应该只返回 Ok=true 且有 ErrCells 的结果
	if len(lines) != 1 {
		t.Errorf("期望 1 行通知, 实际 %d 行", len(lines))
	}

	// 新格式直接输出 Reason，不再添加前缀
	if !strings.Contains(lines[0], "🆕") {
		t.Error("期望包含 Reason 原始内容")
	}

	// 确认不包含旧格式的 📝 前缀
	if strings.Contains(lines[0], "📝 新增行/列通知:") {
		t.Error("不期望包含旧格式的 📝 前缀")
	}
}

func TestErrorFormatter_FormatNotifications_Empty(t *testing.T) {
	formatter := NewErrorFormatter(true)

	// 空结果集
	lines := formatter.FormatNotifications([]*json_rule.TableCheckResult{})
	if len(lines) != 0 {
		t.Errorf("期望 0 行通知, 实际 %d 行", len(lines))
	}
}

func TestCountNotificationResults(t *testing.T) {
	sheetName := "Test"

	results := []*json_rule.TableCheckResult{
		{SheetName: &sheetName, Ok: true, ErrCells: []*json_rule.CellError{{Index: 1}}},  // 有通知
		{SheetName: &sheetName, Ok: true, ErrCells: []*json_rule.CellError{}},            // 无 ErrCells
		{SheetName: &sheetName, Ok: false, ErrCells: []*json_rule.CellError{{Index: 1}}}, // 错误规则
		nil, // nil 结果
	}

	count := countNotificationResults(results)
	if count != 1 {
		t.Errorf("期望通知数=1, 实际=%d", count)
	}
}

func TestGetSummary_WithNotifications(t *testing.T) {
	sheetName := "Test"

	results := []*json_rule.TableCheckResult{
		{SheetName: &sheetName, Ok: true, ErrCells: []*json_rule.CellError{{Index: 1}}},
		{SheetName: &sheetName, Ok: true, ErrCells: []*json_rule.CellError{{Index: 2}}},
	}

	event := &CheckResultEvent{
		TableResults: results,
		CheckTime:    time.Now(),
	}

	summary := GetSummary(event)

	// 检查通知统计
	if summary.TableNotifications != 2 {
		t.Errorf("期望 TableNotifications=2, 实际=%d", summary.TableNotifications)
	}

	// 检查通知标志
	if !summary.HasNotifications {
		t.Error("期望 HasNotifications=true")
	}

	// 通知不应计入错误数
	if summary.TableErrors != 0 {
		t.Errorf("期望 TableErrors=0, 实际=%d", summary.TableErrors)
	}

	// 通知不应触发 HasErrors
	if summary.HasErrors {
		t.Error("期望 HasErrors=false（通知不是错误）")
	}
}

func TestFormatConsoleOutput(t *testing.T) {
	sheetName := "Test"
	colName := "Col"

	results := []*json_rule.ColCheckResult{
		{SheetName: &sheetName, ColName: &colName, Ok: false, Reason: "测试失败"},
	}

	event := &CheckResultEvent{
		ColResults: results,
		CheckTime:  time.Now(),
	}

	formatter := NewErrorFormatter(false)
	output := formatter.FormatConsoleOutput(event)

	// 检查输出包含关键信息
	if !strings.Contains(output, "Excel 配表检查结果") {
		t.Error("期望包含标题")
	}
	if !strings.Contains(output, "检查统计") {
		t.Error("期望包含统计信息")
	}
	if !strings.Contains(output, "⚠️") && !strings.Contains(output, "✅") {
		t.Error("期望包含结果图标")
	}
}
