package functiontest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =====================================================================
// generateAiDesc 反推测试。
//
// 说明：GameExcelService 在测试环境初始化困难（依赖 rain-robot excel_manager_impl
// 和真实 Excel 资源），本测试统一传 nil，heroName/cardName/skillName 全部为空。
// 这恰好精确验证 generateAiDesc 的"分支顺序与拼接格式"——heroName 只是前缀拼接，
// 名字解析是独立环节（getHeroNameById/getCardNameById），不在本测试关注范围内。
// 与前端 use-asset-ai-desc.ts 完全一致的验收点 = 拼接格式一致。
// =====================================================================

// makeInitYanWu1Seat 构造最小 initYanWu（1 个座位，heroId=10105）。
// step.robotIdx=1 时反推到 customHeroes[0]，heroName 经 gameExcel 查（nil 时为空）。
func makeInitYanWu1Seat() InitYanWu {
	return InitYanWu{
		CustomHeroes: []CustomHero{
			{HeroId: 10105, Identity: 1, Color: 2},
		},
	}
}

// step1 构造 robotIdx=1 的 step。
func step1(action string) Step {
	return Step{Action: action, RobotIdx: 1}
}

// TestGenerateAiDesc_AckMsgNames 覆盖所有 Ack 类（isAssetIsAck 分支）。
// 期望：heroName + stepActionChs.label + "成功"。heroName="" 时 = label+"成功"。
func TestGenerateAiDesc_AckMsgNames(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	tests := []struct {
		name    string
		msgName string
		action  string
		want    string
	}{
		{"PlayCardAck + PlayCard", "PlayCardAck", "PlayCard", "出牌成功"},
		{"OptRoomActionAck + OptRoomAction", "OptRoomActionAck", "OptRoomAction", "响应成功"},
		{"DisCardAck + DisCard", "DisCardAck", "DisCard", "弃牌成功"},
		{"UseHeroSkillAck + UseHeroSkill", "UseHeroSkillAck", "UseHeroSkill", "使用技能成功"},
		{"PlayCardOverAck + PlayCardOver", "PlayCardOverAck", "PlayCardOver", "结束打牌成功"},
		// step.action 找不到 label 时返回空字符串
		{"PlayCardAck + 未知action", "PlayCardAck", "Unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{MsgName: tt.msgName, Attr: map[string]string{"Result": "0"}}
			got := generateAiDesc(asset, step1(tt.action), &initYanWu, newNameResolver(nil))
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGenerateAiDesc_CardMsgNames 覆盖 DrawCard/DisCard/PlayCard/GiveCard/CardEnhance/XianBaTouChou。
// generateCardDesc 逻辑：Count!=0 → verb+Count+"张牌"，Count!=0&&Cards → "牌为"，Cards → [id,...]。
func TestGenerateAiDesc_CardMsgNames(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	tests := []struct {
		name    string
		msgName string
		attr    map[string]string
		want    string
	}{
		// Count=1 + Cards=[1019] → "抽到1张牌牌为[1019]"
		{"DrawCard Count+Cards", "DrawCard", map[string]string{"Count": "1", "Cards": "1019"}, "抽到1张牌牌为[1019]"},
		// Count=2 + Cards=[1019,1020] → "弃掉2张牌牌为[1019,1020]"
		{"DisCard Count+2Cards", "DisCard", map[string]string{"Count": "2", "Cards": "1019 1020"}, "弃掉2张牌牌为[1019,1020]"},
		// Count=0 + Cards=[1019] → verb(打出) + [1019] = "打出[1019]"
		{"PlayCard onlyCards", "PlayCard", map[string]string{"Count": "0", "Cards": "1019"}, "打出[1019]"},
		// Count=1 + 无 Cards → "给1张牌"
		{"GiveCard onlyCount", "GiveCard", map[string]string{"Count": "1"}, "给1张牌"},
		// CardEnhance Count=1 + Cards=[1019] → "强化1张牌牌为[1019]"
		{"CardEnhance", "CardEnhance", map[string]string{"Count": "1", "Cards": "1019"}, "强化1张牌牌为[1019]"},
		// XianBaTouChou Count=2 + Cards=[1019,1020] → "先拔头筹拿到2张牌牌为[1019,1020]"
		{"XianBaTouChou", "XianBaTouChou", map[string]string{"Count": "2", "Cards": "1019 1020"}, "先拔头筹拿到2张牌牌为[1019,1020]"},
		// DrawCard 无 attr（全默认）→ Count=0,Cards=[] → ""（verb 都不输出）
		{"DrawCard empty", "DrawCard", map[string]string{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{MsgName: tt.msgName, Attr: tt.attr}
			got := generateAiDesc(asset, step1("PlayCard"), &initYanWu, newNameResolver(nil))
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGenerateAiDesc_CommonHpChange 覆盖 CommonHpChange。
// 期望格式：heroName + "受到来自{HpSrc}的伤害, 体力变化:{ChangeHp}, 当前体力: [{CurHp}], 最大体力: [{MaxHp}]"
func TestGenerateAiDesc_CommonHpChange(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()
	asset := Asset{MsgName: "CommonHpChange", Attr: map[string]string{
		"HpSrc": "HarmHuo", "ChangeHp": "-1", "CurHp": "4", "MaxHp": "5",
	}}
	got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
	assert.Equal(t, "受到来自HarmHuo的伤害, 体力变化:-1, 当前体力: [4], 最大体力: [5]", got)
}

// TestGenerateAiDesc_AttrChange 覆盖 AttrChange。
// 期望：heroName + ", 出杀次数: {ShaCount}, 手牌上限: {HandLimit}" 等（值为 0 的字段跳过）。
func TestGenerateAiDesc_AttrChange(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	t.Run("ShaCount+HandLimit", func(t *testing.T) {
		asset := Asset{MsgName: "AttrChange", Attr: map[string]string{"ShaCount": "2", "HandLimit": "5"}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		assert.Equal(t, ", 出杀次数: 2, 手牌上限: 5", got)
	})

	t.Run("全字段", func(t *testing.T) {
		asset := Asset{MsgName: "AttrChange", Attr: map[string]string{
			"ShaCount": "2", "HandLimit": "5", "EquipLimit": "3", "AttackRange": "1",
			"DistIncr": "1", "DistDecr": "1", "MaxHp": "7", "CurHp": "6",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		assert.Equal(t, ", 出杀次数: 2, 手牌上限: 5, 装备上限: 3, 攻击范围: 1, 距离增加: 1, 距离减少: 1, 最大生命: 7, 当前生命: 6", got)
	})
}

// TestGenerateAiDesc_SkillTrigger 覆盖 SkillTrigger。
// 期望：heroName + "发动技能{技能名}, 目标座位号: [{DestSeatIds}]" + Random 时 "(包含)" + Param 时 ", 参数[{Param}]"。
// nil gameExcel 时技能名为空。
func TestGenerateAiDesc_SkillTrigger(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	t.Run("基本", func(t *testing.T) {
		asset := Asset{MsgName: "SkillTrigger", Attr: map[string]string{
			"ActionValue": "1073", "DestSeatIds": "2",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		// 技能名查不到（nil gameExcel）→ 空
		assert.Equal(t, "发动技能, 目标座位号: 2", got)
	})

	t.Run("Random+Param", func(t *testing.T) {
		asset := Asset{MsgName: "SkillTrigger", Attr: map[string]string{
			"ActionValue": "1073", "DestSeatIds": "2 3", "Random": "true", "Param": "5",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		assert.Equal(t, "发动技能, 目标座位号: 2,3(包含), 参数[5]", got)
	})
}

// TestGenerateAiDesc_EquipChange 覆盖 EquipChange。
// 期望：heroName + "获得装备[{装备名}]({AddEquip})" + Count 时 ", 失去{Count}件装备" + RemoveEquip 时 ", 失去装备[...]"。
func TestGenerateAiDesc_EquipChange(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	t.Run("AddEquip", func(t *testing.T) {
		asset := Asset{MsgName: "EquipChange", Attr: map[string]string{"AddEquip": "1063"}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		// 装备名查不到（nil gameExcel）→ 空
		assert.Equal(t, "获得装备[](1063)", got)
	})

	t.Run("AddEquip+Count+RemoveEquip", func(t *testing.T) {
		asset := Asset{MsgName: "EquipChange", Attr: map[string]string{
			"AddEquip": "1063", "Count": "2", "RemoveEquip": "1001 1002",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		// RemoveEquip 名字查不到 → 走 eid 原值分支
		assert.Equal(t, "获得装备[](1063), 失去2件装备, 失去装备[1001 1002]", got)
	})
}

// TestGenerateAiDesc_ChangeCountry 覆盖 ChangeCountry。
// 期望：heroName + ", 主要势力: {name}" + ", 额外势力: {name}"。
func TestGenerateAiDesc_ChangeCountry(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()
	// MainCountry/ExtraCountry 是 numberArrD，存为数组，取首元素查 countryMap
	asset := Asset{MsgName: "ChangeCountry", Attr: map[string]string{
		"MainCountry": "2", "ExtraCountry": "3",
	}}
	got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
	assert.Equal(t, ", 主要势力: XiChu, 额外势力: XiHan", got)
}

// TestGenerateAiDesc_ReplaceCard 覆盖 ReplaceCard（generateReplaceCardSegment）。
// 这是最复杂的分支，测 2 例。
func TestGenerateAiDesc_ReplaceCard(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	t.Run("仅 ChangeCards 无替换目标", func(t *testing.T) {
		// ReplaceCards_ChangeCards=[1001]，无 ToCards → "替换{卡牌名}(1001)"
		// nil gameExcel 卡牌名查不到 → "替换(1001)"
		asset := Asset{MsgName: "ReplaceCard", Attr: map[string]string{
			"ReplaceCards_ChangeCards": "1001",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		assert.Equal(t, "替换卡牌:替换(1001)", got)
	})

	t.Run("ChangeCards 带 ToCards 单值", func(t *testing.T) {
		// ChangeCards=[1001]，ToCards 1001:1002（单值）→ "[{卡牌名}(1001) 替换为: [1002] ]"
		// nil gameExcel 卡牌名查不到 → "[(1001) 替换为: [1002] ]"；Random → "(任意)"
		asset := Asset{MsgName: "ReplaceCard", Attr: map[string]string{
			"ReplaceCards_ChangeCards":        "1001",
			"ReplaceCards_ChangeCardsToCards": "1001:1002",
			"ReplaceCards_ChangeCardsRandom":  "true",
		}}
		got := generateAiDesc(asset, step1("OnlyAsset"), &initYanWu, newNameResolver(nil))
		assert.Equal(t, "替换卡牌:[(1001) 替换为: [1002] ](任意)", got)
	})
}

// TestGenerateAiDesc_DefaultOther 覆盖未匹配的 msgName → "请输入描述"。
func TestGenerateAiDesc_DefaultOther(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()
	asset := Asset{MsgName: "SomeUnknownMsg", Attr: map[string]string{}}
	got := generateAiDesc(asset, step1("PlayCard"), &initYanWu, newNameResolver(nil))
	assert.Equal(t, "请输入描述", got)
}

// TestGenerateAiDesc_NoCustomHeroes 覆盖 initYanWu 无 customHeroes 时 heroName=""。
func TestGenerateAiDesc_NoCustomHeroes(t *testing.T) {
	emptyInit := InitYanWu{}
	asset := Asset{MsgName: "PlayCardAck", Attr: map[string]string{"Result": "0"}}
	got := generateAiDesc(asset, step1("PlayCard"), &emptyInit, newNameResolver(nil))
	assert.Equal(t, "出牌成功", got)
}

// =====================================================================
// deserializeAssetAttr 测试（验证 attr 字符串反序列化）
// =====================================================================

func TestDeserializeAssetAttr_Ack(t *testing.T) {
	got := deserializeAssetAttr(map[string]string{"Result": "0 1518"}, "PlayCardAck")
	assert.Equal(t, []int{0, 1518}, got["Result"])
}

func TestDeserializeAssetAttr_AckDefault(t *testing.T) {
	// Ack 无 Result 字段 → 默认 [0]
	got := deserializeAssetAttr(map[string]string{}, "PlayCardAck")
	assert.Equal(t, []int{0}, got["Result"])
}

func TestDeserializeAssetAttr_ReplaceDeserialize(t *testing.T) {
	got := deserializeAssetAttr(map[string]string{
		"ReplaceCards_ChangeCards":        "1001 1002",
		"ReplaceCards_ChangeCardsToCards": "1001:1002,1003 1002:1004",
	}, "ReplaceCard")
	assert.Equal(t, []int{1001, 1002}, got["ReplaceCards_ChangeCards"])
	toCards := got["ReplaceCards_ChangeCardsToCards"].(map[int][]int)
	assert.Equal(t, []int{1002, 1003}, toCards[1001])
	assert.Equal(t, []int{1004}, toCards[1002])
}

// =====================================================================
// generateInitYanWuDesc 测试
// =====================================================================

func TestGenerateInitYanWuDesc(t *testing.T) {
	initYanWu := InitYanWu{
		PresentId: 0,
		CardPile:  0,
		Cards:     []int32{1047, 1048},
		CustomHeroes: []CustomHero{
			{HeroId: 10105, Identity: 1, Color: 2, InitCards: []uint32{1040}, DelSkills: []uint32{1073}},
		},
	}
	got := generateInitYanWuDesc(initYanWu, newNameResolver(nil))
	// nil gameExcel → 卡牌/武将名解析为 ?ID
	assert.Contains(t, got, "[牌堆] 摸牌序列：[?1047, ?1048]")
	assert.Contains(t, got, "座位1: ?10105，身份1(主公)，势力2(XiChu)，手牌[?1040]")
	assert.Contains(t, got, "delSkills[?1073]")
	assert.NotContains(t, got, "presentId=") // presentId=0 省略
}

// =====================================================================
// generateStepDesc 测试
// =====================================================================

func TestGenerateStepDesc_PlayCard(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()
	step := Step{Action: "PlayCard", RobotIdx: 1, TargetsId: []int{2}, Cards: []uint32{1040}}
	got := generateStepDesc(step, initYanWu, newNameResolver(nil))
	assert.Equal(t, "座位1() 对 座位2() 打出 ?1040", got)
}

func TestGenerateStepDesc_AllActions(t *testing.T) {
	initYanWu := makeInitYanWu1Seat()

	tests := []struct {
		name string
		step Step
		want string
	}{
		{"PlayCardOver", Step{Action: "PlayCardOver", RobotIdx: 1}, "座位1() 结束出牌阶段"},
		{"OnlyAsset", Step{Action: "OnlyAsset", RobotIdx: 1}, "座位1() 空等待（仅校验断言，不发协议）"},
		{"Sleep", Step{Action: "Sleep", RobotIdx: 1, SleepTime: 1.5}, "座位1() 等待 1.5 秒"},
		{"WaitAction", Step{Action: "WaitAction", RobotIdx: 1}, "座位1() 等待轮到自己行动"},
		{"DisCard", Step{Action: "DisCard", RobotIdx: 1, Cards: []uint32{1040}}, "座位1() 弃牌 ?1040"},
		{"UseHeroSkill UUID", Step{Action: "UseHeroSkill", RobotIdx: 1, HeroSkillUUID: 3000001}, "座位1() 发动技能 技能(UUID:3000001)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateStepDesc(tt.step, initYanWu, newNameResolver(nil))
			assert.Equal(t, tt.want, got)
		})
	}
}

// =====================================================================
// ReverseTranslateCase 集成测试
// =====================================================================

func TestReverseTranslateCase(t *testing.T) {
	qaFuncCase := QAFuncCase{
		Case: "TC-17",
		InitYanWu: InitYanWu{
			Cards: []int32{1047},
			CustomHeroes: []CustomHero{
				{HeroId: 10105, Identity: 1, Color: 2, InitCards: []uint32{1040}},
			},
		},
		Steps: []Step{
			{Action: "PlayCard", RobotIdx: 1, TargetsId: []int{2}, Cards: []uint32{1040}},
			{Action: "OnlyAsset", RobotIdx: 2, Assets: []Asset{
				{MsgName: "PlayCardAck", Attr: map[string]string{"Result": "0"}},
				{MsgName: "CommonHpChange", Attr: map[string]string{"UnExpect": "true"}},
			}},
		},
	}
	got := ReverseTranslateCase(qaFuncCase, newNameResolver(nil))
	assert.Contains(t, got, "===== 前置条件反推")
	assert.Contains(t, got, "===== 步骤反推")
	assert.Contains(t, got, "===== 预期结果反推")
	assert.Contains(t, got, "1. 座位1() 对 座位2() 打出 ?1040")
	// UnExpect=true 的反向断言前缀（CommonHpChange 无字段 → 描述只剩 heroName=""）
	assert.Contains(t, got, "- 不期望：")
	// PlayCardAck 挂在 OnlyAsset step → action 有 label "仅断言" → "仅断言成功"
	assert.Contains(t, got, "- 仅断言成功")
	// 整体包含三段标题
	assert.Equal(t, 3, strings.Count(got, "=====")/2)
}

// =====================================================================
// 全量集成测试：遍历 cases/fight_cases/ 下所有 JSON 文件的所有 case，
// 跑 ReverseTranslateCase，验证三维度反推正常输出。
//
// 说明：
//   - gameExcel 传 nil，heroName/cardName/skillName 为空是正常的（与单元测试一致）。
//   - 测试运行 cwd 是包目录 backend/pkg/function-test/，回退 4 层到 worktree 根。
//   - 单个 case 失败不会让整个测试挂（subtest + recover）。
// =====================================================================

func TestAllFightCases_ReverseTranslate(t *testing.T) {
	// 测试文件位于 backend/pkg/function-test/，回退 3 层到 worktree 根。
	casesDir := filepath.Join("..", "..", "..", "cases", "fight_cases")

	files, err := filepath.Glob(filepath.Join(casesDir, "*.json"))
	if err != nil {
		t.Fatalf("Glob fight_cases 失败: %v", err)
	}
	if len(files) == 0 {
		// 路径不对时，尝试用 os.Getwd 打印调试，帮助定位。
		cwd, _ := os.Getwd()
		t.Fatalf("未找到任何 JSON 文件，casesDir=%q (abs=%q) cwd=%q", casesDir, mustAbs(casesDir), cwd)
	}

	type failure struct {
		file     string
		caseName string
		reason   string
	}

	totalCases := 0
	passedCases := 0
	var failures []failure

	for _, file := range files {
		fileName := filepath.Base(file)
		data, err := os.ReadFile(file)
		if err != nil {
			failures = append(failures, failure{fileName, "<读文件>", err.Error()})
			continue
		}

		var caseList []QAFuncCase
		if err := json.Unmarshal(data, &caseList); err != nil {
			failures = append(failures, failure{fileName, "<解析JSON>", err.Error()})
			continue
		}

		for _, c := range caseList {
			totalCases++
			caseName := c.Case
			if caseName == "" {
				caseName = "<空case名>"
			}

			ok, reason := checkReverseTranslateCase(c)
			if ok {
				passedCases++
				continue
			}
			failures = append(failures, failure{fileName, caseName, reason})
		}
	}

	t.Logf("===== 全量反推统计：总 case 数=%d，通过=%d，失败=%d，JSON 文件数=%d =====",
		totalCases, passedCases, len(failures), len(files))

	for _, f := range failures {
		t.Errorf("失败: %s / %s — %s", f.file, f.caseName, f.reason)
	}
}

// mustAbs 返回绝对路径（用于错误信息调试），失败时返回原路径。
func mustAbs(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}

// checkReverseTranslateCase 跑一次 ReverseTranslateCase，验证三维度标题齐全、
// 有 assets 的 case 其 assets 部分每个断言都有 generateAiDesc 输出。
// 返回 (是否通过, 失败原因)。
func checkReverseTranslateCase(c QAFuncCase) (bool, string) {
	// 1. 不 panic（defer recover）
	var got string
	var panicReason string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicReason = "panic: " + fmt.Sprint(r)
			}
		}()
		got = ReverseTranslateCase(c, newNameResolver(nil))
	}()
	if panicReason != "" {
		return false, "反推 panic: " + panicReason
	}

	// 2-4. 三维度标题齐全
	if !strings.Contains(got, "前置条件反推（initYanWu）") {
		return false, "缺少标题 前置条件反推（initYanWu）"
	}
	if !strings.Contains(got, "步骤反推（steps）") {
		return false, "缺少标题 步骤反推（steps）"
	}
	if !strings.Contains(got, "预期结果反推（assets）") {
		return false, "缺少标题 预期结果反推（assets）"
	}

	// 5. 有 assets 的 case，assets 部分不能为空。
	//    ReverseTranslateCase 的 assets 段对每个 step 的每个 asset 调 generateAiDesc，
	//    生成形如 "- ..." 的行。统计 assets 标题之后的非标题、非空行作为断言输出。
	if totalAssets := countAssets(c); totalAssets > 0 {
		assetsSection := extractAfter(got, "===== 预期结果反推（assets）=====")
		// 去掉标题行本身，剩下应为每个 asset 对应的一行描述。
		lines := splitNonEmptyLines(assetsSection)
		// 允许 generateAiDesc 返回 "请输入描述" 等任意非空内容，但行数应 >= asset 数。
		if len(lines) < totalAssets {
			return false, fmt.Sprintf("assets 断言输出不足: 有 %d 个 asset，但反推只产生 %d 行",
				totalAssets, len(lines))
		}
	}

	return true, ""
}

// countAssets 统计 case 中所有 step 的 assets 总数。
func countAssets(c QAFuncCase) int {
	n := 0
	for _, s := range c.Steps {
		n += len(s.Assets)
	}
	return n
}

// extractAfter 返回 s 中第一次出现 marker 之后的内容（不含 marker 行）。
func extractAfter(s, marker string) string {
	idx := strings.Index(s, marker)
	if idx < 0 {
		return ""
	}
	return s[idx+len(marker):]
}

// splitNonEmptyLines 按换行切分，去掉空行和只含空白/分隔符（=====）的行。
func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "=====") {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
