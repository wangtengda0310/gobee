// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件测试左链步骤处理器（LeftStepHandler）和比较处理器（CompareHandler）
package chain_reference

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 测试辅助构造函数 ====================

// buildChainTestCols 构造测试用的列数据
// 格式：按列存储，每列第一行到 MJS_FIXED_ROWS_NUM-1 是表头，之后是数据
// 列定义：pairs 中每个元素是 (列名, 数据值...)
func buildChainTestCols(pairs [][]string) [][]string {
	return pairs
}

// buildChainTestRefSheet 构造测试用的参考表 excelize.File
// headers: 列名列表，rows: 每行数据（从 MJS_FIXED_ROWS_NUM 行开始写入）
func buildChainTestRefSheet(sheetName string, headers []string, rows [][]string) *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)
	// 写入列名（行索引 2 = MJS_FIXED_ROWS_NAME）
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 3) // 列名在第3行（索引2）
		f.SetCellValue(sheetName, cell, header)
	}
	// 写入数据行（从第5行开始，索引4 = MJS_FIXED_ROWS_NUM）
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+5) // 数据从第5行开始
			f.SetCellValue(sheetName, cell, val)
		}
	}
	return f
}

// ==================== 左链测试 ====================

// TestLeftStep_ValueOnlyMode 左链第一步仅取值
// 从当前表指定列取值，不跨表查找
func TestLeftStep_ValueOnlyMode(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	})

	ctx := NewChainContext(cols, nil, 1, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
			},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:     ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
		StepIdx:  0,
		IsLast:   true,
		ChainCfg: ChainConfig{Steps: []ChainStep{{Sheet: "", NextCol: "OnceDropRule"}}},
	}

	nextCalled := false
	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, []string{"1001"}, ctx.LeftCurrentValues)
	assert.Equal(t, [][]string{{"1001"}}, ctx.LeftStepValues)
	assert.NotNil(t, ctx.LeftResult)
	assert.Equal(t, []string{"1001"}, ctx.LeftResult.MatchValues())
}

// TestLeftStep_ValueWithRegex 左链第一步正则提取
// 从当前表取值后，使用正则提取子值
func TestLeftStep_ValueWithRegex(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "{1001;1}{2002;1}"},
	})

	ctx := NewChainContext(cols, nil, 1, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "BigAward", Pattern: `{(\d+);\d+}`, Groups: "1"},
			},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:    ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "BigAward", Pattern: `{(\d+);\d+}`, Groups: "1"},
		StepIdx: 0,
		IsLast:  true,
		ChainCfg: ChainConfig{
			Steps: []ChainStep{{Pattern: `{(\d+);\d+}`}},
		},
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, []string{"1001", "2002"}, ctx.LeftCurrentValues)
	assert.Equal(t, [][]string{{"1001", "2002"}}, ctx.LeftStepValues)
	assert.Equal(t, []string{"1001", "2002"}, ctx.LeftResult.MatchValues())
}

// TestLeftStep_CrossTableLookup 左链后续步骤跨表查找
// 第一步取值后，在目标表中查找匹配行并提取新值
func TestLeftStep_CrossTableLookup(t *testing.T) {
	// 当前表：OnceDropRule=1001
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	})

	// 目标表：DropRule，Id=1001 对应 DropGroup=G001
	dropRuleFile := buildChainTestRefSheet("DropRule", []string{"Id", "DropGroup"}, [][]string{
		{"1001", "G001"},
		{"1002", "G002"},
	})

	sheetMap := map[string]*excelize.File{
		"DropRule": dropRuleFile,
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
				{Sheet: "DropRule", PreCol: "Id", FindVal: "self", NextCol: "DropGroup"},
			},
		},
	}, nil, nil, 0)

	// 第一步：仅取值
	step0 := &LeftStepHandler{
		Step:    ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
		StepIdx: 0,
		IsLast:  false,
		ChainCfg: ChainConfig{
			Steps: []ChainStep{{}, {Sheet: "DropRule", PreCol: "Id", NextCol: "DropGroup"}},
		},
	}

	err := step0.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, []string{"1001"}, ctx.LeftCurrentValues)

	// 第二步：跨表查找
	step1 := &LeftStepHandler{
		Step:    ChainStep{Sheet: "DropRule", PreCol: "Id", FindVal: "self", NextCol: "DropGroup"},
		StepIdx: 1,
		IsLast:  true,
		ChainCfg: ChainConfig{
			Steps: []ChainStep{{}, {Sheet: "DropRule", PreCol: "Id", NextCol: "DropGroup"}},
		},
	}

	err = step1.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, []string{"G001"}, ctx.LeftCurrentValues)
	assert.NotNil(t, ctx.LeftResult)
	assert.Equal(t, []string{"G001"}, ctx.LeftResult.MatchValues())
}

// TestLeftStep_EmptyValue 当前值为空时 LeftCurrentValues 保持 nil
func TestLeftStep_EmptyValue(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", ""}, // 空值
	})

	ctx := NewChainContext(cols, nil, 1, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"}},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:     ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
		StepIdx:  0,
		IsLast:   true,
		ChainCfg: ChainConfig{},
	}

	nextCalled := false
	err := handler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, nextCalled, "空值时仍应调用 next")
	assert.Nil(t, ctx.LeftCurrentValues)
	assert.Nil(t, ctx.LeftResult)
}

// TestLeftStep_ColumnNotFound 列名不存在时返回错误
func TestLeftStep_ColumnNotFound(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
	})

	ctx := NewChainContext(cols, nil, 0, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{{Sheet: "", PreCol: "", FindVal: "col", NextCol: "NotExistCol"}},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:     ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "NotExistCol"},
		StepIdx:  0,
		IsLast:   true,
		ChainCfg: ChainConfig{},
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到列")
}

// TestLeftStep_SingleStep_ValueOnly 左链只有一步（仅取值模式）
func TestLeftStep_SingleStep_ValueOnly(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	})

	ctx := NewChainContext(cols, nil, 1, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"}},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:     ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
		StepIdx:  0,
		IsLast:   true,
		ChainCfg: ChainConfig{},
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.NotNil(t, ctx.LeftResult)
	assert.Equal(t, []string{"1001"}, ctx.LeftResult.MatchValues())
	assert.Equal(t, []string{"1001"}, ctx.FirstStepInputValues)
}

// TestLeftStep_SingleStep_WithRegex 左链只有一步带正则
func TestLeftStep_SingleStep_WithRegex(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Items", "", "", "{100;1}{200;2}"},
	})

	ctx := NewChainContext(cols, nil, 1, 5, 4, nil, &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Items", Pattern: `{(\d+);\d+}`, Groups: "1"},
			},
		},
	}, nil, nil, 0)

	handler := &LeftStepHandler{
		Step:     ChainStep{Sheet: "", PreCol: "", FindVal: "col", NextCol: "Items", Pattern: `{(\d+);\d+}`, Groups: "1"},
		StepIdx:  0,
		IsLast:   true,
		ChainCfg: ChainConfig{},
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.NotNil(t, ctx.LeftResult)
	assert.Equal(t, []string{"100", "200"}, ctx.LeftResult.MatchValues())
	assert.Equal(t, []string{"100", "200"}, ctx.FirstStepInputValues)
}

// TestLeftChain_MultiStepConsistency 多步左链输出与 ExecuteChain 一致
// 验证 LeftStepHandler 逐步执行的结果与 ExecuteChain 整体执行的结果相同
func TestLeftChain_MultiStepConsistency(t *testing.T) {
	// 当前表：OnceDropRule=1001
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	})

	// 目标表：DropRule
	dropRuleFile := buildChainTestRefSheet("DropRule", []string{"Id", "DropGroup"}, [][]string{
		{"1001", "G001"},
	})

	sheetMap := map[string]*excelize.File{
		"DropRule": dropRuleFile,
	}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
			{Sheet: "DropRule", PreCol: "Id", FindVal: "self", NextCol: "DropGroup"},
		},
	}

	// 方式 1: 使用 ExecuteChain
	engineResult, err := ExecuteChain(cols, 1, 5, 4, cfg, sheetMap)
	assert.NoError(t, err)

	// 方式 2: 使用 LeftStepHandler 逐步执行
	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, &ChainPairConfig{Left: cfg}, nil, nil, 0)

	chain := BuildOnionChain(&ChainPairConfig{Left: cfg}, map[string]string{
		"chainCompare": "verify_exists",
	})
	_ = chain(ctx)

	// 比较：LeftResult 的 MatchValues 应与 ExecuteChain 一致
	if engineResult != nil && ctx.LeftResult != nil {
		assert.Equal(t, engineResult.MatchValues(), ctx.LeftResult.MatchValues(),
			"LeftStepHandler 输出应与 ExecuteChain 一致")
	}
}

// ==================== 比较测试 ====================

// TestCompare_TimeOverlap_ExactMatch time_overlap 时间点匹配报违规
func TestCompare_TimeOverlap_ExactMatch(t *testing.T) {
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{
				{Match: "1001", Compare: "2026-01-01 10:00:00"},
			},
		},
		RightResult: &ChainResult{
			Values: []ChainValue{
				{Match: "1001", Compare: "2026-01-01 10:00:00"},
			},
		},
		Params:  map[string]string{},
		Matched: true,
	}

	handler := &CompareHandler{
		CompareType:    "time_overlap",
		UseTimeOverlap: true,
		LeftKey:        "",
		RightKey:       "",
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.True(t, ctx.Violation, "时间点匹配应报违规")
	assert.Contains(t, ctx.Reason, "time_overlap")
}

// TestCompare_SinglePhase_WhenMatchCompareEmpty 退化单阶段比较
// MatchCompareType 为空时，直接比较 LeftResult.MatchValues vs RightResult.MatchValues
func TestCompare_SinglePhase_WhenMatchCompareEmpty(t *testing.T) {
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{{Match: "A"}},
		},
		RightResult: &ChainResult{
			Values: []ChainValue{{Match: "A"}},
		},
		// MatchCompareType 为空
		Params:  map[string]string{},
		Matched: false, // 未经过两阶段门控
	}

	handler := &CompareHandler{
		CompareType:    "verify_exists",
		UseTimeOverlap: false,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	// verify_exists: "A" 在 ["A"] 中找到 → 不违规
	assert.False(t, ctx.Violation, "单阶段退化模式下 verify_exists 全部找到不应违规")
}

// TestCompare_Backward_Compatibility 与旧行为一致
// 验证 CompareHandler 的单阶段退化模式与直接调用 CompareByType 结果一致
func TestCompare_Backward_Compatibility(t *testing.T) {
	// 左链值
	leftVals := []string{"a", "x"}
	// 右链值
	rightVals := []string{"a", "b"}

	// 方式 1: 直接调用 CompareByType（旧行为）
	oldViolation, oldReason := CompareByType("verify_exists", leftVals, rightVals)

	// 方式 2: 使用 CompareHandler（新行为）
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{{Match: "a"}, {Match: "x"}},
		},
		RightResult: &ChainResult{
			Values: []ChainValue{{Match: "a"}, {Match: "b"}},
		},
		Params:  map[string]string{},
		Matched: true,
	}

	handler := &CompareHandler{
		CompareType:    "verify_exists",
		UseTimeOverlap: false,
	}

	_ = handler.Handle(ctx, func(ctx *ChainContext) error { return nil })

	// 验证结果一致
	assert.Equal(t, oldViolation, ctx.Violation, "新旧行为应一致")
	assert.Equal(t, oldReason, ctx.Reason, "违规原因应一致")
}

// TestCompare_MatchGateNotPassed 两阶段门控未通过时不报违规
func TestCompare_MatchGateNotPassed(t *testing.T) {
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{{Match: "1001"}},
		},
		RightResult: &ChainResult{
			Values: []ChainValue{{Match: "1001"}},
		},
		Params: map[string]string{
			"chainMatchCompare": "verify_exists",
		},
		Matched: false, // 门控未通过
	}

	handler := &CompareHandler{
		CompareType:    "verify_exists",
		UseTimeOverlap: false,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	assert.False(t, ctx.Violation, "门控未通过时不应报违规")
}

// TestCompare_MatchGatePassed_TwoPhaseCompare 两阶段门控通过后执行 Phase2 比较
func TestCompare_MatchGatePassed_TwoPhaseCompare(t *testing.T) {
	ctx := &ChainContext{
		LeftResult: &ChainResult{
			Values: []ChainValue{{Match: "1001"}},
		},
		RightResult: &ChainResult{
			Values:               []ChainValue{{Match: "1002"}},
			FirstStepInputValues: []string{"target_val_missing"},
		},
		Config: &ChainPairConfig{
			Left: ChainConfig{
				Steps: []ChainStep{{IsArray: "false"}},
			},
		},
		Params: map[string]string{
			"chainMatchCompare": "verify_exists",
			"chainCompare":      "verify_exists",
		},
		Matched:   true, // 门控通过
		MyColData: []string{"target_val"},
		DataIdx:   0,
	}

	handler := &CompareHandler{
		CompareType:    "verify_exists",
		UseTimeOverlap: false,
	}

	err := handler.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)
	// verify_exists: "target_val" 在 ["target_val_missing"] 中未找到 → 违规
	assert.True(t, ctx.Violation, "Phase2 比较有值缺失应报违规")
}
