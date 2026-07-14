package buff

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type BuffDiff struct {
	Id                     string
	EBuffId                string
	Name                   string
	NeedRecord             bool
	BuffType               int
	IsDeleteByCasterDead   bool
	IsDeleteByExecutorDead bool
	IsValidByFengJin       bool
	IsReserveByRemoveSkill bool
	TransferSkillBuffType  int
	Round                  int
	Value                  int
	EndByEffect            bool
	EndType                string
	IsServerOnly           bool
	IsCasterOnly           bool
	OwnerType              int
	ShowArea               string
	Icon                   string
	Effect                 int
	FlashEffect            int
	BuffPriority           int
	OverlyingType          int
	IsTrigger              bool
	BuffState              []int
	BuffPro                []int
	ProValue               []int
	CostEffectValue        []int
	TriggerTiming          []int
	TriggerPriority        []int
	TriggerCondition       []string
	TriggerAction          []int
	BuffDot                int
	EffectDescribe         string
}

func (b BuffDiff) GetID() string {
	return b.Id
}

func (b BuffDiff) GetType() string {
	return "BuffDiff"
}

func (b BuffDiff) GetDisplayName() string {
	return b.Name
}

func GetBuffDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]BuffDiff, err error) {
	var buffCols [][]string // Buff表
	if sheet, exist := sheetMap["Buff表|Buff"]; exist {
		var err error
		buffCols, err = sheet.GetCols("Buff表|Buff")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建BuffMap
	buffsDiff := make([]BuffDiff, 0, 300)

	for i, idStr := range buffCols[Id][startRow:helpers.AutoDetectEndIndex(buffCols, Id, startRow, 3)] {
		// 第一列是E#开头的枚举类型，判断规则：不为空且不以#开头
		if idStr == "" || strings.HasPrefix(idStr, "#") {
			continue
		}

		buffDiff := BuffDiff{}

		eBuffId := utils.GetCellValue(buffCols, EBuffId, startRow+i)
		name := utils.GetCellValue(buffCols, Name, startRow+i)
		needRecord := utils.GetCellValue(buffCols, NeedRecord, startRow+i)
		buffType := utils.GetCellValue(buffCols, BuffType, startRow+i)
		isDeleteByCasterDead := utils.GetCellValue(buffCols, IsDeleteByCasterDead, startRow+i)
		isDeleteByExecutorDead := utils.GetCellValue(buffCols, IsDeleteByExecutorDead, startRow+i)
		isValidByFengJin := utils.GetCellValue(buffCols, IsValidByFengJin, startRow+i)
		isReserveByRemoveSkill := utils.GetCellValue(buffCols, IsReserveByRemoveSkill, startRow+i)
		transferSkillBuffType := utils.GetCellValue(buffCols, TransferSkillBuffType, startRow+i)
		round := utils.GetCellValue(buffCols, Round, startRow+i)
		value := utils.GetCellValue(buffCols, Value, startRow+i)
		endByEffect := utils.GetCellValue(buffCols, EndByEffect, startRow+i)
		endType := utils.GetCellValue(buffCols, EndType, startRow+i)
		isServerOnly := utils.GetCellValue(buffCols, IsServerOnly, startRow+i)
		isCasterOnly := utils.GetCellValue(buffCols, IsCasterOnly, startRow+i)
		ownerType := utils.GetCellValue(buffCols, OwnerType, startRow+i)
		showArea := utils.GetCellValue(buffCols, ShowArea, startRow+i)
		icon := utils.GetCellValue(buffCols, Icon, startRow+i)
		effect := utils.GetCellValue(buffCols, Effect, startRow+i)
		flashEffect := utils.GetCellValue(buffCols, FlashEffect, startRow+i)
		buffPriority := utils.GetCellValue(buffCols, BuffPriority, startRow+i)
		overlyingType := utils.GetCellValue(buffCols, OverlyingType, startRow+i)
		isTrigger := utils.GetCellValue(buffCols, IsTrigger, startRow+i)
		buffState := utils.GetCellValue(buffCols, BuffState, startRow+i)
		buffPro := utils.GetCellValue(buffCols, BuffPro, startRow+i)
		proValue := utils.GetCellValue(buffCols, ProValue, startRow+i)
		costEffectValue := utils.GetCellValue(buffCols, CostEffectValue, startRow+i)
		triggerTiming := utils.GetCellValue(buffCols, TriggerTiming, startRow+i)
		triggerPriority := utils.GetCellValue(buffCols, TriggerPriority, startRow+i)
		triggerCondition := utils.GetCellValue(buffCols, TriggerCondition, startRow+i)
		triggerAction := utils.GetCellValue(buffCols, TriggerAction, startRow+i)
		buffDot := utils.GetCellValue(buffCols, BuffDot, startRow+i)
		effectDescribe := utils.GetCellValue(buffCols, EffectDescribe, startRow+i)

		// 录入Buff信息
		buffDiff.Id = idStr
		buffDiff.Name = name
		buffDiff.EBuffId = eBuffId

		if b, err := strconv.ParseBool(needRecord); err == nil {
			buffDiff.NeedRecord = b
		} else {
			buffDiff.NeedRecord = false
		}

		if n, err := strconv.Atoi(buffType); err == nil {
			buffDiff.BuffType = n
		} else {
			buffDiff.BuffType = -1
		}

		if b, err := strconv.ParseBool(isDeleteByCasterDead); err == nil {
			buffDiff.IsDeleteByCasterDead = b
		} else {
			buffDiff.IsDeleteByCasterDead = false
		}

		if b, err := strconv.ParseBool(isDeleteByExecutorDead); err == nil {
			buffDiff.IsDeleteByExecutorDead = b
		} else {
			buffDiff.IsDeleteByExecutorDead = false
		}

		if b, err := strconv.ParseBool(isValidByFengJin); err == nil {
			buffDiff.IsValidByFengJin = b
		} else {
			buffDiff.IsValidByFengJin = false
		}

		if b, err := strconv.ParseBool(isReserveByRemoveSkill); err == nil {
			buffDiff.IsReserveByRemoveSkill = b
		} else {
			buffDiff.IsReserveByRemoveSkill = false
		}

		if n, err := strconv.Atoi(transferSkillBuffType); err == nil {
			buffDiff.TransferSkillBuffType = n
		} else {
			buffDiff.TransferSkillBuffType = -1
		}

		if n, err := strconv.Atoi(round); err == nil {
			buffDiff.Round = n
		} else {
			buffDiff.Round = -1
		}

		if n, err := strconv.Atoi(value); err == nil {
			buffDiff.Value = n
		} else {
			buffDiff.Value = -1
		}

		if b, err := strconv.ParseBool(endByEffect); err == nil {
			buffDiff.EndByEffect = b
		} else {
			buffDiff.EndByEffect = false
		}

		buffDiff.EndType = endType

		if b, err := strconv.ParseBool(isServerOnly); err == nil {
			buffDiff.IsServerOnly = b
		} else {
			buffDiff.IsServerOnly = false
		}

		if b, err := strconv.ParseBool(isCasterOnly); err == nil {
			buffDiff.IsCasterOnly = b
		} else {
			buffDiff.IsCasterOnly = false
		}

		if n, err := strconv.Atoi(ownerType); err == nil {
			buffDiff.OwnerType = n
		} else {
			buffDiff.OwnerType = -1
		}

		buffDiff.ShowArea = showArea
		buffDiff.Icon = icon

		if n, err := strconv.Atoi(effect); err == nil {
			buffDiff.Effect = n
		} else {
			buffDiff.Effect = -1
		}

		if n, err := strconv.Atoi(flashEffect); err == nil {
			buffDiff.FlashEffect = n
		} else {
			buffDiff.FlashEffect = -1
		}

		if n, err := strconv.Atoi(buffPriority); err == nil {
			buffDiff.BuffPriority = n
		} else {
			buffDiff.BuffPriority = -1
		}

		if n, err := strconv.Atoi(overlyingType); err == nil {
			buffDiff.OverlyingType = n
		} else {
			buffDiff.OverlyingType = -1
		}

		if b, err := strconv.ParseBool(isTrigger); err == nil {
			buffDiff.IsTrigger = b
		} else {
			buffDiff.IsTrigger = false
		}

		// 处理数组类型字段
		buffDiff.BuffState = make([]int, 0)
		for _, ss := range strings.Split(buffState, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.BuffState = append(buffDiff.BuffState, n)
			}
		}

		buffDiff.BuffPro = make([]int, 0)
		for _, ss := range strings.Split(buffPro, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.BuffPro = append(buffDiff.BuffPro, n)
			}
		}

		buffDiff.ProValue = make([]int, 0)
		for _, ss := range strings.Split(proValue, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.ProValue = append(buffDiff.ProValue, n)
			}
		}

		buffDiff.CostEffectValue = make([]int, 0)
		for _, ss := range strings.Split(costEffectValue, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.CostEffectValue = append(buffDiff.CostEffectValue, n)
			}
		}

		buffDiff.TriggerTiming = make([]int, 0)
		for _, ss := range strings.Split(triggerTiming, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.TriggerTiming = append(buffDiff.TriggerTiming, n)
			}
		}

		buffDiff.TriggerPriority = make([]int, 0)
		for _, ss := range strings.Split(triggerPriority, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.TriggerPriority = append(buffDiff.TriggerPriority, n)
			}
		}

		buffDiff.TriggerCondition = strings.Split(triggerCondition, ",")

		buffDiff.TriggerAction = make([]int, 0)
		for _, ss := range strings.Split(triggerAction, ",") {
			if n, err := strconv.Atoi(ss); err == nil {
				buffDiff.TriggerAction = append(buffDiff.TriggerAction, n)
			}
		}

		if n, err := strconv.Atoi(buffDot); err == nil {
			buffDiff.BuffDot = n
		} else {
			buffDiff.BuffDot = -1
		}

		buffDiff.EffectDescribe = effectDescribe

		buffsDiff = append(buffsDiff, buffDiff)
	}

	return &buffsDiff, nil
}
