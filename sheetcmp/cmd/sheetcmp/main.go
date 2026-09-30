// sheetcmp — 表格双向比对与合并工具的桌面入口 (Wails v3 + Vue3)。
// 服务层见 internal/service (Compare 比对/Apply 写回), 前端见 frontend/。
//
// 用法:
//
//	sheetcmp                                    # 普通模式: 手动选文件比对
//	sheetcmp -MergeTool <remote> <merged>       # 合并模式 (git mergetool 形态):
//	  左=REMOTE (对方改动), 右=MERGED (本地工作副本);
//	  「完成合并」= 保存写回 + 退出码 0 (trustExitCode);
//	  直接关窗 = 放弃, 退出码 1 (git 保留冲突状态)。
package main

import (
	"flag"
	"log"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	sheetcmp "github.com/wangtengda0310/gobee/sheetcmp"
	"github.com/wangtengda0310/gobee/sheetcmp/internal/service"
)

func main() {
	mergeTool := flag.Bool("MergeTool", false, "合并模式: 需跟两个文件参数 REMOTE MERGED")
	flag.Parse()
	args := flag.Args()

	// 服务实例: merge 模式注入预载文件; 退出码默认 merge=1 (放弃) / normal=0
	svc := &service.CompareService{}
	exitCode := 0
	if *mergeTool {
		if len(args) != 2 {
			log.Fatal("-MergeTool 需要 REMOTE 与 MERGED 两个文件参数")
		}
		svc.Startup = &service.StartupInfo{
			Mode:      "merge",
			LeftPath:  args[0], // REMOTE: 对方改动
			RightPath: args[1], // MERGED: 本地工作副本 (写回目标)
		}
		exitCode = 1 // 未点「完成合并」即退出 = 放弃
	}

	app := application.New(application.Options{
		Name:        "sheetcmp",
		Description: "表格/文本双向比对与合并工具",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(sheetcmp.Assets),
		},
	})

	// 受控退出: 记录退出码后结束应用主循环 (main 尾部统一 os.Exit)
	svc.QuitFn = func(code int) {
		exitCode = code
		app.Quit()
	}

	title := "sheetcmp — 表格双向比对"
	if *mergeTool {
		title = "sheetcmp — 合并模式 (" + filepathBase(args[0]) + " → " + filepathBase(args[1]) + ")"
	}
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  title,
		Width:  1400,
		Height: 860,
		URL:    "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(exitCode)
}

// filepathBase 空安全取文件基名 (窗口标题用)。
func filepathBase(p string) string {
	if i := len(p) - 1; i >= 0 {
		for j := i; j >= 0; j-- {
			if p[j] == '\\' || p[j] == '/' {
				return p[j+1:]
			}
		}
	}
	return p
}
