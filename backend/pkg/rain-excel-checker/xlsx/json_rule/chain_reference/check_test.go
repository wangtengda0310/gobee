package chain_reference

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== ParseChainPairConfig 测试 ====================

// TestParseChainPairConfig_NormalPath 解析完整的两链 JSON
func TestParseChainPairConfig_NormalPath(t *testing.T) {
	jsonStr := `{
		"left": {
			"steps": [
				{"sheet": "SeasonPassReward", "preCol": "HighReward", "findVal": "self", "nextCol": "SeasonPassId"},
				{"sheet": "SeasonPass", "preCol": "Id", "findVal": "self", "nextCol": "StartTime"}
			],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [
				{"sheet": "Item", "preCol": "Id", "findVal": "col", "pattern": "{(\\d+);\\d+}", "groups": "1", "filterCol": "Type", "filterVal": "Hero", "nextCol": "ItemParam"}
			],
			"compareCol": "ItemParam"
		}
	}`

	config, err := ParseChainPairConfig(jsonStr)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// 验证 left 链
	assert.Len(t, config.Left.Steps, 2)
	assert.Equal(t, "SeasonPassReward", config.Left.Steps[0].Sheet)
	assert.Equal(t, "HighReward", config.Left.Steps[0].PreCol)
	assert.Equal(t, "StartTime", config.Left.CompareCol)

	// 验证 right 链
	assert.Len(t, config.Right.Steps, 1)
	assert.Equal(t, "Item", config.Right.Steps[0].Sheet)
	assert.Equal(t, `{(\d+);\d+}`, config.Right.Steps[0].Pattern)
	assert.Equal(t, "ItemParam", config.Right.CompareCol)
}

// TestParseChainPairConfig_WithCompareCol 测试带 CompareCol 的配置
func TestParseChainPairConfig_WithCompareCol(t *testing.T) {
	jsonStr := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "self", "nextCol": "Name"}],
			"compareCol": "Type"
		},
		"right": {
			"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]
		}
	}`

	config, err := ParseChainPairConfig(jsonStr)
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "Type", config.Left.CompareCol)
	assert.Empty(t, config.Right.CompareCol)
}

// TestParseChainPairConfig_Empty 空字符串返回错误
func TestParseChainPairConfig_Empty(t *testing.T) {
	config, err := ParseChainPairConfig("")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "为空")
}

// TestParseChainPairConfig_InvalidJSON 格式错误返回错误
func TestParseChainPairConfig_InvalidJSON(t *testing.T) {
	config, err := ParseChainPairConfig("not json")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "解析失败")
}

// TestParseChainPairConfig_NoSteps 两条链都没有步骤返回错误
func TestParseChainPairConfig_NoSteps(t *testing.T) {
	jsonStr := `{"left": {"steps": []}, "right": {"steps": []}}`
	config, err := ParseChainPairConfig(jsonStr)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "没有步骤")
}

// TestParseChainPairConfig_SingleStep 只有一条链有一个步骤
func TestParseChainPairConfig_SingleStep(t *testing.T) {
	jsonStr := `{
		"left": {"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "self", "nextCol": "Name"}]},
		"right": {"steps": []}
	}`

	config, err := ParseChainPairConfig(jsonStr)
	assert.NoError(t, err)
	assert.Len(t, config.Left.Steps, 1)
	assert.Len(t, config.Right.Steps, 0)
}

// ==================== ExtractByRegexFromString 测试 ====================

// TestExtractByRegexFromString_Normal 从多值格式中提取ID列表
func TestExtractByRegexFromString_Normal(t *testing.T) {
	input := "{1001;1}{1002;10}{1003;5}"
	pattern := `{(\d+);\d+}`
	groups := "1"

	result, err := ExtractByRegexFromString(input, pattern, groups)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1001", "1002", "1003"}, result)
}

// TestExtractByRegexFromString_EmptyInput 空字符串返回 nil
func TestExtractByRegexFromString_EmptyInput(t *testing.T) {
	result, err := ExtractByRegexFromString("", `{(\d+)}`, "1")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestExtractByRegexFromString_EmptyPattern 空正则返回 nil
func TestExtractByRegexFromString_EmptyPattern(t *testing.T) {
	result, err := ExtractByRegexFromString("some text", "", "1")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestExtractByRegexFromString_NoMatch 没有匹配的返回 nil
func TestExtractByRegexFromString_NoMatch(t *testing.T) {
	result, err := ExtractByRegexFromString("hello world", `(\d+)`, "1")
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestExtractByRegexFromString_InvalidPattern 编译错误返回错误
func TestExtractByRegexFromString_InvalidPattern(t *testing.T) {
	result, err := ExtractByRegexFromString("input", "[invalid", "1")
	assert.Error(t, err)
	assert.Empty(t, result)
}

// ==================== FilterRowsByCondition 测试 ====================

// TestFilterRowsByCondition_WithFilter 返回满足条件的行
func TestFilterRowsByCondition_WithFilter(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},            // Id 列
		{"", "", "Type", "", "", "Hero", "Item", "Hero"}, // Type 列
	}
	startRowIdx := 5

	result := FilterRowsByCondition(cols, "Type", "Hero", startRowIdx, "")
	assert.Equal(t, []int{5, 7}, result)
}

// TestFilterRowsByCondition_NoFilter 无过滤条件返回 nil
func TestFilterRowsByCondition_NoFilter(t *testing.T) {
	cols := [][]string{{"", "", "Id", "", "", "1"}}
	result := FilterRowsByCondition(cols, "", "", 5, "")
	assert.Empty(t, result)
}

// TestFilterRowsByCondition_ColNotFound 列不存在返回 nil
func TestFilterRowsByCondition_ColNotFound(t *testing.T) {
	cols := [][]string{{"", "", "Id", "", "", "1"}}
	result := FilterRowsByCondition(cols, "NotExist", "val", 5, "")
	assert.Empty(t, result)
}

// ==================== ExecuteChain 测试 ====================

// TestExecuteChain_EmptySteps 空步骤返回空结果
func TestExecuteChain_EmptySteps(t *testing.T) {
	result, err := ExecuteChain(nil, 0, 0, 0, ChainConfig{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.Values)
}

// TestExecuteChain_NoCompareCol 无 CompareCol 时 Compare 字段为空
func TestExecuteChain_NoCompareCol(t *testing.T) {
	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "TestSheet",
				PreCol:  "Id",
				FindVal: "col",
				NextCol: "Name",
			},
		},
		CompareCol: "",
	}

	assert.Empty(t, cfg.CompareCol)
}

// TestExecuteChain_WithCompareCol 有 CompareCol 时应正确设置
func TestExecuteChain_WithCompareCol(t *testing.T) {
	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "TestSheet",
				PreCol:  "Id",
				FindVal: "col",
				NextCol: "Name",
			},
		},
		CompareCol: "Type",
	}

	assert.Equal(t, "Type", cfg.CompareCol)
}

// ==================== ChainResult 方法测试 ====================

// TestChainResult_MatchValues 正确提取所有 Match 值
func TestChainResult_MatchValues(t *testing.T) {
	result := &ChainResult{
		Values: []ChainValue{
			{Match: "1001", Compare: "Hero"},
			{Match: "1002", Compare: "Item"},
			{Match: "1003", Compare: "Hero"},
		},
	}

	matches := result.MatchValues()
	assert.Equal(t, []string{"1001", "1002", "1003"}, matches)
}

// TestChainResult_MatchValues_NilResult nil 结果返回 nil
func TestChainResult_MatchValues_NilResult(t *testing.T) {
	var result *ChainResult
	assert.Nil(t, result.MatchValues())
}

// TestChainResult_MatchValues_EmptyValues 空值列表返回空切片
func TestChainResult_MatchValues_EmptyValues(t *testing.T) {
	result := &ChainResult{
		Values: []ChainValue{},
	}

	matches := result.MatchValues()
	assert.Equal(t, []string{}, matches)
}

// TestChainResult_GetCompareValues_ExactMatch 精确匹配返回指定键的比较值
func TestChainResult_GetCompareValues_ExactMatch(t *testing.T) {
	result := &ChainResult{
		Values: []ChainValue{
			{Match: "1001", Compare: "Hero"},
			{Match: "1002", Compare: "Item"},
			{Match: "1003", Compare: "Hero"},
		},
	}

	compares := result.GetCompareValues("1001")
	assert.Equal(t, []string{"Hero"}, compares)

	compares = result.GetCompareValues("1002")
	assert.Equal(t, []string{"Item"}, compares)
}

// TestChainResult_GetCompareValues_NotMatch 键不存在返回 nil
func TestChainResult_GetCompareValues_NotMatch(t *testing.T) {
	result := &ChainResult{
		Values: []ChainValue{
			{Match: "1001", Compare: "Hero"},
			{Match: "1002", Compare: "Item"},
		},
	}

	compares := result.GetCompareValues("9999")
	assert.Nil(t, compares)
}

// TestChainResult_GetCompareValues_EmptyCompare 比较值为空时返回 nil
func TestChainResult_GetCompareValues_EmptyCompare(t *testing.T) {
	result := &ChainResult{
		Values: []ChainValue{
			{Match: "1001", Compare: ""},
		},
	}

	compares := result.GetCompareValues("1001")
	assert.Nil(t, compares)
}

// TestChainResult_GetCompareValues_NilResult nil 结果返回 nil
func TestChainResult_GetCompareValues_NilResult(t *testing.T) {
	var result *ChainResult
	assert.Nil(t, result.GetCompareValues("1001"))
}

// TestChainValue 结构体测试
func TestChainValue(t *testing.T) {
	cv := ChainValue{
		Match:   "1001",
		Compare: "Hero",
	}

	assert.Equal(t, "1001", cv.Match)
	assert.Equal(t, "Hero", cv.Compare)
}

// ==================== 左链第一步"仅取值"模式测试 ====================

// TestExecuteChain_LeftFirstStep_ValueOnlyMode 验证左链第一步 sheet="" 时直接从 nextCol 列取值
// 不执行跨表查找，直接将取到的值传递给下一步
func TestExecuteChain_LeftFirstStep_ValueOnlyMode(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "",
				PreCol:  "",
				FindVal: "col",
				NextCol: "OnceDropRule",
			},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, map[string]*excelize.File{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, [][]string{{"1001"}}, result.StepValues)
	assert.Equal(t, []string{"1001"}, result.MatchValues())
}

// TestExecuteChain_LeftFirstStep_ValueOnlyWithRegex 验证仅取值模式下正则提取
func TestExecuteChain_LeftFirstStep_ValueOnlyWithRegex(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "{1001;1}{2002;1}"},
	}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "",
				PreCol:  "",
				FindVal: "col",
				NextCol: "BigAward",
				Pattern: `{(\d+);\d+}`,
				Groups:  "1",
			},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, map[string]*excelize.File{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, [][]string{{"1001", "2002"}}, result.StepValues)
	assert.Equal(t, []string{"1001", "2002"}, result.MatchValues())
}

// TestExecuteChain_FirstStep_ColWithSheet 回归测试：正常跨表跳转仍然正常工作
func TestExecuteChain_FirstStep_ColWithSheet(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	}

	dropRuleFile := excelize.NewFile()
	sheet := "DropRule"
	dropRuleFile.SetSheetName("Sheet1", sheet)
	dropRuleFile.SetCellValue(sheet, "A3", "Id")
	dropRuleFile.SetCellValue(sheet, "B3", "DropGroup")
	dropRuleFile.SetCellValue(sheet, "A5", "1001")
	dropRuleFile.SetCellValue(sheet, "B5", "G001")

	sheetMap := map[string]*excelize.File{
		"DropRule": dropRuleFile,
	}

	// 跨表跳转：colIdx=1 对应 OnceDropRule 列，在 DropRule 中查找
	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "DropRule",
				PreCol:  "Id",
				FindVal: "col",
				NextCol: "DropGroup",
			},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, [][]string{{"G001"}}, result.StepValues)
	assert.Equal(t, []string{"G001"}, result.MatchValues())
}

// ==================== FirstStepInputValues 验证测试 ====================

// TestExecuteChain_FirstStepInputValues_CrossTable 验证跨表查找时 FirstStepInputValues 收集正确
// 右链第一步：colIdx=1（BigAward 列）取值后，在 DropItem.Item 中查找
func TestExecuteChain_FirstStepInputValues_CrossTable(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "{1001;1}"},
	}

	// DropItem: Item={1001;1} (ItemCfg格式)
	dropItemFile := excelize.NewFile()
	dropItemFile.SetSheetName("Sheet1", "DropItem")
	dropItemFile.SetCellValue("DropItem", "A3", "Item")
	dropItemFile.SetCellValue("DropItem", "B3", "DropGroup")
	dropItemFile.SetCellValue("DropItem", "A5", "{1001;1}")
	dropItemFile.SetCellValue("DropItem", "B5", "G001")

	sheetMap := map[string]*excelize.File{
		"DropItem": dropItemFile,
	}

	// 跨表跳转：colIdx=1 对应 BigAward 列，从当前列取值后去 DropItem 查找
	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "DropItem",
				PreCol:  "Item",
				FindVal: "col",
				NextCol: "DropGroup",
			},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// StepValues[0] 应该是 NextCol 的值
	assert.Equal(t, [][]string{{"G001"}}, result.StepValues)
	// FirstStepInputValues 应该是 searchValues（即 inputValues，因为没有正则）
	assert.Equal(t, []string{"{1001;1}"}, result.GetFirstStepInputValues())
}

// TestExecuteChain_FirstStepInputValues_WithRegex 验证有正则时 FirstStepInputValues 是正则提取后的值
func TestExecuteChain_FirstStepInputValues_WithRegex(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "{1001;1}{2002;1}"},
	}

	// DropItem: Item=1001 和 Item=2002
	dropItemFile := excelize.NewFile()
	dropItemFile.SetSheetName("Sheet1", "DropItem")
	dropItemFile.SetCellValue("DropItem", "A3", "Item")
	dropItemFile.SetCellValue("DropItem", "B3", "DropGroup")
	dropItemFile.SetCellValue("DropItem", "A5", "1001")
	dropItemFile.SetCellValue("DropItem", "B5", "G001")
	dropItemFile.SetCellValue("DropItem", "A6", "2002")
	dropItemFile.SetCellValue("DropItem", "B6", "G001")

	sheetMap := map[string]*excelize.File{
		"DropItem": dropItemFile,
	}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:   "DropItem",
				PreCol:  "Item",
				FindVal: "col",
				Pattern: `{(\d+);\d+}`,
				Groups:  "1",
				NextCol: "DropGroup",
			},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// StepValues[0] = NextCol 的值（两个 G001）
	assert.Equal(t, [][]string{{"G001", "G001"}}, result.StepValues)
	// FirstStepInputValues = 正则提取后的查找值
	assert.Equal(t, []string{"1001", "2002"}, result.GetFirstStepInputValues())
}

// TestExecuteChain_FirstStepInputValues_ValueOnly 验证仅取值模式下 FirstStepInputValues
func TestExecuteChain_FirstStepInputValues_ValueOnly(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "OnceDropRule", "", "", "1001"},
	}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{Sheet: "", PreCol: "", FindVal: "col", NextCol: "OnceDropRule"},
		},
	}

	result, err := ExecuteChain(cols, 1, 5, 4, cfg, map[string]*excelize.File{})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, []string{"1001"}, result.GetFirstStepInputValues())
}

// ==================== ExtractChainStepSheets 测试 ====================

// TestExtractChainStepSheets_Normal 提取左右链中所有引用的表名
func TestExtractChainStepSheets_Normal(t *testing.T) {
	params := map[string]string{
		"chainSteps": `{
			"left": {"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]},
			"right": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]}
		}`,
	}

	sheets := ExtractChainStepSheets(params)
	assert.Contains(t, sheets, "Hero")
	assert.Contains(t, sheets, "Item")
	assert.Len(t, sheets, 2)
}

// TestExtractChainStepSheets_Dedup 去重：左右链引用相同表名
func TestExtractChainStepSheets_Dedup(t *testing.T) {
	params := map[string]string{
		"chainSteps": `{
			"left": {"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]},
			"right": {"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]}
		}`,
	}

	sheets := ExtractChainStepSheets(params)
	assert.Len(t, sheets, 1)
	assert.Contains(t, sheets, "Hero")
}

// TestExtractChainStepSheets_EmptyParams chainSteps 为空返回 nil
func TestExtractChainStepSheets_EmptyParams(t *testing.T) {
	sheets := ExtractChainStepSheets(map[string]string{})
	assert.Nil(t, sheets)

	sheets = ExtractChainStepSheets(map[string]string{"chainSteps": ""})
	assert.Nil(t, sheets)
}

// TestExtractChainStepSheets_SkipEmptySheet 跳过 sheet 为空的步骤（左链第一步仅取值模式）
func TestExtractChainStepSheets_SkipEmptySheet(t *testing.T) {
	params := map[string]string{
		"chainSteps": `{
			"left": {"steps": [{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "BigAward"}]},
			"right": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]}
		}`,
	}

	sheets := ExtractChainStepSheets(params)
	assert.Len(t, sheets, 1)
	assert.Contains(t, sheets, "Item")
}

// ==================== FilterRowsByCondition 多值匹配测试 ====================

// TestFilterRowsByCondition_SingleValue 验证原有单值匹配行为不变（filterIsArray=""）
func TestFilterRowsByCondition_SingleValue(t *testing.T) {
	// 构造 3 列 4 行数据的表：Id, Type, Name
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3", "4"},
		{"", "", "Type", "", "", "Hero", "Item", "Hero", "Equip"},
		{"", "", "Name", "", "", "张飞", "长剑", "关羽", "护甲"},
	}
	startRowIdx := 5

	// 单值精确匹配：只匹配 Type=Hero 的行
	result := FilterRowsByCondition(cols, "Type", "Hero", startRowIdx, "")
	assert.Equal(t, []int{5, 7}, result)

	// 单值精确匹配：匹配 Type=Item 的行
	result = FilterRowsByCondition(cols, "Type", "Item", startRowIdx, "")
	assert.Equal(t, []int{6}, result)

	// 单值精确匹配：不存在的值返回空
	result = FilterRowsByCondition(cols, "Type", "NotExist", startRowIdx, "")
	assert.Empty(t, result)
}

// TestFilterRowsByCondition_MultiValue 验证多值匹配（filterIsArray="true"，filterVal="A,B,C"）
func TestFilterRowsByCondition_MultiValue(t *testing.T) {
	// 构造 3 列 5 行数据的表：Id, Type, Name
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3", "4", "5"},
		{"", "", "Type", "", "", "Hero", "Item", "Hero", "Equip", "Skin"},
		{"", "", "Name", "", "", "张飞", "长剑", "关羽", "护甲", "皮肤A"},
	}
	startRowIdx := 5

	// 多值匹配：Type 匹配 Hero 或 Item 或 Equip 的行
	result := FilterRowsByCondition(cols, "Type", "Hero,Item,Equip", startRowIdx, "true")
	assert.Equal(t, []int{5, 6, 7, 8}, result)

	// 多值匹配：只匹配部分值
	result = FilterRowsByCondition(cols, "Type", "Item,Skin", startRowIdx, "true")
	assert.Equal(t, []int{6, 9}, result)

	// 多值匹配：全部不存在的值返回空
	result = FilterRowsByCondition(cols, "Type", "NotExist1,NotExist2", startRowIdx, "true")
	assert.Empty(t, result)
}

// TestFilterRowsByCondition_MultiValue_WithBraces 验证含花括号的多值不被错误拆分
// filterVal="{A,B},C" 应拆分为 "{A,B}" 和 "C"，而不是 "A" "B" "C"
func TestFilterRowsByCondition_MultiValue_WithBraces(t *testing.T) {
	// 构造 3 列 4 行数据的表：Id, Value, Name
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3", "4"},
		{"", "", "Value", "", "", "{A,B}", "C", "A", "D"},
		{"", "", "Name", "", "", "复合值", "单项C", "单项A", "不在列表"},
	}
	startRowIdx := 5

	// filterVal="{A,B},C" 应拆分为 "{A,B}" 和 "C"
	// 匹配 Value="{A,B}" 的行（行5）和 Value="C" 的行（行6）
	result := FilterRowsByCondition(cols, "Value", "{A,B},C", startRowIdx, "true")
	assert.Equal(t, []int{5, 6}, result)

	// 验证单独的 "A" 不在匹配结果中（因为 "{A,B}" 整体是一个值，不会被拆分为 "A"）
	// 行7 Value="A" 不应被匹配
	result = FilterRowsByCondition(cols, "Value", "{A,B}", startRowIdx, "true")
	assert.Equal(t, []int{5}, result)
}

// ==================== FilterRowsByConditionEx 测试 ====================

// TestFilterRowsByConditionEx_EmptyMode filterMode="" 等价于单值匹配
func TestFilterRowsByConditionEx_EmptyMode(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "Type", "", "", "Hero", "Item", "Hero"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Type",
		FilterVal:     "Hero",
		StartRowIdx:   startRowIdx,
		FilterMode:    "",
	})
	assert.Equal(t, []int{5, 7}, result)
}

// TestFilterRowsByConditionEx_MultiMode filterMode="multi" 强制多值匹配
func TestFilterRowsByConditionEx_MultiMode(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3", "4"},
		{"", "", "Type", "", "", "Hero", "Item", "Equip", "Skin"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Type",
		FilterVal:     "Hero,Item",
		StartRowIdx:   startRowIdx,
		FilterMode:    "multi",
	})
	assert.Equal(t, []int{5, 6}, result)
}

// TestFilterRowsByConditionEx_WithinDays_Basic 过去/未来3天/远未来，N=7，只返回未来3天
func TestFilterRowsByConditionEx_WithinDays_Basic(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "Date", "", "", "2026-05-18 00:00:00", "2026-05-22 00:00:00", "2026-06-01 00:00:00"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Date",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "7",
		Now:           now,
	})
	// 行5=过去(5/18)排除, 行6=未来3天(5/22)保留, 行7=远未来(6/1)超出7天排除
	assert.Equal(t, []int{6}, result)
}

// TestFilterRowsByConditionEx_WithinDays_NoMatch 全部过期
func TestFilterRowsByConditionEx_WithinDays_NoMatch(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2"},
		{"", "", "Date", "", "", "2026-05-10 00:00:00", "2026-05-15 00:00:00"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Date",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "7",
		Now:           now,
	})
	assert.Empty(t, result)
}

// TestFilterRowsByConditionEx_WithinDays_AllMatch 全部在未来N天内
func TestFilterRowsByConditionEx_WithinDays_AllMatch(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2"},
		{"", "", "Date", "", "", "2026-05-20 00:00:00", "2026-05-25 00:00:00"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Date",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "7",
		Now:           now,
	})
	assert.Equal(t, []int{5, 6}, result)
}

// TestFilterRowsByConditionEx_WithinDays_InvalidDate 非法日期被排除
func TestFilterRowsByConditionEx_WithinDays_InvalidDate(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "Date", "", "", "not-a-date", "", "2026-05-22 00:00:00"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Date",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "7",
		Now:           now,
	})
	assert.Equal(t, []int{7}, result)
}

// TestFilterRowsByConditionEx_WithinDays_InvalidDays filterDays 非数字返回 nil
func TestFilterRowsByConditionEx_WithinDays_InvalidDays(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Date", "", "", "2026-05-22 00:00:00"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "Date",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "abc",
	})
	assert.Empty(t, result)
}

// TestFilterRowsByConditionEx_WithinDays_EmptyCol filterColName 为空返回 nil
func TestFilterRowsByConditionEx_WithinDays_EmptyCol(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
	}
	startRowIdx := 5

	result, _ := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "",
		StartRowIdx:   startRowIdx,
		FilterMode:    "withinDays",
		FilterDays:    "7",
	})
	assert.Empty(t, result)
}

// TestFilterRowsByConditionEx_BackwardCompat 通过原 FilterRowsByCondition 调用，行为不变
func TestFilterRowsByConditionEx_BackwardCompat(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "Type", "", "", "Hero", "Item", "Hero"},
	}
	startRowIdx := 5

	// 单值匹配
	result := FilterRowsByCondition(cols, "Type", "Hero", startRowIdx, "")
	assert.Equal(t, []int{5, 7}, result)

	// 多值匹配
	result = FilterRowsByCondition(cols, "Type", "Hero,Item", startRowIdx, "true")
	assert.Equal(t, []int{5, 6, 7}, result)

	// 无过滤条件
	result = FilterRowsByCondition(cols, "", "", startRowIdx, "")
	assert.Empty(t, result)
}

// ==================== withinDays 调用点集成测试（回归保护）====================
// 这些测试覆盖 BUG: 5 个生产代码调用点漏传 FilterMode/FilterDays，
// 导致 ChainStep 配置的 withinDays 过滤不生效，整张表数据全部参与匹配。
// 测试构造"未来5天"和"未来100天"两行数据，windowDays=20 时只有前者匹配；
// 修复前测试会因两行都匹配而失败。
// 时间采用 time.Now() 相对值（15 天宽窗口余量），避免日期漂移。
//
// 调用点覆盖矩阵：
//   ✅ engine.go:205                          ← TestExecuteChain_WithinDaysFilter_Regression
//   ✅ onion_left_step.go:174                 ← TestOnion_WithinDaysFilter_LeftStep
//   ✅ onion_right_step.go:122                ← TestOnion_WithinDaysFilter_RightStep
//   ✅ onion_match.go:123                     ← TestOnion_WithinDaysFilter_MatchStep
//   ⚠️ excel_check_chain_reference.go:116    ← 未覆盖（旧路径，需走 Check 入口构造完整 params/sheetMap）
//
// TODO: 若旧路径 (chainMatchCompare 为空) 未来仍可能被触发，补充入口级 smoke test。

// futureDate 返回从当前时间起 days 天后的日期字符串（"2006-01-02 15:04:05" 格式）
func futureDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02 15:04:05")
}

// TestExecuteChain_WithinDaysFilter_Regression 验证 ExecuteChain 调用点正确传递 withinDays 过滤
// 场景：左链第一步跨表查找 DrawSkin（伪），其中只有未来5天的行应匹配，未来100天的行应被过滤
func TestExecuteChain_WithinDaysFilter_Regression(t *testing.T) {
	// 当前表：仅触发链查找的输入列
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "trigger"},
	})

	// 目标表：DrawSkin（伪），包含 StartTime（过滤列）+ ByProduct（提取列）
	// 行5: 未来5天 → 应在 20 天窗口内，被保留
	// 行6: 未来100天 → 超出窗口，应被过滤
	drawSkinFile := buildChainTestRefSheet("DrawSkin",
		[]string{"Id", "StartTime", "ByProduct"},
		[][]string{
			{"trigger", futureDate(5), "ITEM_IN_WINDOW"},
			{"trigger", futureDate(100), "ITEM_OUT_OF_WINDOW"},
		})

	sheetMap := map[string]*excelize.File{"DrawSkin": drawSkinFile}

	cfg := ChainConfig{
		Steps: []ChainStep{
			{
				Sheet:      "DrawSkin",
				PreCol:     "Id",
				FindVal:    "col",
				NextCol:    "ByProduct",
				FilterCol:  "StartTime",
				FilterMode: "withinDays",
				FilterDays: "20",
			},
		},
	}

	result, err := ExecuteChain(cols, 0, 5, 4, cfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// 修复后：只有窗口内的行参与匹配
	assert.Equal(t, []string{"ITEM_IN_WINDOW"}, result.MatchValues(),
		"withinDays 过滤未生效：窗口外的行未被过滤掉（这是修复前的 bug 行为）")
}

// TestOnion_WithinDaysFilter_LeftStep 验证 onion 模型左链步骤正确传递 withinDays 过滤
// 场景：左链跨表查找 DrawSkin，未来5天行应保留、未来100天行应过滤
func TestOnion_WithinDaysFilter_LeftStep(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "trigger"},
	})

	drawSkinFile := buildChainTestRefSheet("DrawSkin",
		[]string{"Id", "StartTime", "ByProduct"},
		[][]string{
			{"trigger", futureDate(5), "ITEM_IN_WINDOW"},
			{"trigger", futureDate(100), "ITEM_OUT_OF_WINDOW"},
		})

	sheetMap := map[string]*excelize.File{"DrawSkin": drawSkinFile}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:      "DrawSkin",
					PreCol:     "Id",
					FindVal:    "col",
					NextCol:    "ByProduct",
					FilterCol:  "StartTime",
					FilterMode: "withinDays",
					FilterDays: "20",
				},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{}},
	}

	params := map[string]string{"chainCompare": "verify_exists"}
	ctx := buildOnionTestContext(cols, 0, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)
	assert.NoError(t, err)
	// 修复后：左链结果只包含窗口内的提取值
	assert.Equal(t, []string{"ITEM_IN_WINDOW"}, ctx.LeftCurrentValues,
		"onion 左链 withinDays 过滤未生效（这是修复前的 bug 行为）")
}

// TestOnion_WithinDaysFilter_RightStep 验证 onion 模型右链反向查找正确传递 withinDays 过滤
// 直接调用 RightStepHandler.Handle()，绕过 Match 门控（参考 TestOnion_ReverseLookup_TwoSteps 模式）
//
// 链路：模拟 Match 已通过，RightCurrentValues="G001" →
// RightStep 在 DrawSkin 表中用 ByProduct="G001" 反向查找 →
// 两行都满足 ByProduct=G001，但 withinDays=20 过滤后只有未来5天行命中
// 修复前 bug：过滤未生效，两行都命中，RightCurrentValues 包含 ["A","B"]
func TestOnion_WithinDaysFilter_RightStep(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "TargetGroup", "", "", "G001"},
	})

	drawSkinFile := buildChainTestRefSheet("DrawSkin",
		[]string{"Id", "StartTime", "ByProduct"},
		[][]string{
			{"A", futureDate(5), "G001"},   // 窗口内：应命中
			{"B", futureDate(100), "G001"}, // 窗口外：应被过滤
		})
	sheetMap := map[string]*excelize.File{"DrawSkin": drawSkinFile}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "TargetGroup"},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:      "DrawSkin",
					PreCol:     "Id", // 反向查找的提取列
					FindVal:    "self",
					NextCol:    "ByProduct", // 反向查找的搜索列
					FilterCol:  "StartTime",
					FilterMode: "withinDays",
					FilterDays: "20",
				},
			},
		},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, config, nil, nil, 0)
	// 模拟 Match 已通过，注入右链初始值
	ctx.Matched = true
	ctx.RightCurrentValues = []string{"G001"}

	// 直接调用 RightStepHandler 执行反向查找
	step0 := &RightStepHandler{
		Step:     config.Right.Steps[0],
		StepIdx:  0,
		IsFirst:  true,
		ChainCfg: config.Right,
	}
	err := step0.Handle(ctx, func(ctx *ChainContext) error { return nil })
	assert.NoError(t, err)

	// 修复后：反向查找在 ByProduct="G001" 行中只有未来5天的行（A）通过 withinDays 过滤
	// 修复前 bug：过滤被忽略，B 也会被包含
	assert.Equal(t, []string{"A"}, ctx.RightCurrentValues,
		"RightStepHandler 反向查找未应用 withinDays 过滤（修复前 bug 行为）")
}

// TestOnion_WithinDaysFilter_MatchStep 验证 onion 模型 Match 阶段正确传递 withinDays 过滤
// 直接调用 MatchHandler.Handle()，验证 getRightFinalValues 中的 FilterRowsByConditionEx 过滤生效
//
// 链路：左链最终值 ["G001"] → MatchHandler 打开 DrawSkin 右链最终表 → 通过 withinDays 过滤
//
//	未来5天行 ByProduct=G001 保留，未来100天行 ByProduct=G002 也保留（如果过滤失效）
//
// 通过让两行有不同的 NextCol 值（G001/G002），观察 ctx.RightCurrentValues 是否只包含窗口内的 G001
// 修复前 bug：MatchHandler.getRightFinalValues 未应用 withinDays，右链最终值集合包含 [G001, G002]
//
//	且左链 G001 仍匹配，传递给右链的 passToRight=[G001,G002]
//
// 修复后：右链最终值集合只剩 [G001]，passToRight=[G001]
func TestOnion_WithinDaysFilter_MatchStep(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "TargetGroup", "", "", "G001"},
	})

	// 右链最终表：DrawSkin，ByProduct 列是 NextCol（右链最终值提取列）
	// 窗口内行的 ByProduct=G001，窗口外行的 ByProduct=G002
	drawSkinFile := buildChainTestRefSheet("DrawSkin",
		[]string{"Id", "StartTime", "ByProduct"},
		[][]string{
			{"A", futureDate(5), "G001"},   // 窗口内：应保留
			{"B", futureDate(100), "G002"}, // 窗口外：应被过滤
		})
	sheetMap := map[string]*excelize.File{"DrawSkin": drawSkinFile}

	rightLastStep := ChainStep{
		Sheet:      "DrawSkin",
		PreCol:     "Id",
		FindVal:    "self",
		NextCol:    "ByProduct",
		FilterCol:  "StartTime",
		FilterMode: "withinDays",
		FilterDays: "20",
	}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{Sheet: "", PreCol: "", FindVal: "col", NextCol: "TargetGroup"},
			},
		},
		Right: ChainConfig{Steps: []ChainStep{rightLastStep}},
	}

	ctx := NewChainContext(cols, nil, 1, 5, 4, sheetMap, config, nil, nil, 0)
	// 模拟左链已完成，注入左链最终值
	ctx.LeftCurrentValues = []string{"G001"}

	matchHandler := &MatchHandler{
		MatchType:     "verify_exists",
		RightLastStep: rightLastStep,
		RightConfig:   config.Right,
	}
	nextCalled := false
	err := matchHandler.Handle(ctx, func(ctx *ChainContext) error {
		nextCalled = true
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, ctx.Matched, "G001 应在窗口内行中存在，Match 应通过")
	assert.True(t, nextCalled, "Match 通过后应调用 next")

	// 修复后：getRightFinalValues 应用 withinDays 过滤，右链最终值集合只剩 [G001]
	//   ctx.RightCurrentValues = passToRight = [G001]
	// 修复前 bug：过滤未生效，右链最终值集合 = [G001, G002]
	//   ctx.RightCurrentValues = [G001, G002]
	assert.Equal(t, []string{"G001"}, ctx.RightCurrentValues,
		"MatchHandler.getRightFinalValues 未应用 withinDays 过滤（修复前 bug 行为）")
}

// ==================== verify_must_exist 强制引用完整性测试 ====================
// 这些测试覆盖方案 D：新增 verify_must_exist 匹配类型，与 verify_exists 区分语义。
//   - verify_exists：门控语义，左链值不在右链时静默跳过（不报错）
//   - verify_must_exist：引用完整性语义，左链值不在右链时报错并报告缺失值

// TestMatchByType_VerifyMustExist_AllExist 全部存在时返回 true、无 reason
func TestMatchByType_VerifyMustExist_AllExist(t *testing.T) {
	matched, reason := MatchByType("verify_must_exist", []string{"a", "b"}, []string{"a", "b", "c"})
	assert.True(t, matched)
	assert.Empty(t, reason)
}

// TestMatchByType_VerifyMustExist_Missing 部分缺失时返回 false、reason 含缺失值列表
func TestMatchByType_VerifyMustExist_Missing(t *testing.T) {
	matched, reason := MatchByType("verify_must_exist", []string{"a", "b", "x"}, []string{"a", "b"})
	assert.False(t, matched)
	assert.Contains(t, reason, "x", "reason 应包含缺失值")
	assert.Contains(t, reason, "不存在", "reason 应说明引用缺失")
}

// TestMatchByType_VerifyExists_Missing 对照：verify_exists 缺失时仅返回 false，不带 reason
func TestMatchByType_VerifyExists_Missing(t *testing.T) {
	matched, reason := MatchByType("verify_exists", []string{"a", "x"}, []string{"a", "b"})
	assert.False(t, matched)
	assert.Empty(t, reason, "verify_exists 是门控类型，失败时不应携带 reason")
}

// TestIsMatchTypeStrict 验证强制类型识别
func TestIsMatchTypeStrict(t *testing.T) {
	assert.True(t, IsMatchTypeStrict("verify_must_exist"))
	assert.False(t, IsMatchTypeStrict("verify_exists"))
	assert.False(t, IsMatchTypeStrict("date_equals"))
	assert.False(t, IsMatchTypeStrict(""))
}

// TestOnion_VerifyMustExist_ReportsMissing 端到端：用户实际场景的简化版
// 链路：DrawSkin.byproduct 提取 itemId → 验证存在于 ShopGoods.Item
// 当 ShopGoods 缺少某个 itemId 时，verify_must_exist 应产生 Violation 并报告缺失值
func TestOnion_VerifyMustExist_ReportsMissing(t *testing.T) {
	// 当前表 DrawSkin，byproduct 列含 "{91001;1}" 引用
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "23"},
		{"", "", "byproduct", "", "", "{91001;1}"},
	})

	// 右链最终表 ShopGoods，Item 列**不含** 91001（用户删除后场景）
	shopGoodsFile := buildChainTestRefSheet("ShopGood",
		[]string{"Id", "Item"},
		[][]string{
			{"1", "{12345;1}"},
			{"2", "{67890;1}"},
		})
	sheetMap := map[string]*excelize.File{"ShopGood": shopGoodsFile}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:   "",
					PreCol:  "",
					FindVal: "col",
					NextCol: "byproduct",
					Pattern: `\{(\d+);\d+\}`,
					Groups:  "1",
				},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:   "ShopGood",
					PreCol:  "Item",
					FindVal: "col",
					Pattern: `\{(\d+);\d+\}`,
					Groups:  "1",
					IsArray: "true",
				},
			},
		},
	}
	params := map[string]string{
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_must_exist",
	}

	ctx := buildOnionTestContext(cols, 1, 5, sheetMap, config, params)
	err := runOnionChain(ctx, config, params)
	assert.NoError(t, err)
	assert.True(t, ctx.Violation, "verify_must_exist 缺失引用时应报错")
	assert.Contains(t, ctx.Reason, "91001", "Reason 应包含缺失的 itemId")
}

// TestOnion_VerifyExists_StaysSilent 对照：同样配置但用 verify_exists（门控），应静默不报错
func TestOnion_VerifyExists_StaysSilent(t *testing.T) {
	cols := buildChainTestCols([][]string{
		{"", "", "Id", "", "", "23"},
		{"", "", "byproduct", "", "", "{91001;1}"},
	})

	shopGoodsFile := buildChainTestRefSheet("ShopGood",
		[]string{"Id", "Item"},
		[][]string{
			{"1", "{12345;1}"},
			{"2", "{67890;1}"},
		})
	sheetMap := map[string]*excelize.File{"ShopGood": shopGoodsFile}

	config := &ChainPairConfig{
		Left: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:   "",
					PreCol:  "",
					FindVal: "col",
					NextCol: "byproduct",
					Pattern: `\{(\d+);\d+\}`,
					Groups:  "1",
				},
			},
		},
		Right: ChainConfig{
			Steps: []ChainStep{
				{
					Sheet:   "ShopGood",
					PreCol:  "Item",
					FindVal: "col",
					Pattern: `\{(\d+);\d+\}`,
					Groups:  "1",
					IsArray: "true",
				},
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
	assert.False(t, ctx.Violation, "verify_exists 是门控类型，缺失时不应报错")
	assert.False(t, ctx.Matched, "门控未通过")
}

// ==================== P2 回归：filterCol 不存在显式报错 ====================

// TestFilterRowsByConditionEx_FilterColNotFound 单值模式 filterCol 不存在应返回 error
func TestFilterRowsByConditionEx_FilterColNotFound(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2"},
		{"", "", "Type", "", "", "Hero", "Item"},
	}
	rows, err := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "NonExistentCol",
		FilterVal:     "Hero",
		StartRowIdx:   5,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NonExistentCol")
	assert.Contains(t, err.Error(), "不存在")
	assert.Nil(t, rows)
}

// TestFilterRowsByConditionEx_WithinDays_FilterColNotFound withinDays 模式 filterCol 不存在应返回 error
func TestFilterRowsByConditionEx_WithinDays_FilterColNotFound(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "Type", "", "", "Hero"},
	}
	rows, err := FilterRowsByConditionEx(cols, FilterOptions{
		FilterColName: "NonExistentTimeCol",
		StartRowIdx:   5,
		FilterMode:    "withinDays",
		FilterDays:    "7",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NonExistentTimeCol")
	assert.Nil(t, rows)
}

// TestFilterRowsByConditionEx_WithinDays_InvalidFilterDays 非法 filterDays 应返回 error
func TestFilterRowsByConditionEx_WithinDays_InvalidFilterDays(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "StartTime", "", "", "2099-01-01 00:00:00"},
	}
	for _, days := range []string{"abc", "", "  "} {
		rows, err := FilterRowsByConditionEx(cols, FilterOptions{
			FilterColName: "StartTime",
			StartRowIdx:   4,
			FilterMode:    "withinDays",
			FilterDays:    days,
		})
		assert.Error(t, err, "filterDays=%q 应报错", days)
		assert.Contains(t, err.Error(), "filterDays")
		assert.Nil(t, rows)
	}
}
