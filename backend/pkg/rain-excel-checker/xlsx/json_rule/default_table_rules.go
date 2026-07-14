// Package json_rule 提供 Excel 配表校验规则的类型定义
// 本包定义了所有表级和列级规则的数据结构、枚举类型和元数据
package json_rule

import "strings"

// DefaultTableRules 默认表级规则配置
// 这些规则会自动应用于指定的表，无需在 JSON 配置文件中手动配置
// Key 是表名的英文后缀（如 "SeasonPassReward"），支持自动匹配 "中文|英文" 格式
var DefaultTableRules = map[string][]ETableRule{
	// 战令武将开放时间检查 - 自动应用于 SeasonPassReward 表
	"SeasonPassReward": {SEASON_PASS_HERO_OPEN_CHECK, SEASON_PASS_REWARD_INTEGRITY_CHECK},

	// 大将军武将开放时间检查 - 自动应用于 ArenaScoreReward 表
	// 注意：实际 sheet 名为 "竞技场积分奖励表|ArenaScoreReward"（单数）
	"ArenaScoreReward": {ARENA_GENERAL_HERO_OPEN_CHECK},

	// 竞技场赛季检查 - 自动应用于 ArenaSeason 表
	"ArenaSeason": {ARENA_SEASON_CHECK},

	// 武将相关检查 - 自动应用于 Hero 表
	// 包括：掉落/熔炼检查 + 战令/大将军武将开放时间检查 + 技能Buff检查 + IsOpen/OpenDate一致性检查
	// 注意：合成检查(HERO_SYNTHESIS_CHECK)已迁移到 Item 表，因为 IsSynthetic 字段在 Item 表中
	"Hero": {HERO_DROP_CHECK, HERO_MELT_CHECK,
		SEASON_PASS_HERO_OPEN_CHECK, ARENA_GENERAL_HERO_OPEN_CHECK, HERO_SKILL_BUFF_CHECK,
		HERO_ISOPEN_OPENDATE_CHECK},

	// 武将合成检查 - 自动应用于 Item 表
	// IsSynthetic 字段在 Item 表中配置，增量模式下修改 Item 表才触发
	"Item": {HERO_SYNTHESIS_CHECK},

	// 活动任务奖励检查 - 自动应用于 ActivityTask 表
	// 检查活动任务中配置的奖励道具是否存在于 Item 表中且数量大于0
	"ActivityTask": {ACTIVITY_TASK_REWARD_CHECK},

	// 掉落道具检查 - 自动应用于 DropItem 表
	// 检查掉落道具表中配置的道具是否存在于 Item 表中
	"DropItem": {DROP_ITEM_MUST_IN_ITEM_CHECK, DROP_ITEM_VALIDITY_CHECK, DATE_VALID_EXPIRE_CHECK, HERO_DROP_VALIDDATE_CHECK},

	// 掉落分组检查 - 自动应用于 DropGroup 表
	// 检查掉落分组表的日期有效性（ValidDate <= ExpireDate）
	"DropGroup": {DATE_VALID_EXPIRE_CHECK},

	// 皮肤抽奖副产品检查 - 自动应用于 DrawSkin 表
	// 检查皮肤抽奖表中配置的副产品道具ID是否存在于 Item 表中
	"DrawSkin": {
		DRAWSKIN_BYPRODUCT_CHECK, DRAW_DROP_RULE_REFERENCE_CHECK,
		DRAWSKIN_DATA_VALIDITY_CHECK, DRAWSKIN_TIME_RANGE_CHECK, DRAWSKIN_TEN_COST_CHECK,
		DRAWSKIN_ONCE_ITEM_COST_CHECK,
		ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK, ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK,
	},

	// 名将册掉落规则引用检查 - 自动应用于 Draw 表
	"Draw": {DRAW_DROP_RULE_REFERENCE_CHECK},

	// 定向招募战令保护期检查 - 自动应用于 DrawFix 表
	// 检查定向招募表中是否错误配置了战令保护期内的武将
	"DrawFix": {DRAWFIX_PROTECTION_CHECK, DRAWFIX_ARENA_PROTECTION_CHECK},

	// 掉落规则组ID引用检查 - 自动应用于 DropRule 表
	// 检查掉落规则中引用的掉落组ID是否在掉落分组表中存在
	"DropRule": {DROP_RULE_GROUP_ID_CHECK},

	// 丹青阁活动相关检查 - 自动应用于 Activity 表
	// 时间校验：检查 ActivityType=ActTypeSkinRaffle 的丹青阁活动是否在生效时间范围内
	// 自定义参数检查：检查丹青阁活动的 CustomParma 值是否在 DrawSkin 表（皮肤抽奖表）中存在
	"Activity": {
		DANQINGGE_TIME_ACTIVE_CHECK, DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK,
		ACTIVITY_DANQINGGE_TIME_CONFIG_CHECK,
		ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK, ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK,
	},
}

// GetDefaultTableRulesForSheet 获取指定表的默认表级规则列表
// 支持以下匹配方式：
// 1. 精确匹配：sheetName 完全等于 Key
// 2. 后缀匹配：sheetName 为 "中文|Key" 格式时，能匹配到 Key
func GetDefaultTableRulesForSheet(sheetName string) []ETableRule {
	// 1. 精确匹配
	if rules, ok := DefaultTableRules[sheetName]; ok {
		return rules
	}

	// 2. 后缀匹配：检查 sheetName 是否以 "|Key" 结尾
	// 例如："赛季战令奖励表|SeasonPassReward" 应该匹配 "SeasonPassReward"
	for key, rules := range DefaultTableRules {
		// 检查是否以 "|key" 结尾
		suffix := "|" + key
		if strings.HasSuffix(sheetName, suffix) {
			return rules
		}
	}

	return nil
}

// HasDefaultTableRule 检查指定表是否有指定的默认表级规则
func HasDefaultTableRule(sheetName string, ruleType ETableRule) bool {
	rules := GetDefaultTableRulesForSheet(sheetName)
	for _, r := range rules {
		if r == ruleType {
			return true
		}
	}
	return false
}

// GeneralRuleOverrides 通用规则的 per-sheet 参数覆盖注册表
// Key 是表名的英文后缀（与 DefaultTableRules 相同的匹配规则）
// Value 是 map[ETableRule]map[参数名]参数值
//
// 同时设置 ID_COL_NAMES 和 ID_COL_NAME 的原因：
// ID_COL_NAMES 控制快照构建的内部行为，ID_COL_NAME 控制飞书通知的展示文案（如 "🔑 AnimationState"），
// 两者必须同步覆盖，否则通知会显示错误的 ID 列名
var GeneralRuleOverrides = map[string]map[ETableRule]map[string]string{
	"PetAudio": {
		NEW_ROW_NOTIFY: {
			string(ID_COL_NAMES): "AnimationState,ItemCfgId",
			string(ID_COL_NAME):  "AnimationState",
		},
		ROW_CHANGE_NOTIFY: {
			string(ID_COL_NAMES): "AnimationState,ItemCfgId",
			string(ID_COL_NAME):  "AnimationState",
		},
	},
	"PetTriggerWeight": {
		NEW_ROW_NOTIFY: {
			string(ID_COL_NAMES): "Id,ItemCfgId",
		},
		ROW_CHANGE_NOTIFY: {
			string(ID_COL_NAMES): "Id,ItemCfgId",
		},
	},
}

// GetGeneralRuleOverrides 获取指定表的通用规则参数覆盖
// 支持精确匹配和后缀匹配（与 GetDefaultTableRulesForSheet 相同规则）
// 同时支持新枚举自动回退到旧枚举的配置：
//   - ADDED_ROW_NOTIFY/REMOVED_ROW_NOTIFY/ADDED_COL_NOTIFY/REMOVED_COL_NOTIFY → 回退到 NEW_ROW_NOTIFY
//   - MODIFIED_ROW_NOTIFY → 回退到 ROW_CHANGE_NOTIFY
func GetGeneralRuleOverrides(sheetName string) map[ETableRule]map[string]string {
	// 1. 精确匹配
	if overrides, ok := GeneralRuleOverrides[sheetName]; ok {
		return expandOverridesWithFallback(overrides)
	}

	// 2. 后缀匹配
	for key, overrides := range GeneralRuleOverrides {
		suffix := "|" + key
		if strings.HasSuffix(sheetName, suffix) {
			return expandOverridesWithFallback(overrides)
		}
	}

	return map[ETableRule]map[string]string{}
}

// expandOverridesWithFallback 为新枚举补充回退配置
// 当 GeneralRuleOverrides 只配置了旧枚举（NEW_ROW_NOTIFY/ROW_CHANGE_NOTIFY）时，
// 自动将配置复制给对应的新枚举，避免需要为每个新枚举重复配置
func expandOverridesWithFallback(overrides map[ETableRule]map[string]string) map[ETableRule]map[string]string {
	result := make(map[ETableRule]map[string]string, len(overrides))

	// 复制原始配置
	for k, v := range overrides {
		result[k] = v
	}

	// 为新枚举补充回退配置（仅在新枚举未单独配置时回退）
	rowNotifyParams, hasRowNotify := overrides[NEW_ROW_NOTIFY]
	changeNotifyParams, hasChangeNotify := overrides[ROW_CHANGE_NOTIFY]

	// NEW_ROW_NOTIFY 的 4 个拆分规则
	if hasRowNotify {
		for _, newRule := range []ETableRule{ADDED_ROW_NOTIFY, REMOVED_ROW_NOTIFY, ADDED_COL_NOTIFY, REMOVED_COL_NOTIFY} {
			if _, exists := result[newRule]; !exists {
				// 复制一份参数（避免共享 map 引用）
				params := make(map[string]string, len(rowNotifyParams))
				for k, v := range rowNotifyParams {
					params[k] = v
				}
				result[newRule] = params
			}
		}
	}

	// ROW_CHANGE_NOTIFY → MODIFIED_ROW_NOTIFY
	if hasChangeNotify {
		if _, exists := result[MODIFIED_ROW_NOTIFY]; !exists {
			params := make(map[string]string, len(changeNotifyParams))
			for k, v := range changeNotifyParams {
				params[k] = v
			}
			result[MODIFIED_ROW_NOTIFY] = params
		}
	}

	return result
}
