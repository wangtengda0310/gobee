// Package check_manager 提供规则过滤功能的单元测试
package engine

import (
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// TestFilterRulesByChangedFiles_HeroOnly 测试只变更 Hero 表
// 业务场景：只修改了 Hero.xlsx，应该只执行 Hero 相关的规则
func TestFilterRulesByChangedFiles_HeroOnly(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			Rules: map[string]*json_rule.SheetColRule{
				"Id": {PropName: "Id"},
			},
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.ARENA_SEASON_CHECK, Enabled: true},
			},
		},
	}

	changedFiles := []string{"Hero.xlsx"}
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	if len(filtered) != 1 {
		t.Errorf("期望 1 条规则，实际 %d", len(filtered))
	}
	if filtered[0].Sheet != "武将|Hero" {
		t.Errorf("期望 Hero 规则，实际 %s", filtered[0].Sheet)
	}
}

// TestFilterRulesByChangedFiles_EmptyChanges 测试无变更文件时返回全部规则
// 业务场景：没有提供变更文件列表，应该执行所有规则（全量检查）
func TestFilterRulesByChangedFiles_EmptyChanges(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{Sheet: "武将|Hero"},
		{Sheet: "竞技场赛季|ArenaSeason"},
	}

	changedFiles := []string{}
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	if len(filtered) != 2 {
		t.Errorf("期望 2 条规则（全量检查），实际 %d", len(filtered))
	}
}

// TestFilterRulesByChangedFiles_GeneralRuleAlwaysRun 测试通用规则始终执行
// 业务场景：即使没有变更文件，通用规则（NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）也应该执行
func TestFilterRulesByChangedFiles_GeneralRuleAlwaysRun(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.NEW_ROW_NOTIFY, Enabled: true},
			},
		},
	}

	changedFiles := []string{"Hero.xlsx"} // 变更的是 Hero，而不是 ArenaSeason
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	// NEW_ROW_NOTIFY 是通用规则，始终应该执行
	if len(filtered) != 1 {
		t.Errorf("期望 1 条规则（通用规则始终执行），实际 %d", len(filtered))
	}
}

// TestIsSheetInChangedFiles_SuffixMatch 测试后缀匹配
// 业务场景：配表名格式为 "中文|英文"，变更文件只有英文部分时也应该匹配
func TestIsSheetInChangedFiles_SuffixMatch(t *testing.T) {
	cases := []struct {
		sheetName    string
		changedFiles []string
		expected     bool
	}{
		{"武将|Hero", []string{"Hero.xlsx"}, true},
		{"Hero", []string{"Hero.xlsx"}, true},
		{"竞技场赛季|ArenaSeason", []string{"Hero.xlsx"}, false},
		{"武将|Hero", []string{"config/excel/Hero.xlsx"}, true},
	}

	for _, tc := range cases {
		result := isSheetInChangedFiles(tc.sheetName, tc.changedFiles)
		if result != tc.expected {
			t.Errorf("isSheetInChangedFiles(%q, %v) = %v, 期望 %v",
				tc.sheetName, tc.changedFiles, result, tc.expected)
		}
	}
}

// TestFilterRulesByChangedFiles_MultipleChanges 测试多个文件变更
// 业务场景：同时修改了 Hero.xlsx 和 ArenaSeason.xlsx，应该执行两个表的规则
func TestFilterRulesByChangedFiles_MultipleChanges(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			Rules: map[string]*json_rule.SheetColRule{
				"Id": {PropName: "Id"},
			},
		},
		{
			Sheet: "竞技场赛季|ArenaSeason",
			Rules: map[string]*json_rule.SheetColRule{
				"SeasonEndTime": {PropName: "SeasonEndTime"},
			},
		},
		{
			Sheet: "卡牌|Card",
			Rules: map[string]*json_rule.SheetColRule{
				"Id": {PropName: "Id"},
			},
		},
	}

	changedFiles := []string{"Hero.xlsx", "ArenaSeason.xlsx"}
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	if len(filtered) != 2 {
		t.Errorf("期望 2 条规则，实际 %d", len(filtered))
	}
}

// TestIsGeneralRule 测试通用规则判断
// 业务场景：NEW_ROW_NOTIFY 和 ROW_CHANGE_NOTIFY 应该被识别为通用规则
func TestIsGeneralRule(t *testing.T) {
	tests := []struct {
		ruleType json_rule.ETableRule
		expected bool
	}{
		{json_rule.NEW_ROW_NOTIFY, true},
		{json_rule.ROW_CHANGE_NOTIFY, true},
		{json_rule.HERO_DROP_CHECK, false},
		{json_rule.ARENA_SEASON_CHECK, false},
	}

	for _, tt := range tests {
		result := isGeneralRule(tt.ruleType)
		if result != tt.expected {
			t.Errorf("isGeneralRule(%v) = %v, 期望 %v", tt.ruleType, result, tt.expected)
		}
	}
}

// TestIsSheetInChangedFiles_CrossTableDependency 测试跨表依赖匹配
// 业务场景：HERO_DROP_CHECK 的 TargetSheets 为 ["Hero"]，Hero.xlsx 变更时应匹配
func TestIsSheetInChangedFiles_CrossTableDependency(t *testing.T) {
	// 模拟 Hero.xlsx 变更
	changedFiles := []string{"excel/Hero.xlsx"}

	// HERO_DROP_CHECK 的 TargetSheets 是 "Hero"
	if !isSheetInChangedFiles("Hero", changedFiles) {
		t.Error("Hero TargetSheets 应匹配 Hero.xlsx")
	}

	// SEASON_PASS_HERO_OPEN_CHECK 的 TargetSheets 是 "SeasonPassReward"
	if isSheetInChangedFiles("SeasonPassReward", changedFiles) {
		t.Error("SeasonPassReward TargetSheets 不应匹配 Hero.xlsx")
	}
}

// TestFilterRulesByChangedFiles_SpaceSeparatedFiles 测试空格分隔的变更文件列表
// 业务场景：CI/CD 可能传入空格分隔的文件列表而非 \n 分隔（main.go 已修复分割逻辑）
func TestFilterRulesByChangedFiles_SpaceSeparatedFiles(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			Rules: map[string]*json_rule.SheetColRule{
				"Id": {PropName: "Id"},
			},
		},
		{
			Sheet: "竞技场赛季|ArenaSeason",
			Rules: map[string]*json_rule.SheetColRule{
				"SeasonEndTime": {PropName: "SeasonEndTime"},
			},
		},
	}

	// 注意：main.go 层面已修复分割逻辑，FilterRulesByChangedFiles 接收的应该是已分割好的 []string
	changedFiles := []string{"Hero.xlsx", "ArenaSeason.xlsx"}
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	if len(filtered) != 2 {
		t.Errorf("期望 2 条规则，实际 %d", len(filtered))
	}
}

// TestFilterRulesByChangedFiles_SupplementedHeroWithCrossTableRules
// 测试 supplementDefaultTableRules 补充的 Hero 规则（含跨表依赖）能正确匹配
// 业务场景：Hero.xlsx 变更时，通过 DefaultTableRules 补充的规则（含战令/大将军）应被匹配
func TestFilterRulesByChangedFiles_SupplementedHeroWithCrossTableRules(t *testing.T) {
	// 模拟 supplementDefaultTableRules 为 Hero 补充的规则结构
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			Rules: map[string]*json_rule.SheetColRule{}, // 空 map（无列级规则）
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
				// HERO_SYNTHESIS_CHECK 已迁移到 Item 表
				{Type: json_rule.HERO_MELT_CHECK, Enabled: true},
				{Type: json_rule.SEASON_PASS_HERO_OPEN_CHECK, Enabled: true},
				{Type: json_rule.ARENA_GENERAL_HERO_OPEN_CHECK, Enabled: true},
			},
		},
	}

	changedFiles := []string{"excel/Hero.xlsx"}
	filtered := FilterRulesByChangedFiles(rules, changedFiles)

	if len(filtered) != 1 {
		t.Errorf("期望 1 条规则匹配（Hero supplemented），实际 %d", len(filtered))
	}
	if len(filtered) > 0 && len(filtered[0].TableRules) != 4 {
		t.Errorf("期望 4 个表级规则（合成检查已迁移到 Item），实际 %d", len(filtered[0].TableRules))
	}
}

// ==================== 通用规则决策测试 ====================

// TestShouldRunGeneralRules_Incremental 增量模式应执行通用规则
func TestShouldRunGeneralRules_Incremental(t *testing.T) {
	assert.True(t, shouldRunGeneralRules([]string{"Hero.xlsx"}),
		"增量模式（changedFiles 非空）应执行通用规则")
	assert.True(t, shouldRunGeneralRules([]string{"Hero.xlsx", "Item.xlsx"}),
		"增量模式（多个 changedFiles）应执行通用规则")
}

// TestShouldRunGeneralRules_Full 全量模式应跳过通用规则
func TestShouldRunGeneralRules_Full(t *testing.T) {
	assert.False(t, shouldRunGeneralRules(nil),
		"全量模式（changedFiles=nil）应跳过通用规则")
	assert.False(t, shouldRunGeneralRules([]string{}),
		"全量模式（changedFiles 为空切片）应跳过通用规则")
}

// ==================== 默认参数补充测试 ====================

// TestSupplementDefaultParams_FillMissingKeys 测试为空 Params 填充默认值
// 业务场景：ArenaSeason 有 JSON 配置但 Params 为空，应从 ParamDefs 获取默认值
func TestSupplementDefaultParams_FillMissingKeys(t *testing.T) {
	// 构造一个模拟规则：TableRule 有 Params 但为空
	rules := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.ARENA_SEASON_CHECK,
					Enabled: true,
					Params:  map[string]string{}, // 空 Params
				},
			},
		},
	}

	SupplementDefaultParams(rules)

	// 验证默认参数被填充
	params := rules[0].TableRules[0].Params
	assert.Equal(t, "SeasonEndTime", params[string(json_rule.SEASON_END_TIME_COL)],
		"应填充 seasonEndTimeCol 默认值")
	assert.Equal(t, "168h", params[string(json_rule.TIME_RANGE_BEFORE)],
		"应填充 timeRangeBefore 默认值")
}

// TestSupplementDefaultParams_NoOverrideExisting 测试不覆盖已有参数
// 业务场景：用户在 JSON 中自定义了参数值，不应被默认值覆盖
func TestSupplementDefaultParams_NoOverrideExisting(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.ARENA_SEASON_CHECK,
					Enabled: true,
					Params: map[string]string{
						string(json_rule.TIME_RANGE_BEFORE): "72h", // 用户自定义值
					},
				},
			},
		},
	}

	SupplementDefaultParams(rules)

	params := rules[0].TableRules[0].Params
	assert.Equal(t, "72h", params[string(json_rule.TIME_RANGE_BEFORE)],
		"不应覆盖用户已有的参数值")
	assert.Equal(t, "SeasonEndTime", params[string(json_rule.SEASON_END_TIME_COL)],
		"应填充缺失的参数默认值")
}

// TestSupplementDefaultParams_SkipNoParamDefs 测试跳过无 ParamDefs 的规则
// 业务场景：某些规则（如 DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK）没有 ParamDefs，不应被修改
func TestSupplementDefaultParams_SkipNoParamDefs(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "活动|Activity",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK,
					Enabled: true,
					Params:  map[string]string{},
				},
			},
		},
	}

	SupplementDefaultParams(rules)

	// DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK 没有 ParamDefs，Params 应保持为空
	params := rules[0].TableRules[0].Params
	assert.Empty(t, params, "无 ParamDefs 的规则不应添加任何参数")
}

// TestSupplementDefaultParams_SkipDisabled 测试跳过已禁用的规则
// 业务场景：用户禁用了某个规则，不应修改其参数
func TestSupplementDefaultParams_SkipDisabled(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.ARENA_SEASON_CHECK,
					Enabled: false, // 已禁用
					Params:  nil,
				},
			},
		},
	}

	SupplementDefaultParams(rules)

	// 已禁用的规则 Params 应保持为 nil
	assert.Nil(t, rules[0].TableRules[0].Params, "已禁用的规则不应被修改")
}

// TestSupplementDefaultParams_NilParams 测试 Params 为 nil 的情况
// 业务场景：JSON 配置中完全没有 params 字段
func TestSupplementDefaultParams_NilParams(t *testing.T) {
	rules := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.ARENA_SEASON_CHECK,
					Enabled: true,
					Params:  nil, // nil Params
				},
			},
		},
	}

	SupplementDefaultParams(rules)

	params := rules[0].TableRules[0].Params
	assert.NotNil(t, params, "应初始化 Params map")
	assert.Equal(t, "168h", params[string(json_rule.TIME_RANGE_BEFORE)],
		"应填充默认参数值")
}

// ==================== 深拷贝测试 ====================

// TestDeepCopyRules_Independence 验证深拷贝后修改副本不影响原始
// 业务场景：merge 场景中多个 commit 共享 allRules，深拷贝确保各 commit 独立
func TestDeepCopyRules_Independence(t *testing.T) {
	original := []*json_rule.SheetRule{
		{
			Sheet: "竞技场赛季|ArenaSeason",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.ARENA_SEASON_CHECK,
					Enabled: true,
					Params:  map[string]string{"timeRangeBefore": "72h"},
				},
			},
		},
	}

	copied := deepCopyRules(original)

	// 修改副本的 Params
	copied[0].TableRules[0].Params["timeRangeBefore"] = "999h"

	// 验证原始不受影响
	assert.Equal(t, "72h", original[0].TableRules[0].Params["timeRangeBefore"],
		"修改副本不应影响原始规则")

	// 验证副本确实被修改
	assert.Equal(t, "999h", copied[0].TableRules[0].Params["timeRangeBefore"],
		"副本应被正确修改")
}

// TestDeepCopyRules_NilParams 验证 Params 为 nil 时深拷贝正确处理
func TestDeepCopyRules_NilParams(t *testing.T) {
	original := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{
					Type:    json_rule.HERO_DROP_CHECK,
					Enabled: true,
					Params:  nil,
				},
			},
		},
	}

	copied := deepCopyRules(original)

	// Params 应保持 nil（不应预创建 map）
	assert.Nil(t, copied[0].TableRules[0].Params, "nil Params 深拷贝后应保持 nil")

	// supplementDefaultParams 给副本创建 Params 不影响原始
	SupplementDefaultParams(copied)
	assert.Nil(t, original[0].TableRules[0].Params, "原始规则的 Params 应保持 nil")
}

// TestDeepCopyRules_EmptySlice 验证空切片和 nil 输入
func TestDeepCopyRules_EmptySlice(t *testing.T) {
	// 空切片
	empty := deepCopyRules([]*json_rule.SheetRule{})
	assert.Empty(t, empty)

	// nil 输入会 panic（len(nil)=0，make 长度为 0），验证不会 panic
	result := deepCopyRules(nil)
	assert.Empty(t, result)
}

// ==================== filterSheetMapByRules 测试 ====================

// TestFilterSheetMapByRules_Basic 测试基本过滤功能
// 业务场景：前端传入只包含 Hero 的规则，sheetMap 应保留 Hero 及其 RequiredSheets
// HERO_DROP_CHECK 的 RequiredSheets 包含 DropItem、SeasonPass、ArenaSeason 等
// 后缀匹配："竞技场赛季|ArenaSeason" 匹配 "ArenaSeason"，"掉落道具|DropItem" 匹配 "DropItem"
func TestFilterSheetMapByRules_Basic(t *testing.T) {
	// 构造模拟 sheetMap（3 个表，使用中文|英文格式）
	sheetMap := map[string]*excelize.File{
		"武将|Hero":           nil,
		"竞技场赛季|ArenaSeason": nil,
		"掉落道具|DropItem":     nil,
	}

	// 只传入 Hero 的规则
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
	}

	filtered := filterSheetMapByRules(sheetMap, rules)

	// Hero 本身 + RequiredSheets(ArenaSeason, DropItem) 通过后缀匹配都被保留
	assert.Len(t, filtered, 3, "应保留 Hero 及其 RequiredSheets")
	assert.Contains(t, filtered, "武将|Hero")
	assert.Contains(t, filtered, "竞技场赛季|ArenaSeason")
	assert.Contains(t, filtered, "掉落道具|DropItem")
}

// TestFilterSheetMapByRules_WithRequiredSheets 测试 RequiredSheets 关联表保留
// 业务场景：Hero 规则需要 DropItem 表做跨表检查，过滤时应保留
func TestFilterSheetMapByRules_WithRequiredSheets(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero":          nil,
		"DropItem":         nil,
		"SeasonPassReward": nil,
		"Item":             nil,
	}

	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
	}

	filtered := filterSheetMapByRules(sheetMap, rules)

	// HERO_DROP_CHECK 的 RequiredSheets 包含 DropItem/SeasonPassReward，应被保留
	assert.Contains(t, filtered, "武将|Hero")
	assert.Contains(t, filtered, "DropItem")
	assert.Contains(t, filtered, "SeasonPassReward")
	assert.NotContains(t, filtered, "Item")
}

// TestFilterSheetMapByRules_WithCrossReference 测试列级跨表引用参数
// 业务场景：列级 FOREIGN_KEY 规则引用了其他表，过滤时应保留
func TestFilterSheetMapByRules_WithCrossReference(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"活动任务|ActivityTask": nil,
		"道具|Item":           nil,
		"武将|Hero":           nil,
	}

	rules := []*json_rule.SheetRule{
		{
			Sheet: "活动任务|ActivityTask",
			Rules: map[string]*json_rule.SheetColRule{
				"RewardId": {
					PropName: "RewardId",
					PropRules: []*json_rule.ColRule{
						{
							Type: json_rule.FOREIGN_KEY,
							Params: map[string]string{
								"targetSheet": "道具|Item",
							},
						},
					},
				},
			},
		},
	}

	filtered := filterSheetMapByRules(sheetMap, rules)

	// 应保留 ActivityTask（规则本身）和 Item（targetSheet 引用）
	assert.Contains(t, filtered, "活动任务|ActivityTask")
	assert.Contains(t, filtered, "道具|Item")
	assert.NotContains(t, filtered, "武将|Hero")
}

// TestFilterSheetMapByRules_EmptyRules 测试空规则列表
// 业务场景：前端传入空规则列表，应返回空 sheetMap
func TestFilterSheetMapByRules_EmptyRules(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero": nil,
	}

	filtered := filterSheetMapByRules(sheetMap, []*json_rule.SheetRule{})

	assert.Empty(t, filtered, "空规则列表应返回空 sheetMap")
}

// TestFilterSheetMapByRules_DisabledRule 测试禁用规则的 RequiredSheets 不被收集
// 业务场景：禁用的规则不应影响过滤结果
func TestFilterSheetMapByRules_DisabledRule(t *testing.T) {
	sheetMap := map[string]*excelize.File{
		"武将|Hero":       nil,
		"掉落道具|DropItem": nil,
	}

	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: false}, // 禁用
			},
		},
	}

	filtered := filterSheetMapByRules(sheetMap, rules)

	// 只保留 Hero（禁用的 HERO_DROP_CHECK 的 RequiredSheets 不应被收集）
	assert.Len(t, filtered, 1)
	assert.Contains(t, filtered, "武将|Hero")
	assert.NotContains(t, filtered, "掉落道具|DropItem")
}

// ==================== supplementDefaultTableRules 测试 ====================

// TestSupplementDefaultTableRules_LoadedSheetMissingDefaultRules 场景A：已加载但缺少默认表级规则的 Sheet
// 业务场景：前端传入了 "定向招募表|DrawFix" 的 SheetRule，但 tableRules 为空
// 预期：supplementDefaultTableRules 应该补充 DRAWFIX_PROTECTION_CHECK 规则
func TestSupplementDefaultTableRules_LoadedSheetMissingDefaultRules(t *testing.T) {
	// 1. 构建 sheetMap（包含 DrawFix 表）
	sheetMap := map[string]*excelize.File{
		"定向招募表|DrawFix": nil,
	}

	// 2. 构建 rules：前端传入了 DrawFix 的 SheetRule，但 TableRules 为空
	rules := []*json_rule.SheetRule{
		{
			Sheet:       "定向招募表|DrawFix",
			ManagerList: nil,
			Rules:       map[string]*json_rule.SheetColRule{},
			TableRules:  []*json_rule.TableRule{}, // 空表级规则
		},
	}

	// 3. 调用 supplementDefaultTableRules
	result := supplementDefaultTableRules(rules, sheetMap)

	// 4. 验证：应只有 1 条 SheetRule，且 DrawFix 的 TableRules 中新增了 DRAWFIX_PROTECTION_CHECK
	assert.Len(t, result, 1, "不应新增 SheetRule，应在原有规则上补充")

	drawFixRule := result[0]
	assert.Equal(t, "定向招募表|DrawFix", drawFixRule.Sheet)
	assert.Len(t, drawFixRule.TableRules, 2, "应补充 2 条默认表级规则")
	assert.Equal(t, json_rule.DRAWFIX_PROTECTION_CHECK, drawFixRule.TableRules[0].Type,
		"应补充 DRAWFIX_PROTECTION_CHECK 规则")
	assert.True(t, drawFixRule.TableRules[0].Enabled, "补充的规则应默认启用")
	assert.NotEmpty(t, drawFixRule.TableRules[0].DisplayName, "补充的规则应有显示名称")
}

// TestSupplementDefaultTableRules_LoadedSheetWithExistingRules 场景B：已加载且已有默认表级规则的 Sheet
// 业务场景：前端传入了 "武将|Hero" 的 SheetRule，且已包含 HERO_DROP_CHECK
// 预期：supplementDefaultTableRules 不应重复补充 HERO_DROP_CHECK，但应补充其他缺失的默认规则
func TestSupplementDefaultTableRules_LoadedSheetWithExistingRules(t *testing.T) {
	// 1. 构建 sheetMap（包含 Hero 表）
	sheetMap := map[string]*excelize.File{
		"武将|Hero": nil,
	}

	// 2. 构建 rules：Hero 已有 HERO_DROP_CHECK，但缺少 HERO_MELT_CHECK 等
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
	}

	// 3. 调用 supplementDefaultTableRules
	result := supplementDefaultTableRules(rules, sheetMap)

	// 4. 验证：Hero 的 TableRules 应包含所有默认规则，不重复
	heroRule := result[0]
	// Hero 的默认规则包括：HERO_DROP_CHECK, HERO_MELT_CHECK, SEASON_PASS_HERO_OPEN_CHECK, ARENA_GENERAL_HERO_OPEN_CHECK, HERO_SKILL_BUFF_CHECK, HERO_ISOPEN_OPENDATE_CHECK
	expectedTypes := []json_rule.ETableRule{
		json_rule.HERO_DROP_CHECK,
		json_rule.HERO_MELT_CHECK,
		json_rule.SEASON_PASS_HERO_OPEN_CHECK,
		json_rule.ARENA_GENERAL_HERO_OPEN_CHECK,
		json_rule.HERO_SKILL_BUFF_CHECK,
		json_rule.HERO_ISOPEN_OPENDATE_CHECK,
	}

	assert.Len(t, heroRule.TableRules, len(expectedTypes), "应补充缺失的默认规则，不重复已有规则")

	// 验证每种规则只出现一次
	existingTypes := make(map[json_rule.ETableRule]int)
	for _, tr := range heroRule.TableRules {
		existingTypes[tr.Type]++
	}
	for _, expected := range expectedTypes {
		assert.Equal(t, 1, existingTypes[expected], "规则 %s 应只出现一次", expected)
	}
}

// TestSupplementDefaultTableRules_UnloadedSheetWithDefaultRules 场景C：未加载但有默认表级规则的 Sheet
// 业务场景：sheetMap 中包含 "竞技场赛季|ArenaSeason"，但 rules 中没有对应的 SheetRule
// 预期：supplementDefaultTableRules 应创建新的 SheetRule 并补充 ARENA_SEASON_CHECK 规则
func TestSupplementDefaultTableRules_UnloadedSheetWithDefaultRules(t *testing.T) {
	// 1. 构建 sheetMap（包含 ArenaSeason 表，但 rules 中没有）
	sheetMap := map[string]*excelize.File{
		"竞技场赛季|ArenaSeason": nil,
	}

	// 2. 构建 rules：不包含 ArenaSeason 的 SheetRule
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
	}

	// 3. 调用 supplementDefaultTableRules
	result := supplementDefaultTableRules(rules, sheetMap)

	// 4. 验证：应新增 1 条 SheetRule（ArenaSeason），共 2 条
	assert.Len(t, result, 2, "应为未加载的 ArenaSeason 创建新的 SheetRule")

	// 找到 ArenaSeason 规则
	var arenaSeasonRule *json_rule.SheetRule
	for _, r := range result {
		if r.Sheet == "竞技场赛季|ArenaSeason" {
			arenaSeasonRule = r
			break
		}
	}
	assert.NotNil(t, arenaSeasonRule, "结果中应包含 ArenaSeason 的 SheetRule")
	assert.Equal(t, "竞技场赛季|ArenaSeason", arenaSeasonRule.Sheet)
	assert.NotNil(t, arenaSeasonRule.Rules, "新创建的 SheetRule 应有初始化的 Rules map")
	assert.Len(t, arenaSeasonRule.TableRules, 1, "应补充 1 条默认表级规则")
	assert.Equal(t, json_rule.ARENA_SEASON_CHECK, arenaSeasonRule.TableRules[0].Type,
		"应补充 ARENA_SEASON_CHECK 规则")
	assert.True(t, arenaSeasonRule.TableRules[0].Enabled, "补充的规则应默认启用")
}

// TestSupplementDefaultTableRules_SheetWithoutDefaultRules 无默认规则的 Sheet 不应被处理
// 业务场景：sheetMap 中包含 "卡牌|Card"，但 DefaultTableRules 中没有 Card 的配置
// 预期：supplementDefaultTableRules 不应为 Card 创建或修改任何规则
func TestSupplementDefaultTableRules_SheetWithoutDefaultRules(t *testing.T) {
	// 1. 构建 sheetMap（包含 Card 表，无默认规则）
	sheetMap := map[string]*excelize.File{
		"卡牌|Card": nil,
	}

	// 2. 构建 rules：不包含 Card
	rules := []*json_rule.SheetRule{
		{
			Sheet: "武将|Hero",
			TableRules: []*json_rule.TableRule{
				{Type: json_rule.HERO_DROP_CHECK, Enabled: true},
			},
		},
	}

	// 3. 调用 supplementDefaultTableRules
	result := supplementDefaultTableRules(rules, sheetMap)

	// 4. 验证：不应为 Card 创建新规则，结果仍只有 1 条
	assert.Len(t, result, 1, "无默认规则的 Sheet 不应被补充")
	assert.Equal(t, "武将|Hero", result[0].Sheet)
}

// TestFilterSheetMapByRules_DefaultRuleRequiredSheets 测试默认规则的 RequiredSheets 被正确加载
// 业务场景：右键点击 SeasonPassReward.xlsx 执行分类，此时 list 中只有 SeasonPassReward 的 SheetRule
// 但 SeasonPassReward 有默认规则 SEASON_PASS_HERO_OPEN_CHECK，需要 Hero 和 SeasonPass 表
// 预期：filterSheetMapByRules 应保留 SeasonPassReward、Hero、SeasonPass
func TestFilterSheetMapByRules_DefaultRuleRequiredSheets(t *testing.T) {
	// 构造模拟 sheetMap（包含 SeasonPassReward 及其依赖表）
	sheetMap := map[string]*excelize.File{
		"赛季战令奖励表|SeasonPassReward": nil,
		"武将|Hero":                  nil,
		"赛季战令|SeasonPass":          nil,
		"道具|Item":                  nil,
	}

	// 模拟右键分类场景：只传入 SeasonPassReward 的规则（无 TableRules，默认规则尚未补充）
	rules := []*json_rule.SheetRule{
		{
			Sheet:      "赛季战令奖励表|SeasonPassReward",
			TableRules: []*json_rule.TableRule{}, // 空，默认规则尚未补充
		},
	}

	filtered := filterSheetMapByRules(sheetMap, rules)

	// 此时 rules 中还没有默认规则，filterSheetMapByRules 只基于传入的规则过滤
	// 所以只保留 SeasonPassReward 本身
	assert.Len(t, filtered, 1, "传入规则无 TableRules，应只保留 SeasonPassReward")
	assert.Contains(t, filtered, "赛季战令奖励表|SeasonPassReward")
	assert.NotContains(t, filtered, "武将|Hero")
	assert.NotContains(t, filtered, "赛季战令|SeasonPass")
	assert.NotContains(t, filtered, "道具|Item")

	// 补充默认规则后再过滤
	rulesWithDefault := supplementDefaultTableRules(rules, filtered)
	filteredWithDefault := filterSheetMapByRules(sheetMap, rulesWithDefault)

	// 补充默认规则后，SEASON_PASS_HERO_OPEN_CHECK 的 RequiredSheets 包含 Hero 和 SeasonPass
	assert.Len(t, filteredWithDefault, 3, "补充默认规则后应保留 SeasonPassReward + Hero + SeasonPass")
	assert.Contains(t, filteredWithDefault, "赛季战令奖励表|SeasonPassReward")
	assert.Contains(t, filteredWithDefault, "武将|Hero")
	assert.Contains(t, filteredWithDefault, "赛季战令|SeasonPass")
	assert.NotContains(t, filteredWithDefault, "道具|Item")
}
