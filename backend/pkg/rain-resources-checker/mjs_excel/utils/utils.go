package utils

import (
	"strconv"
	"strings"
)

func GetIntArr(strArrLine string) []int {
	strArr := strings.Split(strArrLine, ",")
	intArr := make([]int, 0, len(strArr))
	for _, l := range strArr {
		if id, err := strconv.Atoi(l); err != nil {
			continue
		} else {
			intArr = append(intArr, id)
		}
	}
	return intArr
}

// 安全获取单元格值的辅助函数
func GetCellValue(cols [][]string, colIdx, rowIdx int) string {
	if colIdx < 0 || colIdx >= len(cols) {
		return ""
	}
	if rowIdx < 0 || rowIdx >= len(cols[colIdx]) {
		return ""
	}
	return cols[colIdx][rowIdx]
}
