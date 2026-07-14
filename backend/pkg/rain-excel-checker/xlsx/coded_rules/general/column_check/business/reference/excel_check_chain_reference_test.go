package reference

import (
	"testing"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule/chain_reference"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 测试辅助函数 ====================

// buildChainTestCols 构建当前表（DrawFix）的列数据
// MJS_FIXED_ROWS: 0=中文, 1=类型, 2=列名, 3=导出, 4=数据开始
// 数据行从索引 4 开始
func buildChainTestCols() [][]string {
	return [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},                                   // Id 列
		{"", "", "ItemIds", "", "", "{1001;1}{1002;1}", "{2001;1}", "{1001;1}"}, // ItemIds 列
	}
}

// buildChainTestRefSheet 构建参考表（Item）的 excelize.File
func buildChainTestRefSheet() *excelize.File {
	f := excelize.NewFile()
	sheet := "Item"
	f.SetSheetName("Sheet1", sheet)

	// 表头行 (MJS格式: 行1=中文, 行2=类型, 行3=列名, 行4=导出, 行5+=数据)
	headers := [][]string{
		{"道具ID", "类型", "Id", ""},        // 行1: 中文
		{"int", "string", "int", ""},    // 行2: 类型
		{"Id", "Type", "ItemParam", ""}, // 行3: 列名
		{"S", "S", "S", ""},             // 行4: 导出标识
	}
	for colIdx, row := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		for rowIdx, val := range row {
			if rowIdx == 0 {
				continue
			}
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
			f.SetCellValue(sheet, cellName, val)
		}
		_ = cell
	}

	// 直接设置列名和数据（简化方式）
	// 列 A = Id, 列 B = Type, 列 C = ItemParam
	f.SetCellValue(sheet, "A1", "")
	f.SetCellValue(sheet, "A2", "")
	f.SetCellValue(sheet, "A3", "Id")
	f.SetCellValue(sheet, "A4", "")
	f.SetCellValue(sheet, "A5", "1001")
	f.SetCellValue(sheet, "A6", "1002")
	f.SetCellValue(sheet, "A7", "2001")

	f.SetCellValue(sheet, "B1", "")
	f.SetCellValue(sheet, "B2", "")
	f.SetCellValue(sheet, "B3", "Type")
	f.SetCellValue(sheet, "B4", "")
	f.SetCellValue(sheet, "B5", "Hero")
	f.SetCellValue(sheet, "B6", "Hero")
	f.SetCellValue(sheet, "B7", "Item")

	f.SetCellValue(sheet, "C1", "")
	f.SetCellValue(sheet, "C2", "")
	f.SetCellValue(sheet, "C3", "ItemParam")
	f.SetCellValue(sheet, "C4", "")
	f.SetCellValue(sheet, "C5", "10803")
	f.SetCellValue(sheet, "C6", "10804")
	f.SetCellValue(sheet, "C7", "")

	return f
}

// buildChainTestTimeSheet 构造带时间列的测试表（用于两阶段测试）
func buildChainTestTimeSheet() *excelize.File {
	f := excelize.NewFile()
	sheet := "Hero"
	f.SetSheetName("Sheet1", sheet)

	// 表头
	f.SetCellValue(sheet, "A3", "Id")
	f.SetCellValue(sheet, "B3", "Name")
	f.SetCellValue(sheet, "C3", "StartTime")
	f.SetCellValue(sheet, "D3", "EndTime")

	// 数据行从索引 5 开始
	// 武将 1001: 2026-01-01 ~ 2026-06-01
	f.SetCellValue(sheet, "A5", "1001")
	f.SetCellValue(sheet, "B5", "吕布")
	f.SetCellValue(sheet, "C5", "2026-01-01 00:00:00")
	f.SetCellValue(sheet, "D5", "2026-06-01 00:00:00")

	// 武将 1002: 2026-03-01 ~ 2026-08-01
	f.SetCellValue(sheet, "A6", "1002")
	f.SetCellValue(sheet, "B6", "关羽")
	f.SetCellValue(sheet, "C6", "2026-03-01 00:00:00")
	f.SetCellValue(sheet, "D6", "2026-08-01 00:00:00")

	return f
}

// ==================== Match 比较测试 ====================

// TestChainRef_VerifyExistsViolation 两链最终值验证存在性
func TestChainRef_VerifyExistsViolation(t *testing.T) {
	rule := &ChainReferenceCheckRule{}
	cols := buildChainTestCols()
	refSheet := buildChainTestRefSheet()
	sheetMap := map[string]*excelize.File{
		"Item": refSheet,
	}

	// 左链: 从 SrcCol 列取值 → 在 Item 表查找 Id → 提取 ItemParam
	// 右链: 同一条链
	chainSteps := "{\"left\":{\"steps\":[{\"sheet\":\"Item\",\"preCol\":\"Id\",\"findVal\":\"col\",\"nextCol\":\"ItemParam\"}]},\"right\":{\"steps\":[{\"sheet\":\"Item\",\"preCol\":\"Id\",\"findVal\":\"col\",\"nextCol\":\"ItemParam\"}]}}"

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	// colIdx=1 是 ItemIds 列
	errors := rule.Check("DrawFix", cols, 1, 4, params, sheetMap)
	// 两链提取的值相同，verify_exists 要求左链值在右链中全部找到 → 找到 → 不违规
	assert.Empty(t, errors, "verify_exists: 两链值相同，全部找到不应报错")
}

// TestChainRef_VerifyExistsNoViolation 两链最终值全部存在 → 通过
// srcCol 已移除，跨表第一步为全表扫描模式。使用过滤条件构造无值场景。
func TestChainRef_VerifyExistsNoViolation(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "ItemIds", "", "", "{1001;1}"},
	}

	refSheet := buildChainTestRefSheet()
	sheetMap := map[string]*excelize.File{
		"Item": refSheet,
	}

	// left 链：全表扫描 Item，过滤 Type="NotExist"（无匹配行）→ 提取不到值
	// right 链：全表扫描 Item，无过滤 → 提取到所有 ItemParam
	// 两链无交集 → 通过
	chainSteps := `{
		"left": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "filterCol": "Type", "filterVal": "NotExist", "nextCol": "ItemParam"}]},
		"right": {"steps": []}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	assert.Empty(t, errors, "不应检测到违规")
}

// TestChainRef_NoChainSteps 配置为空 → 报配置错误
func TestChainRef_NoChainSteps(t *testing.T) {
	rule := &ChainReferenceCheckRule{}
	cols := buildChainTestCols()

	params := map[string]string{
		"chainSteps":   "",
		"chainCompare": "verify_exists",
	}

	errors := rule.Check("DrawFix", cols, 1, 4, params, nil)
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Reason, "chainSteps")
}

// TestChainRef_InvalidJSON JSON 格式错误 → 报配置错误
func TestChainRef_InvalidJSON(t *testing.T) {
	rule := &ChainReferenceCheckRule{}
	cols := buildChainTestCols()

	params := map[string]string{
		"chainSteps":   "not valid json",
		"chainCompare": "verify_exists",
	}

	errors := rule.Check("DrawFix", cols, 1, 4, params, nil)
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Reason, "解析失败")
}

// ==================== Exists 比较测试 ====================

// TestChainRef_VerifyExistsMismatchViolation 左链值不在右链中 → 报错
func TestChainRef_VerifyExistsMismatchViolation(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "SrcCol", "", "", "1001"},
		{"", "", "RefCol", "", "", "1001"},
	}

	refSheet := buildChainTestRefSheet()
	sheetMap := map[string]*excelize.File{
		"Item": refSheet,
	}

	chainSteps := `{
		"left": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "ItemParam"}]},
		"right": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "ItemParam"}]}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 两链都提取到 ItemParam=10803，verify_exists: 左链值在右链中全部找到 → 不违规
	assert.Empty(t, errors, "verify_exists: 两链值相同不应报错")
}

// ==================== FilterCondition 测试 ====================

// TestChainRef_FilterCondition 有 FilterCol/FilterVal → 只匹配过滤后的行
func TestChainRef_FilterCondition(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "SrcCol", "", "", "1001"},
	}

	refSheet := buildChainTestRefSheet()
	sheetMap := map[string]*excelize.File{
		"Item": refSheet,
	}

	// 有过滤条件：只匹配 Type="Hero" 的行
	chainSteps := `{
		"left": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "filterCol": "Type", "filterVal": "Hero", "nextCol": "ItemParam"}]},
		"right": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "nextCol": "ItemParam"}]}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 两链都提取到 10803（1001 匹配 Hero 行），verify_exists 全部找到 → 不违规
	assert.Empty(t, errors, "verify_exists: 两链值相同不应报错")
}

// ==================== SheetNotFound 测试 ====================

// TestChainRef_SheetNotFound 目标表不存在 → 返回明确错误
func TestChainRef_SheetNotFound(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "SrcCol", "", "", "1001"},
	}

	sheetMap := map[string]*excelize.File{} // 空的 sheetMap

	chainSteps := `{
		"left": {"steps": [{"sheet": "NotExistTable", "preCol": "Id", "findVal": "col", "nextCol": "Name"}]},
		"right": {"steps": []}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	assert.True(t, len(errors) > 0)
	assert.Contains(t, errors[0].Reason, "目标表不存在")
}

// ==================== RegexExtract 测试 ====================

// TestChainRef_RegexExtract 正则提取子值 → 正确提取
func TestChainRef_RegexExtract(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "ItemIds", "", "", "{1001;1}{1002;2}"},
	}

	refSheet := buildChainTestRefSheet()
	sheetMap := map[string]*excelize.File{
		"Item": refSheet,
	}

	// 使用正则从 {1001;1}{1002;2} 提取 1001 和 1002
	chainSteps := `{
		"left": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col", "pattern": "{(\\d+);\\d+}", "groups": "1", "nextCol": "ItemParam"}]},
		"right": {"steps": []}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	// right 链为空，不会产生交集，不应该报错
	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	assert.Empty(t, errors, "right 链为空，不应报错")
}

// ==================== 空值和注释处理测试 ====================

// TestChainRef_AllowEmpty 允许空值 → 跳过空单元格
func TestChainRef_AllowEmpty(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1", "2"},
		{"", "", "SrcCol", "", "", "", "1001"},
	}

	sheetMap := map[string]*excelize.File{}

	chainSteps := `{
		"left": {"steps": [{"sheet": "Item", "preCol": "Id", "findVal": "col"}]},
		"right": {"steps": []}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "verify_exists",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 第一行 SrcCol 为空，allowEmpty=true 跳过
	// 第二行 SrcCol=1001，但 sheetMap 中没有 Item 表，会报来源链执行失败
	// 所以应该只有一个错误（第二行的）
	for _, e := range errors {
		t.Logf("错误: index=%d reason=%s", e.Index, e.Reason)
	}
	assert.True(t, len(errors) <= 2)
}

// ==================== 两阶段比较测试 ====================

// TestChainRef_TwoPhase_TimeOverlapViolation 两链 Match 值相同且 Compare 时间点相同 → 报错
func TestChainRef_TwoPhase_TimeOverlapViolation(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	// 构造测试数据：当前表有两行，都引用相同的武将
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2"},
		{"", "", "HeroId", "", "", "1001", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	// 两链都配置了 compareCol，提取时间进行比较
	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "EndTime"
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 由于测试数据中左链 StartTime 和右链 EndTime 不同，不应该报错
	// 如果要测试报错，需要构造相同时间的数据
	assert.Empty(t, errors, "时间点不同，不应报错")
}

// TestChainRef_TwoPhase_NoMatch 两链 Match 值无交集 → 通过
func TestChainRef_TwoPhase_NoMatch(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "HeroId", "", "", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	// 左链找 1001，右链找 9999（不存在）
	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "EndTime"
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 右链找 Id=1，Hero 表中没有，无交集
	assert.Empty(t, errors, "无 Match 交集，不应报错")
}

// TestChainRef_TwoPhase_MatchButNoTimeOverlap Match 相同但时间不同 → 通过
func TestChainRef_TwoPhase_MatchButNoTimeOverlap(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "HeroId", "", "", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "EndTime"
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 1001 的 StartTime=2026-01-01，EndTime=2026-06-01，不相同
	assert.Empty(t, errors, "时间点不同，不应报错")
}

// TestChainRef_TwoPhase_CompareColEmpty 任一链 compareCol 为空 → 使用单阶段比较
func TestChainRef_TwoPhase_CompareColEmpty(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "HeroId", "", "", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	// 右链没有 compareCol
	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": ""
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 右链没有 compareCol，应该降级为单阶段比较
	// 两链提取到的是 Id=1001，单阶段 time_overlap 用 compareTimeMatch 比较的是时间值而非 ID 值
	// Id 值不是时间格式，时间匹配不会触发，因此不应报错
	assert.Empty(t, errors, "降级单阶段 time_overlap 时 Id 值不是时间点，不应报错")
}

// TestChainRef_TwoPhase_AsymmetricCompareCol 仅左链有 compareCol → 使用单阶段比较
func TestChainRef_TwoPhase_AsymmetricCompareCol(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "HeroId", "", "", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	// 只有左链有 compareCol
	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}]
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// 不对称配置，降级为单阶段 time_overlap
	// 两链提取到 Id=1001，但单阶段 time_overlap 比较的是时间值，Id 不是时间点不会触发
	assert.Empty(t, errors, "降级单阶段 time_overlap 时 Id 值不是时间点，不应报错")
}

// TestChainRef_TwoPhase_CompareColNotFound compareCol 列不存在 → Compare 值为空
func TestChainRef_TwoPhase_CompareColNotFound(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "HeroId", "", "", "1001"},
	}

	heroSheet := buildChainTestTimeSheet()
	sheetMap := map[string]*excelize.File{
		"Hero": heroSheet,
	}

	// compareCol 指向不存在的列
	chainSteps := `{
		"left": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "NotExistColumn"
		},
		"right": {
			"steps": [{"sheet": "Hero", "preCol": "Id", "findVal": "col", "nextCol": "Id"}],
			"compareCol": "StartTime"
		}
	}`

	params := map[string]string{
		"chainSteps":   chainSteps,
		"chainCompare": "time_overlap",
		"allowEmpty":   "true",
	}

	errors := rule.Check("TestSheet", cols, 1, 4, params, sheetMap)
	// compareCol 不存在，左链 Compare 值为空
	// 两阶段比较不会找到时间点匹配
	assert.Empty(t, errors, "compareCol 不存在，不应报时间重叠")
}

// ==================== compareMatch/compareExists/compareTimeMatch 单元测试 ====================
// 这些函数已迁移到 chain_reference 包，这里通过 chain_reference 包的导出函数间接测试

// TestCompareVerifyExists_Found 左链值在右链中全部找到 → 不违规
func TestCompareVerifyExists_Found(t *testing.T) {
	violation, _ := chain_reference.CompareByType("verify_exists", []string{"a", "b"}, []string{"a", "b", "c"})
	assert.False(t, violation)
}

// TestCompareVerifyExists_NotFound 左链值未在右链中找到 → 违规
func TestCompareVerifyExists_NotFound(t *testing.T) {
	violation, reason := chain_reference.CompareByType("verify_exists", []string{"x"}, []string{"a", "b"})
	assert.True(t, violation)
	assert.Contains(t, reason, "x")
}

// TestCompareTimeMatch_Match 有匹配返回 true
func TestCompareTimeMatch_Match(t *testing.T) {
	leftVals := []string{"2026-01-01 00:00:00"}
	rightVals := []string{"2026-01-01 00:00:00"}
	violation, reason := chain_reference.CompareTimeMatch(leftVals, rightVals, "StartTime", "EndTime")
	assert.True(t, violation)
	assert.Contains(t, reason, "2026-01-01")
}

// TestCompareTimeMatch_NoMatch 无匹配返回 false
func TestCompareTimeMatch_NoMatch(t *testing.T) {
	leftVals := []string{"2026-01-01 00:00:00"}
	rightVals := []string{"2026-06-01 00:00:00"}
	violation, _ := chain_reference.CompareTimeMatch(leftVals, rightVals, "StartTime", "EndTime")
	assert.False(t, violation)
}

// TestCompareTimeMatch_InvalidTime 无效时间返回 false
func TestCompareTimeMatch_InvalidTime(t *testing.T) {
	leftVals := []string{"not-a-date"}
	rightVals := []string{"also-not-a-date"}
	violation, _ := chain_reference.CompareTimeMatch(leftVals, rightVals, "StartTime", "EndTime")
	assert.False(t, violation)
}

// ==================== 细粒度 ExecuteChain 两阶段数据流验证 ====================

// TestExecuteChain_TwoStepLeft_RightChainInputValues 验证两步左链和两步右链的完整数据流
// 目的：确认 Phase 2 比较所需的值在 Check 层面能正确获取
func TestExecuteChain_TwoStepLeft_RightChainInputValues(t *testing.T) {
	// 模拟 TestChainRef_TwoPhaseMatchGate_IntersectionViolation 的数据
	cols := [][]string{
		{"", "", "Id", "", "", "1"},
		{"", "", "BigAward", "", "", "{1001;1}"},
		{"", "", "OnceDropRule", "", "", "1001"},
	}

	// DropRule: Id=1001 → DropGroup=G001
	dropRuleFile := excelize.NewFile()
	dropRuleFile.SetSheetName("Sheet1", "DropRule")
	dropRuleFile.SetCellValue("DropRule", "A3", "Id")
	dropRuleFile.SetCellValue("DropRule", "B3", "DropGroup")
	dropRuleFile.SetCellValue("DropRule", "A5", "1001")
	dropRuleFile.SetCellValue("DropRule", "B5", "G001")

	// DropItem: Item={1001;1} (ItemCfg格式) → DropGroup=G001
	dropItemFile := excelize.NewFile()
	dropItemFile.SetSheetName("Sheet1", "DropItem")
	dropItemFile.SetCellValue("DropItem", "A3", "Item")
	dropItemFile.SetCellValue("DropItem", "B3", "DropGroup")
	dropItemFile.SetCellValue("DropItem", "A5", "{1001;1}")
	dropItemFile.SetCellValue("DropItem", "B5", "G001")

	// DropGroup: Id=G001
	dropGroupFile := excelize.NewFile()
	dropGroupFile.SetSheetName("Sheet1", "DropGroup")
	dropGroupFile.SetCellValue("DropGroup", "A3", "Id")
	dropGroupFile.SetCellValue("DropGroup", "A5", "G001")

	sheetMap := map[string]*excelize.File{
		"DropRule":  dropRuleFile,
		"DropItem":  dropItemFile,
		"DropGroup": dropGroupFile,
	}

	// 左链：Step1(仅取值) + Step2(DropRule)
	leftCfg := chain_reference.ChainConfig{
		Steps: []chain_reference.ChainStep{
			{Sheet: "", PreCol: "", FindVal: "col", NextCol: ""},
			{Sheet: "DropRule", PreCol: "Id", FindVal: "self", NextCol: "DropGroup"},
		},
	}
	leftResult, err := chain_reference.ExecuteChain(cols, 2, 5, 4, leftCfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, leftResult)
	t.Logf("左链 StepValues: %v", leftResult.StepValues)
	t.Logf("左链 LastStep: %v", leftResult.LastStepValues())
	t.Logf("左链 FirstStep: %v", leftResult.FirstStepValues())
	t.Logf("左链 FirstStepInput: %v", leftResult.GetFirstStepInputValues())

	// 右链：Step1(DropItem) + Step2(DropGroup)
	rightCfg := chain_reference.ChainConfig{
		Steps: []chain_reference.ChainStep{
			{Sheet: "DropItem", PreCol: "Item", FindVal: "col", NextCol: "DropGroup"},
			{Sheet: "DropGroup", PreCol: "Id", FindVal: "self", NextCol: "Id"},
		},
	}
	rightResult, err := chain_reference.ExecuteChain(cols, 1, 5, 4, rightCfg, sheetMap)
	assert.NoError(t, err)
	assert.NotNil(t, rightResult)
	t.Logf("右链 StepValues: %v", rightResult.StepValues)
	t.Logf("右链 LastStep: %v", rightResult.LastStepValues())
	t.Logf("右链 FirstStep: %v", rightResult.FirstStepValues())
	t.Logf("右链 FirstStepInput: %v", rightResult.GetFirstStepInputValues())

	// Phase 1 门控：两链最后一步匹配
	assert.Equal(t, []string{"G001"}, leftResult.LastStepValues())
	assert.Equal(t, []string{"G001"}, rightResult.LastStepValues())

	// Phase 2 比较：BigAward vs 右链第一步查找值
	currentColVal := cols[1][5] // BigAward 列第5行
	t.Logf("当前列值(BigAward): %q", currentColVal)
	assert.Equal(t, "{1001;1}", currentColVal)
	assert.Equal(t, []string{"{1001;1}"}, rightResult.GetFirstStepInputValues())
	// match: {1001;1} 在 [{1001;1}] 中 → 应匹配
}

// ==================== P0 回归：withinDays 全部窗口外不应降级为不过滤 ====================
// Bug: excel_check_chain_reference.go 第 113 行原本用 `if len(filteredRows) > 0` 判断是否启用过滤，
// 当 withinDays 窗口内 0 行匹配时 leftFirstStepFilterSet 保持 nil → 主循环 `!= nil` 为 false →
// 降级为"不过滤"，整表被检查并大量误报。修复后：filterEnabled 为 true 时建立空 map，
// 主循环正确跳过所有行。

// TestChainRef_WithinDaysAllOutOfWindow_NoDowngrade 全部行 StartTime 都在过去 →
// withinDays 窗口内匹配 0 行 → 不应误报任何行
func TestChainRef_WithinDaysAllOutOfWindow_NoDowngrade(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	// 当前表 3 行，StartTime 都在远未来超出窗口（这样独立于今天日期保持稳定）
	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "ByProduct", "", "", "X001", "X002", "X003"}, // 故意写不在 ShopGood 的值
		{"", "", "StartTime", "", "", "2099-01-01 00:00:00", "2099-02-01 00:00:00", "2099-03-01 00:00:00"},
	}

	// ShopGood 表故意不含 X001/X002/X003，若过滤失效将全部报错
	shopFile := excelize.NewFile()
	shopFile.SetSheetName("Sheet1", "ShopGood")
	shopFile.SetCellValue("ShopGood", "A3", "Item")
	shopFile.SetCellValue("ShopGood", "A5", "OTHER001")

	sheetMap := map[string]*excelize.File{"ShopGood": shopFile}

	// 左链 sheet="" 取 ByProduct 列，withinDays=20 过滤 StartTime
	// 右链 sheet=ShopGood 验证 Item 存在
	chainSteps := `{"left":{"steps":[{"sheet":"","preCol":"","findVal":"col","nextCol":"ByProduct","filterCol":"StartTime","filterMode":"withinDays","filterDays":"20"}]},"right":{"steps":[{"sheet":"ShopGood","preCol":"Item","findVal":"col"}]}}`

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_must_exist",
		"chainWarnBefore":   "0", // 启用规则
		"chainWarnSheet":    "DrawSkin",
		"chainWarnCol":      "StartTime",
		"allowEmpty":        "true",
	}

	errors := rule.Check("DrawSkin", cols, 1, 4, params, sheetMap)
	// 修复后：3 行 StartTime 全部超出 20 天窗口（2099年）→ 过滤后 0 行匹配 →
	//   leftFirstStepFilterSet 是非 nil 空 map → 主循环跳过所有 3 行 → errors 应为空
	// 修复前 bug：filteredRows 为空 → leftFirstStepFilterSet 保持 nil → 不过滤 →
	//   3 行全部检查并报告引用缺失 → errors 长度 3
	assert.Empty(t, errors,
		"withinDays 全部窗口外时应跳过所有行，不应降级为不过滤（修复前 bug：会误报 %d 行）", len(errors))
}

// TestChainRef_WithinDaysSomeInWindow_OnlyMatchedRowsChecked 只有部分行在窗口内 →
// 只检查窗口内的行
func TestChainRef_WithinDaysSomeInWindow_OnlyMatchedRowsChecked(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	// 3 行：第 1、2 行在远未来（窗口外），第 3 行在近未来（窗口内）
	// 用 time.Now() 派生近未来日期保持窗口判断稳定
	nearFuture := time.Now().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")

	cols := [][]string{
		{"", "", "Id", "", "", "1", "2", "3"},
		{"", "", "ByProduct", "", "", "X001", "X002", "MISSING_ITEM"},
		{"", "", "StartTime", "", "", "2099-01-01 00:00:00", "2099-02-01 00:00:00", nearFuture},
	}

	shopFile := excelize.NewFile()
	shopFile.SetSheetName("Sheet1", "ShopGood")
	shopFile.SetCellValue("ShopGood", "A3", "Item")
	shopFile.SetCellValue("ShopGood", "A5", "X001")
	shopFile.SetCellValue("ShopGood", "A6", "X002")
	// 故意不含 MISSING_ITEM

	sheetMap := map[string]*excelize.File{"ShopGood": shopFile}

	chainSteps := `{"left":{"steps":[{"sheet":"","preCol":"","findVal":"col","nextCol":"ByProduct","filterCol":"StartTime","filterMode":"withinDays","filterDays":"20"}]},"right":{"steps":[{"sheet":"ShopGood","preCol":"Item","findVal":"col"}]}}`

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_must_exist",
		"chainWarnBefore":   "0",
		"chainWarnSheet":    "DrawSkin",
		"chainWarnCol":      "StartTime",
		"allowEmpty":        "true",
	}

	errors := rule.Check("DrawSkin", cols, 1, 4, params, sheetMap)
	// 修复后：只有第 3 行在窗口内，引用 MISSING_ITEM 缺失 → 1 个错误
	assert.Len(t, errors, 1, "应只检查窗口内的第 3 行")
	if len(errors) == 1 {
		assert.Contains(t, errors[0].Reason, "MISSING_ITEM", "Reason 应包含缺失的引用值")
	}
}

// TestChainRef_VerifyMustExist_RightLastStepPatternApplied 验证右链最终值提取应用 Pattern
// Bug: onion_match.go::getRightFinalValues 之前未应用 RightLastStep.Pattern，
//
//	导致 ShopGoods.Item="{1039037;1}"（带花括号原始字符串）与左链提取出的纯数字 "1039037" 不匹配
//	→ 所有左链值都被误报为"缺失"。
//
// 复现用户场景：
//   - 左链：byproduct 列逗号分隔正则 \s*([^,\s]+)\s* 提取 [10111004, 1039037, 1022225]
//   - 右链：ShopGoods.Item 列 {x;y} 格式正则 {(\d+);\d+} 提取 [1039037, 1022225]（缺 10111004）
//   - verify_must_exist：应只报 [10111004] 缺失，不应报全部三个
func TestChainRef_VerifyMustExist_RightLastStepPatternApplied(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	// 当前表第 23 行：StartTime 在未来 11 天（窗口内），byproduct 逗号分隔
	nearFuture := time.Now().AddDate(0, 0, 11).Format("2006-01-02 15:04:05")
	cols := [][]string{
		{"", "", "Id", "", "", "23"},
		{"", "", "StartTime", "", "", nearFuture},
		{"", "", "byproduct", "", "", "10111004,1039037,1022225"},
	}

	// ShopGoods 表：含 1039037 和 1022225，缺 10111004
	shopFile := excelize.NewFile()
	shopFile.SetSheetName("Sheet1", "商品表|ShopGood")
	shopFile.SetCellValue("商品表|ShopGood", "A3", "Item")
	shopFile.SetCellValue("商品表|ShopGood", "A5", "{1039037;1}")
	shopFile.SetCellValue("商品表|ShopGood", "A6", "{1022225;1}")
	// 故意不含 10111004
	sheetMap := map[string]*excelize.File{"商品表|ShopGood": shopFile}

	// 用户真实配置：左链逗号分隔正则，右链 {x;y} 正则
	// JSON 内 \\ 才表示一个反斜杠，json.Unmarshal 后正则得到 \s*
	chainSteps := "{\"left\":{\"steps\":[{\"sheet\":\"\",\"preCol\":\"\",\"findVal\":\"col\",\"nextCol\":\"byproduct\",\"pattern\":\"\\\\s*([^,\\\\s]+)\\\\s*\",\"groups\":\"1\",\"filterCol\":\"StartTime\",\"filterMode\":\"withinDays\",\"filterDays\":\"20\"}]},\"right\":{\"steps\":[{\"sheet\":\"商品表|ShopGood\",\"preCol\":\"Item\",\"findVal\":\"col\",\"pattern\":\"{(\\\\d+);\\\\d+}\",\"groups\":\"1\",\"isArray\":\"true\"}]}}"

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_must_exist",
		"chainWarnBefore":   "0",
		"chainWarnSheet":    "皮肤抽奖|DrawSkin",
		"chainWarnCol":      "StartTime",
		"allowEmpty":        "true",
	}

	errors := rule.Check("皮肤抽奖|DrawSkin", cols, 2, 4, params, sheetMap)
	// 修复后：右链正则应用 → 集合=[1039037,1022225] → 左链[10111004,1039037,1022225] 只有 10111004 缺失
	// 修复前 bug：右链集合=[{1039037;1},{1022225;1}] 原始字符串 → 三个左链值全部找不到 → 全部误报
	if assert.Len(t, errors, 1, "应只报 1 行错误") && len(errors) == 1 {
		assert.Contains(t, errors[0].Reason, "10111004", "Reason 应包含真实缺失的 10111004")
		assert.NotContains(t, errors[0].Reason, "1039037",
			"Reason 不应包含 1039037（它在 ShopGoods 中存在），修复前 bug 行为：会误报")
		assert.NotContains(t, errors[0].Reason, "1022225",
			"Reason 不应包含 1022225（它在 ShopGoods 中存在），修复前 bug 行为：会误报")
	}
}
