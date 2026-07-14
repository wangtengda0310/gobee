package task_complete_cond

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type TaskCompleteConditonDiff struct {
	Id                int
	CondDes           string
	CompleteCond      string
	CompleteCondParam []int
	JumpCond          string
	JumpParm          []int
}

func (t TaskCompleteConditonDiff) GetType() string {
	return "TaskCompleteConditonDiff"
}

func (t TaskCompleteConditonDiff) GetDisplayName() string {
	return t.CondDes
}

func GetTaskCompleteConditonDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]TaskCompleteConditonDiff, err error) {
	var taskCols [][]string // 任务完成条件表
	if sheet, exist := sheetMap["任务完成条件表|TaskCompleteConditon"]; exist {
		var err error
		taskCols, err = sheet.GetCols("任务完成条件表|TaskCompleteConditon")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建任务完成条件Map
	tasksDiff := make([]TaskCompleteConditonDiff, 0, 100)

	// 第一列是int类型的Id
	for i, idStr := range taskCols[Id][startRow:helpers.AutoDetectEndIndex(taskCols, Id, startRow, 3)] {
		// 表格第一列是否有id作为读取这一列的标准
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			taskDiff := TaskCompleteConditonDiff{}

			condDes := utils.GetCellValue(taskCols, CondDes, startRow+i)
			completeCond := utils.GetCellValue(taskCols, CompleteCond, startRow+i)
			completeCondParam := utils.GetCellValue(taskCols, CompleteCondParam, startRow+i)
			jumpCond := utils.GetCellValue(taskCols, JumpCond, startRow+i)
			jumpParm := utils.GetCellValue(taskCols, JumpParm, startRow+i)

			// 录入任务完成条件信息
			taskDiff.Id = id
			taskDiff.CondDes = condDes
			taskDiff.CompleteCond = completeCond

			// 处理任务完成条件参数（int数组）
			taskDiff.CompleteCondParam = make([]int, 0, 5)
			if completeCondParam != "" {
				for _, ss := range strings.Split(completeCondParam, ",") {
					if n, err := strconv.Atoi(ss); err == nil {
						taskDiff.CompleteCondParam = append(taskDiff.CompleteCondParam, n)
					}
				}
			}

			taskDiff.JumpCond = jumpCond

			// 处理跳转参数（int数组）
			taskDiff.JumpParm = make([]int, 0, 5)
			if jumpParm != "" {
				for _, ss := range strings.Split(jumpParm, ",") {
					if n, err := strconv.Atoi(ss); err == nil {
						taskDiff.JumpParm = append(taskDiff.JumpParm, n)
					}
				}
			}

			tasksDiff = append(tasksDiff, taskDiff)
		}
	}

	return &tasksDiff, nil
}
