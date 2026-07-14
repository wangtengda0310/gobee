package activitywiki_def

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/accumulated_recharge"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/activity"
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

// ActivityWikiDiff 活动Wiki聚合数据
type ActivityWikiDiff struct {
	Activities   map[int]*ActivityCompleteData   // 按Activity.Id索引
	SeasonPasses map[int]*SeasonPassCompleteData // 按SeasonPass.Id索引（战令数据）
}

func (a ActivityWikiDiff) GetType() string {
	return "ActivityWikiDiff"
}

// ActivityCompleteData 单个活动的完整数据
type ActivityCompleteData struct {
	Basic *activity.ActivityDiff

	// 丹青阁特有数据
	DrawSkin      *draw_skin.DrawSkinDiff
	PrevDrawSkin  *draw_skin.DrawSkinDiff     // 丹青阁上一期抽奖配置（按StartTime排序的前一个），可能为nil
	NextDrawSkin  *draw_skin.DrawSkinDiff     // 丹青阁下一期抽奖配置（按StartTime排序的后一个），可能为nil
	DropRule      *drop_rule.DropRuleDiff     // 单抽掉落规则（OnceDropRule关联）
	DropGroups    []*drop_group.DropGroupDiff // 单抽掉落组
	DropItems     []*drop_item.DropItemDiff   // 单抽掉落项
	TenDropRule   *drop_rule.DropRuleDiff     // 十连掉落规则（TenDropRule关联）
	TenDropGroups []*drop_group.DropGroupDiff // 十连掉落组
	TenDropItems  []*drop_item.DropItemDiff   // 十连掉落项
	TimesRewards  []*limit_skin_times_reward.LimitSkinTimesRewardDiff

	// 商店相关数据
	Shop      *shop.ShopDiff
	ShopGoods []*shop_goods.ShopGoodsDiff

	// 皮肤相关数据
	HeroSkinCollition *hero_skin_collition.HeroSkinCollectionDiff
	ItemHeroSkin      *item_hero_skin.HeroSkinDiff
	HeroSkinItem      *hero_skin_item.HeroSkinItemDiff
	HeroSkinSpine     *hero_skin_spine.HeroSkinSpineDiff

	// 结缘庭相关数据
	DrawPet     *draw_pet.DrawPetDiff // 当前期结缘亭抽奖配置（根据时间判断）
	PrevDrawPet *draw_pet.DrawPetDiff // 上一期结缘亭抽奖配置（按StartTime排序的前一个），可能为nil
	NextDrawPet *draw_pet.DrawPetDiff // 下一期结缘亭抽奖配置（按StartTime排序的后一个），可能为nil
	Pets        []*pet.PetDiff
	PetAudios   []*pet_audio.PetAudioDiff

	// activity-wiki-dev: 新增字段 - 累充活动奖励列表
	AccumulatedRecharges []*accumulated_recharge.AccumulatedRechargeDiff
}

func (a ActivityCompleteData) GetType() string {
	return "ActivityCompleteData"
}

// SeasonPassCompleteData 单个战令的完整数据
type SeasonPassCompleteData struct {
	Basic   *season_pass.SeasonPassDiff
	Bags    []*season_pass_bag.SeasonPassBagDiff
	Rewards []*season_pass_reward.SeasonPassRewardDiff
	Tasks   []*season_pass_task.SeasonPassTaskDiff

	// 战令上一期/下一期（按StartTime排序）
	PrevSeasonPass *season_pass.SeasonPassDiff // 上一期战令，可能为nil
	NextSeasonPass *season_pass.SeasonPassDiff // 下一期战令，可能为nil

	// 本期战令关联数据（已按SeasonPassId过滤）
	CurrentBags    []*season_pass_bag.SeasonPassBagDiff
	CurrentRewards []*season_pass_reward.SeasonPassRewardDiff
	CurrentTasks   []*season_pass_task.SeasonPassTaskDiff
}

func (s SeasonPassCompleteData) GetType() string {
	return "SeasonPassCompleteData"
}
