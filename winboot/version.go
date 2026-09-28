package main

import "strings"

// TOOL_VERSION 工具版本号 (自动更新比对基准).
const TOOL_VERSION = "0.2.0"

// VersionGt 语义化版本比较 a > b ('1.10.0' > '1.9.1'); 非数字段按 0 兜底.
func VersionGt(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var x, y int
		if i < len(pa) {
			x = atoiSafe(pa[i])
		}
		if i < len(pb) {
			y = atoiSafe(pb[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0 // 非数字段兜底
		}
		n = n*10 + int(c-'0')
	}
	return n
}
