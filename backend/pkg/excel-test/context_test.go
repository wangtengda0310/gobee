package exceltest

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
)

func TestCheckContext_AddParseErrors(t *testing.T) {
	ctx := NewCheckContext()

	errors := []*SheetParseError{
		{FileName: "test.xlsx", SheetName: "Sheet1", Error: "表头行数不足"},
		{FileName: "test.xlsx", SheetName: "Sheet2", Error: "类型不匹配"},
	}

	ctx.AddParseErrors(errors)

	assert.Len(t, ctx.ParseErrors, 2)
	assert.Equal(t, "表头行数不足", ctx.ParseErrors[0].Error)
}

func TestCheckContext_AddColResults(t *testing.T) {
	ctx := NewCheckContext()

	sheetName := "Hero"
	results := []*json_rule.ColCheckResult{
		{SheetName: &sheetName, Ok: false, Reason: "唯一性检查失败"},
		{SheetName: &sheetName, Ok: true},
	}

	ctx.AddColResults(results)

	assert.Len(t, ctx.ColResults, 2)
}

func TestCheckContext_AddTableResults(t *testing.T) {
	ctx := NewCheckContext()

	sheetName := "Arena"
	results := []*json_rule.TableCheckResult{
		{SheetName: &sheetName, Ok: false, Reason: "赛季时间无效"},
	}

	ctx.AddTableResults(results)

	assert.Len(t, ctx.TableResults, 1)
}

func TestCheckContext_ToResult(t *testing.T) {
	ctx := NewCheckContext()

	// 添加各种结果
	ctx.AddParseErrors([]*SheetParseError{
		{FileName: "test.xlsx", SheetName: "Sheet1", Error: "解析错误"},
	})

	sheetName := "Hero"
	ctx.AddColResults([]*json_rule.ColCheckResult{
		{SheetName: &sheetName, Ok: false},
	})
	ctx.AddTableResults([]*json_rule.TableCheckResult{
		{SheetName: &sheetName, Ok: true},
	})

	result := ctx.ToResult()

	assert.Len(t, result.ParseErrors, 1)
	assert.Len(t, result.ColResults, 1)
	assert.Len(t, result.TableResults, 1)
}

func TestCheckContext_HasErrors(t *testing.T) {
	t.Run("有解析错误", func(t *testing.T) {
		ctx := NewCheckContext()
		ctx.AddParseErrors([]*SheetParseError{
			{FileName: "test.xlsx", SheetName: "Sheet1", Error: "错误"},
		})
		assert.True(t, ctx.HasErrors())
	})

	t.Run("有列级检查失败", func(t *testing.T) {
		ctx := NewCheckContext()
		sheetName := "Hero"
		ctx.AddColResults([]*json_rule.ColCheckResult{
			{SheetName: &sheetName, Ok: false},
		})
		assert.True(t, ctx.HasErrors())
	})

	t.Run("有表级检查失败", func(t *testing.T) {
		ctx := NewCheckContext()
		sheetName := "Hero"
		ctx.AddTableResults([]*json_rule.TableCheckResult{
			{SheetName: &sheetName, Ok: false},
		})
		assert.True(t, ctx.HasErrors())
	})

	t.Run("无错误", func(t *testing.T) {
		ctx := NewCheckContext()
		assert.False(t, ctx.HasErrors())
	})

	t.Run("全部成功不算错误", func(t *testing.T) {
		ctx := NewCheckContext()
		sheetName := "Hero"
		ctx.AddColResults([]*json_rule.ColCheckResult{
			{SheetName: &sheetName, Ok: true},
		})
		ctx.AddTableResults([]*json_rule.TableCheckResult{
			{SheetName: &sheetName, Ok: true},
		})
		assert.False(t, ctx.HasErrors())
	})
}

func TestCheckContext_Empty(t *testing.T) {
	ctx := NewCheckContext()

	assert.Empty(t, ctx.ParseErrors)
	assert.Empty(t, ctx.ColResults)
	assert.Empty(t, ctx.TableResults)
	assert.False(t, ctx.HasErrors())

	result := ctx.ToResult()
	assert.NotNil(t, result)
	assert.Empty(t, result.ParseErrors)
	assert.Empty(t, result.ColResults)
	assert.Empty(t, result.TableResults)
}
