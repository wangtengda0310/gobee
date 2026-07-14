package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

// 测试辅助函数（非 Test 函数，避免 go test 误识别）
func drawskinDataValidityCols() [][]string {
	return [][]string{
		{"", "", "Id", "", "1", "2", "3"},           // Id列
		{"", "", "Name", "", "池A", "池B", "池C"},      // Name列
		{"", "", "DailyLimit", "", "10", "5", "-1"}, // DailyLimit列
	}
}

func TestDrawskinDataValidity_Meta(t *testing.T) {
	rule := &DrawskinDataValidityCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWSKIN_DATA_VALIDITY_CHECK, meta.Type)
	assert.Equal(t, "DrawSkin基础数据检查", meta.DisplayName)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
}

func TestDrawskinDataValidity_AllValid(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1", "2", "3"},
		{"", "", "Name", "", "池A", "池B", "池C"},
		{"", "", "DailyLimit", "", "10", "5", "0"},
	}
	rule := &DrawskinDataValidityCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    7,
	})
	assert.True(t, result.Ok, "所有数据有效时应该通过")
	assert.Empty(t, result.ErrCells)
}

func TestDrawskinDataValidity_DuplicateId(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1", "2", "1"},
		{"", "", "Name", "", "池A", "池B", "池C"},
		{"", "", "DailyLimit", "", "10", "5", "10"},
	}
	rule := &DrawskinDataValidityCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    7,
	})
	assert.False(t, result.Ok)
	assert.NotEmpty(t, result.ErrCells)
	assert.Contains(t, result.ErrCells[0].Reason, "重复")
}

func TestDrawskinDataValidity_InvalidDailyLimit(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1", "2"},
		{"", "", "Name", "", "池A", "池B"},
		{"", "", "DailyLimit", "", "10", "-1"},
	}
	rule := &DrawskinDataValidityCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    6,
	})
	assert.False(t, result.Ok)
	assert.Len(t, result.ErrCells, 1)
	assert.Contains(t, result.ErrCells[0].Reason, "DailyLimit")
	assert.Contains(t, result.ErrCells[0].Reason, ">= 0")
}

func TestDrawskinDataValidity_MissingIdColumn(t *testing.T) {
	cols := [][]string{
		{"", "", "Name", "", "池A"},
	}
	rule := &DrawskinDataValidityCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.False(t, result.Ok, "缺少 Id 列时应报错")
	assert.Contains(t, result.Reason, "未找到 Id 列")
}
