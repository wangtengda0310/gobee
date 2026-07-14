package functiontest

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/game"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel"
	"git.devcloud.ztgame.com/v-tangfangda/rain-robot/project/xcard/xcard_excel/excel_config"
)

// =====================================================================
// 语义回译（reverse-translate）：把 QAFuncCase（JSON）机械地反推成中文描述。
//
// 三维度反推：
//   1. assets  —— 精确复刻前端 use-asset-ai-desc.ts 的 generateAiDesc
//   2. initYanWu —— 按 reverse-translate.md §2.1 模板反推（座位/武将/身份/势力/手牌/技能/装备）
//   3. steps   —— 按 reverse-translate.md §2.2 模板反推（8 种 Action）
//
// 客观性约束：反推函数是纯函数（JSON in → 中文 out），ID 解析失败写 ?{ID}，不润色不改写。
// =====================================================================

// countryMap 势力ID → 势力名，复刻前端 ECountry.ts 的 countryMap。
// 仅用于 ChangeCountry 的 MainCountry/ExtraCountry 字段反推（这俩才是真正的国家/势力）。
//
// ⚠️ 注意：countryMap 是 ECountry 国家表（1=Qin 秦 ... 16=Yan 燕），与座位颜色是不同维度。
// 座位 CustomHero.Color 的真实语义是"身份阵营颜色"（来自游戏配表 IdentityEncodeRule_身份编码规则表），
// 取值见前端 Identity.ts 的 excelIdentityColorMap（主公1~8、反贼65、内奸97 等）。
// 下方 hero.Color 反推处误用了本表翻译，属已知 bug（详见该处注释）。
var countryMap = map[int]string{
	0:  "CoNone",
	1:  "Qin",
	2:  "XiChu",
	3:  "XiHan",
	4:  "DongHan",
	5:  "Huang",
	6:  "CaoWei",
	7:  "Shu",
	8:  "SunWu",
	9:  "Fei",
	10: "ZhangChu",
	11: "Wei",
	12: "Chu",
	13: "Qi",
	14: "Zhao",
	15: "Han",
	16: "Yan",
}

// identityMap 身份ID → 身份名，复刻前端 identity 颜色映射的语义。
// initYanWu.customHeroes[idx].identity 反推时使用。
var identityMap = map[int32]string{
	1:  "主公",
	2:  "忠臣",
	3:  "反贼",
	4:  "内奸",
	13: "特殊",
}

// actionsSelectOptionLabel Action枚举 → 中文标签，复刻前端 actionsSelectOption。
// assets 的 Ack 类反推时，用 step.action 查这里得到"出牌/弃牌/响应"等动作中文。
func actionsSelectOptionLabel(action string) (string, bool) {
	m := map[string]string{
		"Sleep":         "等待",
		"PlayCard":      "出牌",
		"DisCard":       "弃牌",
		"OptRoomAction": "响应",
		"UseHeroSkill":  "使用技能",
		"PlayCardOver":  "结束打牌",
		"OnlyAsset":     "仅断言",
		"WaitAction":    "等待行动",
	}
	v, ok := m[action]
	return v, ok
}

// assetSelectOptionHas msgName 是否在 assetAllSelectOption 列表中（即 assetActionChs 非空）。
// 复刻前端 assetSelectOption(step).find(o => o.value == assetName) 能否找到。
// Ack 类是动态生成的，统一视为"能找到"（前端 isAssetIsAck 分支单独处理）。
func assetSelectOptionHas(msgName string) bool {
	// 所有在 assetAllSelectOption 里的固定 msgName（来自 StepActionsAndAssetsSelect.ts）
	set := map[string]bool{
		"RoomGameActionAsset": true,
		"UpdateProperty":      true,
		"UpdateHeroSkill":     true,
		"DrawCard":            true,
		"PlayCard":            true,
		"DisCard":             true,
		"GiveCard":            true,
		"CommonHpChange":      true,
		"ReplaceCard":         true,
		"EquipChange":         true,
		"AttrChange":          true,
		"ChangeCountry":       true,
		"SkillTrigger":        true,
		"CardEnhance":         true,
		"XianBaTouChou":       true,
	}
	return set[msgName]
}

// isAssetIsAck 判断 asset.msgName 是否以 "Ack" 结尾，复刻前端 isAssetIsAck。
func isAssetIsAck(msgName string) bool {
	return strings.HasSuffix(msgName, "Ack")
}

// ---------------------------------------------------------------------
// 反序列化：attr map[string]string → 结构化 map[string]interface{}
// 复刻 AssetMapTrans.ts 的 assetMap2AssetList 各 xx_D 函数。
// ---------------------------------------------------------------------

// deserializeAssetAttr 按 msgName 分支，把 asset.attr（字符串）反序列化成结构化字段。
// 对应前端 assetMap2AssetList。返回的 map 中：数字用 int，数字数组用 []int，
// 布尔用 bool，字符串用 string，二维结构（ReplaceCard 的 ToCards 等）用 map[int][]int。
//
// 流程：
//  1. Ack 类 → 只反序列化 Result（numberArrD，默认 [0]）
//  2. 各业务 msgName → 调用对应的 xx_D 反序列化逻辑
//  3. 未匹配的 msgName → 返回空 map（前端 else 分支也为空）
func deserializeAssetAttr(attr map[string]string, msgName string) map[string]interface{} {
	result := make(map[string]interface{})

	if isAssetIsAck(msgName) {
		// ackD: numberArrD("Result", ..., [0])
		result["Result"] = numberArrD(attr, "Result", []int{0})
		return result
	}

	switch msgName {
	case "RoomGameActionAsset":
		// roomGameActionAssetD
		result["ActionType"] = numberD(attr, "ActionType", 0)
		result["ActionValue"] = numberD(attr, "ActionValue", 0)
		result["Params"] = numberArrD(attr, "Params", nil)
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "UpdateProperty":
		// updatePropertyD
		result["IdType"] = strD(attr, "IdType", "None")
		result["PropID"] = strD(attr, "PropID", "None")
		result["PropValue"] = numberD(attr, "PropValue", 0)
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "UpdateHeroSkill":
		// updateHeroSkillD（CardParamMap 反序列化对本反推无用，这里只取 SkillUUid）
		result["SkillUUid"] = numberArrD(attr, "SkillUUid", nil)
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "DrawCard", "PlayCard", "DisCard", "GiveCard", "XianBaTouChou", "CardEnhance":
		// normalCardD
		if msgName == "DrawCard" {
			result["ActionType"] = numberD(attr, "ActionType", 0)
			result["ActionValue"] = numberD(attr, "ActionValue", 0)
		}
		result["Count"] = numberD(attr, "Count", 0)
		result["Cards"] = numberArrD(attr, "Cards", nil)
		result["UnexpectCards"] = numberArrD(attr, "UnexpectCards", nil)
		result["Random"] = boolD(attr, "Random")
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "ReplaceCard":
		// replaceCardD
		result["ReplaceCards_ChangeCards"] = numberArrD(attr, "ReplaceCards_ChangeCards", nil)
		result["ReplaceCards_DrawCards"] = numberArrD(attr, "ReplaceCards_DrawCards", nil)
		result["ReplaceCards_ChangeCardsToCards"] = replaceDeserialize(attr, "ReplaceCards_ChangeCardsToCards")
		result["ReplaceCards_ChangeCardsToPoints"] = replaceDeserialize(attr, "ReplaceCards_ChangeCardsToPoints")
		result["ReplaceCards_ChangeCardsToAttrTypes"] = replaceDeserialize(attr, "ReplaceCards_ChangeCardsToAttrTypes")
		result["ReplaceCards_DrawCardsToCards"] = replaceDeserialize(attr, "ReplaceCards_DrawCardsToCards")
		result["ReplaceCards_DrawCardsToPoints"] = replaceDeserialize(attr, "ReplaceCards_DrawCardsToPoints")
		result["ReplaceCards_DrawCardsToAttrTypes"] = replaceDeserialize(attr, "ReplaceCards_DrawCardsToAttrTypes")
		result["ReplaceCards_ChangeCardsRandom"] = boolD(attr, "ReplaceCards_ChangeCardsRandom")
		result["ReplaceCards_DrawCardsRandom"] = boolD(attr, "ReplaceCards_DrawCardsRandom")
	case "CommonHpChange":
		// commonHpChangeD
		result["HpSrc"] = strD(attr, "HpSrc", "")
		result["ChangeHp"] = numberD(attr, "ChangeHp", 0)
		result["CurHp"] = numberD(attr, "CurHp", 0)
		result["MaxHp"] = numberD(attr, "MaxHp", 0)
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "AttrChange":
		// attrChangeD
		result["ShaCount"] = numberD(attr, "ShaCount", 0)
		result["HandLimit"] = numberD(attr, "HandLimit", 0)
		result["EquipLimit"] = numberD(attr, "EquipLimit", 0)
		result["AttackRange"] = numberD(attr, "AttackRange", 0)
		result["DistIncr"] = numberD(attr, "DistIncr", 0)
		result["DistDecr"] = numberD(attr, "DistDecr", 0)
		result["MaxHp"] = numberD(attr, "MaxHp", 0)
		result["CurHp"] = numberD(attr, "CurHp", 0)
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "ChangeCountry":
		// changeCountryD
		result["MainCountry"] = numberArrD(attr, "MainCountry", nil)
		result["ExtraCountry"] = numberArrD(attr, "ExtraCountry", nil)
	case "SkillTrigger":
		// skillTriggerD
		result["ActionValue"] = numberD(attr, "ActionValue", 0)
		result["DestSeatIds"] = numberArrD(attr, "DestSeatIds", nil)
		result["Param"] = numberD(attr, "Param", 0)
		result["Random"] = boolD(attr, "Random")
		result["UnExpect"] = boolD(attr, "UnExpect")
	case "EquipChange":
		// equipChangeD
		result["AddEquip"] = numberD(attr, "AddEquip", 0)
		result["Count"] = numberD(attr, "Count", 0)
		result["RemoveEquip"] = numberArrD(attr, "RemoveEquip", nil)
		result["Random"] = boolD(attr, "Random")
		result["UnExpect"] = boolD(attr, "UnExpect")
	}

	// 前端 assetMap2AssetList 末尾：if attr['UnExpect'] 则 assetList['UnExpect'] = attr=="true"
	// 上面的 xx_D 已经处理了 UnExpect，这里无需重复。

	return result
}

// numberD 复刻 AssetMapTrans.ts 的 numberD：attr[key] 是合法数字则取 Number||0，否则用 def。
// 注意 TS：Number.isFinite(Number(attr[key])) 为真时赋值 Number(attr[key])||0（空字符串/0 都变 0）。
func numberD(attr map[string]string, key string, def int) int {
	v, ok := attr[key]
	if !ok {
		return def
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return def
	}
	// Number(v)||0 ：NaN||0=0，0||0=0，正数不变。Go 里 parseFloat 成功即非 NaN。
	if n == 0 {
		return 0
	}
	return int(n)
}

// numberArrD 复刻 numberArrD：attr[key]!="" 时按空格切分转数字数组，否则用 def。
// TS: attr[key]=="" → []，否则 split(" ").filter(isFinite).map(Number) || []
func numberArrD(attr map[string]string, key string, def []int) []int {
	v, ok := attr[key]
	if !ok {
		return def
	}
	if v == "" {
		return []int{}
	}
	parts := strings.Fields(v)
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		result = append(result, int(n))
	}
	return result
}

// strD 复刻 strD：attr[key]!=null 用原值，否则用 def。
func strD(attr map[string]string, key, def string) string {
	if v, ok := attr[key]; ok {
		return v
	}
	return def
}

// boolD 复刻 boolD：attr[key]!=null 时 attr[key]=="true"，否则 def（默认 false）。
func boolD(attr map[string]string, key string) bool {
	if v, ok := attr[key]; ok {
		return v == "true"
	}
	return false
}

// replaceDeserialize 复刻 replaceDeserialize：解析 "123:1,2|245:3" 为 map[int][]int。
// TS: split(" ").filter(...).map(cs => [kv[0], kv[1]?kv[1].split(",").filter(Number).map(Number):[]])
// 返回 map[int][]int（key 为源卡牌ID，value 为目标ID列表）。
func replaceDeserialize(attr map[string]string, key string) map[int][]int {
	result := make(map[int][]int)
	v, ok := attr[key]
	if !ok || v == "" {
		return result
	}
	for _, cs := range strings.Fields(v) {
		kv := strings.SplitN(cs, ":", 2)
		if len(kv) == 0 {
			continue
		}
		k, err := strconv.Atoi(kv[0])
		if err != nil {
			continue
		}
		var arr []int
		if len(kv) == 2 && kv[1] != "" {
			for _, c := range strings.Split(kv[1], ",") {
				if n, err := strconv.Atoi(c); err == nil {
					arr = append(arr, n)
				}
			}
		}
		result[k] = arr
	}
	return result
}

// ---------------------------------------------------------------------
// ID → 名字解析（通过 nameResolver 预取）
// ---------------------------------------------------------------------

// nameResolver 预取并持有本次描述生成所需的 ID→配置 map，
// 避免 getHeroNameById 等每个 helper 各自调 GameExcelService.GetAllXxxCfg（全量查询 + 日志噪音）。
// 在 generateAiDesc / ReverseTranslateCase 入口一次性构造，供本次生成的所有 ID→名 解析复用。
type nameResolver struct {
	heroes map[excel.EHeroId]*excel_config.HeroConfig
	cards  map[int32]*excel.CardsTemplate
	skills map[excel.ESkillId]*excel.SkillsTemplate
}

// newNameResolver 从 gameExcel 一次性预取三类配置。
// gameExcel 为 nil（单元测试场景）时返回空 resolver，heroName/cardName/skillName 一律返回空串，与原 helper 的 nil 语义一致。
func newNameResolver(gameExcel *game.GameExcelService) *nameResolver {
	r := &nameResolver{}
	if gameExcel == nil {
		return r
	}
	r.heroes = gameExcel.GetAllHeroCfg()
	r.cards = gameExcel.GetAllCardCfg()
	r.skills = gameExcel.GetAllSkillCfg()
	return r
}

// heroName 通过 heroId 查武将名。未命中返回空字符串。
func (r *nameResolver) heroName(heroId uint32) string {
	if r == nil || r.heroes == nil {
		return ""
	}
	if h, ok := r.heroes[excel.EHeroId(heroId)]; ok && h != nil {
		return h.Name
	}
	return ""
}

// cardName 通过 cardId 查卡牌名。未命中返回空字符串。
func (r *nameResolver) cardName(cardId int) string {
	if r == nil || r.cards == nil {
		return ""
	}
	if c, ok := r.cards[int32(cardId)]; ok && c != nil {
		return c.Name
	}
	return ""
}

// skillName 通过 skillId 查技能名。未命中返回空字符串。
func (r *nameResolver) skillName(skillId int) string {
	if r == nil || r.skills == nil {
		return ""
	}
	if s, ok := r.skills[excel.ESkillId(skillId)]; ok && s != nil {
		return s.SkillName
	}
	return ""
}

// ---------------------------------------------------------------------
// generateAiDesc —— 精确复刻 use-asset-ai-desc.ts 的 generateAiDesc
// ---------------------------------------------------------------------

// generateAiDesc 生成单条 asset 的智能描述，精确复刻前端 generateAiDesc 的分支顺序和拼接格式。
//
// 流程（与 TS 第 83-176 行完全一致）：
//  1. 计算 stepActionChs（action 中文）、assetActionChs（msgName 是否在选项中）、heroName
//  2. isAssetIsAck → heroName + stepActionChs.label + "成功"
//  3. DrawCard/DisCard/PlayCard/GiveCard → heroName + generateCardDesc(动词)
//  4. CommonHpChange → HpSrc/ChangeHp/CurHp/MaxHp 拼接
//  5. ReplaceCard → generateReplaceCardSegment(ChangeCards) + (DrawCards)
//  6. AttrChange → 8 个属性字段拼接
//  7. ChangeCountry → countryMap 查势力名
//  8. CardEnhance → generateCardDesc("强化")
//  9. SkillTrigger → 发动技能 + 目标座位 + Random + Param
// 10. XianBaTouChou → generateCardDesc("先拔头筹拿到")
// 11. EquipChange → 获得装备/失去装备
// 12. 其他 → "请输入描述"
func generateAiDesc(asset Asset, step Step, initYanWu *InitYanWu, resolver *nameResolver) string {
	action := step.Action
	assetName := asset.MsgName

	stepActionChs, stepActionOk := actionsSelectOptionLabel(action)
	assetActionChs := assetSelectOptionHas(assetName)
	heroName := getHeroNameForStep(initYanWu, step, resolver)

	// 先反序列化 attr，供后续字段读取
	assetList := deserializeAssetAttr(asset.Attr, assetName)

	if isAssetIsAck(assetName) {
		if stepActionOk {
			return heroName + stepActionChs + "成功"
		}
		return ""
	}

	if assetName == "DrawCard" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "抽到")
		}
		return ""
	}
	if assetName == "DisCard" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "弃掉")
		}
		return ""
	}
	if assetName == "PlayCard" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "打出")
		}
		return ""
	}
	if assetName == "GiveCard" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "给")
		}
		return ""
	}
	if assetName == "CommonHpChange" {
		if !assetActionChs {
			return ""
		}
		s := heroName
		if v, ok := assetList["HpSrc"].(string); ok && v != "" {
			s += "受到来自" + v + "的伤害"
		}
		if v, ok := assetList["ChangeHp"].(int); ok && v != 0 {
			s += ", 体力变化:" + strconv.Itoa(v)
		}
		if v, ok := assetList["CurHp"].(int); ok && v != 0 {
			s += fmt.Sprintf(", 当前体力: [%d]", v)
		}
		if v, ok := assetList["MaxHp"].(int); ok && v != 0 {
			s += fmt.Sprintf(", 最大体力: [%d]", v)
		}
		return s
	}
	if assetName == "ReplaceCard" {
		if !assetActionChs {
			return ""
		}
		cc := generateReplaceCardSegment(assetList, "ReplaceCards_ChangeCards",
			"ReplaceCards_ChangeCardsToCards", "ReplaceCards_ChangeCardsToPoints",
			"ReplaceCards_ChangeCardsToAttrTypes", "ReplaceCards_ChangeCardsRandom", resolver)
		dc := generateReplaceCardSegment(assetList, "ReplaceCards_DrawCards",
			"ReplaceCards_DrawCardsToCards", "ReplaceCards_DrawCardsToPoints",
			"ReplaceCards_DrawCardsToAttrTypes", "ReplaceCards_DrawCardsRandom", resolver)
		s := heroName
		if cc != "" {
			s += "替换卡牌:" + cc
		}
		if dc != "" && cc != "" {
			s += ", 同时生成卡牌:" + dc
		} else if dc != "" {
			s += "生成卡牌" + dc
		}
		return s
	}
	if assetName == "AttrChange" {
		if !assetActionChs {
			return ""
		}
		s := heroName
		s += appendAttrChangeField(assetList, "ShaCount", ", 出杀次数: ")
		s += appendAttrChangeField(assetList, "HandLimit", ", 手牌上限: ")
		s += appendAttrChangeField(assetList, "EquipLimit", ", 装备上限: ")
		s += appendAttrChangeField(assetList, "AttackRange", ", 攻击范围: ")
		s += appendAttrChangeField(assetList, "DistIncr", ", 距离增加: ")
		s += appendAttrChangeField(assetList, "DistDecr", ", 距离减少: ")
		s += appendAttrChangeField(assetList, "MaxHp", ", 最大生命: ")
		s += appendAttrChangeField(assetList, "CurHp", ", 当前生命: ")
		return s
	}
	if assetName == "ChangeCountry" {
		if !assetActionChs {
			return ""
		}
		s := heroName
		mainCountry := firstInt(assetList, "MainCountry")
		extraCountry := firstInt(assetList, "ExtraCountry")
		if name, ok := countryMap[mainCountry]; ok && name != "" {
			s += ", 主要势力: " + name
		}
		if name, ok := countryMap[extraCountry]; ok && name != "" {
			s += ", 额外势力: " + name
		}
		return s
	}
	if assetName == "CardEnhance" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "强化")
		}
		return ""
	}
	if assetName == "SkillTrigger" {
		if !assetActionChs {
			return ""
		}
		s := heroName
		destSeatIds, hasDest := assetList["DestSeatIds"].([]int)
		if hasDest && len(destSeatIds) > 0 {
			actionValue, _ := assetList["ActionValue"].(int)
			skillName := ""
			if actionValue != 0 {
				skillName = resolver.skillName(actionValue)
			}
			s += fmt.Sprintf("发动技能%s, 目标座位号: %s", skillName, intSliceToString(destSeatIds))
		}
		random, _ := assetList["Random"].(bool)
		if random && hasDest && len(destSeatIds) > 0 {
			s += "(包含)"
		}
		if v, ok := assetList["Param"].(int); ok && v != 0 {
			s += fmt.Sprintf(", 参数[%d]", v)
		}
		return s
	}
	if assetName == "XianBaTouChou" {
		if assetActionChs {
			return heroName + generateCardDesc(assetList, "先拔头筹拿到")
		}
		return ""
	}
	if assetName == "EquipChange" {
		if !assetActionChs {
			return ""
		}
		s := heroName
		if addEquip, ok := assetList["AddEquip"].(int); ok && addEquip != 0 {
			equipName := resolver.cardName(addEquip)
			s += fmt.Sprintf("获得装备[%s](%d)", equipName, addEquip)
		}
		if count, ok := assetList["Count"].(int); ok && count != 0 {
			s += fmt.Sprintf(", 失去%d件装备", count)
		}
		if removeEquip, ok := assetList["RemoveEquip"].([]int); ok && len(removeEquip) > 0 {
			parts := make([]string, 0, len(removeEquip))
			for _, eid := range removeEquip {
				name := resolver.cardName(eid)
				if name != "" {
					parts = append(parts, fmt.Sprintf("[%s](%d)", name, eid))
				} else {
					parts = append(parts, strconv.Itoa(eid))
				}
			}
			s += ", 失去装备[" + strings.Join(parts, " ") + "]"
		}
		return s
	}

	return "请输入描述"
}

// appendAttrChangeField 拼接 AttrChange 的单个字段（值为 0 时跳过，与 TS 的 truthy 判断一致）。
func appendAttrChangeField(assetList map[string]interface{}, key, label string) string {
	if v, ok := assetList[key].(int); ok && v != 0 {
		return label + strconv.Itoa(v)
	}
	return ""
}

// firstInt 取 assetList[key] 的第一个 int 元素（ChangeCountry 的 MainCountry/ExtraCountry 是数组）。
func firstInt(assetList map[string]interface{}, key string) int {
	if arr, ok := assetList[key].([]int); ok && len(arr) > 0 {
		return arr[0]
	}
	return 0
}

// intSliceToString 复刻 TS assetList["DestSeatIds"].toString()，格式 "1,2,3"。
func intSliceToString(arr []int) string {
	parts := make([]string, 0, len(arr))
	for _, v := range arr {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

// generateCardDesc 复刻 TS generateCardDesc（第 30-34 行）。
// verb + Count + "张牌"（Count 非 0），中间连接词（Count&&Cards → "牌为"，!Count&&Cards → verb），
// Cards 数组 toString（格式 "id1,id2"）。
func generateCardDesc(assetList map[string]interface{}, verb string) string {
	count, _ := assetList["Count"].(int)
	cards, _ := assetList["Cards"].([]int)

	s := ""
	if count != 0 {
		s += verb + strconv.Itoa(count) + "张牌"
	}
	// 连接词
	if count != 0 && len(cards) > 0 {
		s += "牌为"
	} else if count == 0 && len(cards) > 0 {
		s += verb
	}
	if len(cards) > 0 {
		s += "[" + intSliceToString(cards) + "]"
	}
	return s
}

// generateReplaceCardSegment 复刻 TS generateReplaceCardSegment（第 39-73 行）。
// 遍历 fromCards（源卡牌数组），对每张牌查 toCards/toPoints/toAttrTypes 映射，拼接替换描述。
// randomKey 为真时整段追加 "(任意)"。
func generateReplaceCardSegment(assetList map[string]interface{}, cardListKey, toCardsKey, toPointsKey, toAttrTypesKey, randomKey string, resolver *nameResolver) string {
	fromCards, _ := assetList[cardListKey].([]int)
	if len(fromCards) == 0 {
		return ""
	}

	toCards, _ := assetList[toCardsKey].(map[int][]int)
	toPoints, _ := assetList[toPointsKey].(map[int][]int)
	toAttrTypes, _ := assetList[toAttrTypesKey].(map[int][]int)

	var segments []string
	for _, c := range fromCards {
		toCardsStr := ""
		if arr, ok := toCards[c]; ok && len(arr) > 0 {
			if len(arr) == 1 {
				toCardsStr = strconv.Itoa(arr[0])
			} else {
				parts := make([]string, 0, len(arr))
				for _, sid := range arr {
					parts = append(parts, fmt.Sprintf("%s(%d)", resolver.skillName(sid), sid))
				}
				toCardsStr = strings.Join(parts, ",") + "(包含)"
			}
		}
		toPointsStr := ""
		if arr, ok := toPoints[c]; ok && len(arr) > 0 {
			if len(arr) == 1 {
				toPointsStr = strconv.Itoa(arr[0])
			} else {
				parts := make([]string, 0, len(arr))
				for _, p := range arr {
					parts = append(parts, strconv.Itoa(p))
				}
				toPointsStr = strings.Join(parts, ",") + "(包含)"
			}
		}
		toAttrTypesStr := ""
		if arr, ok := toAttrTypes[c]; ok && len(arr) > 0 {
			if len(arr) == 1 {
				toAttrTypesStr = strconv.Itoa(arr[0])
			} else {
				parts := make([]string, 0, len(arr))
				for _, a := range arr {
					parts = append(parts, strconv.Itoa(a))
				}
				toAttrTypesStr = strings.Join(parts, ",") + "(包含)"
			}
		}

		cardName := resolver.cardName(c)
		if toCardsStr != "" || toPointsStr != "" || toAttrTypesStr != "" {
			seg := fmt.Sprintf("[%s(%d) 替换为: [%s]", cardName, c, toCardsStr)
			if toPointsStr != "" {
				seg += ", 点数:[" + toPointsStr + "]"
			}
			if toAttrTypesStr != "" {
				seg += ", 属性[" + toAttrTypesStr + "]"
			}
			seg += " ]"
			segments = append(segments, seg)
		} else {
			segments = append(segments, fmt.Sprintf("替换%s(%d)", cardName, c))
		}
	}

	result := strings.Join(segments, "; ")
	if random, _ := assetList[randomKey].(bool); random {
		result += "(任意)"
	}
	return result
}

// getHeroNameForStep 复刻 TS getHeroName（第 18-24 行）。
// step.robotIdx → customHeroes[(robotIdx-1) % len].heroId → 查武将名。
func getHeroNameForStep(initYanWu *InitYanWu, step Step, resolver *nameResolver) string {
	if initYanWu == nil || len(initYanWu.CustomHeroes) == 0 {
		return ""
	}
	idx := (step.RobotIdx - 1) % len(initYanWu.CustomHeroes)
	if idx < 0 {
		idx += len(initYanWu.CustomHeroes)
	}
	heroId := initYanWu.CustomHeroes[idx].HeroId
	if heroId == 0 {
		return ""
	}
	return resolver.heroName(heroId)
}

// ---------------------------------------------------------------------
// generateInitYanWuDesc —— 按 reverse-translate.md §2.1 模板反推 initYanWu
// ---------------------------------------------------------------------

// generateInitYanWuDesc 把 initYanWu 反推成中文，对照文档「前置条件」列。
//
// 流程：
//  1. 输出公共字段（牌堆摸牌序列、身份局、超时），缺省值省略
//  2. 逐个 customHero 输出座位描述（武将/身份/势力/手牌/delSkills/addSkills/装备/augurCards）
//  3. ID 全部解析为名字（格式 "名字(ID)"），解析失败写 "?ID"
func generateInitYanWuDesc(initYanWu InitYanWu, resolver *nameResolver) string {
	var sb strings.Builder

	// 公共字段
	if len(initYanWu.Cards) > 0 {
		sb.WriteString("[牌堆] 摸牌序列：[")
		cardNames := make([]string, 0, len(initYanWu.Cards))
		for _, c := range initYanWu.Cards {
			cardNames = append(cardNames, resolveCardName(resolver, int(c)))
		}
		sb.WriteString(strings.Join(cardNames, ", "))
		sb.WriteString("]\n")
	}
	if initYanWu.PresentId != 0 {
		fmt.Fprintf(&sb, "[身份局] presentId=%d，定制牌堆 cardPile=%d\n", initYanWu.PresentId, initYanWu.CardPile)
	}
	if initYanWu.OperateTimeMs != 0 {
		fmt.Fprintf(&sb, "[超时] operateTimeMs=%d\n", initYanWu.OperateTimeMs)
	}

	// 逐座位
	for idx, hero := range initYanWu.CustomHeroes {
		seat := idx + 1
		heroName := resolveHeroName(resolver, hero.HeroId)
		identityName, identityOk := identityMap[hero.Identity]
		identityStr := strconv.Itoa(int(hero.Identity))
		if identityOk {
			identityStr = fmt.Sprintf("%d(%s)", hero.Identity, identityName)
		}
		// hero.Color 来自座位"颜色下拉框"，真实语义=身份阵营颜色编号（游戏配表 IdentityEncodeRule，
		// 见前端 Identity.ts excelIdentityColorMap：主公1~8、反贼65、内奸97…），并非 ECountry 国家。
		// ⚠️ 已知 bug：此处误用 countryMap（国家表）翻译，导致 color∈1~16 被错译为"秦/西楚…"等国家名，
		//    color>16 查不到则只显示纯数字。正确做法应按身份阵营颜色显示（暂未修复）。
		countryName, countryOk := countryMap[int(hero.Color)]
		colorStr := strconv.Itoa(int(hero.Color))
		if countryOk {
			colorStr = fmt.Sprintf("%d(%s)", hero.Color, countryName)
		}

		fmt.Fprintf(&sb, "座位%d: %s，身份%s，势力%s，", seat, heroName, identityStr, colorStr)
		sb.WriteString("手牌[" + resolveCardIds(resolver, toUint32Slice(hero.InitCards)) + "]，")
		// delSkills 即使为空也输出（矛盾高发区）
		sb.WriteString("delSkills[" + resolveSkillIds(resolver, hero.DelSkills) + "]，")
		sb.WriteString("addSkills[" + resolveSkillIds(resolver, hero.AddSkills) + "]，")
		sb.WriteString("装备 initEquips[" + resolveCardIds(resolver, toUint32Slice(hero.InitEquips)) + "]/exEquips[" + resolveCardIds(resolver, toUint32Slice(hero.ExEquips)) + "]，")
		sb.WriteString("augurCards[" + resolveCardIds(resolver, toUint32Slice(hero.AugurCards)) + "]")
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// resolveHeroName 解析武将ID为"名字(ID)"，解析失败写 "?ID"。
func resolveHeroName(resolver *nameResolver, heroId uint32) string {
	name := resolver.heroName(heroId)
	if name == "" {
		return fmt.Sprintf("?%d", heroId)
	}
	return fmt.Sprintf("%s(%d)", name, heroId)
}

// resolveCardName 解析卡牌ID为"名字(ID)"，解析失败写 "?ID"。
func resolveCardName(resolver *nameResolver, cardId int) string {
	name := resolver.cardName(cardId)
	if name == "" {
		return fmt.Sprintf("?%d", cardId)
	}
	return fmt.Sprintf("%s(%d)", name, cardId)
}

// resolveCardIds 批量解析卡牌ID列表，逗号分隔。
func resolveCardIds(resolver *nameResolver, ids []uint32) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, resolveCardName(resolver, int(id)))
	}
	return strings.Join(parts, ", ")
}

// resolveSkillIds 批量解析技能ID列表，逗号分隔。
func resolveSkillIds(resolver *nameResolver, ids []uint32) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		name := resolver.skillName(int(id))
		if name == "" {
			parts = append(parts, fmt.Sprintf("?%d", id))
		} else {
			parts = append(parts, fmt.Sprintf("%s(%d)", name, id))
		}
	}
	return strings.Join(parts, ", ")
}

// toUint32Slice 把 []uint32 转成 []uint32（直接返回，占位以便后续扩展为通用接口）。
func toUint32Slice(in []uint32) []uint32 {
	if in == nil {
		return []uint32{}
	}
	return in
}

// ---------------------------------------------------------------------
// generateStepDesc —— 按 reverse-translate.md §2.2 模板反推单个 step
// ---------------------------------------------------------------------

// generateStepDesc 把单个 step 反推成一句中文，对照文档「步骤」列。
// 主语永远是"座位{robotIdx}({武将名})"，按 Action 类型套用不同模板。
func generateStepDesc(step Step, initYanWu InitYanWu, resolver *nameResolver) string {
	heroName := getHeroNameForStep(&initYanWu, step, resolver)
	subject := fmt.Sprintf("座位%d(%s)", step.RobotIdx, heroName)

	switch step.Action {
	case "PlayCard":
		targets := resolveTargets(initYanWu, step.TargetsId, resolver)
		cards := resolveCardIds(resolver, step.Cards)
		if targets != "" {
			return fmt.Sprintf("%s 对 %s 打出 %s", subject, targets, cards)
		}
		return fmt.Sprintf("%s 打出 %s", subject, cards)
	case "OptRoomAction":
		confirmStr := "false 不出"
		if step.Confirm {
			confirmStr = "true 出牌"
		}
		cards := resolveCardIds(resolver, step.Cards)
		s := fmt.Sprintf("%s 响应(confirm=%s)", subject, confirmStr)
		if step.Confirm && cards != "" {
			s += "，打出 " + cards
		}
		if step.HeroSkillUUID != 0 {
			s += "，用技能" + resolveSkillForStep(resolver, step.HeroSkillUUID)
		}
		return s
	case "UseHeroSkill":
		targets := resolveTargets(initYanWu, step.TargetsId, resolver)
		cards := resolveCardIds(resolver, step.Cards)
		s := fmt.Sprintf("%s 发动技能 %s", subject, resolveSkillForStep(resolver, step.HeroSkillUUID))
		if targets != "" {
			s += "，目标 " + targets
		}
		if cards != "" {
			s += "，打出 " + cards
		}
		if step.TransCardSkill != 0 {
			s += " (转化)"
		}
		return s
	case "DisCard":
		cards := resolveCardIds(resolver, step.Cards)
		return fmt.Sprintf("%s 弃牌 %s", subject, cards)
	case "PlayCardOver":
		return subject + " 结束出牌阶段"
	case "OnlyAsset":
		return subject + " 空等待（仅校验断言，不发协议）"
	case "Sleep":
		return fmt.Sprintf("%s 等待 %g 秒", subject, step.SleepTime)
	case "WaitAction":
		return subject + " 等待轮到自己行动"
	default:
		return fmt.Sprintf("%s 未知动作(%s)", subject, step.Action)
	}
}

// resolveTargets 解析 targetsId 列表为"座位N(武将名)"，逗号分隔。
func resolveTargets(initYanWu InitYanWu, targetsId []int, resolver *nameResolver) string {
	if len(targetsId) == 0 {
		return ""
	}
	// targetsId 为空或 [0] 时省略（0 表示无目标）
	parts := make([]string, 0, len(targetsId))
	for _, t := range targetsId {
		if t == 0 {
			continue
		}
		var heroName string
		if t >= 1 && t <= len(initYanWu.CustomHeroes) {
			heroName = resolver.heroName(initYanWu.CustomHeroes[t-1].HeroId)
		}
		parts = append(parts, fmt.Sprintf("座位%d(%s)", t, heroName))
	}
	return strings.Join(parts, ", ")
}

// resolveSkillForStep 解析 step 的 heroSkillUuid。
// 值 >= 3000000 视为运行时 UUID，反推为 "技能(UUID:{id})"；否则查技能配置反推为"名字(ID)"。
func resolveSkillForStep(resolver *nameResolver, heroSkillUUID uint32) string {
	if heroSkillUUID >= 3000000 {
		return fmt.Sprintf("技能(UUID:%d)", heroSkillUUID)
	}
	name := resolver.skillName(int(heroSkillUUID))
	if name == "" {
		return fmt.Sprintf("?%d", heroSkillUUID)
	}
	return fmt.Sprintf("%s(%d)", name, heroSkillUUID)
}

// ---------------------------------------------------------------------
// ReverseTranslateCase —— 三维度反推入口
// ---------------------------------------------------------------------

// ReverseTranslateCase 把整个 QAFuncCase 反推成三维度中文描述。
// 返回三段文本：前置条件反推（initYanWu）、步骤反推（steps）、预期结果反推（assets 逐条）。
//
// 流程：
//  1. initYanWu 反推 → 一段文本
//  2. 每个 step 反推 → 编号列表；同时把该 step 下每条 asset 反推成 "- 描述"
//  3. 三段用固定标题分隔，供 AI 对比环节使用
func ReverseTranslateCase(qaFuncCase QAFuncCase, resolver *nameResolver) string {
	var sb strings.Builder

	// ===== 前置条件反推（initYanWu）=====
	sb.WriteString("===== 前置条件反推（initYanWu）=====\n")
	sb.WriteString(generateInitYanWuDesc(qaFuncCase.InitYanWu, resolver))
	sb.WriteString("\n\n")

	// ===== 步骤反推（steps）=====
	sb.WriteString("===== 步骤反推（steps）=====\n")
	for i, step := range qaFuncCase.Steps {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, generateStepDesc(step, qaFuncCase.InitYanWu, resolver))
	}
	sb.WriteString("\n")

	// ===== 预期结果反推（assets 逐条）=====
	sb.WriteString("===== 预期结果反推（assets）=====\n")
	for i, step := range qaFuncCase.Steps {
		if len(step.Assets) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "---- 步骤 %d 的断言 ----\n", i+1)
		for _, asset := range step.Assets {
			desc := generateAiDesc(asset, step, &qaFuncCase.InitYanWu, resolver)
			// 反向断言前缀（UnExpect=true）
			if asset.Attr != nil && asset.Attr["UnExpect"] == "true" {
				desc = "不期望：" + desc
			}
			sb.WriteString("- " + desc + "\n")
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
