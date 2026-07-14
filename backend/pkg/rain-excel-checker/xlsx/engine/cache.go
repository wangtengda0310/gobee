// Package engine 提供检查结果缓存功能
// 本包缓存最近一次配表检查结果，供其他页面（如活动 Wiki）读取错误计数
package engine

import (
	"fmt"
	"sync"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"
)

// cachedCheckResults 缓存最近一次配表检查结果，供其他页面（如活动 Wiki）读取错误计数
var cachedCheckResults struct {
	sync.RWMutex
	ColResults   []*json_rule.ColCheckResult
	TableResults []*json_rule.TableCheckResult
}

// StoreCheckResults 缓存配表检查结果，供活动 Wiki 等页面读取错误计数
func StoreCheckResults(colResults []*json_rule.ColCheckResult, tableResults []*json_rule.TableCheckResult) {
	cachedCheckResults.Lock()
	defer cachedCheckResults.Unlock()
	cachedCheckResults.ColResults = colResults
	cachedCheckResults.TableResults = tableResults

	// [DEBUG] 日志：记录缓存写入
	colWithErr := 0
	for _, cr := range colResults {
		if cr != nil && len(cr.ErrCells) > 0 {
			colWithErr++
		}
	}
	tableWithErr := 0
	for _, tr := range tableResults {
		if tr != nil && len(tr.ErrCells) > 0 {
			tableWithErr++
		}
	}
	fmt.Printf("[StoreCheckResults] 缓存写入: ColResults=%d(有错误=%d), TableResults=%d(有错误=%d)\n",
		len(colResults), colWithErr, len(tableResults), tableWithErr)

	// [DEBUG] 日志：打印有错误的列级结果的 sheet+field
	for _, cr := range colResults {
		if cr != nil && len(cr.ErrCells) > 0 {
			sheetName := "<nil>"
			if cr.SheetName != nil {
				sheetName = *cr.SheetName
			}
			colName := "<nil>"
			if cr.ColName != nil {
				colName = *cr.ColName
			}
			fmt.Printf("[StoreCheckResults]   ColErr: sheet=%s, col=%s, errCells=%d\n", sheetName, colName, len(cr.ErrCells))
		}
	}
}

// GetCachedCheckResults 读取缓存的检查结果
func GetCachedCheckResults() ([]*json_rule.ColCheckResult, []*json_rule.TableCheckResult) {
	cachedCheckResults.RLock()
	defer cachedCheckResults.RUnlock()

	// [DEBUG] 日志：记录缓存读取
	colWithErr := 0
	for _, cr := range cachedCheckResults.ColResults {
		if cr != nil && len(cr.ErrCells) > 0 {
			colWithErr++
		}
	}
	fmt.Printf("[GetCachedCheckResults] 缓存读取: ColResults=%d(有错误=%d), TableResults=%d\n",
		len(cachedCheckResults.ColResults), colWithErr, len(cachedCheckResults.TableResults))

	return cachedCheckResults.ColResults, cachedCheckResults.TableResults
}
