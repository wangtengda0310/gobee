package main

// 前端静态资源嵌入。
// 必须在根目录定义，因为 go:embed 不支持 ../ 路径。

import "embed"

//go:embed all:frontend/dist
var assets embed.FS
