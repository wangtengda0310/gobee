// Package activity 提供活动相关的跨表校验规则
package activity

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestActivityTaskRewardCheckRule_Meta 测试规则元数据
func TestActivityTaskRewardCheckRule_Meta(t *testing.T) {
	rule := &ActivityTaskRewardCheckRule{}
	meta := rule.Meta()

	assert.Equal(t, json_rule.ACTIVITY_TASK_REWARD_CHECK, meta.Type)
	assert.Equal(t, "活动任务奖励检查", meta.DisplayName)
	assert.Contains(t, meta.TargetSheets, "ActivityTask")
	assert.NotEmpty(t, meta.Description)
}

// TestActivityTaskRewardCheckRule_AllRewardsValid 测试所有奖励道具都存在且数量正常的情况
func TestActivityTaskRewardCheckRule_AllRewardsValid(t *testing.T) {
	// 准备 Item 表数据（道具表）
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	// 设置表头
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D1", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "D2", "string")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D3", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "client")
	_ = itemFile.SetCellValue(itemSheetName, "D4", "client")
	// 设置数据
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")
	_ = itemFile.SetCellValue(itemSheetName, "D5", "金币")
	_ = itemFile.SetCellValue(itemSheetName, "A6", "")
	_ = itemFile.SetCellValue(itemSheetName, "B6", "")
	_ = itemFile.SetCellValue(itemSheetName, "C6", "1000002")
	_ = itemFile.SetCellValue(itemSheetName, "D6", "钻石")
	_ = itemFile.SetCellValue(itemSheetName, "A7", "")
	_ = itemFile.SetCellValue(itemSheetName, "B7", "")
	_ = itemFile.SetCellValue(itemSheetName, "C7", "1000003")
	_ = itemFile.SetCellValue(itemSheetName, "D7", "武将卡")

	// 准备 ActivityTask 表数据（活动任务表）
	activityTaskFile := excelize.NewFile()
	activityTaskSheetName := "活动任务表|ActivityTask"
	_, _ = activityTaskFile.NewSheet(activityTaskSheetName)
	// 设置表头
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A3", "Id")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B3", "ActivityId")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C3", "Name")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D3", "Description")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E3", "Class")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F3", "JumpCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G3", "SubType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H3", "CompleteCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I3", "Reward")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E4", "ETaskClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F4", "ETaskJumpType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G4", "ETaskSubClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H4", "TaskCondCfg[]")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I4", "ItemCfg[]")
	// 设置数据
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A5", "100001")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B5", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C5", "每日登录")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D5", "登录游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E5", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F5", "0")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G5", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H5", "{4600}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I5", "{1000001;100}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A6", "100002")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B6", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C6", "每日对战")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D6", "进行一局游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E6", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F6", "1")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G6", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H6", "{4601}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I6", "{1000002;50}{1000003;1}")

	// 构建 sheetMap - 键名必须是完整的 Sheet 名称
	sheetMap := map[string]*excelize.File{
		"道具表|Item":           itemFile,
		"活动任务表|ActivityTask": activityTaskFile,
	}

	// 获取 ActivityTask 列数据
	activityTaskCols, err := activityTaskFile.GetCols(activityTaskSheetName)
	assert.NoError(t, err)

	// 执行检查
	rule := &ActivityTaskRewardCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activityTaskSheetName,
		Cols:        activityTaskCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// 验证结果：所有道具都存在且数量正常，应该通过
	assert.True(t, result.Ok, "所有道具都存在且数量正常时应该通过检查")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}

// TestActivityTaskRewardCheckRule_ItemNotExist 测试道具不存在的情况
func TestActivityTaskRewardCheckRule_ItemNotExist(t *testing.T) {
	// 准备 Item 表数据（道具表）- 只包含 1000001
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	// 设置表头
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D1", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "D2", "string")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D3", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "client")
	_ = itemFile.SetCellValue(itemSheetName, "D4", "client")
	// 设置数据
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")
	_ = itemFile.SetCellValue(itemSheetName, "D5", "金币")

	// 准备 ActivityTask 表数据 - 包含不存在的道具 9999999
	activityTaskFile := excelize.NewFile()
	activityTaskSheetName := "活动任务表|ActivityTask"
	_, _ = activityTaskFile.NewSheet(activityTaskSheetName)
	// 设置表头
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A3", "Id")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B3", "ActivityId")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C3", "Name")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D3", "Description")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E3", "Class")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F3", "JumpCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G3", "SubType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H3", "CompleteCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I3", "Reward")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E4", "ETaskClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F4", "ETaskJumpType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G4", "ETaskSubClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H4", "TaskCondCfg[]")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I4", "ItemCfg[]")
	// 设置数据 - 包含不存在的道具
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A5", "100001")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B5", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C5", "每日登录")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D5", "登录游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E5", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F5", "0")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G5", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H5", "{4600}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I5", "{1000001;100}{9999999;1}")

	// 构建 sheetMap
	sheetMap := map[string]*excelize.File{
		"道具表|Item":           itemFile,
		"活动任务表|ActivityTask": activityTaskFile,
	}

	// 获取 ActivityTask 列数据
	activityTaskCols, err := activityTaskFile.GetCols(activityTaskSheetName)
	assert.NoError(t, err)

	// 执行检查
	rule := &ActivityTaskRewardCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activityTaskSheetName,
		Cols:        activityTaskCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// 验证结果：存在不存在的道具，应该失败
	assert.False(t, result.Ok, "存在不存在的道具时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "9999999", "错误信息应该包含不存在的道具ID")
	assert.Contains(t, result.ErrCells[0].Reason, "每日登录", "错误信息应该包含任务名称")
}

// TestActivityTaskRewardCheckRule_ItemCountInvalid 测试道具数量异常的情况
func TestActivityTaskRewardCheckRule_ItemCountInvalid(t *testing.T) {
	// 准备 Item 表数据
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	// 设置表头
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D1", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "D2", "string")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D3", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "client")
	_ = itemFile.SetCellValue(itemSheetName, "D4", "client")
	// 设置数据
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")
	_ = itemFile.SetCellValue(itemSheetName, "D5", "金币")
	_ = itemFile.SetCellValue(itemSheetName, "A6", "")
	_ = itemFile.SetCellValue(itemSheetName, "B6", "")
	_ = itemFile.SetCellValue(itemSheetName, "C6", "1000002")
	_ = itemFile.SetCellValue(itemSheetName, "D6", "钻石")

	// 准备 ActivityTask 表数据 - 包含数量为0的道具
	activityTaskFile := excelize.NewFile()
	activityTaskSheetName := "活动任务表|ActivityTask"
	_, _ = activityTaskFile.NewSheet(activityTaskSheetName)
	// 设置表头
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A3", "Id")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B3", "ActivityId")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C3", "Name")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D3", "Description")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E3", "Class")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F3", "JumpCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G3", "SubType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H3", "CompleteCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I3", "Reward")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E4", "ETaskClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F4", "ETaskJumpType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G4", "ETaskSubClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H4", "TaskCondCfg[]")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I4", "ItemCfg[]")
	// 设置数据 - 数量为0
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A5", "100001")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B5", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C5", "每日登录")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D5", "登录游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E5", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F5", "0")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G5", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H5", "{4600}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I5", "{1000001;100}{1000002;0}")

	// 构建 sheetMap
	sheetMap := map[string]*excelize.File{
		"道具表|Item":           itemFile,
		"活动任务表|ActivityTask": activityTaskFile,
	}

	// 获取 ActivityTask 列数据
	activityTaskCols, err := activityTaskFile.GetCols(activityTaskSheetName)
	assert.NoError(t, err)

	// 执行检查
	rule := &ActivityTaskRewardCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activityTaskSheetName,
		Cols:        activityTaskCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// 验证结果：存在数量为0的道具，应该失败
	assert.False(t, result.Ok, "存在数量为0的道具时应该失败")
	assert.Len(t, result.ErrCells, 1, "应该有一个错误")
	assert.Contains(t, result.ErrCells[0].Reason, "1000002", "错误信息应该包含数量异常的道具ID")
	assert.Contains(t, result.ErrCells[0].Reason, "数量=0", "错误信息应该包含数量信息")
}

// TestActivityTaskRewardCheckRule_NoItemSheet 测试 Item 表不存在的情况
func TestActivityTaskRewardCheckRule_NoItemSheet(t *testing.T) {
	// 准备 ActivityTask 表数据
	activityTaskFile := excelize.NewFile()
	activityTaskSheetName := "活动任务表|ActivityTask"
	_, _ = activityTaskFile.NewSheet(activityTaskSheetName)
	// 设置表头
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A3", "Id")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B3", "ActivityId")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C3", "Name")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D3", "Description")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E3", "Class")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F3", "JumpCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G3", "SubType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H3", "CompleteCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I3", "Reward")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E4", "ETaskClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F4", "ETaskJumpType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G4", "ETaskSubClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H4", "TaskCondCfg[]")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I4", "ItemCfg[]")
	// 设置数据
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A5", "100001")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B5", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C5", "每日登录")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D5", "登录游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E5", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F5", "0")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G5", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H5", "{4600}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I5", "{1000001;100}")

	// 构建 sheetMap - 不包含 Item 表
	sheetMap := map[string]*excelize.File{
		"活动任务表|ActivityTask": activityTaskFile,
	}

	// 获取 ActivityTask 列数据
	activityTaskCols, err := activityTaskFile.GetCols(activityTaskSheetName)
	assert.NoError(t, err)

	// 执行检查
	rule := &ActivityTaskRewardCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activityTaskSheetName,
		Cols:        activityTaskCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// 验证结果：Item 表不存在时应该报错
	assert.False(t, result.Ok, "Item 表不存在时应该报错")
	assert.Contains(t, result.Reason, "未找到 Item 表")
}

// TestActivityTaskRewardCheckRule_EmptyRewardColumn 测试 Reward 列为空的情况
func TestActivityTaskRewardCheckRule_EmptyRewardColumn(t *testing.T) {
	// 准备 Item 表数据
	itemFile := excelize.NewFile()
	itemSheetName := "道具表|Item"
	_, _ = itemFile.NewSheet(itemSheetName)
	// 设置表头
	_ = itemFile.SetCellValue(itemSheetName, "A1", "")
	_ = itemFile.SetCellValue(itemSheetName, "B1", "")
	_ = itemFile.SetCellValue(itemSheetName, "C1", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D1", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A2", "")
	_ = itemFile.SetCellValue(itemSheetName, "B2", "")
	_ = itemFile.SetCellValue(itemSheetName, "C2", "int")
	_ = itemFile.SetCellValue(itemSheetName, "D2", "string")
	_ = itemFile.SetCellValue(itemSheetName, "A3", "")
	_ = itemFile.SetCellValue(itemSheetName, "B3", "")
	_ = itemFile.SetCellValue(itemSheetName, "C3", "Id")
	_ = itemFile.SetCellValue(itemSheetName, "D3", "Name")
	_ = itemFile.SetCellValue(itemSheetName, "A4", "")
	_ = itemFile.SetCellValue(itemSheetName, "B4", "")
	_ = itemFile.SetCellValue(itemSheetName, "C4", "client")
	_ = itemFile.SetCellValue(itemSheetName, "D4", "client")
	// 设置数据
	_ = itemFile.SetCellValue(itemSheetName, "A5", "")
	_ = itemFile.SetCellValue(itemSheetName, "B5", "")
	_ = itemFile.SetCellValue(itemSheetName, "C5", "1000001")
	_ = itemFile.SetCellValue(itemSheetName, "D5", "金币")

	// 准备 ActivityTask 表数据 - Reward 列为空
	activityTaskFile := excelize.NewFile()
	activityTaskSheetName := "活动任务表|ActivityTask"
	_, _ = activityTaskFile.NewSheet(activityTaskSheetName)
	// 设置表头
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I1", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I2", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A3", "Id")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B3", "ActivityId")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C3", "Name")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D3", "Description")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E3", "Class")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F3", "JumpCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G3", "SubType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H3", "CompleteCond")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I3", "Reward")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B4", "int")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D4", "string")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E4", "ETaskClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F4", "ETaskJumpType")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G4", "ETaskSubClass")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H4", "TaskCondCfg[]")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I4", "ItemCfg[]")
	// Reward 列为空
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "A5", "100001")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "B5", "4")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "C5", "每日登录")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "D5", "登录游戏")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "E5", "ActivityDaily")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "F5", "0")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "G5", "")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "H5", "{4600}")
	_ = activityTaskFile.SetCellValue(activityTaskSheetName, "I5", "")

	// 构建 sheetMap
	sheetMap := map[string]*excelize.File{
		"道具表|Item":           itemFile,
		"活动任务表|ActivityTask": activityTaskFile,
	}

	// 获取 ActivityTask 列数据
	activityTaskCols, err := activityTaskFile.GetCols(activityTaskSheetName)
	assert.NoError(t, err)

	// 执行检查
	rule := &ActivityTaskRewardCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   activityTaskSheetName,
		Cols:        activityTaskCols,
		StartRowIdx: excelio.MJS_FIXED_ROWS_NUM,
		EndIndex:    1000,
		Params:      nil,
		SheetMap:    sheetMap,
	})

	// 验证结果：Reward 列为空时应该通过
	assert.True(t, result.Ok, "Reward 列为空时应该通过")
	assert.Empty(t, result.ErrCells, "不应该有错误")
}
