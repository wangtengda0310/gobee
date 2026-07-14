package herowiki

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/achieve_hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/arena_score_rewards"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/buff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/country"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/drop_group"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_collition"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_spine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/item_hero_skin"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/recommend_bd"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/robot_action"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/season_pass_reward"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_melt"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_tag"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_ui"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/task_complete_cond"
)

// BuildHeroWikiDiff 构建完整的HeroWikiDiff对象
func BuildHeroWikiDiff(container *diff.DataContainer) *herowiki_def.HeroWikiDiff {
	wiki := &herowiki_def.HeroWikiDiff{
		Indexes: &herowiki_def.HeroIndexes{
			HeroByID:              make(map[int]string),
			HeroByEHeroId:         make(map[string]int),
			SkinsByHeroId:         make(map[string][]int),
			LinesBySkinId:         make(map[int][]int),
			ESkillBySkillId:       make(map[int]string),
			SkillsByHeroId:        make(map[string][]string),
			LinesBySkillId:        make(map[string][]int),
			BuffIdByEBuffId:       make(map[string]string),
			HeroesByCountry:       make(map[string][]string),
			AchieveHeroesByHeroId: make(map[string][]string),
			AchieveByHeroId:       make(map[string][]int),
		},
		Heroes: make(map[string]*herowiki_def.HeroCompleteData),
	}

	// 1. 构建基础索引
	buildBasicIndexes(wiki, container)

	// 2.1. 构建英雄ID映射（数字ID -> EHeroId）
	heroIdMap := make(map[string]string)
	for _, heroDiff := range *container.HeroDiff {
		heroIdMap[strconv.Itoa(heroDiff.Id)] = heroDiff.EHeroId
	}

	// 2.2 创建掉落信息构建器
	dropBuilder := NewDropInfoBuilder(container, heroIdMap)

	// 3. 处理每个武将的完整数据
	for _, heroDiff := range *container.HeroDiff {
		heroData := buildHeroCompleteData(heroDiff, container, wiki.Indexes)

		// 添加掉落信息
		heroData.DropInfo = dropBuilder.BuildHeroDropInfo(
			heroDiff.EHeroId,
			heroDiff.Name,
		)

		wiki.Heroes[heroDiff.EHeroId] = heroData
	}

	// ?. 额外处理国家信息（关联武将）
	//for _, countryDiff := range container.CountryDiff {
	//	if heroes, ok := wiki.Indexes.HeroesByCountry[countryDiff.ECountry]; ok {
	//		// 国家信息已经在buildHeroCompleteData中关联
	//		// 这里可以补充国家维度的信息
	//	}
	//}

	return wiki
}

// buildBasicIndexes 构建基础索引
func buildBasicIndexes(wiki *herowiki_def.HeroWikiDiff, container *diff.DataContainer) {
	// HeroDiff索引
	for _, h := range *container.HeroDiff {
		wiki.Indexes.HeroByID[h.Id] = h.EHeroId
		wiki.Indexes.HeroByEHeroId[h.EHeroId] = h.Id

		// 国家索引
		wiki.Indexes.HeroesByCountry[h.Country] = append(
			wiki.Indexes.HeroesByCountry[h.Country],
			h.EHeroId,
		)
	}

	// 皮肤索引
	for _, skin := range *container.HeroSkinItemDiff {
		wiki.Indexes.SkinsByHeroId[skin.HeroId] = append(
			wiki.Indexes.SkinsByHeroId[skin.HeroId],
			skin.SkinItemId,
		)

		// 皮肤台词索引
		wiki.Indexes.LinesBySkinId[skin.SkinItemId] = skin.Lines
	}

	// 技能ID Enum索引
	for _, h := range *container.SkillDiff {
		if atoi, err := strconv.Atoi(h.Id); err == nil {
			wiki.Indexes.ESkillBySkillId[atoi] = h.ESkillId
		}
	}

	// 技能索引
	for _, h := range *container.HeroDiff {
		for _, skillId := range h.Skill {
			skillIdStr := wiki.Indexes.ESkillBySkillId[skillId] // 转换
			wiki.Indexes.SkillsByHeroId[h.EHeroId] = append(
				wiki.Indexes.SkillsByHeroId[h.EHeroId],
				skillIdStr,
			)
		}
	}

	// Buff索引
	for _, buff_ := range *container.BuffDiff {
		wiki.Indexes.BuffIdByEBuffId[buff_.EBuffId] = buff_.Id
	}

	// 技能台词索引
	for _, skillLine := range *container.SkillLineDiff {
		wiki.Indexes.LinesBySkillId[skillLine.SkillId] = append(
			wiki.Indexes.LinesBySkillId[skillLine.SkillId],
			skillLine.Id,
		)
	}

	// 英雄成就索引
	for _, achieve_ := range *container.AchieveHeroDiff {
		for _, EHeroId := range achieve_.UseHero {
			wiki.Indexes.AchieveHeroesByHeroId[EHeroId] = append(
				wiki.Indexes.AchieveHeroesByHeroId[EHeroId],
				achieve_.Id, // 需要确认成就ID字段
			)
		}
	}

	// 成就索引
	for _, achieve_ := range *container.AchieveDiff {
		for _, heroItemId := range achieve_.HeroItemId {
			heroItemIdStr := strconv.FormatInt(int64(heroItemId), 10)
			if len(heroItemIdStr) < 7 {
				continue
			}
			heroIdStr := heroItemIdStr[2:]
			wiki.Indexes.AchieveByHeroId[heroIdStr] = append(
				wiki.Indexes.AchieveByHeroId[heroIdStr],
				achieve_.Id, // 需要确认成就ID字段
			)
		}
	}
}

// buildHeroCompleteData 构建单个武将完整数据
func buildHeroCompleteData(
	heroDiff hero.HeroDiff,
	container *diff.DataContainer,
	indexes *herowiki_def.HeroIndexes,
) *herowiki_def.HeroCompleteData {
	heroData := &herowiki_def.HeroCompleteData{
		Basic:        buildHeroBasicInfo(heroDiff),
		UI:           findHeroUI(heroDiff.Id, container.HeroUIDiff),
		Skins:        buildHeroSkins(strconv.FormatInt(int64(heroDiff.Id), 10), container, indexes),
		Skills:       buildHeroSkills(heroDiff.Skill, container, indexes),
		Country:      findCountry(heroDiff.Country, container.CountryDiff),
		RecommendBd:  findRecommendBd(strconv.FormatInt(int64(heroDiff.Id), 10), container.RecommendBdDiff),
		Achievements: buildHeroAchievements(heroDiff.EHeroId, container, indexes),
		RobotActions: findRobotActions(heroDiff.Skill, container.RobotActionDiff),
	}

	return heroData
}

// buildHeroSkins 构建武将皮肤
func buildHeroSkins(
	heroId string,
	container *diff.DataContainer,
	indexes *herowiki_def.HeroIndexes,
) []*herowiki_def.HeroSkinInfo {
	var skins []*herowiki_def.HeroSkinInfo

	skinIds := indexes.SkinsByHeroId[heroId]
	for _, skinId := range skinIds {
		// 查找皮肤基础信息
		var skinItem *hero_skin_item.HeroSkinItemDiff
		for _, s := range *container.HeroSkinItemDiff {
			if s.SkinItemId == skinId {
				skinItem = &s
				break
			}
		}
		if skinItem == nil {
			continue
		}

		skin := &herowiki_def.HeroSkinInfo{
			ItemId:              skinItem.SkinItemId,
			HeroId:              skinItem.HeroId,
			Name:                skinItem.Name,
			GetWay:              skinItem.GetWay,
			SkinPinYin:          skinItem.SkinPinYin,
			SkinType:            skinItem.SkinType,
			RailType:            skinItem.RailyType,
			SeatSpecialImg:      skinItem.SeatSpecialImg,
			HeroUIExtraIcons:    skinItem.HeroUIExtraIcons,
			OriginalArtDesigner: skinItem.OriginalArtDesigner,
			CollectionType:      skinItem.CollitionType,
			IsOpen:              skinItem.IsOpen,
			OpenDate:            skinItem.OpenDate,
			CollectionTagImg:    skinItem.CollitionTagImg,
			KillShowTime:        skinItem.KillShowTime,
			BodyOffset:          skinItem.BodyOffset,
			HasOutFrameIcon:     skinItem.HasOutFrameIcon,
		}

		// 关联Spine信息
		skin.Spine = findSkinSpine(skinId, container.HeroSkinSpineDiff)

		// 关联资源信息
		skin.Resource = findSkinResource(skinId, container.ItemHeroSkinDiff)

		// 关联台词
		skin.LinesDubbed = skinItem.LinesDubbed
		skin.Lines = findSkinLines(skinItem.Lines, container.HeroLinesDiff)

		// 关联收藏册
		skin.Collection = findSkinCollection(skinItem.CollitionType, container.HeroSkinCollectionDiff)

		skins = append(skins, skin)
	}

	return skins
}

// buildHeroSkills 构建武将技能
func buildHeroSkills(
	skillIds []int,
	container *diff.DataContainer,
	indexes *herowiki_def.HeroIndexes,
) []*herowiki_def.HeroSkillInfo {
	var skills []*herowiki_def.HeroSkillInfo

	for _, skillId := range skillIds {
		skillIdStr := indexes.ESkillBySkillId[skillId]

		// 查找技能基础信息
		var skillBasic *skill.SkillDiff
		for _, s := range *container.SkillDiff {
			if s.ESkillId == skillIdStr {
				skillBasic = &s
				break
			}
		}
		if skillBasic == nil {
			continue
		}

		ui := findSkillUI(skillIdStr, *container.SkillUIDiff)
		tags := make([]int, 0)
		if ui != nil {
			tags = ui.SkillTag
		}
		skill_ := &herowiki_def.HeroSkillInfo{
			Basic: buildSkillBasicInfo(*skillBasic, container.BuffDiff, indexes),
			UI:    ui,
			Melt:  findSkillMelt(strconv.FormatInt(int64(skillId), 10), container.SkillMeltDiff),
			Lines: findSkillLines(skillIdStr, container.SkillLineDiff),
			Tags:  findSkillTags(tags, container.SkillTagDiff),
		}

		skills = append(skills, skill_)
	}

	return skills
}

// 辅助查找函数

func findHeroUI(id int, uiList *[]hero_ui.HeroUIDiff) *herowiki_def.HeroUIInfo {
	for _, ui := range *uiList {
		if ui.Id == id {
			return &herowiki_def.HeroUIInfo{
				HeroDiffId:          ui.Id,
				Name:                ui.Name,
				VoiceType:           ui.VoiceType,
				BelongExpansionPack: ui.BelongExpansionPack,
				AwardDes:            ui.AwardDes,
				LongIntroduction:    ui.LongIntroduction,
				ShortIntroduction:   ui.ShortIntroduction,
				Evidence:            ui.Evidence,
				Evaluation:          ui.Evaluation,
				IsNew:               ui.IsNew,
				CopyWriter:          ui.CopyWriter,
				SkillDesigner:       ui.SkillDesigner,
				GetWay:              ui.GetWay,
				Position:            ui.Position,
				ExclusiveCard:       ui.ExclusiveCard,
				NewbieShowSkillTag:  ui.NewbieShowSkillTag,
				WinningRateIn2v2:    ui.WinningRateIn2v2,
				WinRateShowPriority: ui.WinRateShowPriority,
			}
		}
	}
	return nil
}

func findCountry(eCountry string, countryList *[]country.CountryDiff) *herowiki_def.CountryInfo {
	for _, c := range *countryList {
		if c.ECountry == eCountry {
			return &herowiki_def.CountryInfo{
				ECountry:                  c.ECountry,
				Name:                      c.Name,
				KingName:                  c.KingName,
				SeatFlagImagePath:         c.SeatFlagImagePath,
				PlayerInfoFlagImagePath:   c.PlayerInfoFlagImagePath,
				PlayerInfoFlagTextPath:    c.PlayerInfoFlagTextPath,
				IsWhiteBg:                 c.IsWhiteBg,
				IsOpen:                    c.IsOpen,
				NameArtSelectedFontPath:   c.NameArtSelectedFontPath,
				NameArtUnselectedFontPath: c.NameArtUnselectedFontPath,
				ShiLiSmallTopIcon:         c.ShiLiSmallTopIcon,
				ShiLiSmallBottomIcon:      c.ShiLiSmallBottomIcon,
				SortPriority:              c.SortPriority,
				GuildIcon:                 c.GuildIcon,
			}
		}
	}
	return nil
}

func findRecommendBd(heroId string, bdList *[]recommend_bd.RecommendBdDiff) *herowiki_def.RecommendBdInfo {
	for _, bd := range *bdList {
		if bd.HeroId == heroId {
			return &herowiki_def.RecommendBdInfo{
				HeroId: bd.HeroId,
				Level1: bd.Level1,
				Level2: bd.Level2,
				Level3: bd.Level3,
				Level4: bd.Level4,
				Level5: bd.Level5,
				Level6: bd.Level6,
			}
		}
	}
	return nil
}

// 其他辅助函数类似实现...

// buildHeroBasicInfo 构建武将基础信息
func buildHeroBasicInfo(h hero.HeroDiff) *herowiki_def.HeroBasicInfo {
	return &herowiki_def.HeroBasicInfo{
		Id:                  h.Id,
		EHeroId:             h.EHeroId,
		Name:                h.Name,
		IsOpen:              h.IsOpen,
		OpenDate:            h.OpenDate,
		Gender:              h.Gender,
		Point:               h.Point,
		HpLimit:             h.HpLimit,
		HandLimit:           h.HandLimit,
		EquipLimit:          h.EquipLimit,
		Country:             h.Country,
		IsAlwaysZhuGong:     h.IsAlwaysZhuGong,
		ExcludeIdentity:     h.ExcludeIdentity,
		NotUseModeType:      h.NotUseModeType,
		HeroType:            h.HeroType,
		EHeroType:           h.EHeroType,
		CanMelt:             h.CanMelt,
		MeltName:            h.MeltName,
		IsNewHero:           h.IsNewHero,
		IsGacha:             h.IsGacha,
		BelongExpansionPack: h.BelongExpansionPack,
	}
}

// findSkinSpine 查找皮肤Spine信息
func findSkinSpine(skinItemId int, spineList *[]hero_skin_spine.HeroSkinSpineDiff) *herowiki_def.HeroSkinSpineInfo {
	for _, spine := range *spineList {
		if spine.SkinItemId == skinItemId {
			return &herowiki_def.HeroSkinSpineInfo{
				SkinItemId:                          spine.SkinItemId,
				IsHasSeatSpine:                      spine.IsHasSeatSpine,
				IsHasBookSpine:                      spine.IsHasBookSpine,
				IsHasMainBgSpine:                    spine.IsHasMainBgSpine,
				MainBgFx:                            spine.MainBgFx,
				IsHasSeatKillSpine:                  spine.IsHasSeatKillSpine,
				KillFxId:                            spine.KillFxId,
				IsHasCollectionBgSpine:              spine.IsHasCollitionBgSpine,
				IsHasCollectionCardBgSpine:          spine.IsHasCollitionCardBgSpine,
				SpineAnimAudio:                      spine.SpineAnimAudio,
				IsHasCollectionCardBgSpineDuplicate: spine.IsHasCollitionCardBgSpineDuplicate,
				KillAudio:                           spine.KillAudio,
			}
		}
	}
	return nil
}

// findSkinResource 查找皮肤资源信息
func findSkinResource(skinItemId int, resourceList *[]item_hero_skin.HeroSkinDiff) *herowiki_def.HeroSkinResourceInfo {
	for _, resource := range *resourceList {
		if resource.SkinItemId == skinItemId {
			return &herowiki_def.HeroSkinResourceInfo{
				SkinItemId: resource.SkinItemId,
				Path:       resource.Path,
			}
		}
	}
	return nil
}

// findSkinLines 查找皮肤台词
func findSkinLines(lineIds []int, linesList *[]hero_lines.HeroLinesDiff) []*herowiki_def.HeroLineInfo {
	var lines []*herowiki_def.HeroLineInfo

	for _, lineId := range lineIds {
		for _, line := range *linesList {
			if line.Id == lineId {
				lines = append(lines, &herowiki_def.HeroLineInfo{
					Id:           line.Id,
					Type:         line.Type,
					TabName:      line.TabName,
					Text:         line.Text,
					AudioId:      line.AudioId,
					Achievements: line.Achievements,
					GroupId:      line.GroupId,
				})
				break
			}
		}
	}

	return lines
}

// findSkinCollection 查找皮肤收藏册信息
func findSkinCollection(collectionType string, collectionList *[]hero_skin_collition.HeroSkinCollectionDiff) *herowiki_def.HeroSkinCollectionInfo {
	for _, collection := range *collectionList {
		if collection.Type == collectionType {
			return &herowiki_def.HeroSkinCollectionInfo{
				Type:     collection.Type,
				Name:     collection.Name,
				NameImg:  collection.NameImg,
				NameBg:   collection.NameBg,
				Desc:     collection.Desc,
				Weight:   collection.Weight,
				OpenDate: collection.OpenDate,
			}
		}
	}
	return nil
}

// buildSkillBasicInfo 构建技能基础信息
func buildSkillBasicInfo(s skill.SkillDiff, buffList *[]buff.BuffDiff, indexes *herowiki_def.HeroIndexes) *herowiki_def.SkillBasicInfo {
	return &herowiki_def.SkillBasicInfo{
		Id:                s.Id,
		SkillName:         s.SkillName,
		ShortSkillName:    s.ShortSkillName,
		ESkillId:          s.ESkillId,
		SkillType:         s.SkillType,
		IsFromOther:       s.IsFromOther,
		CounterFormula:    s.CounterFormula,
		IsFromAura:        s.IsFromAura,
		TransCard:         s.TransCard,
		SkillFromType:     s.SkillFromType,
		ResetCounterType:  s.ResetCounterType,
		ResetTimesType:    s.ResetTimesType,
		SkillLimitTimes:   s.SkillLimitTimes,
		TotalLimitTimes:   s.TotalLimitTimes,
		TriggerCondition:  s.TriggerCondition,
		SkillEffect:       s.SkillEffect,
		EmptyWaitTime:     s.EmptyWaitTime,
		Buff:              buildSkillBasicBuffInfo(s.Buff, buffList, indexes),
		ShowPointAndAttr:  s.ShowPointAndAttr,
		MutexSkill:        s.MutexSkill,
		BattleCardClass:   s.BattleCardClass,
		AIJudgeArea:       s.AIJudgeArea,
		DeadType:          s.DeadType,
		InitPro:           s.InitPro,
		MagicSkillID:      s.MagicSkillID,
		IsAutoSelAllOther: s.IsAutoSelAllOther,
		IsForbidCopy:      s.IsForbidCopy,
		IsForbidTrans:     s.IsForbidTrans,
		IsForbidDestroy:   s.IsForbidDestroy,
	}
}

// buildSkillBasicBuffInfo 构建技能基础信息
func buildSkillBasicBuffInfo(buffs []string, buffList *[]buff.BuffDiff, indexes *herowiki_def.HeroIndexes) []*herowiki_def.BuffInfo {
	buffInfos := make([]*herowiki_def.BuffInfo, 0, len(buffs))
	for _, eBuffId := range buffs {
		buffId := indexes.BuffIdByEBuffId[eBuffId]
		for _, buff_ := range *buffList {
			if buffId == buff_.Id {
				buffInfos = append(buffInfos, &herowiki_def.BuffInfo{
					Id:                     buff_.Id,
					EBuffId:                buff_.EBuffId,
					Name:                   buff_.Name,
					NeedRecord:             buff_.NeedRecord,
					BuffType:               buff_.BuffType,
					IsDeleteByCasterDead:   buff_.IsDeleteByCasterDead,
					IsDeleteByExecutorDead: buff_.IsDeleteByExecutorDead,
					IsValidByFengJin:       buff_.IsValidByFengJin,
					IsReserveByRemoveSkill: buff_.IsReserveByRemoveSkill,
					TransferSkillBuffType:  buff_.TransferSkillBuffType,
					Round:                  buff_.Round,
					Value:                  buff_.Value,
					EndByEffect:            buff_.EndByEffect,
					EndType:                buff_.EndType,
					IsServerOnly:           buff_.IsServerOnly,
					IsCasterOnly:           buff_.IsCasterOnly,
					OwnerType:              buff_.OwnerType,
					ShowArea:               buff_.ShowArea,
					Icon:                   buff_.Icon,
					Effect:                 buff_.Effect,
					FlashEffect:            buff_.FlashEffect,
					BuffPriority:           buff_.BuffPriority,
					OverlyingType:          buff_.OverlyingType,
					IsTrigger:              buff_.IsTrigger,
					BuffState:              buff_.BuffState,
					BuffPro:                buff_.BuffPro,
					ProValue:               buff_.ProValue,
					CostEffectValue:        buff_.CostEffectValue,
					TriggerTiming:          buff_.TriggerTiming,
					TriggerPriority:        buff_.TriggerPriority,
					TriggerCondition:       buff_.TriggerCondition,
					TriggerAction:          buff_.TriggerAction,
					BuffDot:                buff_.BuffDot,
					EffectDescribe:         buff_.EffectDescribe,
				})
			}
		}
	}
	return buffInfos
}

// findSkillUI 查找技能UI信息
func findSkillUI(ESkillId string, uiList []skill_ui.SkillUIDiff) *herowiki_def.SkillUIInfo {
	for _, ui := range uiList {
		if ui.Id == ESkillId {
			return &herowiki_def.SkillUIInfo{
				Id:              ui.Id,
				SkillName:       ui.SkillName,
				PlayCardAudio:   ui.PlayCardAudio,
				IdentityLine:    ui.IdentityLine,
				Audio:           ui.Audio,
				SkillText:       ui.SkillText,
				ShortSkillText:  ui.ShortSkillText,
				KeyWords:        ui.KeyWords,
				SettlementDes:   ui.SettlementDes,
				Allusion:        ui.Allusion,
				DesignThought:   ui.DesignThought,
				SkillTag:        ui.SkillTag,
				BattleSkillStep: ui.BattleSkillStep,
				HasRelation:     ui.HasRelation,
				RelatedSkill:    ui.RelatedSkill,
				SpecialAudio:    ui.SpecialAudio,
				ESkillId:        ui.ESkillId,
				EAudio:          ui.EAudio,
				EAudioId:        ui.EAudioId,
				ESkillTag:       ui.ESkillTag,
			}
		}
	}
	return nil
}

// findSkillMelt 查找技能熔炼信息
func findSkillMelt(skillId string, meltList *[]skill_melt.SkillMeltDiff) *herowiki_def.SkillMeltInfo {
	for _, melt := range *meltList {
		if melt.Id == skillId {
			return &herowiki_def.SkillMeltInfo{
				Id:        melt.Id,
				MeltPower: melt.MeltPower,
				CanMelt:   melt.CanMelt,
			}
		}
	}
	return nil
}

// findSkillLines 查找技能台词
func findSkillLines(skillId string, linesList *[]skill_lines.SkillLinesDiff) []*herowiki_def.SkillLineInfo {
	var lines []*herowiki_def.SkillLineInfo

	for _, line := range *linesList {
		if line.SkillId == skillId {
			lines = append(lines, &herowiki_def.SkillLineInfo{
				Id:              line.Id,
				SkillId:         line.SkillId,
				SkinId:          line.SkinId,
				SkillFirstLine:  line.SkillFirstLine,
				SkillSecondLine: line.SkillSecondLine,
				SkillThirdLine:  line.SkillThirdLine,
				SkillForthLine:  line.SkillForthLine,
				SpecialAudio:    line.SpecialAudio,
			})
		}
	}

	return lines
}

// findSkillTags 查找技能标签
func findSkillTags(tagIds []int, tagList *[]skill_tag.SkillTagDiff) []*herowiki_def.SkillTagInfo {
	var tags []*herowiki_def.SkillTagInfo

	// 注意：SkillTagDiff中的SkillTag是string类型，tagIds是int类型
	// 这里可能需要转换，暂时假设tagIds是某种索引
	for _, tagId := range tagIds {
		tagIdStr := strconv.Itoa(tagId)
		for _, tag := range *tagList {
			if tag.SkillTag == tagIdStr {
				tags = append(tags, &herowiki_def.SkillTagInfo{
					SkillTag: tag.SkillTag,
					TagName:  tag.TagName,
					TagColor: tag.TagColor,
					TagDes:   tag.TagDes,
				})
				break
			}
		}
	}

	return tags
}

// buildHeroAchievements 构建武将成就
func buildHeroAchievements(
	EHeroId string,
	container *diff.DataContainer,
	indexes *herowiki_def.HeroIndexes,
) []*herowiki_def.HeroAchievementInfo {
	var achievements []*herowiki_def.HeroAchievementInfo

	heroAchieveIds := indexes.AchieveHeroesByHeroId[EHeroId]
	for _, achieveId := range heroAchieveIds {
		// 查找成就类型
		var heroAchieve *achieve_hero.HeroAchieveDiff
		for _, ha := range *container.AchieveHeroDiff {
			// 需要确认成就ID字段名，假设为Id
			if ha.Id == achieveId {
				heroAchieve = &ha
				break
			}
		}
		if heroAchieve == nil {
			continue
		}

		achievements = append(achievements, &herowiki_def.HeroAchievementInfo{
			HeroAchieve: &herowiki_def.HeroAchieveInfo{
				Id:           heroAchieve.Id,
				Name:         heroAchieve.Name,
				IsMult:       heroAchieve.IsMult,
				Mode:         heroAchieve.Mode,
				MinPlayerNum: heroAchieve.MinPlayerNum,
				UseHero:      heroAchieve.UseHero,
				Class:        heroAchieve.Class,
				Camp:         heroAchieve.Camp,
				Identity:     heroAchieve.Identity,
				Hooker:       heroAchieve.Hooker,
				HookerTarget: heroAchieve.HookerTarget,
				CondParam:    heroAchieve.CondParam,
			},
		})
	}

	achieveIds := indexes.AchieveByHeroId[strconv.Itoa(indexes.HeroByEHeroId[EHeroId])]
	for _, achieveId := range achieveIds {
		// 查找成就类型
		var achieveDiff *achieve.AchieveDiff
		for _, ha := range *container.AchieveDiff {
			// 需要确认成就ID字段名，假设为Id
			if ha.Id == achieveId {
				achieveDiff = &ha
				break
			}
		}

		if achieveDiff == nil {
			continue
		}

		var condParamInfo herowiki_def.TaskCompleteConditionInfo

		for _, conditionDiff := range *container.TaskCompleteCondDiff {
			if conditionDiff.Id == achieveDiff.CompleteCondId {
				condParamInfo.Id = conditionDiff.Id
				condParamInfo.CondDes = conditionDiff.CondDes
				condParamInfo.CompleteCond = conditionDiff.CompleteCond
				condParamInfo.CompleteCondParam = conditionDiff.CompleteCondParam
				condParamInfo.JumpCond = conditionDiff.JumpCond
				condParamInfo.JumpParm = conditionDiff.JumpParm
			}
		}

		achievements = append(achievements, &herowiki_def.HeroAchievementInfo{
			AchieveDetail: &herowiki_def.AchieveDetailInfo{
				Id:             achieveDiff.Id,
				Name:           achieveDiff.Name,
				IsHide:         achieveDiff.IsHide,
				CompleteCondId: achieveDiff.CompleteCondId,
				Reward:         achieveDiff.Reward,
				None:           achieveDiff.None,
				Des:            achieveDiff.Des,
				Condition:      achieveDiff.Condition,
				ConditionInfo:  &condParamInfo,
				HeroItemId:     achieveDiff.HeroItemId,
				OpenDate:       achieveDiff.OpenDate,
			},
		})
	}

	return achievements
}

// findRobotActions 查找机器人行为
func findRobotActions(skillIds []int, robotList *[]robot_action.RobotActionDiff) []*herowiki_def.RobotActionInfo {
	var actions []*herowiki_def.RobotActionInfo

	// 注意：RobotActionDiff的Id是string类型，heroId也是string
	// 假设Id直接对应heroId
	for _, skillId := range skillIds {
		for _, robot := range *robotList {
			if robot.Id == strconv.FormatInt(int64(skillId), 10) {
				actions = append(actions, &herowiki_def.RobotActionInfo{
					Id:               robot.Id,
					Action:           robot.Action,
					TargetNum:        robot.TargetNum,
					TargetType:       robot.TargetType,
					CardNum:          robot.CardNum,
					CardFromType:     robot.CardFromType,
					TransCardSkill:   robot.TransCardSkill,
					DefaultCardSkill: robot.DefaultCardSkill,
				})
			}
		}
	}

	return actions
}

// 以下是对其他表的处理函数（虽然暂时无法完全关联，但提供基础查找）

// FindDropGroupsByItemId 通过物品ID查找掉落组
func FindDropGroupsByItemId(itemId int, container *diff.DataContainer) []*drop_group.DropGroupDiff {
	var dropGroups []*drop_group.DropGroupDiff

	// 遍历DropItemDiff查找包含该物品的掉落项
	for _, dropItem := range *container.DropItemDiff {
		for _, item := range dropItem.Item {
			// 假设ItemCfg结构有ID字段
			_ = item
			//if item.ID == itemId {
			//	// 查找对应的掉落组
			//	for _, dropGroup := range container.DropGroupDiff {
			//		if dropGroup.Id == dropItem.DropGroup {
			//			dropGroups = append(dropGroups, &dropGroup)
			//			break
			//		}
			//	}
			//	break
			//}
		}
	}

	return dropGroups
}

// FindBuffsBySkillId 通过技能ID查找Buff
func FindBuffsBySkillId(skillId string, container *diff.DataContainer) []*buff.BuffDiff {
	var buffs []*buff.BuffDiff

	// 查找技能关联的Buff
	for _, skill := range *container.SkillDiff {
		if skill.Id == skillId {
			// skill.Buff是[]string，需要解析
			for _, buffId := range skill.Buff {
				for _, buff := range *container.BuffDiff {
					// 需要确定BuffDiff的主键，假设有Id字段
					_ = buff
					_ = buffId
					//if buff.Id == buffId {
					//	buffs = append(buffs, &buff)
					//	break
					//}
				}
			}
			break
		}
	}

	return buffs
}

// FindTasksByAchieveId 通过成就ID查找任务
func FindTasksByAchieveId(achieveId int, container *diff.DataContainer) []*task_complete_cond.TaskCompleteConditonDiff {
	var tasks []*task_complete_cond.TaskCompleteConditonDiff

	for _, achieve := range *container.AchieveDiff {
		if achieve.Id == achieveId {
			// 假设通过CompleteCondId关联
			for _, task := range *container.TaskCompleteCondDiff {
				if task.Id == achieve.CompleteCondId {
					tasks = append(tasks, &task)
				}
			}
			break
		}
	}

	return tasks
}

// FindSeasonPassRewardsByHeroId 通过武将ID查找赛季通行证奖励
func FindSeasonPassRewardsByHeroId(heroId string, container *diff.DataContainer) []*season_pass_reward.SeasonPassRewardDiff {
	var rewards []*season_pass_reward.SeasonPassRewardDiff

	for _, reward := range *container.SeasonPassRewardDiff {
		// 检查NormalReward和HighReward是否包含该武将相关的物品
		for _, item := range reward.NormalReward {
			if checkItemIsHeroRelated(item.ItemId, heroId, container) {
				rewards = append(rewards, &reward)
				break
			}
		}

		if len(rewards) > 0 && rewards[len(rewards)-1] == &reward {
			continue
		}

		for _, item := range reward.HighReward {
			if checkItemIsHeroRelated(item.ItemId, heroId, container) {
				rewards = append(rewards, &reward)
				break
			}
		}
	}

	return rewards
}

// checkItemIsHeroRelated 检查物品是否与武将相关
func checkItemIsHeroRelated(itemId int, heroId string, container *diff.DataContainer) bool {
	// 检查皮肤物品
	for _, skin := range *container.HeroSkinItemDiff {
		if skin.SkinItemId == itemId && skin.HeroId == heroId {
			return true
		}
	}

	// 可以添加其他检查逻辑

	return false
}

// FindArenaRewardsByHeroId 通过武将ID查找竞技场奖励
func FindArenaRewardsByHeroId(heroId string, container *diff.DataContainer) []*arena_score_rewards.ArenaScoreRewardDiff {
	var rewards []*arena_score_rewards.ArenaScoreRewardDiff

	for _, reward := range *container.ArenaScoreRewardsDiff {
		// 检查Reward数组是否包含该武将相关的物品
		for _, item := range reward.Reward {
			_ = item
			//if checkItemIsHeroRelated(item.ID, heroId, container) {
			//	rewards = append(rewards, &reward)
			//	break
			//}
		}
	}

	return rewards
}
