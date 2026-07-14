// Package general 提供通用表级别的校验规则
// 本包中的规则适用于所有表，不依赖具体业务逻辑

package coded_rules

import (
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// NewRowNotifyRule 新增行/列通知规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// ## 校验规则（来源：需求确认）
// 1. 检测新增行：对比历史数据，找出新增的数据行
// 2. 检测新增列：对比历史数据，找出新增的列
// 3. 检测删除行：对比历史数据，找出被删除的数据行
// 4. 检测删除列：对比历史数据，找出被删除的列
// 5. 使用 git diff 获取上一个 commit 的文件内容进行对比
//
// ## 需求确认记录
//   - Q: 哪些表需要进行变更检测？
//     A: 所有表都需要进行变更检测
//   - Q: 删除行是否需要通知？
//     A: 需要通知
//   - Q: NEW_ROW_NOTIFY如何获取列级规则？
//     A: 通过sheetMap参数传入SheetRule配置
type NewRowNotifyRule struct{}

// Meta 返回规则元数据
func (c *NewRowNotifyRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.NEW_ROW_NOTIFY,
		DisplayName:  "新增行/列通知",
		Description:  "检测Excel表中的新增行、新增列、删除行、删除列，并生成变更通知。",
		TargetSheets: []string{}, // 空数组表示适用于所有表
		ParamDefs: []json_rule.TableRuleParamDef{
			{
				Key:         json_rule.ID_COL_NAME,
				Label:       "ID列名",
				Description: "用于标识行的ID列名",
				Type:        "string",
				Default:     "Id",
				Required:    false,
			},
			{
				Key:         json_rule.ID_COL_NAMES,
				Label:       "复合主键列名",
				Description: "逗号分隔的多列名（优先级高于ID列名），如 'AnimationState,ItemCfgId'",
				Type:        "string",
				Default:     "",
				Required:    false,
			},
			{
				Key:         json_rule.NAME_COL_NAME,
				Label:       "名称列名",
				Description: "用于标识行的名称列名",
				Type:        "string",
				Default:     "Name",
				Required:    false,
			},
			{
				Key:         json_rule.GIT_REPO_PATH,
				Label:       "Git仓库路径",
				Description: "Git仓库根目录路径",
				Type:        "string",
				Default:     ".",
				Required:    false,
			},
			{
				Key:         json_rule.BASE_COMMIT,
				Label:       "基准Commit",
				Description: "用于对比的基准commit（默认HEAD~1）",
				Type:        "string",
				Default:     "HEAD~1",
				Required:    false,
			},
			{
				Key:         json_rule.NOTIFY_ADDED_ROWS,
				Label:       "通知新增行",
				Description: "是否生成新增行通知",
				Type:        "boolean",
				Default:     "true",
				Required:    false,
			},
			{
				Key:         json_rule.NOTIFY_REMOVED_ROWS,
				Label:       "通知删除行",
				Description: "是否生成删除行通知",
				Type:        "boolean",
				Default:     "true",
				Required:    false,
			},
			{
				Key:         json_rule.NOTIFY_ADDED_COLS,
				Label:       "通知新增列",
				Description: "是否生成新增列通知",
				Type:        "boolean",
				Default:     "true",
				Required:    false,
			},
			{
				Key:         json_rule.NOTIFY_REMOVED_COLS,
				Label:       "通知删除列",
				Description: "是否生成删除列通知",
				Type:        "boolean",
				Default:     "true",
				Required:    false,
			},
		},
	}
}

// Check 执行表级检查（委托模式）
// 根据布尔参数委托到对应的子规则，合并结果
func (c *NewRowNotifyRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	result := &json_rule.TableCheckResult{
		Ok:          true,
		RuleType:    json_rule.NEW_ROW_NOTIFY,
		DisplayName: "新增行/列通知",
		ErrCells:    make([]*json_rule.CellError, 0),
	}

	// 读取布尔参数决定委托到哪些子规则
	notifyAddedRows := param.Params[string(json_rule.NOTIFY_ADDED_ROWS)] != "false"
	notifyRemovedRows := param.Params[string(json_rule.NOTIFY_REMOVED_ROWS)] != "false"
	notifyAddedCols := param.Params[string(json_rule.NOTIFY_ADDED_COLS)] != "false"
	notifyRemovedCols := param.Params[string(json_rule.NOTIFY_REMOVED_COLS)] != "false"

	var reasons []string

	if notifyAddedRows {
		r := (&AddedRowNotifyRule{}).Check(param)
		result.ErrCells = append(result.ErrCells, r.ErrCells...)
		if r.Reason != "" {
			reasons = append(reasons, r.Reason)
		}
		if !r.Ok {
			result.Ok = false
		}
	}
	if notifyRemovedRows {
		r := (&RemovedRowNotifyRule{}).Check(param)
		result.ErrCells = append(result.ErrCells, r.ErrCells...)
		if r.Reason != "" {
			reasons = append(reasons, r.Reason)
		}
		if !r.Ok {
			result.Ok = false
		}
	}
	if notifyAddedCols {
		r := (&AddedColNotifyRule{}).Check(param)
		result.ErrCells = append(result.ErrCells, r.ErrCells...)
		if r.Reason != "" {
			reasons = append(reasons, r.Reason)
		}
		if !r.Ok {
			result.Ok = false
		}
	}
	if notifyRemovedCols {
		r := (&RemovedColNotifyRule{}).Check(param)
		result.ErrCells = append(result.ErrCells, r.ErrCells...)
		if r.Reason != "" {
			reasons = append(reasons, r.Reason)
		}
		if !r.Ok {
			result.Ok = false
		}
	}

	if len(reasons) > 0 {
		result.Reason = strings.Join(reasons, "\n\n---\n\n")
	}

	return result
}
