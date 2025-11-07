package load

import (
	"archive/zip"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

// Zip loads data from a ZIP file and returns it as a slice of maps
func Zip(file string) []dump.Record {
	// 打开 ZIP 文件
	reader, err := zip.OpenReader(file)
	if err != nil {
		panic("failed to open zip file: " + err.Error())
	}
	defer reader.Close()

	// 创建结果切片
	var result []dump.Record

	// 按目录分组文件（每个目录代表一条记录）
	recordGroups := make(map[string][]*zip.File)

	// 遍历 ZIP 文件中的所有文件
	for _, zipFile := range reader.File {
		// 获取文件所在的目录
		dir := filepath.Dir(zipFile.Name)

		// 将文件按目录分组
		recordGroups[dir] = append(recordGroups[dir], zipFile)
	}

	// 处理每个记录组
	for _, files := range recordGroups {
		// 创建一个新的记录
		record := make(dump.Record)

		// 处理组中的每个文件
		for _, zipFile := range files {
			// 获取字段名（文件名）
			fieldName := filepath.Base(zipFile.Name)

			// 打开文件进行读取
			fileReader, err := zipFile.Open()
			if err != nil {
				log.Panic("failed to open file in zip: " + err.Error())
			}

			// 读取文件内容
			content, err := io.ReadAll(fileReader)
			fileReader.Close()
			if err != nil {
				log.Panic("failed to read file in zip: " + err.Error())
			}

			// 将字段和内容添加到记录中
			record[fieldName] = content
		}

		// 将记录添加到结果中
		result = append(result, record)
	}

	log.Println(len(result), "records loaded from zip file", file)
	return result
}

// IsZipFileExist checks if a ZIP file exists
func IsZipFileExist(path string) bool {
	// 尝试打开 ZIP 文件来检查是否存在
	_, err := zip.OpenReader(path)
	if err != nil {
		// 检查是否是文件不存在的错误
		if strings.Contains(err.Error(), "no such file or directory") ||
			strings.Contains(err.Error(), "The system cannot find the file specified") {
			return false
		}
		// 如果是其他错误（如不是有效的 ZIP 文件），仍然认为文件存在
		return true
	}

	// 成功打开 ZIP 文件，说明文件存在
	return true
}
