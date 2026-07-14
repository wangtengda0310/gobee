package hero

import (
	"fmt"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// HeroSkillBuffCheckRule 武将技能和Buff引用完整性检查
//
// ## 校验规则
// 1. Hero.Skill 列中的每个技能ID必须存在于 Skill.xlsx 的 Id 列
// 2. Hero.Buff 列中的值（非空时）必须存在于 Buff.xlsx 的第2列（英文字符串标识符列）
//
// ## 相关表结构
// - Hero: Id, Name, Skill(逗号分隔数字ID), Buff(英文字符串标识符)
// - Skill: Id(数字)
// - Buff: Id(数字), 第2列(英文字符串标识符)
//
// ## 检查流程
// 1. 加载 Skill 表，构建有效技能 ID 集合
// 2. 加载 Buff 表，构建有效 Buff 标识符集合
// 3. 遍历 Hero 表的每一行，分别检查 Skill 和 Buff 列
// 4. 记录所有不存在的引用到错误列表
type HeroSkillBuffCheckRule struct{}

// Meta 返回规则元数据
func (c *HeroSkillBuffCheckRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:           json_rule.HERO_SKILL_BUFF_CHECK,
		DisplayName:    "武将技能和Buff引用检查",
		Description:    "检查武将表(Hero)中配置的技能ID是否存在于技能表(Skill)，Buff标识符是否存在于Buff表",
		TargetSheets:   []string{"Hero"},
		RequiredSheets: []string{"Skill", "Buff"},
		ParamDefs:      []json_rule.TableRuleParamDef{},
	}
}

// Check 执行武将技能和Buff引用完整性检查
//
// 执行流程：
// 1. 加载 Skill 表和 Buff 表数据
// 2. 构建有效技能 ID 集合和 Buff 标识符集合
// 3. 查找 Hero 表中的 Skill、Buff、Id、Name 列
// 4. 遍历 Hero 表数据行，逐行检查 Skill 和 Buff 引用
// 5. 汇总错误并返回
func (c *HeroSkillBuffCheckRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		SheetName:   &param.SheetName,
		DisplayName: "武将技能和Buff引用检查",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 加载 Skill 表
	skillCols := c.loadSheet(param.SheetMap, "Skill")
	if skillCols == nil {
		result.Ok = false
		result.Reason = "未找到 Skill 表，无法执行技能引用检查"
		return result
	}

	// 加载 Buff 表
	buffCols := c.loadSheet(param.SheetMap, "Buff")
	if buffCols == nil {
		result.Ok = false
		result.Reason = "未找到 Buff 表，无法执行Buff引用检查"
		return result
	}

	// 构建有效数据集合
	validSkillIds := c.buildValidSkillIdSet(skillCols, param.StartRowIdx)
	validBuffIds := c.buildValidBuffIdSet(buffCols, param.StartRowIdx)

	// 查找 Hero 表关键列
	skillColIdx := helpers.GetColIndexByName(param.Cols, "Skill")
	buffColIdx := helpers.GetColIndexByName(param.Cols, "Buff")
	idColIdx := helpers.GetColIndexByName(param.Cols, "Id")
	nameColIdx := helpers.GetColIndexByName(param.Cols, "Name")

	// 遍历检查（使用调用方传入的 EndIndex）
	for rowIdx := param.StartRowIdx; rowIdx < param.EndIndex; rowIdx++ {
		rowId := helpers.GetColValue(param.Cols, idColIdx, rowIdx)
		rowName := helpers.GetColValue(param.Cols, nameColIdx, rowIdx)
		var errors []string

		// 检查 Skill 列
		if skillColIdx >= 0 {
			if errs := c.checkSkillRef(param.Cols, skillColIdx, rowIdx, rowId, rowName, validSkillIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		// 检查 Buff 列
		if buffColIdx >= 0 {
			if errs := c.checkBuffRef(param.Cols, buffColIdx, rowIdx, rowId, rowName, validBuffIds); len(errs) > 0 {
				errors = append(errors, errs...)
			}
		}

		if len(errors) > 0 {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    rowIdx,
				ExcelRow: rowIdx + 1,
				Reason:   strings.Join(errors, "; "),
			})
		}
	}

	if len(result.ErrCells) > 0 {
		result.Ok = false
		result.Reason = fmt.Sprintf("发现 %d 个武将的技能或Buff引用问题", len(result.ErrCells))
	}

	return result
}

// loadSheet 从 sheetMap 中加载指定表的数据
func (c *HeroSkillBuffCheckRule) loadSheet(sheetMap map[string]*excelize.File, suffix string) [][]string {
	if file, sheetName, ok := helpers.FindSheetBySuffix(sheetMap, suffix); ok {
		cols, err := file.GetCols(sheetName)
		if err == nil {
			return cols
		}
	}
	return nil
}

// buildValidSkillIdSet 从 Skill 表构建有效的技能 ID 集合
// Skill 表的 Id 列为数字类型
func (c *HeroSkillBuffCheckRule) buildValidSkillIdSet(skillCols [][]string, startRowIdx int) map[int]bool {
	validIds := make(map[int]bool)
	idColIdx := helpers.GetColIndexByName(skillCols, "Id")
	if idColIdx < 0 {
		return validIds
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(skillCols, startRowIdx); rowIdx++ {
		idStr := helpers.GetColValue(skillCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		validIds[id] = true
	}
	return validIds
}

// buildValidBuffIdSet 从 Buff 表构建有效的 Buff 标识符集合
// Buff 表的第2列（列索引1）存储英文字符串标识符，如 BuffBossNianShou2026
// 该列没有列名（第3行为 None），所以直接用索引 1
func (c *HeroSkillBuffCheckRule) buildValidBuffIdSet(buffCols [][]string, startRowIdx int) map[string]bool {
	validIds := make(map[string]bool)
	buffIdColIdx := 1 // Buff 表英文字符串标识符在第2列（0-based 索引 1）
	if len(buffCols) <= buffIdColIdx {
		return validIds
	}

	for rowIdx := startRowIdx; rowIdx < helpers.GetDataEndIndex(buffCols, startRowIdx); rowIdx++ {
		buffId := helpers.GetColValue(buffCols, buffIdColIdx, rowIdx)
		if buffId == "" {
			continue
		}
		validIds[buffId] = true
	}
	return validIds
}

// checkSkillRef 检查单行的 Skill 列引用
func (c *HeroSkillBuffCheckRule) checkSkillRef(cols [][]string, skillColIdx, rowIdx int, rowId, rowName string, validSkillIds map[int]bool) []string {
	skillStr := helpers.GetColValue(cols, skillColIdx, rowIdx)
	if skillStr == "" {
		return nil
	}

	// 解析逗号分隔的技能 ID
	parts := strings.Split(skillStr, ",")
	var missing []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			missing = append(missing, fmt.Sprintf("无效技能ID[%s]", part))
			continue
		}
		if !validSkillIds[id] {
			missing = append(missing, fmt.Sprintf("技能ID[%s]在技能表|Skill中不存在", part))
		}
	}

	if len(missing) > 0 {
		return []string{fmt.Sprintf("武将【%s】(ID=%s) %s", rowName, rowId, strings.Join(missing, ", "))}
	}
	return nil
}

// checkBuffRef 检查单行的 Buff 列引用
func (c *HeroSkillBuffCheckRule) checkBuffRef(cols [][]string, buffColIdx, rowIdx int, rowId, rowName string, validBuffIds map[string]bool) []string {
	buffStr := helpers.GetColValue(cols, buffColIdx, rowIdx)
	if buffStr == "" {
		return nil
	}

	if !validBuffIds[buffStr] {
		return []string{fmt.Sprintf("武将【%s】(ID=%s) Buff标识符[%s]在Buff表中不存在", rowName, rowId, buffStr)}
	}
	return nil
}
