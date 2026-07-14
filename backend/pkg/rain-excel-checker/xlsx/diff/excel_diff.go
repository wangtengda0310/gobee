// Package diff 提供 Excel 配表差异检测与上下文管理功能
// 本包包含列检查、参数解析、表查找等通用辅助函数
package diff

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"github.com/xuri/excelize/v2"
)

// ==================== 枚举表检测 ====================

// isEnumFormatCols 检测列数据是否为枚举表格式
// 枚举表的第0行以 "name" 和 "value" 开头
// 如 GuildCityName_城池名称表.xlsx（不是 _enum.xlsx 后缀但实际是枚举格式）
func isEnumFormatCols(cols [][]string) bool {
	if len(cols) < 2 {
		return false
	}
	// 检查第0列和第1列的第0行是否为 "name" 和 "value"
	firstRow0 := strings.ToLower(strings.TrimSpace(cols[0][0]))
	firstRow1 := strings.ToLower(strings.TrimSpace(cols[1][0]))
	return firstRow0 == "name" && firstRow1 == "value"
}

// getColNameRowIdx 获取列名所在行索引
// 枚举表：第0行是列名（name/value/type/sign/description）
// 普通表：第 MJS_FIXED_ROWS_NAME 行（索引2）是列名（Id/Name/...）
func getColNameRowIdx(cols [][]string, defaultRowIdx int) int {
	if isEnumFormatCols(cols) {
		return 0 // 枚举表第0行就是列名
	}
	return defaultRowIdx
}

// getDataStartRowIdx 获取数据起始行索引
// 枚举表：从第1行开始
// 普通表：从 startRowIdx 开始（通常是 MJS_FIXED_ROWS_NUM=4）
func getDataStartRowIdx(cols [][]string, defaultStartRow int) int {
	if isEnumFormatCols(cols) {
		return 1 // 枚举表数据从第1行开始
	}
	return defaultStartRow
}

// ==================== 常量定义 ====================

// BREAK_LINE 连续空行阈值，名将杀配表格式规范
// 连续遇到 BREAK_LINE 个空单元格时，视为数据结束（注释掉后续所有行）
const BREAK_LINE = 3

// compositeKeySep 复合主键列值之间的分隔符
// 使用不可见字符 \x01 避免与 Excel 数据内容碰撞
const compositeKeySep = "\x01"

// ==================== 数据结构定义 ====================

// ExcelSnapshot Excel数据快照（内存中）
type ExcelSnapshot struct {
	SheetName string                  `json:"sheetName"` // 表名
	ColNames  []string                `json:"colNames"`  // 列名列表（用于检测新增/删除列）
	Rows      map[string]*RowSnapshot `json:"rows"`      // 以ID为key的行数据
}

// RowSnapshot 行数据快照
type RowSnapshot struct {
	Id       string            `json:"id"`  // 展示用 ID（第一个 ID 列的原始值，用于通知消息）
	Key      string            `json:"key"` // 内部 map key（单列时等于 Id，复合主键时为组合值）
	Name     string            `json:"name"`
	Data     map[string]string `json:"data"` // 列名 -> 值
	RowIndex int               `json:"rowIndex"`
}

// ExcelDiffResult Excel差异检测结果
type ExcelDiffResult struct {
	TableName    string
	SheetName    string
	Timestamp    time.Time
	AddedRows    []*RowChange
	RemovedRows  []*RowChange
	ModifiedRows []*RowChange
	AddedCols    []string
	RemovedCols  []string
}

// RowChange 行变更信息
type RowChange struct {
	RowIndex int
	RowId    string
	RowName  string
	Changes  []*FieldChange
}

// FieldChange 字段变更信息
type FieldChange struct {
	ColName  string
	OldValue string
	NewValue string
}

// ColDiffResult 列变化检测结果
type ColDiffResult struct {
	AddedCols   []string // 新增的列名
	RemovedCols []string // 删除的列名
	Reordered   bool     // 列顺序是否变化（但列名相同）
}

// ==================== 快照构建 ====================

// BuildSnapshot 从列数据构建快照
//
// 🆕 适配"连续3个空行注释"规则：使用 helpers.AutoDetectEndIndex 确定数据结束位置，过滤掉注释区的数据
//
// 执行流程：
// 1. 创建快照结构，初始化基本信息（表名）
// 2. 遍历所有列，提取列名列表
// 3. 查找ID列和名称列的索引位置
// 4. 🆕 使用 helpers.AutoDetectEndIndex 确定数据结束位置（连续3个空行视为注释区）
// 5. 遍历数据行（只在有效数据范围内），为每行创建快照：
//   - 跳过空ID的行
//   - 记录行ID、行索引、行名称
//   - 记录所有列的值
//
// 6. 返回完整快照
func BuildSnapshot(sheetName string, cols [][]string, startRowIdx int, idColName, nameColName string) *ExcelSnapshot {
	// 步骤1: 创建快照结构
	snapshot := &ExcelSnapshot{
		SheetName: sheetName,
		ColNames:  make([]string, 0),
		Rows:      make(map[string]*RowSnapshot),
	}

	// 检测枚举表格式，确定列名行和数据起始行
	colNameRowIdx := getColNameRowIdx(cols, excelio.MJS_FIXED_ROWS_NAME)
	actualStartRowIdx := getDataStartRowIdx(cols, startRowIdx)

	// 步骤2: 提取列名列表
	for _, col := range cols {
		if len(col) > colNameRowIdx {
			colName := col[colNameRowIdx]
			if colName != "" {
				snapshot.ColNames = append(snapshot.ColNames, colName)
			}
		}
	}

	// 步骤3: 查找关键列（ID列、名称列）的索引位置
	idColIdx := excelio.GetColIndexByName(cols, idColName)
	nameColIdx := excelio.GetColIndexByName(cols, nameColName)

	// 如果没有找到指定的 ID 列，使用第一列作为行标识符
	useFirstColAsId := false
	if idColIdx < 0 {
		if len(cols) == 0 {
			return snapshot
		}
		idColIdx = 0
		useFirstColAsId = true
	}

	_ = useFirstColAsId // 后续可用于标记降级场景

	// 步骤4: 使用 helpers.AutoDetectEndIndex 确定数据结束位置
	// 连续 BREAK_LINE 个空单元格视为数据结束（注释掉后续所有行）
	commentStartIdx := helpers.AutoDetectEndIndex(cols, idColIdx, actualStartRowIdx, BREAK_LINE)

	// 步骤5: 遍历数据行，构建行快照（只在有效数据范围内）
	for rowIdx := actualStartRowIdx; rowIdx < commentStartIdx && rowIdx < len(cols[idColIdx]); rowIdx++ {
		rowId := strings.TrimSpace(cols[idColIdx][rowIdx])
		if rowId == "" {
			continue
		}

		// 为每行创建快照数据
		rowSnapshot := &RowSnapshot{
			Id:       rowId,
			Key:      rowId, // 单列模式下 Key 等于 Id
			RowIndex: rowIdx,
			Data:     make(map[string]string),
		}

		// 提取行名称
		if nameColIdx >= 0 && rowIdx < len(cols[nameColIdx]) {
			rowSnapshot.Name = strings.TrimSpace(cols[nameColIdx][rowIdx])
		}

		// 提取所有列的值
		for colIdx, col := range cols {
			if len(col) > colNameRowIdx {
				colName := col[colNameRowIdx]
				if colName != "" && rowIdx < len(col) {
					rowSnapshot.Data[colName] = strings.TrimSpace(col[rowIdx])
				}
			}
			_ = colIdx // 避免未使用变量警告
		}

		snapshot.Rows[rowId] = rowSnapshot
	}

	return snapshot
}

// ==================== 复合主键快照构建 ====================

// BuildSnapshotWithCompositeKey 支持复合主键的快照构建函数
//
// 与 BuildSnapshot 的区别：
//   - idColNames 支持多列（如 ["AnimationState", "ItemCfgId"]）
//   - 内部 map key 使用多列值拼接的组合 key，避免单列值重复导致覆盖
//   - RowSnapshot.Id 保持第一个 ID 列的原始值（用于展示）
//
// 降级策略：
//   - 所有 ID 列都找到 → 使用组合 key
//   - 部分列未找到 → 降级为已找到的第一个列
//   - 全部未找到 → 降级为第一列
func BuildSnapshotWithCompositeKey(sheetName string, cols [][]string, startRowIdx int, idColNames []string, nameColName string) *ExcelSnapshot {
	// 步骤1: 创建快照结构
	snapshot := &ExcelSnapshot{
		SheetName: sheetName,
		ColNames:  make([]string, 0),
		Rows:      make(map[string]*RowSnapshot),
	}

	// 检测枚举表格式，确定列名行和数据起始行
	colNameRowIdx := getColNameRowIdx(cols, excelio.MJS_FIXED_ROWS_NAME)
	actualStartRowIdx := getDataStartRowIdx(cols, startRowIdx)

	// 步骤2: 提取列名列表
	for _, col := range cols {
		if len(col) > colNameRowIdx {
			colName := col[colNameRowIdx]
			if colName != "" {
				snapshot.ColNames = append(snapshot.ColNames, colName)
			}
		}
	}

	// 步骤3: 查找所有 ID 列的索引位置
	idColIndices := make([]int, 0, len(idColNames))
	for _, name := range idColNames {
		idx := excelio.GetColIndexByName(cols, name)
		if idx >= 0 {
			idColIndices = append(idColIndices, idx)
		}
	}

	// 降级策略：如果没有找到任何 ID 列，使用第一列
	if len(idColIndices) == 0 {
		if len(cols) == 0 {
			return snapshot
		}
		idColIndices = []int{0}
	}

	// 只有一个 ID 列时，使用单列模式（兼容原逻辑）
	useCompositeKey := len(idColIndices) >= 2

	// 步骤4: 名称列索引
	nameColIdx := excelio.GetColIndexByName(cols, nameColName)

	// 步骤5: 使用第一个 ID 列检测注释区结束位置
	commentStartIdx := helpers.AutoDetectEndIndex(cols, idColIndices[0], actualStartRowIdx, BREAK_LINE)

	// 步骤6: 遍历数据行，构建行快照
	for rowIdx := actualStartRowIdx; rowIdx < commentStartIdx; rowIdx++ {
		// 检查是否超出所有 ID 列的范围
		valid := true
		for _, idx := range idColIndices {
			if rowIdx >= len(cols[idx]) {
				valid = false
				break
			}
		}
		if !valid {
			break
		}

		// 构建组合 key 或单列 key
		var mapKey string
		var displayId string // 展示用 ID（第一个 ID 列的值）

		if useCompositeKey {
			parts := make([]string, 0, len(idColIndices))
			for _, idx := range idColIndices {
				val := strings.TrimSpace(cols[idx][rowIdx])
				parts = append(parts, val)
			}
			mapKey = strings.Join(parts, compositeKeySep)
			displayId = strings.TrimSpace(cols[idColIndices[0]][rowIdx])
		} else {
			displayId = strings.TrimSpace(cols[idColIndices[0]][rowIdx])
			mapKey = displayId
		}

		// 跳过空 key（所有 ID 列都为空的行）
		if mapKey == "" || strings.Trim(mapKey, compositeKeySep) == "" {
			continue
		}

		rowSnapshot := &RowSnapshot{
			Id:       displayId,
			Key:      mapKey, // 内部 map key（单列时等于 Id，复合主键时为组合值）
			RowIndex: rowIdx,
			Data:     make(map[string]string),
		}

		// 提取行名称
		if nameColIdx >= 0 && rowIdx < len(cols[nameColIdx]) {
			rowSnapshot.Name = strings.TrimSpace(cols[nameColIdx][rowIdx])
		}

		// 提取所有列的值
		for colIdx, col := range cols {
			if len(col) > colNameRowIdx {
				colName := col[colNameRowIdx]
				if colName != "" && rowIdx < len(col) {
					rowSnapshot.Data[colName] = strings.TrimSpace(col[rowIdx])
				}
			}
			_ = colIdx // 避免未使用变量警告
		}

		snapshot.Rows[mapKey] = rowSnapshot
	}

	return snapshot
}

// DetectDiffWithCompositeKey 支持复合主键的差异检测函数
// 内部使用 BuildSnapshotWithCompositeKey 构建新快照，对比逻辑与 DetectDiff 相同
func DetectDiffWithCompositeKey(oldSnapshot *ExcelSnapshot, cols [][]string, startRowIdx int, idColNames []string, nameColName string) *ExcelDiffResult {
	// 复用 DetectDiff 逻辑：构建新快照后交给通用对比函数
	newSnapshot := BuildSnapshotWithCompositeKey(oldSnapshot.SheetName, cols, startRowIdx, idColNames, nameColName)
	return detectDiffInternal(oldSnapshot, newSnapshot)
}

// ==================== 差异检测 ====================

// DetectDiff 检测Excel数据差异（单列 ID 版本）
//
// 执行流程：
// 1. 如果旧快照为 nil，返回空结果（表示无历史数据对比）
// 2. 根据当前列数据构建新快照
// 3. 调用 detectDiffInternal 执行通用对比逻辑
func DetectDiff(oldSnapshot *ExcelSnapshot, cols [][]string, startRowIdx int, idColName, nameColName string) *ExcelDiffResult {
	// 步骤1: 如果没有旧快照，直接返回空结果
	if oldSnapshot == nil {
		return &ExcelDiffResult{
			Timestamp:    time.Now(),
			AddedRows:    make([]*RowChange, 0),
			RemovedRows:  make([]*RowChange, 0),
			ModifiedRows: make([]*RowChange, 0),
			AddedCols:    make([]string, 0),
			RemovedCols:  make([]string, 0),
		}
	}

	// 步骤2: 构建新数据的快照
	newSnapshot := BuildSnapshot(oldSnapshot.SheetName, cols, startRowIdx, idColName, nameColName)

	// 步骤3: 通用对比
	return detectDiffInternal(oldSnapshot, newSnapshot)
}

// detectDiffInternal 通用差异对比逻辑
// 由 DetectDiff 和 DetectDiffWithCompositeKey 共用
//
// 执行流程：
// 1. 初始化差异结果结构体
// 2. 检测列变化：调用 DetectColChanges 对比新旧列名列表
// 3. 检测行变化：
//   - 新增行：在新快照中存在但旧快照中不存在的行
//   - 删除行：在旧快照中存在但新快照中不存在的行
//   - 修改行：map key 相同但字段值发生变化的行
//
// 4. 排序结果：按 Excel 原始顺序排列（行按行号，字段按列顺序，列名按字母序）
func detectDiffInternal(oldSnapshot, newSnapshot *ExcelSnapshot) *ExcelDiffResult {
	result := &ExcelDiffResult{
		Timestamp:    time.Now(),
		AddedRows:    make([]*RowChange, 0),
		RemovedRows:  make([]*RowChange, 0),
		ModifiedRows: make([]*RowChange, 0),
		AddedCols:    make([]string, 0),
		RemovedCols:  make([]string, 0),
	}

	// 步骤2: 检测列变化
	colDiff := DetectColChanges(oldSnapshot.ColNames, newSnapshot.ColNames)
	result.AddedCols = colDiff.AddedCols
	result.RemovedCols = colDiff.RemovedCols

	// 步骤3: 检测行变化
	// 3.1 检测新增行：在新快照中存在但旧快照中不存在
	for _, newRow := range newSnapshot.Rows {
		if _, exists := oldSnapshot.Rows[newRow.Key]; !exists {
			result.AddedRows = append(result.AddedRows, &RowChange{
				RowIndex: newRow.RowIndex,
				RowId:    newRow.Id,
				RowName:  newRow.Name,
			})
		}
	}

	// 3.2 检测删除行：在旧快照中存在但新快照中不存在
	for _, oldRow := range oldSnapshot.Rows {
		if _, exists := newSnapshot.Rows[oldRow.Key]; !exists {
			result.RemovedRows = append(result.RemovedRows, &RowChange{
				RowIndex: oldRow.RowIndex,
				RowId:    oldRow.Id,
				RowName:  oldRow.Name,
			})
		}
	}

	// 3.3 检测修改行：map key 相同但字段值发生变化
	for mapKey, newRow := range newSnapshot.Rows {
		oldRow, exists := oldSnapshot.Rows[mapKey]
		if !exists {
			continue // 新增行，已处理
		}

		changes := make([]*FieldChange, 0)

		// 对比所有列的值，只检查新旧快照都有的列
		for colName, newValue := range newRow.Data {
			oldValue, hasOld := oldRow.Data[colName]

			if hasOld && oldValue != newValue {
				changes = append(changes, &FieldChange{
					ColName:  colName,
					OldValue: oldValue,
					NewValue: newValue,
				})
			}
		}

		if len(changes) > 0 {
			result.ModifiedRows = append(result.ModifiedRows, &RowChange{
				RowIndex: newRow.RowIndex,
				RowId:    newRow.Id,
				RowName:  newRow.Name,
				Changes:  changes,
			})
		}
	}

	// 步骤4: 排序结果，保持 Excel 中的原始顺序
	sortDiffResult(result, newSnapshot.ColNames)

	return result
}

// buildColOrderMap 构建列名到顺序索引的映射，用于按 Excel 列顺序排序字段变更
func buildColOrderMap(colNames []string) map[string]int {
	order := make(map[string]int, len(colNames))
	for i, name := range colNames {
		order[name] = i
	}
	return order
}

// sortDiffResult 对差异检测结果排序，保持 Excel 中的原始顺序
// - 行变更（新增/删除/修改）按 RowIndex 升序排列
// - 字段变更按 ColNames 中的列顺序排列
// - 列名变更按字母顺序排列
func sortDiffResult(result *ExcelDiffResult, colNames []string) {
	colOrder := buildColOrderMap(colNames)

	// 行按 RowIndex 升序（Excel 行号）
	sort.Slice(result.AddedRows, func(i, j int) bool {
		return result.AddedRows[i].RowIndex < result.AddedRows[j].RowIndex
	})
	sort.Slice(result.RemovedRows, func(i, j int) bool {
		return result.RemovedRows[i].RowIndex < result.RemovedRows[j].RowIndex
	})
	sort.Slice(result.ModifiedRows, func(i, j int) bool {
		return result.ModifiedRows[i].RowIndex < result.ModifiedRows[j].RowIndex
	})

	// 字段变更按 ColNames 中的列顺序
	for _, row := range result.ModifiedRows {
		sort.Slice(row.Changes, func(i, j int) bool {
			return colOrder[row.Changes[i].ColName] < colOrder[row.Changes[j].ColName]
		})
	}

	// 防御性排序（DetectColChanges 内部也会排序）
	sort.Strings(result.AddedCols)
	sort.Strings(result.RemovedCols)
}

// DetectColChanges 检测列变化
func DetectColChanges(oldColNames, newColNames []string) *ColDiffResult {
	result := &ColDiffResult{
		AddedCols:   make([]string, 0),
		RemovedCols: make([]string, 0),
		Reordered:   false,
	}

	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, name := range oldColNames {
		oldSet[name] = true
	}
	for _, name := range newColNames {
		newSet[name] = true
	}

	// 检测新增列
	for name := range newSet {
		if !oldSet[name] {
			result.AddedCols = append(result.AddedCols, name)
		}
	}

	// 检测删除列
	for name := range oldSet {
		if !newSet[name] {
			result.RemovedCols = append(result.RemovedCols, name)
		}
	}

	// 检测列顺序变化（列名相同但顺序不同）
	if len(oldColNames) == len(newColNames) && len(result.AddedCols) == 0 && len(result.RemovedCols) == 0 {
		for i := range oldColNames {
			if oldColNames[i] != newColNames[i] {
				result.Reordered = true
				break
			}
		}
	}

	// 防御性排序，确保输出顺序确定
	sort.Strings(result.AddedCols)
	sort.Strings(result.RemovedCols)

	return result
}

// ParseExcelFromBytes 从字节数组解析 Excel
// 返回按列存储的二维数组
func ParseExcelFromBytes(content []byte, sheetName string) ([][]string, error) {
	// 从字节数组读取 Excel
	tmpFile, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 文件失败: %w", err)
	}
	defer tmpFile.Close()

	// 读取指定 sheet 的所有行
	rows, err := tmpFile.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取 sheet 失败: %w", err)
	}

	// 转换为按列存储的格式
	if len(rows) == 0 {
		return [][]string{}, nil
	}

	// 计算最大列数
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// 按列组织数据
	cols := make([][]string, maxCols)
	for i := range cols {
		cols[i] = make([]string, len(rows))
	}

	// 填充列数据
	for rowIdx, row := range rows {
		for colIdx, cellValue := range row {
			if colIdx < maxCols {
				cols[colIdx][rowIdx] = cellValue
			}
		}
	}

	return cols, nil
}

// ==================== 通知格式化 ====================

// FormatDiffNotification 格式化差异通知
func FormatDiffNotification(diff *ExcelDiffResult, sheetName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📋 配表变更通知 - %s\n\n", sheetName))

	// 新增行
	if len(diff.AddedRows) > 0 {
		sb.WriteString(fmt.Sprintf("🆕 新增行 (%d条)\n", len(diff.AddedRows)))
		for _, row := range diff.AddedRows {
			sb.WriteString(fmt.Sprintf("• %s\n", excelio.FormatAddRowMessage(row.RowId, row.RowName)))
		}
		sb.WriteString("\n")
	}

	// 删除行
	if len(diff.RemovedRows) > 0 {
		sb.WriteString(fmt.Sprintf("🗑️ 删除行 (%d条)\n", len(diff.RemovedRows)))
		for _, row := range diff.RemovedRows {
			sb.WriteString(fmt.Sprintf("• %s\n", excelio.FormatAddRowMessage(row.RowId, row.RowName)))
		}
		sb.WriteString("\n")
	}

	// 新增列
	if len(diff.AddedCols) > 0 {
		sb.WriteString(fmt.Sprintf("📝 新增列 (%d个)\n", len(diff.AddedCols)))
		for _, col := range diff.AddedCols {
			sb.WriteString(fmt.Sprintf("• %s\n", col))
		}
		sb.WriteString("\n")
	}

	// 删除列
	if len(diff.RemovedCols) > 0 {
		sb.WriteString(fmt.Sprintf("📝 删除列 (%d个)\n", len(diff.RemovedCols)))
		for _, col := range diff.RemovedCols {
			sb.WriteString(fmt.Sprintf("• %s\n", col))
		}
		sb.WriteString("\n")
	}

	// 修改字段
	if len(diff.ModifiedRows) > 0 {
		sb.WriteString(fmt.Sprintf("📝 修改字段 (%d条)\n", len(diff.ModifiedRows)))
		for _, row := range diff.ModifiedRows {
			for _, change := range row.Changes {
				sb.WriteString(fmt.Sprintf("• %s\n",
					excelio.FormatChangeMessage(row.RowName, row.RowId, change.ColName, change.OldValue, change.NewValue)))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// HasChanges 判断是否有变更
func (d *ExcelDiffResult) HasChanges() bool {
	return len(d.AddedRows) > 0 ||
		len(d.RemovedRows) > 0 ||
		len(d.ModifiedRows) > 0 ||
		len(d.AddedCols) > 0 ||
		len(d.RemovedCols) > 0
}
