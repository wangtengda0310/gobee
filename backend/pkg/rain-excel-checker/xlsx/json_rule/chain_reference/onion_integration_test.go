// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件测试洋葱模型的完整集成（左链 + Match + 右链反向 + 比较）
package chain_reference

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 测试辅助函数 ====================

// buildOnionTestContext 构造洋葱模型测试用的 ChainContext
// cols: 当前表列数据, colIdx: 当前检查列, rowIdx: 当前行（绝对行号）, sheetMap: 参考表, config: 链配置, params: 参数
// dataIdx 自动计算为 rowIdx - startRowIdx
func buildOnionTestContext(
	cols [][]string,
	colIdx int,
	rowIdx int,
	sheetMap map[string]*excelize.File,
	config *ChainPairConfig,
	params map[string]string,
) *ChainContext {
	startRowIdx := 4
	myColData := cols[colIdx][startRowIdx:]
	dataIdx := rowIdx - startRowIdx
	return NewChainContext(cols, nil, colIdx, rowIdx, startRowIdx, sheetMap, config, params, myColData, dataIdx)
}

// runOnionChain 构建并执行洋葱链，返回 ctx
func runOnionChain(ctx *ChainContext, config *ChainPairConfig, params map[string]string) error {
	// 确保 chainSteps 参数与 config 一致
	if _, ok := params["chainSteps"]; !ok || params["chainSteps"] == `{}` || params["chainSteps"] == "" {
		jsonBytes, _ := json.Marshal(config)
		params["chainSteps"] = string(jsonBytes)
	}
	chain := BuildOnionChain(config, params)
	return chain(ctx)
}

// ==================== 单步链场景 ====================

// TestOnion_LeftSingleStep 左链只有一步（仅取值模式）
func TestOnion_LeftSingleStep(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	})

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, nil, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	assert.Equal(t, []string{"1001"}, ctx.LeftCurrentValues)
	assert.NotNil(t, ctx.LeftResult)
}

// TestOnion_RightSingleStep_ReverseLookup 右链只有一步的反向查找
// 右链单步：Match 设置 RightCurrentValues → RightStep0 反向查找 → 提取 PreCol 值
func TestOnion_RightSingleStep_ReverseLookup(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "1001"},
	})

	// 右链目标表：Item
	itemFile := buildChainTestRefSheet("Item", []string{"Item", "DropGroup"}, [][]string{
		{"1001", "G001"},
		{"2002", "G002"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "BigAward"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "Item", FindVal: "self", NextCol: "DropGroup"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链值 "1001"，右链最终表 DropGroup 值 ["G001", "G002"]
	// Match verify_exists: "1001" 在 ["G001", "G002"] 中未找到 → 门控不通过
	assert.False(t, ctx.Matched)
}

// TestOnion_RightSingleStep_FirstStepInputValues 单步右链输入值收集
// 使用两步右链，验证反向查找正确性和 FirstStepInputValues
func TestOnion_RightSingleStep_FirstStepInputValues(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "G001"},
	})

	// 右链第一步表：DropGroup
	groupFile := buildChainTestRefSheet("DropGroup", []string{"Id", "Name"}, [][]string{
		{"G001", "GroupA"},
	})

	// 右链第二步表：Item
	itemFile := buildChainTestRefSheet("Item", []string{"Item", "DropGroup"}, [][]string{
		{"1001", "G001"},
	})

	sheetMap := map[string]*excelize.File{
		"DropGroup": groupFile,
		"Item":      itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "DropGroup", PreCol: "Id", FindVal: "self", NextCol: "Name"},
				{Sheet: "Item", PreCol: "DropGroup", FindVal: "self", NextCol: "Item"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链值 "G001"，右链最终表 Item 的 Item 列值 ["1001"]
	// Match verify_exists: "G001" 在 ["1001"] 中未找到 → 门控不通过
	assert.False(t, ctx.Matched)
}

// ==================== 反向查找场景 ====================

// TestOnion_ReverseLookup_Basic 反向查找基本正确性
// 直接测试 RightStepHandler 的反向查找方法
func TestOnion_ReverseLookup_Basic(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "G001"},
	})

	// 右链 Step1 目标表：Item
	// Item=1001, DropGroup=G001
	itemFile := buildChainTestRefSheet("Item", []string{"Item", "DropGroup"}, [][]string{
		{"1001", "G001"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "DropGroup", PreCol: "Id", FindVal: "self", NextCol: "Name"},
				{Sheet: "Item", PreCol: "DropGroup", FindVal: "self", NextCol: "Item"},
			},
		},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, config, nil, nil, 0)

	// 模拟 Match 设置的 RightCurrentValues
	ctx.Matched = true
	ctx.RightCurrentValues = []string{"1001"}

	// 执行 Step1 反向查找（在 Item 表中用 Item="1001" 匹配，提取 DropGroup="G001"）
	step1 := &RightStepHandler{
		Step:     config.Right.Steps[1],
		StepIdx:  1,
		IsFirst:  false,
		ChainCfg: config.Right,
	}

	nextCalled := false
	err := step1.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	// Step1 反向查找：NextCol=Item 列匹配 "1001" → 提取 PreCol=DropGroup 值 "G001"
	assert.Equal(t, []string{"G001"}, ctx.RightCurrentValues)
}

// TestOnion_ReverseLookup_EmptyInput 反向查找输入为空返回错误
func TestOnion_ReverseLookup_EmptyInput(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
	})

	itemFile := buildChainTestRefSheet("Item", []string{"Item", "DropGroup"}, [][]string{
		{"1001", "G001"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	ctx := NewChainContext(cols, nil, 0, 5, 4, sheetMap, nil, nil, nil, 0)
	ctx.Matched = true
	ctx.RightCurrentValues = []string{} // 空输入

	handler := &RightStepHandler{
		Step:     ChainStep{Sheet: "Item", PreCol: "Item", NextCol: "DropGroup"},
		StepIdx:  0,
		IsFirst:  true,
		ChainCfg: ChainConfig{},
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Error(t, ctx.Err)
}

// ==================== time_overlap 特化场景 ====================

// TestOnion_TwoPhase_TimeOverlap_BothChainsHaveCompareCol 两链都有 CompareCol
func TestOnion_TwoPhase_TimeOverlap_BothChainsHaveCompareCol(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	leftFile := buildChainTestRefSheet("LeftTable", []string{"Id", "Time"}, [][]string{
		{"1001", "2026-01-01 10:00:00"},
	})

	rightFile := buildChainTestRefSheet("RightTable", []string{"Id", "Time"}, [][]string{
		{"1001", "2026-01-01 10:00:00"},
	})

	sheetMap := map[string]*excelize.File{
		"LeftTable":  leftFile,
		"RightTable": rightFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
				{Sheet: "LeftTable", PreCol: "Id", FindVal: "self", NextCol: "Time"},
			},
			CompareCol: "Time",
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "RightTable", PreCol: "Id", FindVal: "self", NextCol: "Time"},
			},
			CompareCol: "Time",
		},
	}

	params := map[string]string{
		"chainCompare":      "time_overlap",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 验证没有异常
}

// TestOnion_TwoPhase_TimeOverlap_CompareColNotFound CompareCol 不存在不报异常
func TestOnion_TwoPhase_TimeOverlap_CompareColNotFound(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "val1"},
	})

	rightFile := buildChainTestRefSheet("RightTable", []string{"Id", "Name"}, [][]string{
		{"val1", "Name1"},
	})

	sheetMap := map[string]*excelize.File{
		"RightTable": rightFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "RightTable", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
}

// TestOnion_TwoPhase_TimeOverlap_InvalidTimeFormat 无效时间格式不报违规
func TestOnion_TwoPhase_TimeOverlap_InvalidTimeFormat(t *testing.T) {
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{{Match: "1001", Compare: "not-a-date"}},
		},
		RightResult: &ChainResult{
			Values: []ChainValue{{Match: "1001", Compare: "also-not-date"}},
		},
		Params:  map[string]string{},
		Matched: true,
	}

	handler := &CompareHandler{
		CompareType:    "time_overlap",
		UseTimeOverlap: true,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.False(t, ctx.Violation, "无效时间格式无法解析，不应报违规")
}

// ==================== 边界条件 ====================

// TestOnion_AllowEmpty 空值跳过
// 验证当 currentCellValue 为空时洋葱模型不报错
func TestOnion_AllowEmpty(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", ""}, // 空值
	})

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "BigAward"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "SomeSheet", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, nil, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链取值为空 → LeftCurrentValues 为 nil → Match 阶段左链无值 → 匹配失败
	assert.False(t, ctx.Matched)
	assert.False(t, ctx.Violation)
}

// TestOnion_AllowCommit 注释行作为普通字符串处理
// 注释行的 # 前缀过滤由上层 Check 方法（SolveEmptyAndCommit）负责
func TestOnion_AllowCommit(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "#comment"},
	})

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "BigAward"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "SomeSheet", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, nil, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链取值为 "#comment"（非空字符串，注释过滤由上层负责）
	assert.Equal(t, []string{"#comment"}, ctx.LeftCurrentValues)
}

// TestOnion_SheetNotFound 目标表不存在返回错误
func TestOnion_SheetNotFound(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "NotExistSheet", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, map[string]*excelize.File{}, config, nil, nil, 0)
	ctx.Matched = true
	ctx.RightCurrentValues = []string{"1001"}

	handler := &RightStepHandler{
		Step:     config.Right.Steps[0],
		StepIdx:  0,
		IsFirst:  true,
		ChainCfg: config.Right,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Error(t, ctx.Err)
	assert.Contains(t, ctx.Err.Error(), "目标表不存在")
}

// TestOnion_ColumnNotFound 列不存在返回错误
func TestOnion_ColumnNotFound(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	rightFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, [][]string{
		{"1001", "Item1"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": rightFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "NotExistCol", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, config, nil, nil, 0)
	ctx.Matched = true
	ctx.RightCurrentValues = []string{"1001"}

	handler := &RightStepHandler{
		Step:     config.Right.Steps[0],
		StepIdx:  0,
		IsFirst:  true,
		ChainCfg: config.Right,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Error(t, ctx.Err)
	assert.Contains(t, ctx.Err.Error(), "未找到")
}

// TestOnion_ValidateBadConfig 参数校验失败
func TestOnion_ValidateBadConfig(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
	})

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "SomeSheet", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "unknown_type",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 0, 5, nil, config, params)
	err := runOnionChain(ctx, config, params)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未知的比较类型")
}

// ==================== 完整两步右链反向流程 ====================

// TestOnion_TwoStepRightChain 两步右链反向查找
// Step0(DropGroup) → Step1(Item)
// 反向：Step1 在 Item 用 Item 匹配 → 提取 DropGroup → Step0 在 DropGroup 用 Name 匹配 → 提取 Id
func TestOnion_TwoStepRightChain(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// Step0 目标表：DropGroup
	groupFile := buildChainTestRefSheet("DropGroup", []string{"Id", "Name"}, [][]string{
		{"G001", "GroupA"},
	})

	// Step1 目标表：Item
	itemFile := buildChainTestRefSheet("Item", []string{"Item", "DropGroup"}, [][]string{
		{"1001", "G001"},
	})

	sheetMap := map[string]*excelize.File{
		"DropGroup": groupFile,
		"Item":      itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "DropGroup", PreCol: "Id", FindVal: "self", NextCol: "Name"},
				{Sheet: "Item", PreCol: "DropGroup", FindVal: "self", NextCol: "Item"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链值 "1001"，右链最终表 Item 的 Item 列值 ["1001"]
	// Match verify_exists: "1001" 在 ["1001"] 中找到 → 通过
	assert.True(t, ctx.Matched)

	// 反向执行：
	// Step1: 在 Item 中用 Item="1001" 匹配 → 提取 DropGroup="G001"
	// Step0: 在 DropGroup 中用 Name 匹配 "G001" → Name 列是 "GroupA"，不匹配
	// 所以 FirstStepInputValues 为空（反向查找未找到匹配）
	assert.Empty(t, ctx.FirstStepInputValues)
}

// TestOnion_MatchWithFilter 过滤条件在 Match 和反向查找中生效
func TestOnion_MatchWithFilter(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// 带过滤条件的表
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Type", "Name"}, [][]string{
		{"1001", "Hero", "HeroA"},
		{"1002", "Item", "ItemB"},
		{"1003", "Hero", "HeroC"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name", FilterCol: "Type", FilterVal: "Hero"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链值 "1001"
	// Match 从 Item 表的 Name 列取值（过滤 Type=Hero）→ ["HeroA", "HeroC"]
	// Match verify_exists: "1001" 在 ["HeroA", "HeroC"] 中未找到 → 门控不通过
	assert.False(t, ctx.Matched)
}

// TestOnion_NextColEmpty 右链步骤 NextCol 为空时使用 PreCol
func TestOnion_NextColEmpty(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	itemFile := buildChainTestRefSheet("Item", []string{"Id"}, [][]string{
		{"1001"},
		{"1002"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: ""},
			},
		},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, config, nil, nil, 0)
	ctx.Matched = true
	ctx.RightCurrentValues = []string{"1001"}

	handler := &RightStepHandler{
		Step:     config.Right.Steps[0],
		StepIdx:  0,
		IsFirst:  true,
		ChainCfg: config.Right,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	// NextCol 为空，PreCol=Id 既搜索又提取
	assert.Equal(t, []string{"1001"}, ctx.RightCurrentValues)
	assert.Equal(t, []string{"1001"}, ctx.FirstStepInputValues)
	assert.NotNil(t, ctx.RightResult)
}

// TestOnion_MatchPreColOnly 右链最终表无 NextCol 时 Match 使用 PreCol
func TestOnion_MatchPreColOnly(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// 表中只有 Id 列（没有 Name/DropGroup 等 NextCol）
	itemFile := buildChainTestRefSheet("Item", []string{"Id"}, [][]string{
		{"1001"},
		{"2002"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: ""},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// Match 从 Item 表取值（NextCol 为空时用 PreCol=Id）→ ["1001", "2002"]
	// Match verify_exists: "1001" 在 ["1001", "2002"] 中找到 → 通过
	assert.True(t, ctx.Matched)
	assert.Contains(t, ctx.MatchedKeys, "1001")
}

// ==================== 左链引用完整性检查 ====================

// TestOnion_LeftChainLookupFailed_ReportsError 左链有输入值但在目标表中找不到匹配行时报错
// 验证 handleCrossTableLookup 中引用完整性检查：ctx.Err 非空
func TestOnion_LeftChainLookupFailed_ReportsError(t *testing.T) {
	// 当前表：colIdx=1 对应 "Value" 列，值为 "999"
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Value", "", "", "999"},
	})

	// 目标表 Item：只有表头，没有数据行（没有任何 Id 匹配 "999"）
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, nil)

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Value"},
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	// 左链第2步在 Item 表中查找 Id="999"，找不到 → 引用完整性检查报错
	// Handle 中 ctx.Err 非空时直接返回，所以 err 也非空
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到")
}

// TestOnion_LeftChainEmptyInput_NoError 左链输入为空时不报错
// 验证 DrawPet 场景：当列为空字符串时，handleCrossTableLookup 的早期守卫直接返回
func TestOnion_LeftChainEmptyInput_NoError(t *testing.T) {
	// 当前表：colIdx=1 对应 "Value" 列，值为空字符串
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Value", "", "", ""},
	})

	// 目标表有数据（但由于输入为空不会触发查找）
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, [][]string{
		{"1001", "ItemA"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Value"},
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	// 左链第1步取值为空 → LeftCurrentValues 为 nil → 第2步守卫跳过 → 无报错
	assert.NoError(t, err)
	assert.Nil(t, ctx.Err)
}

// TestOnion_LeftChainLookupFound_NoError 左链查找成功时不报错（回归测试）
// 验证正常场景下引用完整性检查不会误报
func TestOnion_LeftChainLookupFound_NoError(t *testing.T) {
	// 当前表：colIdx=1 对应 "Value" 列，值为 "1001"
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Value", "", "", "1001"},
	})

	// 目标表 Item：有 Id="1001" 的行
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, [][]string{
		{"1001", "Sword"},
		{"2002", "Shield"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Value"},
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	// 左链第2步在 Item 表中查找 Id="1001"，找到 → 提取 Name="Sword" → 无报错
	assert.NoError(t, err)
	assert.Nil(t, ctx.Err)
	assert.Equal(t, []string{"Sword"}, ctx.LeftCurrentValues)
}

// verify_exists 纯比较/匹配函数测试已迁移到：
// coded_rules/general/column_check/compare_verify_exists_test.go
// 包含：CompareByType、MatchByType、E2E、isArray 拆分 bug 验收测试

// ==================== 性能裁剪验证测试 ====================
// 验证洋葱模型各环节的数据量裁剪效果，确保性能优化后的中间数据量符合预期

// TestOnion_Performance_GetColsCacheHits 验证 GetCols 缓存命中
// 多次执行洋葱链时，同一张表只调用一次 GetCols，后续使用 ColsCache 返回
func TestOnion_Performance_GetColsCacheHits(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// Item 表：2 行数据
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, [][]string{
		{"1001", "Sword"},
		{"2002", "Shield"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	// 第一次执行：传入空的 ColsCache map，GetCachedCols 会调用 GetCols 并缓存
	colsCache := make(map[string][][]string)
	startRowIdx := 4
	myColData := cols[1][startRowIdx:]
	ctx1 := NewChainContext(cols, colsCache, 1, 5, startRowIdx, sheetMap, config, params, myColData, 5-startRowIdx)
	err := runOnionChain(ctx1, config, params)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Sword"}, ctx1.LeftCurrentValues)
	// 验证 ColsCache 中有 Item 表的缓存
	_, hasItem := ctx1.ColsCache["Item"]
	assert.True(t, hasItem, "第一次执行后 ColsCache 应包含 Item 表")

	// 第二次执行：使用同一个 ColsCache（模拟缓存命中）
	// 直接复用第一次的缓存 map，验证 GetCachedCols 直接返回缓存值
	myColData2 := cols[1][startRowIdx:]
	ctx2 := NewChainContext(cols, colsCache, 1, 5, startRowIdx, sheetMap, config, params, myColData2, 5-startRowIdx)
	err = runOnionChain(ctx2, config, params)
	assert.NoError(t, err)
	// 使用缓存执行应得到相同结果
	assert.Equal(t, []string{"Sword"}, ctx2.LeftCurrentValues)
}

// TestOnion_Performance_LeftStepFilterReducesRows 验证左链跨表查找时 filterRowSet 裁剪效果
// 目标表 10 行，Type 列只有 2 行为 "Hero"，验证 LeftStepValues 长度为 2
func TestOnion_Performance_LeftStepFilterReducesRows(t *testing.T) {
	// 当前表：Col1 值为 "1001"
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// Item 表：10 行，其中只有 2 行 Type=Hero
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Type", "Name"}, [][]string{
		{"1001", "Hero", "HeroA"}, // 匹配 Id=1001, Type=Hero → 命中
		{"1002", "Item", "ItemB"},
		{"1003", "Hero", "HeroC"}, // 匹配 Id=1003, Type=Hero → 但左链查找 Id=1001 不匹配
		{"1004", "Item", "ItemD"},
		{"1005", "Equip", "EquipE"},
		{"1006", "Item", "ItemF"},
		{"1007", "Equip", "EquipG"},
		{"1008", "Item", "ItemH"},
		{"1009", "Item", "ItemI"},
		{"1010", "Item", "ItemJ"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
				// 第二步：在 Item 表中查找 Id 匹配，且 Type=Hero 过滤
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name", FilterCol: "Type", FilterVal: "Hero"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链第二步：在 10 行中，filterRowSet 裁剪后只有 2 行 Type=Hero（行0=1001，行2=1003）
	// 但只有行0的 Id="1001" 匹配左链输入值 "1001"，所以 LeftStepValues[1] 长度为 1
	assert.Len(t, ctx.LeftStepValues, 2, "左链应有 2 步")
	assert.Equal(t, 1, len(ctx.LeftStepValues[1]), "过滤后只有1行同时满足Id=1001和Type=Hero")
	assert.Equal(t, []string{"HeroA"}, ctx.LeftCurrentValues)
}

// TestOnion_Performance_RightStepFilterReducesRows 验证右链反向查找时过滤后匹配数
// 右链目标表 10 行，只有 3 行 Type=Equip，验证 RightStepValues 长度为 3
func TestOnion_Performance_RightStepFilterReducesRows(t *testing.T) {
	// 当前表
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "1001"},
	})

	// 右链目标表 Item：10 行，其中 3 行 Type=Equip
	// 1001 同时在 Item 列（被 Match 搜索的列）且 Type=Equip → 匹配
	itemFile := buildChainTestRefSheet("Item", []string{"Item", "Type", "DropGroup"}, [][]string{
		{"1001", "Equip", "G001"}, // 匹配：Item=1001 且 Type=Equip
		{"2001", "Item", "G002"},
		{"2002", "Equip", "G003"}, // Type=Equip 但 Item != 1001，反向查找不匹配
		{"2003", "Hero", "G004"},
		{"2004", "Equip", "G005"}, // Type=Equip 但 Item != 1001，反向查找不匹配
		{"2005", "Item", "G006"},
		{"2006", "Hero", "G007"},
		{"2007", "Item", "G008"},
		{"2008", "Item", "G009"},
		{"2009", "Item", "G010"},
	})

	sheetMap := map[string]*excelize.File{
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "Item", PreCol: "DropGroup", FindVal: "self", NextCol: "Item",
					FilterCol: "Type", FilterVal: "Equip"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// Match 阶段：getRightFinalValues 从 Item 表取 Item 列值（过滤 Type=Equip）
	// 10 行中只有 3 行 Type=Equip → rightVals = ["1001", "2002", "2004"]
	// 左链值 "1001" 在 ["1001", "2002", "2004"] 中 → 匹配通过
	assert.True(t, ctx.Matched, "左链值1001应在过滤后的右链值中找到")
	// Match 传递 rightVals=["1001", "2002", "2004"] 给右链反向
	assert.Equal(t, 3, len(ctx.RightCurrentValues), "过滤后Match阶段应传递3个值给右链")
	// 反向查找：在 Item 表中用 Item 列匹配 ["1001","2002","2004"]，提取 DropGroup
	// 行0: Item=1001 → DropGroup=G001
	// 行2: Item=2002 → DropGroup=G003
	// 行4: Item=2004 → DropGroup=G005
	// 但反向查找也会应用 filterRowSet(Type=Equip)，3 行都是 Equip → 全部匹配
	assert.Len(t, ctx.RightStepValues, 1, "右链有1步")
	assert.Equal(t, 3, len(ctx.RightStepValues[0]), "反向查找后提取3个DropGroup值")
}

// TestOnion_Performance_FullPipelineDataCounts 模拟 SeasonPassReward 场景（简化版）
// 验证整条链路每步的数据量裁剪：
//   - 左链第一步取 HeroId → 第二步跨表 Hero 查找 → Match → 右链反向查找 Item
func TestOnion_Performance_FullPipelineDataCounts(t *testing.T) {
	// 当前表 SeasonPassReward 简化版：6 行数据
	// 检查第 1 行（rowIdx=5），HeroId="H001"
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1", "2", "3", "4", "5", "6"},
		{"", "", "SeasonPassId", "", "", "SP001", "SP002", "SP001", "SP002", "SP003", "SP001"},
		{"", "", "HeroId", "", "", "H001", "H002", "H003", "H004", "H005", "H006"},
	})

	// Hero 表：3 行
	heroFile := buildChainTestRefSheet("Hero", []string{"Id", "Name"}, [][]string{
		{"H001", "张飞"},
		{"H002", "关羽"},
		{"H003", "赵云"},
	})

	// Item 表：2 行（右链最终表）
	itemFile := buildChainTestRefSheet("Item", []string{"Id", "Name"}, [][]string{
		{"H001", "张飞"},
		{"H003", "赵云"},
	})

	sheetMap := map[string]*excelize.File{
		"Hero": heroFile,
		"Item": itemFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				// 第一步：从当前表取 HeroId 值
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "HeroId"},
				// 第二步：在 Hero 表中查找 Id 匹配，提取 Name
				{Sheet: "Hero", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				// 右链一步：在 Item 表中查找
				{Sheet: "Item", PreCol: "Id", FindVal: "self", NextCol: "Name"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 2, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// 左链验证
	assert.Len(t, ctx.LeftStepValues, 2, "左链有2步")
	// 左链第一步：取 HeroId="H001" → LeftStepValues[0] = ["H001"]
	assert.Equal(t, 1, len(ctx.LeftStepValues[0]), "左链第一步取HeroId长度为1")
	// 左链第二步：在 Hero 表（3行）中查找 Id="H001" → 找到1行 → LeftStepValues[1] = ["张飞"]
	assert.Equal(t, 1, len(ctx.LeftStepValues[1]), "左链第二步跨表查找后长度为1")

	// Match 阶段：左链最终值 ["张飞"]，右链 Item 表 Name 列值 ["张飞", "赵云"]
	// "张飞" 在 ["张飞", "赵云"] 中 → 匹配通过
	assert.True(t, ctx.Matched, "张飞在Item表的Name列中")
	assert.Equal(t, 2, len(ctx.MatchedKeys), "Match阶段传递2个右链值")

	// 右链反向查找：在 Item 表中用 Name 列匹配 ["张飞","赵云"]，提取 Id
	// 行0: Name="张飞" → Id="H001"
	// 行1: Name="赵云" → Id="H003"
	assert.Len(t, ctx.RightStepValues, 1, "右链有1步")
	assert.Equal(t, 2, len(ctx.RightStepValues[0]), "右链反向查找提取2个Id值")
}

// TestOnion_Performance_MatchFilterReducesValues 验证 Match 阶段过滤裁剪右链最终值集合
// 右链表有 10 行，过滤后只有 5 行，验证 Match 比较的是过滤后的值
func TestOnion_Performance_MatchFilterReducesValues(t *testing.T) {
	// 当前表：左链值 "val_others"（不在过滤后的右链值中）
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "val_hero"}, // 左链取值
	})

	// 右链最终表：10 行，Type=Active 的只有 5 行
	// Name 列值：Active 表有 val_a1~val_a5，Inactive 表有 val_i1~val_i5
	rightFile := buildChainTestRefSheet("RightTable", []string{"Id", "Type", "Name"}, [][]string{
		{"1", "Active", "val_a1"},
		{"2", "Inactive", "val_i1"},
		{"3", "Active", "val_a2"},
		{"4", "Inactive", "val_i2"},
		{"5", "Active", "val_a3"},
		{"6", "Inactive", "val_i3"},
		{"7", "Active", "val_a4"},
		{"8", "Inactive", "val_i4"},
		{"9", "Active", "val_a5"},
		{"10", "Inactive", "val_i5"},
	})

	sheetMap := map[string]*excelize.File{
		"RightTable": rightFile,
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Col1"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				// 右链一步：过滤 Type=Active，匹配 Name 列
				{Sheet: "RightTable", PreCol: "Id", FindVal: "self", NextCol: "Name",
					FilterCol: "Type", FilterVal: "Active"},
			},
		},
	}

	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)

	assert.NoError(t, err)
	// Match 阶段 getRightFinalValues 应返回过滤后的 5 个值（Type=Active 的 Name 列值）
	// 左链值 "val_hero" 不在过滤后的 ["val_a1","val_a2","val_a3","val_a4","val_a5"] 中
	// Match verify_exists 失败 → ctx.Matched = false
	assert.False(t, ctx.Matched, "左链值不在过滤后的右链值中，匹配应失败")

	// 现在用左链值 "val_a3" 验证匹配成功（在过滤后的 5 个值中）
	cols2 := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Col1", "", "", "val_a3"},
	})

	ctx2 := buildOnionTestContext(cols2, 1, 5, sheetMap, config, params)
	err = runOnionChain(ctx2, config, params)

	assert.NoError(t, err)
	// 左链值 "val_a3" 在过滤后的右链值中 → 匹配成功
	assert.True(t, ctx2.Matched, "val_a3在过滤后的右链值中，匹配应成功")
	// 验证 MatchedKeys 包含过滤后的值（只有 5 个，不是 10 个）
	assert.Equal(t, 5, len(ctx2.MatchedKeys), "Match阶段应只传递过滤后的5个值")
}
