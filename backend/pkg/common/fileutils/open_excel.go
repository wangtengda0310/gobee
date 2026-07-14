package fileutils

import (
	"fmt"
	"os/exec"
	"runtime"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
)

// OpenExcel 打开指定 Sheet 对应的 Excel 文件
// sheetName: Sheet 名称（如"活动表|Activity"），为空时直接打开 filePathOrDir 作为文件路径
// filePathOrDir: Excel 配置目录路径，或当 sheetName 为空时为文件完整路径
func OpenExcel(sheetName string, filePathOrDir string) error {
	var filePath string
	if sheetName == "" {
		// 直接打开文件路径
		filePath = filePathOrDir
	} else {
		sheetMap, err := excelio.GetSheetMap(filePathOrDir)
		if err != nil {
			return err
		}

		file, ok := sheetMap[sheetName]
		if !ok {
			return fmt.Errorf("未找到Sheet: %s", sheetName)
		}
		filePath = file.Path
	}

	// 根据操作系统选择打开方式
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}

	return cmd.Start()
}
