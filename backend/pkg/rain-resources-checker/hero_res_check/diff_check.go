package hero_res_check

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/diff"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def"
)

// Comparator 对比器
type Comparator struct {
	dataFile string
}

// NewComparator 创建对比器
func NewComparator(dataFile string) *Comparator {
	return &Comparator{
		dataFile: dataFile,
	}
}

// LoadPreviousData 加载上一次的所有数据
func (c *Comparator) LoadPreviousData() (*diff.DataContainer, error) {
	data, err := os.ReadFile(c.dataFile)
	if err != nil {
		return nil, err
	}

	var container diff.DataContainer
	if err := json.Unmarshal(data, &container); err != nil {
		return nil, err
	}

	return &container, nil
}

// SaveCurrentData 保存当前所有数据
func (c *Comparator) SaveCurrentData(container *diff.DataContainer) error {
	data, err := json.MarshalIndent(container, "", "  ")
	if err != nil {
		return err
	}
	// 确保目录存在
	dir := filepath.Dir(c.dataFile)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return os.WriteFile(c.dataFile, data, 0644)
}

// CompareAll 对比所有数据类型
func (c *Comparator) CompareAll(oldContainer, newContainer *diff.DataContainer) *diff.DiffResult {
	result := &diff.DiffResult{
		Timestamp:     time.Now(),
		DataTypeStats: make(map[string]*diff.TypeDiffResult),
	}

	// 使用反射遍历所有字段
	c.compareRoot(reflect.ValueOf(oldContainer).Elem(),
		reflect.ValueOf(newContainer).Elem(), result)

	return result
}

// compareRoot 对比根结构体
func (c *Comparator) compareRoot(oldVal, newVal reflect.Value, result *diff.DiffResult) {
	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return
	}

	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		if !field.IsExported() {
			continue
		}

		// 根据字段类型处理
		switch oldField.Kind() {
		case reflect.Slice, reflect.Array:
			// 处理切片/数组
			c.processSliceField(field.Name, oldField, newField, result)
		}
	}
}

// processSliceField 处理切片字段
func (c *Comparator) processSliceField(typeName string, oldSlice, newSlice reflect.Value, result *diff.DiffResult) {
	if oldSlice.Kind() != reflect.Slice || newSlice.Kind() != reflect.Slice {
		return
	}

	// 检查切片元素类型
	if oldSlice.Type().Elem().Kind() == reflect.Struct ||
		(oldSlice.Type().Elem().Kind() == reflect.Ptr &&
			oldSlice.Type().Elem().Elem().Kind() == reflect.Struct) {
		// 如果切片元素是结构体，进行深度对比
		typeResult := c.compareStructSlice(typeName, oldSlice, newSlice)
		if typeResult != nil && (len(typeResult.AddedIds) > 0 ||
			len(typeResult.RemovedIds) > 0 ||
			len(typeResult.ChangedItems) > 0) {
			result.DataTypeStats[typeName] = typeResult
		}
	}
}

// compareStructSlice 对比结构体切片
func (c *Comparator) compareStructSlice(typeName string, oldSlice, newSlice reflect.Value) *diff.TypeDiffResult {
	// 构建ID到元素的映射
	oldMap := make(map[int]reflect.Value)
	newMap := make(map[int]reflect.Value)

	// 处理旧数据
	for i := 0; i < oldSlice.Len(); i++ {
		item := oldSlice.Index(i)
		// 如果元素是指针，获取其指向的值
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		if id := c.extractID(item); id > 0 {
			oldMap[id] = item
		}
	}

	// 处理新数据
	for i := 0; i < newSlice.Len(); i++ {
		item := newSlice.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		if id := c.extractID(item); id > 0 {
			newMap[id] = item
		}
	}

	result := &diff.TypeDiffResult{
		DataType:      typeName,
		AddedIds:      make([]int, 0),
		RemovedIds:    make([]int, 0),
		ChangedItems:  make(map[int]*diff.FieldChanges),
		ItemCount:     newSlice.Len(),
		PreviousCount: oldSlice.Len(),
	}

	// 找出新增的
	for id := range newMap {
		if _, exists := oldMap[id]; !exists {
			result.AddedIds = append(result.AddedIds, id)
		}
	}

	// 找出删除的
	for id := range oldMap {
		if _, exists := newMap[id]; !exists {
			result.RemovedIds = append(result.RemovedIds, id)
		}
	}

	// 找出修改的
	for id, newItem := range newMap {
		if oldItem, exists := oldMap[id]; exists {
			// 创建字段变化记录
			fieldChanges := &diff.FieldChanges{
				Id:            id,
				Name:          c.getDisplayName(newItem),
				FieldPath:     fmt.Sprintf("%s[%d]", typeName, id),
				Changes:       make([]*diff.FieldChange, 0),
				NestedStructs: make(map[string]*diff.NestedStructChange),
			}

			// 深度对比两个结构体
			c.compareStructs(oldItem, newItem, fieldChanges.FieldPath, fieldChanges)

			if len(fieldChanges.Changes) > 0 || len(fieldChanges.NestedStructs) > 0 {
				result.ChangedItems[id] = fieldChanges
			}
		}
	}

	return result
}

// compareStructs 对比两个结构体
func (c *Comparator) compareStructs(oldStruct, newStruct reflect.Value, basePath string, fieldChanges *diff.FieldChanges) {
	if oldStruct.Kind() != reflect.Struct || newStruct.Kind() != reflect.Struct {
		return
	}

	oldType := oldStruct.Type()

	for i := 0; i < oldStruct.NumField(); i++ {
		field := oldType.Field(i)
		oldField := oldStruct.Field(i)
		newField := newStruct.Field(i)

		// 跳过未导出字段和ID字段
		if !field.IsExported() || field.Name == "Id" || field.Name == "ID" {
			continue
		}

		fieldPath := basePath + "." + field.Name

		// 根据字段类型处理
		switch oldField.Kind() {
		case reflect.Struct:
			// 处理嵌套结构体
			c.processNestedStruct(field.Name, fieldPath, oldField, newField, fieldChanges)

		case reflect.Slice, reflect.Array:
			// 处理切片
			c.processNestedSlice(field.Name, fieldPath, oldField, newField, fieldChanges)

		case reflect.Ptr:
			// 处理指针
			c.processNestedPtr(field.Name, fieldPath, oldField, newField, fieldChanges)

		case reflect.Map:
			// 处理Map
			c.processMap(field.Name, fieldPath, oldField, newField, fieldChanges)

		default:
			// 基本类型字段
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				change := &diff.FieldChange{
					FieldPath:   fieldPath,
					FieldName:   field.Name,
					StructName:  oldType.Name(),
					NestedLevel: strings.Count(fieldPath, "."),
					OldValue:    oldField.Interface(),
					NewValue:    newField.Interface(),
					ValueType:   oldField.Kind().String(),
					ChangeType:  diff.ChangeTypeModified,
				}
				fieldChanges.Changes = append(fieldChanges.Changes, change)
			}
		}
	}
}

// processNestedStruct 处理嵌套结构体
func (c *Comparator) processNestedStruct(fieldName, fieldPath string, oldField, newField reflect.Value, fieldChanges *diff.FieldChanges) {
	structType := oldField.Type().Name()
	if structType == "" {
		structType = "anonymous"
	}

	// 创建嵌套结构体变化记录
	nestedChange := c.getOrCreateNestedStruct(fieldChanges, fieldPath, structType)

	// 递归对比嵌套结构体的字段
	oldType := oldField.Type()
	hasChanges := false

	for i := 0; i < oldField.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() || field.Name == "Id" || field.Name == "ID" {
			continue
		}

		subOldField := oldField.Field(i)
		subNewField := newField.Field(i)
		subFieldPath := fieldPath + "." + field.Name

		if !reflect.DeepEqual(subOldField.Interface(), subNewField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   subFieldPath,
				FieldName:   field.Name,
				StructName:  structType,
				NestedLevel: strings.Count(subFieldPath, "."),
				OldValue:    subOldField.Interface(),
				NewValue:    subNewField.Interface(),
				ValueType:   subOldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
			hasChanges = true
		}
	}

	// 如果没有变化，从map中移除
	if !hasChanges {
		delete(fieldChanges.NestedStructs, fieldPath)
	}
}

// processNestedSlice 处理嵌套切片
func (c *Comparator) processNestedSlice(fieldName, fieldPath string, oldSlice, newSlice reflect.Value, fieldChanges *diff.FieldChanges) {
	// 检查切片元素类型
	if oldSlice.Len() > 0 && oldSlice.Index(0).Kind() == reflect.Struct {
		elemType := oldSlice.Type().Elem().Name()
		sliceChange := c.getOrCreateNestedStruct(fieldChanges, fieldPath, "[]"+elemType)

		// 构建元素映射
		oldMap := make(map[int]reflect.Value)
		newMap := make(map[int]reflect.Value)

		// 处理旧切片
		for i := 0; i < oldSlice.Len(); i++ {
			item := oldSlice.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			if id := c.extractID(item); id > 0 {
				oldMap[id] = item
			}
		}

		// 处理新切片
		for i := 0; i < newSlice.Len(); i++ {
			item := newSlice.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			if id := c.extractID(item); id > 0 {
				newMap[id] = item
			}
		}

		// 检查新增的元素
		for id, newItem := range newMap {
			if _, exists := oldMap[id]; !exists {
				elemPath := fmt.Sprintf("%s[%d]", fieldPath, id)
				change := &diff.FieldChange{
					FieldPath:   elemPath,
					FieldName:   fmt.Sprintf("元素[%d]", id),
					StructName:  elemType,
					NestedLevel: strings.Count(elemPath, "."),
					OldValue:    nil,
					NewValue:    c.getDisplayName(newItem),
					ValueType:   "struct",
					ChangeType:  diff.ChangeTypeAdded,
				}
				sliceChange.Changes = append(sliceChange.Changes, change)
				sliceChange.FieldCount++
			}
		}

		// 检查删除的元素
		for id, oldItem := range oldMap {
			if _, exists := newMap[id]; !exists {
				elemPath := fmt.Sprintf("%s[%d]", fieldPath, id)
				change := &diff.FieldChange{
					FieldPath:   elemPath,
					FieldName:   fmt.Sprintf("元素[%d]", id),
					StructName:  elemType,
					NestedLevel: strings.Count(elemPath, "."),
					OldValue:    c.getDisplayName(oldItem),
					NewValue:    nil,
					ValueType:   "struct",
					ChangeType:  diff.ChangeTypeRemoved,
				}
				sliceChange.Changes = append(sliceChange.Changes, change)
				sliceChange.FieldCount++
			}
		}

		// 检查修改的元素
		for id, newItem := range newMap {
			if oldItem, exists := oldMap[id]; exists {
				elemPath := fmt.Sprintf("%s[%d]", fieldPath, id)
				// 为每个元素创建子嵌套结构体
				elemChange := c.getOrCreateNestedStruct(fieldChanges, elemPath, elemType)
				c.compareStructs(oldItem, newItem, elemPath, fieldChanges)

				// 如果元素有变化，确保它被记录
				if len(elemChange.Changes) > 0 {
					elemChange.FieldCount = len(elemChange.Changes)
				}
			}
		}

		// 如果没有变化，从map中移除
		if sliceChange.FieldCount == 0 {
			delete(fieldChanges.NestedStructs, fieldPath)
		}
	} else {
		// 基本类型切片
		if !reflect.DeepEqual(oldSlice.Interface(), newSlice.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   fieldName,
				StructName:  "slice",
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldSlice.Interface(),
				NewValue:    newSlice.Interface(),
				ValueType:   "slice",
				ChangeType:  diff.ChangeTypeModified,
			}
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}
}

// processNestedPtr 处理嵌套指针
func (c *Comparator) processNestedPtr(fieldName, fieldPath string, oldPtr, newPtr reflect.Value, fieldChanges *diff.FieldChanges) {
	oldIsNil := oldPtr.IsNil()
	newIsNil := newPtr.IsNil()

	if oldIsNil && newIsNil {
		return
	}

	ptrType := oldPtr.Type().Elem().Name()
	ptrChange := c.getOrCreateNestedStruct(fieldChanges, fieldPath, "*"+ptrType)

	if oldIsNil != newIsNil {
		change := &diff.FieldChange{
			FieldPath:   fieldPath,
			FieldName:   fieldName,
			StructName:  "pointer",
			NestedLevel: strings.Count(fieldPath, "."),
			OldValue:    oldIsNil,
			NewValue:    newIsNil,
			ValueType:   "pointer",
			ChangeType:  diff.ChangeTypeModified,
		}
		ptrChange.Changes = append(ptrChange.Changes, change)
		ptrChange.FieldCount++
		return
	}

	// 两者都不为nil，对比指向的值
	oldElem := oldPtr.Elem()
	newElem := newPtr.Elem()

	if oldElem.Kind() == reflect.Struct {
		// 为指针指向的结构体创建嵌套记录
		c.getOrCreateNestedStruct(fieldChanges, fieldPath+"->", ptrType)
		c.compareStructs(oldElem, newElem, fieldPath+"->", fieldChanges)
	}
}

// processMap 处理Map类型
func (c *Comparator) processMap(fieldName, fieldPath string, oldMap, newMap reflect.Value, fieldChanges *diff.FieldChanges) {
	if !reflect.DeepEqual(oldMap.Interface(), newMap.Interface()) {
		change := &diff.FieldChange{
			FieldPath:   fieldPath,
			FieldName:   fieldName,
			StructName:  "map",
			NestedLevel: strings.Count(fieldPath, "."),
			OldValue:    oldMap.Interface(),
			NewValue:    newMap.Interface(),
			ValueType:   "map",
			ChangeType:  diff.ChangeTypeModified,
		}
		fieldChanges.Changes = append(fieldChanges.Changes, change)
	}
}

// getOrCreateNestedStruct 获取或创建嵌套结构体变化记录
func (c *Comparator) getOrCreateNestedStruct(fieldChanges *diff.FieldChanges, path, structType string) *diff.NestedStructChange {
	if fieldChanges.NestedStructs == nil {
		fieldChanges.NestedStructs = make(map[string]*diff.NestedStructChange)
	}

	if _, exists := fieldChanges.NestedStructs[path]; !exists {
		fieldChanges.NestedStructs[path] = &diff.NestedStructChange{
			StructPath: path,
			StructType: structType,
			Changes:    make([]*diff.FieldChange, 0),
			Children:   make(map[string]*diff.NestedStructChange),
		}
	}

	return fieldChanges.NestedStructs[path]
}

// extractID 从结构体中提取ID字段
func (c *Comparator) extractID(val reflect.Value) int {
	if val.Kind() != reflect.Struct {
		return 0
	}

	// 尝试获取ID字段
	idField := val.FieldByName("Id")
	if !idField.IsValid() {
		idField = val.FieldByName("ID")
	}

	if idField.IsValid() && idField.Kind() == reflect.Int {
		return int(idField.Int())
	}

	return 0
}

// getDisplayName 获取显示名称
func (c *Comparator) getDisplayName(val reflect.Value) string {
	if val.Kind() != reflect.Struct {
		return ""
	}

	// 尝试获取Name字段
	nameField := val.FieldByName("SkillName")
	if nameField.IsValid() && nameField.Kind() == reflect.String {
		return nameField.String()
	}

	// 尝试获取Title字段
	titleField := val.FieldByName("Title")
	if titleField.IsValid() && titleField.Kind() == reflect.String {
		return titleField.String()
	}

	return fmt.Sprintf("ID:%d", c.extractID(val))
}

// ----------------------报告生成功能---------------------------

// ReportGenerator 报告生成器
type ReportGenerator struct{}

// NewReportGenerator 创建报告生成器
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// PrintSummary 打印对比摘要
func (rg *ReportGenerator) PrintSummary(result *diff.DiffResult) {
	fmt.Printf("========== 数据对比报告 ==========\n")
	fmt.Printf("对比时间: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05"))

	totalAdditions := 0
	totalDeletions := 0
	totalModifications := 0

	for dataType, stat := range result.DataTypeStats {
		fmt.Printf("【%s】\n", dataType)
		fmt.Printf("  数量变化: %d → %d (%+d)\n",
			stat.PreviousCount, stat.ItemCount, stat.ItemCount-stat.PreviousCount)
		fmt.Printf("  新增: %d个\n", len(stat.AddedIds))
		fmt.Printf("  删除: %d个\n", len(stat.RemovedIds))
		fmt.Printf("  修改: %d个\n", len(stat.ChangedItems))

		if len(stat.AddedIds) > 0 {
			fmt.Printf("  新增ID列表: %v\n", stat.AddedIds)
		}
		if len(stat.RemovedIds) > 0 {
			fmt.Printf("  删除ID列表: %v\n", stat.RemovedIds)
		}

		totalAdditions += len(stat.AddedIds)
		totalDeletions += len(stat.RemovedIds)
		totalModifications += len(stat.ChangedItems)

		fmt.Println()
	}

	fmt.Printf("========== 总计 ==========\n")
	fmt.Printf("总新增: %d, 总删除: %d, 总修改: %d\n",
		totalAdditions, totalDeletions, totalModifications)
}

// PrintDetailedChanges 打印详细变化
func (rg *ReportGenerator) PrintDetailedChanges(result *diff.DiffResult) {
	for dataType, stat := range result.DataTypeStats {
		if len(stat.ChangedItems) > 0 {
			fmt.Printf("\n📦 【%s】详细修改:\n", dataType)
			fmt.Println(strings.Repeat("─", 50))

			for id, changes := range stat.ChangedItems {
				rg.printFieldChanges(id, changes, "")
			}
		}
	}
}

// printFieldChanges 递归打印字段变化
func (rg *ReportGenerator) printFieldChanges(id int, changes *diff.FieldChanges, indent string) {
	fmt.Printf("%s📌 ID: %d (%s)\n", indent, id, changes.Name)
	fmt.Printf("%s  路径: %s\n", indent, changes.FieldPath)

	// 打印基本字段变化
	for _, change := range changes.Changes {
		fmt.Printf("%s  • %s:\n", indent, change.FieldPath)
		fmt.Printf("%s    旧值: %v\n", indent, rg.formatValue(change.OldValue))
		fmt.Printf("%s    新值: %v\n", indent, rg.formatValue(change.NewValue))
	}

	// 打印嵌套结构体变化
	if len(changes.NestedStructs) > 0 {
		fmt.Printf("%s  嵌套结构体:\n", indent)
		rg.printNestedStructs(changes.NestedStructs, indent+"    ")
	}
}

// printNestedStructs 递归打印嵌套结构体
func (rg *ReportGenerator) printNestedStructs(nestedStructs map[string]*diff.NestedStructChange, indent string) {
	for path, nested := range nestedStructs {
		if nested.FieldCount > 0 {
			fmt.Printf("%s📁 %s [%s] (%d个变化)\n", indent, path, nested.StructType, nested.FieldCount)

			for _, change := range nested.Changes {
				fmt.Printf("%s  • %s:\n", indent, change.FieldName)
				fmt.Printf("%s    旧: %v\n", indent, rg.formatValue(change.OldValue))
				fmt.Printf("%s    新: %v\n", indent, rg.formatValue(change.NewValue))
			}

			// 递归打印子结构体
			if len(nested.Children) > 0 {
				rg.printNestedStructs(nested.Children, indent+"  ")
			}
		}
	}
}

// formatValue 格式化值显示
func (rg *ReportGenerator) formatValue(val interface{}) string {
	if val == nil {
		return "nil"
	}

	switch v := val.(type) {
	case []int:
		return fmt.Sprintf("%v", v)
	case []string:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return "map[...]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GenerateJSONReport 生成JSON格式报告
func (rg *ReportGenerator) GenerateJSONReport(result *diff.DiffResult) (map[string]interface{}, error) {
	report := make(map[string]interface{})
	report["timestamp"] = result.Timestamp
	report["summary"] = make(map[string]interface{})

	for dataType, stat := range result.DataTypeStats {
		typeReport := make(map[string]interface{})
		typeReport["previous_count"] = stat.PreviousCount
		typeReport["current_count"] = stat.ItemCount
		typeReport["added_count"] = len(stat.AddedIds)
		typeReport["removed_count"] = len(stat.RemovedIds)
		typeReport["modified_count"] = len(stat.ChangedItems)
		typeReport["added_ids"] = stat.AddedIds
		typeReport["removed_ids"] = stat.RemovedIds

		// 添加修改详情
		modifications := make([]map[string]interface{}, 0)
		for id, changes := range stat.ChangedItems {
			mod := rg.buildChangeDetail(id, changes)
			modifications = append(modifications, mod)
		}
		typeReport["modifications"] = modifications

		report["summary"].(map[string]interface{})[dataType] = typeReport
	}

	return report, nil
}

// buildChangeDetail 构建变化详情
func (rg *ReportGenerator) buildChangeDetail(id int, changes *diff.FieldChanges) map[string]interface{} {
	detail := make(map[string]interface{})
	detail["id"] = id
	detail["name"] = changes.Name
	detail["field_path"] = changes.FieldPath

	fieldChanges := make([]map[string]interface{}, 0)
	for _, change := range changes.Changes {
		fieldChange := map[string]interface{}{
			"field_path":  change.FieldPath,
			"field_name":  change.FieldName,
			"struct":      change.StructName,
			"old_value":   change.OldValue,
			"new_value":   change.NewValue,
			"change_type": change.ChangeType,
		}
		fieldChanges = append(fieldChanges, fieldChange)
	}
	detail["field_changes"] = fieldChanges

	// 添加嵌套结构体信息
	if len(changes.NestedStructs) > 0 {
		detail["nested_structs"] = rg.buildNestedStructs(changes.NestedStructs)
	}

	return detail
}

// buildNestedStructs 构建嵌套结构体信息
func (rg *ReportGenerator) buildNestedStructs(nestedStructs map[string]*diff.NestedStructChange) map[string]interface{} {
	result := make(map[string]interface{})

	for path, nested := range nestedStructs {
		nestedInfo := make(map[string]interface{})
		nestedInfo["struct_type"] = nested.StructType
		nestedInfo["field_count"] = nested.FieldCount

		changes := make([]map[string]interface{}, 0)
		for _, change := range nested.Changes {
			changes = append(changes, map[string]interface{}{
				"field_name": change.FieldName,
				"old_value":  change.OldValue,
				"new_value":  change.NewValue,
			})
		}
		nestedInfo["changes"] = changes

		if len(nested.Children) > 0 {
			nestedInfo["children"] = rg.buildNestedStructs(nested.Children)
		}

		result[path] = nestedInfo
	}

	return result
}

// / ----------------------- 下面是herowiki专用 -------------------------

// CompareHeroWikiDiff 专门对比HeroWikiDiff
func (c *Comparator) CompareHeroWikiDiff(oldWiki, newWiki *herowiki_def.HeroWikiDiff) *diff.HeroWikiDiffResult {
	if oldWiki == nil || newWiki == nil {
		return nil
	}

	result := &diff.HeroWikiDiffResult{
		Timestamp:         time.Now(),
		HeroesDiff:        make(map[string]*diff.HeroDiffDetail),
		RemovedHeroesData: make(map[string]*herowiki_def.HeroCompleteData),
		Summary: &diff.HeroWikiDiffSummary{
			AddedHeroes:    make([]string, 0),
			RemovedHeroes:  make([]string, 0),
			ModifiedHeroes: make([]string, 0),
		},
	}

	// 找出新增、删除、修改的武将
	for heroId, newHero := range newWiki.Heroes {
		if oldHero, exists := oldWiki.Heroes[heroId]; exists {
			// 对比修改的武将
			changes := c.compareHeroCompleteData(oldHero, newHero, heroId)
			// 检查是否有实际变化
			hasChanges := len(changes.FieldChanges) > 0
			for _, nested := range changes.NestedChanges {
				if nested.FieldCount > 0 {
					hasChanges = true
					break
				}
			}

			if hasChanges {
				changes.ChangeType = diff.ChangeTypeModified
				changes.ChangeCount = len(changes.FieldChanges)
				for _, nested := range changes.NestedChanges {
					changes.ChangeCount += nested.FieldCount
				}
				result.HeroesDiff[heroId] = changes
				result.Summary.ModifiedHeroes = append(result.Summary.ModifiedHeroes, heroId)
				result.Summary.TotalChanges += changes.ChangeCount
			}
		} else {
			// 新增的武将
			changes := &diff.HeroDiffDetail{
				EHeroId:     heroId,
				Name:        newHero.Basic.Name,
				ChangeType:  diff.ChangeTypeAdded,
				ChangeCount: 1,
			}
			result.HeroesDiff[heroId] = changes
			result.Summary.AddedHeroes = append(result.Summary.AddedHeroes, heroId)
			result.Summary.TotalChanges++
		}
	}

	// 找出删除的武将
	for heroId, oldHero := range oldWiki.Heroes {
		if _, exists := newWiki.Heroes[heroId]; !exists {
			// 创建一个空的 newHero，用于生成完整的字段变化
			emptyNewHero := &herowiki_def.HeroCompleteData{}
			changes := c.compareHeroCompleteData(oldHero, emptyNewHero, heroId)
			changes.ChangeType = diff.ChangeTypeRemoved
			// 重新计算 ChangeCount
			changes.ChangeCount = len(changes.FieldChanges)
			for _, nested := range changes.NestedChanges {
				changes.ChangeCount += nested.FieldCount
			}
			if changes.ChangeCount == 0 {
				changes.ChangeCount = 1 // 确保至少有1个变化（整个武将被删除）
			}
			result.HeroesDiff[heroId] = changes
			result.Summary.RemovedHeroes = append(result.Summary.RemovedHeroes, heroId)
			result.Summary.TotalChanges++

			// 存储删除武将的完整数据，供前端显示
			result.RemovedHeroesData[heroId] = oldHero
		}
	}

	result.Summary.TotalHeroes = len(newWiki.Heroes)
	return result
}

// compareHeroCompleteData 对比单个武将的完整数据
func (c *Comparator) compareHeroCompleteData(oldHero, newHero *herowiki_def.HeroCompleteData, heroId string) *diff.HeroDiffDetail {
	// 优先使用 oldHero 的名称（用于删除武将的情况）
	heroName := ""
	heroIdNum := 0
	if oldHero != nil && oldHero.Basic != nil {
		heroName = oldHero.Basic.Name
		heroIdNum = oldHero.Basic.Id
	} else if newHero != nil && newHero.Basic != nil {
		heroName = newHero.Basic.Name
		heroIdNum = newHero.Basic.Id
	}

	detail := &diff.HeroDiffDetail{
		EHeroId:       heroId,
		Name:          heroName,
		FieldChanges:  make([]*diff.FieldChange, 0),
		NestedChanges: make(map[string]*diff.NestedStructChange),
	}

	fieldChanges := &diff.FieldChanges{
		Id:            heroIdNum,
		Name:          heroName,
		FieldPath:     "HeroWikiDiff.Heroes[" + heroId + "]",
		Changes:       make([]*diff.FieldChange, 0),
		NestedStructs: make(map[string]*diff.NestedStructChange),
	}

	// 对比各个部分 - 这些函数会填充fieldChanges.NestedStructs
	c.compareHeroBasic(oldHero.Basic, newHero.Basic, fieldChanges)
	c.compareHeroUI(oldHero.UI, newHero.UI, fieldChanges)
	c.compareHeroSkills(oldHero.Skills, newHero.Skills, fieldChanges)
	c.compareHeroSkins(oldHero.Skins, newHero.Skins, fieldChanges)
	c.compareHeroAchievements(oldHero.Achievements, newHero.Achievements, fieldChanges)
	c.compareCountry(oldHero.Country, newHero.Country, fieldChanges)
	c.compareRecommendBd(oldHero.RecommendBd, newHero.RecommendBd, fieldChanges)
	c.compareHeroRobotActions(oldHero.RobotActions, newHero.RobotActions, fieldChanges)
	c.compareHeroDropInfos(oldHero.DropInfo, newHero.DropInfo, fieldChanges)

	// 转换格式 - 只保留有实际变化的嵌套结构
	detail.FieldChanges = fieldChanges.Changes

	for path, nested := range fieldChanges.NestedStructs {
		// 递归清理空的变化
		c.cleanEmptyNestedStruct(nested)
		if nested.FieldCount > 0 {
			detail.NestedChanges[path] = nested
		}
	}

	return detail
}

// cleanEmptyNestedStruct 递归清理空的嵌套结构
func (c *Comparator) cleanEmptyNestedStruct(nested *diff.NestedStructChange) {
	if nested == nil {
		return
	}

	// 清理子结构
	toDelete := make([]string, 0)
	for path, child := range nested.Children {
		c.cleanEmptyNestedStruct(child)
		if child.FieldCount == 0 && len(child.Changes) == 0 {
			toDelete = append(toDelete, path)
		}
	}
	for _, path := range toDelete {
		delete(nested.Children, path)
	}

	// 更新FieldCount
	nested.FieldCount = len(nested.Changes)
	for _, child := range nested.Children {
		nested.FieldCount += child.FieldCount
	}
}

// compareHeroBasic 对比武将基础信息
func (c *Comparator) compareHeroBasic(oldBasic, newBasic *herowiki_def.HeroBasicInfo, fieldChanges *diff.FieldChanges) {
	if oldBasic == nil && newBasic == nil {
		return
	}

	basePath := fieldChanges.FieldPath + ".Basic"

	if oldBasic == nil && newBasic != nil {
		// 新增基础信息
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroBasicInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "Basic",
			StructName:  "HeroBasicInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    nil,
			NewValue:    "新增基础信息",
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeAdded,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	if oldBasic != nil && newBasic == nil {
		// 删除基础信息
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroBasicInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "Basic",
			StructName:  "HeroBasicInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    "删除基础信息",
			NewValue:    nil,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeRemoved,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	// 两者都存在，对比字段
	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroBasicInfo")

	// 对比所有字段
	oldVal := reflect.ValueOf(oldBasic).Elem()
	newVal := reflect.ValueOf(newBasic).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   field.Name,
				StructName:  "HeroBasicInfo",
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldField.Interface(),
				NewValue:    newField.Interface(),
				ValueType:   oldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}
}

// compareHeroUI 对比武将UI信息
func (c *Comparator) compareHeroUI(oldUI, newUI *herowiki_def.HeroUIInfo, fieldChanges *diff.FieldChanges) {
	if oldUI == nil && newUI == nil {
		return
	}

	basePath := fieldChanges.FieldPath + ".UI"

	if oldUI == nil && newUI != nil {
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroUIInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "UI",
			StructName:  "HeroUIInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    nil,
			NewValue:    "新增UI信息",
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeAdded,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	if oldUI != nil && newUI == nil {
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroUIInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "UI",
			StructName:  "HeroUIInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    "删除UI信息",
			NewValue:    nil,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeRemoved,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroUIInfo")
	oldVal := reflect.ValueOf(oldUI).Elem()
	newVal := reflect.ValueOf(newUI).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   field.Name,
				StructName:  "HeroUIInfo",
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldField.Interface(),
				NewValue:    newField.Interface(),
				ValueType:   oldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}
}

// compareHeroSkills 对比武将技能
func (c *Comparator) compareHeroSkills(oldSkills, newSkills []*herowiki_def.HeroSkillInfo, fieldChanges *diff.FieldChanges) {
	basePath := fieldChanges.FieldPath + ".Skills"

	if len(oldSkills) == 0 && len(newSkills) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "[]HeroSkillInfo")

	// 构建技能映射
	oldMap := make(map[string]*herowiki_def.HeroSkillInfo)
	newMap := make(map[string]*herowiki_def.HeroSkillInfo)

	for _, skill := range oldSkills {
		if skill != nil && skill.Basic != nil {
			oldMap[skill.Basic.Id] = skill
		}
	}

	for _, skill := range newSkills {
		if skill != nil && skill.Basic != nil {
			newMap[skill.Basic.Id] = skill
		}
	}

	// 处理新增的技能
	for id, newSkill := range newMap {
		if _, exists := oldMap[id]; !exists {
			skillPath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   skillPath,
				FieldName:   fmt.Sprintf("技能[%s]", id),
				StructName:  "HeroSkillInfo",
				NestedLevel: strings.Count(skillPath, "."),
				OldValue:    nil,
				NewValue:    newSkill.Basic.SkillName,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 处理删除的技能
	for id, oldSkill := range oldMap {
		if _, exists := newMap[id]; !exists {
			skillPath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   skillPath,
				FieldName:   fmt.Sprintf("技能[%s]", id),
				StructName:  "HeroSkillInfo",
				NestedLevel: strings.Count(skillPath, "."),
				OldValue:    oldSkill.Basic.SkillName,
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 处理修改的技能
	for id, newSkill := range newMap {
		if oldSkill, exists := oldMap[id]; exists {
			skillPath := fmt.Sprintf("%s[%s]", basePath, id)
			c.compareSingleSkill(oldSkill, newSkill, skillPath, fieldChanges)
		}
	}
}

// compareSingleSkill 对比单个技能
func (c *Comparator) compareSingleSkill(oldSkill, newSkill *herowiki_def.HeroSkillInfo, basePath string, fieldChanges *diff.FieldChanges) {
	// 对比技能基础信息
	if oldSkill.Basic != nil || newSkill.Basic != nil {
		if oldSkill.Basic != nil && newSkill.Basic != nil {
			basicPath := basePath + ".Basic"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, basicPath, "SkillBasicInfo")
			c.compareStructFields(
				reflect.ValueOf(oldSkill.Basic).Elem(),
				reflect.ValueOf(newSkill.Basic).Elem(),
				basicPath,
				nestedChange,
			)
		} else if oldSkill.Basic == nil && newSkill.Basic != nil {
			basicPath := basePath + ".Basic"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, basicPath, "SkillBasicInfo")
			change := &diff.FieldChange{
				FieldPath:   basicPath,
				FieldName:   "Basic",
				StructName:  "SkillBasicInfo",
				NestedLevel: strings.Count(basicPath, "."),
				OldValue:    nil,
				NewValue:    "新增技能基础信息",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		} else if oldSkill.Basic != nil && newSkill.Basic == nil {
			basicPath := basePath + ".Basic"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, basicPath, "SkillBasicInfo")
			change := &diff.FieldChange{
				FieldPath:   basicPath,
				FieldName:   "Basic",
				StructName:  "SkillBasicInfo",
				NestedLevel: strings.Count(basicPath, "."),
				OldValue:    "删除技能基础信息",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}

	// 对比技能UI信息
	if oldSkill.UI != nil || newSkill.UI != nil {
		if oldSkill.UI != nil && newSkill.UI != nil {
			uiPath := basePath + ".UI"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, uiPath, "SkillUIInfo")
			c.compareStructFields(
				reflect.ValueOf(oldSkill.UI).Elem(),
				reflect.ValueOf(newSkill.UI).Elem(),
				uiPath,
				nestedChange,
			)
		} else if oldSkill.UI == nil && newSkill.UI != nil {
			uiPath := basePath + ".UI"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, uiPath, "SkillUIInfo")
			change := &diff.FieldChange{
				FieldPath:   uiPath,
				FieldName:   "UI",
				StructName:  "SkillUIInfo",
				NestedLevel: strings.Count(uiPath, "."),
				OldValue:    nil,
				NewValue:    "新增技能UI信息",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		} else if oldSkill.UI != nil && newSkill.UI == nil {
			uiPath := basePath + ".UI"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, uiPath, "SkillUIInfo")
			change := &diff.FieldChange{
				FieldPath:   uiPath,
				FieldName:   "UI",
				StructName:  "SkillUIInfo",
				NestedLevel: strings.Count(uiPath, "."),
				OldValue:    "删除技能UI信息",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}

	// 对比技能熔炼信息
	if oldSkill.Melt != nil || newSkill.Melt != nil {
		if oldSkill.Melt != nil && newSkill.Melt != nil {
			meltPath := basePath + ".Melt"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, meltPath, "SkillMeltInfo")
			c.compareStructFields(
				reflect.ValueOf(oldSkill.Melt).Elem(),
				reflect.ValueOf(newSkill.Melt).Elem(),
				meltPath,
				nestedChange,
			)
		} else if oldSkill.Melt == nil && newSkill.Melt != nil {
			meltPath := basePath + ".Melt"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, meltPath, "SkillMeltInfo")
			change := &diff.FieldChange{
				FieldPath:   meltPath,
				FieldName:   "Melt",
				StructName:  "SkillMeltInfo",
				NestedLevel: strings.Count(meltPath, "."),
				OldValue:    nil,
				NewValue:    "新增技能熔炼信息",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		} else if oldSkill.Melt != nil && newSkill.Melt == nil {
			meltPath := basePath + ".Melt"
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, meltPath, "SkillMeltInfo")
			change := &diff.FieldChange{
				FieldPath:   meltPath,
				FieldName:   "Melt",
				StructName:  "SkillMeltInfo",
				NestedLevel: strings.Count(meltPath, "."),
				OldValue:    "删除技能熔炼信息",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}

	// 对比技能台词
	c.compareSkillLines(oldSkill.Lines, newSkill.Lines, basePath, fieldChanges)

	// 对比技能标签
	c.compareSkillTags(oldSkill.Tags, newSkill.Tags, basePath, fieldChanges)
}

// compareStructFields 通用的结构体字段对比函数
// isEmptyEquivalent 判断两个值是否"空等价"
// 避免 JSON 序列化导致的 [] vs null 误报
func isEmptyEquivalent(oldField, newField reflect.Value) bool {
	kind := oldField.Kind()
	// slice/map: 两边都为空（nil 或 length 0）视为等价
	if kind == reflect.Slice || kind == reflect.Map {
		if oldField.Len() == 0 && newField.Len() == 0 {
			return true
		}
	}
	return false
}

func (c *Comparator) compareStructFields(oldVal, newVal reflect.Value, basePath string, nestedChange *diff.NestedStructChange) {
	if oldVal.Kind() != reflect.Struct || newVal.Kind() != reflect.Struct {
		return
	}

	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() || field.Name == "Id" || field.Name == "ID" {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		// 空 slice/map 视为等价（避免 JSON 序列化 [] vs null 产生误报）
		if isEmptyEquivalent(oldField, newField) {
			continue
		}

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   field.Name,
				StructName:  oldType.Name(),
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldField.Interface(),
				NewValue:    newField.Interface(),
				ValueType:   oldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}
}

// compareSkillLines 对比技能台词
func (c *Comparator) compareSkillLines(oldLines, newLines []*herowiki_def.SkillLineInfo, basePath string, fieldChanges *diff.FieldChanges) {
	linesPath := basePath + ".Lines"

	if len(oldLines) == 0 && len(newLines) == 0 {
		return
	}

	// 构建台词映射（使用ID作为key）
	oldMap := make(map[int]*herowiki_def.SkillLineInfo)
	newMap := make(map[int]*herowiki_def.SkillLineInfo)

	for _, line := range oldLines {
		if line != nil {
			oldMap[line.Id] = line
		}
	}

	for _, line := range newLines {
		if line != nil {
			newMap[line.Id] = line
		}
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, linesPath, "[]SkillLineInfo")

	// 处理新增、删除、修改
	for id := range newMap {
		if _, exists := oldMap[id]; !exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			change := &diff.FieldChange{
				FieldPath:   linePath,
				FieldName:   fmt.Sprintf("台词[%d]", id),
				StructName:  "SkillLineInfo",
				NestedLevel: strings.Count(linePath, "."),
				OldValue:    nil,
				NewValue:    "新增台词",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id := range oldMap {
		if _, exists := newMap[id]; !exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			change := &diff.FieldChange{
				FieldPath:   linePath,
				FieldName:   fmt.Sprintf("台词[%d]", id),
				StructName:  "SkillLineInfo",
				NestedLevel: strings.Count(linePath, "."),
				OldValue:    "删除台词",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, newLine := range newMap {
		if oldLine, exists := oldMap[id]; exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, linePath, "SkillLineInfo")
			c.compareStructFields(
				reflect.ValueOf(oldLine).Elem(),
				reflect.ValueOf(newLine).Elem(),
				linePath,
				nestedChange,
			)
		}
	}
}

// compareSkillTags 对比技能标签
func (c *Comparator) compareSkillTags(oldTags, newTags []*herowiki_def.SkillTagInfo, basePath string, fieldChanges *diff.FieldChanges) {
	tagsPath := basePath + ".Tags"

	if len(oldTags) == 0 && len(newTags) == 0 {
		return
	}

	// 构建标签映射（使用SkillTag作为key）
	oldMap := make(map[string]*herowiki_def.SkillTagInfo)
	newMap := make(map[string]*herowiki_def.SkillTagInfo)

	for _, tag := range oldTags {
		if tag != nil {
			oldMap[tag.SkillTag] = tag
		}
	}

	for _, tag := range newTags {
		if tag != nil {
			newMap[tag.SkillTag] = tag
		}
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, tagsPath, "[]SkillTagInfo")

	// 处理新增、删除、修改
	for id, newTag := range newMap {
		if _, exists := oldMap[id]; !exists {
			tagPath := fmt.Sprintf("%s[%s]", tagsPath, id)
			change := &diff.FieldChange{
				FieldPath:   tagPath,
				FieldName:   fmt.Sprintf("标签[%s]", id),
				StructName:  "SkillTagInfo",
				NestedLevel: strings.Count(tagPath, "."),
				OldValue:    nil,
				NewValue:    newTag.TagName,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, oldTag := range oldMap {
		if _, exists := newMap[id]; !exists {
			tagPath := fmt.Sprintf("%s[%s]", tagsPath, id)
			change := &diff.FieldChange{
				FieldPath:   tagPath,
				FieldName:   fmt.Sprintf("标签[%s]", id),
				StructName:  "SkillTagInfo",
				NestedLevel: strings.Count(tagPath, "."),
				OldValue:    oldTag.TagName,
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, newTag := range newMap {
		if oldTag, exists := oldMap[id]; exists {
			tagPath := fmt.Sprintf("%s[%s]", tagsPath, id)
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, tagPath, "SkillTagInfo")
			c.compareStructFields(
				reflect.ValueOf(oldTag).Elem(),
				reflect.ValueOf(newTag).Elem(),
				tagPath,
				nestedChange,
			)
		}
	}
}

// compareHeroSkins 对比武将皮肤
func (c *Comparator) compareHeroSkins(oldSkins, newSkins []*herowiki_def.HeroSkinInfo, fieldChanges *diff.FieldChanges) {
	basePath := fieldChanges.FieldPath + ".Skins"

	// 构建皮肤映射
	oldMap := make(map[int]*herowiki_def.HeroSkinInfo)
	newMap := make(map[int]*herowiki_def.HeroSkinInfo)

	for _, skin := range oldSkins {
		if skin != nil {
			oldMap[skin.ItemId] = skin
		}
	}

	for _, skin := range newSkins {
		if skin != nil {
			newMap[skin.ItemId] = skin
		}
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "[]HeroSkinInfo")

	// 检查新增的皮肤
	for id, newSkin := range newMap {
		if _, exists := oldMap[id]; !exists {
			skinPath := fmt.Sprintf("%s[%d]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   skinPath,
				FieldName:   fmt.Sprintf("皮肤[%d]", id),
				StructName:  "HeroSkinInfo",
				NestedLevel: strings.Count(skinPath, "."),
				OldValue:    nil,
				NewValue:    newSkin.Name,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 检查删除的皮肤
	for id, oldSkin := range oldMap {
		if _, exists := newMap[id]; !exists {
			skinPath := fmt.Sprintf("%s[%d]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   skinPath,
				FieldName:   fmt.Sprintf("皮肤[%d]", id),
				StructName:  "HeroSkinInfo",
				NestedLevel: strings.Count(skinPath, "."),
				OldValue:    oldSkin.Name,
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 检查修改的皮肤
	for id, newSkin := range newMap {
		if oldSkin, exists := oldMap[id]; exists {
			skinPath := fmt.Sprintf("%s[%d]", basePath, id)
			c.compareSingleSkin(oldSkin, newSkin, skinPath, fieldChanges)
		}
	}
}

// compareSingleSkin 对比单个皮肤
func (c *Comparator) compareSingleSkin(oldSkin, newSkin *herowiki_def.HeroSkinInfo, basePath string, fieldChanges *diff.FieldChanges) {
	// 对比皮肤基础字段
	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroSkinInfo")
	c.compareStructFields(reflect.ValueOf(oldSkin).Elem(), reflect.ValueOf(newSkin).Elem(), basePath, nestedChange)

	// 对比Spine信息
	if oldSkin.Spine != nil || newSkin.Spine != nil {
		spinePath := basePath + ".Spine"
		spineChange := c.getOrCreateNestedStruct(fieldChanges, spinePath, "HeroSkinSpineInfo")
		if oldSkin.Spine != nil && newSkin.Spine != nil {
			c.compareStructFields(
				reflect.ValueOf(oldSkin.Spine).Elem(),
				reflect.ValueOf(newSkin.Spine).Elem(),
				spinePath,
				spineChange,
			)
		} else if oldSkin.Spine == nil && newSkin.Spine != nil {
			change := &diff.FieldChange{
				FieldPath:   spinePath,
				FieldName:   "Spine",
				StructName:  "HeroSkinSpineInfo",
				NestedLevel: strings.Count(spinePath, "."),
				OldValue:    nil,
				NewValue:    "新增Spine信息",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			spineChange.Changes = append(spineChange.Changes, change)
			spineChange.FieldCount++
		} else if oldSkin.Spine != nil && newSkin.Spine == nil {
			change := &diff.FieldChange{
				FieldPath:   spinePath,
				FieldName:   "Spine",
				StructName:  "HeroSkinSpineInfo",
				NestedLevel: strings.Count(spinePath, "."),
				OldValue:    "删除Spine信息",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			spineChange.Changes = append(spineChange.Changes, change)
			spineChange.FieldCount++
		}
	}

	// 对比皮肤台词
	c.compareSkinLines(oldSkin.Lines, newSkin.Lines, basePath, fieldChanges)

	// 对比收藏册信息
	if oldSkin.Collection != nil || newSkin.Collection != nil {
		collectionPath := basePath + ".Collection"
		collectionChange := c.getOrCreateNestedStruct(fieldChanges, collectionPath, "HeroSkinCollectionInfo")
		if oldSkin.Collection != nil && newSkin.Collection != nil {
			c.compareStructFields(
				reflect.ValueOf(oldSkin.Collection).Elem(),
				reflect.ValueOf(newSkin.Collection).Elem(),
				collectionPath,
				collectionChange,
			)
		}
	}
}

// compareSkinLines 对比皮肤台词
func (c *Comparator) compareSkinLines(oldLines, newLines []*herowiki_def.HeroLineInfo, basePath string, fieldChanges *diff.FieldChanges) {
	linesPath := basePath + ".Lines"

	if len(oldLines) == 0 && len(newLines) == 0 {
		return
	}

	// 构建台词映射
	oldMap := make(map[int]*herowiki_def.HeroLineInfo)
	newMap := make(map[int]*herowiki_def.HeroLineInfo)

	for _, line := range oldLines {
		if line != nil {
			oldMap[line.Id] = line
		}
	}

	for _, line := range newLines {
		if line != nil {
			newMap[line.Id] = line
		}
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, linesPath, "[]HeroLineInfo")

	// 处理新增、删除、修改
	for id, newLine := range newMap {
		if _, exists := oldMap[id]; !exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			change := &diff.FieldChange{
				FieldPath:   linePath,
				FieldName:   fmt.Sprintf("台词[%d]", id),
				StructName:  "HeroLineInfo",
				NestedLevel: strings.Count(linePath, "."),
				OldValue:    nil,
				NewValue:    newLine.Text,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, oldLine := range oldMap {
		if _, exists := newMap[id]; !exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			change := &diff.FieldChange{
				FieldPath:   linePath,
				FieldName:   fmt.Sprintf("台词[%d]", id),
				StructName:  "HeroLineInfo",
				NestedLevel: strings.Count(linePath, "."),
				OldValue:    oldLine.Text,
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, newLine := range newMap {
		if oldLine, exists := oldMap[id]; exists {
			linePath := fmt.Sprintf("%s[%d]", linesPath, id)
			nestedChange := c.getOrCreateNestedStruct(fieldChanges, linePath, "HeroLineInfo")
			c.compareStructFields(
				reflect.ValueOf(oldLine).Elem(),
				reflect.ValueOf(newLine).Elem(),
				linePath,
				nestedChange,
			)
		}
	}
}

// compareHeroAchievements 对比武将成就
func (c *Comparator) compareHeroAchievements(oldAchieves, newAchieves []*herowiki_def.HeroAchievementInfo, fieldChanges *diff.FieldChanges) {
	basePath := fieldChanges.FieldPath + ".Achievements"

	if len(oldAchieves) == 0 && len(newAchieves) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "[]HeroAchievementInfo")

	// 构建成就映射（使用成就ID作为key）
	oldMap := make(map[string]*herowiki_def.HeroAchievementInfo)
	newMap := make(map[string]*herowiki_def.HeroAchievementInfo)

	for i, achieve := range oldAchieves {
		if achieve != nil {
			if achieve.HeroAchieve != nil {
				oldMap[achieve.HeroAchieve.Id] = achieve
			} else {
				oldMap[fmt.Sprintf("index_%d", i)] = achieve
			}
		}
	}

	for i, achieve := range newAchieves {
		if achieve != nil {
			if achieve.HeroAchieve != nil {
				newMap[achieve.HeroAchieve.Id] = achieve
			} else {
				newMap[fmt.Sprintf("index_%d", i)] = achieve
			}
		}
	}

	// 处理新增、删除、修改
	for id := range newMap {
		if _, exists := oldMap[id]; !exists {
			achievePath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   achievePath,
				FieldName:   fmt.Sprintf("成就[%s]", id),
				StructName:  "HeroAchievementInfo",
				NestedLevel: strings.Count(achievePath, "."),
				OldValue:    nil,
				NewValue:    "新增成就",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id := range oldMap {
		if _, exists := newMap[id]; !exists {
			achievePath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   achievePath,
				FieldName:   fmt.Sprintf("成就[%s]", id),
				StructName:  "HeroAchievementInfo",
				NestedLevel: strings.Count(achievePath, "."),
				OldValue:    "删除成就",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	for id, newAchieve := range newMap {
		if oldAchieve, exists := oldMap[id]; exists {
			achievePath := fmt.Sprintf("%s[%s]", basePath, id)
			c.compareSingleAchievement(oldAchieve, newAchieve, achievePath, fieldChanges)
		}
	}
}

// compareSingleAchievement 对比单个成就
func (c *Comparator) compareSingleAchievement(oldAchieve, newAchieve *herowiki_def.HeroAchievementInfo, basePath string, fieldChanges *diff.FieldChanges) {
	// 对比英雄成就
	if oldAchieve.HeroAchieve != nil || newAchieve.HeroAchieve != nil {
		heroPath := basePath + ".HeroAchieve"
		heroChange := c.getOrCreateNestedStruct(fieldChanges, heroPath, "HeroAchieveInfo")
		if oldAchieve.HeroAchieve != nil && newAchieve.HeroAchieve != nil {
			c.compareStructFields(
				reflect.ValueOf(oldAchieve.HeroAchieve).Elem(),
				reflect.ValueOf(newAchieve.HeroAchieve).Elem(),
				heroPath,
				heroChange,
			)
		}
	}

	// 对比成就详情
	if oldAchieve.AchieveDetail != nil || newAchieve.AchieveDetail != nil {
		detailPath := basePath + ".AchieveDetail"
		detailChange := c.getOrCreateNestedStruct(fieldChanges, detailPath, "AchieveDetailInfo")
		if oldAchieve.AchieveDetail != nil && newAchieve.AchieveDetail != nil {
			c.compareStructFields(
				reflect.ValueOf(oldAchieve.AchieveDetail).Elem(),
				reflect.ValueOf(newAchieve.AchieveDetail).Elem(),
				detailPath,
				detailChange,
			)

			// 对比任务完成条件
			if oldAchieve.AchieveDetail.ConditionInfo.Id != 0 || newAchieve.AchieveDetail.ConditionInfo.Id != 0 {
				condPath := detailPath + ".ConditionInfo"
				condChange := c.getOrCreateNestedStruct(fieldChanges, condPath, "TaskCompleteConditionInfo")
				c.compareStructFields(
					reflect.ValueOf(oldAchieve.AchieveDetail.ConditionInfo).Elem(),
					reflect.ValueOf(newAchieve.AchieveDetail.ConditionInfo).Elem(),
					condPath,
					condChange,
				)
			}
		}
	}
}

// compareCountry 对比国家信息
func (c *Comparator) compareCountry(oldCountry, newCountry *herowiki_def.CountryInfo, fieldChanges *diff.FieldChanges) {
	if oldCountry == nil && newCountry == nil {
		return
	}

	basePath := fieldChanges.FieldPath + ".Country"

	if oldCountry == nil && newCountry != nil {
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "Country",
			StructName:  "CountryInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    nil,
			NewValue:    newCountry.Name,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeAdded,
		}
		fieldChanges.Changes = append(fieldChanges.Changes, change)
		return
	}

	if oldCountry != nil && newCountry == nil {
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "Country",
			StructName:  "CountryInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    oldCountry.Name,
			NewValue:    nil,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeRemoved,
		}
		fieldChanges.Changes = append(fieldChanges.Changes, change)
		return
	}

	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "CountryInfo")
	c.compareStructFields(reflect.ValueOf(oldCountry).Elem(), reflect.ValueOf(newCountry).Elem(), basePath, nestedChange)
}

// compareRecommendBd 对比推荐布阵
func (c *Comparator) compareRecommendBd(oldBd, newBd *herowiki_def.RecommendBdInfo, fieldChanges *diff.FieldChanges) {
	if oldBd == nil && newBd == nil {
		return
	}

	basePath := fieldChanges.FieldPath + ".RecommendBd"

	if oldBd == nil && newBd != nil {
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "RecommendBd",
			StructName:  "RecommendBdInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    nil,
			NewValue:    "新增推荐布阵",
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeAdded,
		}
		fieldChanges.Changes = append(fieldChanges.Changes, change)
		return
	}

	if oldBd != nil && newBd == nil {
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "RecommendBd",
			StructName:  "RecommendBdInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    "删除推荐布阵",
			NewValue:    nil,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeRemoved,
		}
		fieldChanges.Changes = append(fieldChanges.Changes, change)
		return
	}

	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "RecommendBdInfo")
	c.compareStructFields(reflect.ValueOf(oldBd).Elem(), reflect.ValueOf(newBd).Elem(), basePath, nestedChange)
}

// compareHeroRobotActions 对比机器人行为
func (c *Comparator) compareHeroRobotActions(oldActions, newActions []*herowiki_def.RobotActionInfo, fieldChanges *diff.FieldChanges) {
	basePath := fieldChanges.FieldPath + ".RobotActions"

	if len(oldActions) == 0 && len(newActions) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "[]RobotActionInfo")

	// 构建行为映射（使用ID作为key）
	oldMap := make(map[string]*herowiki_def.RobotActionInfo)
	newMap := make(map[string]*herowiki_def.RobotActionInfo)

	for i, action := range oldActions {
		if action != nil {
			if action.Id != "" {
				oldMap[action.Id] = action
			} else {
				oldMap[fmt.Sprintf("index_%d", i)] = action
			}
		}
	}

	for i, action := range newActions {
		if action != nil {
			if action.Id != "" {
				newMap[action.Id] = action
			} else {
				newMap[fmt.Sprintf("index_%d", i)] = action
			}
		}
	}

	// 处理新增的行为
	for id := range newMap {
		if _, exists := oldMap[id]; !exists {
			actionPath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   actionPath,
				FieldName:   fmt.Sprintf("行为[%s]", id),
				StructName:  "RobotActionInfo",
				NestedLevel: strings.Count(actionPath, "."),
				OldValue:    nil,
				NewValue:    "新增机器人行为",
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 处理删除的行为
	for id := range oldMap {
		if _, exists := newMap[id]; !exists {
			actionPath := fmt.Sprintf("%s[%s]", basePath, id)
			change := &diff.FieldChange{
				FieldPath:   actionPath,
				FieldName:   fmt.Sprintf("行为[%s]", id),
				StructName:  "RobotActionInfo",
				NestedLevel: strings.Count(actionPath, "."),
				OldValue:    "删除机器人行为",
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
		}
	}

	// 处理修改的行为
	for id, newAction := range newMap {
		if oldAction, exists := oldMap[id]; exists {
			actionPath := fmt.Sprintf("%s[%s]", basePath, id)
			c.compareSingleRobotAction(oldAction, newAction, actionPath, fieldChanges)
		}
	}
}

// compareSingleRobotAction 对比单个机器人行为
func (c *Comparator) compareSingleRobotAction(oldAction, newAction *herowiki_def.RobotActionInfo, basePath string, fieldChanges *diff.FieldChanges) {
	if oldAction == nil || newAction == nil {
		return
	}

	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "RobotActionInfo")

	// 使用反射对比所有字段
	oldVal := reflect.ValueOf(oldAction).Elem()
	newVal := reflect.ValueOf(newAction).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		// 特殊处理切片类型
		if oldField.Kind() == reflect.Slice || oldField.Kind() == reflect.Array {
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				oldSlice := oldField.Interface()
				newSlice := newField.Interface()

				// 格式化切片显示
				oldStr := fmt.Sprintf("%v", oldSlice)
				newStr := fmt.Sprintf("%v", newSlice)

				if oldStr != newStr {
					change := &diff.FieldChange{
						FieldPath:   fieldPath,
						FieldName:   field.Name,
						StructName:  "RobotActionInfo",
						NestedLevel: strings.Count(fieldPath, "."),
						OldValue:    oldSlice,
						NewValue:    newSlice,
						ValueType:   "slice",
						ChangeType:  diff.ChangeTypeModified,
					}
					nestedChange.Changes = append(nestedChange.Changes, change)
					nestedChange.FieldCount++
				}
			}
		} else {
			// 基本类型字段
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				change := &diff.FieldChange{
					FieldPath:   fieldPath,
					FieldName:   field.Name,
					StructName:  "RobotActionInfo",
					NestedLevel: strings.Count(fieldPath, "."),
					OldValue:    oldField.Interface(),
					NewValue:    newField.Interface(),
					ValueType:   oldField.Kind().String(),
					ChangeType:  diff.ChangeTypeModified,
				}
				nestedChange.Changes = append(nestedChange.Changes, change)
				nestedChange.FieldCount++
			}
		}
	}
}

// compareHeroDropInfos 对比武将掉落信息
func (c *Comparator) compareHeroDropInfos(oldDropInfo, newDropInfo *herowiki_def.HeroDropInfo, fieldChanges *diff.FieldChanges) {
	if oldDropInfo == nil && newDropInfo == nil {
		return
	}

	basePath := fieldChanges.FieldPath + ".DropInfo"

	// 处理新增或删除的情况
	if oldDropInfo == nil && newDropInfo != nil {
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroDropInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "DropInfo",
			StructName:  "HeroDropInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    nil,
			NewValue:    "新增掉落信息",
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeAdded,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	if oldDropInfo != nil && newDropInfo == nil {
		nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroDropInfo")
		change := &diff.FieldChange{
			FieldPath:   basePath,
			FieldName:   "DropInfo",
			StructName:  "HeroDropInfo",
			NestedLevel: strings.Count(basePath, "."),
			OldValue:    "删除掉落信息",
			NewValue:    nil,
			ValueType:   "struct",
			ChangeType:  diff.ChangeTypeRemoved,
		}
		nestedChange.Changes = append(nestedChange.Changes, change)
		nestedChange.FieldCount++
		return
	}

	// 两者都存在，对比各个字段
	nestedChange := c.getOrCreateNestedStruct(fieldChanges, basePath, "HeroDropInfo")

	// 对比直接掉落规则
	c.compareDropRuleSlice(
		oldDropInfo.DirectDropRules,
		newDropInfo.DirectDropRules,
		basePath+".DirectDropRules",
		"DirectDropRules",
		nestedChange,
		fieldChanges,
	)

	// 对比保底掉落规则
	c.compareDropRuleSlice(
		oldDropInfo.GuaranteeDropRules,
		newDropInfo.GuaranteeDropRules,
		basePath+".GuaranteeDropRules",
		"GuaranteeDropRules",
		nestedChange,
		fieldChanges,
	)

	// 对比掉落组
	c.compareDropGroupSlice(
		oldDropInfo.DropGroups,
		newDropInfo.DropGroups,
		basePath+".DropGroups",
		"DropGroups",
		nestedChange,
		fieldChanges,
	)

	// 对比按类型分类的掉落信息
	c.compareDropTypeMap(
		oldDropInfo.ByDropType,
		newDropInfo.ByDropType,
		basePath+".ByDropType",
		"ByDropType",
		nestedChange,
		fieldChanges,
	)

	// 如果没有任何变化，清理空的嵌套结构
	if len(nestedChange.Changes) == 0 && nestedChange.FieldCount == 0 {
		delete(fieldChanges.NestedStructs, basePath)
	}
}

// compareDropRuleSlice 对比掉落规则切片 - 修复版
func (c *Comparator) compareDropRuleSlice(oldRules, newRules []*herowiki_def.DropRuleSummary, basePath, fieldName string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if len(oldRules) == 0 && len(newRules) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "[]DropRuleSummary")

	// 构建规则映射 - 使用RuleId作为key，如果没有ID则使用索引
	oldMap := make(map[int]*herowiki_def.DropRuleSummary)
	newMap := make(map[int]*herowiki_def.DropRuleSummary)

	for i, rule := range oldRules {
		if rule != nil {
			// 如果有RuleId，使用它；否则使用索引+偏移量避免冲突
			if rule.RuleId > 0 {
				oldMap[rule.RuleId] = rule
			} else {
				oldMap[1000000+i] = rule // 使用大偏移量避免与真实ID冲突
			}
		}
	}

	for i, rule := range newRules {
		if rule != nil {
			if rule.RuleId > 0 {
				newMap[rule.RuleId] = rule
			} else {
				newMap[1000000+i] = rule
			}
		}
	}

	// 处理新增的规则
	for id, newRule := range newMap {
		if _, exists := oldMap[id]; !exists {
			rulePath := fmt.Sprintf("%s[%d]", basePath, id)

			// 创建变化记录
			change := &diff.FieldChange{
				FieldPath:   rulePath,
				FieldName:   fmt.Sprintf("规则[%d]", id),
				StructName:  "DropRuleSummary",
				NestedLevel: strings.Count(rulePath, "."),
				OldValue:    nil,
				NewValue:    c.formatDropRuleSummary(newRule),
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}

			// 添加到sliceChange
			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++

			// 同时添加到fieldChanges.Changes以便上层能看到
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理删除的规则
	for id, oldRule := range oldMap {
		if _, exists := newMap[id]; !exists {
			rulePath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   rulePath,
				FieldName:   fmt.Sprintf("规则[%d]", id),
				StructName:  "DropRuleSummary",
				NestedLevel: strings.Count(rulePath, "."),
				OldValue:    c.formatDropRuleSummary(oldRule),
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理修改的规则
	for id, newRule := range newMap {
		if oldRule, exists := oldMap[id]; exists && oldRule != nil && newRule != nil {
			rulePath := fmt.Sprintf("%s[%d]", basePath, id)

			// 创建规则自己的嵌套结构
			ruleChange := c.getOrCreateNestedStruct(fieldChanges, rulePath, "DropRuleSummary")

			// 对比规则的所有字段
			oldVal := reflect.ValueOf(oldRule).Elem()
			newVal := reflect.ValueOf(newRule).Elem()
			oldType := oldVal.Type()

			for i := 0; i < oldVal.NumField(); i++ {
				field := oldType.Field(i)
				if !field.IsExported() {
					continue
				}

				oldField := oldVal.Field(i)
				newField := newVal.Field(i)
				fieldPath := rulePath + "." + field.Name

				if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
					change := &diff.FieldChange{
						FieldPath:   fieldPath,
						FieldName:   field.Name,
						StructName:  "DropRuleSummary",
						NestedLevel: strings.Count(fieldPath, "."),
						OldValue:    oldField.Interface(),
						NewValue:    newField.Interface(),
						ValueType:   oldField.Kind().String(),
						ChangeType:  diff.ChangeTypeModified,
					}

					ruleChange.Changes = append(ruleChange.Changes, change)
					ruleChange.FieldCount++

					// 同时添加到sliceChange和fieldChanges
					sliceChange.Changes = append(sliceChange.Changes, change)
					fieldChanges.Changes = append(fieldChanges.Changes, change)
				}
			}

			// 如果规则有变化，确保它被记录
			if ruleChange.FieldCount > 0 {
				// 已经在上面添加了变化
			}
		}
	}

	// 如果sliceChange有变化，确保它被保留
	if sliceChange.FieldCount == 0 {
		delete(parentChange.Children, basePath)
	}
}

// compareDropGroupSlice 对比掉落组切片 - 修复版
func (c *Comparator) compareDropGroupSlice(oldGroups, newGroups []*herowiki_def.DropGroupSummary, basePath, fieldName string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if len(oldGroups) == 0 && len(newGroups) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "[]DropGroupSummary")

	// 构建组映射
	oldMap := make(map[int]*herowiki_def.DropGroupSummary)
	newMap := make(map[int]*herowiki_def.DropGroupSummary)

	for i, group := range oldGroups {
		if group != nil {
			if group.GroupId > 0 {
				oldMap[group.GroupId] = group
			} else {
				oldMap[1000000+i] = group
			}
		}
	}

	for i, group := range newGroups {
		if group != nil {
			if group.GroupId > 0 {
				newMap[group.GroupId] = group
			} else {
				newMap[1000000+i] = group
			}
		}
	}

	// 处理新增的组
	for id, newGroup := range newMap {
		if _, exists := oldMap[id]; !exists {
			groupPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   groupPath,
				FieldName:   fmt.Sprintf("掉落组[%d]", id),
				StructName:  "DropGroupSummary",
				NestedLevel: strings.Count(groupPath, "."),
				OldValue:    nil,
				NewValue:    c.formatDropGroupSummary(newGroup),
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理删除的组
	for id, oldGroup := range oldMap {
		if _, exists := newMap[id]; !exists {
			groupPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   groupPath,
				FieldName:   fmt.Sprintf("掉落组[%d]", id),
				StructName:  "DropGroupSummary",
				NestedLevel: strings.Count(groupPath, "."),
				OldValue:    c.formatDropGroupSummary(oldGroup),
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理修改的组
	for id, newGroup := range newMap {
		if oldGroup, exists := oldMap[id]; exists && oldGroup != nil && newGroup != nil {
			groupPath := fmt.Sprintf("%s[%d]", basePath, id)

			// 创建组的嵌套结构
			groupChange := c.getOrCreateNestedStruct(fieldChanges, groupPath, "DropGroupSummary")

			// 对比组的字段
			oldVal := reflect.ValueOf(oldGroup).Elem()
			newVal := reflect.ValueOf(newGroup).Elem()
			oldType := oldVal.Type()

			for i := 0; i < oldVal.NumField(); i++ {
				field := oldType.Field(i)
				if !field.IsExported() {
					continue
				}

				oldField := oldVal.Field(i)
				newField := newVal.Field(i)
				fieldPath := groupPath + "." + field.Name

				// 特殊处理 DropItems 切片
				if field.Name == "DropItems" {
					c.compareHeroDropItemSlice(oldGroup.DropItems, newGroup.DropItems, fieldPath, "DropItems", groupChange, fieldChanges)
					continue
				}

				if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
					change := &diff.FieldChange{
						FieldPath:   fieldPath,
						FieldName:   field.Name,
						StructName:  "DropGroupSummary",
						NestedLevel: strings.Count(fieldPath, "."),
						OldValue:    oldField.Interface(),
						NewValue:    newField.Interface(),
						ValueType:   oldField.Kind().String(),
						ChangeType:  diff.ChangeTypeModified,
					}

					groupChange.Changes = append(groupChange.Changes, change)
					groupChange.FieldCount++
					sliceChange.Changes = append(sliceChange.Changes, change)
					fieldChanges.Changes = append(fieldChanges.Changes, change)
				}
			}
		}
	}
}

// compareDropTypeMap 对比按类型分类的掉落信息
func (c *Comparator) compareDropTypeMap(oldMap, newMap map[string]*herowiki_def.DropTypeInfo, basePath, fieldName string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if len(oldMap) == 0 && len(newMap) == 0 {
		return
	}

	mapChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "map[string]DropTypeInfo")

	// 检查所有key
	allKeys := make(map[string]bool)
	for k := range oldMap {
		allKeys[k] = true
	}
	for k := range newMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		oldTypeInfo := oldMap[key]
		newTypeInfo := newMap[key]
		typePath := basePath + "[" + key + "]"

		if oldTypeInfo == nil && newTypeInfo != nil {
			// 新增
			change := &diff.FieldChange{
				FieldPath:   typePath,
				FieldName:   key,
				StructName:  "DropTypeInfo",
				NestedLevel: strings.Count(typePath, "."),
				OldValue:    nil,
				NewValue:    c.formatDropTypeInfo(newTypeInfo),
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}
			mapChange.Changes = append(mapChange.Changes, change)
			mapChange.FieldCount++
		} else if oldTypeInfo != nil && newTypeInfo == nil {
			// 删除
			change := &diff.FieldChange{
				FieldPath:   typePath,
				FieldName:   key,
				StructName:  "DropTypeInfo",
				NestedLevel: strings.Count(typePath, "."),
				OldValue:    c.formatDropTypeInfo(oldTypeInfo),
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}
			mapChange.Changes = append(mapChange.Changes, change)
			mapChange.FieldCount++
		} else if oldTypeInfo != nil && newTypeInfo != nil {
			// 修改
			c.compareSingleDropTypeInfo(oldTypeInfo, newTypeInfo, typePath, parentChange, fieldChanges)
		}
	}
}

// compareSingleDropRule 对比单个掉落规则
func (c *Comparator) compareSingleDropRule(oldRule, newRule *herowiki_def.DropRuleSummary, basePath string, parentChange *diff.NestedStructChange) {
	if oldRule == nil || newRule == nil {
		return
	}

	nestedChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "DropRuleSummary")

	oldVal := reflect.ValueOf(oldRule).Elem()
	newVal := reflect.ValueOf(newRule).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		// 特殊处理切片类型
		if oldField.Kind() == reflect.Slice {
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				change := &diff.FieldChange{
					FieldPath:   fieldPath,
					FieldName:   field.Name,
					StructName:  "DropRuleSummary",
					NestedLevel: strings.Count(fieldPath, "."),
					OldValue:    fmt.Sprintf("%v", oldField.Interface()),
					NewValue:    fmt.Sprintf("%v", newField.Interface()),
					ValueType:   "slice",
					ChangeType:  diff.ChangeTypeModified,
				}
				nestedChange.Changes = append(nestedChange.Changes, change)
				nestedChange.FieldCount++
			}
		} else {
			if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
				change := &diff.FieldChange{
					FieldPath:   fieldPath,
					FieldName:   field.Name,
					StructName:  "DropRuleSummary",
					NestedLevel: strings.Count(fieldPath, "."),
					OldValue:    oldField.Interface(),
					NewValue:    newField.Interface(),
					ValueType:   oldField.Kind().String(),
					ChangeType:  diff.ChangeTypeModified,
				}
				nestedChange.Changes = append(nestedChange.Changes, change)
				nestedChange.FieldCount++
			}
		}
	}
}

// compareSingleDropGroup 对比单个掉落组
func (c *Comparator) compareSingleDropGroup(oldGroup, newGroup *herowiki_def.DropGroupSummary, basePath string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if oldGroup == nil || newGroup == nil {
		return
	}

	nestedChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "DropGroupSummary")

	oldVal := reflect.ValueOf(oldGroup).Elem()
	newVal := reflect.ValueOf(newGroup).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		// 特殊处理 DropItems 切片
		if field.Name == "DropItems" {
			c.compareHeroDropItemSlice(oldGroup.DropItems, newGroup.DropItems, fieldPath, "DropItems", parentChange, fieldChanges)
			continue
		}

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   field.Name,
				StructName:  "DropGroupSummary",
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldField.Interface(),
				NewValue:    newField.Interface(),
				ValueType:   oldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}
}

// compareHeroDropItemSlice 对比英雄掉落项切片 - 修复版
func (c *Comparator) compareHeroDropItemSlice(oldItems, newItems []*herowiki_def.HeroDropItem, basePath, fieldName string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if len(oldItems) == 0 && len(newItems) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "[]HeroDropItem")

	// 构建物品映射
	oldMap := make(map[int]*herowiki_def.HeroDropItem)
	newMap := make(map[int]*herowiki_def.HeroDropItem)

	for i, item := range oldItems {
		if item != nil {
			if item.ItemId > 0 {
				oldMap[item.ItemId] = item
			} else {
				oldMap[1000000+i] = item
			}
		}
	}

	for i, item := range newItems {
		if item != nil {
			if item.ItemId > 0 {
				newMap[item.ItemId] = item
			} else {
				newMap[1000000+i] = item
			}
		}
	}

	// 处理新增的物品
	for id, newItem := range newMap {
		if _, exists := oldMap[id]; !exists {
			itemPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   itemPath,
				FieldName:   fmt.Sprintf("掉落项[%d]", id),
				StructName:  "HeroDropItem",
				NestedLevel: strings.Count(itemPath, "."),
				OldValue:    nil,
				NewValue:    c.formatHeroDropItem(newItem),
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理删除的物品
	for id, oldItem := range oldMap {
		if _, exists := newMap[id]; !exists {
			itemPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   itemPath,
				FieldName:   fmt.Sprintf("掉落项[%d]", id),
				StructName:  "HeroDropItem",
				NestedLevel: strings.Count(itemPath, "."),
				OldValue:    c.formatHeroDropItem(oldItem),
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理修改的物品
	for id, newItem := range newMap {
		if oldItem, exists := oldMap[id]; exists && oldItem != nil && newItem != nil {
			itemPath := fmt.Sprintf("%s[%d]", basePath, id)

			// 创建物品的嵌套结构
			itemChange := c.getOrCreateNestedStruct(fieldChanges, itemPath, "HeroDropItem")

			oldVal := reflect.ValueOf(oldItem).Elem()
			newVal := reflect.ValueOf(newItem).Elem()
			oldType := oldVal.Type()

			for i := 0; i < oldVal.NumField(); i++ {
				field := oldType.Field(i)
				if !field.IsExported() {
					continue
				}

				oldField := oldVal.Field(i)
				newField := newVal.Field(i)
				fieldPath := itemPath + "." + field.Name

				// 特殊处理 ItemConfigs 切片
				if field.Name == "ItemConfigs" {
					c.compareItemConfigSlice(oldItem.ItemConfigs, newItem.ItemConfigs, fieldPath, "ItemConfigs", itemChange, fieldChanges)
					continue
				}

				if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
					change := &diff.FieldChange{
						FieldPath:   fieldPath,
						FieldName:   field.Name,
						StructName:  "HeroDropItem",
						NestedLevel: strings.Count(fieldPath, "."),
						OldValue:    oldField.Interface(),
						NewValue:    newField.Interface(),
						ValueType:   oldField.Kind().String(),
						ChangeType:  diff.ChangeTypeModified,
					}

					itemChange.Changes = append(itemChange.Changes, change)
					itemChange.FieldCount++
					sliceChange.Changes = append(sliceChange.Changes, change)
					fieldChanges.Changes = append(fieldChanges.Changes, change)
				}
			}
		}
	}
}

// compareItemConfigSlice 对比物品配置切片 - 修复版
func (c *Comparator) compareItemConfigSlice(oldConfigs, newConfigs []*herowiki_def.ItemConfig, basePath, fieldName string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if len(oldConfigs) == 0 && len(newConfigs) == 0 {
		return
	}

	sliceChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "[]ItemConfig")

	// 构建配置映射
	oldMap := make(map[int]*herowiki_def.ItemConfig)
	newMap := make(map[int]*herowiki_def.ItemConfig)

	for i, config := range oldConfigs {
		if config != nil {
			if config.ItemId > 0 {
				oldMap[config.ItemId] = config
			} else {
				oldMap[1000000+i] = config
			}
		}
	}

	for i, config := range newConfigs {
		if config != nil {
			if config.ItemId > 0 {
				newMap[config.ItemId] = config
			} else {
				newMap[1000000+i] = config
			}
		}
	}

	// 处理新增的配置
	for id, newConfig := range newMap {
		if _, exists := oldMap[id]; !exists {
			configPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   configPath,
				FieldName:   fmt.Sprintf("物品[%d]", id),
				StructName:  "ItemConfig",
				NestedLevel: strings.Count(configPath, "."),
				OldValue:    nil,
				NewValue:    c.formatItemConfig(newConfig),
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeAdded,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理删除的配置
	for id, oldConfig := range oldMap {
		if _, exists := newMap[id]; !exists {
			configPath := fmt.Sprintf("%s[%d]", basePath, id)

			change := &diff.FieldChange{
				FieldPath:   configPath,
				FieldName:   fmt.Sprintf("物品[%d]", id),
				StructName:  "ItemConfig",
				NestedLevel: strings.Count(configPath, "."),
				OldValue:    c.formatItemConfig(oldConfig),
				NewValue:    nil,
				ValueType:   "struct",
				ChangeType:  diff.ChangeTypeRemoved,
			}

			sliceChange.Changes = append(sliceChange.Changes, change)
			sliceChange.FieldCount++
			fieldChanges.Changes = append(fieldChanges.Changes, change)
		}
	}

	// 处理修改的配置
	for id, newConfig := range newMap {
		if oldConfig, exists := oldMap[id]; exists && oldConfig != nil && newConfig != nil {
			configPath := fmt.Sprintf("%s[%d]", basePath, id)

			// 创建配置的嵌套结构
			configChange := c.getOrCreateNestedStruct(fieldChanges, configPath, "ItemConfig")

			oldVal := reflect.ValueOf(oldConfig).Elem()
			newVal := reflect.ValueOf(newConfig).Elem()
			oldType := oldVal.Type()

			for i := 0; i < oldVal.NumField(); i++ {
				field := oldType.Field(i)
				if !field.IsExported() {
					continue
				}

				oldField := oldVal.Field(i)
				newField := newVal.Field(i)
				fieldPath := configPath + "." + field.Name

				if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
					change := &diff.FieldChange{
						FieldPath:   fieldPath,
						FieldName:   field.Name,
						StructName:  "ItemConfig",
						NestedLevel: strings.Count(fieldPath, "."),
						OldValue:    oldField.Interface(),
						NewValue:    newField.Interface(),
						ValueType:   oldField.Kind().String(),
						ChangeType:  diff.ChangeTypeModified,
					}

					configChange.Changes = append(configChange.Changes, change)
					configChange.FieldCount++
					sliceChange.Changes = append(sliceChange.Changes, change)
					fieldChanges.Changes = append(fieldChanges.Changes, change)
				}
			}
		}
	}
}

// compareSingleDropTypeInfo 对比单个掉落类型信息
func (c *Comparator) compareSingleDropTypeInfo(oldInfo, newInfo *herowiki_def.DropTypeInfo, basePath string, parentChange *diff.NestedStructChange, fieldChanges *diff.FieldChanges) {
	if oldInfo == nil || newInfo == nil {
		return
	}

	nestedChange := c.getOrCreateNestedStructForParent(parentChange, basePath, "DropTypeInfo")

	oldVal := reflect.ValueOf(oldInfo).Elem()
	newVal := reflect.ValueOf(newInfo).Elem()
	oldType := oldVal.Type()

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		if !field.IsExported() {
			continue
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)
		fieldPath := basePath + "." + field.Name

		// 特殊处理 DropRules 切片
		if field.Name == "DropRules" {
			c.compareDropRuleSlice(oldInfo.DropRules, newInfo.DropRules, fieldPath, "DropRules", parentChange, fieldChanges)
			continue
		}

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			change := &diff.FieldChange{
				FieldPath:   fieldPath,
				FieldName:   field.Name,
				StructName:  "DropTypeInfo",
				NestedLevel: strings.Count(fieldPath, "."),
				OldValue:    oldField.Interface(),
				NewValue:    newField.Interface(),
				ValueType:   oldField.Kind().String(),
				ChangeType:  diff.ChangeTypeModified,
			}
			nestedChange.Changes = append(nestedChange.Changes, change)
			nestedChange.FieldCount++
		}
	}
}

// 辅助函数：为父嵌套结构创建子嵌套结构
func (c *Comparator) getOrCreateNestedStructForParent(parent *diff.NestedStructChange, path, structType string) *diff.NestedStructChange {
	if parent.Children == nil {
		parent.Children = make(map[string]*diff.NestedStructChange)
	}

	if _, exists := parent.Children[path]; !exists {
		parent.Children[path] = &diff.NestedStructChange{
			StructPath: path,
			StructType: structType,
			Changes:    make([]*diff.FieldChange, 0),
			Children:   make(map[string]*diff.NestedStructChange),
		}
	}

	return parent.Children[path]
}

// 格式化辅助函数
func (c *Comparator) formatDropRuleSummary(rule *herowiki_def.DropRuleSummary) string {
	if rule == nil {
		return "nil"
	}
	return fmt.Sprintf("规则[%d]: %s (次数:%d)", rule.RuleId, rule.RuleName, rule.DropCount)
}

func (c *Comparator) formatDropGroupSummary(group *herowiki_def.DropGroupSummary) string {
	if group == nil {
		return "nil"
	}
	return fmt.Sprintf("组[%d]: %s (权重:%d)", group.GroupId, group.GroupName, group.Weight)
}

func (c *Comparator) formatDropTypeInfo(info *herowiki_def.DropTypeInfo) string {
	if info == nil {
		return "nil"
	}
	return fmt.Sprintf("%s (%d个规则)", info.TypeName, info.TotalCount)
}

func (c *Comparator) formatHeroDropItem(item *herowiki_def.HeroDropItem) string {
	if item == nil {
		return "nil"
	}
	return fmt.Sprintf("掉落项[%d]: %s (组:%d)", item.ItemId, item.ItemName, item.DropGroupId)
}

func (c *Comparator) formatItemConfig(config *herowiki_def.ItemConfig) string {
	if config == nil {
		return "nil"
	}
	if config.IsHero {
		return fmt.Sprintf("武将[%s]: %d个", config.HeroId, config.Count)
	}
	return fmt.Sprintf("物品[%d]: %d个", config.ItemId, config.Count)
}
