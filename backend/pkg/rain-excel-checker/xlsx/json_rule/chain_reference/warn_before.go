// Package chain_reference 提供关系链检查（CHAIN_REFERENCE）的公共数据结构和执行引擎
// 本文件实现预警窗口过滤：当违规目标的生效时间距今超过 chainWarnBefore 时静默
package chain_reference

import (
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"github.com/xuri/excelize/v2"
)

// ShouldSuppressByWarnBefore 判断是否因预警窗口而静默违规
// 当所有预警时间值都距今超过 warnBefore 时返回 true（静默）
// 返回 false 表示不静默，正常报错
func ShouldSuppressByWarnBefore(ctx *ChainContext, now time.Time) bool {
	warnBefore := ctx.WarnBefore()
	if warnBefore <= 0 {
		return false
	}
	if len(ctx.WarnValues) == 0 {
		// 链路断开时 WarnValues 为空，尝试从当前行值独立查找预警时间
		return shouldSuppressByCurrentValue(ctx, now)
	}

	// 解析所有 WarnValues 为时间，找最近的未来时间
	var nearest time.Time
	for _, v := range ctx.WarnValues {
		t := helpers.ParseDate(v)
		if t.IsZero() || t.Before(now) {
			continue
		}
		if nearest.IsZero() || t.Before(nearest) {
			nearest = t
		}
	}

	if nearest.IsZero() {
		return false
	}

	return nearest.Sub(now) > warnBefore
}

// ShouldSuppressWarnBeforeLegacy 旧路径的预警窗口过滤（无 ChainContext）
// 从 sheetMap 中全表扫描指定表/列，找最近未来时间判断是否超过 warnBefore
func ShouldSuppressWarnBeforeLegacy(sheetMap map[string]*excelize.File, warnSheet, warnCol string, warnBefore time.Duration, now time.Time) bool {
	if warnBefore <= 0 || warnSheet == "" || warnCol == "" {
		return false
	}

	targetFile, targetSheetName, found := helpers.FindSheetBySuffix(sheetMap, warnSheet)
	if !found {
		return false
	}

	targetCols, err := targetFile.GetCols(targetSheetName)
	if err != nil {
		return false
	}

	warnColIdx := helpers.GetColIndexByName(targetCols, warnCol)
	if warnColIdx < 0 {
		return false
	}

	var nearest time.Time
	dataStart := excelio.MJS_FIXED_ROWS_NUM
	dataEnd := helpers.GetDataEndIndex(targetCols, dataStart)
	for rowIdx := dataStart; rowIdx < dataEnd; rowIdx++ {
		v := strings.TrimSpace(targetCols[warnColIdx][rowIdx])
		if v == "" {
			continue
		}
		t := helpers.ParseDate(v)
		if t.IsZero() || t.Before(now) {
			continue
		}
		if nearest.IsZero() || t.Before(nearest) {
			nearest = t
		}
	}

	if nearest.IsZero() {
		return false
	}

	return nearest.Sub(now) > warnBefore
}

// shouldSuppressByCurrentValue 链路断开时的逐行预警过滤
// 用当前单元格值（如 SeasonPassId="5"）在 warnSheet 的 Id 列查找匹配行，提取 warnCol 值
func shouldSuppressByCurrentValue(ctx *ChainContext, now time.Time) bool {
	warnSheet := ctx.WarnSheet()
	warnCol := ctx.WarnCol()
	if warnSheet == "" || warnCol == "" {
		return false
	}

	targetFile, targetSheetName, found := helpers.FindSheetBySuffix(ctx.SheetMap, warnSheet)
	if !found {
		return false
	}

	targetCols, err := targetFile.GetCols(targetSheetName)
	if err != nil {
		return false
	}

	idColIdx := helpers.GetColIndexByName(targetCols, "Id")
	if idColIdx < 0 {
		return false
	}

	warnColIdx := helpers.GetColIndexByName(targetCols, warnCol)
	if warnColIdx < 0 {
		return false
	}

	currentVal := ctx.CurrentCellValue()
	dataStart := excelio.MJS_FIXED_ROWS_NUM
	dataEnd := helpers.GetDataEndIndex(targetCols, dataStart)

	for rowIdx := dataStart; rowIdx < dataEnd; rowIdx++ {
		idVal := strings.TrimSpace(helpers.GetColValue(targetCols, idColIdx, rowIdx))
		if idVal != currentVal {
			continue
		}
		warnVal := strings.TrimSpace(helpers.GetColValue(targetCols, warnColIdx, rowIdx))
		if warnVal == "" {
			continue
		}
		t := helpers.ParseDate(warnVal)
		if t.IsZero() || t.Before(now) {
			return false
		}
		return t.Sub(now) > ctx.WarnBefore()
	}

	return false
}
