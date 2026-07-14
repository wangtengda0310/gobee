// Package engine 中的 column_sheetmap 实现字段卡片「执行检查」的 Excel 按需加载。
//
// CheckSingleColumn 需要 sheetMap 做跨表引用校验，但不能 GetSheetMap(dir) 全量加载
// 整个配表目录（200+ 文件）。加载策略与「执行分类」的列级引用收集对齐：
//
//  1. 从列规则收集关联表名（含 chainRequiredSheets 逗号分隔、chainSteps JSON）
//  2. 按表名启发式筛选 xlsx 文件名后首轮加载
//  3. 若 sheetMap 仍缺表，仅补载缺失表对应的文件（禁止回退全量）
//
// 表名与文件名：规则里常见「赛季战令表|SeasonPass」，磁盘为 SeasonPass_赛季战令表.xlsx，
// 需将 | 后的英文短名展开后再做文件名匹配（见 expandRequiredSheetsForFileMatch）。
package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
	"github.com/xuri/excelize/v2"
)

// CollectRelevantSheetsForColumn 从单列规则收集检查所需的关联表名
// 包括：目标 Sheet + 列级跨表引用参数 + CHAIN_REFERENCE 的 chainSteps 引用表
func CollectRelevantSheetsForColumn(sheetName string, colRule *json_rule.SheetColRule) map[string]bool {
	relevantSheets := map[string]bool{sheetName: true}
	if colRule == nil {
		return relevantSheets
	}
	for _, cr := range colRule.PropRules {
		for _, paramKey := range json_rule.ReferenceSheetParamKeys {
			if refSheet, ok := cr.Params[paramKey]; ok && refSheet != "" {
				addSheetNamesFromParam(relevantSheets, refSheet)
			}
		}
		if cr.Params["chainSteps"] != "" {
			for _, sheet := range extractChainStepSheets(cr.Params) {
				relevantSheets[sheet] = true
			}
		}
	}
	return relevantSheets
}

// addSheetNamesFromParam 解析跨表参数中的表名。
// chainRequiredSheets 等参数由前端填充，值为逗号分隔，如：
// "赛季战令表|SeasonPass,赛季战令奖励表|SeasonPassReward"
func addSheetNamesFromParam(sheets map[string]bool, value string) {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			sheets[part] = true
		}
	}
}

// expandRequiredSheetsForFileMatch 展开表名用于 xlsx 文件名匹配
// 将「赛季战令表|SeasonPass」额外展开为 SeasonPass，以便匹配 SeasonPass_赛季战令表.xlsx
func expandRequiredSheetsForFileMatch(requiredSheets map[string]bool) map[string]bool {
	expanded := make(map[string]bool, len(requiredSheets)*2)
	for sheet := range requiredSheets {
		expanded[sheet] = true
		if idx := strings.LastIndex(sheet, "|"); idx >= 0 {
			suffix := strings.TrimSpace(sheet[idx+1:])
			if suffix != "" {
				expanded[suffix] = true
			}
		}
	}
	return expanded
}

// filterXlsxFileNamesByRequiredSheets 根据关联表名启发式筛选待加载的 xlsx 文件名。
// 实际匹配键经 expandRequiredSheetsForFileMatch 展开，再委托 mayFileContainRequiredSheet
// （与 git 增量场景 supplementFromCommit 使用同一套文件名启发式）。
func filterXlsxFileNamesByRequiredSheets(allNames []string, requiredSheets map[string]bool) []string {
	if len(requiredSheets) == 0 {
		return allNames
	}
	matchKeys := expandRequiredSheetsForFileMatch(requiredSheets)
	filtered := make([]string, 0, len(requiredSheets)+1)
	for _, name := range allNames {
		fileName := strings.TrimSuffix(name, ".xlsx")
		if mayFileContainRequiredSheet(fileName, matchKeys) {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

// findMissingRequiredSheets 检查 sheetMap 中是否包含所有需要的表（支持后缀匹配）
func findMissingRequiredSheets(sheetMap map[string]*excelize.File, requiredSheets map[string]bool) []string {
	var missing []string
	for sheet := range requiredSheets {
		found := false
		for name := range sheetMap {
			if isRelevantSheet(name, map[string]bool{sheet: true}) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, sheet)
		}
	}
	return missing
}

// loadedXlsxBaseNames 返回 sheetMap 中已加载文件的基础文件名集合
func loadedXlsxBaseNames(sheetMap map[string]*excelize.File) map[string]bool {
	names := make(map[string]bool)
	for _, f := range sheetMap {
		if f == nil || f.Path == "" {
			continue
		}
		names[filepath.Base(f.Path)] = true
	}
	return names
}

// mergeSheetMap 将 src 中缺失的 sheet 合并到 dest
func mergeSheetMap(dest, src map[string]*excelize.File) {
	for sheetName, f := range src {
		if _, exists := dest[sheetName]; !exists {
			dest[sheetName] = f
		}
	}
}

// closeSheetMap 关闭 sheetMap 中所有 Excel 文件（同一 *excelize.File 可能被多个 sheet 引用，去重后只 Close 一次）
func closeSheetMap(sheetMap map[string]*excelize.File) {
	closed := make(map[*excelize.File]bool)
	for _, f := range sheetMap {
		if f == nil || closed[f] {
			continue
		}
		closed[f] = true
		_ = f.Close()
	}
}

// supplementSheetMapForMissing 仅补载缺失关联表对应的 xlsx 文件。
// 避免首轮文件名启发式过窄时回退全量 GetSheetMap(dir)；只打开尚未加载过的文件。
func supplementSheetMapForMissing(sheetMap map[string]*excelize.File, dir string, allNames []string, missing []string) error {
	if len(missing) == 0 {
		return nil
	}
	missingSet := make(map[string]bool, len(missing))
	for _, s := range missing {
		missingSet[s] = true
	}
	extraNames := filterXlsxFileNamesByRequiredSheets(allNames, missingSet)
	loaded := loadedXlsxBaseNames(sheetMap)
	var toLoad []string
	for _, name := range extraNames {
		if !loaded[name] {
			toLoad = append(toLoad, name)
		}
	}
	if len(toLoad) == 0 {
		return nil
	}
	extraMap, err := excelio.GetSheetMap(dir, toLoad...)
	if err != nil {
		return err
	}
	mergeSheetMap(sheetMap, extraMap)
	return nil
}

// GetSheetMapForColumnCheck 为单列检查按需加载 Excel，只读取与规则相关的文件。
// 由 backend/pkg/excel-test.CheckSingleColumn 调用。
func GetSheetMapForColumnCheck(dir string, sheetName string, colRule *json_rule.SheetColRule) (map[string]*excelize.File, error) {
	requiredSheets := CollectRelevantSheetsForColumn(sheetName, colRule)

	// 先列目录内所有 xlsx 文件名（不打开文件），再按关联表筛选后加载
	allNames, err := excelio.ListXlsxFileNames(dir)
	if err != nil {
		return nil, err
	}
	if len(allNames) == 0 {
		return nil, fmt.Errorf("excel file is empty")
	}

	filtered := filterXlsxFileNamesByRequiredSheets(allNames, requiredSheets)
	if len(filtered) == 0 {
		// 启发式未命中任何文件名时兜底：避免返回空 sheetMap
		filtered = allNames
	}

	sheetMap, err := excelio.GetSheetMap(dir, filtered...)
	if err != nil {
		return nil, err
	}

	if missing := findMissingRequiredSheets(sheetMap, requiredSheets); len(missing) > 0 {
		fmt.Printf("[单列检查] 首轮加载缺少关联表 %v，尝试补载\n", missing)
		if err := supplementSheetMapForMissing(sheetMap, dir, allNames, missing); err != nil {
			closeSheetMap(sheetMap)
			return nil, fmt.Errorf("补载关联表失败: %w", err)
		}
		if stillMissing := findMissingRequiredSheets(sheetMap, requiredSheets); len(stillMissing) > 0 {
			closeSheetMap(sheetMap)
			return nil, fmt.Errorf("缺少关联表且无法从磁盘加载: %v", stillMissing)
		}
	}

	fmt.Printf("[单列检查] 按需加载 %d/%d 个 Excel 文件（关联表 %d 个）\n",
		len(loadedXlsxBaseNames(sheetMap)), len(allNames), len(requiredSheets))
	return sheetMap, nil
}
