package skill_ui

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillUIDiff struct {
	Id              string   // ESkillId 作为主键
	SkillName       string   // 技能名称
	PlayCardAudio   string   // 出牌音效
	IdentityLine    []string // 身份技能台词
	None            string   // 空列2
	Audio           []int    // 打牌时触发技能的台词
	SkillText       string   // 技能文案
	ShortSkillText  string   // 技能文案（短）
	KeyWords        []int    // 技能关键字
	SettlementDes   string   // 结算详情
	Allusion        string   // 技能典故
	DesignThought   string   // 设计思路
	SkillTag        []int    // 技能标签
	BattleSkillStep []string // 战斗内阶段显示
	HasRelation     bool     // 是否有关联技能
	RelatedSkill    []string // 关联技能技能ID
	SpecialAudio    string   // 技能特殊音效
	ESkillId        string   // E#SkillId
	EAudio          string   // E#AudioId
	EAudioId        string   // E#AudioId[]
	ESkillTag       string   // E#SkillTag[]
}

func (s SkillUIDiff) GetID() string {
	return s.Id
}

func (s SkillUIDiff) GetType() string {
	return "SkillUIDiff"
}

func (s SkillUIDiff) GetDisplayName() string {
	return s.SkillName
}

func GetSkillUIDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SkillUIDiff, err error) {
	var skillCols [][]string
	if sheet, exist := sheetMap["技能表现配置表|SkillUI"]; exist {
		var err error
		skillCols, err = sheet.GetCols("技能表现配置表|SkillUI")
		if err != nil {
			return nil, err
		}
	}

	// 固定行数
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能Map
	skillsDiff := make([]SkillUIDiff, 0, 300)

	for i, idStr := range skillCols[Id][startRow:helpers.AutoDetectEndIndex(skillCols, Id, startRow, 3)] {
		// 第一列是ESkillId（枚举类型），判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		skillDiff := SkillUIDiff{}

		// 获取所有字段的值
		skillName := utils.GetCellValue(skillCols, SkillName, startRow+i)
		playCardAudio := utils.GetCellValue(skillCols, PlayCardAudio, startRow+i)
		identityLine := utils.GetCellValue(skillCols, IdentityLine, startRow+i)
		none := utils.GetCellValue(skillCols, None, startRow+i)
		audio := utils.GetCellValue(skillCols, Audio, startRow+i)
		skillText := utils.GetCellValue(skillCols, SkillText, startRow+i)
		shortSkillText := utils.GetCellValue(skillCols, ShortSkillText, startRow+i)
		keyWords := utils.GetCellValue(skillCols, KeyWords, startRow+i)
		settlementDes := utils.GetCellValue(skillCols, SettlementDes, startRow+i)
		allusion := utils.GetCellValue(skillCols, Allusion, startRow+i)
		designThought := utils.GetCellValue(skillCols, DesignThought, startRow+i)
		skillTag := utils.GetCellValue(skillCols, SkillTag, startRow+i)
		battleSkillStep := utils.GetCellValue(skillCols, BattleSkillStep, startRow+i)
		hasRelation := utils.GetCellValue(skillCols, HasRelation, startRow+i)
		relatedSkill := utils.GetCellValue(skillCols, RelatedSkill, startRow+i)
		specialAudio := utils.GetCellValue(skillCols, SpecialAudio, startRow+i)
		eSkillId := utils.GetCellValue(skillCols, ESkillId, startRow+i)
		eAudio := utils.GetCellValue(skillCols, EAudio, startRow+i)
		eAudioId := utils.GetCellValue(skillCols, EAudioId, startRow+i)
		eSkillTag := utils.GetCellValue(skillCols, ESkillTag, startRow+i)

		// 赋值
		skillDiff.Id = idStr
		skillDiff.SkillName = skillName
		skillDiff.PlayCardAudio = playCardAudio

		// BattleSkillStep是string数组
		skillDiff.IdentityLine = strings.Split(identityLine, ",")
		if len(skillDiff.IdentityLine) == 1 && skillDiff.IdentityLine[0] == "" {
			skillDiff.IdentityLine = []string{}
		}

		skillDiff.None = none

		// Audio是int数组
		skillDiff.Audio = make([]int, 0, 5)
		for _, ss := range strings.Split(audio, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					skillDiff.Audio = append(skillDiff.Audio, n)
				}
			}
		}

		skillDiff.SkillText = skillText
		skillDiff.ShortSkillText = shortSkillText

		// KeyWords是int数组
		skillDiff.KeyWords = make([]int, 0, 5)
		for _, ss := range strings.Split(keyWords, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					skillDiff.KeyWords = append(skillDiff.KeyWords, n)
				}
			}
		}

		skillDiff.SettlementDes = settlementDes
		skillDiff.Allusion = allusion
		skillDiff.DesignThought = designThought

		// SkillTag是int数组
		skillDiff.SkillTag = make([]int, 0, 5)
		for _, ss := range strings.Split(skillTag, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					skillDiff.SkillTag = append(skillDiff.SkillTag, n)
				}
			}
		}

		// BattleSkillStep是string数组
		skillDiff.BattleSkillStep = strings.Split(battleSkillStep, ",")
		if len(skillDiff.BattleSkillStep) == 1 && skillDiff.BattleSkillStep[0] == "" {
			skillDiff.BattleSkillStep = []string{}
		}

		// HasRelation是bool
		if b, err := strconv.ParseBool(hasRelation); err == nil {
			skillDiff.HasRelation = b
		} else {
			skillDiff.HasRelation = false
		}

		// RelatedSkill是string数组
		skillDiff.RelatedSkill = strings.Split(relatedSkill, ",")
		if len(skillDiff.RelatedSkill) == 1 && skillDiff.RelatedSkill[0] == "" {
			skillDiff.RelatedSkill = []string{}
		}

		skillDiff.SpecialAudio = specialAudio
		skillDiff.ESkillId = eSkillId
		skillDiff.EAudio = eAudio
		skillDiff.EAudioId = eAudioId
		skillDiff.ESkillTag = eSkillTag

		skillsDiff = append(skillsDiff, skillDiff)
	}

	return &skillsDiff, nil
}
