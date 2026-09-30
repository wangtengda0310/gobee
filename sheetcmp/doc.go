// Package sheetcmp 是表格双向比对与合并工具的根包。
// 桌面入口 main.go 见后续阶段; 本文件仅提供内嵌 README 供 CLI -doc 参数查看。
package sheetcmp

import _ "embed"

//go:embed README.md
var readme string

// Readme 返回内嵌的项目 README 全文 (CLI -doc 输出用)。
func Readme() string { return readme }
