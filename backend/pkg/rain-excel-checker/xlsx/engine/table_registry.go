// Package engine 提供表级检查器管理功能
// 本包负责管理所有表级检查器的注册、获取和调度
package engine

import (
	crossactivity "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/cross_table/activity"
	crossdraw "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/cross_table/draw"
	crossdrop "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/cross_table/drop"
	crosshero "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/cross_table/hero"
	generalrules "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general"
	tablecheck "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/table"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// TableManager 全局表级检查器管理器实例
var TableManager = NewTableCheckerManager()

// TableCheckerManager 表级检查器管理器
// 负责管理所有表级检查器的注册和获取
type TableCheckerManager struct {
	checkerMap map[json_rule.ETableRule]TableChecker // 检查器映射表
}

// NewTableCheckerManager 创建新的表级检查器管理器
// 返回一个初始化完成的表级检查器管理器实例
func NewTableCheckerManager() *TableCheckerManager {
	return &TableCheckerManager{
		checkerMap: make(map[json_rule.ETableRule]TableChecker),
	}
}

// Reg 注册表级检查器
// 同时将检查器的元数据注册到全局元数据映射中
// 参数:
//   - rule: 表级规则类型枚举
//   - checker: 对应的表级检查器实例
func (cm *TableCheckerManager) Reg(rule json_rule.ETableRule, checker TableChecker) {
	cm.checkerMap[rule] = checker
	// 同时注册元数据，方便前端和 MCP 获取规则信息
	json_rule.TableRuleMetas[rule] = checker.Meta()
}

// GetChecker 获取指定类型的表级检查器
// 参数:
//   - rule: 表级规则类型枚举
//
// 返回:
//   - 表级检查器实例，如果不存在则返回 nil
func (cm *TableCheckerManager) GetChecker(rule json_rule.ETableRule) TableChecker {
	ck, exist := cm.checkerMap[rule]
	if !exist {
		return nil
	}
	return ck
}

// GetAllMetas 获取所有表级规则元数据
// 返回:
//   - 所有已注册表级规则的元数据列表
func (cm *TableCheckerManager) GetAllMetas() []*json_rule.TableRuleMeta {
	metas := make([]*json_rule.TableRuleMeta, 0, len(cm.checkerMap))
	for _, checker := range cm.checkerMap {
		metas = append(metas, checker.Meta())
	}
	return metas
}

// init 包初始化函数
// 自动注册所有表级检查器到管理器中
// 包括：
//   - 竞技场赛季检查
//   - 通用通知规则（新增行、行变更）
//   - 武将开放时间检查（战令、大将军）
//   - 武将掉落、合成、熔炼检查
func init() {
	// 竞技场赛季检查
	TableManager.Reg(json_rule.ARENA_SEASON_CHECK, new(tablecheck.ArenaSeasonCheckRule))

	// 通用通知规则（旧枚举，保留向后兼容）
	TableManager.Reg(json_rule.NEW_ROW_NOTIFY, new(generalrules.NewRowNotifyRule))
	TableManager.Reg(json_rule.ROW_CHANGE_NOTIFY, new(generalrules.RowChangeNotifyRule))

	// 拆分后的通用通知规则（5 个独立规则）
	TableManager.Reg(json_rule.ADDED_ROW_NOTIFY, new(generalrules.AddedRowNotifyRule))
	TableManager.Reg(json_rule.REMOVED_ROW_NOTIFY, new(generalrules.RemovedRowNotifyRule))
	TableManager.Reg(json_rule.ADDED_COL_NOTIFY, new(generalrules.AddedColNotifyRule))
	TableManager.Reg(json_rule.REMOVED_COL_NOTIFY, new(generalrules.RemovedColNotifyRule))
	TableManager.Reg(json_rule.MODIFIED_ROW_NOTIFY, new(generalrules.ModifiedRowNotifyRule))

	// 武将开放时间检查规则
	TableManager.Reg(json_rule.SEASON_PASS_HERO_OPEN_CHECK, new(tablecheck.SeasonPassHeroOpenCheckRule))
	TableManager.Reg(json_rule.ARENA_GENERAL_HERO_OPEN_CHECK, new(tablecheck.ArenaGeneralHeroOpenCheckRule))

	// 武将掉落检查规则（跨表）
	TableManager.Reg(json_rule.HERO_DROP_CHECK, new(crosshero.HeroDropCheckRule))

	// 武将合成检查规则（跨表）— 挂载在 Item 表上（IsSynthetic 字段在 Item 表中）
	TableManager.Reg(json_rule.HERO_SYNTHESIS_CHECK, new(crosshero.ItemSynthesisCheckRule))

	// 武将熔炼检查规则（跨表）
	TableManager.Reg(json_rule.HERO_MELT_CHECK, new(crosshero.HeroMeltCheckRule))

	// 活动任务奖励检查规则（跨表）
	// TODO: ActivityTaskRewardCheckRule 尚未实现
	// TableManager.Reg(json_rule.ACTIVITY_TASK_REWARD_CHECK, new(crossactivity.ActivityTaskRewardCheckRule))

	// 掉落道具必须在道具表中检查规则（跨表）
	TableManager.Reg(json_rule.DROP_ITEM_MUST_IN_ITEM_CHECK, new(crossdrop.DropItemMustInItemCheckRule))

	// 皮肤抽奖副产品检查规则（跨表）
	TableManager.Reg(json_rule.DRAWSKIN_BYPRODUCT_CHECK, new(crossdraw.DrawskinByproductCheckRule))

	// 丹青阁活动时间校验规则
	TableManager.Reg(json_rule.DANQINGGE_TIME_ACTIVE_CHECK, new(tablecheck.DanQingGeTimeActiveCheckRule))

	// 丹青阁活动自定义参数检查规则（跨表）
	TableManager.Reg(json_rule.DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK, new(crossactivity.ActivityDanqinggeCustomParamIsItemidCheckRule))

	// 武将技能和Buff引用完整性检查规则（跨表）
	TableManager.Reg(json_rule.HERO_SKILL_BUFF_CHECK, new(crosshero.HeroSkillBuffCheckRule))

	// 抽奖池掉落规则引用检查规则（跨表）
	TableManager.Reg(json_rule.DRAW_DROP_RULE_REFERENCE_CHECK, new(crossdraw.DrawDropRuleReferenceCheckRule))

	// 掉落规则组ID引用检查规则（跨表）
	TableManager.Reg(json_rule.DROP_RULE_GROUP_ID_CHECK, new(crossdrop.DropRuleGroupIdCheckRule))

	// 丹青阁活动时间配置检查（ACT-04/05/06）
	TableManager.Reg(json_rule.ACTIVITY_DANQINGGE_TIME_CONFIG_CHECK, new(tablecheck.ActivityDanqinggeTimeConfigCheckRule))

	// DrawSkin 单抽消耗检查（DSK-06/07/08）
	TableManager.Reg(json_rule.DRAWSKIN_ONCE_ITEM_COST_CHECK, new(crossdraw.DrawskinOnceItemCostCheckRule))

	// DrawSkin 基础数据检查（DSK-01/12）
	TableManager.Reg(json_rule.DRAWSKIN_DATA_VALIDITY_CHECK, new(tablecheck.DrawskinDataValidityCheckRule))

	// DrawSkin 十连消耗检查（DSK-09）
	TableManager.Reg(json_rule.DRAWSKIN_TEN_COST_CHECK, new(tablecheck.DrawskinTenCostCheckRule))

	// DrawSkin 时间范围检查（DSK-10/11）
	TableManager.Reg(json_rule.DRAWSKIN_TIME_RANGE_CHECK, new(tablecheck.DrawskinTimeRangeCheckRule))

	// Activity 与 DrawSkin 交叉引用检查（DSK-05/XC-01/XC-04）
	TableManager.Reg(json_rule.ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK, new(crossactivity.ActivityDrawskinCrossReferenceCheckRule))

	// Activity 与 DrawSkin 时间交集检查（XC-02）
	TableManager.Reg(json_rule.ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK, new(crossactivity.ActivityDrawskinTimeOverlapCheckRule))

	// DropRule 基础数据检查（Count>0、DropGroup非空、EnsureItemCount<=Count、EnsureSmall<EnsureBig）
	TableManager.Reg(json_rule.DROP_RULE_DATA_VALIDITY_CHECK, new(tablecheck.DropRuleDataValidityCheckRule))

	// DropRule 条件引用检查（保底组条件检查、EnsureItem条件检查）
	TableManager.Reg(json_rule.DROP_RULE_CONDITIONAL_CHECK, new(crossdrop.DropRuleConditionalCheckRule))

	// DropItem 条件和互斥检查（ReplaceGroup引用、布尔互斥、Item道具有效性）
	TableManager.Reg(json_rule.DROP_ITEM_VALIDITY_CHECK, new(crossdrop.DropItemValidityCheckRule))

	// 日期有效性检查（ValidDate <= ExpireDate）
	TableManager.Reg(json_rule.DATE_VALID_EXPIRE_CHECK, new(tablecheck.DateValidExpireCheckRule))

	// 列连续性检查（严格递增、单调递增、日期间隔一致、ID递增、提取数字递增、拆分后全局唯一）
	TableManager.Reg(json_rule.COL_CONTINUOUS_CHECK, new(tablecheck.ColContinuousCheckRule))

	// 定向招募战令保护期检查规则（跨表）
	TableManager.Reg(json_rule.DRAWFIX_PROTECTION_CHECK, new(crossdraw.DrawFixProtectionCheckRule))

	// 定向招募大将军保护期检查规则（跨表）
	TableManager.Reg(json_rule.DRAWFIX_ARENA_PROTECTION_CHECK, new(crossdraw.DrawFixArenaProtectionCheckRule))

	// 武将 IsOpen 与 OpenDate 一致性检查
	TableManager.Reg(json_rule.HERO_ISOPEN_OPENDATE_CHECK, new(tablecheck.HeroIsOpenOpenDateCheckRule))

	// 普通武将掉落时间检查（跨表）— 挂载在 DropItem 表（ValidDate 字段在 DropItem 表中）
	TableManager.Reg(json_rule.HERO_DROP_VALIDDATE_CHECK, new(crosshero.HeroDropValidDateCheckRule))

	// 战令高级奖励道具引用完整性检查（跨表）— 挂载在 SeasonPassReward 表，带赛季开始前N天时间门控
	TableManager.Reg(json_rule.SEASON_PASS_REWARD_INTEGRITY_CHECK, new(crossactivity.SeasonPassRewardIntegrityCheckRule))
}
