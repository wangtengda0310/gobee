package drop_group

import (
	"strconv"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type DropGroupDiff struct {
	Id            int
	Name          string
	Weight        int
	WeightInc     int
	Deduplication bool
	ValidDate     string
	ExpireDate    string
}

func (d DropGroupDiff) GetType() string {
	return "DropGroupDiff"
}

func (d DropGroupDiff) GetDisplayName() string {
	return d.Name
}

func GetDropGroupDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]DropGroupDiff, err error) {
	var dropGroupCols [][]string // 掉落分组表
	if sheet, exist := sheetMap["掉落分组表|DropGroup"]; exist {
		var err error
		dropGroupCols, err = sheet.GetCols("掉落分组表|DropGroup")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建掉落分组Map
	dropGroupsDiff := make([]DropGroupDiff, 0, 300)

	for i, idStr := range dropGroupCols[Id][startRow:helpers.AutoDetectEndIndex(dropGroupCols, Id, startRow, 3)] {
		// 表格第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			dropGroupDiff := DropGroupDiff{}

			name := utils.GetCellValue(dropGroupCols, Name, startRow+i)
			weight := utils.GetCellValue(dropGroupCols, Weight, startRow+i)
			weightInc := utils.GetCellValue(dropGroupCols, WeightInc, startRow+i)
			deduplication := utils.GetCellValue(dropGroupCols, Deduplication, startRow+i)
			validDate := utils.GetCellValue(dropGroupCols, ValidDate, startRow+i)
			expireDate := utils.GetCellValue(dropGroupCols, ExpireDate, startRow+i)

			// 录入掉落分组信息
			dropGroupDiff.Id = id
			dropGroupDiff.Name = name

			if n, err := strconv.Atoi(weight); err == nil {
				dropGroupDiff.Weight = n
			} else {
				dropGroupDiff.Weight = -1
			}

			if n, err := strconv.Atoi(weightInc); err == nil {
				dropGroupDiff.WeightInc = n
			} else {
				dropGroupDiff.WeightInc = -1
			}

			if b, err := strconv.ParseBool(deduplication); err == nil && b {
				dropGroupDiff.Deduplication = b
			} else {
				dropGroupDiff.Deduplication = false
			}

			dropGroupDiff.ValidDate = validDate
			dropGroupDiff.ExpireDate = expireDate

			dropGroupsDiff = append(dropGroupsDiff, dropGroupDiff)
		}
	}

	return &dropGroupsDiff, nil
}
