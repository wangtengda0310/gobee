package activitywiki

import (
	"sort"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/activitywiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/accumulated_recharge"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_collition"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_spine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/item_hero_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/limit_skin_times_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet_audio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_bag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_task"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop_goods"
)

// BuildActivityWikiDiff 构建活动Wiki聚合数据
// 遍历所有活动，处理丹青阁、结缘亭、累充等活动类型
func BuildActivityWikiDiff(container *diff.DataContainer) *activitywiki_def.ActivityWikiDiff {
	wiki := &activitywiki_def.ActivityWikiDiff{
		Activities:   make(map[int]*activitywiki_def.ActivityCompleteData),
		SeasonPasses: make(map[int]*activitywiki_def.SeasonPassCompleteData),
	}

	if container.ActivityDiff == nil {
		return wiki
	}

	// 构建索引映射
	dropRuleById := buildDropRuleIndex(container.DropRuleDiff)
	dropGroupById := buildDropGroupIndex(container.DropGroupDiff)
	dropItemsByGroupId := buildDropItemsIndex(container.DropItemDiff)
	timesRewardsByActIdStr := buildTimesRewardsByStrIndex(container.LimitSkinTimesRewardDiff)
	shopByType := buildShopIndex(container.ShopDiff)
	shopGoodsByShopType := buildShopGoodsIndex(container.ShopGoodsDiff)
	heroSkinCollitionByType := buildHeroSkinCollitionIndex(container.HeroSkinCollectionDiff)
	itemHeroSkinById := buildItemHeroSkinIndex(container.ItemHeroSkinDiff)
	heroSkinItemById := buildHeroSkinItemIndex(container.HeroSkinItemDiff)
	heroSkinSpineById := buildHeroSkinSpineIndex(container.HeroSkinSpineDiff)
	drawPetById := buildDrawPetIndex(container.DrawPetDiff)
	petById := buildPetIndex(container.PetDiff)
	petAudioByItemCfgId := buildPetAudioIndex(container.PetAudioDiff)
	accumulatedRechargeByActId := buildAccumulatedRechargeByActIdIndex(container.AccumulatedRechargeDiff)
	// 战令相关索引
	seasonPassBagBySeasonId := buildSeasonPassBagIndex(container.SeasonPassBagDiff)
	seasonPassRewardBySeasonId := buildSeasonPassRewardIndex(container.SeasonPassRewardDiff)
	seasonPassTaskBySeasonId := buildSeasonPassTaskIndex(container.SeasonPassTaskDiff)

	sortedDrawSkins := buildAllSortedDrawSkins(container.DrawSkinDiff)
	sortedDrawPets := buildAllSortedDrawPets(container.DrawPetDiff)
	sortedSeasonPasses := buildAllSortedSeasonPasses(container.SeasonPassDiff)

	for i := range *container.ActivityDiff {
		act := &(*container.ActivityDiff)[i]
		data := &activitywiki_def.ActivityCompleteData{
			Basic: act,
		}

		// 丹青阁活动
		if act.ActivityType == "ActTypeSkinRaffle" {
			// 通过时间找当前进行中的 DrawSkin，而不是直接用 CustomParma[0]
			currentDs := findCurrentDrawSkin(sortedDrawSkins)
			if currentDs != nil {
				data.DrawSkin = currentDs
				prev, next := findPrevNextDrawSkin(sortedDrawSkins, currentDs.Id)
				data.PrevDrawSkin = prev
				data.NextDrawSkin = next

				// 关联单抽掉落规则
				if currentDs.OnceDropRule > 0 {
					linkDropRuleData(currentDs.OnceDropRule, data, dropRuleById, dropGroupById, dropItemsByGroupId, false)
				}
				// 关联十连掉落规则
				if currentDs.TenDropRule > 0 {
					linkDropRuleData(currentDs.TenDropRule, data, dropRuleById, dropGroupById, dropItemsByGroupId, true)
				}
			}

			if rewards, ok := timesRewardsByActIdStr[act.EActivityId]; ok {
				data.TimesRewards = rewards
			}
			if s, ok := shopByType["ShopTypeSkinRaffle"]; ok {
				data.Shop = s
			}
			if goods, ok := shopGoodsByShopType["ShopTypeSkinRaffle"]; ok {
				data.ShopGoods = goods
			}

			if data.DrawSkin != nil {
				bigAwardItemId := data.DrawSkin.BigAwardItemId
				if bigAwardItemId > 0 {
					if skin, ok := itemHeroSkinById[bigAwardItemId]; ok {
						data.ItemHeroSkin = skin
					}
					if skin, ok := heroSkinItemById[bigAwardItemId]; ok {
						data.HeroSkinItem = skin
					}
					if spine, ok := heroSkinSpineById[bigAwardItemId]; ok {
						data.HeroSkinSpine = spine
					}
					if data.HeroSkinItem != nil && data.HeroSkinItem.CollitionType != "" {
						if collition, ok := heroSkinCollitionByType[data.HeroSkinItem.CollitionType]; ok {
							data.HeroSkinCollition = collition
						}
					}
				}
			}

		} else if act.ActivityType == "ActTypeAccumulatedRecharge" {
			if rewards, ok := accumulatedRechargeByActId[act.EActivityId]; ok {
				data.AccumulatedRecharges = rewards
			}

		} else if act.ActivityType == "ActTypeDrawPet" {
			var relatedDrawPets []*draw_pet.DrawPetDiff
			if len(act.CustomParma) > 0 {
				if found, ok := drawPetById[act.CustomParma[0]]; ok {
					relatedDrawPets = append(relatedDrawPets, found)
				}
			} else {
				for _, dp := range sortedDrawPets {
					for _, aid := range dp.ActivityId {
						if aid == act.Id {
							relatedDrawPets = append(relatedDrawPets, dp)
							break
						}
					}
				}
			}

			currentDp := findCurrentDrawPet(relatedDrawPets)
			if currentDp != nil {
				data.DrawPet = currentDp
				prev, next := findPrevNextDrawPet(sortedDrawPets, currentDp.Id)
				data.PrevDrawPet = prev
				data.NextDrawPet = next

				// 关联单抽掉落规则
				if currentDp.OnceDropRule > 0 {
					linkDropRuleData(currentDp.OnceDropRule, data, dropRuleById, dropGroupById, dropItemsByGroupId, false)
				}
				// 关联十连掉落规则
				if currentDp.TenDropRule > 0 {
					linkDropRuleData(currentDp.TenDropRule, data, dropRuleById, dropGroupById, dropItemsByGroupId, true)
				}

				// 通过 PartnerItem.ItemId 关联 Pet 数据
				if currentDp.PartnerItem != nil && currentDp.PartnerItem.ItemId > 0 {
					petItemId := currentDp.PartnerItem.ItemId
					if p, ok := petById[petItemId]; ok {
						data.Pets = append(data.Pets, p)
					}
					if audios, ok := petAudioByItemCfgId[petItemId]; ok {
						data.PetAudios = append(data.PetAudios, audios...)
					}
				}
			}

			if s, ok := shopByType["ShopTypeDrawPet"]; ok {
				data.Shop = s
			}
			if goods, ok := shopGoodsByShopType["ShopTypeDrawPet"]; ok {
				data.ShopGoods = goods
			}
		}

		wiki.Activities[act.Id] = data
	}

	// 构建战令数据（战令没有对应的ActivityType，独立索引）
	// 只展示当前期战令（进行中的），如果没有则展示最新一期
	currentSeasonPass := findCurrentSeasonPass(sortedSeasonPasses)
	if currentSeasonPass != nil {
		sp := currentSeasonPass
		seasonPassData := &activitywiki_def.SeasonPassCompleteData{
			Basic: sp,
		}
		// 关联战令礼包
		if bags, ok := seasonPassBagBySeasonId[sp.Id]; ok {
			seasonPassData.Bags = bags
			seasonPassData.CurrentBags = bags
		}
		// 关联战令奖励
		if rewards, ok := seasonPassRewardBySeasonId[sp.Id]; ok {
			seasonPassData.Rewards = rewards
			seasonPassData.CurrentRewards = rewards
		}
		// 关联战令任务
		if tasks, ok := seasonPassTaskBySeasonId[sp.Id]; ok {
			seasonPassData.Tasks = tasks
			seasonPassData.CurrentTasks = tasks
		}
		// 设置上一期/下一期战令
		prev, next := findPrevNextSeasonPass(sortedSeasonPasses, sp.Id)
		seasonPassData.PrevSeasonPass = prev
		seasonPassData.NextSeasonPass = next

		wiki.SeasonPasses[sp.Id] = seasonPassData
	}

	return wiki
}

// linkDropRuleData 关联 DropRule 及其下属的 DropGroup 和 DropItem 到 ActivityCompleteData
// isTen=true 写入十连字段（TenDropRule/TenDropGroups/TenDropItems），否则写入单抽字段
func linkDropRuleData(
	ruleId int,
	data *activitywiki_def.ActivityCompleteData,
	dropRuleById map[int]*drop_rule.DropRuleDiff,
	dropGroupById map[int]*drop_group.DropGroupDiff,
	dropItemsByGroupId map[int][]*drop_item.DropItemDiff,
	isTen bool,
) {
	dr, ok := dropRuleById[ruleId]
	if !ok {
		return
	}
	if isTen {
		data.TenDropRule = dr
	} else {
		data.DropRule = dr
	}
	for _, dgId := range dr.DropGroup {
		if dg, ok := dropGroupById[dgId]; ok {
			if isTen {
				data.TenDropGroups = append(data.TenDropGroups, dg)
			} else {
				data.DropGroups = append(data.DropGroups, dg)
			}
			if items, ok := dropItemsByGroupId[dgId]; ok {
				if isTen {
					data.TenDropItems = append(data.TenDropItems, items...)
				} else {
					data.DropItems = append(data.DropItems, items...)
				}
			}
		}
	}
}

// buildAllSortedDrawSkins 构建 DrawSkin 全表按 StartTime 排序的切片
func buildAllSortedDrawSkins(drawSkinDiff *[]draw_skin.DrawSkinDiff) []*draw_skin.DrawSkinDiff {
	var sorted []*draw_skin.DrawSkinDiff
	if drawSkinDiff == nil {
		return sorted
	}
	for i := range *drawSkinDiff {
		sorted = append(sorted, &(*drawSkinDiff)[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime < sorted[j].StartTime
	})
	return sorted
}

// findCurrentDrawSkin 从排序后的 DrawSkin 列表中找到当前期
// 优先选择正在进行中的（StartTime <= now <= EndTime），如果没有则选最新一期
func findCurrentDrawSkin(sorted []*draw_skin.DrawSkinDiff) *draw_skin.DrawSkinDiff {
	if len(sorted) == 0 {
		return nil
	}
	now := time.Now()
	for _, ds := range sorted {
		start, err1 := time.Parse("2006-01-02 15:04:05", ds.StartTime)
		end, err2 := time.Parse("2006-01-02 15:04:05", ds.EndTime)
		if err1 == nil && err2 == nil && (now.Equal(start) || now.After(start)) && (now.Equal(end) || now.Before(end)) {
			return ds
		}
	}
	// 没有进行中的，返回最新一期（StartTime 最大的）
	return sorted[len(sorted)-1]
}

// findPrevNextDrawSkin 在按 StartTime 排序的切片中找到当前 DrawSkin 的前一个和后一个
func findPrevNextDrawSkin(sorted []*draw_skin.DrawSkinDiff, currentId int) (prev *draw_skin.DrawSkinDiff, next *draw_skin.DrawSkinDiff) {
	for i, ds := range sorted {
		if ds.Id == currentId {
			if i > 0 {
				prev = sorted[i-1]
			}
			if i < len(sorted)-1 {
				next = sorted[i+1]
			}
			return
		}
	}
	return nil, nil
}

// buildAllSortedDrawPets 构建 DrawPet 全表按 StartTime 排序的切片
func buildAllSortedDrawPets(drawPetDiff *[]draw_pet.DrawPetDiff) []*draw_pet.DrawPetDiff {
	var sorted []*draw_pet.DrawPetDiff
	if drawPetDiff == nil {
		return sorted
	}
	for i := range *drawPetDiff {
		sorted = append(sorted, &(*drawPetDiff)[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime < sorted[j].StartTime
	})
	return sorted
}

// findCurrentDrawPet 从关联的 DrawPet 列表中找到当前期
// 优先选择正在进行中的（StartTime <= now <= EndTime），如果没有则选最新一期
func findCurrentDrawPet(pets []*draw_pet.DrawPetDiff) *draw_pet.DrawPetDiff {
	if len(pets) == 0 {
		return nil
	}
	now := time.Now()
	for _, dp := range pets {
		start, err1 := time.Parse("2006-01-02 15:04:05", dp.StartTime)
		end, err2 := time.Parse("2006-01-02 15:04:05", dp.EndTime)
		if err1 == nil && err2 == nil && (now.Equal(start) || now.After(start)) && (now.Equal(end) || now.Before(end)) {
			return dp
		}
	}
	var latest *draw_pet.DrawPetDiff
	for _, dp := range pets {
		if latest == nil || dp.StartTime > latest.StartTime {
			latest = dp
		}
	}
	return latest
}

// findPrevNextDrawPet 在按 StartTime 排序的切片中找到当前 DrawPet 的前一个和后一个
func findPrevNextDrawPet(sorted []*draw_pet.DrawPetDiff, currentId int) (prev *draw_pet.DrawPetDiff, next *draw_pet.DrawPetDiff) {
	for i, dp := range sorted {
		if dp.Id == currentId {
			if i > 0 {
				prev = sorted[i-1]
			}
			if i < len(sorted)-1 {
				next = sorted[i+1]
			}
			return
		}
	}
	return nil, nil
}

// buildAllSortedSeasonPasses 构建 SeasonPass 全表按 StartTime 排序的切片
func buildAllSortedSeasonPasses(seasonPassDiff *[]season_pass.SeasonPassDiff) []*season_pass.SeasonPassDiff {
	var sorted []*season_pass.SeasonPassDiff
	if seasonPassDiff == nil {
		return sorted
	}
	for i := range *seasonPassDiff {
		sorted = append(sorted, &(*seasonPassDiff)[i])
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime < sorted[j].StartTime
	})
	return sorted
}

// findPrevNextSeasonPass 在按 StartTime 排序的切片中找到当前 SeasonPass 的前一个和后一个
func findPrevNextSeasonPass(sorted []*season_pass.SeasonPassDiff, currentId int) (prev *season_pass.SeasonPassDiff, next *season_pass.SeasonPassDiff) {
	for i, sp := range sorted {
		if sp.Id == currentId {
			if i > 0 {
				prev = sorted[i-1]
			}
			if i < len(sorted)-1 {
				next = sorted[i+1]
			}
			return
		}
	}
	return nil, nil
}

// findCurrentSeasonPass 从排序后的 SeasonPass 列表中找到当前期
// 优先选择正在进行中的（StartTime <= now <= EndTime），如果没有则选最新一期
func findCurrentSeasonPass(sorted []*season_pass.SeasonPassDiff) *season_pass.SeasonPassDiff {
	if len(sorted) == 0 {
		return nil
	}
	now := time.Now()
	for _, sp := range sorted {
		start, err1 := time.Parse("2006-01-02 15:04:05", sp.StartTime)
		end, err2 := time.Parse("2006-01-02 15:04:05", sp.EndTime)
		if err1 == nil && err2 == nil && (now.Equal(start) || now.After(start)) && (now.Equal(end) || now.Before(end)) {
			return sp
		}
	}
	// 没有进行中的，返回最新一期（StartTime 最大的）
	return sorted[len(sorted)-1]
}

// buildDrawSkinIndex 构建 DrawSkin 索引（按 Id）
func buildDrawSkinIndex(drawSkinDiff *[]draw_skin.DrawSkinDiff) map[int]*draw_skin.DrawSkinDiff {
	index := make(map[int]*draw_skin.DrawSkinDiff)
	if drawSkinDiff == nil {
		return index
	}
	for i := range *drawSkinDiff {
		ds := &(*drawSkinDiff)[i]
		index[ds.Id] = ds
	}
	return index
}

func buildDropRuleIndex(dropRuleDiff *[]drop_rule.DropRuleDiff) map[int]*drop_rule.DropRuleDiff {
	index := make(map[int]*drop_rule.DropRuleDiff)
	if dropRuleDiff == nil {
		return index
	}
	for i := range *dropRuleDiff {
		dr := &(*dropRuleDiff)[i]
		index[dr.Id] = dr
	}
	return index
}

func buildDropGroupIndex(dropGroupDiff *[]drop_group.DropGroupDiff) map[int]*drop_group.DropGroupDiff {
	index := make(map[int]*drop_group.DropGroupDiff)
	if dropGroupDiff == nil {
		return index
	}
	for i := range *dropGroupDiff {
		dg := &(*dropGroupDiff)[i]
		index[dg.Id] = dg
	}
	return index
}

func buildDropItemsIndex(dropItemDiff *[]drop_item.DropItemDiff) map[int][]*drop_item.DropItemDiff {
	index := make(map[int][]*drop_item.DropItemDiff)
	if dropItemDiff == nil {
		return index
	}
	for i := range *dropItemDiff {
		di := &(*dropItemDiff)[i]
		index[di.DropGroup] = append(index[di.DropGroup], di)
	}
	return index
}

func buildTimesRewardsByStrIndex(timesRewardDiff *[]limit_skin_times_reward.LimitSkinTimesRewardDiff) map[string][]*limit_skin_times_reward.LimitSkinTimesRewardDiff {
	index := make(map[string][]*limit_skin_times_reward.LimitSkinTimesRewardDiff)
	if timesRewardDiff == nil {
		return index
	}
	for i := range *timesRewardDiff {
		tr := &(*timesRewardDiff)[i]
		key := tr.ActIdStr
		if key == "" {
			key = "invalid"
		}
		index[key] = append(index[key], tr)
	}
	return index
}

func buildShopIndex(shopDiff *[]shop.ShopDiff) map[string]*shop.ShopDiff {
	index := make(map[string]*shop.ShopDiff)
	if shopDiff == nil {
		return index
	}
	for i := range *shopDiff {
		s := &(*shopDiff)[i]
		index[s.ShopType] = s
	}
	return index
}

func buildShopGoodsIndex(shopGoodsDiff *[]shop_goods.ShopGoodsDiff) map[string][]*shop_goods.ShopGoodsDiff {
	index := make(map[string][]*shop_goods.ShopGoodsDiff)
	if shopGoodsDiff == nil {
		return index
	}
	for i := range *shopGoodsDiff {
		g := &(*shopGoodsDiff)[i]
		index[g.ShopType] = append(index[g.ShopType], g)
	}
	return index
}

func buildHeroSkinCollitionIndex(collitionDiff *[]hero_skin_collition.HeroSkinCollectionDiff) map[string]*hero_skin_collition.HeroSkinCollectionDiff {
	index := make(map[string]*hero_skin_collition.HeroSkinCollectionDiff)
	if collitionDiff == nil {
		return index
	}
	for i := range *collitionDiff {
		c := &(*collitionDiff)[i]
		index[c.Type] = c
	}
	return index
}

func buildItemHeroSkinIndex(itemHeroSkinDiff *[]item_hero_skin.HeroSkinDiff) map[int]*item_hero_skin.HeroSkinDiff {
	index := make(map[int]*item_hero_skin.HeroSkinDiff)
	if itemHeroSkinDiff == nil {
		return index
	}
	for i := range *itemHeroSkinDiff {
		s := &(*itemHeroSkinDiff)[i]
		index[s.SkinItemId] = s
	}
	return index
}

func buildHeroSkinItemIndex(heroSkinItemDiff *[]hero_skin_item.HeroSkinItemDiff) map[int]*hero_skin_item.HeroSkinItemDiff {
	index := make(map[int]*hero_skin_item.HeroSkinItemDiff)
	if heroSkinItemDiff == nil {
		return index
	}
	for i := range *heroSkinItemDiff {
		s := &(*heroSkinItemDiff)[i]
		index[s.SkinItemId] = s
	}
	return index
}

func buildHeroSkinSpineIndex(heroSkinSpineDiff *[]hero_skin_spine.HeroSkinSpineDiff) map[int]*hero_skin_spine.HeroSkinSpineDiff {
	index := make(map[int]*hero_skin_spine.HeroSkinSpineDiff)
	if heroSkinSpineDiff == nil {
		return index
	}
	for i := range *heroSkinSpineDiff {
		s := &(*heroSkinSpineDiff)[i]
		index[s.SkinItemId] = s
	}
	return index
}

func buildDrawPetIndex(drawPetDiff *[]draw_pet.DrawPetDiff) map[int]*draw_pet.DrawPetDiff {
	index := make(map[int]*draw_pet.DrawPetDiff)
	if drawPetDiff == nil {
		return index
	}
	for i := range *drawPetDiff {
		dp := &(*drawPetDiff)[i]
		index[dp.Id] = dp
	}
	return index
}

func buildPetIndex(petDiff *[]pet.PetDiff) map[int]*pet.PetDiff {
	index := make(map[int]*pet.PetDiff)
	if petDiff == nil {
		return index
	}
	for i := range *petDiff {
		p := &(*petDiff)[i]
		index[p.Id] = p
	}
	return index
}

func buildPetAudioIndex(petAudioDiff *[]pet_audio.PetAudioDiff) map[int][]*pet_audio.PetAudioDiff {
	index := make(map[int][]*pet_audio.PetAudioDiff)
	if petAudioDiff == nil {
		return index
	}
	for i := range *petAudioDiff {
		pa := &(*petAudioDiff)[i]
		index[pa.ItemCfgId] = append(index[pa.ItemCfgId], pa)
	}
	return index
}

func buildAccumulatedRechargeByActIdIndex(diff *[]accumulated_recharge.AccumulatedRechargeDiff) map[string][]*accumulated_recharge.AccumulatedRechargeDiff {
	index := make(map[string][]*accumulated_recharge.AccumulatedRechargeDiff)
	if diff == nil {
		return index
	}
	for i := range *diff {
		r := &(*diff)[i]
		key := r.ActId
		if key == "" {
			continue
		}
		index[key] = append(index[key], r)
	}
	return index
}

// buildSeasonPassBagIndex 构建 SeasonPassBag 索引（按 SeasonPassId）
func buildSeasonPassBagIndex(diff *[]season_pass_bag.SeasonPassBagDiff) map[int][]*season_pass_bag.SeasonPassBagDiff {
	index := make(map[int][]*season_pass_bag.SeasonPassBagDiff)
	if diff == nil {
		return index
	}
	for i := range *diff {
		b := &(*diff)[i]
		index[b.SeasonPassId] = append(index[b.SeasonPassId], b)
	}
	return index
}

// buildSeasonPassRewardIndex 构建 SeasonPassReward 索引（按 SeasonPassId）
func buildSeasonPassRewardIndex(diff *[]season_pass_reward.SeasonPassRewardDiff) map[int][]*season_pass_reward.SeasonPassRewardDiff {
	index := make(map[int][]*season_pass_reward.SeasonPassRewardDiff)
	if diff == nil {
		return index
	}
	for i := range *diff {
		r := &(*diff)[i]
		index[r.SeasonPassId] = append(index[r.SeasonPassId], r)
	}
	return index
}

// buildSeasonPassTaskIndex 构建 SeasonPassTask 索引（按 SeasonPassId）
func buildSeasonPassTaskIndex(diff *[]season_pass_task.SeasonPassTaskDiff) map[int][]*season_pass_task.SeasonPassTaskDiff {
	index := make(map[int][]*season_pass_task.SeasonPassTaskDiff)
	if diff == nil {
		return index
	}
	for i := range *diff {
		t := &(*diff)[i]
		index[t.SeasonPassId] = append(index[t.SeasonPassId], t)
	}
	return index
}
