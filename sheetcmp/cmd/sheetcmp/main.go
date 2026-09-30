// sheetcmp — 表格双向比对与合并工具的桌面入口 (Wails v3 + Vue3)。
// 服务层见 internal/service (Compare 比对), 前端见 frontend/ (构建产物内嵌于根包 Assets)。
package main

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	sheetcmp "github.com/wangtengda0310/gobee/sheetcmp"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/service"
)

func main() {
	// 创建应用: 绑定比对服务, 资源服务内嵌的前端构建产物
	app := application.New(application.Options{
		Name:        "sheetcmp",
		Description: "表格/文本双向比对与合并工具",
		Services: []application.Service{
			application.NewService(&service.CompareService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(sheetcmp.Assets),
		},
	})

	// 主窗口: 双栏对比界面需要宽窗口
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "sheetcmp — 表格双向比对",
		Width:  1400,
		Height: 860,
		URL:    "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
