package skill

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type SkillDiff struct {
	Id                string // ESkillId类型，使用string
	SkillName         string
	ShortSkillName    string
	ESkillId          string // ESkillId def
	SkillType         string // ESkillType
	IsFromOther       bool
	CounterFormula    string
	IsFromAura        bool
	TransCard         string // ESkillId
	SkillFromType     string // ESkillFromType
	ResetCounterType  string // ESkillResetCountType
	ResetTimesType    string // ESkillResetCountType
	SkillLimitTimes   int
	TotalLimitTimes   int
	TriggerCondition  []int
	SkillEffect       []string // 结构体类型，暂时用string处理
	EmptyWaitTime     string   // EEmptyWaitTime
	Buff              []string // EBuffId[]
	ShowPointAndAttr  bool
	MutexSkill        []string // ESkillId[]
	BattleCardClass   string   // ECardMarkClass
	AIJudgeArea       []string // EAIJudgeArea[]
	DeadType          int
	InitPro           []string // {int Pro;int Value}[] 结构体数组
	MagicSkillID      string   // ESkillId
	IsAutoSelAllOther bool
	IsForbidCopy      bool
	IsForbidTrans     bool
	IsForbidDestroy   bool
}

func (s SkillDiff) GetID() string {
	return s.Id
}

func (s SkillDiff) GetType() string {
	return "SkillDiff"
}

func (s SkillDiff) GetDisplayName() string {
	return s.SkillName
}

func GetSkillDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]SkillDiff, err error) {
	var skillCols [][]string // 技能表
	if sheet, exist := sheetMap["技能表|Skill"]; exist {
		var err error
		skillCols, err = sheet.GetCols("技能表|Skill")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建技能Map
	skillsDiff := make([]SkillDiff, 0, 300)

	for i, idStr := range skillCols[Id][startRow:helpers.AutoDetectEndIndex(skillCols, Id, startRow, 3)] {
		// 第一列是ESkillId枚举类型，判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		skillDiff := SkillDiff{}

		// 获取所有字段的值
		skillName := utils.GetCellValue(skillCols, SkillName, startRow+i)
		shortSkillName := utils.GetCellValue(skillCols, ShortSkillName, startRow+i)
		eSkillId := utils.GetCellValue(skillCols, ESkillId, startRow+i)
		skillType := utils.GetCellValue(skillCols, SkillType, startRow+i)
		isFromOther := utils.GetCellValue(skillCols, IsFromOther, startRow+i)
		counterFormula := utils.GetCellValue(skillCols, CounterFormula, startRow+i)
		isFromAura := utils.GetCellValue(skillCols, IsFromAura, startRow+i)
		transCard := utils.GetCellValue(skillCols, TransCard, startRow+i)
		skillFromType := utils.GetCellValue(skillCols, SkillFromType, startRow+i)
		resetCounterType := utils.GetCellValue(skillCols, ResetCounterType, startRow+i)
		resetTimesType := utils.GetCellValue(skillCols, ResetTimesType, startRow+i)
		skillLimitTimes := utils.GetCellValue(skillCols, SkillLimitTimes, startRow+i)
		totalLimitTimes := utils.GetCellValue(skillCols, TotalLimitTimes, startRow+i)
		triggerCondition := utils.GetCellValue(skillCols, TriggerCondition, startRow+i)
		skillEffect := utils.GetCellValue(skillCols, SkillEffect, startRow+i)
		emptyWaitTime := utils.GetCellValue(skillCols, EmptyWaitTime, startRow+i)
		buff := utils.GetCellValue(skillCols, Buff, startRow+i)
		showPointAndAttr := utils.GetCellValue(skillCols, ShowPointAndAttr, startRow+i)
		mutexSkill := utils.GetCellValue(skillCols, MutexSkill, startRow+i)
		battleCardClass := utils.GetCellValue(skillCols, BattleCardClass, startRow+i)
		aiJudgeArea := utils.GetCellValue(skillCols, AIJudgeArea, startRow+i)
		deadType := utils.GetCellValue(skillCols, DeadType, startRow+i)
		initPro := utils.GetCellValue(skillCols, InitPro, startRow+i)
		magicSkillID := utils.GetCellValue(skillCols, MagicSkillID, startRow+i)
		isAutoSelAllOther := utils.GetCellValue(skillCols, IsAutoSelAllOther, startRow+i)
		isForbidCopy := utils.GetCellValue(skillCols, IsForbidCopy, startRow+i)
		isForbidTrans := utils.GetCellValue(skillCols, IsForbidTrans, startRow+i)
		isForbidDestroy := utils.GetCellValue(skillCols, IsForbidDestroy, startRow+i)

		// 填充数据
		skillDiff.Id = idStr
		skillDiff.SkillName = skillName
		skillDiff.ShortSkillName = shortSkillName
		skillDiff.ESkillId = eSkillId
		skillDiff.SkillType = skillType

		if b, err := strconv.ParseBool(isFromOther); err == nil {
			skillDiff.IsFromOther = b
		} else {
			skillDiff.IsFromOther = false
		}

		skillDiff.CounterFormula = counterFormula

		if b, err := strconv.ParseBool(isFromAura); err == nil {
			skillDiff.IsFromAura = b
		} else {
			skillDiff.IsFromAura = false
		}

		skillDiff.TransCard = transCard
		skillDiff.SkillFromType = skillFromType
		skillDiff.ResetCounterType = resetCounterType
		skillDiff.ResetTimesType = resetTimesType

		if n, err := strconv.Atoi(skillLimitTimes); err == nil {
			skillDiff.SkillLimitTimes = n
		} else {
			skillDiff.SkillLimitTimes = -1
		}

		if n, err := strconv.Atoi(totalLimitTimes); err == nil {
			skillDiff.TotalLimitTimes = n
		} else {
			skillDiff.TotalLimitTimes = -1
		}

		// TriggerCondition是int[]类型
		skillDiff.TriggerCondition = make([]int, 0)
		for _, ss := range strings.Split(triggerCondition, ",") {
			if ss != "" {
				if n, err := strconv.Atoi(ss); err == nil {
					skillDiff.TriggerCondition = append(skillDiff.TriggerCondition, n)
				}
			}
		}

		// SkillEffect是结构体数组，暂时作为string处理
		skillDiff.SkillEffect = make([]string, 0)
		for _, ss := range strings.Split(skillEffect, ";") {
			if ss != "" {
				skillDiff.SkillEffect = append(skillDiff.SkillEffect, ss)
			}
		}

		skillDiff.EmptyWaitTime = emptyWaitTime

		// Buff是EBuffId[]类型
		skillDiff.Buff = make([]string, 0)
		for _, ss := range strings.Split(buff, ",") {
			if ss != "" {
				skillDiff.Buff = append(skillDiff.Buff, ss)
			}
		}

		if b, err := strconv.ParseBool(showPointAndAttr); err == nil {
			skillDiff.ShowPointAndAttr = b
		} else {
			skillDiff.ShowPointAndAttr = false
		}

		// MutexSkill是ESkillId[]类型
		skillDiff.MutexSkill = make([]string, 0)
		for _, ss := range strings.Split(mutexSkill, ",") {
			if ss != "" {
				skillDiff.MutexSkill = append(skillDiff.MutexSkill, ss)
			}
		}

		skillDiff.BattleCardClass = battleCardClass

		// AIJudgeArea是EAIJudgeArea[]类型
		skillDiff.AIJudgeArea = make([]string, 0)
		for _, ss := range strings.Split(aiJudgeArea, ",") {
			if ss != "" {
				skillDiff.AIJudgeArea = append(skillDiff.AIJudgeArea, ss)
			}
		}

		if n, err := strconv.Atoi(deadType); err == nil {
			skillDiff.DeadType = n
		} else {
			skillDiff.DeadType = -1
		}

		// InitPro是结构体数组，暂时作为string处理
		skillDiff.InitPro = make([]string, 0)
		for _, ss := range strings.Split(initPro, ";") {
			if ss != "" {
				skillDiff.InitPro = append(skillDiff.InitPro, ss)
			}
		}

		skillDiff.MagicSkillID = magicSkillID

		if b, err := strconv.ParseBool(isAutoSelAllOther); err == nil {
			skillDiff.IsAutoSelAllOther = b
		} else {
			skillDiff.IsAutoSelAllOther = false
		}

		if b, err := strconv.ParseBool(isForbidCopy); err == nil {
			skillDiff.IsForbidCopy = b
		} else {
			skillDiff.IsForbidCopy = false
		}

		if b, err := strconv.ParseBool(isForbidTrans); err == nil {
			skillDiff.IsForbidTrans = b
		} else {
			skillDiff.IsForbidTrans = false
		}

		if b, err := strconv.ParseBool(isForbidDestroy); err == nil {
			skillDiff.IsForbidDestroy = b
		} else {
			skillDiff.IsForbidDestroy = false
		}

		skillsDiff = append(skillsDiff, skillDiff)
	}

	return &skillsDiff, nil
}
