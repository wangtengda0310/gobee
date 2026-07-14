// Package excel_internal 提供 Excel 文件读取和解析功能
// 本包负责从文件系统读取 Excel 文件并构建 Sheet 映射
package excelio

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

// GetSheetMap 从指定目录读取所有 Excel 文件，构建 Sheet 名称到文件对象的映射
//
// 此函数执行以下操作：
// 1. 读取目录下的所有 Excel 文件（支持文件名过滤）
// 2. 过滤出符合项目规范的 Sheet
// 3. 构建 Sheet 名称 -> Excel 文件对象的映射
//
// 参数：
//   - dir: Excel 文件所在目录路径
//   - sheetNamesFilter: 可选的文件名过滤列表，只处理指定文件名的 Excel
//
// 返回值：
//   - sheetMap: Sheet 名称到 Excel 文件对象的映射
//   - err: 读取或解析过程中的错误
//
// 注意：
//   - 返回的 sheetMap 中的 *excelize.File 对象由调用方管理生命周期
//   - 调用方使用完毕后应调用 File.Close() 释放资源
//   - 与 GetSheetMapFromBytes 保持一致的生命周期管理策略
func GetSheetMap(dir string, sheetNamesFilter ...string) (sheetMap map[string]*excelize.File, err error) {
	// 读取目录下的所有 Excel 文件
	excels, err := ReadFileOrDir(dir, sheetNamesFilter...)
	if err != nil {
		return nil, err
	}
	// 注意：不再 defer Close 所有文件
	// 原因：返回的 sheetMap 中的 value 指向这些 File 对象
	// 如果在这里关闭，调用方使用时会操作已关闭的 File 对象
	// File 对象的生命周期由调用方管理（与 GetSheetMapFromBytes 保持一致）

	// 检查是否成功读取到 Excel 文件
	if len(excels) == 0 {
		return nil, errors.New("excel file is empty")
	}

	fmt.Println("加载了", len(excels), "个excel文件")

	// 过滤出符合规范的 Sheet
	filter, err := ExcelFilter(excels)
	if err != nil {
		return nil, err
	}

	// 构建 Sheet 名称到文件对象的映射
	// 因为文件名并不完全是表名，所以需要遍历每个文件里的 Sheet 名
	sheetMap = make(map[string]*excelize.File)
	for file, sheets := range filter {
		for _, sheet := range sheets {
			sheetMap[sheet.Name] = file
		}
	}

	return sheetMap, nil
}

// GetSheetMapFromBytes 从字节数据构建 Sheet 映射（不读本地磁盘）
// 用于 merge 遍历场景：通过 git show 获取历史版本后构建 sheetMap
//
// 参数：
//   - files: git 相对路径 → 文件内容 ([]byte) 的映射，key 如 "excel/Hero.xlsx"
//   - repoPath: git 仓库根目录的绝对路径
//
// file.Path 会被设置为 repoPath + gitRelPath 拼接的绝对路径，
// 确保通用规则（NEW_ROW_NOTIFY）内部调用 GetFileAtCommit 时路径转换正确。
func GetSheetMapFromBytes(files map[string][]byte, repoPath string) (map[string]*excelize.File, error) {
	var excels []*excelize.File

	for gitRelPath, data := range files {
		f, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			fmt.Printf("[警告] 解析 git 文件 %s 失败: %v\n", gitRelPath, err)
			continue
		}
		// 设置绝对路径，确保 GetFileAtCommit 的 filepath.Abs→filepath.Rel 转换正确
		f.Path = filepath.Join(repoPath, gitRelPath)
		excels = append(excels, f)
	}

	if len(excels) == 0 {
		return nil, errors.New("没有可解析的 Excel 文件")
	}

	// ⚠️ 移除 defer 关闭逻辑
	// 原因：返回的 sheetMap 中的 value 指向这些 File 对象
	// 如果在这里关闭，调用方使用缓存时会操作已关闭的 File 对象
	// File 对象的生命周期由调用方管理（见 main.go 的 handleMergeCommit）
	//
	// defer func() {
	//     for _, excel := range excels {
	//         excel.Close()
	//     }
	// }()

	fmt.Println("从 git 历史加载了", len(excels), "个 Excel 文件")

	// 过滤出符合规范的 Sheet
	filter, err := ExcelFilter(excels)
	if err != nil {
		return nil, err
	}

	// 构建 Sheet 名称到文件对象的映射
	sheetMap := make(map[string]*excelize.File)
	for file, sheets := range filter {
		for _, sheet := range sheets {
			sheetMap[sheet.Name] = file
		}
	}

	return sheetMap, nil
}
