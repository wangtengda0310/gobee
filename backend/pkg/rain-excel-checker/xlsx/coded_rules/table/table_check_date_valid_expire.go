package coded_rules

import (
	"fmt"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// DateValidExpireCheckRule 日期有效性检查规则（ValidDate <= ExpireDate）
//
// ## 校验规则
// - ValidDate 和 ExpireDate 均非空时，ValidDate 必须 <= ExpireDate
//
// ## 相关表结构
// - DropGroup: ValidDate, ExpireDate（掉落分组表）
// - DropItem: ValidDate, ExpireDate（掉落道具表）
//
// ## 检查流程
// 1. 查找 ValidDate 和 ExpireDate 列
// 2. 遍历数据行，当两列均非空时比较日期大小
// 3. 汇总错误并返回
type DateValidExpireCheckRule struct{}

// Meta 返回规则元数据
func (c *DateValidExpireCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.DATE_VALID_EXPIRE_CHECK,
		DisplayName:  "日期有效性检查",
		Description:  "检查 ValidDate 是否 <= ExpireDate（当两者均非空时）",
		TargetSheets: []string{"DropGroup", "DropItem"},
		ParamDefs:    []json_rule.TableRuleParamDef{},
	}
}

// Check 执行日期有效性检查
func (c *DateValidExpireCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "日期有效性检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 获取列索引
	validDateColIdx := helpers.GetColIndexByName(param.Cols, "ValidDate")
	expireDateColIdx := helpers.GetColIndexByName(param.Cols, "ExpireDate")
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")

	// 如果找不到任一日期列，跳过检查
	if validDateColIdx == -1 || expireDateColIdx == -1 {
		result.Ok = true
		result.Reason = "未找到 ValidDate 或 ExpireDate 列，跳过检查"
		return result
	}

	errorCount := 0

	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		validDateStr := helpers.GetColValue(param.Cols, validDateColIdx, rowIdx)
		expireDateStr := helpers.GetColValue(param.Cols, expireDateColIdx, rowIdx)

		// 如果任一日期为空，跳过检查
		if validDateStr == "" || expireDateStr == "" {
			continue
		}

		// 解析日期
		validDate := excelio.ParseDate(validDateStr)
		expireDate := excelio.ParseDate(expireDateStr)

		// 检查日期解析是否成功（到达此处时 validDateStr/expireDateStr 必然非空）
		if validDate.IsZero() {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("ValidDate 格式错误: %s", validDateStr),
			})
			errorCount++
			continue
		}

		if expireDate.IsZero() {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   fmt.Sprintf("ExpireDate 格式错误: %s", expireDateStr),
			})
			errorCount++
			continue
		}

		// 检查 ValidDate <= ExpireDate
		if validDate.After(expireDate) {
			rowId := ""
			if idColIdx != -1 {
				rowId = helpers.GetColValue(param.Cols, idColIdx, rowIdx)
			}

			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason: fmt.Sprintf("行 %s: ValidDate(%s) 必须 <= ExpireDate(%s)",
					rowId,
					validDate.Format("2006-01-02 15:04:05"),
					expireDate.Format("2006-01-02 15:04:05")),
			})
			errorCount++
		}
	}

	if errorCount > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个日期有效性问题", errorCount)
	}

	return result
}
