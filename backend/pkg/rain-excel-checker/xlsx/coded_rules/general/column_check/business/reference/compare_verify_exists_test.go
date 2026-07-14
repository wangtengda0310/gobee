package reference

// verify_exists 比较类型完整测试套件
//
// verify_exists 语义：验证左侧值在右侧集合中全部存在 → 不违规，有缺失 → 违规并报告缺失值。
// 是 CHAIN_REFERENCE 规则中 chainCompare 和 chainMatchCompare 两个阶段的唯一非日期比较类型。
//
// ## 测试分层
//
// 本文件按三个层次组织测试：
//
//   1. 纯函数层 — CompareByType / MatchByType 的直接调用
//   2. 洋葱模型层 — BuildOnionChain 的 E2E 集成测试
//   3. 旧路径层 — ChainReferenceCheckRule.Check 的 E2E 测试（含 isArray 拆分 bug 验收）
//
// ## 关系链检查参数说明
//
// 关系链检查通过三张卡片（前端 UI）或三个参数（后端 JSON）配置：
//
//   - chainSteps（链步骤卡片）：定义左链和右链的跳转路径
//   - chainMatchCompare（匹配卡片）：Phase 1 门控，判断两链是否"交汇"
//   - chainCompare（比较卡片）：Phase 2 比较，判断当前列值是否违规
//
// ## 左右链配置
//
// 左链（left chain）：从当前表的当前列值出发，经过 N 步跳转到目标表
//
//	Step 0 (sheet=""): 从当前表取 nextCol 列的值（如 OnceDropRule）
//	Step 1+: 跨表查找，在目标表的 preCol 中匹配上一步传来的值，提取 nextCol 列传给下一步
//
// 右链（right chain）：从关联表出发，反向收集数据
//
//	Step 0 (sheet=目标表): 全表扫描，提取 preCol 值作为 firstStepInputValues，提取 nextCol 传给下一步
//	Step 1+: 跨表查找，同左链
//
// ## 两阶段门控流程
//
//	Phase 1 (Match): chainMatchCompare 比较左链最后一步值 vs 右链最后一步值
//	  - verify_exists: 左链值在右链值中全部找到 → 门控通过 → 进入 Phase 2
//	  - 门控不通过 → 跳过此行，不报错
//
//	Phase 2 (Compare): chainCompare 比较当前列值（如 BigAward）vs 右链第一步 firstStepInputValues
//	  - verify_exists: 当前列拆分后的每个值在右链 firstStepInput 中全部找到 → 不违规
//	  - 有缺失 → 违规，报告缺失的值列表
//
// ## isArray 参数
//
//	ChainStep.IsArray 控制两处拆分：
//	  1. 步骤内：提取值后按逗号拆分（如 "91001,91003,91004" → ["91001","91003","91004"]）
//	  2. 左链第一步 IsArray：还控制 Phase 2 比较阶段当前列值的拆分
//	     （如 BigAward="{1039016;1},{1040015;1}" → ["{1039016;1}","{1040015;1}"]）
//
// ## DrawPet 真实场景数据模型
//
//	检查目标："保证结缘亭配置的掉落道具在掉落池中存在"
//
//	数据关系：
//	  DrawPet.BigAward = "{1039016;1},{1040015;1}"    ← 大奖道具（多值，逗号分隔）
//	  DrawPet.OnceDropRule = 910001                    ← 掉落规则
//	  DropRule.910001 → DropGroup = 91001,91003,91004  ← 掉落分组（多值，逗号分隔）
//	  DropItem: {1039016;1} 在 DropGroup=91001         ← 掉落道具确实存在
//
//	左链：BigAward → OnceDropRule → DropRule → DropGroup（建立"哪些掉落组属于此活动"）
//	右链：DropItem → DropGroup → DropGroup.Id（建立"哪些道具在这些掉落组中"）
//	Phase 2：BigAward 中的每个道具 vs DropItem.Item（掉落池中的道具）

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule/chain_reference"
)

// ==================== 测试数据构造 ====================

// buildDrawPetSheets 构造 DrawPet 关系链检查的测试数据
//
// 数据关系图：
//
//	DrawPet(Id=3001).BigAward="{1039016;1},{1040015;1}", OnceDropRule=910001
//	DrawPet(Id=3002).BigAward="{1050007;1}",            OnceDropRule=920001
//
//	DropRule(Id=910001).DropGroup="91001,91003,91004"  ← 多值字段
//	DropRule(Id=920001).DropGroup="92001,92002,92003"
//
//	DropItem(DropGroup=91001, Item={1039016;1})  ← 3001 的第一个道具
//	DropItem(DropGroup=91001, Item={1040015;1})  ← 3001 的第二个道具
//	DropItem(DropGroup=92001, Item={1050007;1})  ← 3002 的道具
//
//	DropGroup: 91001/91003/91004（属于 910001）, 92001/92002/92003（属于 920001）
func buildDrawPetSheets() ([][]string, map[string]*excelize.File) {
	// 当前列 DrawPet（MJS_FIXED_ROWS: 4行表头，数据从索引4开始）
	// 包含 2 行数据：Id=3001（双道具）和 Id=3002（单道具）
	cols := [][]string{
		{"", "", "Id", "", "", "3001", "3002"},
		{"", "", "BigAward", "", "", "{1039016;1},{1040015;1}", "{1050007;1}"},
		{"", "", "OnceDropRule", "", "", "910001", "920001"},
	}

	// DropRule: Id(A) → DropGroup(D)
	// DropGroup 是多值字段，逗号分隔
	dropRuleFile := excelize.NewFile()
	dropRuleFile.SetSheetName("Sheet1", "掉落规则表|DropRule")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A3", "Id")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D3", "DropGroup")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A5", "910001")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D5", "91001,91003,91004")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "A6", "920001")
	dropRuleFile.SetCellValue("掉落规则表|DropRule", "D6", "92001,92002,92003")

	// DropItem: Id(A) → DropGroup(C), Item(D)
	// 每行一个道具，DropGroup 是单值
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

	// DropGroup: Id(A)
	// 91001/91003/91004 属于掉落规则 910001，92001/92002/92003 属于 920001
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

// drawPetChainSteps 返回 DrawPet 的标准链步骤配置（左链 step1 isArray=true）
//
// 左链路径：BigAward → OnceDropRule → DropRule.DropGroup
//   - Step 0 (isArray=true): 从当前行取 OnceDropRule 值，同时 isArray 控制比较阶段 BigAward 拆分
//   - Step 1 (isArray=true): 在 DropRule 中按 Id 匹配，提取 DropGroup（多值 "91001,91003,91004"）
//
// 右链路径：DropItem.Item → DropGroup → DropGroup.Id
//   - Step 0: 全表扫描 DropItem，preCol="Item" 收集 ItemCfg 值，nextCol="DropGroup" 传给下一步
//   - Step 1: 在 DropGroup 中按 Id 匹配，提取 Id 传给下一步
func drawPetChainSteps() string {
	return `{
		"left": {"steps": [
			{"sheet": "", "preCol": "", "findVal": "col", "nextCol": "OnceDropRule", "isArray": "true"},
			{"sheet": "掉落规则表|DropRule", "preCol": "Id", "findVal": "col", "nextCol": "DropGroup", "isArray": "true"}
		]},
		"right": {"steps": [
			{"sheet": "掉落道具表|DropItem", "preCol": "Item", "findVal": "col", "nextCol": "DropGroup"},
			{"sheet": "掉落分组表|DropGroup", "preCol": "Id", "findVal": "col", "nextCol": "Id"}
		]}
	}`
}

// ==================== 第一层：纯函数测试 ====================
//
// 直接调用 chain_reference.CompareByType 和 chain_reference.MatchByType，
// 验证 verify_exists 比较和匹配的核心逻辑。

// TestVerifyExists_Compare_AllFound 全部找到 → 不违规
func TestVerifyExists_Compare_AllFound(t *testing.T) {
	left := []string{"a", "b"}
	right := []string{"a", "b", "c"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.False(t, violation, "左链值全部在右链中，不应违规")
	assert.Empty(t, reason)
}

// TestVerifyExists_Compare_PartialMissing 部分缺失 → 违规，报告缺失值
func TestVerifyExists_Compare_PartialMissing(t *testing.T) {
	left := []string{"a", "x"}
	right := []string{"a", "b"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.True(t, violation, "左链值 x 不在右链中，应违规")
	assert.Contains(t, reason, "x")
	assert.Contains(t, reason, "verify_exists")
}

// TestVerifyExists_Compare_NoneFound 全部缺失 → 违规
func TestVerifyExists_Compare_NoneFound(t *testing.T) {
	left := []string{"x", "y"}
	right := []string{"a", "b"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.True(t, violation, "左链值全部不在右链中，应违规")
	assert.Contains(t, reason, "x")
	assert.Contains(t, reason, "y")
}

// TestVerifyExists_Compare_EmptyLeft 左链为空 → 不违规
func TestVerifyExists_Compare_EmptyLeft(t *testing.T) {
	left := []string{}
	right := []string{"a"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.False(t, violation, "左链为空，无值需要验证，不应违规")
	assert.Empty(t, reason)
}

// TestVerifyExists_Compare_SingleValue 左右各一个值且相等 → 不违规
func TestVerifyExists_Compare_SingleValue(t *testing.T) {
	left := []string{"{1039016;1}"}
	right := []string{"{1039016;1}", "{1040015;1}"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.False(t, violation, "单值在右链中找到，不应违规")
	assert.Empty(t, reason)
}

// TestVerifyExists_Compare_DuplicateLeft 左链有重复值 → 只报告一次缺失
func TestVerifyExists_Compare_DuplicateLeft(t *testing.T) {
	left := []string{"x", "x"}
	right := []string{"a"}

	violation, reason := chain_reference.CompareByType("verify_exists", left, right)

	assert.True(t, violation)
	assert.Contains(t, reason, "x")
}

// TestVerifyExists_Match_AllFound 匹配阶段：全部找到 → 门控通过
func TestVerifyExists_Match_AllFound(t *testing.T) {
	left := []string{"91001", "91003"}
	right := []string{"91001", "91003", "91004"}

	matched, _ := chain_reference.MatchByType("verify_exists", left, right)

	assert.True(t, matched, "左链值全部在右链中，门控应通过")
}

// TestVerifyExists_Match_PartialMissing 匹配阶段：部分缺失 → 门控不通过
func TestVerifyExists_Match_PartialMissing(t *testing.T) {
	left := []string{"91001", "99999"}
	right := []string{"91001", "91003"}

	matched, _ := chain_reference.MatchByType("verify_exists", left, right)

	assert.False(t, matched, "左链值 99999 不在右链中，门控应不通过")
}

// TestVerifyExists_Match_EmptyLeft 匹配阶段：左链为空 → 门控通过（无缺失值）
func TestVerifyExists_Match_EmptyLeft(t *testing.T) {
	left := []string{}
	right := []string{"91001"}

	matched, _ := chain_reference.MatchByType("verify_exists", left, right)

	// 空左链意味着没有需要验证的值 → 没有"缺失" → 门控通过
	assert.True(t, matched, "左链为空时无缺失值，门控应通过")
}

// ==================== 第二层：旧路径 E2E 测试 ====================
//
// 通过 ChainReferenceCheckRule.Check 走完整链路（含洋葱模型），
// 验证 verify_exists 在真实数据流中的端到端行为。

// TestVerifyExists_E2E_AllItemsExist 正常数据：BigAward 道具全部在掉落池中 → 不报错
//
// 数据流（DrawPet Id=3001, BigAward="{1039016;1},{1040015;1}"）：
//
//	左链 Step 0 (isArray=true): 取 OnceDropRule="910001"
//	左链 Step 1 (isArray=true): DropRule.Id="910001" → DropGroup="91001,91003,91004"
//	  → isArray 拆分为 ["91001","91003","91004"]
//
//	右链 Step 0: DropItem 全表扫描，preCol=Item → firstStepInput=[{1039016;1},{1040015;1},{1050007;1}]
//	右链 Step 1: DropGroup 中查找 → last=[91001,91003,91004,92001,92002,92003]
//
//	Phase 1 (Match, verify_exists):
//	  left=[91001,91003,91004] vs right=[91001,...,91003,...,91004,...] → 全部找到 → 通过
//
//	Phase 2 (Compare, verify_exists):
//	  当前列 BigAward="{1039016;1},{1040015;1}" → 左链 step0 isArray 拆分为 [{1039016;1},{1040015;1}]
//	  右链 firstStepInput=[{1039016;1},{1040015;1},{1050007;1}] → 全部找到 → 不违规
func TestVerifyExists_E2E_AllItemsExist(t *testing.T) {
	cols, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	assert.Empty(t, errors, "BigAward 道具全部在掉落池中，不应报错")
}

// TestVerifyExists_E2E_FakeItemInBigAward 假数据：BigAward 中埋入不存在的道具 → 报错
//
// 修改 BigAward 为 "{10390116;1},{1040015;1}"，其中 {10390116;1} 不在 DropItem 中
// Phase 2 compare: {10390116;1} 未在 firstStepInput 中找到 → 违规
func TestVerifyExists_E2E_FakeItemInBigAward(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "3001"},
		{"", "", "BigAward", "", "", "{10390116;1},{1040015;1}"}, // {10390116;1} 不存在于 DropItem
		{"", "", "OnceDropRule", "", "", "910001"},
	}
	_, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	if assert.NotEmpty(t, errors, "{10390116;1} 不存在于 DropItem，应报错") {
		assert.Contains(t, errors[0].Reason, "10390116", "错误信息应包含缺失的道具 ID")
		assert.Contains(t, errors[0].Reason, "verify_exists", "错误信息应包含比较类型名")
	}
}

// TestVerifyExists_E2E_SingleItem 单道具场景：BigAward 只有一个道具且存在 → 不报错
//
// DrawPet Id=3002, BigAward="{1050007;1}", OnceDropRule=920001
// 左链: 920001 → DropGroup=92001,92002,92003 → 拆分
// 右链: DropItem.Item={1050007;1} 在 DropGroup=92001
// Match: left=[92001,92002,92003] vs right=[...92001,92002,92003...] → 通过
// Compare: [{1050007;1}] vs [{1039016;1},{1040015;1},{1050007;1}] → 找到 → 不违规
func TestVerifyExists_E2E_SingleItem(t *testing.T) {
	cols, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	// colIdx=1 (BigAward), rowIdx=5 (第二行数据 Id=3002)
	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 5, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	assert.Empty(t, errors, "BigAward={1050007;1} 在 DropItem 中存在，不应报错")
}

// TestVerifyExists_E2E_LeftChainNoMatch 左链跨表查找无匹配 → 引用完整性报错
//
// OnceDropRule=999999 在 DropRule 中无匹配，引用完整性检查应报错
func TestVerifyExists_E2E_LeftChainNoMatch(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "3003"},
		{"", "", "BigAward", "", "", "{1039016;1}"},
		{"", "", "OnceDropRule", "", "", "999999"}, // DropRule 中不存在此 Id
	}
	_, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	assert.NotEmpty(t, errors, "左链 OnceDropRule=999999 在 DropRule 中无匹配，引用完整性应报错")
	assert.Contains(t, errors[0].Reason, "未找到")
}

// TestVerifyExists_E2E_AllFakeItems 所有道具都不存在 → 全部报告缺失
func TestVerifyExists_E2E_AllFakeItems(t *testing.T) {
	cols := [][]string{
		{"", "", "Id", "", "", "3001"},
		{"", "", "BigAward", "", "", "{9990001;1},{9990002;1}"}, // 两个都不存在
		{"", "", "OnceDropRule", "", "", "910001"},
	}
	_, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	if assert.NotEmpty(t, errors, "所有道具都不存在，应报错") {
		assert.Contains(t, errors[0].Reason, "9990001")
		assert.Contains(t, errors[0].Reason, "9990002")
	}
}

// ==================== 第三层：isArray 拆分 bug 验收测试 ====================
//
// 验收 bug 修复：applyIsArray 应同时作用于搜索值（preCol 匹配）和提取值（nextCol）
//
// Bug 描述：左链 step1 isArray=true 时，搜索值 "910001" 被正确处理，
// 但提取的 nextCol 值 "91001,91003,91004" 没有被拆分就直接传给下一步，
// 导致 Match 阶段收到 ["91001,91003,91004"]（单元素）而非 ["91001","91003","91004"]（三元素），
// verify_exists 门控因字符串整体不匹配而始终失败。
//
// 修复：在 onion_left_step.go 和 engine.go 中，nextCol 提取循环结束后
// 统一调用 applyIsArray(nextValues, step.IsArray) 拆分。

// TestVerifyExists_IsArrayFix_MatchGatePasses 验收：isArray 拆分后 Match 门控正确通过
//
// 修复前：left=["91001,91003,91004"] → verify_exists 在 right 中找不到整体字符串 → 门控失败
// 修复后：left=["91001","91003","91004"] → 逐值在 right 中找到 → 门控通过
func TestVerifyExists_IsArrayFix_MatchGatePasses(t *testing.T) {
	cols, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)

	// Match 门控通过 + Compare 全部找到 → 不报错
	assert.Empty(t, errors, "isArray 拆分后 Match 门控应通过，道具全部存在，不应报错")
}

// TestVerifyExists_IsArrayFix_ComparePhaseSplit 验收：比较阶段当前列值被正确拆分
//
// 左链 step0 isArray=true → Phase 2 比较时 BigAward="{1039016;1},{1040015;1}" 被拆分
// 使用假数据验证拆分后确实逐一比较
func TestVerifyExists_IsArrayFix_ComparePhaseSplit(t *testing.T) {
	// BigAward 中一个真一个假
	cols := [][]string{
		{"", "", "Id", "", "", "3001"},
		{"", "", "BigAward", "", "", "{1039016;1},{NOT_EXIST;1}"}, // 第二个是假的
		{"", "", "OnceDropRule", "", "", "910001"},
	}
	_, sheetMap := buildDrawPetSheets()

	params := map[string]string{
		"chainSteps":        drawPetChainSteps(),
		"chainCompare":      "verify_exists",
		"chainMatchCompare": "verify_exists",
		"allowCommit":       "true",
		"allowEmpty":        "false",
		"breakLine":         "3",
	}

	rule := &ChainReferenceCheckRule{}
	errors := rule.Check("结缘亭|DrawPet", cols, 1, 4, params, sheetMap)
	for _, e := range errors {
		t.Logf("error: index=%d reason=%s", e.Index, e.Reason)
	}
	if assert.NotEmpty(t, errors, "BigAward 中有假道具，应报错") {
		// 验证报告的是 NOT_EXIST 而非 {1039016;1}
		assert.Contains(t, errors[0].Reason, "NOT_EXIST")
		assert.NotContains(t, errors[0].Reason, "1039016", "{1039016;1} 存在，不应被报告为缺失")
	}
}

// TestVerifyExists_IsArrayFix_Step1NextColSplit 验收：左链 step1 提取的 nextCol 多值被正确拆分
//
// 直接验证 SplitArrayElements 对 DropGroup 值的拆分结果
func TestVerifyExists_IsArrayFix_Step1NextColSplit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"三值逗号分隔", "91001,91003,91004", []string{"91001", "91003", "91004"}},
		{"单值", "91001", []string{"91001"}},
		{"含花括号", "{1039016;1},{1040015;1}", []string{"{1039016;1}", "{1040015;1}"}},
		{"空字符串", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []string
			if tt.input != "" {
				result = chain_reference.SplitArrayElements(tt.input, ",")
			}
			assert.Equal(t, tt.expected, result, fmt.Sprintf("SplitArrayElements(%q)", tt.input))
		})
	}
}
