// Package coded_rules 提供表级别的校验规则
// 本文件为丹青阁活动时间校验规则的单元测试

package coded_rules

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 丹青阁活动时间校验测试 ====================

// TestDanQingGeTimeActive_ShouldPass 活动距结束超过7天，应该通过
func TestDanQingGeTimeActive_ShouldPass(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_Future()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (Future) Result:", string(jsonData))

	assert.True(t, result.Ok, "活动距结束超过7天，应该通过")
	assert.Equal(t, 0, len(result.ErrCells), "不应该有错误单元格")
}

// TestDanQingGeTimeActive_ShouldWarn 活动离结束不到7天，应该警告
func TestDanQingGeTimeActive_ShouldWarn(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_Soon()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (Soon) Result:", string(jsonData))

	assert.False(t, result.Ok, "活动即将结束，应该不通过")
	assert.True(t, len(result.ErrCells) > 0, "应该有警告单元格")
	assert.True(t, strings.Contains(result.Reason, "即将结束"), "原因应包含'即将结束'")
}

// TestDanQingGeTimeActive_ShouldFail_Expired 活动已结束，应该报错
func TestDanQingGeTimeActive_ShouldFail_Expired(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_Expired()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (Expired) Result:", string(jsonData))

	assert.False(t, result.Ok, "活动已结束，应该不通过")
	assert.True(t, len(result.ErrCells) > 0, "应该有错误单元格")

	found := false
	for _, cell := range result.ErrCells {
		if strings.Contains(cell.Reason, "已结束") {
			found = true
			break
		}
	}
	assert.True(t, found, "应该有'已结束'的错误信息")
}

// TestDanQingGeTimeActive_MissingColumns 缺少必要列
func TestDanQingGeTimeActive_MissingColumns(t *testing.T) {
	// 缺少 ActivityType 列
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "测试"},
		{"", "", "EndTime", "", "2026-12-31 23:59:59"},
	}
	params := map[string]string{string(json_rule.TIME_RANGE_BEFORE): "168h"}

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    map[string]*excelize.File{},
	})

	assert.False(t, result.Ok, "缺少 ActivityType 列，应该不通过")
	assert.Contains(t, result.Reason, "ActivityType")

	// 缺少 EndTime 列
	cols2 := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "ActivityType", "", "ActTypeSkinRaffle"},
	}
	result2 := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols2,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    map[string]*excelize.File{},
	})
	assert.False(t, result2.Ok, "缺少 EndTime 列，应该不通过")
	assert.Contains(t, result2.Reason, "EndTime")
}

// TestDanQingGeTimeActive_EmptyData 空数据行
func TestDanQingGeTimeActive_EmptyData(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_Empty()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    excelio.MJS_FIXED_ROWS_NUM, // 空数据：EndIndex == StartRowIdx
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (Empty) Result:", string(jsonData))

	assert.False(t, result.Ok, "空数据应该不通过")
	assert.Contains(t, result.Reason, "没有有效的数据行")
}

// TestDanQingGeTimeActive_InvalidTime 无法解析的时间格式
func TestDanQingGeTimeActive_InvalidTime(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_InvalidTime()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (InvalidTime) Result:", string(jsonData))

	assert.False(t, result.Ok, "无效时间格式应该不通过")
	found := false
	for _, cell := range result.ErrCells {
		if strings.Contains(cell.Reason, "无法解析结束时间") {
			found = true
			break
		}
	}
	assert.True(t, found, "应该有'无法解析结束时间'的错误信息")
}

// TestDanQingGeTimeActive_MultipleRows 多行数据混合场景（含非丹青阁活动被跳过）
func TestDanQingGeTimeActive_MultipleRows(t *testing.T) {
	cols, params, sheetMap := fakeDanQingGeData_Multiple()

	rule := new(DanQingGeTimeActiveCheckRule)
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "Activity",
		Cols:        cols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      params,
		SheetMap:    sheetMap,
	})

	jsonData, _ := json.MarshalIndent(result, "", " ")
	t.Log("DanQingGeTimeActive (Multiple) Result:", string(jsonData))

	assert.False(t, result.Ok, "混合数据中包含已结束和即将结束的丹青阁活动，应该不通过")
	// 应该有至少 2 个错误/警告（1个已结束 + 1个即将结束），非丹青阁活动不应计入
	assert.True(t, len(result.ErrCells) >= 2, "应该有多个错误/警告")
}

// TestDanQingGeTimeActive_Meta 测试规则元数据
func TestDanQingGeTimeActive_Meta(t *testing.T) {
	rule := new(DanQingGeTimeActiveCheckRule)
	meta := rule.Meta()

	assert.Equal(t, json_rule.DANQINGGE_TIME_ACTIVE_CHECK, meta.Type)
	assert.Equal(t, "丹青阁活动时间校验", meta.DisplayName)
	assert.Contains(t, meta.TargetSheets, "Activity")
	assert.True(t, len(meta.ParamDefs) > 0, "应该有参数定义")
}

// ==================== 丹青阁测试数据构造函数 ====================
// Activity 表结构：Id, ActivityType, TimeType, Name, StartTime, EndTime
// 丹青阁活动特征：ActivityType == "ActTypeSkinRaffle", TimeType == "1"（绝对时间）

// fakeDanQingGeData_Future 构造活动距结束超过7天的测试数据
func fakeDanQingGeData_Future() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	futureEnd := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")
	pastStart := time.Now().AddDate(0, 0, -5).Format("2006-01-02 15:04:05")

	cols = [][]string{
		{"", "", "Id", "", "36"},
		{"", "", "ActivityType", "", "ActTypeSkinRaffle"},
		{"", "", "TimeType", "", "1"},
		{"", "", "Name", "", "红妆夜弋"},
		{"", "", "StartTime", "", pastStart},
		{"", "", "EndTime", "", futureEnd},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

// fakeDanQingGeData_Soon 构造活动离结束不到7天的测试数据
func fakeDanQingGeData_Soon() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	soonEnd := time.Now().AddDate(0, 0, 3).Format("2006-01-02 15:04:05")
	pastStart := time.Now().AddDate(0, 0, -10).Format("2006-01-02 15:04:05")

	cols = [][]string{
		{"", "", "Id", "", "37"},
		{"", "", "ActivityType", "", "ActTypeSkinRaffle"},
		{"", "", "TimeType", "", "1"},
		{"", "", "Name", "", "丹青阁测试活动"},
		{"", "", "StartTime", "", pastStart},
		{"", "", "EndTime", "", soonEnd},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

// fakeDanQingGeData_Expired 构造活动已结束的测试数据
func fakeDanQingGeData_Expired() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	pastEnd := time.Now().AddDate(0, 0, -2).Format("2006-01-02 15:04:05")
	pastStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")

	cols = [][]string{
		{"", "", "Id", "", "38"},
		{"", "", "ActivityType", "", "ActTypeSkinRaffle"},
		{"", "", "TimeType", "", "1"},
		{"", "", "Name", "", "已结束活动"},
		{"", "", "StartTime", "", pastStart},
		{"", "", "EndTime", "", pastEnd},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

// fakeDanQingGeData_Empty 构造空数据的测试数据
func fakeDanQingGeData_Empty() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	cols = [][]string{
		{"", "", "Id", ""},
		{"", "", "ActivityType", ""},
		{"", "", "EndTime", ""},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

// fakeDanQingGeData_InvalidTime 构造无效时间格式的测试数据
func fakeDanQingGeData_InvalidTime() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	cols = [][]string{
		{"", "", "Id", "", "39"},
		{"", "", "ActivityType", "", "ActTypeSkinRaffle"},
		{"", "", "TimeType", "", "1"},
		{"", "", "Name", "", "时间格式错误"},
		{"", "", "StartTime", "", "2024-01-01 00:00:00"},
		{"", "", "EndTime", "", "不是有效的时间格式"},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}

// fakeDanQingGeData_Multiple 构造多行混合数据：
// 行1: 非丹青阁活动（应被跳过）
// 行2: 丹青阁活动已结束
// 行3: 丹青阁活动即将结束
// 行4: 丹青阁活动正常
func fakeDanQingGeData_Multiple() (cols [][]string, params map[string]string, sheetMap map[string]*excelize.File) {
	params = make(map[string]string)
	params[string(json_rule.TIME_RANGE_BEFORE)] = "168h"

	pastStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02 15:04:05")
	expiredEnd := time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")
	soonEnd := time.Now().AddDate(0, 0, 3).Format("2006-01-02 15:04:05")
	futureEnd := time.Now().AddDate(0, 0, 30).Format("2006-01-02 15:04:05")

	cols = [][]string{
		// Id: 非丹青阁, 已结束丹青阁, 即将结束丹青阁, 正常丹青阁
		{"", "", "Id", "", "1", "36", "37", "38"},
		// ActivityType: Other, SkinRaffle, SkinRaffle, SkinRaffle
		{"", "", "ActivityType", "", "ActTypeLogin", "ActTypeSkinRaffle", "ActTypeSkinRaffle", "ActTypeSkinRaffle"},
		// TimeType
		{"", "", "TimeType", "", "1", "1", "1", "1"},
		// Name
		{"", "", "Name", "", "登录活动", "已结束", "即将结束", "正常活动"},
		// StartTime
		{"", "", "StartTime", "", pastStart, pastStart, pastStart, pastStart},
		// EndTime
		{"", "", "EndTime", "", expiredEnd, expiredEnd, soonEnd, futureEnd},
	}

	sheetMap = map[string]*excelize.File{}
	return cols, params, sheetMap
}
