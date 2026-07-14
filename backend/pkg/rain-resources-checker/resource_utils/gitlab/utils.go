package gitlab

import "fmt"

// CheckFileExistsFromList 检查特定文件是否存在 - 从已获取的列表中检查
// dirPath := "Master/Card/Audio/Voice"
func CheckFileExistsFromList(filename, dirPath string, files []FileInfo) bool {
	fullPath := fmt.Sprintf("%s/%s", dirPath, filename)

	for _, file := range files {
		if file.Path == fullPath || file.Name == filename {
			return true
		}
	}
	return false
}
