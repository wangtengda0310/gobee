package achieve

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

// AchieveDiff 成就表数据结构
type AchieveDiff struct {
	Id             int      // 成就id
	Name           string   // 成就名称
	IsHide         bool     // 是否隐藏
	CompleteCondId int      // 成就完成条件
	Reward         []string // 成就奖励
	None           string   // 空列
	Des            string   // 成就描述
	Condition      string   // 成就完成条件
	HeroItemId     []int    // 关联武将ID
	OpenDate       string   // 开放时间
}

// GetType 获取类型
func (a AchieveDiff) GetType() string {
	return "AchieveDiff"
}

// GetDisplayName 获取显示名称
func (a AchieveDiff) GetDisplayName() string {
	return a.Name
}

// GetAchieveDiffMap 获取成就表数据
func GetAchieveDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]AchieveDiff, err error) {
	var achieveCols [][]string // 成就表
	if sheet, exist := sheetMap["成就表|Achieve"]; exist {
		var err error
		achieveCols, err = sheet.GetCols("成就表|Achieve")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建成就Map
	achievesDiff := make([]AchieveDiff, 0, 100)

	for i, idStr := range achieveCols[Id][startRow:helpers.AutoDetectEndIndex(achieveCols, Id, startRow, 3)] {
		// 第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			achieveDiff := AchieveDiff{}

			// 读取各列数据
			name := utils.GetCellValue(achieveCols, Name, startRow+i)
			isHide := utils.GetCellValue(achieveCols, IsHide, startRow+i)
			completeCondId := utils.GetCellValue(achieveCols, CompleteCondId, startRow+i)
			reward := utils.GetCellValue(achieveCols, Reward, startRow+i)
			none := utils.GetCellValue(achieveCols, None, startRow+i)
			des := utils.GetCellValue(achieveCols, Des, startRow+i)
			condition := utils.GetCellValue(achieveCols, Condition, startRow+i)
			heroItemId := utils.GetCellValue(achieveCols, HeroItemId, startRow+i)
			openDate := utils.GetCellValue(achieveCols, OpenDate, startRow+i)

			// 赋值
			achieveDiff.Id = id
			achieveDiff.Name = name

			// IsHide 布尔值处理
			if b, err := strconv.ParseBool(isHide); err == nil {
				achieveDiff.IsHide = b
			} else {
				achieveDiff.IsHide = false
			}

			// CompleteCondId 整数处理
			if n, err := strconv.Atoi(completeCondId); err == nil {
				achieveDiff.CompleteCondId = n
			} else {
				achieveDiff.CompleteCondId = -1
			}

			// Reward 字符串数组处理（ItemCfg[]类型）
			achieveDiff.Reward = make([]string, 0, 5)
			for _, ss := range strings.Split(reward, ",") {
				if ss != "" {
					achieveDiff.Reward = append(achieveDiff.Reward, strings.TrimSpace(ss))
				}
			}

			// 空列处理
			achieveDiff.None = none

			// 字符串字段直接赋值
			achieveDiff.Des = des
			achieveDiff.Condition = condition

			// HeroItemId 整数数组处理
			achieveDiff.HeroItemId = make([]int, 0, 5)
			for _, ss := range strings.Split(heroItemId, ",") {
				if n, err := strconv.Atoi(strings.TrimSpace(ss)); err == nil {
					achieveDiff.HeroItemId = append(achieveDiff.HeroItemId, n)
				}
			}

			// OpenDate 字符串处理
			achieveDiff.OpenDate = openDate

			achievesDiff = append(achievesDiff, achieveDiff)
		}
	}

	return &achievesDiff, nil
}
