// Package json_rule 提供 Excel 配表校验规则的类型定义
// 本包定义了所有表级和列级规则的数据结构、枚举类型和元数据
package json_rule

import (
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ================== 表级规则相关定义 ==================

// CheckParam 表级检查器参数结构
// 封装表级检查所需的所有参数，避免参数列表过长
type CheckParam struct {
	SheetName   string                    // 表名
	Cols        [][]string                // 所有列数据
	StartRowIdx int                       // 数据起始行索引
	EndIndex    int                       // 数据结束行索引（不含），由调用方通过 AutoDetectEndIndex 计算
	Params      map[string]string         // 规则参数
	SheetMap    map[string]*excelize.File // 其他表的数据（用于跨表检查）
	Now         time.Time                 // 注入当前时间（零值使用 time.Now()）。仅用于单元测试注入固定时间，生产环境请勿设置
}

// ETableRule 表级规则类型枚举
type ETableRule string

const (
	ARENA_SEASON_CHECK                      ETableRule = "ARENA_SEASON_CHECK"                      // 竞技场赛季检查
	NEW_ROW_NOTIFY                          ETableRule = "NEW_ROW_NOTIFY"                          // Deprecated: 使用 ADDED_ROW_NOTIFY + REMOVED_ROW_NOTIFY + ADDED_COL_NOTIFY + REMOVED_COL_NOTIFY
	ROW_CHANGE_NOTIFY                       ETableRule = "ROW_CHANGE_NOTIFY"                       // Deprecated: 使用 MODIFIED_ROW_NOTIFY
	SEASON_PASS_HERO_OPEN_CHECK             ETableRule = "SEASON_PASS_HERO_OPEN_CHECK"             // 战令武将开放时间检查
	ARENA_GENERAL_HERO_OPEN_CHECK           ETableRule = "ARENA_GENERAL_HERO_OPEN_CHECK"           // 大将军武将开放时间检查
	HERO_DROP_CHECK                         ETableRule = "HERO_DROP_CHECK"                         // 武将抽卡掉落检查
	HERO_SYNTHESIS_CHECK                    ETableRule = "HERO_SYNTHESIS_CHECK"                    // 武将合成检查
	HERO_MELT_CHECK                         ETableRule = "HERO_MELT_CHECK"                         // 武将熔炼检查
	ADDED_ROW_NOTIFY                        ETableRule = "ADDED_ROW_NOTIFY"                        // 新增行通知（拆分自 NEW_ROW_NOTIFY）
	REMOVED_ROW_NOTIFY                      ETableRule = "REMOVED_ROW_NOTIFY"                      // 删除行通知（拆分自 NEW_ROW_NOTIFY）
	ADDED_COL_NOTIFY                        ETableRule = "ADDED_COL_NOTIFY"                        // 新增列通知（拆分自 NEW_ROW_NOTIFY）
	REMOVED_COL_NOTIFY                      ETableRule = "REMOVED_COL_NOTIFY"                      // 删除列通知（拆分自 NEW_ROW_NOTIFY）
	MODIFIED_ROW_NOTIFY                     ETableRule = "MODIFIED_ROW_NOTIFY"                     // 修改行通知（原 ROW_CHANGE_NOTIFY）
	ACTIVITY_TASK_REWARD_CHECK              ETableRule = "ACTIVITY_TASK_REWARD_CHECK"              // 活动任务奖励检查
	DROP_ITEM_MUST_IN_ITEM_CHECK            ETableRule = "DROP_ITEM_MUST_IN_ITEM_CHECK"            // 掉落道具必须在道具表中
	DRAWSKIN_BYPRODUCT_CHECK                ETableRule = "DRAWSKIN_BYPRODUCT_CHECK"                // 皮肤抽奖副产品检查
	DANQINGGE_TIME_ACTIVE_CHECK             ETableRule = "DANQINGGE_TIME_ACTIVE_CHECK"             // 丹青阁活动时间校验
	DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK  ETableRule = "DANQINGGE_CUSTOM_PARAM_IS_ITEMID_CHECK"  // 丹青阁活动自定义参数检查
	HERO_SKILL_BUFF_CHECK                   ETableRule = "HERO_SKILL_BUFF_CHECK"                   // 武将技能和Buff引用完整性检查
	DRAW_DROP_RULE_REFERENCE_CHECK          ETableRule = "DRAW_DROP_RULE_REFERENCE_CHECK"          // 抽奖池掉落规则引用检查
	DROP_RULE_GROUP_ID_CHECK                ETableRule = "DROP_RULE_GROUP_ID_CHECK"                // 掉落规则组ID引用检查
	ACTIVITY_DANQINGGE_TIME_CONFIG_CHECK    ETableRule = "ACTIVITY_DANQINGGE_TIME_CONFIG_CHECK"    // 丹青阁活动时间配置检查(ACT-04/05/06)
	DRAWSKIN_ONCE_ITEM_COST_CHECK           ETableRule = "DRAWSKIN_ONCE_ITEM_COST_CHECK"           // DrawSkin单抽消耗检查(DSK-06/07/08)
	DRAWSKIN_DATA_VALIDITY_CHECK            ETableRule = "DRAWSKIN_DATA_VALIDITY_CHECK"            // DrawSkin基础数据检查(DSK-01/12)
	DRAWSKIN_TEN_COST_CHECK                 ETableRule = "DRAWSKIN_TEN_COST_CHECK"                 // DrawSkin十连消耗检查(DSK-09)
	DRAWSKIN_TIME_RANGE_CHECK               ETableRule = "DRAWSKIN_TIME_RANGE_CHECK"               // DrawSkin时间范围检查(DSK-10/11)
	ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK ETableRule = "ACTIVITY_DRAWSKIN_CROSS_REFERENCE_CHECK" // Activity与DrawSkin交叉引用检查(DSK-05/XC-01/XC-04)
	ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK    ETableRule = "ACTIVITY_DRAWSKIN_TIME_OVERLAP_CHECK"    // Activity与DrawSkin时间交集检查(XC-02)
	DROP_RULE_DATA_VALIDITY_CHECK           ETableRule = "DROP_RULE_DATA_VALIDITY_CHECK"           // DropRule基础数据检查
	DROP_RULE_CONDITIONAL_CHECK             ETableRule = "DROP_RULE_CONDITIONAL_CHECK"             // DropRule条件引用检查
	DROP_ITEM_VALIDITY_CHECK                ETableRule = "DROP_ITEM_VALIDITY_CHECK"                // DropItem条件和互斥检查
	DATE_VALID_EXPIRE_CHECK                 ETableRule = "DATE_VALID_EXPIRE_CHECK"                 // 日期有效性检查(ValidDate <= ExpireDate)
	COL_CONTINUOUS_CHECK                    ETableRule = "COL_CONTINUOUS_CHECK"                    // 列连续性检查
	DRAWFIX_PROTECTION_CHECK                ETableRule = "DRAWFIX_PROTECTION_CHECK"                // 定向招募战令保护期检查
	DRAWFIX_ARENA_PROTECTION_CHECK          ETableRule = "DRAWFIX_ARENA_PROTECTION_CHECK"          // 定向招募大将军保护期检查
	HERO_ISOPEN_OPENDATE_CHECK              ETableRule = "HERO_ISOPEN_OPENDATE_CHECK"              // 武将IsOpen与OpenDate一致性检查
	HERO_DROP_VALIDDATE_CHECK               ETableRule = "HERO_DROP_VALIDDATE_CHECK"               // 普通武将掉落时间检查
	SEASON_PASS_REWARD_INTEGRITY_CHECK      ETableRule = "SEASON_PASS_REWARD_INTEGRITY_CHECK"      // 战令高级奖励道具引用完整性检查
)

// TableRule 表级校验规则
type TableRule struct {
	Type        ETableRule        `json:"type"`        // 规则类型（机器识别）
	DisplayName string            `json:"displayName"` // 显示名称（人类可读，如"竞技场赛季不足"）
	Uuid        string            `json:"uuid"`        // 唯一标识
	Description string            `json:"description"` // 规则描述
	Params      map[string]string `json:"params"`      // 规则参数
	Enabled     bool              `json:"enabled"`     // 是否启用
}

// TableCheckResult 表级校验结果
type TableCheckResult struct {
	TableName   *string      `json:"tableName"`
	SheetName   *string      `json:"sheetName"`
	RuleType    ETableRule   `json:"ruleType"`
	DisplayName string       `json:"displayName"`
	Ok          bool         `json:"ok"`
	Reason      string       `json:"reason"`
	ErrCells    []*CellError `json:"errCells"`
}

// SelectOption 下拉选项
type SelectOption struct {
	Label string `json:"label"` // 显示文本
	Value string `json:"value"` // 实际值
}

// TableRuleParamDef 表级规则参数定义
type TableRuleParamDef struct {
	Key         ERuleParam     `json:"key"`         // 参数键
	Label       string         `json:"label"`       // UI 显示标签
	Description string         `json:"description"` // 参数说明
	Type        string         `json:"type"`        // 参数类型: string, duration, number, select
	Default     string         `json:"default"`     // 默认值
	Required    bool           `json:"required"`    // 是否必填
	Options     []SelectOption `json:"options"`     // 下拉选项（type="select" 时使用）
}

// TableRuleMeta 表级规则元数据（用于 MCP 和前端）
type TableRuleMeta struct {
	Type           ETableRule          `json:"type"`
	DisplayName    string              `json:"displayName"`
	Description    string              `json:"description"`
	TargetSheets   []string            `json:"targetSheets"`   // 适用的表名列表（支持后缀匹配，如 "ArenaSeason" 匹配 "竞技场赛季|ArenaSeason"）
	RequiredSheets []string            `json:"requiredSheets"` // 规则执行所需的关联表名列表（如 HERO_DROP_CHECK 需要 "DropItem"、"Item"）；为空表示仅依赖 TargetSheets
	ParamDefs      []TableRuleParamDef `json:"paramDefs"`
}

// ResolveParams 从 ParamDefs 构建默认参数 map，并合并用户覆盖值
// 统一替代各调用方重复的 for range ParamDefs 初始化逻辑
func (m *TableRuleMeta) ResolveParams(overrides map[string]string) map[string]string {
	params := make(map[string]string)
	for _, def := range m.ParamDefs {
		if def.Default != "" {
			params[string(def.Key)] = def.Default
		}
	}
	for k, v := range overrides {
		params[k] = v
	}
	return params
}

// TableRuleMetas 已注册的规则元数据（在 init 中填充）
var TableRuleMetas = map[ETableRule]*TableRuleMeta{}

// GetAllTableRuleMetas 获取所有表级规则元数据
func GetAllTableRuleMetas() []*TableRuleMeta {
	metas := make([]*TableRuleMeta, 0, len(TableRuleMetas))
	for _, meta := range TableRuleMetas {
		metas = append(metas, meta)
	}
	return metas
}

// GetTableRuleMetasForSheet 获取指定表适用的表级规则元数据
// 根据 TargetSheets 字段过滤，只返回适用于指定 Sheet 的规则
func GetTableRuleMetasForSheet(sheetName string) []*TableRuleMeta {
	allMetas := GetAllTableRuleMetas()
	result := make([]*TableRuleMeta, 0, len(allMetas))

	for _, meta := range allMetas {
		if isRuleApplicableToSheet(sheetName, meta.TargetSheets) {
			result = append(result, meta)
		}
	}

	return result
}

// isRuleApplicableToSheet 检查规则是否适用于指定的表
// 规则匹配逻辑：
// 1. 如果 TargetSheets 为空，适用于所有表
// 2. 如果 TargetSheets 包含精确匹配的表名，适用
// 3. 如果 TargetSheets 包含后缀匹配（如 "ArenaSeason" 匹配 "竞技场赛季|ArenaSeason"），适用
func isRuleApplicableToSheet(sheetName string, targetSheets []string) bool {
	// 没有指定目标表，适用于所有表
	if len(targetSheets) == 0 {
		return true
	}

	for _, target := range targetSheets {
		// 精确匹配
		if sheetName == target {
			return true
		}
		// 后缀匹配（配表名格式如 "中文名|英文名"）
		if strings.HasSuffix(sheetName, "|"+target) {
			return true
		}
	}

	return false
}

// ================== 列级规则相关定义 ==================

type ColCheckResult struct {
	// 定位
	TableName *string
	SheetName *string
	ColName   *string
	ColIndex  *int // 定位 第几列
	ErrCells  []*CellError

	// 结果
	Ok     bool
	Reason string
}

// CellError 单元格错误/变更信息
type CellError struct {
	Index    int         // 定位 第几行（列级规则为相对偏移，表级规则为绝对索引）
	Reason   string      // 错误日志
	Detail   interface{} // 变更详情（RowChangeDetail 或 ColumnChangeDetail）
	ExcelRow int         // 实际 Excel 行号（1-indexed），0 表示未设置，展示层应 fallback 到 Index+1
}

// RowChangeDetail 行变更详情
type RowChangeDetail struct {
	ChangeType string // 变更类型: added, removed
	RowId      string // 行ID
	RowName    string // 行名称
}

// ColumnChangeDetail 列变更详情
type ColumnChangeDetail struct {
	ChangeType string // 变更类型: added, removed
	ColName    string // 列名
}

// FieldChangeDetail 字段变更详情（行内字段值变化）
type FieldChangeDetail struct {
	RowId    string // 行ID
	RowName  string // 行名称
	ColName  string // 变更的列名
	OldValue string // 原值
	NewValue string // 新值
}

type ManagerList struct {
	QA         []*GiantLabel // 负责QA
	Designer   []*GiantLabel // 负责策划
	Programmer []*GiantLabel // 负责程序
}

type GiantLabel struct {
	Name string
}

// SheetRule 单个Sheet的完整规则配置
// ⚠️ 以下字段名与 cases/excel_cases/*.json 配置文件强耦合，修改需同步更新所有JSON配置
// ⚠️ 历史问题参考: commit 2ee0e9c (字段名不一致导致功能失效)
type SheetRule struct {
	Sheet       string                   `json:"sheet"`       // @json-critical 配表名，如 "角色成就|HeroAchieve"
	ManagerList *ManagerList             `json:"managerList"` // @json-critical 管理人员配置
	Rules       map[string]*SheetColRule `json:"rules"`       // @json-critical 列级检查规则，key为列名
	TableRules  []*TableRule             `json:"tableRules"`  // @json-critical 表级规则列表
}

type SheetColRule struct {
	PropName  string     `json:"PropName"`  // @json-critical 列属性名
	PropType  string     `json:"PropType"`  // @json-critical 列数据类型，如 "int", "string", "EHeroAchieve"
	PropRules []*ColRule `json:"PropRules"` // @json-critical 该列应用的检查规则列表
}

type ColRule struct {
	Type        EColRule `json:"Type"`        // @json-critical 规则类型，如 "NUMERIC_RANGE", "ALL_BASE", "ENUM"
	Uuid        string   `json:"Uuid"`        // @json-critical 规则唯一标识
	DisplayName string   `json:"DisplayName"` // @json-critical 规则显示名称
	Enabled     bool     `json:"Enabled"`     // @json-critical 是否启用

	//Params map[ERuleParam]string 这个wails3的ts自动绑定生成为 {[_:ERuleParam]:string} 有问题, github有issue 待解决
	// ⚠️ Params 的 key 必须与 ERuleParam 枚举值一致，否则规则参数读取会失败
	// ⚠️ 历史问题: NUMERIC_RANGE 规则曾使用 "min"/"max" 而非 "minValue"/"maxValue" 导致检查失效
	Params map[string]string `json:"Params"` // @json-critical 规则参数，key对应ERuleParam枚举值
}

type EColRule string

//go:generate enumer -type=EColRule -json -text -sql
const (
	// ---------基础规则----------

	TEST             EColRule = "TEST"             // 测试
	INCREASE_ID      EColRule = "INCREASE_ID"      // 从1依次自增的ID
	UNIQUE           EColRule = "UNIQUE"           // 唯一不重复
	CHS_ONLY         EColRule = "CHS_ONLY"         // 仅中文
	NOT_EMPTY        EColRule = "NOT_EMPTY"        // 不为空
	SERVER_OR_CLIENT EColRule = "SERVER_OR_CLIENT" // 前端还是后端
	ALL_BASE         EColRule = "ALL_BASE"         // 以上所有基础规则合并

	// ---------基础规则----------

	// ---------数据类型规则----------

	NUMERIC EColRule = "NUMERIC" // 数值类型
	DATE    EColRule = "DATE"    // 检测日期类型, 日期格式、日期范围等
	BOOLEAN EColRule = "BOOLEAN" // 布尔值类型
	STRING  EColRule = "STRING"  // 单元格应为字符串
	// 通过 excelize.GetCellType() 检测实际存储格式：
	// - CellTypeSharedString(7)/CellTypeInlineString(5): 文本格式（正确）
	// - CellTypeUnset(0)/CellTypeNumber(6): 数值格式（错误）

	// ---------数据类型规则----------

	// ---------业务关系规则----------

	// ---日期---

	DATE_RANGE    EColRule = "DATE_RANGE"    // 日期范围
	DATE_DURATION EColRule = "DATE_DURATION" // 日期跨度

	// ---日期---
	// ---数值---

	NUMERIC_RANGE EColRule = "NUMERIC_RANGE" // 范围内整数
	ENUM          EColRule = "ENUM"          // 固定枚举

	// ---数值---
	// ---关联表---

	FOREIGN_KEY     EColRule = "FOREIGN_KEY"     // 关联表
	CROSS_REFERENCE EColRule = "CROSS_REFERENCE" // 跨表引用检查
	SPLIT_REGERENCE EColRule = "SPLIT_REGERENCE" // 拆分引用检查
	CHAIN_REFERENCE EColRule = "CHAIN_REFERENCE" // 跨表关系链检查

	// ---关联表---
	// ---特殊格式---

	SPECIAL_FORMAT EColRule = "SPECIAL_FORMAT" // 特殊格式检查 - 道具id+数量：{1000005;1}{1000011;10}{9000001;5}
	REGEX          EColRule = "REGEX"          // 自定义正则检查 - 道具id+数量：{1000005;1}{1000011;10}{9000001;5}

	// ---特殊格式---

	// ---数值计算---

	WEIGHT_SUM       EColRule = "WEIGHT_SUM"
	DATE_CONSISTENCY EColRule = "DATE_CONSISTENCY"

	// ---数值计算---

	// ---资源检查---

	RESOURCE EColRule = "RESOURCE"

	// ---资源检查---

	// ---拼音和汉字匹配---

	PIN_YIN_CHS EColRule = "PIN_YIN_CHS"

	// ---拼音和汉字匹配---

	// ---富文本检查---

	RICH_TEXT EColRule = "RICH_TEXT"

	// ---富文本检查---

	// ---------业务关系规则----------
)

type ERuleParam string

const (
	NONE ERuleParam = "none"

	// ---一般传参---
	ALLOW_EMPTY        ERuleParam = "allowEmpty"     // 允许空值
	EXCEPTS            ERuleParam = "excepts"        // 排除的值 逗号间隔
	ALLOW_COMMIT       ERuleParam = "allowCommit"    // 允许注释值 正则，比如 ^#.*
	ENUMS              ERuleParam = "enums"          // 枚举值 逗号间隔
	BREAK_LINE         ERuleParam = "breakLine"      // 列截断检测连续空单元格格数
	USE_ID_COL_FOR_END ERuleParam = "useIdColForEnd" // 使用ID列确定数据结束位置（避免不同列截断位置不一致）
	STRICT             ERuleParam = "strict"         // 严格检查
	COMPARE_RULE       ERuleParam = "compareRule"    // 比较规则 eq（等于）、gt（大于）、lt（小于）、ge（大于等于）、le（小于等于）
	TOLERANCE          ERuleParam = "tolerance"      // 容差
	MIN                ERuleParam = "min"            // 最小值
	MAX                ERuleParam = "max"            // 最大值
	PATTERN            ERuleParam = "pattern"        // 正则表达式
	GROUPS             ERuleParam = "groups"         // 正则表达式捕获组
	FILTER_COL         ERuleParam = "filterCol"      // 条件过滤列名（仅当该行指定列的值等于filterVal时才执行检查）
	FILTER_VAL         ERuleParam = "filterVal"      // 条件过滤值
	FILTER_IS_ARRAY    ERuleParam = "filterIsArray"  // 过滤值是否数组（"true" 时按逗号拆分多值匹配）
	FILTER_MODE        ERuleParam = "filterMode"     // 过滤模式: ""(单值) | "multi"(多值) | "withinDays"(距今<N天)
	FILTER_DAYS        ERuleParam = "filterDays"     // 距今天数（仅 filterMode="withinDays" 时使用）
	// ---一般传参---

	START_DATE ERuleParam = "startDate" // 开始日期
	END_DATE   ERuleParam = "endDate"   // 结束日期

	// ---表级规则参数---
	SEASON_END_TIME_COL ERuleParam = "seasonEndTimeCol" // SeasonEndTime 列名（默认: SeasonEndTime）
	TIME_RANGE_BEFORE   ERuleParam = "timeRangeBefore"  // 提前警告时间（默认: 168h 即7天）

	// ---通用通知规则参数---
	ID_COL_NAME         ERuleParam = "idColName"         // ID列名（默认: Id）
	ID_COL_NAMES        ERuleParam = "idColNames"        // 复合主键列名（逗号分隔，如 "AnimationState,ItemCfgId"），优先级高于 ID_COL_NAME
	NAME_COL_NAME       ERuleParam = "nameColName"       // 名称列名（默认: Name）
	OLD_DATA_PATH       ERuleParam = "oldDataPath"       // 历史数据JSON路径
	NOTIFY_ADDED_ROWS   ERuleParam = "notifyAddedRows"   // 是否通知新增行（默认: true）
	NOTIFY_REMOVED_ROWS ERuleParam = "notifyRemovedRows" // 是否通知删除行（默认: true）
	NOTIFY_ADDED_COLS   ERuleParam = "notifyAddedCols"   // 是否通知新增列（默认: true）
	NOTIFY_REMOVED_COLS ERuleParam = "notifyRemovedCols" // 是否通知删除列（默认: true）
	// ---Git diff 模式参数---
	USE_GIT_DIFF  ERuleParam = "useGitDiff"  // 使用 git diff 替代快照（默认: true）
	GIT_REPO_PATH ERuleParam = "gitRepoPath" // Git 仓库路径（默认: Excel 文件所在目录）
	BASE_COMMIT   ERuleParam = "baseCommit"  // 基准 commit（默认: HEAD~1）
	HEAD_COMMIT   ERuleParam = "headCommit"  // 当前检查的 commit hash（merge 场景为子 commit，普通场景为 HEAD）
	// ---通用通知规则参数---

	// ---列连续性检查规则参数---
	TARGET_COL   ERuleParam = "targetCol"   // 检查列名
	CHECK_MODE   ERuleParam = "checkMode"   // 检查模式
	SCOPE        ERuleParam = "scope"       // 检查范围
	START_VALUE  ERuleParam = "startValue"  // 起始值
	EXCLUDE_ROWS ERuleParam = "excludeRows" // 排除行号
	SEPARATOR    ERuleParam = "separator"   // 拆分分隔符（SPLIT_UNIQUE 模式）
	// ---列连续性检查规则参数---

	// ---武将检查规则参数---
	WARN_DAYS_BEFORE   ERuleParam = "warnDaysBefore"  // 提前警告天数（默认: 5）
	DROP_MONTHS_DELAY  ERuleParam = "dropMonthsDelay" // 掉落延迟月数（默认: 3）
	SKILL_COL_NAME     ERuleParam = "skillColName"    // 技能列名（默认: Skill）
	CAN_MELT_COL_NAME  ERuleParam = "canMeltColName"  // 可熔炼列名（默认: CanMelt）
	IS_OPEN_COL_NAME   ERuleParam = "isOpenColName"   // 是否开放列名（默认: IsOpen）
	OPEN_DATE_COL_NAME ERuleParam = "openDateColName" // 开放时间列名（默认: OpenDate）
	PROTECT_MONTHS     ERuleParam = "protectMonths"   // 战令保护期月数（默认: 4）
	// ---武将检查规则参数---

	// ---关系链检查规则参数---
	CHAIN_STEPS           ERuleParam = "chainSteps"          // 关系链步骤配置（JSON字符串）
	CHAIN_COMPARE         ERuleParam = "chainCompare"        // 比较类型：match/not_match/time_overlap/exists/not_exists
	CHAIN_MATCH_COMPARE   ERuleParam = "chainMatchCompare"   // 匹配阶段类型（两阶段门控模型的匹配规则）
	CHAIN_LEFT_KEY        ERuleParam = "chainLeftKey"        // 来源链最终字段（用于时间比较）
	CHAIN_RIGHT_KEY       ERuleParam = "chainRightKey"       // 目标链最终字段（用于时间比较）
	CHAIN_REQUIRED_SHEETS ERuleParam = "chainRequiredSheets" // 关系链所需的关联表（逗号分隔），由前端从 chainSteps 中提取目标表名后填充
	CHAIN_WARN_BEFORE     ERuleParam = "chainWarnBefore"     // 预警窗口提前量（Go duration，如"168h"）；"0"表示无条件启用
	CHAIN_WARN_SHEET      ERuleParam = "chainWarnSheet"      // 预警时间所在表名
	CHAIN_WARN_COL        ERuleParam = "chainWarnCol"        // 预警时间所在列名
	// ---关系链检查规则参数---
	// ---表级规则参数---
)

// ReferenceSheetParamKeys 列级规则中引用其他表的参数名
// FOREIGN_KEY 用 targetSheet, CROSS_REFERENCE/SPLIT_REFERENCE 用 refSheet, DATE_CONSISTENCY 用 descSheet,
// CHAIN_REFERENCE 用 chainRequiredSheets（目标表名嵌套在 chainSteps JSON 内部，需由前端提取后单独传递）
var ReferenceSheetParamKeys = []string{"targetSheet", "refSheet", "descSheet", "chainRequiredSheets"}
