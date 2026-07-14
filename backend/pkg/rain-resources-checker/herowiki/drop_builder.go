// herowiki/drop_builder.go - 修正版

package herowiki

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
)

// DropInfoBuilder 掉落信息构建器
type DropInfoBuilder struct {
	dropRules  map[int]drop_rule.DropRuleDiff
	dropGroups map[int]drop_group.DropGroupDiff
	dropItems  map[int]drop_item.DropItemDiff

	// 反向索引
	heroToDropItems  map[string][]int // EHeroId -> []DropItemId
	heroToDropRules  map[string][]int // EHeroId -> []DropRuleId
	heroToDropGroups map[string][]int // EHeroId -> []DropGroupId

	// 英雄ID映射（用于快速查找）
	heroIdMap map[string]string // 数字ID字符串 -> EHeroId
}

// NewDropInfoBuilder 创建掉落信息构建器
func NewDropInfoBuilder(container *diff.DataContainer, heroIdMap map[string]string) *DropInfoBuilder {
	builder := &DropInfoBuilder{
		dropRules:        make(map[int]drop_rule.DropRuleDiff),
		dropGroups:       make(map[int]drop_group.DropGroupDiff),
		dropItems:        make(map[int]drop_item.DropItemDiff),
		heroToDropItems:  make(map[string][]int),
		heroToDropRules:  make(map[string][]int),
		heroToDropGroups: make(map[string][]int),
		heroIdMap:        heroIdMap,
	}

	// 建立索引
	builder.buildIndexes(container)

	return builder
}

// buildIndexes 建立所有索引
func (b *DropInfoBuilder) buildIndexes(container *diff.DataContainer) {
	// 索引掉落规则
	for _, rule := range *container.DropRuleDiff {
		b.dropRules[rule.Id] = rule
	}

	// 索引掉落组
	for _, group := range *container.DropGroupDiff {
		b.dropGroups[group.Id] = group
	}

	// 索引掉落项并建立英雄关联
	for _, item := range *container.DropItemDiff {
		b.dropItems[item.Id] = item

		// 解析掉落项中的物品（直接使用已有的ItemCfg结构）
		for _, itemCfg := range item.Item {
			heroId := b.extractHeroIdFromItemCfg(itemCfg)
			if heroId != "" {
				b.heroToDropItems[heroId] = append(b.heroToDropItems[heroId], item.Id)
				b.heroToDropGroups[heroId] = append(b.heroToDropGroups[heroId], item.DropGroup)
			}
		}
	}

	// 建立掉落规则到英雄的关联
	for ruleId, rule := range b.dropRules {
		// 检查普通掉落组
		for _, groupId := range rule.DropGroup {
			b.addHeroesFromGroupToRule(groupId, ruleId, false)
		}

		// 检查保底掉落组
		for _, groupId := range rule.EnsureSmallGroup {
			b.addHeroesFromGroupToRule(groupId, ruleId, true)
		}
		for _, groupId := range rule.EnsureBigGroup {
			b.addHeroesFromGroupToRule(groupId, ruleId, true)
		}

		// 检查保底物品
		if rule.EnsureItemID > 0 {
			if heroId := b.extractHeroIdFromItemId(rule.EnsureItemID); heroId != "" {
				b.heroToDropRules[heroId] = append(b.heroToDropRules[heroId], ruleId)
			}
		}
	}
}

// extractHeroIdFromItemCfg 从ItemCfg中提取英雄ID
func (b *DropInfoBuilder) extractHeroIdFromItemCfg(itemCfg drop_item.ItemCfg) string {
	itemIdStr := strconv.Itoa(itemCfg.ItemId)
	return b.extractHeroIdFromItemIdStr(itemIdStr)
}

// extractHeroIdFromItemId 从物品ID提取英雄ID
func (b *DropInfoBuilder) extractHeroIdFromItemId(itemId int) string {
	itemIdStr := strconv.Itoa(itemId)
	return b.extractHeroIdFromItemIdStr(itemIdStr)
}

// extractHeroIdFromItemIdStr 从物品ID字符串提取英雄ID
func (b *DropInfoBuilder) extractHeroIdFromItemIdStr(itemIdStr string) string {
	// 匹配规则：去掉前缀10开头，剩下的如果是英雄ID则返回
	if len(itemIdStr) > 2 && strings.HasPrefix(itemIdStr, "10") {
		heroNumStr := itemIdStr[2:]
		// 通过heroIdMap查找对应的EHeroId
		if eHeroId, exists := b.heroIdMap[heroNumStr]; exists {
			return eHeroId
		}
	}
	return ""
}

// addHeroesFromGroupToRule 从掉落组添加英雄到规则关联
func (b *DropInfoBuilder) addHeroesFromGroupToRule(groupId int, ruleId int, isGuarantee bool) {
	// 查找该掉落组下的所有掉落项
	for _, item := range b.dropItems {
		if item.DropGroup == groupId {
			for _, itemCfg := range item.Item {
				if heroId := b.extractHeroIdFromItemCfg(itemCfg); heroId != "" {
					b.heroToDropRules[heroId] = append(b.heroToDropRules[heroId], ruleId)
				}
			}
		}
	}
}

// BuildHeroDropInfo 构建单个武将的掉落信息
func (b *DropInfoBuilder) BuildHeroDropInfo(heroId string, heroName string) *herowiki_def.HeroDropInfo {
	info := &herowiki_def.HeroDropInfo{
		DirectDropRules:    make([]*herowiki_def.DropRuleSummary, 0),
		GuaranteeDropRules: make([]*herowiki_def.DropRuleSummary, 0),
		DropGroups:         make([]*herowiki_def.DropGroupSummary, 0),
		ByDropType:         make(map[string]*herowiki_def.DropTypeInfo),
	}

	// 处理掉落规则
	ruleIds := b.heroToDropRules[heroId]
	ruleMap := make(map[int]bool)

	for _, ruleId := range ruleIds {
		if ruleMap[ruleId] {
			continue
		}
		ruleMap[ruleId] = true

		if rule, exists := b.dropRules[ruleId]; exists {
			summary := b.buildRuleSummary(rule, heroId)
			if summary.IsGuarantee {
				info.GuaranteeDropRules = append(info.GuaranteeDropRules, summary)
			} else {
				info.DirectDropRules = append(info.DirectDropRules, summary)
			}
		}
	}

	// 处理掉落组
	groupIds := b.heroToDropGroups[heroId]
	groupMap := make(map[int]bool)

	for _, groupId := range groupIds {
		if groupMap[groupId] {
			continue
		}
		groupMap[groupId] = true

		if group, exists := b.dropGroups[groupId]; exists {
			groupSummary := b.buildGroupSummary(group, heroId)
			info.DropGroups = append(info.DropGroups, groupSummary)
		}
	}

	// 按类型分类
	info.ByDropType["direct"] = &herowiki_def.DropTypeInfo{
		TypeName:   "直接掉落",
		DropRules:  info.DirectDropRules,
		TotalCount: len(info.DirectDropRules),
	}

	info.ByDropType["guarantee"] = &herowiki_def.DropTypeInfo{
		TypeName:   "保底掉落",
		DropRules:  info.GuaranteeDropRules,
		TotalCount: len(info.GuaranteeDropRules),
	}

	return info
}

// buildRuleSummary 构建规则摘要
func (b *DropInfoBuilder) buildRuleSummary(rule drop_rule.DropRuleDiff, heroId string) *herowiki_def.DropRuleSummary {
	summary := &herowiki_def.DropRuleSummary{
		RuleId:      rule.Id,
		RuleName:    rule.Name,
		DropCount:   rule.Count,
		DropGroups:  rule.DropGroup,
		IsGuarantee: false,
	}

	// 检查是否在保底中
	for _, groupId := range rule.EnsureSmallGroup {
		if b.groupContainsHero(groupId, heroId) {
			summary.IsGuarantee = true
			break
		}
	}

	if !summary.IsGuarantee {
		for _, groupId := range rule.EnsureBigGroup {
			if b.groupContainsHero(groupId, heroId) {
				summary.IsGuarantee = true
				break
			}
		}
	}

	// 检查保底物品
	if rule.EnsureItemID > 0 {
		if extractedId := b.extractHeroIdFromItemId(rule.EnsureItemID); extractedId == heroId {
			summary.IsGuarantee = true
		}
	}

	return summary
}

// buildGroupSummary 构建组摘要
func (b *DropInfoBuilder) buildGroupSummary(group drop_group.DropGroupDiff, heroId string) *herowiki_def.DropGroupSummary {
	summary := &herowiki_def.DropGroupSummary{
		GroupId:     group.Id,
		GroupName:   group.Name,
		Weight:      group.Weight,
		WeightInc:   group.WeightInc,
		Deduplicate: group.Deduplication,
		ValidDate:   group.ValidDate,
		ExpireDate:  group.ExpireDate,
		DropItems:   make([]*herowiki_def.HeroDropItem, 0),
	}

	// 查找该组中涉及此英雄的掉落项
	for _, itemId := range b.heroToDropItems[heroId] {
		if item, exists := b.dropItems[itemId]; exists && item.DropGroup == group.Id {
			dropItem := b.buildHeroDropItem(item, heroId)
			if dropItem != nil {
				summary.DropItems = append(summary.DropItems, dropItem)
			}
		}
	}

	return summary
}

// buildHeroDropItem 构建英雄掉落项
func (b *DropInfoBuilder) buildHeroDropItem(item drop_item.DropItemDiff, heroId string) *herowiki_def.HeroDropItem {
	dropItem := &herowiki_def.HeroDropItem{
		ItemId:       item.Id,
		ItemName:     item.Name,
		DropGroupId:  item.DropGroup,
		ItemConfigs:  make([]*herowiki_def.ItemConfig, 0),
		Weight:       item.Weight,
		WeightInc:    item.WeightInc,
		Deduplicate:  item.Deduplication,
		CheckExist:   item.CheckExist,
		ExcludeExist: item.ExcludeExist,
		MustHave:     item.MustHave,
		ReplaceGroup: item.ReplaceGroup,
		ValidDate:    item.ValidDate,
		ExpireDate:   item.ExpireDate,
	}

	// 直接使用已有的ItemCfg结构
	for _, itemCfg := range item.Item {
		config := &herowiki_def.ItemConfig{
			ItemId: itemCfg.ItemId,
			Count:  itemCfg.Count,
			IsHero: false,
		}

		// 检查是否是英雄
		itemIdStr := strconv.Itoa(itemCfg.ItemId)
		if extractedId := b.extractHeroIdFromItemIdStr(itemIdStr); extractedId == heroId {
			config.IsHero = true
			config.HeroId = heroId
		}

		dropItem.ItemConfigs = append(dropItem.ItemConfigs, config)
	}

	return dropItem
}

// groupContainsHero 检查掉落组是否包含指定英雄
func (b *DropInfoBuilder) groupContainsHero(groupId int, heroId string) bool {
	for _, item := range b.dropItems {
		if item.DropGroup == groupId {
			for _, itemCfg := range item.Item {
				if extractedId := b.extractHeroIdFromItemCfg(itemCfg); extractedId == heroId {
					return true
				}
			}
		}
	}
	return false
}
