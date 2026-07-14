// Package check_manager 提供规则过滤功能
// 本包根据变更文件列表过滤需要执行的校验规则
package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule/chain_reference"
	"github.com/xuri/excelize/v2"
)

// FilterRulesByChangedFiles 根据变更文件过滤规则列表
// 返回需要执行的规则子集
//
// 参数:
//   - rules: 所有规则列表
//   - changedFiles: 变更文件列表（相对路径或绝对路径）
//
// 返回:
//   - 需要执行的规则子集
//
// 过滤逻辑:
//  1. 如果 changedFiles 为空，返回全部规则（全量检查）
//  2. 列级规则：通过 rule.Sheet 匹配变更文件
//  3. 表级规则：
//     - 通用规则（NEW_ROW_NOTIFY、ROW_CHANGE_NOTIFY）始终执行
//     - 其他规则：通过 Meta().TargetSheets 匹配变更文件
func FilterRulesByChangedFiles(rules []*json_rule.SheetRule, changedFiles []string) []*json_rule.SheetRule {
	if len(changedFiles) == 0 {
		return rules // 无变更文件限制，返回全部规则
	}

	var filtered []*json_rule.SheetRule
	for _, rule := range rules {
		if shouldRunRule(rule, changedFiles) {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// shouldRunRule 判断规则是否应该执行
func shouldRunRule(rule *json_rule.SheetRule, changedFiles []string) bool {
	// 1. 检查列级规则（通过 rule.Sheet 绑定）
	if len(rule.Rules) > 0 {
		if isSheetInChangedFiles(rule.Sheet, changedFiles) {
			return true
		}
	}

	// 2. 检查表级规则（通过 Meta().TargetSheets 绑定）
	for _, tableRule := range rule.TableRules {
		if !tableRule.Enabled {
			continue
		}

		// 通用规则始终执行
		if isGeneralRule(tableRule.Type) {
			return true
		}

		// 从 TableRuleMetas 获取元数据（代码注册）
		meta := json_rule.TableRuleMetas[tableRule.Type]
		if meta == nil {
			continue
		}

		// 检查 TargetSheets 是否与变更文件匹配
		for _, targetSheet := range meta.TargetSheets {
			if isSheetInChangedFiles(targetSheet, changedFiles) {
				return true
			}
		}
	}

	return false
}

// isSheetInChangedFiles 检查表名是否在变更文件列表中
// 支持精确匹配和后缀匹配（如 "武将|Hero" 可通过 "Hero" 匹配）
// 如果 sheetMap 不为空，还会检查表所在的文件是否在变更列表中
func isSheetInChangedFiles(sheetName string, changedFiles []string) bool {
	for _, file := range changedFiles {
		// 提取文件名（不含路径和扩展名）
		fileName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(fileName, ".xlsx")

		// 精确匹配: "Hero" == "Hero"
		if sheetName == nameWithoutExt {
			return true
		}

		// 后缀匹配: "武将|Hero" 后缀匹配 "Hero"
		if strings.HasSuffix(sheetName, "|"+nameWithoutExt) {
			return true
		}
	}
	return false
}

// isGeneralRule 判断是否为通用规则（始终执行）
func isGeneralRule(ruleType json_rule.ETableRule) bool {
	return ruleType == json_rule.NEW_ROW_NOTIFY ||
		ruleType == json_rule.ROW_CHANGE_NOTIFY
}

// extractChainStepSheets 从 CHAIN_REFERENCE 规则的 chainSteps 参数中提取所有引用的表名
// 后端自主解析，不依赖前端传递 chainRequiredSheets 参数
// 委托给 chain_reference 包实现
func extractChainStepSheets(params map[string]string) []string {
	return chain_reference.ExtractChainStepSheets(params)
}

// supplementDefaultTableRules 为有默认表级规则但没有 JSON 文件的表创建 SheetRule
// 这样可以确保代码中注册的表级规则（如 Hero 的 HERO_DROP_CHECK）能够被执行
//
// 参数:
//   - rules: 已从 JSON 文件加载的规则列表
//   - sheetMap: 所有 Excel 表的映射
//
// 返回:
//   - 补充后的规则列表
func supplementDefaultTableRules(rules []*json_rule.SheetRule, sheetMap map[string]*excelize.File) []*json_rule.SheetRule {
	// 构建已加载的表名 -> SheetRule 映射
	loadedSheetMap := make(map[string]*json_rule.SheetRule)
	for _, rule := range rules {
		loadedSheetMap[rule.Sheet] = rule
	}

	// 遍历所有 Excel 表，检查是否有默认表级规则
	supplemented := rules
	for sheetName := range sheetMap {
		// 检查是否有默认表级规则
		defaultRuleTypes := json_rule.GetDefaultTableRulesForSheet(sheetName)
		if len(defaultRuleTypes) == 0 {
			continue
		}

		if existingRule, ok := loadedSheetMap[sheetName]; ok {
			// 表已加载，检查是否缺少默认表级规则
			existingTypes := make(map[json_rule.ETableRule]bool)
			for _, tr := range existingRule.TableRules {
				existingTypes[tr.Type] = true
			}

			added := 0
			for _, ruleType := range defaultRuleTypes {
				if existingTypes[ruleType] {
					continue // 已存在，跳过
				}
				meta := json_rule.TableRuleMetas[ruleType]
				if meta == nil {
					continue
				}
				existingRule.TableRules = append(existingRule.TableRules, &json_rule.TableRule{
					Type:        ruleType,
					DisplayName: meta.DisplayName,
					Uuid:        "",
					Description: meta.Description,
					Params:      meta.ResolveParams(nil),
					Enabled:     true,
				})
				added++
			}
			if added > 0 {
				fmt.Printf("自动补充规则: %s (%d 条表级规则)\n", sheetName, added)
			}
		} else {
			// 表未加载，创建新的 SheetRule
			newRule := &json_rule.SheetRule{
				Sheet:       sheetName,
				ManagerList: nil,
				Rules:       make(map[string]*json_rule.SheetColRule),
				TableRules:  make([]*json_rule.TableRule, 0, len(defaultRuleTypes)),
			}

			for _, ruleType := range defaultRuleTypes {
				meta := json_rule.TableRuleMetas[ruleType]
				if meta == nil {
					continue
				}

				newRule.TableRules = append(newRule.TableRules, &json_rule.TableRule{
					Type:        ruleType,
					DisplayName: meta.DisplayName,
					Uuid:        "",
					Description: meta.Description,
					Params:      meta.ResolveParams(nil),
					Enabled:     true,
				})
			}

			supplemented = append(supplemented, newRule)
			fmt.Printf("自动补充规则: %s (%d 条表级规则)\n", sheetName, len(newRule.TableRules))
		}
	}

	return supplemented
}

// supplementDefaultParams 为已有 JSON 配置的 TableRule 补充 ParamDefs 中定义的默认参数
// 仅填充 TableRule.Params 中缺失的 key，不覆盖已有值
//
// 解决的问题：supplementDefaultTableRules 采用"全有或全无"策略，
// 跳过已有 JSON 配置的表。但 JSON 配置中的 TableRule.Params 可能为空，
// 导致检查器读取不到 ParamDefs 定义的默认值（如 ArenaSeason 的 timeRangeBefore="168h"）。
//
// 参数:
//   - rules: 规则列表（已包含 supplementDefaultTableRules 补充的规则）
func SupplementDefaultParams(rules []*json_rule.SheetRule) {
	for _, rule := range rules {
		for _, tableRule := range rule.TableRules {
			if !tableRule.Enabled {
				continue
			}
			meta := json_rule.TableRuleMetas[tableRule.Type]
			if meta == nil || len(meta.ParamDefs) == 0 {
				continue
			}
			if tableRule.Params == nil {
				tableRule.Params = make(map[string]string)
			}
			for _, def := range meta.ParamDefs {
				key := string(def.Key)
				if _, exists := tableRule.Params[key]; !exists && def.Default != "" {
					tableRule.Params[key] = def.Default
				}
			}
		}
	}
}

// deepCopyRules 深拷贝规则列表，隔离 TableRule.Params 的写入
// 用于 merge 场景：handleMergeCommit 中 allRules 被多个 commit 共享，
// supplementDefaultParams 会修改 TableRule.Params map，深拷贝确保各 commit 互不影响。
//
// 拷贝范围：
//   - SheetRule 结构体：值拷贝（Sheet/ManagerList/Rules/TableRules 字段各自拷贝）
//   - TableRules 切片：每个 TableRule 指针独立（新 *TableRule）
//   - TableRule.Params map：新 map + 逐项复制（唯一会被 supplementDefaultParams 修改的字段）
//
// 不拷贝的部分（流程中仅读取不修改）：
//   - Rules map（列级规则，map[string]*SheetColRule）
//   - ManagerList 指针
func deepCopyRules(rules []*json_rule.SheetRule) []*json_rule.SheetRule {
	result := make([]*json_rule.SheetRule, len(rules))
	for i, rule := range rules {
		newRule := *rule // 值拷贝 SheetRule

		// 深拷贝 TableRules（supplementDefaultParams 会修改其中的 Params map）
		newRule.TableRules = make([]*json_rule.TableRule, len(rule.TableRules))
		for j, tr := range rule.TableRules {
			newTR := *tr // 值拷贝 TableRule
			// 深拷贝 Params map
			if tr.Params != nil {
				newTR.Params = make(map[string]string, len(tr.Params))
				for k, v := range tr.Params {
					newTR.Params[k] = v
				}
			}
			newRule.TableRules[j] = &newTR
		}

		// Rules map 不拷贝（列级规则在流程中仅读取，不被修改）

		result[i] = &newRule
	}
	return result
}

// filterSheetMapByRules 根据规则列表过滤 sheetMap，只保留规则中涉及的 Sheet
// 包括：rule.Sheet 本身 + 表级规则的 RequiredSheets + 列级规则的跨表引用参数 + CHAIN_REFERENCE 的 chainSteps 引用表
// 用于"执行分类"场景：前端传入限定规则时，后端只加载相关 Sheet 的数据
func filterSheetMapByRules(sheetMap map[string]*excelize.File, rules []*json_rule.SheetRule) map[string]*excelize.File {
	relevantSheets := make(map[string]bool)

	for _, rule := range rules {
		// 1. 规则本身的 Sheet
		relevantSheets[rule.Sheet] = true

		// 2. 表级规则的 RequiredSheets
		for _, tableRule := range rule.TableRules {
			if !tableRule.Enabled {
				continue
			}
			meta := json_rule.TableRuleMetas[tableRule.Type]
			if meta == nil {
				continue
			}
			for _, sheet := range meta.RequiredSheets {
				relevantSheets[sheet] = true
			}
		}

		// 3. 列级规则中的跨表引用参数
		for _, colRule := range rule.Rules {
			for _, cr := range colRule.PropRules {
				for _, paramKey := range json_rule.ReferenceSheetParamKeys {
					if refSheet, ok := cr.Params[paramKey]; ok && refSheet != "" {
						relevantSheets[refSheet] = true
					}
				}

				// 4. CHAIN_REFERENCE: 直接从 chainSteps JSON 中提取引用的表名
				if cr.Params["chainSteps"] != "" {
					for _, sheet := range extractChainStepSheets(cr.Params) {
						relevantSheets[sheet] = true
					}
				}
			}
		}
	}

	filtered := make(map[string]*excelize.File)
	for sheetName, file := range sheetMap {
		if isRelevantSheet(sheetName, relevantSheets) {
			filtered[sheetName] = file
		}
	}

	fmt.Printf("[过滤] sheetMap %d→%d（根据预加载规则过滤）\n", len(sheetMap), len(filtered))
	return filtered
}

// isRelevantSheet 检查 sheetName 是否在相关的短名集合中
// 支持精确匹配和后缀匹配（如 "道具表|Item" 可通过 "Item" 匹配）
func isRelevantSheet(sheetName string, relevantSheets map[string]bool) bool {
	if relevantSheets[sheetName] {
		return true
	}
	// 后缀匹配："道具表|Item" 匹配 "Item"
	for shortName := range relevantSheets {
		if strings.HasSuffix(sheetName, "|"+shortName) {
			return true
		}
	}
	return false
}
