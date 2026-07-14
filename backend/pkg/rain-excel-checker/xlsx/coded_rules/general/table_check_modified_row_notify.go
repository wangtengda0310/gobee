package coded_rules

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ModifiedRowNotifyRule 修改行通知规则
// 检测 Excel 表中字段值发生变化的行，生成变更通知
// 原 ROW_CHANGE_NOTIFY 的拆分版本
type ModifiedRowNotifyRule struct{}

func (c *ModifiedRowNotifyRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.MODIFIED_ROW_NOTIFY,
		DisplayName:  "修改行通知",
		Description:  "检测Excel表中字段值发生变化的行，生成变更通知。仅在有字段变更时产生通知。",
		TargetSheets: []string{},
		ParamDefs: []json_rule.TableRuleParamDef{
			{Key: json_rule.ID_COL_NAME, Label: "ID列名", Description: "行ID列名", Type: "string", Default: "Id", Required: false},
			{Key: json_rule.ID_COL_NAMES, Label: "复合主键列名", Description: "复合主键列名（逗号分隔），优先级高于ID列名", Type: "string", Default: "", Required: false},
			{Key: json_rule.NAME_COL_NAME, Label: "名称列名", Description: "行名称列名", Type: "string", Default: "Name", Required: false},
			{Key: json_rule.GIT_REPO_PATH, Label: "Git仓库路径", Description: "Git仓库路径", Type: "string", Default: ".", Required: false},
			{Key: json_rule.BASE_COMMIT, Label: "基准commit", Description: "对比的基准commit", Type: "string", Default: "HEAD~1", Required: false},
			{Key: json_rule.HEAD_COMMIT, Label: "当前commit", Description: "当前检查的commit", Type: "string", Default: "", Required: false},
		},
	}
}

func (c *ModifiedRowNotifyRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		RuleType:    json_rule.MODIFIED_ROW_NOTIFY,
		DisplayName: "修改行通知",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	idColNames, nameColName, gitRepoPath, baseCommit, headCommit := diff.ParseDiffParams(param.Params)
	excelPath, xlsxFile := diff.ResolveExcelPath(param.SheetMap, param.SheetName)
	if xlsxFile == nil {
		result.Ok = false
		result.Reason = fmt.Sprintf("未找到包含 sheet %s 的 Excel 文件", param.SheetName)
		return result
	}

	key := diff.BuildDiffCacheKey(excelPath, baseCommit, param.SheetName, nameColName, idColNames)
	entry := diff.GetOrComputeDiff(key, diff.DiffComputeParam{
		GitRepoPath: gitRepoPath,
		BaseCommit:  baseCommit,
		HeadCommit:  headCommit,
		SheetName:   param.SheetName,
		Cols:        param.Cols,
		StartRowIdx: param.StartRowIdx,
		IdColNames:  idColNames,
		NameColName: nameColName,
		ExcelPath:   excelPath,
	})

	if entry.Err != nil {
		result.Ok = false
		result.Reason = entry.Err.Error()
		return result
	}

	// 新文件不触发修改行通知
	if entry.IsNewFile {
		return result
	}

	if entry.DiffResult == nil || len(entry.DiffResult.ModifiedRows) == 0 {
		return result
	}

	idColName := "Id"
	if len(idColNames) > 0 {
		idColName = idColNames[0]
	}
	ctx := &formatNotifyContext{
		sheetName:   param.SheetName,
		idColName:   idColName,
		nameColName: nameColName,
		commitTime:  entry.GitCtx.CommitTime,
		committer:   entry.GitCtx.Committer,
		baseCommit:  entry.GitCtx.BaseCommit,
		headCommit:  entry.GitCtx.HeadCommit,
	}

	// 按 modifiedRow 构建变更记录并格式化
	records := buildFieldChangeRecords(entry.DiffResult.ModifiedRows)
	var notification strings.Builder
	notification.WriteString(formatFieldChangeReason(records, ctx))
	result.Reason = notification.String()

	// 构建 ErrCells（每个字段变更一条）
	for _, row := range entry.DiffResult.ModifiedRows {
		for _, change := range row.Changes {
			result.ErrCells = append(result.ErrCells, &json_rule.CellError{
				Index:    row.RowIndex,
				ExcelRow: row.RowIndex + 1,
				Reason: fmt.Sprintf("%s行，id为%s，字段%s从%s改成了%s",
					row.RowName, row.RowId, change.ColName, change.OldValue, change.NewValue),
				Detail: &json_rule.FieldChangeDetail{
					RowId:    row.RowId,
					RowName:  row.RowName,
					ColName:  change.ColName,
					OldValue: change.OldValue,
					NewValue: change.NewValue,
				},
			})
		}
	}

	return result
}

// buildFieldChangeRecords 将 ModifiedRows 转换为 fieldChangeRecord 列表
// 供 formatFieldChangeReason 格式化使用
func buildFieldChangeRecords(modifiedRows []*diff.RowChange) []*fieldChangeRecord {
	records := make([]*fieldChangeRecord, 0, len(modifiedRows))
	for _, row := range modifiedRows {
		record := &fieldChangeRecord{
			rowId:   row.RowId,
			rowName: row.RowName,
			lineNo:  row.RowIndex + 1, // 转为 Excel 行号（1-based 显示）
		}
		for _, change := range row.Changes {
			record.changes = append(record.changes, struct {
				colName  string
				oldValue string
				newValue string
			}{change.ColName, change.OldValue, change.NewValue})
		}
		records = append(records, record)
	}
	return records
}
