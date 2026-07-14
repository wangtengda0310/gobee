package mjs_excel

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/audio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_spine"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_lines"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/skill_ui"
	"github.com/xuri/excelize/v2"
)

type AudioRefExcel struct {
	HeroIdOpenMap map[int]bool   // 武将Id=>是否开放
	HeroSkillsMap map[int][]int  // 武将Id=>技能INT[]
	HeroNameMap   map[int]string // 武将Id=>技能INT[]

	SkillEnumIdMap map[string]int // 技能枚举=>技能ID
	SkillIdEnumMap map[int]string // 技能ID=>技能枚举
	SkillIdNameMap map[int]string // 技能ID=>中文名

	SkillIdLineVoiceIdMap map[int]skill_lines.SkillLineVoice // 技能枚举INT=>台词ID

	HeroSkinLineVoiceIdMap map[int]hero_skin_item.SkinLineVoiceMap // 武将Id=>皮肤枚举INT=>台词ID
	SkinIdNameMap          map[int]string                          // 皮肤枚举INT=>台词枚举
	SkinIdOpenMap          map[int]bool                            // 皮肤枚举INT=>皮肤是否开放

	LineVoiceIdEnumMap map[int]string // 台词Id=>音效枚举
	LineVoiceEnumIdMap map[string]int // 音效枚举=>台词Id

	AudioEnumPathMap     map[string][]string // 音效枚举=>台词路径数组
	AudioEnumTypeEnumMap map[string]string   // 音效枚举=>台词类型(用于确定资源目录)

	SkinIdSpineAnimAudioMap map[int][]string // 皮肤道具ID=>动画音效枚举
	SkinIdKillAudioMap      map[int]string   // 皮肤道具ID=>击杀动画音效枚举

	SkillEnumPlayCardAudioEnumMap     map[string]string   // 技能ESkillId枚举=>出牌音效枚举
	SkillEnumIdentityLineVoiceEnumMap map[string][]string // 技能ESkillId枚举=>身份技能音效枚举数组
	SkillEnumLineVoiceIdMap           map[string][]int    // 技能ESkillId枚举=>出发技能台词ID数组
	SkillEnumSpecialAudioEnumMap      map[string]string   // 技能ESkillId枚举=>特殊音效枚举
}

func InitAudioRefExcel(sheetMap map[string]*excelize.File) (*AudioRefExcel, error) {
	heroIdOpenMap, heroSkillsMap, heroNameMap, err := hero.GetHeroMap(sheetMap)
	if err != nil {
		return nil, err
	}

	skillEnumIdMap, skillIdEnumMap, skillIdNameMap, err := skill.GetSkillMap(sheetMap)
	if err != nil {
		return nil, err
	}

	heroSkinLineVoiceMap, skinIdNameMap, skinIdOpenMap, err := hero_skin_item.GetHeroSkinsItemVoiceMap(sheetMap)
	if err != nil {
		return nil, err
	}

	skillIdLineVoiceMap, err := skill_lines.GetSkillLinesMap(sheetMap, skillEnumIdMap)
	if err != nil {
		return nil, err
	}

	lineVoiceIdEnumMap, lineVoiceEnumIdMap, err := hero_lines.GetHeroLinesMap(sheetMap)
	if err != nil {
		return nil, err
	}

	audioEnumPathMap, audioEnumTypeEnumMap, err := audio.GetAudioMap(sheetMap)
	if err != nil {
		return nil, err
	}

	skinIdSpineAnimAudioMap, skinIdKillAudioMap, err := hero_skin_spine.GetHeroSkinSpineVoiceMap(sheetMap)
	if err != nil {
		return nil, err
	}

	skillEnumPlayCardAudioEnumMap, skillEnumIdentityLineVoiceEnumMap, skillEnumLineVoiceIdMap, skillEnumSpecialAudioEnumMap, err := skill_ui.GetSkillUIMap(sheetMap)
	if err != nil {
		return nil, err
	}

	excel := &AudioRefExcel{
		heroIdOpenMap,
		heroSkillsMap,
		heroNameMap,
		skillEnumIdMap,
		skillIdEnumMap,
		skillIdNameMap,
		skillIdLineVoiceMap,
		heroSkinLineVoiceMap,
		skinIdNameMap,
		skinIdOpenMap,
		lineVoiceIdEnumMap,
		lineVoiceEnumIdMap,
		audioEnumPathMap,
		audioEnumTypeEnumMap,
		skinIdSpineAnimAudioMap,
		skinIdKillAudioMap,
		skillEnumPlayCardAudioEnumMap,
		skillEnumIdentityLineVoiceEnumMap,
		skillEnumLineVoiceIdMap,
		skillEnumSpecialAudioEnumMap,
	}
	return excel, nil
}
