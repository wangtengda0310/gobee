// Package general 提供通用表级别的校验规则
// 本包中的规则适用于所有表，不依赖具体业务逻辑

package coded_rules

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// RowChangeNotifyRule 行变更字段通知规则
//
// 原始需求：https://ztgame.feishu.cn/wiki/TOJlw8ucyiNkfMkuoyjcJi8XnLc
//
// ## 校验规则（来源：需求确认）
// 1. 检测修改的行：对比历史数据，找出字段值发生变化的行
// 2. 生成变更详情：格式为 "xx行，id为xx，字段y从a改成了b"
// 3. 使用 git diff 获取上一个 commit 的文件内容进行对比
//
// ## 需求确认记录
//   - Q: 哪些表需要进行变更检测？
//     A: 所有表都需要进行变更检测
//   - Q: 变更字段格式？
//     A: "xx行，id为xx，字段y从a改成了b"
type RowChangeNotifyRule struct{}

// Meta 返回规则元数据
func (c *RowChangeNotifyRule) Meta() *json_rule.TableRuleMeta {
	return &json_rule.TableRuleMeta{
		Type:         json_rule.ROW_CHANGE_NOTIFY,
		DisplayName:  "行变更字段通知",
		Description:  "检测Excel表中已修改行的字段变更，生成详细的变更通知。格式：xx行，id为xx，字段y从a改成了b。",
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
		},
	}
}

// Check 执行表级检查（委托模式）
// 直接委托到 ModifiedRowNotifyRule
func (c *RowChangeNotifyRule) Check(param json_rule.CheckParam) *json_rule.TableCheckResult {
	return (&ModifiedRowNotifyRule{}).Check(param)
}
