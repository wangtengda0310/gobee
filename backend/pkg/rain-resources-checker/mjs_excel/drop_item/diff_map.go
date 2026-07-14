package drop_item

import (
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type DropItemDiff struct {
	Id            int
	Name          string
	DropGroup     int
	Item          []ItemCfg
	Weight        int
	WeightInc     int
	Deduplication bool
	CheckExist    bool
	ExcludeExist  bool
	MustHave      bool
	ReplaceGroup  int
	ValidDate     string
	ExpireDate    string
}

type ItemCfg struct {
	ItemId int
	Count  int
	// 如果有其他字段可以在这里添加
}

func (d DropItemDiff) GetType() string {
	return "DropItemDiff"
}

func (d DropItemDiff) GetDisplayName() string {
	return d.Name
}

func GetDropItemDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]DropItemDiff, err error) {
	var dropItemCols [][]string // 掉落道具表
	if sheet, exist := sheetMap["掉落道具表|DropItem"]; exist {
		var err error
		dropItemCols, err = sheet.GetCols("掉落道具表|DropItem")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建掉落道具Map
	dropItemsDiff := make([]DropItemDiff, 0, 300)

	reg := regexp.MustCompile(`\{(\d+);\d+}`)

	for i, idStr := range dropItemCols[Id][startRow:helpers.AutoDetectEndIndex(dropItemCols, Id, startRow, 3)] {
		// 表格第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			dropItemDiff := DropItemDiff{}

			name := utils.GetCellValue(dropItemCols, Name, startRow+i)
			dropGroup := utils.GetCellValue(dropItemCols, DropGroup, startRow+i)
			item := utils.GetCellValue(dropItemCols, Item, startRow+i)
			weight := utils.GetCellValue(dropItemCols, Weight, startRow+i)
			weightInc := utils.GetCellValue(dropItemCols, WeightInc, startRow+i)
			deduplication := utils.GetCellValue(dropItemCols, Deduplication, startRow+i)
			checkExist := utils.GetCellValue(dropItemCols, CheckExist, startRow+i)
			excludeExist := utils.GetCellValue(dropItemCols, ExcludeExist, startRow+i)
			mustHave := utils.GetCellValue(dropItemCols, MustHave, startRow+i)
			replaceGroup := utils.GetCellValue(dropItemCols, ReplaceGroup, startRow+i)
			validDate := utils.GetCellValue(dropItemCols, ValidDate, startRow+i)
			expireDate := utils.GetCellValue(dropItemCols, ExpireDate, startRow+i)

			// 录入掉落道具信息
			dropItemDiff.Id = id
			dropItemDiff.Name = name

			if n, err := strconv.Atoi(dropGroup); err == nil {
				dropItemDiff.DropGroup = n
			} else {
				dropItemDiff.DropGroup = -1
			}

			// 解析掉落道具，格式可能是 "itemId1:count1,itemId2:count2" 或类似格式
			dropItemDiff.Item = make([]ItemCfg, 0, 5)
			if item != "" {
				// 使用正则找出所有匹配项
				matches := reg.FindAllStringSubmatch(item, -1)
				for _, match := range matches {
					if len(match) >= 2 {
						// 解析完整格式 {itemId;count}
						fullMatch := match[0] // 完整的 {数字;数字}
						itemIdStr := match[1] // 第一个数字

						// 解析数量
						count := 1
						parts := strings.Split(strings.Trim(fullMatch, "{}"), ";")
						if len(parts) >= 2 {
							if c, err := strconv.Atoi(parts[1]); err == nil {
								count = c
							}
						}

						itemId, _ := strconv.Atoi(itemIdStr)

						dropItemDiff.Item = append(dropItemDiff.Item, ItemCfg{
							ItemId: itemId,
							Count:  count,
						})
					}
				}
			}

			if n, err := strconv.Atoi(weight); err == nil {
				dropItemDiff.Weight = n
			} else {
				dropItemDiff.Weight = -1
			}

			if n, err := strconv.Atoi(weightInc); err == nil {
				dropItemDiff.WeightInc = n
			} else {
				dropItemDiff.WeightInc = -1
			}

			if b, err := strconv.ParseBool(deduplication); err == nil {
				dropItemDiff.Deduplication = b
			} else {
				dropItemDiff.Deduplication = false
			}

			if b, err := strconv.ParseBool(checkExist); err == nil {
				dropItemDiff.CheckExist = b
			} else {
				dropItemDiff.CheckExist = false
			}

			if b, err := strconv.ParseBool(excludeExist); err == nil {
				dropItemDiff.ExcludeExist = b
			} else {
				dropItemDiff.ExcludeExist = false
			}

			if b, err := strconv.ParseBool(mustHave); err == nil {
				dropItemDiff.MustHave = b
			} else {
				dropItemDiff.MustHave = false
			}

			if n, err := strconv.Atoi(replaceGroup); err == nil {
				dropItemDiff.ReplaceGroup = n
			} else {
				dropItemDiff.ReplaceGroup = -1
			}

			dropItemDiff.ValidDate = validDate
			dropItemDiff.ExpireDate = expireDate

			dropItemsDiff = append(dropItemsDiff, dropItemDiff)
		}
	}

	return &dropItemsDiff, nil
}
