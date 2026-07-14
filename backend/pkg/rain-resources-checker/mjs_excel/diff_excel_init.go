package mjs_excel

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/accumulated_recharge"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve_hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/activity"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/arena_score_rewards"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/buff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/country"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/draw_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_collition"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_spine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/item_hero_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/limit_skin_times_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/pet_audio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/recommend_bd"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/robot_action"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_bag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_task"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/shop_goods"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_melt"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_tag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/task_complete_cond"
	"github.com/xuri/excelize/v2"
)

func InitDiffRefExcel(sheetMap map[string]*excelize.File) (*diff.DataContainer, error) {
	heroDiff, err := hero.GetHeroDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	countryDiff, err := country.GetCountryDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	heroLinesDiff, err := hero_lines.GetHeroLinesDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	heroUIDiff, err := hero_ui.GetHeroUIDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	heroSkinCollectionDiff, err := hero_skin_collition.GetHeroSkinCollectionDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	heroSkinItemDiff, err := hero_skin_item.GetHeroSkinItemDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	heroSkinSpineDiff, err := hero_skin_spine.GetHeroSkinSpineDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	itemHeroSkinDiff, err := item_hero_skin.GetHeroSkinDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	dropGroupDiff, err := drop_group.GetDropGroupDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	dropItemDiff, err := drop_item.GetDropItemDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	dropRuleDiff, err := drop_rule.GetDropRuleDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	seasonPassDiff, err := season_pass.GetSeasonPassDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	seasonPassBagDiff, err := season_pass_bag.GetSeasonPassBagDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	seasonPassRewardDiff, err := season_pass_reward.GetSeasonPassRewardDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	seasonPassTaskDiff, err := season_pass_task.GetSeasonPassTaskDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	arenaScoreRewardsDiff, err := arena_score_rewards.GetArenaScoreRewardDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	skillDiff, err := skill.GetSkillDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	skillLineDiff, err := skill_lines.GetSkillLinesDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	skillMeltDiff, err := skill_melt.GetSkillMeltDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	skillTagDiff, err := skill_tag.GetSkillTagDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	skillUIDiff, err := skill_ui.GetSkillUIDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	buffDiff, err := buff.GetBuffDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	recommendBDDiff, err := recommend_bd.GetRecommendBdDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	robotActionDiff, err := robot_action.GetRobotActionDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	taskCompleteCondDiff, err := task_complete_cond.GetTaskCompleteConditonDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	achieveHeroDiff, err := achieve_hero.GetHeroAchieveDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	achieveDiff, err := achieve.GetAchieveDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	activityDiff, err := activity.GetActivityDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	drawSkinDiff, err := draw_skin.GetDrawSkinDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	limitSkinTimesRewardDiff, err := limit_skin_times_reward.GetLimitSkinTimesRewardDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	shopDiff, err := shop.GetShopDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	shopGoodsDiff, err := shop_goods.GetShopGoodsDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	drawPetDiff, err := draw_pet.GetDrawPetDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	petDiff, err := pet.GetPetDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	petAudioDiff, err := pet_audio.GetPetAudioDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}
	accumulatedRechargeDiff, err := accumulated_recharge.GetAccumulatedRechargeDiffMap(sheetMap)
	if err != nil {
		return nil, err
	}

	excel := &diff.DataContainer{
		HeroWikiDiff:             nil,
		HeroDiff:                 heroDiff,
		CountryDiff:              countryDiff,
		HeroLinesDiff:            heroLinesDiff,
		HeroUIDiff:               heroUIDiff,
		HeroSkinCollectionDiff:   heroSkinCollectionDiff,
		HeroSkinItemDiff:         heroSkinItemDiff,
		HeroSkinSpineDiff:        heroSkinSpineDiff,
		ItemHeroSkinDiff:         itemHeroSkinDiff,
		DropGroupDiff:            dropGroupDiff,
		DropItemDiff:             dropItemDiff,
		DropRuleDiff:             dropRuleDiff,
		SeasonPassDiff:           seasonPassDiff,
		SeasonPassBagDiff:        seasonPassBagDiff,
		SeasonPassRewardDiff:     seasonPassRewardDiff,
		SeasonPassTaskDiff:       seasonPassTaskDiff,
		ArenaScoreRewardsDiff:    arenaScoreRewardsDiff,
		SkillDiff:                skillDiff,
		SkillLineDiff:            skillLineDiff,
		SkillMeltDiff:            skillMeltDiff,
		SkillTagDiff:             skillTagDiff,
		SkillUIDiff:              skillUIDiff,
		BuffDiff:                 buffDiff,
		RecommendBdDiff:          recommendBDDiff,
		RobotActionDiff:          robotActionDiff,
		TaskCompleteCondDiff:     taskCompleteCondDiff,
		AchieveHeroDiff:          achieveHeroDiff,
		AchieveDiff:              achieveDiff,
		ActivityWikiDiff:         nil, // 稍后构建
		ActivityDiff:             activityDiff,
		DrawSkinDiff:             drawSkinDiff,
		LimitSkinTimesRewardDiff: limitSkinTimesRewardDiff,
		ShopDiff:                 shopDiff,
		ShopGoodsDiff:            shopGoodsDiff,
		DrawPetDiff:              drawPetDiff,
		PetDiff:                  petDiff,
		PetAudioDiff:             petAudioDiff,
		AccumulatedRechargeDiff:  accumulatedRechargeDiff,
	}

	return excel, nil
}
