package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

func TestDrawskinTimeRange_Meta(t *testing.T) {
	rule := &DrawskinTimeRangeCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWSKIN_TIME_RANGE_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
}

func TestDrawskinTimeRange_AllValid(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "StartTime", "", "2025-01-01 00:00:00"},
		{"", "", "EndTime", "", "2025-06-01 00:00:00"},
	}
	rule := &DrawskinTimeRangeCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.True(t, result.Ok)
	assert.Empty(t, result.ErrCells)
}

func TestDrawskinTimeRange_InvalidFormat(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "StartTime", "", "invalid-date"},
		{"", "", "EndTime", "", "2025-06-01 00:00:00"},
	}
	rule := &DrawskinTimeRangeCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.True(t, result.Ok, "格式非法是 Warning，Ok 应为 true")
	assert.NotEmpty(t, result.ErrCells)
}

func TestDrawskinTimeRange_StartAfterEnd(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "StartTime", "", "2025-12-01 00:00:00"},
		{"", "", "EndTime", "", "2025-01-01 00:00:00"},
	}
	rule := &DrawskinTimeRangeCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.False(t, result.Ok, "StartTime 晚于 EndTime 是 Error")
	assert.NotEmpty(t, result.ErrCells)
}

func TestDrawskinTimeRange_EmptyTimes(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "StartTime", "", ""},
		{"", "", "EndTime", "", ""},
	}
	rule := &DrawskinTimeRangeCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.False(t, result.Ok, "StartTime 和 EndTime 均为空时应报错")
	assert.NotEmpty(t, result.ErrCells)
}
