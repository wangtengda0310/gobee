package coded_rules

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

func TestDrawskinTenCost_Meta(t *testing.T) {
	rule := &DrawskinTenCostCheckRule{}
	meta := rule.Meta()
	assert.Equal(t, json_rule.DRAWSKIN_TEN_COST_CHECK, meta.Type)
	assert.Contains(t, meta.TargetSheets, "DrawSkin")
}

func TestDrawskinTenCost_AllValid(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "TenDropRule", "", "501"},
		{"", "", "TenItemCost", "", "{1000005;10}"},
	}
	rule := &DrawskinTenCostCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.True(t, result.Ok)
	assert.Empty(t, result.ErrCells)
}

func TestDrawskinTenCost_MissingTenCost(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "TenDropRule", "", "501"},
		{"", "", "TenItemCost", "", ""},
	}
	rule := &DrawskinTenCostCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.True(t, result.Ok, "Warning 性质 Ok=true")
	assert.NotEmpty(t, result.ErrCells)
	assert.Contains(t, result.ErrCells[0].Reason, "TenItemCost 未配置")
}

func TestDrawskinTenCost_NoTenDropRule(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "Name", "", "池A"},
		{"", "", "TenDropRule", "", "0"},
		{"", "", "TenItemCost", "", ""},
	}
	rule := &DrawskinTenCostCheckRule{}
	result := rule.Check(json_rule.CheckParam{
		SheetName:   "皮肤抽奖|DrawSkin",
		Cols:        cols,
		StartRowIdx: 4,
		EndIndex:    5,
	})
	assert.True(t, result.Ok, "TenDropRule=0 时不报错")
	assert.Empty(t, result.ErrCells)
}
