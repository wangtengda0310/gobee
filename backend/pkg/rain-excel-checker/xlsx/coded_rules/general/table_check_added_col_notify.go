package coded_rules

import (
	"fmt"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// AddedColNotifyRule 新增列通知规则
// 检测 Excel 表中新增的列，生成变更通知
type AddedColNotifyRule struct{}

func (c *AddedColNotifyRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.ADDED_COL_NOTIFY,
		DisplayName:  "新增列通知",
		Description:  "检测Excel表中新增的列，生成变更通知。仅在有新增列时产生通知。",
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

func (c *AddedColNotifyRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		RuleType:    json_rule.ADDED_COL_NOTIFY,
		DisplayName: "新增列通知",
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

	// 新文件不触发列通知
	if entry.IsNewFile {
		return result
	}

	if entry.DiffResult == nil || len(entry.DiffResult.AddedCols) == 0 {
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

	var notification strings.Builder
	notification.WriteString(formatAddedColsReason(entry.DiffResult.AddedCols, ctx))
	result.Reason = notification.String()

	for _, colName := range entry.DiffResult.AddedCols {
		result.ErrCells = append(result.ErrCells, &json_rule.CellError{
			Index:  -1,
			Reason: fmt.Sprintf("新增列: %s", colName),
			Detail: &json_rule.ColumnChangeDetail{
				ChangeType: "colAdded",
				ColName:    colName,
			},
		})
	}

	return result
}
