package herowiki

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/arena_score_rewards"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/buff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/task_complete_cond"
)

// 扩展HeroWikiDiff结构，添加全局数据
type GlobalData struct {
	// 掉落系统
	DropGroups map[int]*DropGroupData
	DropItems  map[int]*DropItemData
	DropRules  map[int]*DropRuleData

	// 奖励系统
	SeasonPassRewards map[int]*SeasonPassRewardData
	ArenaRewards      map[int]*ArenaRewardData

	// Buff系统
	Buffs map[string]*BuffData

	// 任务系统
	Tasks map[int]*TaskData
}

// DropGroupData 掉落组数据
type DropGroupData struct {
	Basic     drop_group.DropGroupDiff
	DropItems []*drop_item.DropItemDiff
	DropRules []*drop_rule.DropRuleDiff
}

// SeasonPassRewardData 赛季通行证奖励数据
type SeasonPassRewardData struct {
	Basic       season_pass_reward.SeasonPassRewardDiff
	RewardItems []*ItemDetail
}

// ItemDetail 物品详情
type ItemDetail struct {
	ItemId     int
	ItemType   string // "hero", "skin", "other"
	HeroId     string // 如果是英雄相关物品
	SkinItemId int    // 如果是皮肤相关物品
}

// BuildGlobalData 构建全局数据
func BuildGlobalData(container *diff.DataContainer) *GlobalData {
	global := &GlobalData{
		DropGroups:        make(map[int]*DropGroupData),
		DropItems:         make(map[int]*DropItemData),
		DropRules:         make(map[int]*DropRuleData),
		SeasonPassRewards: make(map[int]*SeasonPassRewardData),
		ArenaRewards:      make(map[int]*ArenaRewardData),
		Buffs:             make(map[string]*BuffData),
		Tasks:             make(map[int]*TaskData),
	}

	// 构建掉落组数据
	for _, dropGroup := range *container.DropGroupDiff {
		global.DropGroups[dropGroup.Id] = &DropGroupData{
			Basic: dropGroup,
		}
	}

	// 关联掉落物品到掉落组
	for _, dropItem := range *container.DropItemDiff {
		if group, ok := global.DropGroups[dropItem.DropGroup]; ok {
			group.DropItems = append(group.DropItems, &dropItem)
		}
		global.DropItems[dropItem.Id] = &DropItemData{Basic: dropItem}
	}

	// 关联掉落规则
	for _, dropRule := range *container.DropRuleDiff {
		global.DropRules[dropRule.Id] = &DropRuleData{Basic: dropRule}

		// 关联掉落规则中的掉落组
		for _, groupId := range dropRule.DropGroup {
			if group, ok := global.DropGroups[groupId]; ok {
				group.DropRules = append(group.DropRules, &dropRule)
			}
		}
	}

	// 构建赛季通行证奖励数据
	for _, reward := range *container.SeasonPassRewardDiff {
		data := &SeasonPassRewardData{
			Basic: reward,
		}

		// 解析奖励物品（ItemCfg结构体，取ItemId字段）
		for _, item := range reward.NormalReward {
			data.RewardItems = append(data.RewardItems, resolveItem(item.ItemId, container))
		}
		for _, item := range reward.HighReward {
			data.RewardItems = append(data.RewardItems, resolveItem(item.ItemId, container))
		}

		global.SeasonPassRewards[reward.Id] = data
	}

	// 构建竞技场奖励数据
	for _, reward := range *container.ArenaScoreRewardsDiff {
		data := &ArenaRewardData{
			Basic: reward,
		}

		for _, item := range reward.Reward {
			_ = item
			//data.RewardItems = append(data.RewardItems, resolveItem(item.ID, container))
		}

		global.ArenaRewards[reward.Id] = data
	}

	return global
}

// resolveItem 解析物品ID对应的具体物品
func resolveItem(itemId int, container *diff.DataContainer) *ItemDetail {
	detail := &ItemDetail{
		ItemId: itemId,
	}

	// 检查是否是皮肤物品
	for _, skin := range *container.HeroSkinItemDiff {
		if skin.SkinItemId == itemId {
			detail.ItemType = "skin"
			detail.HeroId = skin.HeroId
			detail.SkinItemId = skin.SkinItemId
			return detail
		}
	}

	// 检查是否是英雄
	for _, hero := range *container.HeroDiff {
		if hero.Id == itemId {
			detail.ItemType = "hero"
			detail.HeroId = hero.EHeroId
			return detail
		}
	}

	// 其他类型物品
	detail.ItemType = "other"
	return detail
}

// DropItemData 掉落物品数据
type DropItemData struct {
	Basic       drop_item.DropItemDiff
	ItemDetails []*ItemDetail
}

// DropRuleData 掉落规则数据
type DropRuleData struct {
	Basic      drop_rule.DropRuleDiff
	DropGroups []*DropGroupData
}

// BuffData Buff数据
type BuffData struct {
	Basic         buff.BuffDiff
	RelatedSkills []string // 使用该Buff的技能
}

// ArenaRewardData 竞技场奖励数据
type ArenaRewardData struct {
	Basic       arena_score_rewards.ArenaScoreRewardDiff
	RewardItems []*ItemDetail
}

// TaskData 任务数据
type TaskData struct {
	Basic           task_complete_cond.TaskCompleteConditonDiff
	RelatedAchieves []int // 关联的成就ID
}
