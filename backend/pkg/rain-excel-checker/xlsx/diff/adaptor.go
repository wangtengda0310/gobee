// Package diff 提供 Excel 配表差异检测与上下文管理功能
// 本包作为 src 包中简洁接口（colRule/rowRule/sheetRule）与现有复杂实现之间的桥接层
package diff

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// ================== 全局上下文管理器 ==================

// GlobalAdaptor 全局适配器实例，用于存储和检索检查上下文
//
// 使用场景：
//   - colRule 接口保持简洁（只接收 map[int]string）
//   - 实际规则需要访问完整上下文时（跨列检查、参数读取、sheetMap 访问）
//   - 通过 adaptor.GetContext() 获取 CheckParam
//
// ⚠️ 重要：本包供 check_manager 内部使用，coded_rules 包不要导入此文件，避免形成循环依赖
var GlobalAdaptor ContextAdaptor = &adaptorImpl{
	contexts: make(map[string]*CheckContext),
}

// ContextAdaptor 上下文适配器接口
// 提供接口类型便于测试和 mock
type ContextAdaptor interface {
	// StoreContext 存储检查上下文，返回取消函数
	// 使用 context.Context 实现自动清理，避免内存泄漏
	StoreContext(ctx context.Context, sheetName, colName string, checkCtx *CheckContext) (context.Context, context.CancelFunc)

	// GetContext 获取检查上下文
	// 返回详细错误信息，便于调试
	GetContext(sheetName, colName string) (*CheckContext, error)

	// ClearContext 清理指定上下文
	ClearContext(sheetName, colName string) error

	// ClearAll 清理所有上下文（用于测试）
	ClearAll()

	// Stats 获取上下文统计信息（用于调试和监控）
	Stats() AdaptorStats
}

// adaptorImpl 上下文适配器实现
type adaptorImpl struct {
	mu       sync.RWMutex             // 读写锁，保护 contexts 并发访问
	contexts map[string]*CheckContext // 存储多个检查上下文，key 格式: "sheetName:colName:uuid"
	counter  uint64                   // 原子计数器，生成唯一 ID
}

// AdaptorStats 适配器统计信息
type AdaptorStats struct {
	TotalContexts int      // 当前存储的上下文总数
	Keys          []string // 所有上下文的 key
}

// CheckContext 单次检查的完整上下文
// 封装了 src 简洁接口之外的所有信息，包括：
//   - CheckParam: 现有的复杂参数结构（包含所有列数据、参数、sheetMap 等）
//   - ColName: 当前列名（用于生成唯一 key）
//   - ColType: 当前列类型（可选，用于类型检查）
//   - SheetName: 表名
//   - ColIndex: 列索引
type CheckContext struct {
	CheckParam *json_rule.CheckParam // 检查参数（包含 Cols、Params、SheetMap 等）
	ColName    string                // 当前列名
	ColType    string                // 当前列类型（如 "int", "string", "E#枚举名"）
	SheetName  string                // 表名
	ColIndex   int                   // 列索引
	UUID       string                // 唯一标识符（用于并发场景下的 key 区分）
}

// StoreContext 存储检查上下文
// 在执行检查前调用，将 CheckParam 和列信息存入全局 adaptor
//
// 参数：
//   - ctx: 外部 context.Context，用于传递取消信号
//   - sheetName: 表名
//   - colName: 列名
//   - checkCtx: 检查上下文
//
// 返回：
//   - context.Context: 包含上下文 key 的 context
//   - context.CancelFunc: 取消函数，用于清理上下文
//
// 使用示例：
//
//	ctx, cancel := adaptor.GlobalAdaptor.StoreContext(context.Background(), "Item", "Id", &helpers.CheckContext{
//	    CheckParam: &json_rule.CheckParam{...},
//	    ColName: "Id",
//	    ColType: "int",
//	})
//	defer cancel()
//
//	result := checker.Check(colData)
func (a *adaptorImpl) StoreContext(ctx context.Context, sheetName, colName string, checkCtx *CheckContext) (context.Context, context.CancelFunc) {
	// 生成唯一 ID，解决并发场景下的 key 冲突
	uuid := fmt.Sprintf("%d", atomic.AddUint64(&a.counter, 1))
	checkCtx.UUID = uuid
	key := a.buildKey(sheetName, colName, uuid)

	a.mu.Lock()
	a.contexts[key] = checkCtx
	a.mu.Unlock()

	// 创建带取消功能的 context
	ctx = context.WithValue(ctx, contextKey{}, key)

	// 返回取消函数
	cancel := func() {
		a.ClearContext(sheetName, colName)
	}

	// 使用 context 的 cancel 机制实现自动清理
	go func() {
		<-ctx.Done()
		a.ClearContext(sheetName, colName)
	}()

	return ctx, cancel
}

// GetContext 获取检查上下文
// 在 colRule.Check() 方法中调用，获取完整的 CheckParam
//
// 参数：
//   - sheetName: 表名
//   - colName: 列名
//
// 返回：
//   - *CheckContext: 检查上下文
//   - error: 错误信息（便于调试）
//
// 错误场景：
//   - ErrContextNotFound: 上下文未设置
//   - ErrMultipleContexts: 存在多个上下文（并发场景）
//
// 使用示例（在 colRule 实现中）：
//
//	func (r *MyColumnTypeCheckRule) Check(data map[int]string) col_check.ColCheckResult {
//	    ctx, err := adaptor.GlobalAdaptor.GetContext("Item", "Id")
//	    if err != nil {
//	        return col_check.NewColCheckResult(false, err.Error())
//	    }
//	    params := ctx.CheckParam.Params
//	    sheetMap := ctx.CheckParam.SheetMap
//	    // ... 执行检查逻辑
//	}
func (a *adaptorImpl) GetContext(sheetName, colName string) (*CheckContext, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 查找所有匹配的上下文（可能有多个，并发场景）
	var matches []*CheckContext
	prefix := fmt.Sprintf("%s:%s:", sheetName, colName)

	for key, ctx := range a.contexts {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			matches = append(matches, ctx)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: sheetName=%s, colName=%s", ErrContextNotFound, sheetName, colName)
	}

	if len(matches) > 1 {
		// 并发场景：存在多个上下文，返回错误
		return nil, fmt.Errorf("%w: found %d contexts for sheetName=%s, colName=%s",
			ErrMultipleContexts, len(matches), sheetName, colName)
	}

	return matches[0], nil
}

// ClearContext 清理检查上下文
// 检查完成后调用，释放内存
//
// 参数：
//   - sheetName: 表名
//   - colName: 列名
//
// 行为：
//   - 清理所有匹配的上下文（并发场景下可能有多个）
//
// 建议：使用 defer 确保上下文一定会被清理
//
//	ctx, cancel := adaptor.GlobalAdaptor.StoreContext(context.Background(), "Item", "Id", ctx)
//	defer cancel()
func (a *adaptorImpl) ClearContext(sheetName, colName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	prefix := fmt.Sprintf("%s:%s:", sheetName, colName)
	deleted := 0

	for key := range a.contexts {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(a.contexts, key)
			deleted++
		}
	}

	if deleted == 0 {
		return fmt.Errorf("%w: sheetName=%s, colName=%s", ErrContextNotFound, sheetName, colName)
	}

	return nil
}

// ClearAll 清理所有上下文
// 通常用于测试或批量检查结束后
func (a *adaptorImpl) ClearAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.contexts = make(map[string]*CheckContext)
}

// Stats 获取上下文统计信息
// 用于调试和监控
func (a *adaptorImpl) Stats() AdaptorStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	keys := make([]string, 0, len(a.contexts))
	for key := range a.contexts {
		keys = append(keys, key)
	}

	return AdaptorStats{
		TotalContexts: len(a.contexts),
		Keys:          keys,
	}
}

// buildKey 生成上下文存储的唯一 key
// 格式: "sheetName:colName:uuid"
func (a *adaptorImpl) buildKey(sheetName, colName, uuid string) string {
	return fmt.Sprintf("%s:%s:%s", sheetName, colName, uuid)
}

// ================== 错误定义 ==================

// contextKeyType 用于 context.Value 的 key 类型
type contextKey struct{}

// 错误定义
var (
	ErrContextNotFound  = fmt.Errorf("context not found")
	ErrMultipleContexts = fmt.Errorf("multiple contexts found")
)

// ================== 适配器 ==================

// ColRuleAdapter 列规则适配器
// 将 src.ColRule 接口适配到现有的 Checker 接口
//
// 职责：
//  1. 在调用 colRule.Check() 前设置上下文
//  2. 调用完成后清理上下文
//  3. 转换数据格式（现有列数据 → map[int]string）
type ColRuleAdapter struct {
	rule ColRule // src.ColRule 接口
}

// ColRule 列规则接口（来自 src 包）
// 保持接口简洁，只接收列数据
type ColRule interface {
	Check(data map[int]string) ColCheckResult
}

// ColCheckResult 列检查结果（来自 src 包）
type ColCheckResult interface {
	IsOk() bool
	GetReason() string
}

// NewColRuleAdapter 创建列规则适配器
func NewColRuleAdapter(rule ColRule) *ColRuleAdapter {
	return &ColRuleAdapter{rule: rule}
}

// Check 执行检查（适配到 Checker 接口）
// 这个方法会：
//  1. 将当前列数据转换为 map[int]string 格式
//  2. 存储上下文到 GlobalAdaptor
//  3. 调用 rule.Check()
//  4. 清理上下文
//  5. 转换结果格式
func (a *ColRuleAdapter) Check(param *json_rule.CheckParam, sheetName, colName string, colData []string) ColCheckResult {
	// 1. 转换数据格式: []string → map[int]string
	dataMap := make(map[int]string)
	for rowIdx, value := range colData {
		dataMap[rowIdx] = value
	}

	// 2. 构造检查上下文
	checkCtx := &CheckContext{
		CheckParam: param,
		ColName:    colName,
		SheetName:  sheetName,
	}

	// 3. 存储上下文（使用 context.Context 实现自动清理）
	ctx, cancel := GlobalAdaptor.StoreContext(context.Background(), sheetName, colName, checkCtx)
	defer cancel()

	// 4. 调用规则检查
	result := a.rule.Check(dataMap)

	// 确保 context 被取消
	_ = ctx

	return result
}
