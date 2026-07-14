// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件定义洋葱模型的接口和数据上下文
// 洋葱模型将关系链检查拆分为多个独立的处理器层，每层通过 NextFunc 委托给内层，
// 形成从外到内再从内到外的执行流：
//
//	Validate → LeftStep0 → ... → LeftStepM → Match → RightStepN → ... → RightStep0 → Compare
//	   ↑                                                                              ↓
//	   └──────────────────── 结果回传（Violation/Reason）←─────────────────────────────┘
package chain_reference

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// ==================== 洋葱模型接口 ====================

// ChainHandler 洋葱模型的单层处理器
// Handle 执行当前层逻辑，通过调用 next 委托给内层
type ChainHandler interface {
	Handle(ctx *ChainContext, next NextFunc) error
}

// NextFunc 内层执行函数
// 洋葱模型的核心：每个 handler 在执行完自身逻辑后调用 next，
// 控制权逐层向内传递，内层返回后再逐层向外回传
type NextFunc func(ctx *ChainContext) error

// ==================== 洋葱模型数据上下文 ====================
// ChainContext 洋葱模型的数据上下文
// 承载从外层到内层全部中间状态，各层 handler 通过读写此结构传递数据
//
// 数据流向（洋葱模型的"切片"视角）：
//
//	外层（Validate）→ 填充输入参数 → 传递给内层
//	左链步骤 → 累积 LeftCurrentValues/LeftStepValues → 传递给内层
//	Match → 读取左链结果，写入 Matched/MatchedKeys → 传递给内层
//	右链步骤 → 累积 RightCurrentValues/RightStepValues → 传递给内层
//	Compare → 读取左右结果，写入 Violation/Reason → 返回外层
type ChainContext struct {
	// ---- 输入参数（由 Validate 层填充，后续层只读） ----
	// Cols 当前表的列数据（按列存储的二维数组）
	Cols [][]string
	// ColIdx 当前检查列索引
	ColIdx int
	// RowIdx 当前行索引（绝对行号）
	RowIdx int
	// StartRowIdx 数据起始行索引
	StartRowIdx int
	// SheetMap 所有表的数据映射（key 为 Sheet 名称，value 为 excelize.File）
	SheetMap map[string]*excelize.File
	// Config 两链配置（left + right 各包含 steps 和 compareCol）
	Config *ChainPairConfig
	// Params 规则参数（chainCompare、chainMatchCompare 等）
	Params map[string]string
	// MyColData 当前列数据（从 startRowIdx 到 endIdx 的切片）
	MyColData []string
	// DataIdx 当前行在 myColData 中的索引（相对行号）
	DataIdx int
	// ---- 左链中间结果（前向累积） ----
	// LeftCurrentValues 左链当前步提取的值集合，作为下一步的输入
	LeftCurrentValues []string
	// LeftStepValues 左链每一步的 nextValues（StepValues[0]=第一步, ..., StepValues[last]=最后一步）
	LeftStepValues [][]string
	// LeftResult 左链最终执行结果（包含 MatchValues、CompareValues 等）
	LeftResult *ChainResult
	// ---- 匹配结果 ----
	// Matched 两链是否交汇（Phase 1 门控结果）
	Matched bool
	// MatchedKeys 匹配阶段找到的交集键列表（用于调试和原因描述）
	MatchedKeys []string
	// ---- 右链中间结果（反向累积） ----
	// RightCurrentValues 右链当前步提取的值集合，作为下一步的输入
	RightCurrentValues []string
	// RightStepValues 右链每一步的 nextValues
	RightStepValues [][]string
	// FirstStepInputValues 右链第一步的查找值（用于两阶段比较的 Phase 2 右侧值）
	FirstStepInputValues []string
	// RightResult 右链最终执行结果
	RightResult *ChainResult
	// ---- 比较结果 ----
	// Violation 是否违规（true = 检查不通过，需要报错）
	Violation bool
	// Reason 违规原因描述（Violation 为 true 时填充）
	Reason string
	// ---- 错误 ----
	// Err 执行过程中的错误（与 Violation 不同：Err 表示程序异常，Violation 表示业务违规）
	Err error

	// ---- 预警窗口参数 ----
	// WarnValues 预警时间值集合（由左链/右链步骤在匹配到 chainWarnSheet 时提取）
	// 用于 ShouldSuppressByWarnBefore 判定是否因时间过早而静默违规
	WarnValues []string
	// ColsCache GetCols 缓存，避免同一张表在每行重复调用 GetCols
	ColsCache map[string][][]string
}

// key 为 "sheetName"，value 为 GetCols 返回的列数据
// NewChainContext 创建新的 ChainContext
// 所有输入参数在此处设置，中间结果由各层 handler 在执行过程中填充
//
// 参数：
//   - cols: 当前表的列数据
//   - colIdx: 当前检查列索引
//   - rowIdx: 当前行索引（绝对行号）
//   - startRowIdx: 数据起始行索引
//   - sheetMap: 所有表的数据映射
//   - config: 两链配置
//   - params: 规则参数
//   - myColData: 当前列数据
//   - dataIdx: 当前行在 myColData 中的索引
func NewChainContext(
	cols [][]string,
	colsCache map[string][][]string,
	colIdx int,
	rowIdx int,
	startRowIdx int,
	sheetMap map[string]*excelize.File,
	config *ChainPairConfig,
	params map[string]string,
	myColData []string,
	dataIdx int,
) *ChainContext {
	return &ChainContext{
		Cols:        cols,
		ColsCache:   colsCache,
		ColIdx:      colIdx,
		RowIdx:      rowIdx,
		StartRowIdx: startRowIdx,
		SheetMap:    sheetMap,
		Config:      config,
		Params:      params,
		MyColData:   myColData,
		DataIdx:     dataIdx,
	}
}

// GetCachedCols 获取目标表的列数据，使用缓存避免重复调用 GetCols
// 所有洋葱模型层应使用此方法而非直接调用 targetFile.GetCols()
func (ctx *ChainContext) GetCachedCols(sheetName string, file *excelize.File) ([][]string, error) {
	if ctx.ColsCache != nil {
		if cached, ok := ctx.ColsCache[sheetName]; ok {
			return cached, nil
		}
	}
	cols, err := file.GetCols(sheetName)
	if err != nil {
		return nil, err
	}
	if ctx.ColsCache != nil {
		ctx.ColsCache[sheetName] = cols
	}
	return cols, nil
}

// CurrentCellValue 获取当前行当前列的单元格值
// 等价于 MyColData[DataIdx]，提供语义化的访问方式
func (ctx *ChainContext) CurrentCellValue() string {
	if ctx.DataIdx >= 0 && ctx.DataIdx < len(ctx.MyColData) {
		return ctx.MyColData[ctx.DataIdx]
	}
	return ""
}

// CompareType 获取比较类型参数（chainCompare）
// 默认值为 "verify_exists"
func (ctx *ChainContext) CompareType() string {
	if ct, ok := ctx.Params["chainCompare"]; ok && ct != "" {
		return ct
	}
	return "verify_exists"
}

// MatchCompareType 获取匹配阶段类型参数（chainMatchCompare）
// 为空时表示不使用两阶段门控，退化为单阶段比较
func (ctx *ChainContext) MatchCompareType() string {
	if mct, ok := ctx.Params["chainMatchCompare"]; ok && mct != "" {
		return mct
	}
	return ""
}

// ChainStepsJSON 获取链配置的原始 JSON 字符串
func (ctx *ChainContext) ChainStepsJSON() string {
	return ctx.Params["chainSteps"]
}

// WarnBefore 获取预警窗口提前量（Go duration 格式，如 "168h"）
// 返回 0 表示未配置或解析失败，不启用预警过滤
func (ctx *ChainContext) WarnBefore() time.Duration {
	val := ctx.Params["chainWarnBefore"]
	if val == "" {
		return 0
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0
	}
	return d
}

// WarnSheet 获取预警时间所在的表名（如 "赛季战令表|SeasonPass"）
// 为空表示不启用预警过滤
func (ctx *ChainContext) WarnSheet() string {
	return ctx.Params["chainWarnSheet"]
}

// WarnCol 获取预警时间所在的列名（如 "StartTime"）
// 为空表示不启用预警过滤
func (ctx *ChainContext) WarnCol() string {
	return ctx.Params["chainWarnCol"]
}

// String 提供 ChainContext 的简要描述，用于日志和调试
func (ctx *ChainContext) String() string {
	return fmt.Sprintf("ChainContext{Row=%d, Col=%d, DataIdx=%d, LeftSteps=%d, RightSteps=%d}",
		ctx.RowIdx, ctx.ColIdx, ctx.DataIdx,
		len(ctx.Config.Left.Steps), len(ctx.Config.Right.Steps),
	)
}
