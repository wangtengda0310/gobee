// Package sheetcmp 是表格双向比对与合并工具的根包。
// 根包集中放置 go:embed 资源 (README 与前端构建产物), 供 GUI 主程序 (cmd/sheetcmp)
// 与 CLI (cmd/sheetcmp-merge) 共同引用; 桌面入口见 cmd/sheetcmp。
package sheetcmp

import (
	"embed"

	_ "embed"
)

//go:embed README.md
var readme string

// Readme 返回内嵌的项目 README 全文 (CLI -doc 输出用)。
func Readme() string { return readme }

// Assets 前端构建产物 (frontend/dist), GUI 主程序作为 wails 资源服务。
// 约定: dist 随仓库提交 (与 winboot 同模式), 保证新克隆 go build 即可用。
//
//go:embed all:frontend/dist
var Assets embed.FS
