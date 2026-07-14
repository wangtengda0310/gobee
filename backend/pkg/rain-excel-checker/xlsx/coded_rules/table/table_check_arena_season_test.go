// Package table 提供表级别的校验规则
// 本包中的规则针对单个 Excel 表的特定业务逻辑进行校验

package coded_rules

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/xuri/excelize/v2"
)

// ==================== 竞技场赛季检查测试 ====================

func TestArenaSeasonCheckRule_ShouldPass(t *testing.T) {
	// 赛季结束时间在 7 天后，应该通过
	cols, params, sheetMap := fakeArenaSeasonData_Future()

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (Future) Result:", string(jsonData))

	if !result.Ok {
		t.Errorf("预期通过，但失败: %s", result.Reason)
	}
}

func TestArenaSeasonCheckRule_ShouldWarn(t *testing.T) {
	// 赛季结束时间在 3 天后，应该警告
	cols, params, sheetMap := fakeArenaSeasonData_Soon()

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (Soon) Result:", string(jsonData))

	if result.Ok {
		t.Error("预期警告，但通过了")
	}
	if result.Reason == "" {
		t.Error("警告应该有原因说明")
	}
}

func TestArenaSeasonCheckRule_ShouldFail(t *testing.T) {
	// 赛季已结束，应该报错
	cols, params, sheetMap := fakeArenaSeasonData_Past()

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (Past) Result:", string(jsonData))

	if result.Ok {
		t.Error("预期失败，但通过了")
	}
	if len(result.ErrCells) == 0 {
		t.Error("应该有错误单元格")
	}
}

func TestArenaSeasonCheckRule_MissingColumn(t *testing.T) {
	// 缺少必要列
	cols, params, sheetMap := fakeArenaSeasonData_MissingColumn()

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (MissingColumn) Result:", string(jsonData))

	if result.Ok {
		t.Error("预期失败，但通过了")
	}
	if result.Reason == "" {
		t.Error("缺少列时应该有原因说明")
	}
}

func TestArenaSeasonCheckRule_EmptyData(t *testing.T) {
	// 空数据
	cols, params, sheetMap := fakeArenaSeasonData_Empty()

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    excelio.MJS_FIXED_ROWS_NUM, // 空数据：EndIndex == StartRowIdx
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (Empty) Result:", string(jsonData))

	if result.Ok {
		t.Error("预期失败，但通过了")
	}
}

// ==================== 测试数据构造函数 ====================
// 注意：
// - MJS_FIXED_ROWS_NAME = 2，列名在第 3 行（索引 2）
// - MJS_FIXED_ROWS_NUM = 4，数据从第 5 行（索引 4）开始
// - cols 是按列存储的二维数组，每个元素是一列的所有行数据

func fakeArenaSeasonData_Future() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h" // 7天

	// 赛季结束时间在 10 天后
	futureTime := time.Now().AddDate(0, 0, 10).Format("2006-01-02 15:04:05")

	// 模拟 ArenaSeason 表数据（按列存储）
	// 索引 0-3 是表头，索引 4+ 是数据
	// 索引 2 是列名（MJS_FIXED_ROWS_NAME）
	cols = [][]string{
		// Id 列: [第1行, 第2行, "Id", 第4行, 数据1, 数据2]
		{"", "", "Id", "", "1", "2"},
		// SeasonStartTime 列
		{"", "", "SeasonStartTime", "", "2024-01-01 00:00:00", "2024-02-01 00:00:00"},
		// SeasonEndTime 列
		{"", "", "SeasonEndTime", "", "2024-02-01 00:00:00", futureTime},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

func fakeArenaSeasonData_Soon() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h" // 7天

	// 赛季结束时间在 3 天后（小于 7 天警告阈值）
	soonTime := time.Now().AddDate(0, 0, 3).Format("2006-01-02 15:04:05")

	cols = [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "SeasonStartTime", "", "2024-01-01 00:00:00"},
		{"", "", "SeasonEndTime", "", soonTime},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

func fakeArenaSeasonData_Past() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	// 赛季已结束（1 天前）
	pastTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")

	cols = [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "SeasonStartTime", "", "2024-01-01 00:00:00"},
		{"", "", "SeasonEndTime", "", pastTime},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

func fakeArenaSeasonData_MissingColumn() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	// 缺少 SeasonEndTime 列
	cols = [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "SeasonStartTime", "", "2024-01-01 00:00:00"},
		// 没有 SeasonEndTime 列
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

func fakeArenaSeasonData_Empty() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	// 空数据（只有表头，没有数据行）
	cols = [][]string{
		{"", "", "Id", ""}, // 只有 4 行，没有数据
		{"", "", "SeasonStartTime", ""},
		{"", "", "SeasonEndTime", ""},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

func TestArenaSeasonCheckRule_InvalidDurationParam(t *testing.T) {
	// 测试容错处理：使用错误的 duration 格式，应该回退到默认值 7天
	cols, _, sheetMap := fakeArenaSeasonData_Future()

	// 使用错误的 duration 格式（中文单位）
	params := make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "7天" // ❌ 错误格式，应该被容错

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (InvalidDuration) Result:", string(jsonData))

	// 容错处理：应该使用默认值 7天，而不是报错
	// 由于测试数据中赛季结束时间在 10 天后，大于 7 天阈值，所以应该通过
	if !result.Ok {
		t.Errorf("容错处理失败，应该使用默认值 7天 继续检查，但报错: %s", result.Reason)
	}
}

func TestArenaSeasonCheckRule_BugReproduction_InvalidParamWithSoonSeason(t *testing.T) {
	// 🔴 Bug 复现测试：修复前此测试会失败
	//
	// 场景：用户输入了错误的参数格式 "7天"（而不是 "168h"），
	//      同时赛季结束时间在 3 天后（应该触发警告）
	//
	// 修复前行为：
	//   - time.ParseDuration("7天") 返回 error
	//   - 直接返回 result.Ok=false, Reason="无法解析时间参数: 7天"
	//   - 即使赛季确实即将结束，也无法给出正确的警告信息
	//
	// 修复后行为：
	//   - 解析失败时回退到默认值 168h（7天）
	//   - 继续检查赛季时间，正确报告"赛季即将结束"
	//   - 用户能获得有意义的警告信息

	params := make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "7天" // ❌ Bug 触发点：中文单位

	// 赛季结束时间在 3 天后（小于默认的 7 天阈值，应该警告）
	soonTime := time.Now().AddDate(0, 0, 3).Format("2006-01-02 15:04:05")

	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "SeasonStartTime", "", "2024-01-01 00:00:00"},
		{"", "", "SeasonEndTime", "", soonTime},
	}

	sheetMap := map[string]*excelize.File{}

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("Bug Reproduction Test Result:", string(jsonData))

	// ✅ 修复后的预期行为
	// 1. 检查应该失败（赛季即将结束）
	// 2. Reason 应该包含有意义的警告信息（而非"无法解析时间参数"）
	// 3. ErrCells 应该包含具体的错误单元格信息

	if result.Ok {
		t.Error("修复后测试失败：赛季即将结束，但检查通过了")
	}

	// 验证错误信息不是参数解析错误
	if strings.Contains(result.Reason, "无法解析时间参数") {
		t.Error("修复后测试失败：仍然报告参数解析错误，应该回退到默认值并继续检查")
	}

	// 验证有具体的错误单元格信息
	if len(result.ErrCells) == 0 {
		t.Error("修复后测试失败：应该有错误单元格信息")
	}

	// 验证错误信息提到了"即将结束"（而不是参数错误）
	if !strings.Contains(result.Reason, "即将结束") {
		t.Errorf("修复后测试失败：Reason 应该包含'即将结束'，实际: %s", result.Reason)
	}

	t.Logf("✅ Bug 修复验证成功：即使参数格式错误，仍能正确报告赛季即将结束")
}

func TestArenaSeasonCheckRule_EmptyDurationParam(t *testing.T) {
	// 测试容错处理：空参数，应该回退到默认值 7天
	cols, _, sheetMap := fakeArenaSeasonData_Future()

	// 使用空的 duration 参数
	params := make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "" // 空值

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (EmptyDuration) Result:", string(jsonData))

	// 容错处理：应该使用默认值 7天
	if !result.Ok {
		t.Errorf("空参数容错处理失败，应该使用默认值 7天，但报错: %s", result.Reason)
	}
}

func TestArenaSeasonCheckRule_NumericOnlyParam(t *testing.T) {
	// 测试容错处理：纯数字（缺少单位），应该回退到默认值 7天
	cols, _, sheetMap := fakeArenaSeasonData_Future()

	// 使用纯数字（Go duration 不支持）
	params := make(map[string]string)
	params[string(json_rule.SEASON_END_TIME_COL)] = "SeasonEndTime"
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168" // ❌ 缺少单位 h

	rule := new(ArenaSeasonCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "ArenaSeason",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("ArenaSeasonCheckRule (NumericOnly) Result:", string(jsonData))

	// 容错处理：应该使用默认值 7天
	if !result.Ok {
		t.Errorf("纯数字参数容错处理失败，应该使用默认值 7天，但报错: %s", result.Reason)
	}
}
