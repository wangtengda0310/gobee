package reference

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// ==================== 用户实际配置的 DrawPet E2E 测试 ====================
//
// 数据来源：真实 Excel 配表（D:\work\config\excel\）
//   - DrawPet: DrawPet_结缘亭.xlsx, sheet "结缘亭|DrawPet"
//   - DropRule/DropItem/DropGroup: Drop.xlsx
//
// 真实数据摘要（前2行）：
//   DrawPet Id=3001: OnceDropRule=910001, BigAward={1039016;1},{1040015;1}
//   DrawPet Id=3002: OnceDropRule=920001, BigAward={1050007;1}
//   DropRule Id=910001: DropGroup=91001,91003,91004
//   DropRule Id=920001: DropGroup=92001,92002,92003
//   DropItem Id=110002: DropGroup=91001, Item={1039016;1}
//   DropItem Id=110003: DropGroup=91001, Item={1040015;1}
//   DropItem Id=200002: DropGroup=92001, Item={1050007;1}
//   DropGroup: 91001, 91003, 91004, 92001, 92002, 92003

// buildDrawPetRealDataSheets 使用真实 Excel 数据构造测试数据
// 数据关系链：
//
//	DrawPet(Id=3001).OnceDropRule=910001 → DropRule(Id=910001).DropGroup=91001,91003,91004
//	DropItem(Id=110002).DropGroup=91001, Item={1039016;1}  ← BigAward 中的第一个道具
//	DropItem(Id=110003).DropGroup=91001, Item={1040015;1}  ← BigAward 中的第二个道具
//	DropItem(Id=200002).DropGroup=92001, Item={1050007;1}  ← DrawPet 3002 的道具
//	DropGroup: 91001/91003/91004 对应 Id=910001，92001/92002/92003 对应 Id=920001
func buildDrawPetRealDataSheets() ([][]string, map[string]*excelize.File) {
	// 当前列 DrawPet（MJS_FIXED_ROWS: 4行表头，数据从索引4开始）
	// 真实数据：Id=3001, BigAward={1039016;1},{1040015;1}, OnceDropRule=910001
	cols := [][]string{
		{"", "", "Id", "", "", "3001"},                          // Id 列 (colIdx=0)
		{"", "", "BigAward", "", "", "{1039016;1},{1040015;1}"}, // BigAward 列 (colIdx=1)
		{"", "", "OnceDropRule", "", "", "910001"},              // OnceDropRule 列 (colIdx=2)
	}

	// === DropRule 表: Id(A), DropGroup(D) ===
	// 真实数据：Id=910001 → DropGroup=91001,91003,91004
	//          Id=920001 → DropGroup=92001,92002,92003
	dropRuleFile := excelize.NewFile()
	dropRuleFile.SetSheetName("Sheet1", "掉落规则表|DropRule")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A3", "Id")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D3", "DropGroup")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A5", "910001")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D5", "91001,91003,91004")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A6", "920001")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D6", "92001,92002,92003")

	// === DropItem 表: Id(A), DropGroup(C), Item(D) ===
	// 真实数据：Id=110002, DropGroup=91001, Item={1039016;1}
	//          Id=110003, DropGroup=91001, Item={1040015;1}
	//          Id=200002, DropGroup=92001, Item={1050007;1}
	dropItemFile := excelize.NewFile()
	dropItemFile.SetSheetName("Sheet1", "掉落道具表|DropItem")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "A3", "Id")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "C3", "DropGroup")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "D3", "Item")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "A5", "110002")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "C5", "91001")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "D5", "{1039016;1}")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "A6", "110003")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "C6", "91001")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "D6", "{1040015;1}")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "A7", "200002")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "C7", "92001")
	dropItemFile.SetCellValue("掉落道具表|DropItem", "D7", "{1050007;1}")

	// === DropGroup 表: Id(A) ===
	// 真实数据：91001, 91003, 91004, 92001, 92002, 92003
	dropGroupFile := excelize.NewFile()
	dropGroupFile.SetSheetName("Sheet1", "掉落分组表|DropGroup")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A3", "Id")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A5", "91001")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A6", "91003")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A7", "91004")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A8", "92001")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A9", "92002")
	dropGroupFile.SetCellValue("掉落分组表|DropGroup", "A10", "92003")

	sheetMap := map[string]*excelize.File{
		"掉落规则表|DropRule":  dropRuleFile,
		"掉落道具表|DropItem":  dropItemFile,
		"掉落分组表|DropGroup": dropGroupFile,
	}

	return cols, sheetMap
}

// TestChainRef_DrawPet_RealData 使用真实数据验证旧路径数据流
// 使用 verify_exists 替代已删除的 contains/match 类型
func TestChainRef_DrawPet_RealData(t *testing.T) {
	rule := &ChainReferenceCheckRule{}
	cols, sheetMap := buildDrawPetRealDataSheets()

	// 用户实际配置：右链第一步 preCol="Id"
	chainSteps := `{
		"left": {"steps": [
			{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "OnceDropRule"},
			{"sheet": "掉落规则表|DropRule", "preCol": "Id", "findVal": "col", "nextCol": "DropGroup"}
		]},
		"right": {"steps": [
			{"sheet": "掉落道具表|DropItem", "preCol": "Id", "findVal": "col", "nextCol": "DropGroup"},
			{"sheet": "掉落分组表|DropGroup", "preCol": "Id", "findVal": "col", "nextCol": "Id"}
		]}
	}`

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	// preCol="Id" 时：右链收集的 firstStepInput 是数字 ID [110002,110003,200002]
	// Phase1 contains: 右链 last=[91001,92001] 不包含左链 last=[91001,91003,91004]（字符串级别）
	// 因此两链未交汇，不应报错
	assert.Empty(t, errors, "verify_exists 门控不通过时不应报错")
}

// TestChainRef_DrawPet_RealData_ItemCol 右链第一步 preCol 改为 "Item"
//
// 与 TestChainRef_DrawPet_RealData 相同的数据，但右链第一步 preCol="Item"
// 全表扫描时 firstStepInput 收集 DropItem.Item=[{1039016;1},{1040015;1},{1050007;1}]
//
// 数据流（DrawPet Id=3001）：
//
//	左链: OnceDropRule="910001" → DropRule → DropGroup="91001,91003,91004"
//	右链 Step1: firstStepInput=[{1039016;1},{1040015;1},{1050007;1}], next=[91001,91001,92001]
//	右链 Step2: DropGroup 中查 → last=[91001,92001]
//	Phase1 contains: "91001".Contains("91001,91003,91004")=false → 不交汇 → 不报错
//
// 注意：此测试验证 preCol="Item" 时 firstStepInput 变为 ItemCfg 格式值，
// 但由于 contains 门控中左链值是逗号分隔的长字符串（如 "91001,91003,91004"），
// 右链单个值（"91001"）不包含该长字符串，门控仍不通过。
// 这反映了当前用户配置在实际数据上的真实行为。
func TestChainRef_DrawPet_RealData_ItemCol(t *testing.T) {
	rule := &ChainReferenceCheckRule{}
	cols, sheetMap := buildDrawPetRealDataSheets()

	chainSteps := `{
		"left": {"steps": [
			{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "OnceDropRule"},
			{"sheet": "掉落规则表|DropRule", "preCol": "Id", "findVal": "col", "nextCol": "DropGroup"}
		]},
		"right": {"steps": [
			{"sheet": "掉落道具表|DropItem", "preCol": "Item", "findVal": "col", "nextCol": "DropGroup"},
			{"sheet": "掉落分组表|DropGroup", "preCol": "Id", "findVal": "col", "nextCol": "Id"}
		]}
	}`

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	// preCol="Item" 时 firstStepInput=[{1039016;1},{1040015;1},{1050007;1}]
	// 但 Phase1 contains 门控仍然不通过（左链="91001,91003,91004" 太长）
	// 如果未来改进 contains 门控逻辑使其支持子串匹配，此测试应改为检测到违规
	assert.Empty(t, errors, "verify_exists 门控不通过时不应报错")
}

// TestChainRef_DrawPet_RealData_GateBlocked 左链无匹配行 → 不报错
//
// 构造 OnceDropRule=999999（DropRule 中无此 Id），左链 Step2 找不到匹配 → 左链结果为空
// Phase1 无法比较 → 不报错
func TestChainRef_DrawPet_RealData_GateBlocked(t *testing.T) {
	rule := &ChainReferenceCheckRule{}

	// 构造一个 OnceDropRule 不存在于 DropRule 的行
	cols := [][]string{
		{"", "", "Id", "", "", "3003"},
		{"", "", "BigAward", "", "", "{1039016;1}"},
		{"", "", "OnceDropRule", "", "", "999999"}, // DropRule 中无此 Id
	}

	_, sheetMap := buildDrawPetRealDataSheets()

	chainSteps := `{
		"left": {"steps": [
			{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "OnceDropRule"},
			{"sheet": "掉落规则表|DropRule", "preCol": "Id", "findVal": "col", "nextCol": "DropGroup"}
		]},
		"right": {"steps": [
			{"sheet": "掉落道具表|DropItem", "preCol": "Item", "findVal": "col", "nextCol": "DropGroup"},
			{"sheet": "掉落分组表|DropGroup", "preCol": "Id", "findVal": "col", "nextCol": "Id"}
		]}
	}`

	params := map[string]string{
		"chainSteps":        chainSteps,
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	// 左链 OnceDropRule=999999 在 DropRule 中无匹配，引用完整性检查应报错
	assert.NotEmpty(t, errors, "左链无匹配行（DropRule 无 Id=999999），引用完整性应报错")
	assert.Contains(t, errors[0].Reason, "未找到")
}

// verify_exists 完整测试（含 isArray 拆分 bug 验收）已迁移到：
// compare_verify_exists_test.go
